package manga

import (
	"errors"
	"fmt"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/manga/providers"
	"github.com/seanime-app/seanime/internal/util"
	"sync"
)

var (
	ErrNoResults       = errors.New("no results found for this media")
	ErrNoChapters      = errors.New("no manga chapters found")
	ErrChapterNotFound = errors.New("chapter not found")
)

type (
	ChapterContainer struct {
		MediaId  int                               `json:"mediaId"`
		Provider string                            `json:"provider"`
		Chapters []*manga_providers.ChapterDetails `json:"chapters"`
	}

	PageContainer struct {
		MediaId        int                            `json:"mediaId"`
		Provider       string                         `json:"provider"`
		ChapterId      string                         `json:"chapterId"`
		Pages          []*manga_providers.ChapterPage `json:"pages"`
		PageDimensions map[int]PageDimension          `json:"pageDimensions"`
		IsDownloaded   bool                           `json:"isDownloaded"` // TODO
	}

	PageDimension struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	}
)

// GetMangaChapters returns the chapters for a manga entry based on the provider.
func (r *Repository) GetMangaChapters(provider manga_providers.Provider, mediaId int, titles []*string) (*ChapterContainer, error) {

	key := fmt.Sprintf("%s$%d", provider, mediaId)

	r.logger.Debug().
		Str("provider", string(provider)).
		Int("mediaId", mediaId).
		Str("key", key).
		Msgf("manga: getting chapters")

	var container *ChapterContainer

	bucket := r.getFcProviderBucket(provider, mediaId, bucketTypeChapter)

	// Check if the container is in the cache
	if found, _ := r.fileCacher.Get(bucket, key, &container); found {
		r.logger.Debug().Str("key", key).Msg("manga: Cache HIT")
		return container, nil
	}

	titles = lo.Filter(titles, func(title *string, _ int) bool {
		return util.IsMostlyLatinString(*title)
	})

	// 1. Search

	var searchRes []*manga_providers.SearchResult

	var err error
	for _, title := range titles {
		var _searchRes []*manga_providers.SearchResult
		switch provider {
		case manga_providers.ComickProvider:
			_searchRes, err = r.comick.Search(manga_providers.SearchOptions{
				Query: *title,
			})
		case manga_providers.MangaseeProvider:
			_searchRes, err = r.mangasee.Search(manga_providers.SearchOptions{
				Query: *title,
			})
		}
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

	// 2. Get chapters
	bestRes := searchRes[0]
	for _, res := range searchRes {
		if res.SearchRating > bestRes.SearchRating {
			bestRes = res
		}
	}

	var chapterList []*manga_providers.ChapterDetails

	switch provider {
	case manga_providers.ComickProvider:
		chapterList, err = r.comick.FindChapters(bestRes.ID)
	case manga_providers.MangaseeProvider:
		chapterList, err = r.mangasee.FindChapters(bestRes.ID)
	}

	if err != nil {
		r.logger.Error().Err(err).Msg("manga: find chapters failed")
		return nil, ErrNoChapters
	}

	container = &ChapterContainer{
		MediaId:  mediaId,
		Provider: string(provider),
		Chapters: chapterList,
	}

	// Set cache
	err = r.fileCacher.Set(bucket, key, container)
	if err != nil {
		r.logger.Warn().Err(err).Msg("manga: failed to set cache")
	}

	return container, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) GetMangaPageContainer(
	provider manga_providers.Provider,
	mediaId int,
	chapterId string,
	backup bool, // Whether to download/retrieve downloaded pages
) (*PageContainer, error) {

	if backup {

		r.RefreshBackups()

		var container *PageContainer

		foundDownloadedChapterId := false
		storedChapterIds, found := r.backupMap[DownloadID{Provider: string(provider), MediaID: mediaId}]
		if found {
			for _, storedChapterId := range storedChapterIds {
				if storedChapterId == chapterId {
					foundDownloadedChapterId = true
					break
				}
			}
		}

		pageList := make([]*manga_providers.ChapterPage, 0)

		//
		// Chapter is downloaded
		//
		if foundDownloadedChapterId {

			// Get the downloaded pages
			pageMap, err := r.downloader.getPageMap(string(provider), mediaId, chapterId, r.backupDir)
			if err != nil {
				r.logger.Error().Err(err).Msg("manga: failed to get downloaded pages")
				return nil, err
			}

			for _, pageInfo := range *pageMap {
				pageList = append(pageList, pageInfo.ToChapterPage())
			}

			pageDimensions := make(map[int]PageDimension)

			for _, pageInfo := range *pageMap {
				pageList = append(pageList, pageInfo.ToChapterPage())
				pageDimensions[pageInfo.Index] = PageDimension{
					Width:  pageInfo.Width,
					Height: pageInfo.Height,
				}
			}

			container = &PageContainer{
				MediaId:        mediaId,
				Provider:       string(provider),
				ChapterId:      chapterId,
				Pages:          pageList,
				PageDimensions: pageDimensions,
			}

			return container, nil

		} else {
			//
			// Chapter is not downloaded
			//

			// Get the chapter pages from the online source
			pc, err := r.GetMangaChapterPagesFromOnline(provider, mediaId, chapterId)
			if err != nil {
				r.logger.Error().Err(err).Msg("manga: failed to get online pages")
				return nil, err
			}

			// Download the images
			err = r.downloader.downloadImages(string(provider), mediaId, chapterId, pc.Pages, r.backupDir)
			if err != nil {
				r.logger.Error().Err(err).Msg("manga: failed to download images")
				return nil, err
			}

			// Get the downloaded pages
			pageMap, err := r.downloader.getPageMap(string(provider), mediaId, chapterId, r.backupDir)
			if err != nil {
				r.logger.Error().Err(err).Msg("manga: failed to get downloaded pages")
				return nil, err
			}

			pageDimensions := make(map[int]PageDimension)

			for _, pageInfo := range *pageMap {
				pageList = append(pageList, pageInfo.ToChapterPage())
				pageDimensions[pageInfo.Index] = PageDimension{
					Width:  pageInfo.Width,
					Height: pageInfo.Height,
				}
			}

			container = &PageContainer{
				MediaId:        mediaId,
				Provider:       string(provider),
				ChapterId:      chapterId,
				Pages:          pageList,
				PageDimensions: pageDimensions,
			}

			return container, nil

		}

	}

	return r.GetMangaChapterPagesFromOnline(provider, mediaId, chapterId)
}

// GetMangaChapterPagesFromOnline returns the pages for a manga chapter based on the provider.
func (r *Repository) GetMangaChapterPagesFromOnline(provider manga_providers.Provider, mediaId int, chapterId string) (*PageContainer, error) {

	key := fmt.Sprintf("%s$%d$%s", provider, mediaId, chapterId)

	r.logger.Debug().
		Str("provider", string(provider)).
		Int("mediaId", mediaId).
		Str("key", key).
		Str("chapterId", chapterId).
		Msgf("manga: getting pages")

	var container *PageContainer

	bucket := r.getFcProviderBucket(provider, mediaId, bucketTypePage)

	// Check if the container is in the cache
	if found, _ := r.fileCacher.Get(bucket, key, &container); found {
		return container, nil
	}

	// Search for the chapter in the cache
	chapterBucket := r.getFcProviderBucket(provider, mediaId, bucketTypeChapter)

	var chapterContainer *ChapterContainer
	if found, _ := r.fileCacher.Get(chapterBucket, fmt.Sprintf("%s$%d", provider, mediaId), &chapterContainer); !found {
		r.logger.Error().Msg("manga: chapter container not found")
		return nil, ErrNoChapters
	}

	// Get the chapter from the container
	var chapter *manga_providers.ChapterDetails
	for _, c := range chapterContainer.Chapters {
		if c.ID == chapterId {
			chapter = c
			break
		}
	}

	if chapter == nil {
		r.logger.Error().Msg("manga: chapter not found")
		return nil, ErrChapterNotFound
	}

	// Get the chapter pages
	var pageList []*manga_providers.ChapterPage
	var err error

	switch provider {
	case manga_providers.ComickProvider:
		pageList, err = r.comick.FindChapterPages(chapter.ID)
	case manga_providers.MangaseeProvider:
		pageList, err = r.mangasee.FindChapterPages(chapter.ID)
	}

	if err != nil {
		r.logger.Error().Err(err).Msg("manga: could not get chapter pages")
		return nil, err
	}

	// Get the page dimensions
	pageDimensions := make(map[int]PageDimension)
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	for _, page := range pageList {
		wg.Add(1)
		go func(page *manga_providers.ChapterPage) {
			defer wg.Done()
			width, height, err := getImageNaturalSize(page.URL)
			if err != nil {
				// DEVNOTE: Fails for Mangasee
				//r.logger.Warn().Err(err).Msg("manga: failed to get image size")
				return
			}

			mu.Lock()
			pageDimensions[page.Index] = PageDimension{
				Width:  width,
				Height: height,
			}
			mu.Unlock()
		}(page)
	}
	wg.Wait()

	container = &PageContainer{
		MediaId:        mediaId,
		Provider:       string(provider),
		ChapterId:      chapterId,
		Pages:          pageList,
		PageDimensions: nil,
	}

	// Set cache
	err = r.fileCacher.Set(bucket, key, container)
	if err != nil {
		r.logger.Warn().Err(err).Msg("manga: failed to set cache")
	}

	return container, nil
}
