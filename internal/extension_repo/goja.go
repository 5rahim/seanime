package extension_repo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"seanime/internal/extension"
	goja_bindings "seanime/internal/goja/goja_bindings"
	"seanime/internal/library/anime"
	"seanime/internal/plugin"
	"sync"
	"time"

	"github.com/5rahim/habari"
	"github.com/dop251/goja"
	gojabuffer "github.com/dop251/goja_nodejs/buffer"
	gojarequire "github.com/dop251/goja_nodejs/require"
	gojaurl "github.com/dop251/goja_nodejs/url"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/rs/zerolog"
	"github.com/spf13/cast"
)

// GojaExtension is stored in the repository extension map, giving access to the VMs.
// Current use: Kill the VM when the extension is unloaded.
type GojaExtension interface {
	PutVM(*goja.Runtime)
	ClearInterrupt()
	GetExtension() *extension.Extension
}

var cachedArrayOfTypes = plugin.NewStore[reflect.Type, reflect.Type](nil)

func BindUserConfig(vm *goja.Runtime, ext *extension.Extension, logger *zerolog.Logger) {
	vm.Set("$getUserPreference", func(call goja.FunctionCall) goja.Value {
		if ext.SavedUserConfig == nil {
			return goja.Undefined()
		}

		key := call.Argument(0).String()
		value, ok := ext.SavedUserConfig.Values[key]
		if !ok {
			// Check if the field has a default value
			for _, field := range ext.UserConfig.Fields {
				if field.Name == key && field.Default != "" {
					return vm.ToValue(field.Default)
				}
			}

			return goja.Undefined()
		}

		return vm.ToValue(value)
	})
}

