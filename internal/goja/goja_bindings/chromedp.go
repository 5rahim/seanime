package goja_bindings

import (
	"context"
	"fmt"
	gojautil "seanime/internal/util/goja"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/dop251/goja"
	"github.com/rs/zerolog/log"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// ChromeDP
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

const (
	defaultBrowserTimeout = 30 * time.Second
)

type ChromeDP struct {
	vm         *goja.Runtime
	chromeSem  chan struct{}
	responseCh chan func()
	browsers   map[string]*Browser
	browserMu  sync.RWMutex
	closeOnce  sync.Once
}

// Browser represents a browser instance
type Browser struct {
	id          string
	chromedp    *ChromeDP
	allocCtx    context.Context
	ctx         context.Context
	cancel      context.CancelFunc
	allocCancel context.CancelFunc
	timeout     time.Duration
}

func NewChromeDP(vm *goja.Runtime) *ChromeDP {
	return &ChromeDP{
		vm:         vm,
		chromeSem:  make(chan struct{}, 5), // limit concurrent browser instances
		responseCh: make(chan func(), 10),
		browsers:   make(map[string]*Browser),
	}
}

func (c *ChromeDP) ResponseChannel() <-chan func() {
	return c.responseCh
}

func (c *ChromeDP) Close() {
	defer func() {
		if r := recover(); r != nil {
		}
	}()

	// Close all browsers
	c.browserMu.Lock()
	for _, browser := range c.browsers {
		browser.close()
	}
	c.browserMu.Unlock()

	c.closeOnce.Do(func() {
		close(c.responseCh)
	})
}

type chromeOptions struct {
	Timeout      int    // seconds
	WaitSelector string // CSS selector to wait for
	WaitDuration int    // milliseconds to wait after page load
	UserAgent    string // custom user agent
	Headless     bool   // run in headless mode (default: true)
}

// BindChromeDP binds the ChromeDP utilities to the VM
func BindChromeDP(vm *goja.Runtime) *ChromeDP {
	return BindChromeDPWithScheduler(vm, nil)
}

// BindChromeDPWithScheduler binds the ChromeDP utilities to the VM
func BindChromeDPWithScheduler(vm *goja.Runtime, scheduler *gojautil.Scheduler) *ChromeDP {
	c := NewChromeDP(vm)

	// Create ChromeDP object with methods
	chromeDPObj := vm.NewObject()

	// modular API
	chromeDPObj.Set("newBrowser", c.NewBrowser)

	// one-shot functions
	chromeDPObj.Set("scrape", c.Scrape)
	chromeDPObj.Set("screenshot", c.Screenshot)
	chromeDPObj.Set("evaluate", c.Evaluate)

	vm.Set("ChromeDP", chromeDPObj)

	// Start response handler
	go func() {
		for fn := range c.ResponseChannel() {
			if scheduler != nil {
				scheduler.ScheduleAsync(func() error {
					defer func() {
						if r := recover(); r != nil {
							log.Warn().Msgf("extension: chromedp response channel panic: %v", r)
						}
					}()
					fn()
					return nil
				})
			} else {
				defer func() {
					if r := recover(); r != nil {
						log.Warn().Msgf("extension: chromedp response channel panic: %v", r)
					}
				}()
				fn()
			}
		}
	}()

	return c
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Browser Instance API
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// NewBrowser creates a new browser instance
func (c *ChromeDP) NewBrowser(call goja.FunctionCall) goja.Value {
	opts := chromeOptions{
		Timeout:  int(defaultBrowserTimeout.Seconds()),
		Headless: true,
	}

	if len(call.Arguments) > 0 {
		if optObj := call.Argument(0).ToObject(c.vm); optObj != nil {
			if v := optObj.Get("timeout"); gojaValueIsDefined(v) {
				opts.Timeout = int(v.ToInteger())
			}
			if v := optObj.Get("userAgent"); gojaValueIsDefined(v) {
				opts.UserAgent = v.String()
			}
			if v := optObj.Get("headless"); gojaValueIsDefined(v) {
				opts.Headless = v.ToBoolean()
			}
		}
	}

	promise, resolve, reject := c.vm.NewPromise()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				c.responseCh <- func() {
					reject(NewErrorString(c.vm, fmt.Sprintf("chromedp panic: %v", r)))
				}
			}
		}()

		c.chromeSem <- struct{}{}

		browser, err := c.createBrowser(opts)
		if err != nil {
			<-c.chromeSem
			c.responseCh <- func() {
				reject(NewError(c.vm, err))
			}
			return
		}

		c.responseCh <- func() {
			browserObj := c.vm.NewObject()
			browserObj.Set("navigate", browser.Navigate)
			browserObj.Set("waitVisible", browser.WaitVisible)
			browserObj.Set("waitReady", browser.WaitReady)
			browserObj.Set("click", browser.Click)
			browserObj.Set("sendKeys", browser.SendKeys)
			browserObj.Set("evaluate", browser.EvaluateJS)
			browserObj.Set("innerHTML", browser.InnerHTML)
			browserObj.Set("outerHTML", browser.OuterHTML)
			browserObj.Set("text", browser.Text)
			browserObj.Set("attribute", browser.Attribute)
			browserObj.Set("screenshot", browser.Screenshot)
			browserObj.Set("fullScreenshot", browser.FullScreenshot)
			browserObj.Set("sleep", browser.Sleep)
			browserObj.Set("close", browser.Close)

			resolve(browserObj)
		}
	}()

	return c.vm.ToValue(promise)
}

