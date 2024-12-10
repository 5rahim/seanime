package goja_bindings

import (
	"bytes"
	"fmt"
	"github.com/dop251/goja"
	"github.com/goccy/go-json"
	"io"
	"net/http"
	"seanime/internal/util"
	"strings"
	"time"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fetch
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func BindFetch(vm *goja.Runtime) error {
	err := vm.Set("fetch", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(gojaFetch(vm, call))
	})
	if err != nil {
		return err
	}

	return nil
}

func gojaFetch(vm *goja.Runtime, call goja.FunctionCall) (ret *goja.Promise) {
	defer func() {
		if r := recover(); r != nil {
			promise, _, reject := vm.NewPromise()
			reject(vm.ToValue(fmt.Sprintf("extension: Panic from fetch: %v", r)))
			ret = promise
		}
	}()

	if len(call.Arguments) < 1 {
		promise, _, reject := vm.NewPromise()
		reject(vm.ToValue("TypeError: fetch requires at least 1 argument"))
		return promise
	}

	// Convert the URL parameter to a string
	urlArg, ok := call.Argument(0).Export().(string)
	if !ok {
		promise, _, reject := vm.NewPromise()
		reject(vm.ToValue("TypeError: URL parameter must be a string"))
		return promise
	}

	// Check if the second parameter (options) is provided
	var options *goja.Object
	if len(call.Arguments) > 1 {
		optionsVal := call.Argument(1)
		options = optionsVal.ToObject(vm)
	} else {
		options = vm.NewObject() // Create an empty object if no options are provided
	}

	promise, resolve, reject := vm.NewPromise()

	go func() {
		method := "GET"
		if m := options.Get("method"); m != nil && gojaValueIsDefined(m) {
			method = strings.ToUpper(m.String())
		}

		headers := make(map[string]string)
		if h := options.Get("headers"); h != nil && gojaValueIsDefined(h) {
			headerObj := h.ToObject(vm)
			for _, key := range headerObj.Keys() {
				headers[key] = headerObj.Get(key).String()
			}
		}

		var body io.Reader
		if b := options.Get("body"); b != nil && !goja.IsUndefined(b) {
			body = bytes.NewBufferString(b.String())
		}

		req, err := http.NewRequest(method, urlArg, body)
		if err != nil {
			reject(vm.ToValue(err.Error()))
			return
		}

		for key, value := range headers {
			req.Header.Set(key, value)
		}

		client := &http.Client{
			Timeout: 60 * time.Second,
		}
		client.Transport = util.AddCloudFlareByPass(client.Transport)

		resp, err := client.Do(req)
		if err != nil {
			reject(vm.ToValue(err.Error()))
			return
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			reject(vm.ToValue(err.Error()))
			return
		}

		// Unmarshal the response body to an interface
		var jsonInterface interface{}
		canUnmarshal := true
		if err := json.Unmarshal(bodyBytes, &jsonInterface); err != nil {
			canUnmarshal = false
		}

		responseObj := vm.NewObject()
		responseObj.Set("status", resp.StatusCode)
		responseObj.Set("statusText", resp.Status)
		responseObj.Set("ok", resp.StatusCode >= 200 && resp.StatusCode < 300)
		responseObj.Set("url", resp.Request.URL.String())

		// Set the response headers
		respHeadersObj := vm.NewObject()
		for key, values := range resp.Header {
			respHeadersObj.Set(key, values[0])
		}
		responseObj.Set("headers", respHeadersObj)

		// Set the response body
		responseObj.Set("text", func(call goja.FunctionCall) goja.Value {
			return vm.ToValue(string(bodyBytes))
		})

		// Set the response JSON
		responseObj.Set("json", func(call goja.FunctionCall) goja.Value {
			if !canUnmarshal {
				return goja.Undefined()
			}
			return vm.ToValue(jsonInterface)
		})

		resolve(responseObj)
	}()

	return promise
}
