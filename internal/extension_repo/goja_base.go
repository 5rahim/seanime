package extension_repo

import (
	"context"
	"encoding/json"
	"fmt"
	"seanime/internal/extension"
	"seanime/internal/goja/goja_runtime"

	"github.com/dop251/goja"
	"github.com/rs/zerolog"
)

type gojaProviderBase struct {
	ext            *extension.Extension
	logger         *zerolog.Logger
	pool           *goja_runtime.Pool
	runtimeManager *goja_runtime.Manager
}

func initializeProviderBase(ext *extension.Extension, language extension.Language, logger *zerolog.Logger, runtimeManager *goja_runtime.Manager) (*gojaProviderBase, error) {
	initFn, err := SetupGojaExtensionVM(ext, language, logger)
	if err != nil {
		return nil, err
	}

	pool, err := runtimeManager.GetOrCreatePool(ext.ID, initFn)
	if err != nil {
		return nil, err
	}

	return &gojaProviderBase{
		ext:            ext,
		logger:         logger,
		pool:           pool,
		runtimeManager: runtimeManager,
	}, nil
}

func (g *gojaProviderBase) callClassMethod(ctx context.Context, methodName string, args ...interface{}) (goja.Value, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	vm, err := g.pool.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get VM: %w", err)
	}
	defer g.pool.Put(vm)

	// Create a new instance of the Provider class
	providerInstance, err := vm.RunString("new Provider()")
	if err != nil {
		return nil, fmt.Errorf("failed to create Provider instance: %w", err)
	}
	if providerInstance == nil {
		return nil, fmt.Errorf("Provider constructor returned nil")
	}

	// Get the method from the instance
	method, ok := goja.AssertFunction(providerInstance.ToObject(vm).Get(methodName))
	if !ok {
		return nil, fmt.Errorf("method %s not found or not a function", methodName)
	}

	// Convert arguments to Goja values
	gojaArgs := make([]goja.Value, len(args))
	for i, arg := range args {
		gojaArgs[i] = vm.ToValue(arg)
	}

	// Call the method
	result, err := method(providerInstance, gojaArgs...)
	if err != nil {
		return nil, fmt.Errorf("method %s execution failed: %w", methodName, err)
	}

	g.runtimeManager.PrintMetrics()

	return result, nil
}

func (g *gojaProviderBase) unmarshalValue(value goja.Value, target interface{}) error {
	if value == nil {
		return fmt.Errorf("cannot unmarshal nil value")
	}

	exported := value.Export()
	if exported == nil {
		return fmt.Errorf("exported value is nil")
	}

	data, err := json.Marshal(exported)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	return json.Unmarshal(data, target)
}

func (g *gojaProviderBase) waitForPromise(value goja.Value) (goja.Value, error) {
	if value == nil {
		return nil, fmt.Errorf("cannot wait for nil promise")
	}

	if promise, ok := value.Export().(*goja.Promise); ok {
		result := promise.Result()
		if result == nil {
			return nil, fmt.Errorf("promise result is nil")
		}

		switch promise.State() {
		case goja.PromiseStatePending:
			return nil, fmt.Errorf("promise is still pending")
		case goja.PromiseStateRejected:
			if err, ok := result.Export().(error); ok {
				return nil, fmt.Errorf("promise rejected: %w", err)
			}
			return nil, fmt.Errorf("promise rejected: %v", result)
		case goja.PromiseStateFulfilled:
			return result, nil
		default:
			return nil, fmt.Errorf("unknown promise state: %v", promise.State())
		}
	}
	return value, nil
}

func (g *gojaProviderBase) GetVM() *goja.Runtime {
	vm, _ := g.pool.Get(context.Background())
	return vm
}

func (g *gojaProviderBase) PutVM(vm *goja.Runtime) {
	g.pool.Put(vm)
}
