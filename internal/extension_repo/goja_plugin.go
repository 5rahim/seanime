package extension_repo

import (
	"context"
	"fmt"
	"reflect"
	"seanime/internal/events"
	"seanime/internal/extension"
	"seanime/internal/goja/goja_runtime"
	"seanime/internal/hook"
	"seanime/internal/plugin"
	plugin_ui "seanime/internal/plugin/ui"
	"seanime/internal/util"
	goja_util "seanime/internal/util/goja"
	"slices"
	"strings"

	"github.com/dop251/goja"
	"github.com/dop251/goja/parser"
	"github.com/rs/zerolog"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Load Plugin
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) loadPluginExtension(ext *extension.Extension) (err error) {
	defer util.HandlePanicInModuleWithError("extension_repo/loadPluginExtension", &err)

	loader := NewGojaPluginLoader(ext, r.logger, r.gojaRuntimeManager)

	_, err = NewGojaPlugin(loader, ext, ext.Language, r.logger, r.gojaRuntimeManager, r.wsEventManager)
	if err != nil {
		return err
	}

	// Add the extension to the map
	retExt := extension.NewPluginExtension(ext)
	r.extensionBank.Set(ext.ID, retExt)

	return
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Plugin
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type GojaPlugin struct {
	ext            *extension.Extension
	logger         *zerolog.Logger
	pool           *goja_runtime.Pool
	runtimeManager *goja_runtime.Manager
	store          *plugin.Store[string, any]
	ui             *plugin_ui.UI
	scheduler      *goja_util.Scheduler
}

func (p *GojaPlugin) PutVM(vm *goja.Runtime) {
	p.pool.Put(vm)
}

// ClearInterrupt stops the UI VM and other modules.
// It is called when the extension is unloaded.
func (p *GojaPlugin) ClearInterrupt() {
	p.ui.ClearInterrupt()
	p.store.Stop()
}

func NewGojaPluginLoader(ext *extension.Extension, logger *zerolog.Logger, runtimeManager *goja_runtime.Manager) *goja.Runtime {
	loader := goja.New()
	ShareBinds(loader, logger)

	// Bind hooks to the loader
	BindHooks(loader, runtimeManager, ext)

	return loader
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func NewGojaPlugin(
	loader *goja.Runtime,
	ext *extension.Extension,
	language extension.Language,
	logger *zerolog.Logger,
	runtimeManager *goja_runtime.Manager,
	wsEventManager events.WSEventManagerInterface,
) (*GojaPlugin, error) {
	logger.Trace().Str("id", ext.ID).Msg("extensions: Loading plugin")

	p := &GojaPlugin{
		ext:            ext,
		logger:         logger,
		runtimeManager: runtimeManager,
		store:          plugin.NewStore[string, any](nil),
		scheduler:      goja_util.NewScheduler(),
		ui:             nil,
	}

	// Convert the payload to JavaScript if necessary
	source := ext.Payload
	if language == extension.LanguageTypescript {
		var err error
		source, err = JSVMTypescriptToJS(ext.Payload)
		if err != nil {
			logger.Error().Err(err).Str("id", ext.ID).Msg("extensions: Failed to convert typescript")
			return nil, err
		}
	}

	// Create a new pool for the plugin
	pool, err := runtimeManager.GetOrCreatePluginPool(ext.ID, func() *goja.Runtime {
		runtime := goja.New()
		ShareBinds(runtime, logger)
		p.BindPluginAPIs(runtime, logger)
		return runtime
	})
	if err != nil {
		return nil, err
	}

	//////// UI

	// Create a new VM for the UI
	// We only need one VM for the UI because it is registered once and there's no need to be thread safe
	uiVM := goja.New()
	fm := FieldMapper{}
	uiVM.SetParserOptions(parser.WithDisableSourceMaps)
	uiVM.SetFieldNameMapper(fm)
	ShareBinds(uiVM, logger)
	// Bind the store to the UI VM
	p.BindPluginAPIs(uiVM, logger)
	// Create a new UI instance
	p.ui = plugin_ui.NewUI(plugin_ui.NewUIOptions{
		Extension: ext,
		Logger:    logger,
		VM:        uiVM,
		WSManager: wsEventManager,
		Scheduler: p.scheduler,
	})

	////////

	// Bind the UI API to the loader so the plugin can register a new UI
	// Create a new object for the UI
	uiObj := loader.NewObject()
	// Set the register method on the UI object
	uiObj.Set("register", p.ui.Register)
	// Set the UI object in the loader
	loader.Set("$ui", uiObj)

	////////

	// Load the extension payload in the loader runtime
	_, err = loader.RunString(source)
	if err != nil {
		return nil, err
	}

	// Call init() if it exists, so that plugin initialization runs
	if initFunc := loader.Get("init"); initFunc != nil && initFunc != goja.Undefined() {
		_, err = loader.RunString("init();")
		if err != nil {
			return nil, fmt.Errorf("failed to run init: %w", err)
		}
		logger.Debug().Str("id", ext.ID).Msg("extensions: Plugin initialized")
	}

	p.pool = pool

	return p, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// BindPluginAPIs adds plugin-specific APIs
func (p *GojaPlugin) BindPluginAPIs(vm *goja.Runtime, logger *zerolog.Logger) {
	// Bind the app context
	//_ = vm.Set("$ctx", hook.GlobalHookManager.AppContext())

	// Bind the store
	p.store.Bind(vm, p.scheduler)
	// Bind mutable bindings
	goja_util.BindMutable(vm)

	// Bind permission-specific APIs
	if p.ext.Plugin != nil {
		for _, permission := range p.ext.Plugin.Permissions {
			switch permission.String() {
			case extension.PluginPermissionStorage.String():
				plugin.GlobalAppContext.BindStorage(vm, logger, p.ext)
			case extension.PluginPermissionAnilist.String():
				plugin.GlobalAppContext.BindAnilist(vm, logger, p.ext)
			case extension.PluginPermissionDatabase.String():
				plugin.GlobalAppContext.BindDatabase(vm, logger, p.ext)
			case extension.PluginPermissionOS.String():
				plugin.GlobalAppContext.BindOS(vm, logger, p.ext)
				plugin.GlobalAppContext.BindFilepath(vm, logger, p.ext)
				plugin.GlobalAppContext.BindFilesystem(vm, logger, p.ext)
			}
		}
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// BindHooks sets up hooks for the Goja runtime
func BindHooks(loader *goja.Runtime, runtimeManager *goja_runtime.Manager, ext *extension.Extension) {
	// Create a FieldMapper instance for method name mapping
	fm := FieldMapper{}

	// Get the type of the global hook manager
	appType := reflect.TypeOf(hook.GlobalHookManager)
	// Get the value of the global hook manager
	appValue := reflect.ValueOf(hook.GlobalHookManager)
	// Get the total number of methods in the global hook manager
	// i.e. OnGetAnime, OnGetAnimeDetails, etc.
	totalMethods := appType.NumMethod()
	// Define methods to exclude from binding
	excludeHooks := []string{"OnServe", ""}

	// Create a new JavaScript object to hold the hooks ($app)
	appObj := loader.NewObject()

	// Iterate through all methods of the global hook manager
	// i.e. OnGetAnime, OnGetAnimeDetails, etc.
	for i := 0; i < totalMethods; i++ {
		// Get the method at the current index
		method := appType.Method(i)

		// Check that the method name starts with "On" and is not excluded
		if !strings.HasPrefix(method.Name, "On") || slices.Contains(excludeHooks, method.Name) {
			continue // Skip to the next method if not a hook or excluded
		}

		// Map the method name to a JavaScript-friendly name
		// e.g. OnGetAnime -> onGetAnime
		jsName := fm.MethodName(appType, method)

		// Set the method on the app object with a callback function
		// e.g. $app.onGetAnime(callback, "tag1", "tag2")
		appObj.Set(jsName, func(callback string, tags ...string) {
			// Create a wrapper JavaScript function that calls the provided callback
			// This is necessary because the callback will be called with the provided args
			callback = `function(e) { return (` + callback + `).call(undefined, e); }`
			// Compile the callback into a Goja program
			pr := goja.MustCompile("", "{("+callback+").apply(undefined, __args)}", true)

			// Prepare the tags as reflect.Values for method invocation
			tagsAsValues := make([]reflect.Value, len(tags))
			for i, tag := range tags {
				tagsAsValues[i] = reflect.ValueOf(tag)
			}

			// Get the hook function from the global hook manager and invokes it with the provided tags
			// The invokation returns a hook instance
			// i.e. OnTaggedHook(tags...) -> TaggedHook / OnHook() -> Hook
			hookInstance := appValue.MethodByName(method.Name).Call(tagsAsValues)[0]

			// Get the BindFunc method from the hook instance
			hookBindFunc := hookInstance.MethodByName("BindFunc")

			// Get the expected handler type for the hook
			// i.e. func(e *hook_resolver.Resolver) error
			handlerType := hookBindFunc.Type().In(0)

			// Create a new handler function for the hook
			// - returns a new handler of the given handlerType that wraps the function
			handler := reflect.MakeFunc(handlerType, func(args []reflect.Value) (results []reflect.Value) {
				// Prepare arguments for the handler
				handlerArgs := make([]any, len(args))

				// Run the handler in an isolated "executor" runtime for concurrency
				err := runtimeManager.Run(context.Background(), ext.ID, func(executor *goja.Runtime) error {
					// Set the field name mapper for the executor
					executor.SetFieldNameMapper(fm)
					// Convert each argument (event property) to the appropriate type
					for i, arg := range args {
						handlerArgs[i] = arg.Interface()
					}
					// Set the global variable $ctx in the executor
					//executor.Set("$ctx", hook.GlobalHookManager.AppContext())
					executor.Set("__args", handlerArgs)
					// Execute the handler program
					res, err := executor.RunProgram(pr)
					// Clear the __args variable for this executor
					executor.Set("__args", goja.Undefined())
					// executor.Set("$ctx", goja.Undefined())

					// Check for returned Go error value
					if res != nil {
						if resErr, ok := res.Export().(error); ok {
							return resErr
						}
					}

					return normalizeException(err)
				})

				// Return the error as a reflect.Value
				return []reflect.Value{reflect.ValueOf(&err).Elem()}
			})

			// Register the wrapped hook handler
			hookBindFunc.Call([]reflect.Value{handler})

		})
	}

	// Set the $app object in the loader for JavaScript access
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

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