func (c *ChromeDP) createBrowser(opts chromeOptions) (*Browser, error) {
	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", opts.Headless),
	)

	if opts.UserAgent != "" {
		allocOpts = append(allocOpts, chromedp.UserAgent(opts.UserAgent))
	}

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), allocOpts...)
	ctx, cancel := chromedp.NewContext(allocCtx)

	browser := &Browser{
		id:          fmt.Sprintf("browser-%d", time.Now().UnixNano()),
		chromedp:    c,
		allocCtx:    allocCtx,
		ctx:         ctx,
		cancel:      cancel,
		allocCancel: allocCancel,
		timeout:     time.Duration(opts.Timeout) * time.Second,
	}

	if err := chromedp.Run(ctx); err != nil {
		cancel()
		allocCancel()
		return nil, fmt.Errorf("failed to start browser: %w", err)
	}

	c.browserMu.Lock()
	c.browsers[browser.id] = browser
	c.browserMu.Unlock()

	return browser, nil
}

func (b *Browser) close() {
	if b.cancel != nil {
		b.cancel()
	}
	if b.allocCancel != nil {
		b.allocCancel()
	}
}

// Navigate navigates to a URL
func (b *Browser) Navigate(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		PanicThrowErrorString(b.chromedp.vm, "navigate requires a URL argument")
	}

	url := call.Argument(0).String()
	promise, resolve, reject := b.chromedp.vm.NewPromise()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				b.chromedp.responseCh <- func() {
					reject(NewErrorString(b.chromedp.vm, fmt.Sprintf("navigate panic: %v", r)))
				}
			}
		}()

		ctx, cancel := context.WithTimeout(b.ctx, b.timeout)
		defer cancel()

		err := chromedp.Run(ctx, chromedp.Navigate(url))
		if err != nil {
			b.chromedp.responseCh <- func() {
				reject(NewError(b.chromedp.vm, err))
			}
			return
		}

		b.chromedp.responseCh <- func() {
			resolve(goja.Undefined())
		}
	}()

	return b.chromedp.vm.ToValue(promise)
}

// WaitVisible waits for a selector to be visible
func (b *Browser) WaitVisible(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		PanicThrowErrorString(b.chromedp.vm, "waitVisible requires a selector argument")
	}

	selector := call.Argument(0).String()
	promise, resolve, reject := b.chromedp.vm.NewPromise()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				b.chromedp.responseCh <- func() {
					reject(NewErrorString(b.chromedp.vm, fmt.Sprintf("waitVisible panic: %v", r)))
				}
			}
		}()

		ctx, cancel := context.WithTimeout(b.ctx, b.timeout)
		defer cancel()

		err := chromedp.Run(ctx, chromedp.WaitVisible(selector))
		if err != nil {
			b.chromedp.responseCh <- func() {
				reject(NewError(b.chromedp.vm, err))
			}
			return
		}

		b.chromedp.responseCh <- func() {
			resolve(goja.Undefined())
		}
	}()

	return b.chromedp.vm.ToValue(promise)
}

