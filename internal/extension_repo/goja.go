package extension_repo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"seanime/internal/extension"
	goja_bindings "seanime/internal/goja/goja_bindings"
	"seanime/internal/plugin"
	"time"

	"github.com/dop251/goja"
	"github.com/dop251/goja/parser"
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
}

// SetupGojaExtensionVM creates a new JavaScript VM with the extension source code loaded
func SetupGojaExtensionVM(ext *extension.Extension, language extension.Language, logger *zerolog.Logger) (func() *goja.Runtime, *goja.Program, error) {
	logger.Trace().Str("id", ext.ID).Any("language", language).Msgf("extensions: Creating javascript VM")

	source := ext.Payload
	if language == extension.LanguageTypescript {
		var err error
		source, err = JSVMTypescriptToJS(ext.Payload)
		if err != nil {
			logger.Error().Err(err).Str("id", ext.ID).Msg("extensions: Failed to convert typescript")
			return nil, nil, err
		}
	}

	ext.Payload = source

	// Compile the program once, to be reused by all VMs
	program, err := goja.Compile("", source, false)
	if err != nil {
		logger.Error().Err(err).Str("id", ext.ID).Msg("extensions: Failed to compile program")
		return nil, nil, fmt.Errorf("compilation failed: %w", err)
	}

	return func() *goja.Runtime {
		vm := goja.New()
		vm.SetParserOptions(parser.WithDisableSourceMaps)
		// Bind the shared bindings
		ShareBinds(vm, logger)
		return vm
	}, program, nil
}

var cachedArrayOfTypes = plugin.NewStore[reflect.Type, reflect.Type](nil)

// ShareBinds binds the shared bindings to the VM
// This is called once per VM
func ShareBinds(vm *goja.Runtime, logger *zerolog.Logger) {
	registry := new(gojarequire.Registry)
	registry.Enable(vm)

	bindings := []struct {
		name string
		fn   func(*goja.Runtime) error
	}{
		{"url", func(vm *goja.Runtime) error { gojaurl.Enable(vm); return nil }},
		{"buffer", func(vm *goja.Runtime) error { gojabuffer.Enable(vm); return nil }},
		{"fetch", goja_bindings.BindFetch},
		{"console", func(vm *goja.Runtime) error { return goja_bindings.BindConsole(vm, logger) }},
		{"formData", goja_bindings.BindFormData},
		{"document", goja_bindings.BindDocument},
		{"crypto", goja_bindings.BindCrypto},
		{"torrentUtils", goja_bindings.BindTorrentUtils},
	}

	for _, binding := range bindings {
		if err := binding.fn(vm); err != nil {
			logger.Error().Err(err).Str("name", binding.name).Msg("failed to bind")
		}
	}

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
