package manga

import (
	"errors"
	"fmt"
	"seanime/internal/extension"
	manga_providers "seanime/internal/manga/providers"
	"seanime/internal/util"
	"sync"

	hibikemanga "seanime/internal/extension/hibike/manga"
)

type (
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

// GetMangaPageContainer returns the PageContainer for a manga chapter based on the provider.
func (r *Repository) GetMangaPageContainer(
	provider string,
	mediaId int,
	chapterId string,
	doublePage bool,
	isOffline *bool,
) (ret *PageContainer, err error) {
	defer util.HandlePanicInModuleWithError("manga/GetMangaPageContainer", &err)

	// +---------------------+
	// |      Downloads      |
	// +---------------------+

	providerExtension, ok := extension.GetExtension[extension.MangaProviderExtension](r.providerExtensionBank, provider)
	if !ok {
		r.logger.Error().Str("provider", provider).Msg("manga: Provider not found")
		return nil, errors.New("manga: Provider not found")
	}

	_, isLocalProvider := providerExtension.GetProvider().(*manga_providers.Local)

	if *isOffline && !isLocalProvider {
		ret, err = r.getDownloadedMangaPageContainer(provider, mediaId, chapterId)
		if err != nil {
			return nil, err
		}
		return ret, nil
	}

	if !isLocalProvider {
		ret, _ = r.getDownloadedMangaPageContainer(provider, mediaId, chapterId)
		if ret != nil {
			return ret, nil
		}
	}

	// +---------------------+
	// |      Get Pages      |
	// +---------------------+

	// PageContainer key
	pageContainerKey := fmt.Sprintf("%s$%d$%s", provider, mediaId, chapterId)

	r.logger.Trace().
		Str("provider", provider).
		Int("mediaId", mediaId).
		Str("key", pageContainerKey).
		Str("chapterId", chapterId).
		Msgf("manga: Getting pages")

	// +---------------------+
	// |       Cache         |
	// +---------------------+

	var container *PageContainer

	// PageContainer bucket
	// e.g., manga_comick_pages_123
	//         -> { "comick$123$10010": PageContainer }, { "comick$123$10011": PageContainer }
	pageBucket := r.getFcProviderBucket(provider, mediaId, bucketTypePage)

	// Check if the container is in the cache
	if found, _ := r.fileCacher.Get(pageBucket, pageContainerKey, &container); found && !isLocalProvider {

		// Hydrate page dimensions
		pageDimensions, _ := r.getPageDimensions(doublePage, provider, mediaId, chapterId, container.Pages)
		container.PageDimensions = pageDimensions

		r.logger.Debug().Str("key", pageContainerKey).Msg("manga: Page Container Cache HIT")
		return container, nil
	}

	// +---------------------+
	// |     Fetch pages     |
	// +---------------------+

	// Search for the chapter in the cache
	containerBucket := r.getFcProviderBucket(provider, mediaId, bucketTypeChapter)

	chapterContainerKey := getMangaChapterContainerCacheKey(provider, mediaId)

	var chapterContainer *ChapterContainer
	if found, _ := r.fileCacher.Get(containerBucket, chapterContainerKey, &chapterContainer); !found {
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

	// Get the chapter pages
	var pages []*hibikemanga.ChapterPage

	pages, err = providerExtension.GetProvider().FindChapterPages(chapter.ID)
	if err != nil {
		r.logger.Error().Err(err).Msg("manga: Could not get chapter pages")
		return nil, err
	}

	if pages == nil || len(pages) == 0 {
		r.logger.Error().Msg("manga: No pages found")
		return nil, fmt.Errorf("manga: No pages found")
	}

	// Overwrite provider just in case
	for _, page := range pages {
		page.Provider = provider
	}

	pageDimensions, _ := r.getPageDimensions(doublePage, provider, mediaId, chapterId, pages)

	container = &PageContainer{
		MediaId:        mediaId,
		Provider:       provider,
		ChapterId:      chapterId,
		Pages:          pages,
		PageDimensions: pageDimensions,
		IsDownloaded:   false,
	}

	// Set cache only if not local provider
	if !isLocalProvider {
		err = r.fileCacher.Set(pageBucket, pageContainerKey, container)
		if err != nil {
			r.logger.Warn().Err(err).Msg("manga: Failed to populate cache")
		}
	}

	r.logger.Debug().Str("key", pageContainerKey).Msg("manga: Retrieved pages")

	return container, nil
}

func (r *Repository) getPageDimensions(enabled bool, provider string, mediaId int, chapterId string, pages []*hibikemanga.ChapterPage) (ret map[int]*PageDimension, err error) {
	defer util.HandlePanicInModuleWithError("manga/getPageDimensions", &err)

	if !enabled {
		return nil, nil
	}

	// e.g. comick$123$10010
	key := fmt.Sprintf("%s$%d$%s", provider, mediaId, chapterId)

	// Page dimensions bucket
	// e.g., manga_comick_page-dimensions_123
	//         -> { "comick$123$10010": PageDimensions }, { "comick$123$10011": PageDimensions }
	dimensionBucket := r.getFcProviderBucket(provider, mediaId, bucketTypePageDimensions)

	if found, _ := r.fileCacher.Get(dimensionBucket, key, &ret); found {
		r.logger.Debug().Str("key", key).Msg("manga: Page Dimensions Cache HIT")
		return
	}

	r.logger.Trace().Str("key", key).Msg("manga: Getting page dimensions")

	// Get the page dimensions
	pageDimensions := make(map[int]*PageDimension)
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	for _, page := range pages {
		wg.Add(1)
		go func(page *hibikemanga.ChapterPage) {
			defer wg.Done()
			var buf []byte
			if page.Buf != nil {
				buf = page.Buf
			} else {
				buf, err = manga_providers.GetImageByProxy(page.URL, page.Headers)
				if err != nil {
					return
				}
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

	_ = r.fileCacher.Set(dimensionBucket, key, pageDimensions)

	r.logger.Info().Str("bucket", dimensionBucket.Name()).Msg("manga: Retrieved page dimensions")

	return pageDimensions, nil
}