// WaitReady waits for the page to be ready
func (b *Browser) WaitReady(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		PanicThrowErrorString(b.chromedp.vm, "waitReady requires a selector argument")
	}

	selector := call.Argument(0).String()
	promise, resolve, reject := b.chromedp.vm.NewPromise()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				b.chromedp.responseCh <- func() {
					reject(NewErrorString(b.chromedp.vm, fmt.Sprintf("waitReady panic: %v", r)))
				}
			}
		}()

		ctx, cancel := context.WithTimeout(b.ctx, b.timeout)
		defer cancel()

		err := chromedp.Run(ctx, chromedp.WaitReady(selector))
		if err != nil {
			b.chromedp.responseCh <- func() {
				reject(NewError(b.chromedp.vm, err))
			}
			return
		}

		b.chromedp.responseCh <- func() {
			resolve(goja.Undefined())
		}
	}()

	return b.chromedp.vm.ToValue(promise)
}

// Click clicks on an element
func (b *Browser) Click(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		PanicThrowErrorString(b.chromedp.vm, "click requires a selector argument")
	}

	selector := call.Argument(0).String()
	promise, resolve, reject := b.chromedp.vm.NewPromise()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				b.chromedp.responseCh <- func() {
					reject(NewErrorString(b.chromedp.vm, fmt.Sprintf("click panic: %v", r)))
				}
			}
		}()

		ctx, cancel := context.WithTimeout(b.ctx, b.timeout)
		defer cancel()

		err := chromedp.Run(ctx, chromedp.Click(selector))
		if err != nil {
			b.chromedp.responseCh <- func() {
				reject(NewError(b.chromedp.vm, err))
			}
			return
		}

		b.chromedp.responseCh <- func() {
			resolve(goja.Undefined())
		}
	}()

	return b.chromedp.vm.ToValue(promise)
}

// SendKeys types into an element
func (b *Browser) SendKeys(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		PanicThrowErrorString(b.chromedp.vm, "sendKeys requires selector and keys arguments")
	}

	selector := call.Argument(0).String()
	keys := call.Argument(1).String()
	promise, resolve, reject := b.chromedp.vm.NewPromise()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				b.chromedp.responseCh <- func() {
					reject(NewErrorString(b.chromedp.vm, fmt.Sprintf("sendKeys panic: %v", r)))
				}
			}
		}()

		ctx, cancel := context.WithTimeout(b.ctx, b.timeout)
		defer cancel()

		err := chromedp.Run(ctx, chromedp.SendKeys(selector, keys))
		if err != nil {
			b.chromedp.responseCh <- func() {
				reject(NewError(b.chromedp.vm, err))
			}
			return
		}

		b.chromedp.responseCh <- func() {
			resolve(goja.Undefined())
		}
	}()

	return b.chromedp.vm.ToValue(promise)
}

// EvaluateJS evaluates JavaScript in the browser context
func (b *Browser) EvaluateJS(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		PanicThrowErrorString(b.chromedp.vm, "evaluate requires a JavaScript expression")
	}

	jsCode := call.Argument(0).String()
	promise, resolve, reject := b.chromedp.vm.NewPromise()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				b.chromedp.responseCh <- func() {
					reject(NewErrorString(b.chromedp.vm, fmt.Sprintf("evaluate panic: %v", r)))
				}
			}
		}()

		ctx, cancel := context.WithTimeout(b.ctx, b.timeout)
		defer cancel()

		var result interface{}
		wrappedJS := jsCode
		if !strings.Contains(jsCode, "return") {
			wrappedJS = fmt.Sprintf("(function() { return %s; })()", jsCode)
		}

		err := chromedp.Run(ctx, chromedp.Evaluate(wrappedJS, &result))
		if err != nil {
			b.chromedp.responseCh <- func() {
				reject(NewError(b.chromedp.vm, err))
			}
			return
		}

		b.chromedp.responseCh <- func() {
			resolve(b.chromedp.vm.ToValue(result))
		}
	}()

	return b.chromedp.vm.ToValue(promise)
}

