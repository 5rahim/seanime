package manga

import (
	"cmp"
	"errors"
	"fmt"
	"math"
	"os"
	"seanime/internal/api/anilist"
	"seanime/internal/extension"
	hibikemanga "seanime/internal/extension/hibike/manga"
	"seanime/internal/hook"
	manga_providers "seanime/internal/manga/providers"
	"seanime/internal/util"
	"seanime/internal/util/comparison"
	"seanime/internal/util/result"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/samber/lo"
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

	providerExtension, ok := extension.GetExtension[extension.MangaProviderExtension](r.providerExtensionBank, provider)
	if !ok {
		r.logger.Error().Str("provider", provider).Msg("manga: Provider not found")
		return nil, errors.New("manga: Provider not found")
	}

	// DEVNOTE: Local chapters can be cached
	localProvider, isLocalProvider := providerExtension.GetProvider().(*manga_providers.Local)

	// Set the source directory for local provider
	if isLocalProvider && r.settings.Manga.LocalSourceDirectory != "" {
		localProvider.SetSourceDirectory(r.settings.Manga.LocalSourceDirectory)
	}

	r.logger.Trace().
		Str("provider", provider).
		Int("mediaId", mediaId).
		Msgf("manga: Getting chapters")

	chapterContainerKey := getMangaChapterContainerCacheKey(provider, mediaId)

	// +---------------------+
	// |     Hook event      |
	// +---------------------+

	// Trigger hook event
	reqEvent := &MangaChapterContainerRequestedEvent{
		Provider: provider,
		MediaId:  mediaId,
		Titles:   titles,
		Year:     opts.Year,
		ChapterContainer: &ChapterContainer{
			MediaId:  mediaId,
			Provider: provider,
			Chapters: []*hibikemanga.ChapterDetails{},
		},
	}
	err = hook.GlobalHookManager.OnMangaChapterContainerRequested().Trigger(reqEvent)
	if err != nil {
		r.logger.Error().Err(err).Msg("manga: Exception occurred while triggering hook event")
		return nil, fmt.Errorf("manga: Error in hook, %w", err)
	}

	// Default prevented, return the chapter container
	if reqEvent.DefaultPrevented {
		if reqEvent.ChapterContainer == nil {
			return nil, fmt.Errorf("manga: No chapter container returned by hook event")
		}
		return reqEvent.ChapterContainer, nil
	}

	// +---------------------+
	// |       Cache         |
	// +---------------------+

	var container *ChapterContainer
	containerBucket := r.getFcProviderBucket(provider, mediaId, bucketTypeChapter)

	// Check if the container is in the cache
	if found, _ := r.fileCacher.Get(containerBucket, chapterContainerKey, &container); found {
		r.logger.Info().Str("bucket", containerBucket.Name()).Msg("manga: Chapter Container Cache HIT")

		// Trigger hook event
		ev := &MangaChapterContainerEvent{
			ChapterContainer: container,
		}
		err = hook.GlobalHookManager.OnMangaChapterContainer().Trigger(ev)
		if err != nil {
			r.logger.Error().Err(err).Msg("manga: Exception occurred while triggering hook event")
		}
		container = ev.ChapterContainer

		return container, nil
	}

	// Delete the map cache
	mangaLatestChapterNumberMap.Delete(ChapterCountMapCacheKey)

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

		if len(searchRes) == 0 {
			r.logger.Error().Msg("manga: No search results found")
			if err != nil {
				return nil, fmt.Errorf("%w, %w", ErrNoResults, err)
			} else {
				return nil, ErrNoResults
			}
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

	// Trigger hook event
	ev := &MangaChapterContainerEvent{
		ChapterContainer: container,
	}
	err = hook.GlobalHookManager.OnMangaChapterContainer().Trigger(ev)
	if err != nil {
		r.logger.Error().Err(err).Msg("manga: Exception occurred while triggering hook event")
	}
	container = ev.ChapterContainer

	// Cache the container only if it has chapters
	if len(container.Chapters) > 0 {
		err = r.fileCacher.Set(containerBucket, chapterContainerKey, container)
		if err != nil {
			r.logger.Warn().Err(err).Msg("manga: Failed to populate cache")
		}
	}

	r.logger.Info().Str("bucket", containerBucket.Name()).Msg("manga: Retrieved chapters")
	return container, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// RefreshChapterContainers deletes all cached chapter containers and refetches them based on the selected provider map.
func (r *Repository) RefreshChapterContainers(mangaCollection *anilist.MangaCollection, selectedProviderMap map[int]string) (err error) {
	defer util.HandlePanicInModuleWithError("manga/RefreshChapterContainers", &err)

	// Read the cache directory
	entries, err := os.ReadDir(r.cacheDir)
	if err != nil {
		return err
	}

	removedMediaIds := make(map[int]struct{})
	mu := sync.Mutex{}

	wg := sync.WaitGroup{}
	wg.Add(len(entries))
	for _, entry := range entries {
		go func(entry os.DirEntry) {
			defer wg.Done()

			if entry.IsDir() {
				return
			}

			provider, bucketType, mediaId, ok := ParseChapterContainerFileName(entry.Name())
			if !ok {
				return
			}
			// If the bucket type is not chapter, skip
			if bucketType != bucketTypeChapter {
				return
			}

			r.logger.Trace().Str("provider", provider).Int("mediaId", mediaId).Msg("manga: Refetching chapter container")

			mu.Lock()
			// Remove the container from the cache if it hasn't been removed yet
			if _, ok := removedMediaIds[mediaId]; !ok {
				_ = r.EmptyMangaCache(mediaId)
				removedMediaIds[mediaId] = struct{}{}
			}
			mu.Unlock()

			// If a selectedProviderMap is provided, check if the provider is in the map
			if selectedProviderMap != nil {
				// If the manga is not in the map, continue
				if _, ok := selectedProviderMap[mediaId]; !ok {
					return
				}

				// If the provider is not the one selected, continue
				if selectedProviderMap[mediaId] != provider {
					return
				}
			}

			// Get the manga from the collection
			mangaEntry, found := mangaCollection.GetListEntryFromMangaId(mediaId)
			if !found {
				return
			}

			// If the manga is not currently reading or repeating, continue
			if *mangaEntry.GetStatus() != anilist.MediaListStatusCurrent && *mangaEntry.GetStatus() != anilist.MediaListStatusRepeating {
				return
			}

			// Refetch the container
			_, err = r.GetMangaChapterContainer(&GetMangaChapterContainerOptions{
				Provider: provider,
				MediaId:  mediaId,
				Titles:   mangaEntry.GetMedia().GetAllTitles(),
				Year:     mangaEntry.GetMedia().GetStartYearSafe(),
			})
			if err != nil {
				r.logger.Error().Err(err).Msg("manga: Failed to refetch chapter container")
				return
			}

			r.logger.Trace().Str("provider", provider).Int("mediaId", mediaId).Msg("manga: Refetched chapter container")
		}(entry)
	}
	wg.Wait()

	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

const ChapterCountMapCacheKey = 1

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

	// Trigger hook event
	ev := &MangaLatestChapterNumbersMapEvent{
		LatestChapterNumbersMap: ret,
	}
	err = hook.GlobalHookManager.OnMangaLatestChapterNumbersMap().Trigger(ev)
	if err != nil {
		r.logger.Error().Err(err).Msg("manga: Exception occurred while triggering hook event")
	}
	ret = ev.LatestChapterNumbersMap

	mangaLatestChapterNumberMap.Set(ChapterCountMapCacheKey, ret)
	return
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

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
