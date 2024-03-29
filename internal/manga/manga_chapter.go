package manga

import (
	"errors"
	"fmt"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/manga/providers"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/seanime-app/seanime/internal/util/filecache"
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
		MediaId   int                            `json:"mediaId"`
		Provider  string                         `json:"provider"`
		ChapterId string                         `json:"chapterId"`
		Pages     []*manga_providers.ChapterPage `json:"pages"`
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

	var bucket filecache.Bucket
	switch provider {
	case manga_providers.ComickProvider:
		bucket = r.fcComicKChapterContainerBucket
	case manga_providers.MangaseeProvider:
		bucket = r.fcMangaseeChapterContainerBucket
	}

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

// GetMangaChapterPages returns the pages for a manga chapter based on the provider.
func (r *Repository) GetMangaChapterPages(provider manga_providers.Provider, mediaId int, chapterId string) (*PageContainer, error) {

	key := fmt.Sprintf("%s$%d$%s", provider, mediaId, chapterId)

	r.logger.Debug().
		Str("provider", string(provider)).
		Int("mediaId", mediaId).
		Str("key", key).
		Str("chapterId", chapterId).
		Msgf("manga: getting pages")

	var container *PageContainer

	var bucket filecache.Bucket
	switch provider {
	case manga_providers.ComickProvider:
		bucket = r.fcComicKPageContainerBucket
	case manga_providers.MangaseeProvider:
		bucket = r.fcMangaseePageContainerBucket
	}

	// Check if the container is in the cache
	if found, _ := r.fileCacher.Get(bucket, key, &container); found {
		return container, nil
	}

	// Search for the chapter in the cache
	var chapterBucket filecache.Bucket
	switch provider {
	case manga_providers.ComickProvider:
		chapterBucket = r.fcComicKChapterContainerBucket
	case manga_providers.MangaseeProvider:
		chapterBucket = r.fcMangaseeChapterContainerBucket
	}

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

	container = &PageContainer{
		MediaId:   mediaId,
		Provider:  string(provider),
		ChapterId: chapterId,
		Pages:     pageList,
	}

	// Set cache
	err = r.fileCacher.Set(bucket, key, container)
	if err != nil {
		r.logger.Warn().Err(err).Msg("manga: failed to set cache")
	}

	return nil, nil
}
