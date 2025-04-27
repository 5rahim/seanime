package extension_repo

import (
	"context"
	"fmt"
	"reflect"
	"seanime/internal/events"
	"seanime/internal/extension"
	goja_bindings "seanime/internal/goja/goja_bindings"
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
	"github.com/samber/lo"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Load Plugin
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) loadPluginExtension(ext *extension.Extension) (err error) {
	defer util.HandlePanicInModuleWithError("extension_repo/loadPluginExtension", &err)

	_, gojaExt, err := NewGojaPlugin(ext, ext.Language, r.logger, r.gojaRuntimeManager, r.wsEventManager)
	if err != nil {
		return err
	}

	// Add the extension to the map
	retExt := extension.NewPluginExtension(ext)
	r.extensionBank.Set(ext.ID, retExt)
	r.gojaExtensions.Set(ext.ID, gojaExt)

	return
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Plugin
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type GojaPlugin struct {
	ext             *extension.Extension
	logger          *zerolog.Logger
	pool            *goja_runtime.Pool
	runtimeManager  *goja_runtime.Manager
	store           *plugin.Store[string, any]
	storage         *plugin.Storage
	ui              *plugin_ui.UI
	scheduler       *goja_util.Scheduler
	loader          *goja.Runtime
	unbindHookFuncs []func()
	interrupted     bool
	wsEventManager  events.WSEventManagerInterface
}

func (p *GojaPlugin) GetExtension() *extension.Extension {
	return p.ext
}

func (p *GojaPlugin) PutVM(vm *goja.Runtime) {
	p.pool.Put(vm)
}

// ClearInterrupt stops the UI VM and other modules.
// It is called when the extension is unloaded.
func (p *GojaPlugin) ClearInterrupt() {
	if p.interrupted {
		return
	}

	p.interrupted = true

	p.logger.Debug().Msg("plugin: Interrupting plugin")
	// Unload the UI
	if p.ui != nil {
		p.ui.Unload(false)
	}
	// Clear the interrupt
	if p.loader != nil {
		p.loader.ClearInterrupt()
	}
	// Stop the store
	if p.store != nil {
		p.store.Stop()
	}
	// Stop the storage
	if p.storage != nil {
		p.storage.Stop()
	}
	// Delete the plugin pool
	if p.runtimeManager != nil {
		p.runtimeManager.DeletePluginPool(p.ext.ID)
	}
	p.logger.Debug().Msgf("plugin: Unbinding hooks (%d)", len(p.unbindHookFuncs))
	// Unbind all hooks
	for _, unbindHookFunc := range p.unbindHookFuncs {
		unbindHookFunc()
	}
	// Run garbage collection
	// runtime.GC()
	p.logger.Debug().Msg("plugin: Interrupted plugin")
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func NewGojaPlugin(
	ext *extension.Extension,
	language extension.Language,
	mLogger *zerolog.Logger,
	runtimeManager *goja_runtime.Manager,
	wsEventManager events.WSEventManagerInterface,
) (*GojaPlugin, GojaExtension, error) {
	logger := lo.ToPtr(mLogger.With().Str("id", ext.ID).Logger())
	defer util.HandlePanicInModuleThen("extension_repo/NewGojaPlugin", func() {
		logger.Error().Msg("extensions: Failed to create Goja plugin")
	})

	logger.Trace().Msg("extensions: Loading plugin")

	// 1. Create a new plugin instance
	p := &GojaPlugin{
		ext:             ext,
		logger:          logger,
		runtimeManager:  runtimeManager,
		store:           plugin.NewStore[string, any](nil), // Create a store (must be stopped when unloading)
		scheduler:       goja_util.NewScheduler(),          // Create a scheduler (must be stopped when unloading)
		ui:              nil,                               // To be initialized
		loader:          goja.New(),                        // To be initialized
		unbindHookFuncs: []func(){},
		wsEventManager:  wsEventManager,
	}

	// 2. Create a new loader for the plugin
	// Bind shared APIs to the loader
	ShareBinds(p.loader, logger)
	BindUserConfig(p.loader, ext, logger)
	// Bind hooks to the loader
	p.bindHooks()

	// 3. Convert the payload to JavaScript if necessary
	source := ext.Payload
	if language == extension.LanguageTypescript {
		var err error
		source, err = JSVMTypescriptToJS(ext.Payload)
		if err != nil {
			logger.Error().Err(err).Msg("extensions: Failed to convert typescript")
			return nil, nil, err
		}
	}

	// 4. Create a new pool for the plugin hooks (must be deleted when unloading)
	var err error
	p.pool, err = runtimeManager.GetOrCreatePrivatePool(ext.ID, func() *goja.Runtime {
		runtime := goja.New()
		ShareBinds(runtime, logger)
		BindUserConfig(runtime, ext, logger)
		p.BindPluginAPIs(runtime, logger)
		return runtime
	})
	if err != nil {
		return nil, nil, err
	}

	//////// UI

	// 5. Create a new VM for the UI (The UI uses a single VM instead of a pool in order to share state)
	// (must be interrupted when unloading)
	uiVM := goja.New()
	uiVM.SetParserOptions(parser.WithDisableSourceMaps)
	// Bind shared APIs
	ShareBinds(uiVM, logger)
	BindUserConfig(uiVM, ext, logger)
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

	go func() {
		<-p.ui.Destroyed()
		p.logger.Warn().Msg("plugin: UI interrupted, interrupting plugin")
		p.ClearInterrupt()
	}()

	// 6. Bind the UI API to the loader so the plugin can register a new UI
	//	$ui.register(callback)
	uiObj := p.loader.NewObject()
	_ = uiObj.Set("register", p.ui.Register)
	_ = p.loader.Set("$ui", uiObj)

	// 7. Load the plugin source code in the VM (nothing will execute)
	_, err = p.loader.RunString(source)
	if err != nil {
		logger.Error().Err(err).Msg("extensions: Failed to load plugin")
		return nil, nil, err
	}

	// 8. Get and call the init function to actually run the plugin
	if initFunc := p.loader.Get("init"); initFunc != nil && initFunc != goja.Undefined() {
		_, err = p.loader.RunString("init();")
		if err != nil {
			logger.Error().Err(err).Msg("extensions: Failed to run plugin")
			return nil, nil, fmt.Errorf("failed to run plugin: %w", err)
		}
		logger.Debug().Msg("extensions: Plugin initialized")
	}

	return p, p, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// BindPluginAPIs adds plugin-specific APIs
func (p *GojaPlugin) BindPluginAPIs(vm *goja.Runtime, logger *zerolog.Logger) {
	// Bind the app context
	//_ = vm.Set("$ctx", hook.GlobalHookManager.AppContext())

	fm := FieldMapper{}
	vm.SetFieldNameMapper(fm)

	// Bind the store
	p.store.Bind(vm, p.scheduler)
	// Bind mutable bindings
	goja_util.BindMutable(vm)
	// Bind await bindings
	goja_util.BindAwait(vm)
	// Bind console bindings
	_ = goja_bindings.BindConsoleWithWS(p.ext, vm, logger, p.wsEventManager)

	// Bind the app context
	plugin.GlobalAppContext.BindApp(vm, logger, p.ext)

	// Bind permission-specific APIs
	if p.ext.Plugin != nil {
		for _, permission := range p.ext.Plugin.Permissions.Scopes {
			switch permission {
			case extension.PluginPermissionStorage: // Storage
				p.storage = plugin.GlobalAppContext.BindStorage(vm, logger, p.ext, p.scheduler)

			case extension.PluginPermissionAnilist: // Anilist
				plugin.GlobalAppContext.BindAnilist(vm, logger, p.ext)

			case extension.PluginPermissionDatabase: // Database
				plugin.GlobalAppContext.BindDatabase(vm, logger, p.ext)

			case extension.PluginPermissionSystem: // System
				plugin.GlobalAppContext.BindSystem(vm, logger, p.ext, p.scheduler)
			}
		}
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// bindHooks sets up hooks for the Goja runtime
func (p *GojaPlugin) bindHooks() {
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
	appObj := p.loader.NewObject()

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
			unbindHookFunc := hookInstance.MethodByName("Unbind")

			// Get the expected handler type for the hook
			// i.e. func(e *hook_resolver.Resolver) error
			handlerType := hookBindFunc.Type().In(0)

			// Create a new handler function for the hook
			// - returns a new handler of the given handlerType that wraps the function
			handler := reflect.MakeFunc(handlerType, func(args []reflect.Value) (results []reflect.Value) {
				// Prepare arguments for the handler
				handlerArgs := make([]any, len(args))

				// var err error
				// if p.interrupted {
				// 	return []reflect.Value{reflect.ValueOf(&err).Elem()}
				// }

				// Run the handler in an isolated "executor" runtime for concurrency
				err := p.runtimeManager.Run(context.Background(), p.ext.ID, func(executor *goja.Runtime) error {
					// Set the field name mapper for the executor
					executor.SetFieldNameMapper(fm)
					// Convert each argument (event property) to the appropriate type
					for i, arg := range args {
						handlerArgs[i] = arg.Interface()
					}
					// Set the global variable $ctx in the executor
					// executor.Set("$$app", plugin.GlobalAppContext)
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

			// Bind the hook if the plugin is not interrupted
			if p.interrupted {
				return
			}

			// Register the wrapped hook handler
			callRet := hookBindFunc.Call([]reflect.Value{handler})
			// Get the ID from the return value
			id, ok := callRet[0].Interface().(string)
			if ok {
				p.unbindHookFuncs = append(p.unbindHookFuncs, func() {
					p.logger.Trace().Str("id", p.ext.ID).Msgf("plugin: Unbinding hook %s", id)
					unbindHookFunc.Call([]reflect.Value{reflect.ValueOf(id)})
				})
			}
		})
	}

	// Set the $app object in the loader for JavaScript access
	p.loader.Set("$app", appObj)
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