// InnerHTML gets the inner HTML of an element
func (b *Browser) InnerHTML(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		PanicThrowErrorString(b.chromedp.vm, "innerHTML requires a selector argument")
	}

	selector := call.Argument(0).String()
	promise, resolve, reject := b.chromedp.vm.NewPromise()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				b.chromedp.responseCh <- func() {
					reject(NewErrorString(b.chromedp.vm, fmt.Sprintf("innerHTML panic: %v", r)))
				}
			}
		}()

		ctx, cancel := context.WithTimeout(b.ctx, b.timeout)
		defer cancel()

		var html string
		err := chromedp.Run(ctx, chromedp.InnerHTML(selector, &html))
		if err != nil {
			b.chromedp.responseCh <- func() {
				reject(NewError(b.chromedp.vm, err))
			}
			return
		}

		b.chromedp.responseCh <- func() {
			resolve(b.chromedp.vm.ToValue(html))
		}
	}()

	return b.chromedp.vm.ToValue(promise)
}

// OuterHTML gets the outer HTML of an element
func (b *Browser) OuterHTML(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		PanicThrowErrorString(b.chromedp.vm, "outerHTML requires a selector argument")
	}

	selector := call.Argument(0).String()
	promise, resolve, reject := b.chromedp.vm.NewPromise()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				b.chromedp.responseCh <- func() {
					reject(NewErrorString(b.chromedp.vm, fmt.Sprintf("outerHTML panic: %v", r)))
				}
			}
		}()

		ctx, cancel := context.WithTimeout(b.ctx, b.timeout)
		defer cancel()

		var html string
		err := chromedp.Run(ctx, chromedp.OuterHTML(selector, &html))
		if err != nil {
			b.chromedp.responseCh <- func() {
				reject(NewError(b.chromedp.vm, err))
			}
			return
		}

		b.chromedp.responseCh <- func() {
			resolve(b.chromedp.vm.ToValue(html))
		}
	}()

	return b.chromedp.vm.ToValue(promise)
}

// Text gets the text content of an element
func (b *Browser) Text(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		PanicThrowErrorString(b.chromedp.vm, "text requires a selector argument")
	}

	selector := call.Argument(0).String()
	promise, resolve, reject := b.chromedp.vm.NewPromise()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				b.chromedp.responseCh <- func() {
					reject(NewErrorString(b.chromedp.vm, fmt.Sprintf("text panic: %v", r)))
				}
			}
		}()

		ctx, cancel := context.WithTimeout(b.ctx, b.timeout)
		defer cancel()

		var text string
		err := chromedp.Run(ctx, chromedp.Text(selector, &text))
		if err != nil {
			b.chromedp.responseCh <- func() {
				reject(NewError(b.chromedp.vm, err))
			}
			return
		}

		b.chromedp.responseCh <- func() {
			resolve(b.chromedp.vm.ToValue(text))
		}
	}()

	return b.chromedp.vm.ToValue(promise)
}

// Attribute gets an attribute value of an element
func (b *Browser) Attribute(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		PanicThrowErrorString(b.chromedp.vm, "attribute requires selector and attribute name arguments")
	}

	selector := call.Argument(0).String()
	attrName := call.Argument(1).String()
	promise, resolve, reject := b.chromedp.vm.NewPromise()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				b.chromedp.responseCh <- func() {
					reject(NewErrorString(b.chromedp.vm, fmt.Sprintf("attribute panic: %v", r)))
				}
			}
		}()

		ctx, cancel := context.WithTimeout(b.ctx, b.timeout)
		defer cancel()

		var attrValue string
		var ok bool
		err := chromedp.Run(ctx, chromedp.AttributeValue(selector, attrName, &attrValue, &ok))
		if err != nil {
			b.chromedp.responseCh <- func() {
				reject(NewError(b.chromedp.vm, err))
			}
			return
		}

		b.chromedp.responseCh <- func() {
			if ok {
				resolve(b.chromedp.vm.ToValue(attrValue))
			} else {
				resolve(goja.Null())
			}
		}
	}()

	return b.chromedp.vm.ToValue(promise)
}

