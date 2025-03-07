package goja_bindings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"seanime/internal/util"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/rs/zerolog/log"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fetch
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

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
	_ = obj.Set("rawHeaders", fr.response.Header)
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
	_ = obj.Set("text", func() string {
		return string(fr.body)
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

var client = &http.Client{
	Timeout:   defaultTimeout,
	Transport: util.AddCloudFlareByPass(http.DefaultTransport),
}

type vmFetchState struct {
	fetchSem     chan struct{}
	vmResponseCh chan func()
}

func BindFetch(vm *goja.Runtime) error {
	state := &vmFetchState{
		fetchSem:     make(chan struct{}, maxConcurrentRequests),
		vmResponseCh: make(chan func(), maxConcurrentRequests),
	}

	// Start a goroutine to handle VM responses for this specific VM
	go func() {
		for fn := range state.vmResponseCh {
			fn()
		}
	}()

	err := vm.Set("fetch", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(gojaFetch(vm, call, state))
	})
	if err != nil {
		close(state.vmResponseCh)
		return err
	}

	return nil
}

type fetchResult struct {
	response *fetchResponse
	err      error
}

func gojaFetch(vm *goja.Runtime, call goja.FunctionCall, state *vmFetchState) *goja.Promise {
	promise, resolve, reject := vm.NewPromise()

	// Input validation
	if len(call.Arguments) < 1 {
		_ = reject(vm.ToValue("TypeError: fetch requires at least 1 argument"))
		return promise
	}

	urlArg, ok := call.Argument(0).Export().(string)
	if !ok {
		_ = reject(vm.ToValue("TypeError: URL parameter must be a string"))
		return promise
	}

	// Parse options
	options := parseOptions(vm, call)

	// Execute request in a separate goroutine to not block the VM
	go func() {
		var result fetchResult

		// Acquire semaphore
		state.fetchSem <- struct{}{}
		defer func() { <-state.fetchSem }()

		// Create request
		req, err := createRequest(urlArg, options)
		if err != nil {
			state.vmResponseCh <- func() {
				_ = reject(vm.ToValue(err.Error()))
			}
			return
		}

		// Execute request
		resp, body, err := executeRequest(req)
		if err != nil {
			state.vmResponseCh <- func() {
				_ = reject(vm.ToValue(err.Error()))
			}
			return
		}
		defer resp.Body.Close()

		// Create response object
		result.response = &fetchResponse{
			response: resp,
			body:     body,
		}

		// Schedule the resolution through the VM response channel
		state.vmResponseCh <- func() {
			_ = resolve(result.response.toGojaObject(vm))
		}
	}()

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
