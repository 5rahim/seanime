package extension_repo

import (
	"fmt"
	"github.com/dop251/goja"
	"github.com/rs/zerolog"
	"seanime/internal/extension"
	"seanime/internal/util"

	hibiketorrent "github.com/5rahim/hibike/pkg/extension/torrent"
)

type (
	GojaAnimeTorrentProvider struct {
		gojaExtensionImpl
	}
)

func NewGojaAnimeTorrentProvider(ext *extension.Extension, language extension.Language, logger *zerolog.Logger) (hibiketorrent.AnimeProvider, *GojaAnimeTorrentProvider, error) {
	logger.Trace().Str("id", ext.ID).Any("language", language).Msg("extensions: Loading external anime torrent provider")

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

	ret := &GojaAnimeTorrentProvider{
		gojaExtensionImpl: gojaExtensionImpl{
			vm:       vm,
			logger:   logger,
			ext:      ext,
			classObj: classObj,
		},
	}
	return ret, ret, nil
}

func (g *GojaAnimeTorrentProvider) GetVM() *goja.Runtime {
	return g.vm
}

func (g *GojaAnimeTorrentProvider) Search(opts hibiketorrent.AnimeSearchOptions) (ret []*hibiketorrent.AnimeTorrent, err error) {
	defer util.HandlePanicInModuleWithError(g.ext.ID, &err)

	method, err := g.callClassMethod("search", g.vm.ToValue(structToMap(opts)))

	promiseRes, err := g.waitForPromise(method)
	if err != nil {
		return nil, err
	}

	err = g.unmarshalValue(promiseRes, &ret)
	if err != nil {
		return nil, err
	}

	for i := range ret {
		ret[i].Provider = g.ext.ID
	}

	return
}
func (g *GojaAnimeTorrentProvider) SmartSearch(opts hibiketorrent.AnimeSmartSearchOptions) (ret []*hibiketorrent.AnimeTorrent, err error) {
	defer util.HandlePanicInModuleWithError(g.ext.ID, &err)

	method, err := g.callClassMethod("smartSearch", g.vm.ToValue(structToMap(opts)))

	promiseRes, err := g.waitForPromise(method)
	if err != nil {
		return nil, err
	}

	err = g.unmarshalValue(promiseRes, &ret)
	if err != nil {
		return nil, err
	}

	for i := range ret {
		ret[i].Provider = g.ext.ID
	}

	return
}
func (g *GojaAnimeTorrentProvider) GetTorrentInfoHash(torrent *hibiketorrent.AnimeTorrent) (ret string, err error) {
	defer util.HandlePanicInModuleWithError(g.ext.ID, &err)

	res, err := g.callClassMethod("getTorrentInfoHash", g.vm.ToValue(structToMap(torrent)))
	if err != nil {
		return "", err
	}

	promiseRes, err := g.waitForPromise(res)
	if err != nil {
		return "", err
	}

	err = g.unmarshalValue(promiseRes, &ret)
	if err != nil {
		return "", err
	}

	return
}
func (g *GojaAnimeTorrentProvider) GetTorrentMagnetLink(torrent *hibiketorrent.AnimeTorrent) (ret string, err error) {
	defer util.HandlePanicInModuleWithError(g.ext.ID, &err)

	res, err := g.callClassMethod("getTorrentMagnetLink", g.vm.ToValue(structToMap(torrent)))
	if err != nil {
		return "", err
	}

	promiseRes, err := g.waitForPromise(res)
	if err != nil {
		return "", err
	}

	err = g.unmarshalValue(promiseRes, &ret)
	if err != nil {
		return "", err
	}

	return

}
func (g *GojaAnimeTorrentProvider) GetLatest() (ret []*hibiketorrent.AnimeTorrent, err error) {
	defer util.HandlePanicInModuleWithError(g.ext.ID, &err)

	method, err := g.callClassMethod("getLatest")
	if err != nil {
		return nil, err
	}

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
func (g *GojaAnimeTorrentProvider) GetSettings() (ret hibiketorrent.AnimeProviderSettings) {

	res, err := g.callClassMethod("getSettings")
	if err != nil {
		return hibiketorrent.AnimeProviderSettings{}
	}

	err = g.unmarshalValue(res, &ret)
	if err != nil {
		return hibiketorrent.AnimeProviderSettings{}
	}

	return
}
