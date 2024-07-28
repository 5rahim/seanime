package extension_repo

import (
	"bytes"
	"github.com/dop251/goja"
	"github.com/goccy/go-json"
	"io"
	"net/http"
	"seanime/internal/util"
	"seanime/internal/util/comparison"
	"strings"
	"time"
)

func gojaBindFetch(vm *goja.Runtime) error {
	err := vm.Set("fetch", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(fetch(vm, call))
	})
	if err != nil {
		return err
	}

	return nil
}

func gojaBindFindBestMatchWithSorensenDice(vm *goja.Runtime) error {
	err := vm.Set("$findBestMatchWithSorensenDice", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return vm.ToValue("TypeError: Expected 2 arguments")
		}

		// Convert the first argument to string
		inputStr, ok := call.Argument(0).Export().(string)
		if !ok {
			return vm.ToValue("TypeError: First argument must be a string")
		}
		input := inputStr

		// Convert the second argument to an array of strings
		vals := call.Argument(1).ToObject(vm)

		// Check if the second argument is an array of strings
		length := vals.Get("length")
		if length == nil || goja.IsUndefined(length) {
			return vm.ToValue("TypeError: Second argument must be an array of strings")
		}

		var strVals []*string
		for _, key := range vals.Keys() {
			val := vals.Get(key)
			valStr := val.ToString()
			str := valStr.String()
			strVals = append(strVals, &str)
		}

		// Call the Go function
		result, ok := comparison.FindBestMatchWithSorensenDice(&input, strVals)
		if !ok {
			return vm.ToValue(nil) // No match found
		}

		// Create a JavaScript object to return
		jsResult := vm.NewObject()
		jsResult.Set("originalValue", result.OriginalValue)
		jsResult.Set("value", result.Value)
		jsResult.Set("rating", result.Rating)

		return jsResult
	})
	if err != nil {
		return err
	}

	return nil
}

func fetch(vm *goja.Runtime, call goja.FunctionCall) *goja.Promise {
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
		m := options.Get("method")
		if m != nil && !goja.IsUndefined(m) {
			method = strings.ToUpper(m.String())
		}

		headers := make(map[string]string)
		if h := options.Get("headers"); h != nil && !goja.IsUndefined(h) {
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
			Timeout: 10 * time.Second,
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

		var jsonResult interface{}
		if err := json.Unmarshal(bodyBytes, &jsonResult); err != nil {
			reject(vm.ToValue(err.Error()))
			return
		}

		responseObj := vm.NewObject()
		responseObj.Set("status", resp.StatusCode)
		responseObj.Set("statusText", resp.Status)
		responseObj.Set("ok", resp.StatusCode >= 200 && resp.StatusCode < 300)

		headersObj := vm.NewObject()
		for key, values := range resp.Header {
			headersObj.Set(key, values[0])
		}
		responseObj.Set("headers", headersObj)

		responseObj.Set("text", func(call goja.FunctionCall) goja.Value {
			return vm.ToValue(string(bodyBytes))
		})

		responseObj.Set("json", func(call goja.FunctionCall) goja.Value {
			return vm.ToValue(jsonResult)
		})

		resolve(responseObj)
	}()

	return promise
}
