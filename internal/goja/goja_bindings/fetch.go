package goja_bindings

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"seanime/internal/util"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog/log"
)

const (
	maxConcurrentRequests = 50
	defaultTimeout        = 35 * time.Second
)

// fetchResponse encapsulates the response object creation
type fetchResponse struct {
	response *http.Response
	body     []byte
}

// toGojaObject converts the response to a Goja object with all necessary properties
// Must be called from the original VM's goroutine
func (fr *fetchResponse) toGojaObject(vm *goja.Runtime) *goja.Object {
	obj := vm.NewObject()
	_ = obj.Set("status", fr.response.StatusCode)
	_ = obj.Set("statusText", fr.response.Status)
	_ = obj.Set("method", fr.response.Request.Method)
	_ = obj.Set("ok", fr.response.StatusCode >= 200 && fr.response.StatusCode < 300)
	_ = obj.Set("url", fr.response.Request.URL.String())

	// Set headers
	headers := vm.NewObject()
	for key, values := range fr.response.Header {
		if len(values) > 0 {
			_ = headers.Set(key, values[0])
		}
	}
	_ = obj.Set("headers", headers)

	// Set body methods
	_ = obj.Set("text", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(string(fr.body))
	})

	// Set JSON method
	_ = obj.Set("json", func(call goja.FunctionCall) goja.Value {
		var jsonInterface interface{}
		if err := json.Unmarshal(fr.body, &jsonInterface); err != nil {
			return goja.Undefined()
		}
		return vm.ToValue(jsonInterface)
	})

	return obj
}

var (
	fetchSem = make(chan struct{}, maxConcurrentRequests)
	client   = &http.Client{
		Timeout:   defaultTimeout,
		Transport: util.AddCloudFlareByPass(http.DefaultTransport),
	}
)

func BindFetch(vm *goja.Runtime) error {
	return vm.Set("fetch", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(gojaFetch(vm, call))
	})
}

type fetchResult struct {
	response *fetchResponse
	err      error
}

func gojaFetch(vm *goja.Runtime, call goja.FunctionCall) *goja.Promise {
	promise, resolve, reject := vm.NewPromise()

	// Input validation
	if len(call.Arguments) < 1 {
		reject(vm.ToValue("TypeError: fetch requires at least 1 argument"))
		return promise
	}

	urlArg, ok := call.Argument(0).Export().(string)
	if !ok {
		reject(vm.ToValue("TypeError: URL parameter must be a string"))
		return promise
	}

	// Parse options
	options := parseOptions(vm, call)

	// channel to receive the result
	resultCh := make(chan fetchResult, 1)

	go func() {
		var result fetchResult
		defer func() {
			if r := recover(); r != nil {
				result.err = fmt.Errorf("JS VM: Panic from fetch: %v", r)
			}
			resultCh <- result
		}()

		// Acquire semaphore
		fetchSem <- struct{}{}
		defer func() { <-fetchSem }()

		// Create request
		req, err := createRequest(urlArg, options)
		if err != nil {
			result.err = err
			return
		}

		// Execute request
		resp, body, err := executeRequest(req)
		if err != nil {
			result.err = err
			return
		}
		defer resp.Body.Close()

		// Create response object
		result.response = &fetchResponse{
			response: resp,
			body:     body,
		}
	}()

	// Handle the result in the original goroutine
	result := <-resultCh
	if result.err != nil {
		reject(vm.ToValue(result.err.Error()))
		return promise
	}
	resolve(result.response.toGojaObject(vm))

	return promise
}

func parseOptions(vm *goja.Runtime, call goja.FunctionCall) *goja.Object {
	if len(call.Arguments) > 1 {
		return call.Argument(1).ToObject(vm)
	}
	return vm.NewObject()
}

func createRequest(url string, options *goja.Object) (*http.Request, error) {
	method := "GET"
	if m := options.Get("method"); m != nil && gojaValueIsDefined(m) {
		method = strings.ToUpper(m.String())
	}

	var body io.Reader
	if b := options.Get("body"); b != nil && !goja.IsUndefined(b) {
		body = bytes.NewBufferString(b.String())
	}

	req, err := http.NewRequestWithContext(context.Background(), method, url, body)
	if err != nil {
		return nil, err
	}

	// Set headers
	if h := options.Get("headers"); h != nil && gojaValueIsDefined(h) {
		headerObj := h.ToObject(nil)
		for _, key := range headerObj.Keys() {
			req.Header.Set(key, headerObj.Get(key).String())
		}
	}

	log.Trace().Str("url", url).Str("method", method).Msgf("extension: Fetching using JS VM")
	return req, nil
}

func executeRequest(req *http.Request) (*http.Response, []byte, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("request failed: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("reading response body failed: %w", err)
	}

	return resp, body, nil
}
