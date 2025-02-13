package extension_repo

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"seanime/internal/extension"
	"seanime/internal/goja/goja_runtime"
	"seanime/internal/hook"
	"seanime/internal/util"
	"slices"
	"strings"

	"github.com/dop251/goja"
	"github.com/rs/zerolog"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Anime Torrent provider
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) loadPluginExtension(loader *goja.Runtime, ext *extension.Extension, hm hook.HookManager) (err error) {
	defer util.HandlePanicInModuleWithError("extension_repo/loadPluginExtension", &err)

	switch ext.Language {
	case extension.LanguageJavascript:
		err = r.loadPluginExtensionJS(loader, ext, extension.LanguageJavascript, hm)
	case extension.LanguageTypescript:
		err = r.loadPluginExtensionJS(loader, ext, extension.LanguageTypescript, hm)
	default:
		err = fmt.Errorf("unsupported language: %v", ext.Language)
	}

	if err != nil {
		return
	}

	return
}

func (r *Repository) loadPluginExtensionJS(loader *goja.Runtime, ext *extension.Extension, language extension.Language, hm hook.HookManager) error {
	_, err := NewGojaPlugin(loader, ext, language, r.logger, r.gojaRuntimeManager, hm)
	if err != nil {
		return err
	}

	// Add the extension to the map
	retExt := extension.NewPluginExtension(ext)
	r.extensionBank.Set(ext.ID, retExt)
	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Plugin
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type GojaPlugin struct {
	ext            *extension.Extension
	logger         *zerolog.Logger
	pool           *goja_runtime.Pool
	runtimeManager *goja_runtime.Manager
}

func NewGojaPluginLoader(logger *zerolog.Logger, runtimeManager *goja_runtime.Manager, hm hook.HookManager) *goja.Runtime {
	runtime := goja.New()
	ShareBinds(runtime, logger)
	BindHooks(runtime, hm, runtimeManager)

	// Preinitialize the runtime pool so that runtimeManager.pool is not nil
	if _, err := runtimeManager.GetOrCreatePool(func() *goja.Runtime {
		rt := goja.New()
		ShareBinds(rt, logger)
		return rt
	}); err != nil {
		logger.Error().Err(err).Msg("failed to initialize runtime pool")
	}

	return runtime
}

func NewGojaPlugin(loader *goja.Runtime, ext *extension.Extension, language extension.Language, logger *zerolog.Logger, runtimeManager *goja_runtime.Manager, hm hook.HookManager) (*GojaPlugin, error) {
	pool, err := runtimeManager.GetOrCreatePool(func() *goja.Runtime {
		runtime := goja.New()
		ShareBinds(runtime, logger)
		return runtime
	})
	if err != nil {
		return nil, err
	}

	_, err = loader.RunString(ext.Payload)
	if err != nil {
		return nil, err
	}

	// Call init() if it exists, so that plugin initialization runs
	if initFunc := loader.Get("init"); initFunc != nil && initFunc != goja.Undefined() {
		_, err = loader.RunString("init();")
		if err != nil {
			return nil, fmt.Errorf("failed to run init: %w", err)
		}
	}

	return &GojaPlugin{
		ext:            ext,
		logger:         logger,
		pool:           pool,
		runtimeManager: runtimeManager,
	}, nil
}

// SetupGojaPluginExtensionVM creates a new JavaScript VM with the plugin source code loaded
// func SetupGojaPluginExtensionVM(ext *extension.Extension, language extension.Language, logger *zerolog.Logger, runtimeManager *goja_runtime.Manager, hm hook.HookManager) (func() (*goja.Runtime, error), error) {
// 	logger.Trace().Str("id", ext.ID).Any("language", language).Msgf("extensions: Creating javascript VM")

// 	source := ext.Payload
// 	if language == extension.LanguageTypescript {
// 		var err error
// 		source, err = JSVMTypescriptToJS(ext.Payload)
// 		if err != nil {
// 			logger.Error().Err(err).Str("id", ext.ID).Msg("extensions: Failed to convert typescript")
// 			return nil, err
// 		}
// 	}

// 	// Compile the program once, to be reused by all VMs
// 	program, err := goja.Compile("", source, false)
// 	if err != nil {
// 		return nil, fmt.Errorf("compilation failed: %w", err)
// 	}

// 	return func() (*goja.Runtime, error) {
// 		vm := goja.New()
// 		vm.SetParserOptions(parser.WithDisableSourceMaps)

// 		ShareBinds(vm, logger)
// 		BindHooks(vm, hm, runtimeManager)
// 		_, err := vm.RunProgram(program)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to run program: %w", err)
// 		}

// 		// Call init() if defined to trigger initialization and subscribe hooks
// 		if initFunc := vm.Get("init"); initFunc != nil && initFunc != goja.Undefined() {
// 			_, err = vm.RunString("init();")
// 			if err != nil {
// 				return nil, fmt.Errorf("failed to call init: %w", err)
// 			}
// 		}

// 		return vm, nil
// 	}, nil
// }

type PluginContext struct {
	HookManager hook.HookManager
}

func BindHooks(loader *goja.Runtime, hm hook.HookManager, runtimeManager *goja_runtime.Manager) {
	fm := FieldMapper{}

	appType := reflect.TypeOf(hm)
	appValue := reflect.ValueOf(hm)
	totalMethods := appType.NumMethod()
	excludeHooks := []string{"OnServe"}

	for i := 0; i < totalMethods; i++ {
		method := appType.Method(i)
		if !strings.HasPrefix(method.Name, "On") || slices.Contains(excludeHooks, method.Name) {
			continue // not a hook or excluded
		}

		jsName := fm.MethodName(appType, method)

		loader.Set(jsName, func(callback string, tags ...string) {
			callback = `function(e) { $app = e.app; return (` + callback + `).call(undefined, e); }`
			pr := goja.MustCompile("", "{("+callback+").apply(undefined, __args)}", true)

			tagsAsValues := make([]reflect.Value, len(tags))
			for i, tag := range tags {
				tagsAsValues[i] = reflect.ValueOf(tag)
			}

			hookInstance := appValue.MethodByName(method.Name).Call(tagsAsValues)[0]
			hookBindFunc := hookInstance.MethodByName("BindFunc")

			handlerType := hookBindFunc.Type().In(0)

			handler := reflect.MakeFunc(handlerType, func(args []reflect.Value) (results []reflect.Value) {
				handlerArgs := make([]any, len(args))

				err := runtimeManager.Run(context.Background(), func(executor *goja.Runtime) error {
					for i, arg := range args {
						handlerArgs[i] = convertArg(executor, arg)
					}
					executor.Set("$app", goja.Undefined())
					executor.Set("__args", handlerArgs)
					res, err := executor.RunProgram(pr)
					executor.Set("__args", goja.Undefined())

					// (legacy) check for returned Go error value
					if res != nil {
						if resErr, ok := res.Export().(error); ok {
							return resErr
						}
					}

					return normalizeException(err)
				})

				return []reflect.Value{reflect.ValueOf(&err).Elem()}
			})

			// register the wrapped hook handler
			hookBindFunc.Call([]reflect.Value{handler})

		})
	}

	/////

	// for i := 0; i < totalMethods; i++ {
	// 	method := appType.Method(i)
	// 	if !strings.HasPrefix(method.Name, "On") || slices.Contains(excludeHooks, method.Name) {
	// 		continue // not a hook or excluded
	// 	}

	// 	jsName := fm.MethodName(appType, method)

	// 	hookWrapper := func(callback goja.Value, tags ...string) {
	// 		callbackStr := callback.String()
	// 		compiledCallback := "function(e) { $app = e.app; return (" + callbackStr + ").call(undefined, e); }"
	// 		//			compiledCallback := "function(e) { if(!e.next && typeof e.Next === 'function') { e.next = e.Next; } $app = e.app; return (" + callbackStr + ").call(undefined, e); }"
	// 		pr := goja.MustCompile("", "{("+compiledCallback+").apply(undefined, __args)}", true)

	// 		tagsAsValues := make([]reflect.Value, len(tags))
	// 		for i, tag := range tags {
	// 			tagsAsValues[i] = reflect.ValueOf(tag)
	// 		}

	// 		hookInstance := appValue.MethodByName(method.Name).Call(tagsAsValues)[0]
	// 		hookBindFunc := hookInstance.MethodByName("BindFunc")

	// 		handlerType := hookBindFunc.Type().In(0)

	// 		handler := reflect.MakeFunc(handlerType, func(args []reflect.Value) (results []reflect.Value) {
	// 			handlerArgs := make([]any, len(args))
	// 			for i, arg := range args {
	// 				//util.Spew(arg.Interface())
	// 				//handlerArgs[i] = arg.Interface()
	// 				n := convertArg(vm, arg)
	// 				//util.Spew(n)
	// 				handlerArgs[i] = n
	// 			}

	// 			vm.Set("$app", goja.Undefined())
	// 			vm.Set("__args", handlerArgs)
	// 			_, err := vm.RunProgram(pr)
	// 			vm.Set("__args", goja.Undefined())

	// 			return []reflect.Value{reflect.ValueOf(&err).Elem()}
	// 		})

	// 		// register the wrapped hook handler
	// 		hookBindFunc.Call([]reflect.Value{handler})
	// 	}

	// 	// set the hook under its original name
	// 	ctxObj.Set(jsName, hookWrapper)
	// 	// also set the hook with a lower-case initial (e.g., onGetBaseAnime) if not already the same
	// 	lowerJsName := strings.ToLower(jsName[0:1]) + jsName[1:]
	// 	if lowerJsName != jsName {
	// 		ctxObj.Set(lowerJsName, hookWrapper)
	// 	}
	// }

}

////

// normalizeException checks if the provided error is a goja.Exception
// and attempts to return its underlying Go error.
//
// note: using just goja.Exception.Unwrap() is insufficient and may falsely result in nil.
func normalizeException(err error) error {
	if err == nil {
		return nil
	}

	jsException, ok := err.(*goja.Exception)
	if !ok {
		return err // no exception
	}

	switch v := jsException.Value().Export().(type) {
	case error:
		err = v
	case map[string]any: // goja.GoError
		if vErr, ok := v["value"].(error); ok {
			err = vErr
		}
	}

	return err
}

func mapStructToJSObject(vm *goja.Runtime, value interface{}) *goja.Object {
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return vm.NewObject()
		}
		v = v.Elem()
	}
	t := v.Type()
	obj := vm.NewObject()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" { // unexported
			continue
		}
		var fieldName string
		tag := field.Tag.Get("json")
		if tag != "" {
			parts := strings.Split(tag, ",")
			if parts[0] != "" && parts[0] != "-" {
				fieldName = parts[0]
			} else {
				fieldName = convertGoToJSName(field.Name)
			}
		} else {
			fieldName = convertGoToJSName(field.Name)
		}

		fieldVal := v.Field(i)
		if fieldVal.CanSet() {
			// Create live getter/setter to reflect changes back to the Go struct
			iCopy := i // capture loop variable
			getter := vm.ToValue(func(call goja.FunctionCall) goja.Value {
				return vm.ToValue(convertArg(vm, v.Field(iCopy)))
			})
			setter := vm.ToValue(func(call goja.FunctionCall) goja.Value {
				if len(call.Arguments) > 0 {
					newVal := call.Arguments[0]
					exported := newVal.Export()
					field := v.Field(iCopy)
					newGoVal := reflect.ValueOf(exported)

					// Handle pointer types
					if field.Kind() == reflect.Ptr {
						if newGoVal.Type().AssignableTo(field.Type().Elem()) {
							// Create new pointer and set value
							newPtr := reflect.New(field.Type().Elem())
							newPtr.Elem().Set(newGoVal)
							field.Set(newPtr)
						} else if newGoVal.Type().ConvertibleTo(field.Type().Elem()) {
							// Create new pointer and set converted value
							newPtr := reflect.New(field.Type().Elem())
							newPtr.Elem().Set(newGoVal.Convert(field.Type().Elem()))
							field.Set(newPtr)
						}
					} else if newGoVal.Type().AssignableTo(field.Type()) {
						field.Set(newGoVal)
					} else if newGoVal.Type().ConvertibleTo(field.Type()) {
						field.Set(newGoVal.Convert(field.Type()))
					}
				}
				return goja.Undefined()
			})
			obj.DefineAccessorProperty(fieldName, getter, setter, goja.Flag(1), goja.Flag(1))
		} else {
			obj.Set(fieldName, convertArg(vm, fieldVal))
		}
	}

	// Attempt to fetch the 'Next' method from both pointer and value
	method := reflect.ValueOf(value).MethodByName("Next")
	if !method.IsValid() {
		method = v.MethodByName("Next")
	}

	if method.IsValid() {
		nextFn := func(call goja.FunctionCall) goja.Value {
			results := method.Call(nil)
			if len(results) > 0 {
				return vm.ToValue(results[0].Interface())
			}
			return goja.Undefined()
		}
		obj.Set("next", vm.ToValue(nextFn))
		obj.Set("Next", vm.ToValue(nextFn))
	}

	// Attach a custom toString method to return formatted representation
	obj.Set("toString", vm.ToValue(func(call goja.FunctionCall) goja.Value {
		bs, err := json.Marshal(value)
		if err != nil {
			return vm.ToValue(fmt.Sprintf("%+v", value))
		}
		return vm.ToValue(string(bs))
	}))

	return obj
}

