package extension_repo

import (
	"bytes"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/dop251/goja"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
	"io"
	"mime/multipart"
	"net/http"
	"reflect"
	"seanime/internal/util"
	"strconv"
	"strings"
	"time"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fetch
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func gojaBindFetch(vm *goja.Runtime) error {
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

		// Unmarshal the response body to an interface
		var jsonInterface interface{}
		if err := json.Unmarshal(bodyBytes, &jsonInterface); err != nil {
			reject(vm.ToValue(err.Error()))
			return
		}

		responseObj := vm.NewObject()
		responseObj.Set("status", resp.StatusCode)
		responseObj.Set("statusText", resp.Status)
		responseObj.Set("ok", resp.StatusCode >= 200 && resp.StatusCode < 300)

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
			return vm.ToValue(jsonInterface)
		})

		resolve(responseObj)
	}()

	return promise
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Console
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type gojaConsole struct {
	logger *zerolog.Logger
}

func gojaBindConsole(vm *goja.Runtime, logger *zerolog.Logger) error {
	console := &gojaConsole{
		logger: logger,
	}
	consoleObj := vm.NewObject()
	consoleObj.Set("log", console.Log("log"))
	consoleObj.Set("error", console.Log("error"))
	consoleObj.Set("warn", console.Log("warn"))
	consoleObj.Set("info", console.Log("info"))
	consoleObj.Set("debug", console.Log("debug"))

	vm.Set("console", consoleObj)

	return nil
}

