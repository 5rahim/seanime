package extension_repo

import (
	"fmt"
	"seanime/internal/extension"

	"github.com/dop251/goja"
	"github.com/rs/zerolog"
)

type GojaPlugin struct {
	gojaExtensionImpl
}

func NewGojaPlugin(ext *extension.Extension, language extension.Language, logger *zerolog.Logger) (*GojaPlugin, error) {
	logger.Trace().Str("id", ext.ID).Any("language", language).Msg("extensions: Loading plugin")

	vm, err := SetupGojaExtensionVM(ext, language, logger)
	if err != nil {
		logger.Error().Err(err).Str("id", ext.ID).Msg("extensions: Failed to create javascript VM")
		return nil, err
	}

	_, err = vm.RunString(`function NewPlugin() {	
		return new Plugin()
	}`)
	if err != nil {
		vm.ClearInterrupt()
		logger.Error().Err(err).Str("id", ext.ID).Msg("extensions: Failed to create plugin")
		return nil, err
	}

	newPluginFunc, ok := goja.AssertFunction(vm.Get("NewPlugin"))
	if !ok {
		vm.ClearInterrupt()
		logger.Error().Str("id", ext.ID).Msg("extensions: Failed to invoke plugin constructor")
		return nil, fmt.Errorf("failed to invoke plugin constructor")
	}

	classObjVal, err := newPluginFunc(goja.Undefined())
	if err != nil {
		vm.ClearInterrupt()
		logger.Error().Err(err).Str("id", ext.ID).Msg("extensions: Failed to create plugin")
		return nil, err
	}

	classObj := classObjVal.ToObject(vm)

	ret := &GojaPlugin{
		gojaExtensionImpl: gojaExtensionImpl{
			vm:       vm,
			classObj: classObj,
		},
	}

	return ret, nil
}

func (p *GojaPlugin) GetVM() *goja.Runtime {
	return p.vm
}