// Screenshot captures a screenshot of a specific element
func (b *Browser) Screenshot(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		PanicThrowErrorString(b.chromedp.vm, "screenshot requires a selector argument")
	}

	selector := call.Argument(0).String()
	promise, resolve, reject := b.chromedp.vm.NewPromise()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				b.chromedp.responseCh <- func() {
					reject(NewErrorString(b.chromedp.vm, fmt.Sprintf("screenshot panic: %v", r)))
				}
			}
		}()

		ctx, cancel := context.WithTimeout(b.ctx, b.timeout)
		defer cancel()

		var buf []byte
		err := chromedp.Run(ctx, chromedp.Screenshot(selector, &buf))
		if err != nil {
			b.chromedp.responseCh <- func() {
				reject(NewError(b.chromedp.vm, err))
			}
			return
		}

		b.chromedp.responseCh <- func() {
			resolve(b.chromedp.vm.ToValue(buf))
		}
	}()

	return b.chromedp.vm.ToValue(promise)
}

// FullScreenshot captures a full page screenshot
func (b *Browser) FullScreenshot(call goja.FunctionCall) goja.Value {
	promise, resolve, reject := b.chromedp.vm.NewPromise()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				b.chromedp.responseCh <- func() {
					reject(NewErrorString(b.chromedp.vm, fmt.Sprintf("fullScreenshot panic: %v", r)))
				}
			}
		}()

		ctx, cancel := context.WithTimeout(b.ctx, b.timeout)
		defer cancel()

		var buf []byte
		err := chromedp.Run(ctx, chromedp.FullScreenshot(&buf, 100))
		if err != nil {
			b.chromedp.responseCh <- func() {
				reject(NewError(b.chromedp.vm, err))
			}
			return
		}

		b.chromedp.responseCh <- func() {
			resolve(b.chromedp.vm.ToValue(buf))
		}
	}()

	return b.chromedp.vm.ToValue(promise)
}

// Sleep waits for a duration
func (b *Browser) Sleep(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		PanicThrowErrorString(b.chromedp.vm, "sleep requires a duration in milliseconds")
	}

	duration := call.Argument(0).ToInteger()
	promise, resolve, reject := b.chromedp.vm.NewPromise()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				b.chromedp.responseCh <- func() {
					reject(NewErrorString(b.chromedp.vm, fmt.Sprintf("sleep panic: %v", r)))
				}
			}
		}()

		ctx, cancel := context.WithTimeout(b.ctx, b.timeout)
		defer cancel()

		err := chromedp.Run(ctx, chromedp.Sleep(time.Duration(duration)*time.Millisecond))
		if err != nil {
			b.chromedp.responseCh <- func() {
				reject(NewError(b.chromedp.vm, err))
			}
			return
		}

		b.chromedp.responseCh <- func() {
			resolve(goja.Undefined())
		}
	}()

	return b.chromedp.vm.ToValue(promise)
}

// Close closes the browser instance
func (b *Browser) Close(call goja.FunctionCall) goja.Value {
	promise, resolve, _ := b.chromedp.vm.NewPromise()

	go func() {
		b.chromedp.browserMu.Lock()
		delete(b.chromedp.browsers, b.id)
		b.chromedp.browserMu.Unlock()

		b.close()
		<-b.chromedp.chromeSem

		b.chromedp.responseCh <- func() {
			resolve(goja.Undefined())
		}
	}()

	return b.chromedp.vm.ToValue(promise)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// One-Shot Functions
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Scrape navigates to a URL and returns the HTML content
func (c *ChromeDP) Scrape(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		PanicThrowErrorString(c.vm, "scrape requires at least a URL argument")
	}

	url := call.Argument(0).String()

	// Parse options
	opts := chromeOptions{
		Timeout:  int(defaultBrowserTimeout.Seconds()),
		Headless: true,
	}

	if len(call.Arguments) > 1 {
		if optObj := call.Argument(1).ToObject(c.vm); optObj != nil {
			if v := optObj.Get("timeout"); gojaValueIsDefined(v) {
				opts.Timeout = int(v.ToInteger())
			}
			if v := optObj.Get("waitSelector"); gojaValueIsDefined(v) {
				opts.WaitSelector = v.String()
			}
			if v := optObj.Get("waitDuration"); gojaValueIsDefined(v) {
				opts.WaitDuration = int(v.ToInteger())
			}
			if v := optObj.Get("userAgent"); gojaValueIsDefined(v) {
				opts.UserAgent = v.String()
			}
			if v := optObj.Get("headless"); gojaValueIsDefined(v) {
				opts.Headless = v.ToBoolean()
			}
		}
	}

	promise, resolve, reject := c.vm.NewPromise()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				c.responseCh <- func() {
					reject(NewErrorString(c.vm, fmt.Sprintf("chromedp panic: %v", r)))
				}
			}
		}()

		c.chromeSem <- struct{}{}
		defer func() { <-c.chromeSem }()

		html, err := c.scrapeURL(url, opts)
		if err != nil {
			c.responseCh <- func() {
				reject(NewError(c.vm, err))
			}
			return
		}

		c.responseCh <- func() {
			resolve(c.vm.ToValue(html))
		}
	}()

	return c.vm.ToValue(promise)
}