func convertArg(vm *goja.Runtime, arg reflect.Value) interface{} {
	if !arg.IsValid() {
		return nil
	}

	// Handle pointer types recursively
	if arg.Kind() == reflect.Ptr {
		if arg.IsNil() {
			return nil
		}
		if arg.Elem().Kind() == reflect.Struct {
			return mapStructToJSObject(vm, arg.Interface())
		}
		return convertArg(vm, arg.Elem())
	}

	// Handle struct types as JS objects
	if arg.Kind() == reflect.Struct {
		return mapStructToJSObject(vm, arg.Interface())
	}

	// Handle slices and arrays recursively
	if arg.Kind() == reflect.Slice || arg.Kind() == reflect.Array {
		n := arg.Len()
		result := make([]interface{}, n)
		for i := 0; i < n; i++ {
			result[i] = convertArg(vm, arg.Index(i))
		}
		return vm.ToValue(result)
	}

	// Handle maps
	if arg.Kind() == reflect.Map {
		obj := vm.NewObject()
		iter := arg.MapRange()
		for iter.Next() {
			k := iter.Key()
			v := iter.Value()
			if k.Kind() == reflect.String {
				obj.Set(k.String(), convertArg(vm, v))
			}
		}
		return obj
	}

	// Convert primitive types
	return vm.ToValue(arg.Interface())
}
