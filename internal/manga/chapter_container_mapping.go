package manga

import (
	"errors"
	"seanime/internal/extension"
	"seanime/internal/util"
	"seanime/internal/util/result"
	"strings"

	hibikemanga "seanime/internal/extension/hibike/manga"
)

var searchResultCache = result.NewCache[string, []*hibikemanga.SearchResult]()

func (r *Repository) ManualSearch(provider string, query string) (ret []*hibikemanga.SearchResult, err error) {
	defer util.HandlePanicInModuleWithError("manga/ManualSearch", &err)

	if query == "" {
		return make([]*hibikemanga.SearchResult, 0), nil
	}

	// Get the search results
	providerExtension, ok := extension.GetExtension[extension.MangaProviderExtension](r.providerExtensionBank, provider)
	if !ok {
		r.logger.Error().Str("provider", provider).Msg("manga: Provider not found")
		return nil, errors.New("manga: Provider not found")
	}

	normalizedQuery := strings.ToLower(strings.TrimSpace(query))

	searchRes, found := searchResultCache.Get(provider + normalizedQuery)
	if found {
		return searchRes, nil
	}

	searchRes, err = providerExtension.GetProvider().Search(hibikemanga.SearchOptions{
		Query: normalizedQuery,
	})
	if err != nil {
		r.logger.Error().Err(err).Str("query", normalizedQuery).Msg("manga: Search failed")
		return nil, err
	}

	// Overwrite the provider just in case
	for _, res := range searchRes {
		res.Provider = provider
	}

	searchResultCache.Set(provider+normalizedQuery, searchRes)

	return searchRes, nil
}

// ManualMapping is used to manually map a manga to a provider.
// After calling this, the client should re-fetch the chapter container.
func (r *Repository) ManualMapping(provider string, mediaId int, mangaId string) (err error) {
	defer util.HandlePanicInModuleWithError("manga/ManualMapping", &err)

	r.logger.Trace().Msgf("manga: Removing cached bucket for %s, media ID: %d", provider, mediaId)

	// Delete the cached chapter container if any
	bucket := r.getFcProviderBucket(provider, mediaId, bucketTypeChapter)
	_ = r.fileCacher.Remove(bucket.Name())

	r.logger.Trace().
		Str("provider", provider).
		Int("mediaId", mediaId).
		Str("mangaId", mangaId).
		Msg("manga: Manual mapping")

	// Insert the mapping into the database
	err = r.db.InsertMangaMapping(provider, mediaId, mangaId)
	if err != nil {
		r.logger.Error().Err(err).Msg("manga: Failed to insert mapping")
		return err
	}

	r.logger.Debug().Msg("manga: Manual mapping successful")

	return nil
}

type MappingResponse struct {
	MangaID *string `json:"mangaId"`
}

func (r *Repository) GetMapping(provider string, mediaId int) (ret MappingResponse) {
	defer util.HandlePanicInModuleThen("manga/GetMapping", func() {
		ret = MappingResponse{}
	})

	mapping, found := r.db.GetMangaMapping(provider, mediaId)
	if !found {
		return MappingResponse{}
	}

	return MappingResponse{
		MangaID: &mapping.MangaID,
	}
}

func (r *Repository) RemoveMapping(provider string, mediaId int) (err error) {
	defer util.HandlePanicInModuleWithError("manga/RemoveMapping", &err)

	// Delete the mapping from the database
	err = r.db.DeleteMangaMapping(provider, mediaId)
	if err != nil {
		r.logger.Error().Err(err).Msg("manga: Failed to delete mapping")
		return err
	}

	r.logger.Debug().Msg("manga: Mapping removed")

	r.logger.Trace().Msgf("manga: Removing cached bucket for %s, media ID: %d", provider, mediaId)
	// Delete the cached chapter container if any
	bucket := r.getFcProviderBucket(provider, mediaId, bucketTypeChapter)
	_ = r.fileCacher.Remove(bucket.Name())

	return nil
}