// Screenshot captures a screenshot of a webpage
func (c *ChromeDP) Screenshot(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		PanicThrowErrorString(c.vm, "screenshot requires at least a URL argument")
	}

	url := call.Argument(0).String()

	opts := chromeOptions{
		Timeout:  int(defaultBrowserTimeout.Seconds()),
		Headless: true,
	}

	if len(call.Arguments) > 1 {
		if optObj := call.Argument(1).ToObject(c.vm); optObj != nil {
			if v := optObj.Get("timeout"); gojaValueIsDefined(v) {
				opts.Timeout = int(v.ToInteger())
			}
			if v := optObj.Get("waitSelector"); gojaValueIsDefined(v) {
				opts.WaitSelector = v.String()
			}
			if v := optObj.Get("waitDuration"); gojaValueIsDefined(v) {
				opts.WaitDuration = int(v.ToInteger())
			}
			if v := optObj.Get("userAgent"); gojaValueIsDefined(v) {
				opts.UserAgent = v.String()
			}
			if v := optObj.Get("headless"); gojaValueIsDefined(v) {
				opts.Headless = v.ToBoolean()
			}
		}
	}

	promise, resolve, reject := c.vm.NewPromise()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				c.responseCh <- func() {
					reject(NewErrorString(c.vm, fmt.Sprintf("chromedp panic: %v", r)))
				}
			}
		}()

		c.chromeSem <- struct{}{}
		defer func() { <-c.chromeSem }()

		screenshot, err := c.screenshotURL(url, opts)
		if err != nil {
			c.responseCh <- func() {
				reject(NewError(c.vm, err))
			}
			return
		}

		c.responseCh <- func() {
			resolve(c.vm.ToValue(screenshot))
		}
	}()

	return c.vm.ToValue(promise)
}

