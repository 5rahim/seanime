package extension_repo

import (
	"context"
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
	loader := goja.New()
	// Add bindings to the loader
	ShareBinds(loader, logger)
	// PluginBinds(loader, logger)
	// Bind hooks to the loader
	BindHooks(loader, hm, runtimeManager)

	// Preinitialize the runtime pool so that runtimeManager.pool is not nil
	if _, err := runtimeManager.GetOrCreatePool(func() *goja.Runtime {
		rt := goja.New()
		ShareBinds(rt, logger)
		PluginBinds(rt, logger)
		return rt
	}); err != nil {
		logger.Error().Err(err).Msg("failed to initialize runtime pool")
	}

	return loader
}

func NewGojaPlugin(loader *goja.Runtime, ext *extension.Extension, language extension.Language, logger *zerolog.Logger, runtimeManager *goja_runtime.Manager, hm hook.HookManager) (*GojaPlugin, error) {
	pool, err := runtimeManager.GetOrCreatePool(func() *goja.Runtime {
		runtime := goja.New()
		ShareBinds(runtime, logger)
		PluginBinds(runtime, logger)
		return runtime
	})
	if err != nil {
		return nil, err
	}

	// Load the extension payload
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

type PluginContext struct {
}

func BindHooks(loader *goja.Runtime, hm hook.HookManager, runtimeManager *goja_runtime.Manager) {
	fm := FieldMapper{}

	appType := reflect.TypeOf(hm)
	appValue := reflect.ValueOf(hm)
	totalMethods := appType.NumMethod()
	excludeHooks := []string{"OnServe"}

	appObj := loader.NewObject()

	for i := 0; i < totalMethods; i++ {
		method := appType.Method(i)
		if !strings.HasPrefix(method.Name, "On") || slices.Contains(excludeHooks, method.Name) {
			continue // not a hook or excluded
		}

		jsName := fm.MethodName(appType, method)

		appObj.Set(jsName, func(callback string, tags ...string) {
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

				// Run the handler in a isolated "executor" runtime to allow for concurrency
				// This runtime has shared bindings and plugin bindings
				err := runtimeManager.Run(context.Background(), func(executor *goja.Runtime) error {
					executor.SetFieldNameMapper(fm)
					for i, arg := range args {
						// handlerArgs[i] = convertArg(executor, arg)
						handlerArgs[i] = arg.Interface()
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

	loader.Set("$app", appObj)
}

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
