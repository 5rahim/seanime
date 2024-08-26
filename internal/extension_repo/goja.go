package extension_repo

import (
	"errors"
	"github.com/dop251/goja"
	"github.com/dop251/goja/parser"
	gojabuffer "github.com/dop251/goja_nodejs/buffer"
	gojarequire "github.com/dop251/goja_nodejs/require"
	gojaurl "github.com/dop251/goja_nodejs/url"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/rs/zerolog"
	"seanime/internal/extension"
	"seanime/internal/extension_repo/goja_bindings"
)

// GojaExtension is stored in the repository extension map, giving access to the VMs.
// Current use: Kill the VM when the extension is unloaded.
type GojaExtension interface {
	GetVM() *goja.Runtime
}

// SetupGojaExtensionVM creates a new JavaScript VM with the extension source code loaded
func SetupGojaExtensionVM(ext *extension.Extension, language extension.Language, logger *zerolog.Logger) (*goja.Runtime, error) {
	logger.Trace().Str("id", ext.ID).Any("language", language).Msgf("extensions: Creating javascript VM for external manga provider")

	vm, err := CreateJSVM(logger)
	if err != nil {
		logger.Error().Err(err).Str("id", ext.ID).Msg("extensions: Failed to create javascript VM")
		return nil, err
	}

	source := ext.Payload

	if language == extension.LanguageTypescript {
		source, err = JSVMTypescriptToJS(ext.Payload)
		if err != nil {
			logger.Error().Err(err).Str("id", ext.ID).Msg("extensions: Failed to convert typescript to javascript")
			return nil, err
		}
	}

	// Run the program on the VM
	_, err = vm.RunString(source)
	if err != nil {
		logger.Error().Err(err).Str("id", ext.ID).Msg("extensions: Failed to run javascript code")
		return nil, err
	}

	return vm, nil
}

// CreateJSVM creates a new JavaScript VM for SetupGojaExtensionVM
func CreateJSVM(logger *zerolog.Logger) (*goja.Runtime, error) {

	vm := goja.New()
	vm.SetParserOptions(parser.WithDisableSourceMaps)

	registry := new(gojarequire.Registry)
	registry.Enable(vm)

	gojaurl.Enable(vm)
	gojabuffer.Enable(vm)
	err := goja_bindings.BindFetch(vm)
	if err != nil {
		return nil, err
	}
	err = goja_bindings.BindConsole(vm, logger)
	if err != nil {
		return nil, err
	}
	err = goja_bindings.BindFormData(vm)
	if err != nil {
		return nil, err
	}
	err = goja_bindings.BindDocument(vm)
	if err != nil {
		return nil, err
	}
	err = goja_bindings.BindCrypto(vm)
	if err != nil {
		return nil, err
	}

	return vm, nil
}

func JSVMTypescriptToJS(ts string) (string, error) {
	scriptJSTransform := api.Transform(ts, api.TransformOptions{
		Target: api.ES2018,
		Loader: api.LoaderTS,
		Format: api.FormatDefault,
	})

	if scriptJSTransform.Errors != nil && len(scriptJSTransform.Errors) > 0 {
		return "", errors.New(scriptJSTransform.Errors[0].Text)
	}

	return string(scriptJSTransform.Code), nil
}
