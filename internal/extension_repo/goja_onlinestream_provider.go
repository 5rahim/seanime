package extension_repo

import (
	"fmt"
	"github.com/dop251/goja"
	"github.com/rs/zerolog"
	"seanime/internal/extension"
	"seanime/internal/util"

	hibikeonlinestream "github.com/5rahim/hibike/pkg/extension/onlinestream"
)

type (
	GojaOnlinestreamProvider struct {
		gojaExtensionImpl
	}
)

func NewGojaOnlinestreamProvider(ext *extension.Extension, language extension.Language, logger *zerolog.Logger) (hibikeonlinestream.Provider, *GojaOnlinestreamProvider, error) {
	logger.Trace().Str("id", ext.ID).Any("language", language).Msg("extensions: Loading external online streaming provider")

	vm, err := SetupGojaExtensionVM(ext, language, logger)
	if err != nil {
		logger.Error().Err(err).Str("id", ext.ID).Msg("extensions: Failed to create javascript VM")
		return nil, nil, err
	}

	// Create the provider
	_, err = vm.RunString(`function NewProvider() {
   return new Provider()
}`)
	if err != nil {
		vm.ClearInterrupt()
		logger.Error().Err(err).Str("id", ext.ID).Msg("extensions: Failed to create online streaming provider")
		return nil, nil, err
	}

	newProviderFunc, ok := goja.AssertFunction(vm.Get("NewProvider"))
	if !ok {
		vm.ClearInterrupt()
		logger.Error().Str("id", ext.ID).Msg("extensions: Failed to invoke online streaming provider constructor")
		return nil, nil, fmt.Errorf("failed to invoke online streaming provider constructor")
	}

	classObjVal, err := newProviderFunc(goja.Undefined())
	if err != nil {
		vm.ClearInterrupt()
		logger.Error().Err(err).Str("id", ext.ID).Msg("extensions: Failed to create online streaming provider")
		return nil, nil, err
	}

	classObj := classObjVal.ToObject(vm)

	ret := &GojaOnlinestreamProvider{
		gojaExtensionImpl: gojaExtensionImpl{
			vm:       vm,
			logger:   logger,
			ext:      ext,
			classObj: classObj,
		},
	}
	return ret, ret, nil
}

func (g *GojaOnlinestreamProvider) GetVM() *goja.Runtime {
	return g.vm
}

func (g *GojaOnlinestreamProvider) GetEpisodeServers() (ret []string) {
	ret = make([]string, 0)

	method, err := g.callClassMethod("getEpisodeServers")

	promiseRes, err := g.waitForPromise(method)
	if err != nil {
		return
	}

	err = g.unmarshalValue(promiseRes, &ret)
	if err != nil {
		return
	}

	return
}

func (g *GojaOnlinestreamProvider) Search(query string, dub bool) (ret []*hibikeonlinestream.SearchResult, err error) {
	defer util.HandlePanicInModuleWithError(g.ext.ID, &err)

	method, err := g.callClassMethod("search", g.vm.ToValue(query), g.vm.ToValue(dub))

	promiseRes, err := g.waitForPromise(method)
	if err != nil {
		return nil, err
	}

	err = g.unmarshalValue(promiseRes, &ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (g *GojaOnlinestreamProvider) FindEpisode(id string) (ret []*hibikeonlinestream.EpisodeDetails, err error) {
	defer util.HandlePanicInModuleWithError(g.ext.ID, &err)

	method, err := g.callClassMethod("findEpisode", g.vm.ToValue(id))

	promiseRes, err := g.waitForPromise(method)
	if err != nil {
		return nil, err
	}

	err = g.unmarshalValue(promiseRes, &ret)
	if err != nil {
		return nil, err
	}

	return
}

func (g *GojaOnlinestreamProvider) FindEpisodeServer(episode *hibikeonlinestream.EpisodeDetails, server string) (ret *hibikeonlinestream.EpisodeServer, err error) {
	defer util.HandlePanicInModuleWithError(g.ext.ID, &err)

	method, err := g.callClassMethod("findEpisodeServer", g.vm.ToValue(structToMap(episode)), g.vm.ToValue(server))

	promiseRes, err := g.waitForPromise(method)
	if err != nil {
		return nil, err
	}

	err = g.unmarshalValue(promiseRes, &ret)
	if err != nil {
		return nil, err
	}

	return
}
