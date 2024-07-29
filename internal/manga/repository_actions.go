package manga

import (
	"cmp"
	"errors"
	"fmt"
	hibikemanga "github.com/5rahim/hibike/pkg/extension/manga"
	"github.com/goccy/go-json"
	"github.com/samber/lo"
	"os"
	"path/filepath"
	"seanime/internal/api/anilist"
	"seanime/internal/extension"
	"seanime/internal/manga/downloader"
	"seanime/internal/manga/providers"
	"seanime/internal/util"
	"slices"
	"strconv"
	"strings"
	"sync"
)

var (
	ErrNoResults            = errors.New("no results found for this media")
	ErrNoChapters           = errors.New("no manga chapters found")
	ErrChapterNotFound      = errors.New("chapter not found")
	ErrChapterNotDownloaded = errors.New("chapter not downloaded")
	ErrNoTitlesProvided     = errors.New("no titles provided")
)

type (
	// ChapterContainer is used to display the list of chapters from a provider in the client.
	// It is cached in a unique file cache bucket with a key of the format: {provider}${mediaId}
	ChapterContainer struct {
		MediaId  int                           `json:"mediaId"`
		Provider string                        `json:"provider"`
		Chapters []*hibikemanga.ChapterDetails `json:"chapters"`
	}

	// PageContainer is used to display the list of pages from a chapter in the client.
	// It is cached in the file cache bucket with a key of the format: {provider}${mediaId}${chapterId}
	PageContainer struct {
		MediaId        int                        `json:"mediaId"`
		Provider       string                     `json:"provider"`
		ChapterId      string                     `json:"chapterId"`
		Pages          []*hibikemanga.ChapterPage `json:"pages"`
		PageDimensions map[int]*PageDimension     `json:"pageDimensions"` // Indexed by page number
		IsDownloaded   bool                       `json:"isDownloaded"`   // TODO remove
	}

	// PageDimension is used to store the dimensions of a page.
	// It is used by the client for 'Double Page' mode.
	PageDimension struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	}
)

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

	// +---------------------+
	// |       Search        |
	// +---------------------+

	if titles == nil {
		return nil, ErrNoTitlesProvided
	}

	titles = lo.Filter(titles, func(title *string, _ int) bool {
		return util.IsMostlyLatinString(*title)
	})

	var searchRes []*hibikemanga.SearchResult

	providerExtension, ok := extension.GetExtension[extension.MangaProviderExtension](r.providerExtensionBank, provider)
	if !ok {
		r.logger.Error().Str("provider", provider).Msg("manga: Provider not found")
		return nil, errors.New("manga: Provider not found")
	}

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

	bestRes := searchRes[0]
	for _, res := range searchRes {
		if res.SearchRating > bestRes.SearchRating {
			bestRes = res
		}
	}

	// +---------------------+
	// |    Get chapters     |
	// +---------------------+

	var chapterList []*hibikemanga.ChapterDetails

	chapterList, err = providerExtension.GetProvider().FindChapters(bestRes.ID)

	if err != nil {
		r.logger.Error().Err(err).Msg("manga: Failed to get chapters")
		return nil, ErrNoChapters
	}

	container = &ChapterContainer{
		MediaId:  mediaId,
		Provider: string(provider),
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

// +-------------------------------------------------------------------------------------------------------------------+

// GetMangaPageContainer returns the PageContainer for a manga chapter based on the provider.
func (r *Repository) GetMangaPageContainer(
	provider string,
	mediaId int,
	chapterId string,
	doublePage bool,
	isOffline bool,
) (*PageContainer, error) {

	// +---------------------+
	// |      Downloads      |
	// +---------------------+

	if isOffline {
		ret, err := r.getDownloadedMangaPageContainer(provider, mediaId, chapterId)
		if err != nil {
			return nil, err
		}
		return ret, nil
	} else {
		ret, _ := r.getDownloadedMangaPageContainer(provider, mediaId, chapterId)
		if ret != nil {
			return ret, nil
		}
	}

	// +---------------------+
	// |      Get Pages      |
	// +---------------------+

	// PageContainer key
	key := fmt.Sprintf("%s$%d$%s", provider, mediaId, chapterId)

	r.logger.Debug().
		Str("provider", string(provider)).
		Int("mediaId", mediaId).
		Str("key", key).
		Str("chapterId", chapterId).
		Msgf("manga: getting pages")

	// +---------------------+
	// |       Cache         |
	// +---------------------+

	var container *PageContainer

	// PageContainer bucket
	// e.g., manga_comick_pages_123
	//         -> { "comick$123$10010": PageContainer }, { "comick$123$10011": PageContainer }
	bucket := r.getFcProviderBucket(provider, mediaId, bucketTypePage)

	// Check if the container is in the cache
	if found, _ := r.fileCacher.Get(bucket, key, &container); found {

		// Hydrate page dimensions
		pageDimensions, _ := r.getPageDimensions(doublePage, string(provider), mediaId, chapterId, container.Pages)
		container.PageDimensions = pageDimensions

		r.logger.Info().Str("key", key).Msg("manga: Page Container Cache HIT")
		return container, nil
	}

	// +---------------------+
	// |     Fetch pages     |
	// +---------------------+

	// Search for the chapter in the cache
	chapterBucket := r.getFcProviderBucket(provider, mediaId, bucketTypeChapter)

	var chapterContainer *ChapterContainer
	if found, _ := r.fileCacher.Get(chapterBucket, fmt.Sprintf("%s$%d", provider, mediaId), &chapterContainer); !found {
		r.logger.Error().Msg("manga: Chapter Container not found")
		return nil, ErrNoChapters
	}

	// Get the chapter from the container
	var chapter *hibikemanga.ChapterDetails
	for _, c := range chapterContainer.Chapters {
		if c.ID == chapterId {
			chapter = c
			break
		}
	}

	if chapter == nil {
		r.logger.Error().Msg("manga: Chapter not found")
		return nil, ErrChapterNotFound
	}

	providerExtension, ok := extension.GetExtension[extension.MangaProviderExtension](r.providerExtensionBank, provider)
	if !ok {
		r.logger.Error().Str("provider", provider).Msg("manga: Provider not found")
		return nil, errors.New("manga: Provider not found")
	}

	// Get the chapter pages
	var pageList []*hibikemanga.ChapterPage
	var err error

	pageList, err = providerExtension.GetProvider().FindChapterPages(chapter.ID)

	if err != nil {
		r.logger.Error().Err(err).Msg("manga: Could not get chapter pages")
		return nil, err
	}

	pageDimensions, _ := r.getPageDimensions(doublePage, string(provider), mediaId, chapterId, pageList)

	container = &PageContainer{
		MediaId:        mediaId,
		Provider:       string(provider),
		ChapterId:      chapterId,
		Pages:          pageList,
		PageDimensions: pageDimensions,
		IsDownloaded:   false,
	}

	// Set cache
	err = r.fileCacher.Set(bucket, key, container)
	if err != nil {
		r.logger.Warn().Err(err).Msg("manga: Failed to populate cache")
	}

	r.logger.Info().Str("key", key).Msg("manga: Retrieved pages")

	return container, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) getPageDimensions(enabled bool, provider string, mediaId int, chapterId string, pages []*hibikemanga.ChapterPage) (ret map[int]*PageDimension, err error) {
	util.HandlePanicInModuleThen("manga/getPageDimensions", func() {
		err = fmt.Errorf("failed to get page dimensions")
	})

	if !enabled {
		return nil, nil
	}

	key := fmt.Sprintf("%s$%d$%s", provider, mediaId, chapterId)

	// Page dimensions bucket
	// e.g., manga_comick_page-dimensions_123
	//         -> { "comick$123$10010": PageDimensions }, { "comick$123$10011": PageDimensions }
	bucket := r.getFcProviderBucket(string(provider), mediaId, bucketTypePageDimensions)

	if found, _ := r.fileCacher.Get(bucket, fmt.Sprintf(key, provider, mediaId), &ret); found {
		r.logger.Info().Str("key", key).Msg("manga: Page Dimensions Cache HIT")
		return
	}

	r.logger.Debug().Str("key", key).Msg("manga: Getting page dimensions")

	// Get the page dimensions
	pageDimensions := make(map[int]*PageDimension)
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	for _, page := range pages {
		wg.Add(1)
		go func(page *hibikemanga.ChapterPage) {
			defer wg.Done()
			buf, err := manga_providers.GetImageByProxy(page.URL, page.Headers)
			if err != nil {
				return
			}
			width, height, err := getImageNaturalSizeB(buf)
			if err != nil {
				//r.logger.Warn().Err(err).Int("index", page.Index).Msg("manga: failed to get image size")
				return
			}

			mu.Lock()
			// DEVNOTE: Index by page index
			pageDimensions[page.Index] = &PageDimension{
				Width:  width,
				Height: height,
			}
			mu.Unlock()
		}(page)
	}
	wg.Wait()

	_ = r.fileCacher.Set(bucket, key, pageDimensions)

	r.logger.Info().Str("key", key).Msg("manga: Retrieved page dimensions")

	return pageDimensions, nil
}

// +---------------------+
// |      Downloads      |
// +---------------------+

// getDownloadedMangaPageContainer returns the PageContainer for a downloaded manga chapter based on the provider.
func (r *Repository) getDownloadedMangaPageContainer(
	provider string,
	mediaId int,
	chapterId string,
) (*PageContainer, error) {

	// Check if the chapter is downloaded
	found := false

	// Read download directory
	files, err := os.ReadDir(r.downloadDir)
	if err != nil {
		r.logger.Error().Err(err).Msg("manga: Failed to read download directory")
		return nil, err
	}

	chapterDir := "" // e.g., manga_comick_123_10010_13
	for _, file := range files {
		if file.IsDir() {
			parts := strings.SplitN(file.Name(), "_", 4)
			if len(parts) != 4 {
				continue
			}

			mId, _ := strconv.Atoi(parts[1])

			if parts[0] == string(provider) && mId == mediaId && parts[2] == chapterId {
				found = true
				chapterDir = file.Name()
				break
			}
		}
	}

	if !found {
		return nil, ErrChapterNotDownloaded
	}

	r.logger.Debug().Msg("manga: Found downloaded chapter directory")

	// Open registry file
	registryFile, err := os.Open(filepath.Join(r.downloadDir, chapterDir, "registry.json"))
	if err != nil {
		r.logger.Error().Err(err).Msg("manga: Failed to open registry file")
		return nil, err
	}
	defer registryFile.Close()

	r.logger.Info().Str("chapterId", chapterId).Msg("manga: Reading registry file")

	// Read registry file
	var pageRegistry *chapter_downloader.Registry
	err = json.NewDecoder(registryFile).Decode(&pageRegistry)
	if err != nil {
		r.logger.Error().Err(err).Msg("manga: Failed to decode registry file")
		return nil, err
	}

	pageList := make([]*hibikemanga.ChapterPage, 0)
	pageDimensions := make(map[int]*PageDimension)

	// Get the downloaded pages
	for pageIndex, pageInfo := range *pageRegistry {
		pageList = append(pageList, &hibikemanga.ChapterPage{
			Index:    pageIndex,
			URL:      filepath.Join(chapterDir, pageInfo.Filename),
			Provider: string(provider),
		})
		pageDimensions[pageIndex] = &PageDimension{
			Width:  pageInfo.Width,
			Height: pageInfo.Height,
		}
	}

	slices.SortStableFunc(pageList, func(i, j *hibikemanga.ChapterPage) int {
		return cmp.Compare(i.Index, j.Index)
	})

	container := &PageContainer{
		MediaId:        mediaId,
		Provider:       string(provider),
		ChapterId:      chapterId,
		Pages:          pageList,
		PageDimensions: pageDimensions,
		IsDownloaded:   true,
	}

	r.logger.Info().Str("chapterId", chapterId).Msg("manga: Found downloaded chapter")

	return container, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) GetDownloadedChapterContainers(mangaCollection *anilist.MangaCollection) (ret []*ChapterContainer, err error) {
	ret = make([]*ChapterContainer, 0)

	// Read download directory
	files, err := os.ReadDir(r.downloadDir)
	if err != nil {
		r.logger.Error().Err(err).Msg("manga: Failed to read download directory")
		return nil, err
	}

	// Get all chapter directories
	// e.g. manga_comick_123_10010_13
	chapterDirs := make([]string, 0)
	for _, file := range files {
		if file.IsDir() {
			parts := strings.SplitN(file.Name(), "_", 4)
			if len(parts) != 4 {
				continue
			}
			chapterDirs = append(chapterDirs, file.Name())
		}
	}

	if len(chapterDirs) == 0 {
		return nil, nil
	}

	// Now that we have all the chapter directories, we can get the chapter containers

	keys := make([]*chapter_downloader.DownloadID, 0)
	for _, dir := range chapterDirs {
		parts := strings.SplitN(dir, "_", 4)
		provider := parts[0]
		mediaId, _ := strconv.Atoi(parts[1])
		chapterId := parts[2]
		chapterNumber := parts[3]

		keys = append(keys, &chapter_downloader.DownloadID{
			Provider:      provider,
			MediaId:       mediaId,
			ChapterId:     chapterId,
			ChapterNumber: chapterNumber,
		})
	}

	providerAndMediaIdPairs := make(map[struct {
		provider string
		mediaId  int
	}]bool)

	for _, key := range keys {
		providerAndMediaIdPairs[struct {
			provider string
			mediaId  int
		}{
			provider: key.Provider,
			mediaId:  key.MediaId,
		}] = true
	}

	// Get the chapter containers
	for pair := range providerAndMediaIdPairs {
		provider := string(pair.provider)
		mediaId := pair.mediaId

		// Get the manga chapter container (downloaded and non-downloaded)
		container, err := r.GetMangaChapterContainer(provider, mediaId, nil)
		if err != nil {
			if errors.Is(err, ErrNoTitlesProvided) { // This means the cache has expired
				// Get the manga from the collection
				mangaEntry, ok := mangaCollection.GetListEntryFromMediaId(mediaId)
				if !ok {
					r.logger.Warn().Int("mediaId", mediaId).Msg("manga: [GetDownloadedChapterContainers] Manga not found in collection")
					continue
				}

				container, err = r.GetMangaChapterContainer(provider, mediaId, mangaEntry.GetMedia().GetAllTitles())
				if err != nil {
					r.logger.Error().Err(err).Msg("manga: [GetDownloadedChapterContainers] Failed to get chapter container")
					continue
				}
			} else {
				r.logger.Error().Err(err).Msg("manga: [GetDownloadedChapterContainers] Failed to get chapter container")
				continue
			}
		}

		// Now that we have the container, we'll filter out the chapters that are not downloaded
		chapters := make([]*hibikemanga.ChapterDetails, 0)
		// Go through each chapter and check if it's downloaded
		for _, chapter := range container.Chapters {
			// For each chapter, check if the chapter directory exists
			for _, dir := range chapterDirs {
				if dir == fmt.Sprintf("%s_%d_%s_%s", provider, mediaId, chapter.ID, chapter.Chapter) {
					chapters = append(chapters, chapter)
					break
				}
			}
		}

		if len(chapters) == 0 {
			continue
		}

		container.Chapters = chapters

		ret = append(ret, container)
	}

	return ret, nil
}
