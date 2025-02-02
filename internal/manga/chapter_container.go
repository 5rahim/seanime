package manga

import (
	"cmp"
	"errors"
	"fmt"
	"math"
	"os"
	"seanime/internal/extension"
	"seanime/internal/util"
	"seanime/internal/util/comparison"
	"seanime/internal/util/result"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/samber/lo"

	hibikemanga "github.com/5rahim/hibike/pkg/extension/manga"
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

func getMangaChapterContainerCacheKey(provider string, mediaId int) string {
	return fmt.Sprintf("%s$%d", provider, mediaId)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type GetMangaChapterContainerOptions struct {
	Provider string
	MediaId  int
	Titles   []*string
	Year     int
}

// GetMangaChapterContainer returns the ChapterContainer for a manga entry based on the provider.
// If it isn't cached, it will search for the manga, create a ChapterContainer and cache it.
func (r *Repository) GetMangaChapterContainer(opts *GetMangaChapterContainerOptions) (ret *ChapterContainer, err error) {
	defer util.HandlePanicInModuleWithError("manga/GetMangaChapterContainer", &err)

	provider := opts.Provider
	mediaId := opts.MediaId
	titles := opts.Titles

	r.logger.Trace().
		Str("provider", provider).
		Int("mediaId", mediaId).
		Msgf("manga: Getting chapters")

	chapterContainerKey := getMangaChapterContainerCacheKey(provider, mediaId)

	// +---------------------+
	// |       Cache         |
	// +---------------------+

	var container *ChapterContainer
	containerBucket := r.getFcProviderBucket(provider, mediaId, bucketTypeChapter)

	// Check if the container is in the cache
	if found, _ := r.fileCacher.Get(containerBucket, chapterContainerKey, &container); found {
		r.logger.Info().Str("bucket", containerBucket.Name()).Msg("manga: Chapter Container Cache HIT")
		return container, nil
	}

	// Delete the map cache
	mangaChapterCountMap.Delete(ChapterCountMapCacheKey)

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
				Year:  opts.Year,
			})
			if err == nil {

				HydrateSearchResultSearchRating(_searchRes, title)

				searchRes = append(searchRes, _searchRes...)
			} else {
				r.logger.Warn().Err(err).Msg("manga: Search failed")
			}
		}

		if searchRes == nil || len(searchRes) == 0 {
			r.logger.Error().Msg("manga: No search results found")
			return nil, ErrNoResults
		}

		// Overwrite the provider just in case
		for _, res := range searchRes {
			res.Provider = provider
		}

		bestRes := GetBestSearchResult(searchRes)

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
	err = r.fileCacher.Set(containerBucket, chapterContainerKey, container)
	if err != nil {
		r.logger.Warn().Err(err).Msg("manga: Failed to populate cache")
	}

	r.logger.Info().Str("bucket", containerBucket.Name()).Msg("manga: Retrieved chapters")

	return container, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

const ChapterCountMapCacheKey = 1

var mangaChapterCountMap = result.NewResultMap[int, map[int]int]()

func (r *Repository) GetMangaChapterCountMap() (ret map[int]int, err error) {
	defer util.HandlePanicInModuleThen("manga/GetMangaCurrentChapterCountMap", func() {})
	ret = make(map[int]int)

	if m, ok := mangaChapterCountMap.Get(ChapterCountMapCacheKey); ok {
		ret = m
		return
	}

	// Go through all chapter container caches
	entries, err := os.ReadDir(r.cacheDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		provider, mediaId, ok := parseChapterFileName(entry.Name())
		if !ok {
			continue
		}

		var container *ChapterContainer
		containerBucket := r.getFcProviderBucket(provider, mediaId, bucketTypeChapter)

		// Check if the container is in the cache
		chapterContainerKey := getMangaChapterContainerCacheKey(provider, mediaId)
		if found, _ := r.fileCacher.Get(containerBucket, chapterContainerKey, &container); !found {
			continue
		}

		lastChapter := slices.MaxFunc(container.Chapters, func(a *hibikemanga.ChapterDetails, b *hibikemanga.ChapterDetails) int {
			return cmp.Compare(a.Index, b.Index)
		})
		chapterNumFloat, _ := strconv.ParseFloat(lastChapter.Chapter, 32)
		chapterCount := int(math.Floor(chapterNumFloat))

		ret[mediaId] = chapterCount
	}

	mangaChapterCountMap.Set(ChapterCountMapCacheKey, ret)

	return
}

func parseChapterFileName(dirName string) (provider string, mId int, ok bool) {
	if !strings.HasPrefix(dirName, "manga_") {
		return "", 0, false
	}
	dirName = strings.TrimSuffix(dirName, ".cache")
	parts := strings.Split(dirName, "_")
	if len(parts) != 4 {
		return "", 0, false
	}

	provider = parts[1]
	mId, err := strconv.Atoi(parts[3])
	if err != nil {
		return "", 0, false
	}

	return provider, mId, true
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func GetBestSearchResult(searchRes []*hibikemanga.SearchResult) *hibikemanga.SearchResult {
	bestRes := searchRes[0]
	for _, res := range searchRes {
		if res.SearchRating > bestRes.SearchRating {
			bestRes = res
		}
	}
	return bestRes
}

// HydrateSearchResultSearchRating rates the search results based on the provided title
// It checks if all search results have a rating of 0 and if so, it calculates ratings
// using the Sorensen-Dice
func HydrateSearchResultSearchRating(_searchRes []*hibikemanga.SearchResult, title *string) {
	// Rate the search results if all ratings are 0
	if noRatings := lo.EveryBy(_searchRes, func(res *hibikemanga.SearchResult) bool {
		return res.SearchRating == 0
	}); noRatings {
		wg := sync.WaitGroup{}
		wg.Add(len(_searchRes))
		for _, res := range _searchRes {
			go func(res *hibikemanga.SearchResult) {
				defer wg.Done()

				compTitles := []*string{&res.Title}
				if res.Synonyms == nil || len(res.Synonyms) == 0 {
					return
				}
				for _, syn := range res.Synonyms {
					compTitles = append(compTitles, &syn)
				}

				compRes, ok := comparison.FindBestMatchWithSorensenDice(title, compTitles)
				if !ok {
					return
				}

				res.SearchRating = compRes.Rating
				return
			}(res)
		}
		wg.Wait()
	}
}