// Log method for console.log
func (c *gojaConsole) Log(t string) (ret func(c goja.FunctionCall) goja.Value) {
	defer func() {
		if r := recover(); r != nil {
			c.logger.Error().Msgf("extension: Panic from console.log: %v", r)
			ret = func(call goja.FunctionCall) goja.Value {
				return goja.Undefined()
			}
		}
	}()

	return func(call goja.FunctionCall) goja.Value {
		var ret []string
		for _, arg := range call.Arguments {
			if arg == nil || arg.Export() == nil || arg.ExportType() == nil {
				ret = append(ret, "undefined")
				continue
			}
			if arg.ExportType().Kind() == reflect.Struct || arg.ExportType().Kind() == reflect.Map || arg.ExportType().Kind() == reflect.Slice {
				ret = append(ret, strings.ReplaceAll(spew.Sdump(arg.Export()), "\n", ""))
			} else {
				ret = append(ret, fmt.Sprintf("%v", arg.Export()))
			}
		}

		switch t {
		case "log", "warn", "info", "debug":
			c.logger.Debug().Msgf("extension: [console.%s] %s", t, strings.Join(ret, " "))
		case "error":
			c.logger.Error().Msgf("extension: [console.error] %s", strings.Join(ret, " "))
		}
		return goja.Undefined()
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// GojaFormData
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func gojaBindFormData(vm *goja.Runtime) error {
	err := vm.Set("FormData", func(call goja.ConstructorCall) *goja.Object {
		fd := NewFormData(vm)
		obj := call.This
		obj.Set("append", fd.Append)
		obj.Set("delete", fd.Delete)
		obj.Set("entries", fd.Entries)
		obj.Set("get", fd.Get)
		obj.Set("getAll", fd.GetAll)
		obj.Set("has", fd.Has)
		obj.Set("keys", fd.Keys)
		obj.Set("set", fd.Set)
		obj.Set("values", fd.Values)
		obj.Set("getContentType", fd.GetContentType)
		obj.Set("getBuffer", fd.GetBuffer)
		return obj
	})
	if err != nil {
		return err
	}
	return nil
}

type GojaFormData struct {
	runtime    *goja.Runtime
	buf        *bytes.Buffer
	writer     *multipart.Writer
	fieldNames map[string]struct{}
	values     map[string][]string
	closed     bool
}

func NewFormData(runtime *goja.Runtime) *GojaFormData {
	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	return &GojaFormData{
		runtime:    runtime,
		buf:        buf,
		writer:     writer,
		fieldNames: make(map[string]struct{}),
		values:     make(map[string][]string),
		closed:     false,
	}
}

func (fd *GojaFormData) Append(call goja.FunctionCall) goja.Value {
	if fd.closed {
		return fd.runtime.ToValue("cannot append to closed GojaFormData")
	}

	fieldName := call.Argument(0).String()
	value := call.Argument(1).String()

	fieldName = strings.TrimSpace(fieldName)
	fd.values[fieldName] = append(fd.values[fieldName], value)

	if _, exists := fd.fieldNames[fieldName]; !exists {
		fd.fieldNames[fieldName] = struct{}{}
		writer, err := fd.writer.CreateFormField(fieldName)
		if err != nil {
			return fd.runtime.ToValue(err.Error())
		}
		_, err = writer.Write([]byte(value))
		if err != nil {
			return fd.runtime.ToValue(err.Error())
		}
	}

	return goja.Undefined()
}

func (fd *GojaFormData) Delete(call goja.FunctionCall) goja.Value {
	if fd.closed {
		return fd.runtime.ToValue("cannot delete from closed GojaFormData")
	}

	fieldName := call.Argument(0).String()
	fieldName = strings.TrimSpace(fieldName)

	delete(fd.fieldNames, fieldName)
	delete(fd.values, fieldName)

	return goja.Undefined()
}

func (fd *GojaFormData) Entries(call goja.FunctionCall) goja.Value {
	if fd.closed {
		return fd.runtime.ToValue("cannot get entries from closed GojaFormData")
	}

	iter := fd.runtime.NewArray()
	index := 0
	for key, values := range fd.values {
		for _, value := range values {
			entry := fd.runtime.NewObject()
			entry.Set("0", key)
			entry.Set("1", value)
			iter.Set(strconv.Itoa(index), entry)
			index++
		}
	}

	return iter
}

func (fd *GojaFormData) Get(call goja.FunctionCall) goja.Value {
	if fd.closed {
		return fd.runtime.ToValue("cannot get value from closed GojaFormData")
	}

	fieldName := call.Argument(0).String()
	fieldName = strings.TrimSpace(fieldName)

	if values, exists := fd.values[fieldName]; exists && len(values) > 0 {
		return fd.runtime.ToValue(values[0])
	}

	return goja.Undefined()
}

func (fd *GojaFormData) GetAll(call goja.FunctionCall) goja.Value {
	if fd.closed {
		return fd.runtime.ToValue("cannot get all values from closed GojaFormData")
	}

	fieldName := call.Argument(0).String()
	fieldName = strings.TrimSpace(fieldName)

	iter := fd.runtime.NewArray()
	if values, exists := fd.values[fieldName]; exists {
		for i, value := range values {
			iter.Set(strconv.Itoa(i), value)
		}
	}

	return iter
}

func (fd *GojaFormData) Has(call goja.FunctionCall) goja.Value {
	if fd.closed {
		return fd.runtime.ToValue("cannot check key in closed GojaFormData")
	}

	fieldName := call.Argument(0).String()
	fieldName = strings.TrimSpace(fieldName)

	_, exists := fd.fieldNames[fieldName]
	return fd.runtime.ToValue(exists)
}

func (fd *GojaFormData) Keys(call goja.FunctionCall) goja.Value {
	if fd.closed {
		return fd.runtime.ToValue("cannot get keys from closed GojaFormData")
	}

	iter := fd.runtime.NewArray()
	index := 0
	for key := range fd.fieldNames {
		iter.Set(strconv.Itoa(index), key)
		index++
	}

	return iter
}

func (fd *GojaFormData) Set(call goja.FunctionCall) goja.Value {
	if fd.closed {
		return fd.runtime.ToValue("cannot set value in closed GojaFormData")
	}

	fieldName := call.Argument(0).String()
	value := call.Argument(1).String()

	fieldName = strings.TrimSpace(fieldName)
	fd.values[fieldName] = []string{value}

	if _, exists := fd.fieldNames[fieldName]; !exists {
		fd.fieldNames[fieldName] = struct{}{}
		writer, err := fd.writer.CreateFormField(fieldName)
		if err != nil {
			return fd.runtime.ToValue(err.Error())
		}
		_, err = writer.Write([]byte(value))
		if err != nil {
			return fd.runtime.ToValue(err.Error())
		}
	}

	return goja.Undefined()
}

func (fd *GojaFormData) Values(call goja.FunctionCall) goja.Value {
	if fd.closed {
		return fd.runtime.ToValue("cannot get values from closed GojaFormData")
	}

	iter := fd.runtime.NewArray()
	index := 0
	for _, values := range fd.values {
		for _, value := range values {
			iter.Set(strconv.Itoa(index), value)
			index++
		}
	}

	return iter
}

func (fd *GojaFormData) GetContentType() goja.Value {
	if !fd.closed {
		fd.writer.Close()
		fd.closed = true
	}
	return fd.runtime.ToValue(fd.writer.FormDataContentType())
}

func (fd *GojaFormData) GetBuffer() goja.Value {
	if !fd.closed {
		fd.writer.Close()
		fd.closed = true
	}
	return fd.runtime.ToValue(fd.buf.String())
}
