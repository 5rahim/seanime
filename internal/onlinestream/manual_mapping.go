package onlinestream

import (
	"errors"
	"fmt"
	"seanime/internal/extension"
	"seanime/internal/util"
	"seanime/internal/util/result"
	"strings"

	hibikeonlinestream "seanime/internal/extension/hibike/onlinestream"
)

var searchResultCache = result.NewCache[string, []*hibikeonlinestream.SearchResult]()

func (r *Repository) ManualSearch(provider string, query string, dub bool) (ret []*hibikeonlinestream.SearchResult, err error) {
	defer util.HandlePanicInModuleWithError("onlinestream/ManualSearch", &err)

	if query == "" {
		return make([]*hibikeonlinestream.SearchResult, 0), nil
	}

	// Get the search results
	providerExtension, ok := extension.GetExtension[extension.OnlinestreamProviderExtension](r.providerExtensionBank, provider)
	if !ok {
		r.logger.Error().Str("provider", provider).Msg("onlinestream: Provider not found")
		return nil, errors.New("onlinestream: Provider not found")
	}

	normalizedQuery := strings.ToLower(strings.TrimSpace(query))

	searchRes, found := searchResultCache.Get(provider + normalizedQuery + fmt.Sprintf("%t", dub))
	if found {
		return searchRes, nil
	}

	searchRes, err = providerExtension.GetProvider().Search(hibikeonlinestream.SearchOptions{
		Query: normalizedQuery,
		Dub:   dub,
		Year:  0,
	})
	if err != nil {
		r.logger.Error().Err(err).Str("query", normalizedQuery).Msg("onlinestream: Search failed")
		return nil, err
	}

	searchResultCache.Set(provider+normalizedQuery+fmt.Sprintf("%t", dub), searchRes)

	return searchRes, nil
}

// ManualMapping is used to manually map an anime to a provider.
// After calling this, the client should re-fetch the episode list.
func (r *Repository) ManualMapping(provider string, mediaId int, animeId string) (err error) {
	defer util.HandlePanicInModuleWithError("onlinestream/ManualMapping", &err)

	r.logger.Trace().Msgf("onlinestream: Removing cached bucket for %s, media ID: %d", provider, mediaId)

	// Delete the cached data if any
	epListBucket := r.getFcEpisodeListBucket(provider, mediaId)
	_ = r.fileCacher.Remove(epListBucket.Name())
	epDataBucket := r.getFcEpisodeDataBucket(provider, mediaId)
	_ = r.fileCacher.Remove(epDataBucket.Name())

	r.logger.Trace().
		Str("provider", provider).
		Int("mediaId", mediaId).
		Str("animeId", animeId).
		Msg("onlinestream: Manual mapping")

	// Insert the mapping into the database
	err = r.db.InsertOnlinestreamMapping(provider, mediaId, animeId)
	if err != nil {
		r.logger.Error().Err(err).Msg("onlinestream: Failed to insert mapping")
		return err
	}

	r.logger.Debug().Msg("onlinestream: Manual mapping successful")

	return nil
}

type MappingResponse struct {
	AnimeId *string `json:"animeId"`
}

func (r *Repository) GetMapping(provider string, mediaId int) (ret MappingResponse) {
	defer util.HandlePanicInModuleThen("onlinestream/GetMapping", func() {
		ret = MappingResponse{}
	})

	mapping, found := r.db.GetOnlinestreamMapping(provider, mediaId)
	if !found {
		return MappingResponse{}
	}

	return MappingResponse{
		AnimeId: &mapping.AnimeID,
	}
}

func (r *Repository) RemoveMapping(provider string, mediaId int) (err error) {
	defer util.HandlePanicInModuleWithError("onlinestream/RemoveMapping", &err)

	// Delete the mapping from the database
	err = r.db.DeleteOnlinestreamMapping(provider, mediaId)
	if err != nil {
		r.logger.Error().Err(err).Msg("onlinestream: Failed to delete mapping")
		return err
	}

	r.logger.Debug().Msg("onlinestream: Mapping removed")

	r.logger.Trace().Msgf("onlinestream: Removing cached bucket for %s, media ID: %d", provider, mediaId)
	// Delete the cached data if any
	epListBucket := r.getFcEpisodeListBucket(provider, mediaId)
	_ = r.fileCacher.Remove(epListBucket.Name())
	epDataBucket := r.getFcEpisodeDataBucket(provider, mediaId)
	_ = r.fileCacher.Remove(epDataBucket.Name())

	return nil
}
