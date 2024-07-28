package extension_repo

import (
	"fmt"
	"github.com/dop251/goja"
	"github.com/rs/zerolog"
	"seanime/internal/extension"
	"seanime/internal/util"
	"seanime/internal/util/comparison"

	hibikemanga "github.com/5rahim/hibike/pkg/extension/manga"
)

type (
	GojaMangaProvider struct {
		gojaExtensionImpl
	}
)

func NewGojaMangaProvider(ext *extension.Extension, language extension.Language, logger *zerolog.Logger) (hibikemanga.Provider, *GojaMangaProvider, error) {
	logger.Trace().Str("id", ext.ID).Any("language", language).Msg("extensions: Loading external manga provider")

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
		logger.Error().Err(err).Str("id", ext.ID).Msg("extensions: Failed to create manga provider")
		return nil, nil, err
	}

	newProviderFunc, ok := goja.AssertFunction(vm.Get("NewProvider"))
	if !ok {
		vm.ClearInterrupt()
		logger.Error().Str("id", ext.ID).Msg("extensions: Failed to invoke manga provider constructor")
		return nil, nil, fmt.Errorf("failed to invoke manga provider constructor")
	}

	classObjVal, err := newProviderFunc(goja.Undefined())
	if err != nil {
		vm.ClearInterrupt()
		logger.Error().Err(err).Str("id", ext.ID).Msg("extensions: Failed to create manga provider")
		return nil, nil, err
	}

	classObj := classObjVal.ToObject(vm)

	ret := &GojaMangaProvider{
		gojaExtensionImpl: gojaExtensionImpl{
			vm:       vm,
			logger:   logger,
			ext:      ext,
			classObj: classObj,
		},
	}
	return ret, ret, nil
}

func (g *GojaMangaProvider) GetVM() *goja.Runtime {
	return g.vm
}

func (g *GojaMangaProvider) Search(query hibikemanga.SearchOptions) (ret []*hibikemanga.SearchResult, err error) {
	defer util.HandlePanicInModuleWithError(g.ext.ID, &err)

	method, err := g.callClassMethod("search", g.vm.ToValue(structToMap(query)))

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

		compRes, ok := comparison.FindBestMatchWithSorensenDice(&query.Query, compTitles)
		if ok {
			ret[i].SearchRating = compRes.Rating
		}
	}

	return ret, nil
}

func (g *GojaMangaProvider) FindChapters(id string) (ret []*hibikemanga.ChapterDetails, err error) {
	defer util.HandlePanicInModuleWithError(g.ext.ID, &err)

	method, err := g.callClassMethod("findChapters", g.vm.ToValue(id))

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
	defer util.HandlePanicInModuleWithError(g.ext.ID, &err)

	method, err := g.callClassMethod("findChapterPages", g.vm.ToValue(id))

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
