package goja_bindings

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"seanime/internal/util"
	"strings"
	"sync/atomic"
	"time"

	"github.com/dop251/goja"
	"github.com/imroc/req/v3"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

const (
	maxConcurrentRequests = 50
	defaultTimeout        = 35 * time.Second
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fetch
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var (
	clientWithCloudFlareBypass = req.C().
					SetUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36").
					SetTimeout(defaultTimeout).
					EnableInsecureSkipVerify().
					ImpersonateChrome()

	clientWithoutBypass = req.C().
				SetTimeout(defaultTimeout)
)

type Fetch struct {
	vm             *goja.Runtime
	fetchSem       chan struct{}
	vmResponseCh   chan func()
	closed         atomic.Bool
	allowedDomains []string // empty = allow all domains
	rules          []accessRule
	anilistToken   string
}

func (f *Fetch) SetAnilistToken(token string) {
	f.anilistToken = token
}

// accessRule represents a pre-parsed allowed domain pattern
type accessRule struct {
	scheme     string
	host       string
	path       string
	wildcard   bool // matches *
	subdomain  bool // matches *.example.com
	pathPrefix bool // matches /path/* instead of exact /path
	port       string
}

var whitelistedDomains = []string{
	"api.github.com",
	"raw.githubusercontent.com",
	"shikimori.one",
	"anilist.co",
	"graphql.anilist.co",
	"myanimelist.net",
	"*.myanimelist.net",
	"seanime.app",
	"trakt.tv",
	"*.trakt.tv",
	"kitsu.io",
	"api.simkl.com",
	"simkl.com",
	"*.gstatic.com",
	"*.googleapis.com",
}

func NewFetch(vm *goja.Runtime, allowedDomains []string) *Fetch {
	f := &Fetch{
		vm:             vm,
		fetchSem:       make(chan struct{}, maxConcurrentRequests),
		vmResponseCh:   make(chan func(), maxConcurrentRequests),
		allowedDomains: allowedDomains,
	}
	f.allowedDomains = lo.Uniq(append(f.allowedDomains, whitelistedDomains...))
	f.compileRules()

	return f
}

// CompileRules parses allowedDomains into efficient accessRules.
// Call this once when initializing Fetch or updating configuration.
func (f *Fetch) compileRules() {
	f.rules = make([]accessRule, 0, len(f.allowedDomains))

	for _, pattern := range f.allowedDomains {
		if pattern == "*" {
			f.rules = append(f.rules, accessRule{wildcard: true})
			continue
		}

		// Ensure we can parse it by adding a scheme if missing
		parseStr := pattern
		if !strings.Contains(pattern, "://") {
			parseStr = "https://" + pattern
		}

		u, err := url.Parse(parseStr)
		if err != nil {
			continue
		}

		rule := accessRule{
			host: u.Hostname(),
			path: u.Path,
			port: u.Port(),
		}

		// If the original pattern had a scheme, enforce it
		if strings.Contains(pattern, "://") {
			rule.scheme = u.Scheme
		}

		// Check for subdomain wildcard (*.example.com)
		// We check the original string because url.Parse might mess up the asterisk
		if strings.HasPrefix(pattern, "*.") || strings.HasPrefix(u.Hostname(), "*.") {
			rule.subdomain = true
			rule.host = strings.TrimPrefix(rule.host, "*.")
		}

		// Trailing slash indicates a prefix match for paths
		if strings.HasSuffix(pattern, "/") {
			rule.pathPrefix = true
		}

		f.rules = append(f.rules, rule)
	}
}

func (f *Fetch) ResponseChannel() <-chan func() {
	return f.vmResponseCh
}

func (f *Fetch) Close() {
	defer func() {
		if r := recover(); r != nil {
		}
	}()
	f.closed.Store(true)
	close(f.vmResponseCh)
}

type fetchOptions struct {
	Method             string
	Body               goja.Value
	Headers            map[string]string
	Timeout            int // seconds
	NoCloudFlareBypass bool
	Signal             *goja.Object // AbortSignal
}

type fetchResult struct {
	body     []byte
	request  *req.Request
	response *req.Response
	json     interface{}
}

// BindFetch binds the fetch function to the VM
func BindFetch(vm *goja.Runtime, allowedDomains ...[]string) *Fetch {

	ad := []string{"*"}
	if len(allowedDomains) > 0 {
		ad = allowedDomains[0]
	}

	// Create a new Fetch instance
	f := NewFetch(vm, ad)
	_ = vm.Set("fetch", f.Fetch)

	go func() {
		for fn := range f.ResponseChannel() {
			defer func() {
				if r := recover(); r != nil {
					log.Warn().Msgf("extension: response channel panic: %v", r)
				}
			}()
			fn()
		}
	}()

	return f
}

func (f *Fetch) isURLAllowed(urlStr string) bool {
	if len(f.rules) == 0 {
		return false
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	targetHost := u.Hostname()
	targetPort := u.Port()

	for _, rule := range f.rules {
		if rule.wildcard {
			return true
		}

		// 1. Scheme Check (only if rule specifies one)
		if rule.scheme != "" && rule.scheme != u.Scheme {
			continue
		}

		// 2. Host Check
		hostMatches := false
		if rule.subdomain {
			// Matches domain.com or sub.domain.com
			if targetHost == rule.host || strings.HasSuffix(targetHost, "."+rule.host) {
				hostMatches = true
			}
		} else {
			// Exact match or localhost special handling
			if targetHost == rule.host {
				hostMatches = true
			} else if rule.host == "localhost" && (targetHost == "127.0.0.1" || targetHost == "::1") {
				// simple localhost alias handling
				hostMatches = true
			}
		}

		if !hostMatches {
			continue
		}

		// 3. Port Check (only if rule specifies one)
		if rule.port != "" && rule.port != targetPort {
			continue
		}

		// 4. Path Check
		if rule.path != "" && rule.path != "/" {
			if rule.pathPrefix {
				// /api/ allows /api/v1
				if !strings.HasPrefix(u.Path, rule.path) {
					continue
				}
			} else {
				// /api allows /api but not /api/v1
				// handle trailing slash ambiguity for exact matches
				cleanRule := strings.TrimSuffix(rule.path, "/")
				cleanTarget := strings.TrimSuffix(u.Path, "/")
				if cleanRule != cleanTarget {
					continue
				}
			}
		}

		return true
	}

	return false
}

func (f *Fetch) Fetch(call goja.FunctionCall) goja.Value {
	defer func() {
		if r := recover(); r != nil {
			log.Warn().Msgf("extension: fetch panic: %v", r)
		}
	}()

	promise, resolve, reject := f.vm.NewPromise()

	// Input validation
	if len(call.Arguments) == 0 {
		PanicThrowTypeError(f.vm, "fetch requires at least 1 argument")
	}

	url, ok := call.Argument(0).Export().(string)
	if !ok {
		PanicThrowTypeError(f.vm, "URL parameter must be a string")
	}

	// Parse options
	options := fetchOptions{
		Method:             "GET",
		Timeout:            int(defaultTimeout.Seconds()),
		NoCloudFlareBypass: false,
	}

	var reqBody interface{}
	var reqContentType string

	if len(call.Arguments) > 1 {
		rawOpts := call.Argument(1).ToObject(f.vm)
		if rawOpts != nil && !goja.IsUndefined(rawOpts) {

			if o := rawOpts.Get("method"); o != nil && !goja.IsUndefined(o) {
				if v, ok := o.Export().(string); ok {
					options.Method = strings.ToUpper(v)
				}
			}
			if o := rawOpts.Get("timeout"); o != nil && !goja.IsUndefined(o) {
				if v, ok := o.Export().(int); ok {
					options.Timeout = v
				}
			}
			if o := rawOpts.Get("headers"); o != nil && !goja.IsUndefined(o) {
				if v, ok := o.Export().(map[string]interface{}); ok {
					for k, interf := range v {
						if str, ok := interf.(string); ok {
							if options.Headers == nil {
								options.Headers = make(map[string]string)
							}
							options.Headers[k] = str
						}
					}
				}
			}

			options.Body = rawOpts.Get("body")

			if o := rawOpts.Get("noCloudflareBypass"); o != nil && !goja.IsUndefined(o) {
				if v, ok := o.Export().(bool); ok {
					options.NoCloudFlareBypass = v
				}
			}

			if o := rawOpts.Get("signal"); o != nil && !goja.IsUndefined(o) {
				if signalObj := o.ToObject(f.vm); signalObj != nil {
					options.Signal = signalObj
				}
			}
		}
	}

	// Check if URL is allowed based on domain restrictions
	if !f.isURLAllowed(url) {
		reject(NewError(f.vm, fmt.Errorf("network access denied: URL '%s' does not match any allowed domain patterns", url)))
		return f.vm.ToValue(promise)
	}

	if options.Body != nil && !goja.IsUndefined(options.Body) {
		switch v := options.Body.Export().(type) {
		case string:
			reqBody = v
		case io.Reader:
			reqBody = v
		case []byte:
			reqBody = v
		case *goja.ArrayBuffer:
			reqBody = v.Bytes()
		case goja.ArrayBuffer:
			reqBody = v.Bytes()
		case *formData:
			body, mp := v.GetBuffer()
			reqBody = body
			reqContentType = mp.FormDataContentType()
		case map[string]interface{}:
			reqBody = v
			reqContentType = "application/json"
		default:
			reqBody = options.Body.String()
		}
	}

	go func() {
		defer util.HandlePanicInModuleThen("goja/goja_bindings/Fetch", func() {})
		if f.closed.Load() {
			return
		}
		// Acquire semaphore
		f.fetchSem <- struct{}{}
		defer func() { <-f.fetchSem }()

		// Check if signal is already aborted
		if options.Signal != nil {
			abortedValue := options.Signal.Get("aborted")
			if !goja.IsUndefined(abortedValue) && abortedValue.ToBoolean() {
				f.vmResponseCh <- func() {
					reason := options.Signal.Get("reason")
					if !goja.IsUndefined(reason) && !goja.IsNull(reason) {
						_ = reject(f.vm.ToValue(reason))
					} else {
						_ = reject(NewError(f.vm, fmt.Errorf("AbortError: The operation was aborted")))
					}
				}
				return
			}
		}

		log.Trace().Str("url", url).Str("method", options.Method).Msgf("extension: Network request")

		var client *req.Client
		if options.NoCloudFlareBypass {
			client = clientWithoutBypass
		} else {
			client = clientWithCloudFlareBypass
		}

		// Create request with timeout
		reqClient := client.Clone().SetTimeout(time.Duration(options.Timeout) * time.Second)

		request := reqClient.R()

		// Set headers
		for k, v := range options.Headers {
			request.SetHeader(k, v)
		}

		if reqContentType != "" {
			request.SetContentType(reqContentType)
		}

		// Set body if present
		if reqBody != nil {
			request.SetBody(reqBody)
		}

		// Set context from AbortSignal if provided
		if options.Signal != nil {
			// Extract the context from the AbortSignal
			getContextFunc := options.Signal.Get("_getContext")
			if callable, ok := goja.AssertFunction(getContextFunc); ok {
				ctxVal, err := callable(goja.Undefined())
				if err == nil {
					if ctx, ok := ctxVal.Export().(context.Context); ok {
						request.SetContext(ctx)
					}
				}
			}
		}

		var result fetchResult
		var resp *req.Response
		var err error

		switch options.Method {
		case "GET":
			resp, err = request.Get(url)
		case "POST":
			resp, err = request.Post(url)
		case "PUT":
			resp, err = request.Put(url)
		case "DELETE":
			resp, err = request.Delete(url)
		case "PATCH":
			resp, err = request.Patch(url)
		case "HEAD":
			resp, err = request.Head(url)
		case "OPTIONS":
			resp, err = request.Options(url)
		default:
			resp, err = request.Send(options.Method, url)
		}

		if err != nil {
			f.vmResponseCh <- func() {
				_ = reject(NewError(f.vm, err))
			}
			return
		}

		rawBody := resp.Bytes()
		result.body = rawBody
		result.response = resp
		result.request = request

		if len(rawBody) > 0 {
			var data interface{}
			if err := json.Unmarshal(rawBody, &data); err != nil {
				result.json = nil
			} else {
				result.json = data
			}
		}

		f.vmResponseCh <- func() {
			_ = resolve(result.toGojaObject(f.vm))
			return
		}
	}()

	return f.vm.ToValue(promise)
}

func (f *fetchResult) toGojaObject(vm *goja.Runtime) *goja.Object {
	obj := vm.NewObject()
	_ = obj.Set("status", f.response.StatusCode)
	_ = obj.Set("statusText", f.response.Status)
	_ = obj.Set("method", f.request.Method)
	_ = obj.Set("rawHeaders", f.response.Header)
	_ = obj.Set("ok", f.response.IsSuccessState())
	_ = obj.Set("url", f.response.Request.URL.String())
	_ = obj.Set("body", f.body)

	headers := make(map[string]string)
	for k, v := range f.response.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}
	_ = obj.Set("headers", headers)

	cookies := make(map[string]string)
	for _, cookie := range f.response.Cookies() {
		cookies[cookie.Name] = cookie.Value
	}
	_ = obj.Set("cookies", cookies)
	_ = obj.Set("redirected", f.response.Request.URL != f.response.Request.URL) // req handles redirects automatically
	_ = obj.Set("contentType", f.response.Header.Get("Content-Type"))
	_ = obj.Set("contentLength", f.response.ContentLength)

	_ = obj.Set("text", func() string {
		return string(f.body)
	})

	_ = obj.Set("json", func(call goja.FunctionCall) (ret goja.Value) {
		return vm.ToValue(f.json)
	})

	return obj
}