// ShareBinds binds the shared bindings to the VM
// This is called once per VM
func ShareBinds(vm *goja.Runtime, logger *zerolog.Logger) {
	registry := new(gojarequire.Registry)
	registry.Enable(vm)

	fm := goja_bindings.DefaultFieldMapper{}
	vm.SetFieldNameMapper(fm)
	// goja.TagFieldNameMapper("json", true)

	bindings := []struct {
		name string
		fn   func(*goja.Runtime) error
	}{
		{"url", func(vm *goja.Runtime) error { gojaurl.Enable(vm); return nil }},
		{"buffer", func(vm *goja.Runtime) error { gojabuffer.Enable(vm); return nil }},
		{"fetch", func(vm *goja.Runtime) error { goja_bindings.BindFetch(vm); return nil }},
		{"console", func(vm *goja.Runtime) error { goja_bindings.BindConsole(vm, logger); return nil }},
		{"formData", func(vm *goja.Runtime) error { goja_bindings.BindFormData(vm); return nil }},
		{"document", func(vm *goja.Runtime) error { goja_bindings.BindDocument(vm); return nil }},
		{"crypto", func(vm *goja.Runtime) error { goja_bindings.BindCrypto(vm); return nil }},
		{"torrentUtils", func(vm *goja.Runtime) error { goja_bindings.BindTorrentUtils(vm); return nil }},
	}

	for _, binding := range bindings {
		if err := binding.fn(vm); err != nil {
			logger.Error().Err(err).Str("name", binding.name).Msg("failed to bind")
		}
	}

	vm.Set("__isOffline__", plugin.GlobalAppContext.IsOffline())

	vm.Set("$toString", func(raw any, maxReaderBytes int) (string, error) {
		switch v := raw.(type) {
		case io.Reader:
			if maxReaderBytes == 0 {
				maxReaderBytes = 32 << 20 // 32 MB
			}

			limitReader := io.LimitReader(v, int64(maxReaderBytes))

			bodyBytes, readErr := io.ReadAll(limitReader)
			if readErr != nil {
				return "", readErr
			}

			return string(bodyBytes), nil
		default:
			str, err := cast.ToStringE(v)
			if err == nil {
				return str, nil
			}

			// as a last attempt try to json encode the value
			rawBytes, _ := json.Marshal(raw)

			return string(rawBytes), nil
		}
	})

	vm.Set("$toBytes", func(raw any) ([]byte, error) {
		switch v := raw.(type) {
		case io.Reader:
			bodyBytes, readErr := io.ReadAll(v)
			if readErr != nil {
				return nil, readErr
			}

			return bodyBytes, nil
		case string:
			return []byte(v), nil
		case []byte:
			return v, nil
		case []rune:
			return []byte(string(v)), nil
		default:
			// as a last attempt try to json encode the value
			rawBytes, _ := json.Marshal(raw)
			return rawBytes, nil
		}
	})

	vm.Set("$toError", func(raw any) error {
		if err, ok := raw.(error); ok {
			return err
		}

		return fmt.Errorf("%v", raw)
	})

	vm.Set("$sleep", func(milliseconds int64) {
		time.Sleep(time.Duration(milliseconds) * time.Millisecond)
	})

	vm.Set("$arrayOf", func(model any) any {
		mt := reflect.TypeOf(model)
		st := cachedArrayOfTypes.GetOrSet(mt, func() reflect.Type {
			return reflect.SliceOf(mt)
		})

		return reflect.New(st).Elem().Addr().Interface()
	})

	vm.Set("$unmarshal", func(data, dst any) error {
		raw, err := json.Marshal(data)
		if err != nil {
			return err
		}

		return json.Unmarshal(raw, &dst)
	})

	vm.Set("$toPointer", func(data interface{}) interface{} {
		if data == nil {
			return nil
		}
		v := data
		return &v
	})

	vm.Set("$Context", func(call goja.ConstructorCall) *goja.Object {
		var instance context.Context

		oldCtx, ok := call.Argument(0).Export().(context.Context)
		if ok {
			instance = oldCtx
		} else {
			instance = context.Background()
		}

		key := call.Argument(1).Export()
		if key != nil {
			instance = context.WithValue(instance, key, call.Argument(2).Export())
		}

		instanceValue := vm.ToValue(instance).(*goja.Object)
		instanceValue.SetPrototype(call.This.Prototype())

		return instanceValue
	})

	//
	// Habari
	//
	habariObj := vm.NewObject()
	_ = habariObj.Set("parse", func(filename string) *habari.Metadata {
		return habari.Parse(filename)
	})
	vm.Set("$habari", habariObj)

	//
	// Anime Utils
	//
	animeUtilsObj := vm.NewObject()
	_ = animeUtilsObj.Set("newLocalFileWrapper", func(lfs []*anime.LocalFile) *anime.LocalFileWrapper {
		return anime.NewLocalFileWrapper(lfs)
	})
	vm.Set("$animeUtils", animeUtilsObj)

	vm.Set("$waitGroup", func() *sync.WaitGroup {
		return &sync.WaitGroup{}
	})

	// Run a function in a new goroutine
	// The Goja runtime is not thread safe, so nothing related to the VM should be done in this goroutine
	// You can use the $waitGroup to wait for multiple goroutines to finish
	// You can use $store to communicate with the main thread
	vm.Set("$unsafeGoroutine", func(fn func()) {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error().Err(fmt.Errorf("%v", r)).Msg("goroutine panic")
				}
			}()
			fn()
		}()
	})
}

// JSVMTypescriptToJS converts typescript to javascript
func JSVMTypescriptToJS(ts string) (string, error) {
	result := api.Transform(ts, api.TransformOptions{
		Target:           api.ES2018,
		Loader:           api.LoaderTS,
		Format:           api.FormatDefault,
		MinifyWhitespace: true,
		MinifySyntax:     true,
		Sourcemap:        api.SourceMapNone,
	})

	if len(result.Errors) > 0 {
		var errMsgs []string
		for _, err := range result.Errors {
			errMsgs = append(errMsgs, err.Text)
		}
		return "", fmt.Errorf("typescript compilation errors: %v", errMsgs)
	}

	return string(result.Code), nil
}

// structToMap converts a struct to "JSON-like" map for Goja extensions
// This is used to pass structs to Goja extensions
func structToMap(obj interface{}) map[string]interface{} {
	// Convert the struct to a map
	jsonData, err := json.Marshal(obj)
	if err != nil {
		return nil
	}

	var data map[string]interface{}
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		return nil
	}

	return data
}
