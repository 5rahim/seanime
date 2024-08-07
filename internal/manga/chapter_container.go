package manga

import (
	"errors"
	"fmt"
	"github.com/samber/lo"
	"seanime/internal/extension"
	"seanime/internal/util"
	"seanime/internal/util/result"
	"strings"

	hibikemanga "github.com/5rahim/hibike/pkg/extension/manga"
)

var (
	ErrNoMapping = errors.New("manga: No mapping found")
)

type (
	// ChapterContainer is used to display the list of chapters from a provider in the client.
	// It is cached in a unique file cache bucket with a key of the format: {provider}${mediaId}
	ChapterContainer struct {
		MediaId  int                           `json:"mediaId"`
		Provider string                        `json:"provider"`
		Chapters []*hibikemanga.ChapterDetails `json:"chapters"`
	}
)

var searchResultCache = result.NewCache[string, []*hibikemanga.SearchResult]()

func (r *Repository) ManualSearch(provider string, query string) ([]*hibikemanga.SearchResult, error) {

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

	searchRes, found := searchResultCache.Get(normalizedQuery)
	if found {
		return searchRes, nil
	}

	searchRes, err := providerExtension.GetProvider().Search(hibikemanga.SearchOptions{
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

	searchResultCache.Set(normalizedQuery, searchRes)

	return searchRes, nil
}

// ManualMapping is used to manually map a manga to a provider.
// After calling this, the client should re-fetch the chapter container.
func (r *Repository) ManualMapping(provider string, mediaId int, mangaId string) error {

	r.logger.Trace().Msgf("manga: Removing cached bucket for %s, media ID: %d", provider, mediaId)

	// Delete the cached chapter container if any
	bucket := r.getFcProviderBucket(provider, mediaId, bucketTypeChapter)
	_ = r.fileCacher.Delete(bucket, fmt.Sprintf("%s$%d", provider, mediaId))

	r.logger.Trace().
		Str("provider", provider).
		Int("mediaId", mediaId).
		Str("mangaId", mangaId).
		Msg("manga: Manual mapping")

	// Insert the mapping into the database
	err := r.db.InsertMangaMapping(provider, mediaId, mangaId)
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

func (r *Repository) GetMapping(provider string, mediaId int) MappingResponse {
	mapping, found := r.db.GetMangaMapping(provider, mediaId)
	if !found {
		return MappingResponse{}
	}

	return MappingResponse{
		MangaID: &mapping.MangaID,
	}
}

func (r *Repository) RemoveMapping(provider string, mediaId int) error {

	// Delete the mapping from the database
	err := r.db.DeleteMangaMapping(provider, mediaId)
	if err != nil {
		r.logger.Error().Err(err).Msg("manga: Failed to delete mapping")
		return err
	}

	r.logger.Debug().Msg("manga: Mapping removed")

	r.logger.Trace().Msgf("manga: Removing cached bucket for %s, media ID: %d", provider, mediaId)
	// Delete the cached chapter container if any
	bucket := r.getFcProviderBucket(provider, mediaId, bucketTypeChapter)
	_ = r.fileCacher.Delete(bucket, fmt.Sprintf("%s$%d", provider, mediaId))

	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// GetMangaChapterContainer returns the ChapterContainer for a manga entry based on the provider.
// If it isn't cached, it will search for the manga, create a ChapterContainer and cache it.
func (r *Repository) GetMangaChapterContainer(provider string, mediaId int, titles []*string) (*ChapterContainer, error) {

	key := fmt.Sprintf("%s$%d", provider, mediaId)

	r.logger.Debug().
		Str("provider", provider).
		Int("mediaId", mediaId).
		Str("key", key).
		Msgf("manga: getting chapters")

	// +---------------------+
	// |       Cache         |
	// +---------------------+

	var container *ChapterContainer
	bucket := r.getFcProviderBucket(provider, mediaId, bucketTypeChapter)

	// Check if the container is in the cache
	if found, _ := r.fileCacher.Get(bucket, key, &container); found {
		r.logger.Info().Str("key", key).Msg("manga: Chapter Container Cache HIT")
		return container, nil
	}

	providerExtension, ok := extension.GetExtension[extension.MangaProviderExtension](r.providerExtensionBank, provider)
	if !ok {
		r.logger.Error().Str("provider", provider).Msg("manga: Provider not found")
		return nil, errors.New("manga: Provider not found")
	}

	var mangaId string

	// +---------------------+
	// |      Database       |
	// +---------------------+

	// Search for the mapping in the database
	mapping, found := r.db.GetMangaMapping(provider, mediaId)
	if found {
		r.logger.Debug().Str("mangaId", mapping.MangaID).Msg("manga: Using manual mapping")
		mangaId = mapping.MangaID
	}

	if mangaId == "" {
		// +---------------------+
		// |       Search        |
		// +---------------------+

		r.logger.Trace().Msg("manga: Searching for manga")

		if titles == nil {
			return nil, ErrNoTitlesProvided
		}

		titles = lo.Filter(titles, func(title *string, _ int) bool {
			return util.IsMostlyLatinString(*title)
		})

		var searchRes []*hibikemanga.SearchResult

		var err error
		for _, title := range titles {
			var _searchRes []*hibikemanga.SearchResult

			_searchRes, err = providerExtension.GetProvider().Search(hibikemanga.SearchOptions{
				Query: *title,
			})
			if err == nil {
				searchRes = append(searchRes, _searchRes...)
			} else {
				r.logger.Warn().Err(err).Msg("manga: search failed")
			}
		}

		if searchRes == nil || len(searchRes) == 0 {
			r.logger.Error().Msg("manga: no search results found")
			return nil, ErrNoResults
		}

		// Overwrite the provider just in case
		for _, res := range searchRes {
			res.Provider = provider
		}

		bestRes := searchRes[0]
		for _, res := range searchRes {
			if res.SearchRating > bestRes.SearchRating {
				bestRes = res
			}
		}

		mangaId = bestRes.ID
	}

	// +---------------------+
	// |    Get chapters     |
	// +---------------------+

	chapterList, err := providerExtension.GetProvider().FindChapters(mangaId)
	if err != nil {
		r.logger.Error().Err(err).Msg("manga: Failed to get chapters")
		return nil, ErrNoChapters
	}

	// Overwrite the provider just in case
	for _, chapter := range chapterList {
		chapter.Provider = provider
	}

	container = &ChapterContainer{
		MediaId:  mediaId,
		Provider: provider,
		Chapters: chapterList,
	}

	// DEVNOTE: This might cache container with empty chapters, however the user can reload sources, so it's fine
	err = r.fileCacher.Set(bucket, key, container)
	if err != nil {
		r.logger.Warn().Err(err).Msg("manga: Failed to populate cache")
	}

	r.logger.Info().Str("key", key).Msg("manga: Retrieved chapters")

	return container, nil
}
