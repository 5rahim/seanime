package extension_repo

import (
	"fmt"
	hibikemanga "github.com/5rahim/hibike/pkg/extension/manga"
	"github.com/dop251/goja"
	"github.com/rs/zerolog"
	"seanime/internal/extension"
	"seanime/internal/util"
)

type (
	GojaMangaProvider struct {
		ext         *extension.Extension
		vm          *goja.Runtime
		logger      *zerolog.Logger
		providerObj *goja.Object
	}
)

func NewGojaMangaProvider(ext *extension.Extension, logger *zerolog.Logger) (hibikemanga.Provider, error) {
	logger.Trace().Str("id", ext.ID).Msgf("extensions: Creating javascript VM for external manga provider")

	vm, err := CreateJSVM()
	if err != nil {
		logger.Error().Err(err).Str("id", ext.ID).Msg("extensions: Failed to create javascript VM")
		return nil, err
	}

	source, err := JSVMTypescriptToJS(ext.Payload)
	if err != nil {
		logger.Error().Err(err).Str("id", ext.ID).Msg("extensions: Failed to convert typescript to javascript")
		return nil, err
	}

	// Run the program on the VM
	_, err = vm.RunString(source)
	if err != nil {
		logger.Error().Err(err).Str("id", ext.ID).Msg("extensions: Failed to run javascript source")
		return nil, err
	}

	// Create the provider
	_, err = vm.RunString(`function NewProvider() {
    return new Provider()
}`)
	if err != nil {
		logger.Error().Err(err).Str("id", ext.ID).Msg("extensions: Failed to create manga provider")
		return nil, err
	}

	newProviderFunc, ok := goja.AssertFunction(vm.Get("NewProvider"))
	if !ok {
		logger.Error().Str("id", ext.ID).Msg("extensions: Failed to invoke manga provider constructor")
		return nil, fmt.Errorf("failed to invoke manga provider constructor")
	}

	providerObjVal, err := newProviderFunc(goja.Undefined())
	if err != nil {
		logger.Error().Err(err).Str("id", ext.ID).Msg("extensions: Failed to create manga provider")
		return nil, err
	}

	providerObj := providerObjVal.ToObject(vm)

	return &GojaMangaProvider{
		vm:          vm,
		logger:      logger,
		ext:         ext,
		providerObj: providerObj,
	}, nil
}

func (g *GojaMangaProvider) Search(query hibikemanga.SearchOptions) (ret []*hibikemanga.SearchResult, err error) {
	defer util.HandlePanicInModuleWithError(g.ext.ID, &err)

	searchFunc, ok := goja.AssertFunction(g.providerObj.Get("search"))
	if !ok {
		g.logger.Error().Str("id", g.ext.ID).Msg("extensions: Failed to get search function")
		return nil, fmt.Errorf("failed to get search function")
	}

	// Call the search function
	searchResult, err := searchFunc(g.providerObj, g.vm.ToValue(query.Query))
	if err != nil {
		g.logger.Error().Err(err).Str("id", g.ext.ID).Msg("extensions: Failed to call search function")
		return nil, err
	}

	promiseRes, err := gojaWaitForPromise(g.vm, searchResult)
	if err != nil {
		g.logger.Error().Err(err).Str("id", g.ext.ID).Msg("extensions: Unexpected response")
		return nil, err
	}

	if promiseRes == nil || goja.IsUndefined(promiseRes) {
		g.logger.Error().Str("id", g.ext.ID).Msg("extensions: Result is undefined")
		return nil, fmt.Errorf("result is undefined")
	}

	searchResArr, ok := promiseRes.Export().([]interface{})
	if !ok {
		g.logger.Error().Str("id", g.ext.ID).Msg("extensions: Failed to cast return value")
		return nil, fmt.Errorf("failed to cast return value")
	}

	// Convert the results
	for _, objMap := range searchResArr {
		obj, ok := objMap.(map[string]interface{})
		if !ok {
			g.logger.Error().Str("id", g.ext.ID).Msg("extensions: Failed to cast results from extension")
			return nil, fmt.Errorf("failed to cast results from extension")
		}

		searchRes := &hibikemanga.SearchResult{
			Provider: g.ext.ID,
		}

		searchRes.ID = obj["id"].(string)
		//searchRes.Provider = obj["provider"].(string)
		searchRes.Title = obj["title"].(string)
		searchRes.Image = obj["image"].(string)

		_year, ok := obj["year"].(int64)
		if ok {
			searchRes.Year = int(_year)
		}

		_rating, ok := obj["searchRating"].(interface{})
		if ok {
			searchRatingFloat, ok := _rating.(float64)
			if ok {
				searchRes.SearchRating = searchRatingFloat
			} else {
				searchRatingInt, ok := _rating.(int64)
				if ok {
					searchRes.SearchRating = float64(searchRatingInt)
				}
			}
		}

		_synonyms, ok := obj["synonyms"].([]interface{})
		if ok {
			for _, syn := range _synonyms {
				searchRes.Synonyms = append(searchRes.Synonyms, syn.(string))
			}
		}

		ret = append(ret, searchRes)
	}

	return ret, nil
}

