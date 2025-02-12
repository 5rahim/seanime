package extension_repo

import (
	"context"
	"fmt"
	"seanime/internal/extension"
	hibikeonlinestream "seanime/internal/extension/hibike/onlinestream"
	"seanime/internal/goja/goja_runtime"
	"seanime/internal/util"

	"github.com/rs/zerolog"
)

type GojaOnlinestreamProvider struct {
	*gojaProviderBase
}

func NewGojaOnlinestreamProvider(ext *extension.Extension, language extension.Language, logger *zerolog.Logger, runtimeManager *goja_runtime.Manager) (hibikeonlinestream.Provider, *GojaOnlinestreamProvider, error) {
	base, err := initializeProviderBase(ext, language, logger, runtimeManager)
	if err != nil {
		return nil, nil, err
	}

	provider := &GojaOnlinestreamProvider{
		gojaProviderBase: base,
	}
	return provider, provider, nil
}

func (g *GojaOnlinestreamProvider) GetEpisodeServers() (ret []string) {
	ret = make([]string, 0)

	method, err := g.callClassMethod(context.Background(), "getEpisodeServers")

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

func (g *GojaOnlinestreamProvider) Search(opts hibikeonlinestream.SearchOptions) (ret []*hibikeonlinestream.SearchResult, err error) {
	defer util.HandlePanicInModuleWithError(g.ext.ID+".Search", &err)

	method, err := g.callClassMethod(context.Background(), "search", structToMap(opts))
	if err != nil {
		return nil, fmt.Errorf("failed to call search method: %w", err)
	}

	promiseRes, err := g.waitForPromise(method)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for promise: %w", err)
	}

	ret = make([]*hibikeonlinestream.SearchResult, 0)
	err = g.unmarshalValue(promiseRes, &ret)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal search results: %w", err)
	}

	return ret, nil
}

func (g *GojaOnlinestreamProvider) FindEpisodes(id string) (ret []*hibikeonlinestream.EpisodeDetails, err error) {
	defer util.HandlePanicInModuleWithError(g.ext.ID+".FindEpisodes", &err)

	method, err := g.callClassMethod(context.Background(), "findEpisodes", id)

	promiseRes, err := g.waitForPromise(method)
	if err != nil {
		return nil, err
	}

	err = g.unmarshalValue(promiseRes, &ret)
	if err != nil {
		return nil, err
	}

	for _, episode := range ret {
		episode.Provider = g.ext.ID
	}

	return
}

func (g *GojaOnlinestreamProvider) FindEpisodeServer(episode *hibikeonlinestream.EpisodeDetails, server string) (ret *hibikeonlinestream.EpisodeServer, err error) {
	defer util.HandlePanicInModuleWithError(g.ext.ID+".FindEpisodeServer", &err)

	method, err := g.callClassMethod(context.Background(), "findEpisodeServer", structToMap(episode), server)

	promiseRes, err := g.waitForPromise(method)
	if err != nil {
		return nil, err
	}

	err = g.unmarshalValue(promiseRes, &ret)
	if err != nil {
		return nil, err
	}

	ret.Provider = g.ext.ID

	return
}

func (g *GojaOnlinestreamProvider) GetSettings() (ret hibikeonlinestream.Settings) {
	defer util.HandlePanicInModuleThen(g.ext.ID+".GetSettings", func() {
		ret = hibikeonlinestream.Settings{}
	})

	method, err := g.callClassMethod(context.Background(), "getSettings")
	if err != nil {
		return
	}

	err = g.unmarshalValue(method, &ret)
	if err != nil {
		return
	}

	return
}
