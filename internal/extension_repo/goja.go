package extension_repo

import (
	"errors"
	"fmt"
	"github.com/dop251/goja"
	"github.com/dop251/goja/parser"
	gojaconsole "github.com/dop251/goja_nodejs/console"
	gojarequire "github.com/dop251/goja_nodejs/require"
	gojaurl "github.com/dop251/goja_nodejs/url"
	"github.com/evanw/esbuild/pkg/api"
	"time"
)

// CreateJSVM creates a new Javascript VM with the necessary bindings
func CreateJSVM() (*goja.Runtime, error) {

	vm := goja.New()
	vm.SetParserOptions(parser.WithDisableSourceMaps)

	registry := new(gojarequire.Registry)
	registry.Enable(vm)

	gojaurl.Enable(vm)
	gojaconsole.Enable(vm)

	err := gojaBindFetch(vm)
	if err != nil {
		return nil, err
	}

	err = gojaBindFindBestMatchWithSorensenDice(vm)
	if err != nil {
		return nil, err
	}

	return vm, nil
}

// JSVMTypescriptToJS converts Typescript to Javascript
func JSVMTypescriptToJS(ts string) (string, error) {
	scriptJSTransform := api.Transform(ts, api.TransformOptions{
		Target: api.ESNext,
		Loader: api.LoaderTS,
		Format: api.FormatDefault,
	})

	if scriptJSTransform.Errors != nil && len(scriptJSTransform.Errors) > 0 {
		return "", errors.New(scriptJSTransform.Errors[0].Text)
	}

	return string(scriptJSTransform.Code), nil
}

func gojaWaitForPromise(vm *goja.Runtime, value goja.Value) (goja.Value, error) {
	promise, ok := value.Export().(*goja.Promise)
	if !ok {
		return nil, errors.New("value is not a promise")
	}

	doneCh := make(chan struct{})

	go func() {
		for promise.State() == goja.PromiseStatePending {
			time.Sleep(10 * time.Millisecond)
		}
		close(doneCh)
	}()

	<-doneCh

	if promise.State() == goja.PromiseStateRejected {
		err := promise.Result()
		return nil, fmt.Errorf("promise rejected: %v", err)
	}

	return promise.Result(), nil
}
