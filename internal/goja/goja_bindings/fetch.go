package goja_bindings

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"seanime/internal/util"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/rs/zerolog/log"
)

const (
	maxConcurrentRequests = 50
	defaultTimeout        = 35 * time.Second
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fetch
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var (
	clientWithCloudFlareBypass = &http.Client{
		Transport: util.AddCloudFlareByPass(http.DefaultTransport),
	}
)

type Fetch struct {
	vm           *goja.Runtime
	fetchSem     chan struct{}
	vmResponseCh chan func()
}

func NewFetch(vm *goja.Runtime) *Fetch {
	return &Fetch{
		vm:           vm,
		fetchSem:     make(chan struct{}, maxConcurrentRequests),
		vmResponseCh: make(chan func(), maxConcurrentRequests),
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
	close(f.vmResponseCh)
}

type fetchOptions struct {
	Method             string
	Body               goja.Value
	Headers            map[string]string
	Timeout            int // seconds
	NoCloudFlareBypass bool
}

type fetchResult struct {
	body     []byte
	request  *http.Request
	response *http.Response
	json     interface{}
}

// BindFetch binds the fetch function to the VM
func BindFetch(vm *goja.Runtime) *Fetch {
	// Create a new Fetch instance
	f := NewFetch(vm)
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

func (f *Fetch) Fetch(call goja.FunctionCall) goja.Value {
	defer func() {
		if r := recover(); r != nil {
			log.Warn().Msgf("extension: fetch panic: %v", r)
		}
	}()

	promise, resolve, reject := f.vm.NewPromise()

	// Input validation
	if len(call.Arguments) == 0 {
		//_ = reject(NewErrorString(f.vm, "TypeError: fetch requires at least 1 argument"))
		//return f.vm.ToValue(promise)
		PanicThrowTypeError(f.vm, "fetch requires at least 1 argument")
	}

	url, ok := call.Argument(0).Export().(string)
	if !ok {
		//_ = reject(NewErrorString(f.vm, "TypeError: URL parameter must be a string"))
		//return f.vm.ToValue(promise)
		PanicThrowTypeError(f.vm, "URL parameter must be a string")
	}

	// Parse options
	options := fetchOptions{
		Method:             "GET",
		Timeout:            int(defaultTimeout.Seconds()),
		NoCloudFlareBypass: false,
	}

	var reqBody io.Reader
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
		}
	}

	//gojaValue := f.vm.ToValue(options.Body)
	//reqBody = bytes.NewBufferString(gojaValue.String())
	if options.Body != nil && !goja.IsUndefined(options.Body) {
		switch v := options.Body.Export().(type) {
		case string:
			reqBody = strings.NewReader(v)
		case io.Reader:
			reqBody = v
		case []byte:
			reqBody = bytes.NewReader(v)
		case *goja.ArrayBuffer:
			reqBody = bytes.NewReader(v.Bytes())
		case goja.ArrayBuffer:
			reqBody = bytes.NewReader(v.Bytes())
		case *formData:
			body, mp := v.GetBuffer()
			reqBody = body
			reqContentType = mp.FormDataContentType()
		case map[string]interface{}:
			jsonBody, err := json.Marshal(v)
			if err != nil {
				_ = reject(NewError(f.vm, err))
				return f.vm.ToValue(promise)
			}
			reqBody = bytes.NewReader(jsonBody)
			reqContentType = "application/json"
		default:
			reqBody = bytes.NewBufferString(options.Body.String())
		}
	}

	go func() {
		// Acquire semaphore
		f.fetchSem <- struct{}{}
		defer func() { <-f.fetchSem }()

		log.Trace().Str("url", url).Str("method", options.Method).Msgf("extension: Network request")

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(options.Timeout)*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, options.Method, url, reqBody)
		if err != nil {
			f.vmResponseCh <- func() {
				_ = reject(NewError(f.vm, err))
			}
			return
		}

		for k, v := range options.Headers {
			req.Header.Set(k, v)
		}

		if reqContentType != "" {
			req.Header.Set("Content-Type", reqContentType)
		}

		var result fetchResult

		var client *http.Client
		if options.NoCloudFlareBypass {
			client = http.DefaultClient
		} else {
			client = clientWithCloudFlareBypass
		}

		resp, err := client.Do(req)
		if err != nil {
			f.vmResponseCh <- func() {
				_ = reject(NewError(f.vm, err))
			}
			return
		}
		defer resp.Body.Close()

		rawBody, err := io.ReadAll(resp.Body)
		if err != nil {
			f.vmResponseCh <- func() {
				_ = reject(NewError(f.vm, err))
			}
			return
		}

		result.body = rawBody
		result.response = resp
		result.request = req

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
	_ = obj.Set("method", f.response.Request.Method)
	_ = obj.Set("rawHeaders", f.response.Header)
	_ = obj.Set("ok", f.response.StatusCode >= 200 && f.response.StatusCode < 300)
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
	_ = obj.Set("redirected", f.request.URL != f.response.Request.URL)
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

//------------------------------------------------------------------------

//// fetchResponse encapsulates the response object creation
//type fetchResponse struct {
//	response *http.Response
//	body     []byte
//}
//
//// toGojaObject converts the response to a Goja object with all necessary properties
//// Must be called from the original VM's goroutine
//func (fr *fetchResponse) toGojaObject(vm *goja.Runtime) *goja.Object {
//	obj := vm.NewObject()
//	_ = obj.Set("status", fr.response.StatusCode)
//	_ = obj.Set("statusText", fr.response.Status)
//	_ = obj.Set("method", fr.response.Request.Method)
//	_ = obj.Set("rawHeaders", fr.response.Header)
//	_ = obj.Set("ok", fr.response.StatusCode >= 200 && fr.response.StatusCode < 300)
//	_ = obj.Set("url", fr.response.Request.URL.String())
//
//	// Set headers
//	headers := vm.NewObject()
//	for key, values := range fr.response.Header {
//		if len(values) > 0 {
//			_ = headers.Set(key, values[0])
//		}
//	}
//	_ = obj.Set("headers", headers)
//
//	// Set body methods
//	_ = obj.Set("text", func() string {
//		return string(fr.body)
//	})
//
//	// Set JSON method
//	_ = obj.Set("json", func(call goja.FunctionCall) goja.Value {
//		var jsonInterface interface{}
//		if err := json.Unmarshal(fr.body, &jsonInterface); err != nil {
//			return goja.Undefined()
//		}
//		return vm.ToValue(jsonInterface)
//	})
//
//	return obj
//}
//
//var client = &http.Client{
//	Timeout:   defaultTimeout,
//	Transport: util.AddCloudFlareByPass(http.DefaultTransport),
//}
//
//type vmFetchState struct {
//	fetchSem     chan struct{}
//	vmResponseCh chan func()
//}
//
//func BindFetch(vm *goja.Runtime) error {
//	state := &vmFetchState{
//		fetchSem:     make(chan struct{}, maxConcurrentRequests),
//		vmResponseCh: make(chan func(), maxConcurrentRequests),
//	}
//
//	// Start a goroutine to handle VM responses for this specific VM
//	go func() {
//		for fn := range state.vmResponseCh {
//			fn()
//		}
//	}()
//
//	err := vm.Set("fetch", func(call goja.FunctionCall) goja.Value {
//		return vm.ToValue(gojaFetch(vm, call, state))
//	})
//	if err != nil {
//		close(state.vmResponseCh)
//		return err
//	}
//
//	return nil
//}
//
//type fetchResult struct {
//	response *fetchResponse
//	err      error
//}
//
//func gojaFetch(vm *goja.Runtime, call goja.FunctionCall, state *vmFetchState) *goja.Promise {
//	promise, resolve, reject := vm.NewPromise()
//
//	// Input validation
//	if len(call.Arguments) < 1 {
//		_ = reject(vm.ToValue("TypeError: fetch requires at least 1 argument"))
//		return promise
//	}
//
//	urlArg, ok := call.Argument(0).Export().(string)
//	if !ok {
//		_ = reject(vm.ToValue("TypeError: URL parameter must be a string"))
//		return promise
//	}
//
//	// Parse options
//	options := parseOptions(vm, call)
//
//	// Execute request in a separate goroutine to not block the VM
//	go func() {
//		var result fetchResult
//
//		// Acquire semaphore
//		state.fetchSem <- struct{}{}
//		defer func() { <-state.fetchSem }()
//
//		// Create request
//		req, err := createRequest(urlArg, options)
//		if err != nil {
//			state.vmResponseCh <- func() {
//				_ = reject(vm.ToValue(err.Error()))
//			}
//			return
//		}
//
//		// Execute request
//		resp, body, err := executeRequest(req)
//		if err != nil {
//			state.vmResponseCh <- func() {
//				_ = reject(vm.ToValue(err.Error()))
//			}
//			return
//		}
//		defer resp.Body.Close()
//
//		// Create response object
//		result.response = &fetchResponse{
//			response: resp,
//			body:     body,
//		}
//
//		// Schedule the resolution through the VM response channel
//		state.vmResponseCh <- func() {
//			_ = resolve(result.response.toGojaObject(vm))
//		}
//	}()
//
//	return promise
//}
//
//func parseOptions(vm *goja.Runtime, call goja.FunctionCall) *goja.Object {
//	if len(call.Arguments) > 1 {
//		return call.Argument(1).ToObject(vm)
//	}
//	return vm.NewObject()
//}
//
//func createRequest(url string, options *goja.Object) (*http.Request, error) {
//	method := "GET"
//	if m := options.Get("method"); m != nil && gojaValueIsDefined(m) {
//		method = strings.ToUpper(m.String())
//	}
//
//	var body io.Reader
//	if b := options.Get("body"); b != nil && !goja.IsUndefined(b) {
//		body = bytes.NewBufferString(b.String())
//	}
//
//	req, err := http.NewRequestWithContext(context.Background(), method, url, body)
//	if err != nil {
//		return nil, err
//	}
//
//	// Set headers
//	if h := options.Get("headers"); h != nil && gojaValueIsDefined(h) {
//		headerObj := h.ToObject(nil)
//		for _, key := range headerObj.Keys() {
//			req.Header.Set(key, headerObj.Get(key).String())
//		}
//	}
//
//	log.Trace().Str("url", url).Str("method", method).Msgf("extension: Fetching using JS VM")
//	return req, nil
//}
//
//func executeRequest(req *http.Request) (*http.Response, []byte, error) {
//	resp, err := client.Do(req)
//	if err != nil {
//		return nil, nil, fmt.Errorf("request failed: %w", err)
//	}
//
//	body, err := io.ReadAll(resp.Body)
//	if err != nil {
//		return nil, nil, fmt.Errorf("reading response body failed: %w", err)
//	}
//
//	return resp, body, nil
//}