// Evaluate runs JavaScript code in the browser context and returns the result
func (c *ChromeDP) Evaluate(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		PanicThrowErrorString(c.vm, "evaluate requires a URL and a JavaScript expression")
	}

	url := call.Argument(0).String()
	jsCode := call.Argument(1).String()

	opts := chromeOptions{
		Timeout:  int(defaultBrowserTimeout.Seconds()),
		Headless: true,
	}

	if len(call.Arguments) > 2 {
		if optObj := call.Argument(2).ToObject(c.vm); optObj != nil {
			if v := optObj.Get("timeout"); gojaValueIsDefined(v) {
				opts.Timeout = int(v.ToInteger())
			}
			if v := optObj.Get("waitSelector"); gojaValueIsDefined(v) {
				opts.WaitSelector = v.String()
			}
			if v := optObj.Get("waitDuration"); gojaValueIsDefined(v) {
				opts.WaitDuration = int(v.ToInteger())
			}
			if v := optObj.Get("userAgent"); gojaValueIsDefined(v) {
				opts.UserAgent = v.String()
			}
			if v := optObj.Get("headless"); gojaValueIsDefined(v) {
				opts.Headless = v.ToBoolean()
			}
		}
	}

	promise, resolve, reject := c.vm.NewPromise()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				c.responseCh <- func() {
					reject(NewErrorString(c.vm, fmt.Sprintf("chromedp panic: %v", r)))
				}
			}
		}()

		c.chromeSem <- struct{}{}
		defer func() { <-c.chromeSem }()

		result, err := c.evaluateJS(url, jsCode, opts)
		if err != nil {
			c.responseCh <- func() {
				reject(NewError(c.vm, err))
			}
			return
		}

		c.responseCh <- func() {
			resolve(c.vm.ToValue(result))
		}
	}()

	return c.vm.ToValue(promise)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Helper functions
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (c *ChromeDP) scrapeURL(url string, opts chromeOptions) (string, error) {
	// Create context with options
	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", opts.Headless),
	)

	if opts.UserAgent != "" {
		allocOpts = append(allocOpts, chromedp.UserAgent(opts.UserAgent))
	}

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), allocOpts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, time.Duration(opts.Timeout)*time.Second)
	defer cancel()

	var html string
	var tasks []chromedp.Action

	// Navigate to URL
	tasks = append(tasks, chromedp.Navigate(url))

	// Wait for selector if provided
	if opts.WaitSelector != "" {
		tasks = append(tasks, chromedp.WaitVisible(opts.WaitSelector))
	}

	// Wait for additional duration if specified
	if opts.WaitDuration > 0 {
		tasks = append(tasks, chromedp.Sleep(time.Duration(opts.WaitDuration)*time.Millisecond))
	}

	// Get the outer HTML
	tasks = append(tasks, chromedp.OuterHTML("html", &html))

	err := chromedp.Run(ctx, tasks...)
	if err != nil {
		return "", fmt.Errorf("chromedp run failed: %w", err)
	}

	return html, nil
}

func (c *ChromeDP) screenshotURL(url string, opts chromeOptions) ([]byte, error) {
	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", opts.Headless),
	)

	if opts.UserAgent != "" {
		allocOpts = append(allocOpts, chromedp.UserAgent(opts.UserAgent))
	}

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), allocOpts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, time.Duration(opts.Timeout)*time.Second)
	defer cancel()

	var buf []byte
	var tasks []chromedp.Action

	tasks = append(tasks, chromedp.Navigate(url))

	if opts.WaitSelector != "" {
		tasks = append(tasks, chromedp.WaitVisible(opts.WaitSelector))
	}

	if opts.WaitDuration > 0 {
		tasks = append(tasks, chromedp.Sleep(time.Duration(opts.WaitDuration)*time.Millisecond))
	}

	tasks = append(tasks, chromedp.FullScreenshot(&buf, 100))

	err := chromedp.Run(ctx, tasks...)
	if err != nil {
		return nil, fmt.Errorf("chromedp screenshot failed: %w", err)
	}

	return buf, nil
}

func (c *ChromeDP) evaluateJS(url string, jsCode string, opts chromeOptions) (interface{}, error) {
	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", opts.Headless),
	)

	if opts.UserAgent != "" {
		allocOpts = append(allocOpts, chromedp.UserAgent(opts.UserAgent))
	}

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), allocOpts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, time.Duration(opts.Timeout)*time.Second)
	defer cancel()

	var result interface{}
	var tasks []chromedp.Action

	tasks = append(tasks, chromedp.Navigate(url))

	if opts.WaitSelector != "" {
		tasks = append(tasks, chromedp.WaitVisible(opts.WaitSelector))
	}

	if opts.WaitDuration > 0 {
		tasks = append(tasks, chromedp.Sleep(time.Duration(opts.WaitDuration)*time.Millisecond))
	}

	// Wrap the JS code to ensure it returns a value
	wrappedJS := jsCode
	if !strings.Contains(jsCode, "return") {
		wrappedJS = fmt.Sprintf("(function() { return %s; })()", jsCode)
	}

	tasks = append(tasks, chromedp.Evaluate(wrappedJS, &result))

	err := chromedp.Run(ctx, tasks...)
	if err != nil {
		return nil, fmt.Errorf("chromedp evaluate failed: %w", err)
	}

	return result, nil
}
