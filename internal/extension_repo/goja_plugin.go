package extension_repo

import (
	"context"
	"reflect"
	"seanime/internal/extension"
	"seanime/internal/goja/goja_runtime"
	"seanime/internal/hook"
	"slices"
	"strings"

	"github.com/dop251/goja"
	"github.com/rs/zerolog"
)

type GojaPlugin struct {
	ext            *extension.Extension
	pool           *goja_runtime.Pool
	runtimeManager *goja_runtime.Manager
}

func NewGojaPlugin(ext *extension.Extension, language extension.Language, logger *zerolog.Logger, runtimeManager *goja_runtime.Manager, hm hook.HookManager) (*GojaPlugin, error) {
	initFn := func() (*goja.Runtime, error) {
		vm := goja.New()
		ShareBinds(vm, logger)
		return vm, nil
	}

	pool, err := runtimeManager.GetOrCreatePool(initFn)
	if err != nil {
		return nil, err
	}

	loader := goja.New()
	ShareBinds(loader, logger)
	hooksBinds(hm, loader, pool)

	_, err = loader.RunString(ext.Payload)
	if err != nil {
		logger.Error().Err(err).Msg("Error init")
		return nil, err
	}

	p := &GojaPlugin{
		ext:            ext,
		pool:           pool,
		runtimeManager: runtimeManager,
	}

	return p, nil
}

// hooksBinds adds wrapped "on*" hook methods by reflecting on the hook manager
func hooksBinds(app hook.HookManager, loader *goja.Runtime, executors *goja_runtime.Pool) {
	fm := FieldMapper{}

	appType := reflect.TypeOf(app)
	appValue := reflect.ValueOf(app)
	totalMethods := appType.NumMethod()
	excludeHooks := []string{"OnServe"}

	for i := 0; i < totalMethods; i++ {
		method := appType.Method(i)
		if !strings.HasPrefix(method.Name, "On") || slices.Contains(excludeHooks, method.Name) {
			continue // not a hook or excluded
		}

		jsName := fm.MethodName(appType, method)

		// register the hook to the loader
		loader.Set(jsName, func(callback string, tags ...string) {
			// overwrite the global $app with the hook scoped instance
			callback = `function(e) { $app = e.app; return (` + callback + `).call(undefined, e) }`
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
				for i, arg := range args {
					handlerArgs[i] = arg.Interface()
				}

				vm, err := executors.Get(context.Background())
				if err != nil {
					return []reflect.Value{reflect.ValueOf(err)}
				}
				err = func(executor *goja.Runtime) error {
					executor.Set("$app", goja.Undefined())
					executor.Set("__args", handlerArgs)
					res, err := executor.RunProgram(pr)
					executor.Set("__args", goja.Undefined())
					if res != nil {
						if resErr, ok := res.Export().(error); ok {
							return resErr
						}
					}
					return normalizeException(err)
				}(vm)
				executors.Put(vm)

				return []reflect.Value{reflect.ValueOf(&err).Elem()}
			})

			// register the wrapped hook handler
			hookBindFunc.Call([]reflect.Value{handler})
		})
	}
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
