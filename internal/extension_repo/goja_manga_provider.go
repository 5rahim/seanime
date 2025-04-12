package extension_repo

import (
	"context"
	"seanime/internal/extension"
	hibikemanga "seanime/internal/extension/hibike/manga"
	"seanime/internal/goja/goja_runtime"
	"seanime/internal/util"
	"seanime/internal/util/comparison"

	"github.com/rs/zerolog"
)

type GojaMangaProvider struct {
	*gojaProviderBase
}

func NewGojaMangaProvider(ext *extension.Extension, language extension.Language, logger *zerolog.Logger, runtimeManager *goja_runtime.Manager) (hibikemanga.Provider, *GojaMangaProvider, error) {
	base, err := initializeProviderBase(ext, language, logger, runtimeManager)
	if err != nil {
		return nil, nil, err
	}

	provider := &GojaMangaProvider{
		gojaProviderBase: base,
	}
	return provider, provider, nil
}

func (g *GojaMangaProvider) GetSettings() (ret hibikemanga.Settings) {
	defer util.HandlePanicInModuleThen(g.ext.ID+".GetSettings", func() {
		ret = hibikemanga.Settings{}
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

func (g *GojaMangaProvider) Search(opts hibikemanga.SearchOptions) (ret []*hibikemanga.SearchResult, err error) {
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

	// Set the provider & search rating
	for i := range ret {
		ret[i].Provider = g.ext.ID

		synonyms := ret[i].Synonyms
		if synonyms == nil {
			continue
		}

		compTitles := []*string{&ret[i].Title}
		for _, syn := range synonyms {
			compTitles = append(compTitles, &syn)
		}

		compRes, ok := comparison.FindBestMatchWithSorensenDice(&opts.Query, compTitles)
		if ok {
			ret[i].SearchRating = compRes.Rating
		}
	}

	return ret, nil
}

func (g *GojaMangaProvider) FindChapters(id string) (ret []*hibikemanga.ChapterDetails, err error) {
	defer util.HandlePanicInModuleWithError(g.ext.ID+".FindChapters", &err)

	method, err := g.callClassMethod(context.Background(), "findChapters", id)

	promiseRes, err := g.waitForPromise(method)
	if err != nil {
		return nil, err
	}

	err = g.unmarshalValue(promiseRes, &ret)
	if err != nil {
		return nil, err
	}

	// Set the provider
	for i := range ret {
		ret[i].Provider = g.ext.ID
	}

	return ret, nil
}

func (g *GojaMangaProvider) FindChapterPages(id string) (ret []*hibikemanga.ChapterPage, err error) {
	defer util.HandlePanicInModuleWithError(g.ext.ID+".FindChapterPages", &err)

	method, err := g.callClassMethod(context.Background(), "findChapterPages", id)

	promiseRes, err := g.waitForPromise(method)
	if err != nil {
		return nil, err
	}

	err = g.unmarshalValue(promiseRes, &ret)
	if err != nil {
		return nil, err
	}

	// Set the provider
	for i := range ret {
		ret[i].Provider = g.ext.ID
	}

	return ret, nil
}
