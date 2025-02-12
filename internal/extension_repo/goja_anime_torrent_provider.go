package extension_repo

import (
	"context"
	"seanime/internal/extension"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/goja/goja_runtime"
	"seanime/internal/util"

	"github.com/rs/zerolog"
)

type GojaAnimeTorrentProvider struct {
	*gojaProviderBase
}

func NewGojaAnimeTorrentProvider(ext *extension.Extension, language extension.Language, logger *zerolog.Logger, runtimeManager *goja_runtime.Manager) (hibiketorrent.AnimeProvider, *GojaAnimeTorrentProvider, error) {
	base, err := initializeProviderBase(ext, language, logger, runtimeManager)
	if err != nil {
		return nil, nil, err
	}

	provider := &GojaAnimeTorrentProvider{
		gojaProviderBase: base,
	}
	return provider, provider, nil
}

func (g *GojaAnimeTorrentProvider) Search(opts hibiketorrent.AnimeSearchOptions) (ret []*hibiketorrent.AnimeTorrent, err error) {
	defer util.HandlePanicInModuleWithError(g.ext.ID+".Search", &err)

	method, err := g.callClassMethod(context.Background(), "search", structToMap(opts))

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
	defer util.HandlePanicInModuleWithError(g.ext.ID+".SmartSearch", &err)

	method, err := g.callClassMethod(context.Background(), "smartSearch", structToMap(opts))

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
	defer util.HandlePanicInModuleWithError(g.ext.ID+".GetTorrentInfoHash", &err)

	res, err := g.callClassMethod(context.Background(), "getTorrentInfoHash", structToMap(torrent))
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
	defer util.HandlePanicInModuleWithError(g.ext.ID+".GetTorrentMagnetLink", &err)

	res, err := g.callClassMethod(context.Background(), "getTorrentMagnetLink", structToMap(torrent))
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
	defer util.HandlePanicInModuleWithError(g.ext.ID+".GetLatest", &err)

	method, err := g.callClassMethod(context.Background(), "getLatest")
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
	defer util.HandlePanicInModuleThen(g.ext.ID+".GetSettings", func() {
		ret = hibiketorrent.AnimeProviderSettings{}
	})

	res, err := g.callClassMethod(context.Background(), "getSettings")
	if err != nil {
		return
	}

	err = g.unmarshalValue(res, &ret)
	if err != nil {
		return
	}

	return
}
