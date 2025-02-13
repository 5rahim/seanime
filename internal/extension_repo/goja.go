package extension_repo

import (
	"encoding/json"
	"fmt"
	"seanime/internal/extension"
	goja_bindings "seanime/internal/goja/goja_bindings"

	"github.com/dop251/goja"
	"github.com/dop251/goja/parser"
	gojabuffer "github.com/dop251/goja_nodejs/buffer"
	gojarequire "github.com/dop251/goja_nodejs/require"
	gojaurl "github.com/dop251/goja_nodejs/url"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/rs/zerolog"
)

// GojaExtension is stored in the repository extension map, giving access to the VMs.
// Current use: Kill the VM when the extension is unloaded.
type GojaExtension interface {
	GetVM() *goja.Runtime
	PutVM(*goja.Runtime)
}

// SetupGojaExtensionVM creates a new JavaScript VM with the extension source code loaded
func SetupGojaExtensionVM(ext *extension.Extension, language extension.Language, logger *zerolog.Logger) (func() (*goja.Runtime, error), error) {
	logger.Trace().Str("id", ext.ID).Any("language", language).Msgf("extensions: Creating javascript VM")

	source := ext.Payload
	if language == extension.LanguageTypescript {
		var err error
		source, err = JSVMTypescriptToJS(ext.Payload)
		if err != nil {
			logger.Error().Err(err).Str("id", ext.ID).Msg("extensions: Failed to convert typescript")
			return nil, err
		}
	}

	// Compile the program once, to be reused by all VMs
	program, err := goja.Compile("", source, false)
	if err != nil {
		return nil, fmt.Errorf("compilation failed: %w", err)
	}

	return func() (*goja.Runtime, error) {
		vm := goja.New()
		vm.SetParserOptions(parser.WithDisableSourceMaps)

		ShareBinds(vm, logger)

		_, err := vm.RunProgram(program)
		if err != nil {
			return nil, fmt.Errorf("failed to run program: %w", err)
		}

		return vm, nil
	}, nil
}

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
}

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