func (g *GojaMangaProvider) FindChapters(id string) (ret []*hibikemanga.ChapterDetails, err error) {
	defer util.HandlePanicInModuleWithError(g.ext.ID, &err)

	searchFunc, ok := goja.AssertFunction(g.providerObj.Get("findChapters"))
	if !ok {
		g.logger.Error().Str("id", g.ext.ID).Msg("extensions: Failed to get search function")
		return nil, fmt.Errorf("failed to get search function")
	}

	// Call the findChapters function
	findChaptersRes, err := searchFunc(g.providerObj, g.vm.ToValue(id))
	if err != nil {
		g.logger.Error().Err(err).Str("id", g.ext.ID).Msg("extensions: Failed to call search function")
		return nil, err
	}

	promiseRes, err := gojaWaitForPromise(g.vm, findChaptersRes)
	if err != nil {
		g.logger.Error().Err(err).Str("id", g.ext.ID).Msg("extensions: Unexpected response")
		return nil, err
	}

	if promiseRes == nil || goja.IsUndefined(promiseRes) {
		g.logger.Error().Str("id", g.ext.ID).Msg("extensions: Result is undefined")
		return nil, fmt.Errorf("result is undefined")
	}

	findChaptersResArr, ok := promiseRes.Export().([]interface{})
	if !ok {
		g.logger.Error().Str("id", g.ext.ID).Msg("extensions: Failed to cast return value")
		return nil, fmt.Errorf("failed to cast return value")
	}

	// Convert the results
	for _, objMap := range findChaptersResArr {
		obj, ok := objMap.(map[string]interface{})
		if !ok {
			g.logger.Error().Str("id", g.ext.ID).Msg("extensions: Failed to cast results from extension")
			return nil, fmt.Errorf("failed to cast results from extension")
		}

		chapter := &hibikemanga.ChapterDetails{
			Provider: g.ext.ID,
		}

		chapter.ID = obj["id"].(string)
		chapter.URL = obj["url"].(string)
		//chapter.Provider = obj["provider"].(string)
		chapter.Title = obj["title"].(string)
		chapter.Chapter = obj["chapter"].(string)
		chapter.Index = uint(obj["index"].(int64))

		_rating, ok := obj["rating"].(int64)
		if ok {
			chapter.Rating = int(_rating)
		}

		_updateAt, ok := obj["updatedAt"].(string)
		if ok {
			chapter.UpdatedAt = _updateAt
		}

		ret = append(ret, chapter)
	}

	return ret, nil
}

func (g *GojaMangaProvider) FindChapterPages(id string) (ret []*hibikemanga.ChapterPage, err error) {
	defer util.HandlePanicInModuleWithError(g.ext.ID, &err)

	searchFunc, ok := goja.AssertFunction(g.providerObj.Get("findChapterPages"))
	if !ok {
		g.logger.Error().Str("id", g.ext.ID).Msg("extensions: Failed to get search function")
		return nil, fmt.Errorf("failed to get search function")
	}

	// Call the findChapterPages function
	findChapterPagesRes, err := searchFunc(g.providerObj, g.vm.ToValue(id))
	if err != nil {
		g.logger.Error().Err(err).Str("id", g.ext.ID).Msg("extensions: Failed to call search function")
		return nil, err
	}

	promiseRes, err := gojaWaitForPromise(g.vm, findChapterPagesRes)
	if err != nil {
		g.logger.Error().Err(err).Str("id", g.ext.ID).Msg("extensions: Unexpected response")
		return nil, err
	}

	if promiseRes == nil || goja.IsUndefined(promiseRes) {
		g.logger.Error().Str("id", g.ext.ID).Msg("extensions: Result is undefined")
		return nil, fmt.Errorf("result is undefined")
	}

	findChapterPagesResArr, ok := promiseRes.Export().([]interface{})
	if !ok {
		g.logger.Error().Str("id", g.ext.ID).Msg("extensions: Failed to cast return value")
		return nil, fmt.Errorf("failed to cast return value")
	}

	// Convert the results
	for _, objMap := range findChapterPagesResArr {
		obj, ok := objMap.(map[string]interface{})
		if !ok {
			g.logger.Error().Str("id", g.ext.ID).Msg("extensions: Failed to cast results from extension")
			return nil, fmt.Errorf("failed to cast results from extension")
		}

		chapterPage := &hibikemanga.ChapterPage{
			Provider: g.ext.ID,
			Headers:  make(map[string]string),
		}

		chapterPage.URL = obj["url"].(string)
		//chapterPage.Provider = obj["provider"].(string)
		chapterPage.Index = int(obj["index"].(int64))

		_headers, ok := obj["headers"].(map[string]interface{})
		if ok {
			for key, value := range _headers {
				chapterPage.Headers[key] = value.(string)
			}
		}

		ret = append(ret, chapterPage)
	}

	return ret, nil
}
