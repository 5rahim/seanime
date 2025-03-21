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

	hibikemanga "seanime/internal/extension/hibike/manga"
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
	mangaLatestChapterNumberMap.Delete(ChapterCountMapCacheKey)

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
var mangaLatestChapterNumberMap = result.NewResultMap[int, map[int][]MangaLatestChapterNumberItem]()

type MangaLatestChapterNumberItem struct {
	Provider  string `json:"provider"`
	Scanlator string `json:"scanlator"`
	Language  string `json:"language"`
	Number    int    `json:"number"`
}

// GetMangaLatestChapterNumbersMap retrieves the latest chapter number for all manga entries.
// It scans the cache directory for chapter containers and counts the number of chapters fetched from the provider for each manga.
//
// Unlike [GetMangaLatestChapterNumberMap], it will segregate the chapter numbers by scanlator and language.
func (r *Repository) GetMangaLatestChapterNumbersMap() (ret map[int][]MangaLatestChapterNumberItem, err error) {
	defer util.HandlePanicInModuleThen("manga/GetMangaLatestChapterNumbersMap", func() {})
	ret = make(map[int][]MangaLatestChapterNumberItem)

	if m, ok := mangaLatestChapterNumberMap.Get(ChapterCountMapCacheKey); ok {
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

		// Get the provider and mediaId from the file cache name
		provider, mediaId, ok := parseChapterFileName(entry.Name())
		if !ok {
			continue
		}

		fmt.Println(entry.Name(), provider, mediaId)

		containerBucket := r.getFcProviderBucket(provider, mediaId, bucketTypeChapter)

		// Get the container from the file cache
		var container *ChapterContainer
		chapterContainerKey := getMangaChapterContainerCacheKey(provider, mediaId)
		if found, _ := r.fileCacher.Get(containerBucket, chapterContainerKey, &container); !found {
			continue
		}

		// Create groups
		groupByScanlator := lo.GroupBy(container.Chapters, func(c *hibikemanga.ChapterDetails) string {
			return c.Scanlator
		})

		for scanlator, chapters := range groupByScanlator {
			groupByLanguage := lo.GroupBy(chapters, func(c *hibikemanga.ChapterDetails) string {
				return c.Language
			})

			for language, chapters := range groupByLanguage {
				lastChapter := slices.MaxFunc(chapters, func(a *hibikemanga.ChapterDetails, b *hibikemanga.ChapterDetails) int {
					return cmp.Compare(a.Index, b.Index)
				})

				chapterNumFloat, _ := strconv.ParseFloat(lastChapter.Chapter, 32)
				chapterCount := int(math.Floor(chapterNumFloat))

				if _, ok := ret[mediaId]; !ok {
					ret[mediaId] = []MangaLatestChapterNumberItem{}
				}

				ret[mediaId] = append(ret[mediaId], MangaLatestChapterNumberItem{
					Provider:  provider,
					Scanlator: scanlator,
					Language:  language,
					Number:    chapterCount,
				})
			}
		}
	}

	mangaLatestChapterNumberMap.Set(ChapterCountMapCacheKey, ret)

	return
}

// GetMangaLatestChapterNumberMap retrieves the latest chapter number for all manga entries.
// It scans the cache directory for chapter containers and counts the number of chapters fetched from the provider for each manga.
//
// Note that this doesn't take into account selected scanlators, so the chapter count might be inaccurate.
func (r *Repository) GetMangaLatestChapterNumberMap() (ret map[int]int, err error) {
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

		// Get the provider and mediaId from the file cache name
		provider, mediaId, ok := parseChapterFileName(entry.Name())
		if !ok {
			continue
		}

		containerBucket := r.getFcProviderBucket(provider, mediaId, bucketTypeChapter)

		// Get the container from the file cache
		var container *ChapterContainer
		chapterContainerKey := getMangaChapterContainerCacheKey(provider, mediaId)
		if found, _ := r.fileCacher.Get(containerBucket, chapterContainerKey, &container); !found {
			continue
		}

		// Get the last chapter from the container by sorting the chapters by index
		// This is more accurate than counting the number of chapters in the container
		lastChapter := slices.MaxFunc(container.Chapters, func(a *hibikemanga.ChapterDetails, b *hibikemanga.ChapterDetails) int {
			return cmp.Compare(a.Index, b.Index)
		})

		// Convert the last chapter number to a float and round down to get the chapter count
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
