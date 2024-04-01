package handlers

import (
	"context"
	"errors"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/core"
	"github.com/seanime-app/seanime/internal/manga"
	"github.com/seanime-app/seanime/internal/manga/providers"
	"github.com/seanime-app/seanime/internal/util/result"
	"time"
)

var (
	ErrMangaFeatureDisabled = errors.New("manga feature not enabled")
	baseMangaCache          = result.NewCache[int, *anilist.BaseManga]()
	mangaDetailsCache       = result.NewCache[int, *anilist.MangaDetailsById_Media]()
)

func checkMangaFlag(a *core.App) error {
	if !a.Settings.Library.EnableManga {
		return ErrMangaFeatureDisabled
	}

	return nil
}

// HandleGetAnilistMangaCollection return the user's Anilist manga collection.
//
//	POST /api/v1/manga/anilist/collection
func HandleGetAnilistMangaCollection(c *RouteCtx) error {

	type body struct {
		BypassCache bool `json:"bypassCache"`
	}

	if err := checkMangaFlag(c.App); err != nil {
		return c.RespondWithError(err)
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	collection, err := c.App.GetMangaCollection(b.BypassCache)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(collection)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// HandleGetMangaCollection return the user's manga collection.
//
//	GET /api/v1/manga/collection
func HandleGetMangaCollection(c *RouteCtx) error {

	if err := checkMangaFlag(c.App); err != nil {
		return c.RespondWithError(err)
	}

	anilistCollection, err := c.App.GetMangaCollection(false)
	if err != nil {
		return c.RespondWithError(err)
	}

	collection, err := manga.NewCollection(&manga.NewCollectionOptions{
		MangaCollection:      anilistCollection,
		AnilistClientWrapper: c.App.AnilistClientWrapper,
	})
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(collection)
}

// HandleGetMangaEntry
//
//	GET /api/v1/manga/entry/:id
func HandleGetMangaEntry(c *RouteCtx) error {

	if err := checkMangaFlag(c.App); err != nil {
		return c.RespondWithError(err)
	}

	id, err := c.Fiber.ParamsInt("id")
	if err != nil {
		return c.RespondWithError(err)
	}

	collection, err := c.App.GetMangaCollection(false)
	if err != nil {
		return c.RespondWithError(err)
	}

	entry, err := manga.NewEntry(&manga.NewEntryOptions{
		MediaId:              id,
		Logger:               c.App.Logger,
		FileCacher:           c.App.FileCacher,
		AnilistClientWrapper: c.App.AnilistClientWrapper,
		MangaCollection:      collection,
	})
	if err != nil {
		return c.RespondWithError(err)
	}

	baseMangaCache.SetT(entry.MediaId, entry.Media, time.Hour)

	return c.RespondWithData(entry)
}

// HandleGetMangaEntryDetails return additional details for a manga entry.
//
//	GET /api/v1/manga/entry/:id/details
func HandleGetMangaEntryDetails(c *RouteCtx) error {

	if err := checkMangaFlag(c.App); err != nil {
		return c.RespondWithError(err)
	}

	id, err := c.Fiber.ParamsInt("id")
	if err != nil {
		return c.RespondWithError(err)
	}

	if detailsMedia, found := mangaDetailsCache.Get(id); found {
		return c.RespondWithData(detailsMedia)
	}

	details, err := c.App.AnilistClientWrapper.MangaDetailsByID(context.Background(), &id)
	if err != nil {
		return c.RespondWithError(err)
	}

	mangaDetailsCache.SetT(id, details.GetMedia(), time.Hour)

	return c.RespondWithData(details.GetMedia())
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// HandleEmptyMangaEntryCache will empty the cache for a manga entry.
// HandleGetMangaEntryChapters should be called after this to refresh the client.
//
//	DELETE /api/v1/manga/entry/cache
func HandleEmptyMangaEntryCache(c *RouteCtx) error {

	if err := checkMangaFlag(c.App); err != nil {
		return c.RespondWithError(err)
	}

	type body struct {
		MediaId int `json:"mediaId"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	err := c.App.MangaRepository.EmptyMangaCache(b.MediaId)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}

// HandleGetMangaEntryChapters return the chapters for a manga entry based on the provider.
//
//	POST /api/v1/manga/entry/:id/chapters
func HandleGetMangaEntryChapters(c *RouteCtx) error {

	if err := checkMangaFlag(c.App); err != nil {
		return c.RespondWithError(err)
	}

	type body struct {
		MediaId  int                      `json:"mediaId"`
		Provider manga_providers.Provider `json:"provider"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	var titles []*string
	baseManga, found := baseMangaCache.Get(b.MediaId)
	if !found {
		mangaF, err := c.App.AnilistClientWrapper.BaseMangaByID(context.Background(), &b.MediaId)
		if err != nil {
			return c.RespondWithError(err)
		}
		titles = mangaF.GetMedia().GetAllTitles()
	} else {
		titles = baseManga.GetAllTitles()
	}

	container, err := c.App.MangaRepository.GetMangaChapters(b.Provider, b.MediaId, titles)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(container)
}

// HandleGetMangaEntryPages return the pages for a manga entry chapter based on the provider.
//
//	POST /api/v1/manga/pages
func HandleGetMangaEntryPages(c *RouteCtx) error {

	if err := checkMangaFlag(c.App); err != nil {
		return c.RespondWithError(err)
	}

	type body struct {
		MediaId   int                      `json:"mediaId"`
		Provider  manga_providers.Provider `json:"provider"`
		ChapterId string                   `json:"chapterId"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	container, err := c.App.MangaRepository.GetMangaChapterPagesFromOnline(b.Provider, b.MediaId, b.ChapterId)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(container)
}

// HandleGetMangaEntryPageContainer return the pages for a manga entry chapter based on the provider.
// FIXME SHELVED
//
//	POST /api/v1/manga/pages
func HandleGetMangaEntryPageContainer(c *RouteCtx) error {

	if err := checkMangaFlag(c.App); err != nil {
		return c.RespondWithError(err)
	}

	type body struct {
		MediaId    int                      `json:"mediaId"`
		Provider   manga_providers.Provider `json:"provider"`
		ChapterId  string                   `json:"chapterId"`
		Downloaded bool                     `json:"downloaded"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	container, err := c.App.MangaRepository.GetMangaPageContainer(b.Provider, b.MediaId, b.ChapterId, true)
	if err != nil {
		container, err = c.App.MangaRepository.GetMangaChapterPagesFromOnline(b.Provider, b.MediaId, b.ChapterId)
		if err != nil {
			return c.RespondWithError(err)
		}
	}

	return c.RespondWithData(container)
}

// HandleDownloadMangaChapter download the pages for a manga entry chapter based on the provider.
// FIXME SHELVED
//
//	POST /api/v1/manga/download-chapter
func HandleDownloadMangaChapter(c *RouteCtx) error {

	if err := checkMangaFlag(c.App); err != nil {
		return c.RespondWithError(err)
	}

	type body struct {
		MediaId   int                      `json:"mediaId"`
		Provider  manga_providers.Provider `json:"provider"`
		ChapterId string                   `json:"chapterId"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	err := c.App.MangaRepository.DownloadMangaChapter(b.Provider, b.MediaId, b.ChapterId)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}

func HandleGetMangaEntryBackups(c *RouteCtx) error {

	if err := checkMangaFlag(c.App); err != nil {
		return c.RespondWithError(err)
	}

	type body struct {
		MediaId  int                      `json:"mediaId"`
		Provider manga_providers.Provider `json:"provider"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	backupContainer := c.App.MangaRepository.GetMangaEntryBackups(b.Provider, b.MediaId)

	return c.RespondWithData(backupContainer)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var (
	anilistListMangaCache = result.NewCache[string, *anilist.ListManga]()
)

func HandleAnilistListManga(c *RouteCtx) error {

	if err := checkMangaFlag(c.App); err != nil {
		return c.RespondWithError(err)
	}

	type body struct {
		Page                *int                   `json:"page,omitempty"`
		Search              *string                `json:"search,omitempty"`
		PerPage             *int                   `json:"perPage,omitempty"`
		Sort                []*anilist.MediaSort   `json:"sort,omitempty"`
		Status              []*anilist.MediaStatus `json:"status,omitempty"`
		Genres              []*string              `json:"genres,omitempty"`
		AverageScoreGreater *int                   `json:"averageScoreGreater,omitempty"`
		Season              *anilist.MediaSeason   `json:"season,omitempty"`
		SeasonYear          *int                   `json:"seasonYear,omitempty"`
		Format              *anilist.MediaFormat   `json:"format,omitempty"`
	}

	p := new(body)
	if err := c.Fiber.BodyParser(p); err != nil {
		return c.RespondWithError(err)
	}

	if p.Page == nil || p.PerPage == nil {
		*p.Page = 1
		*p.PerPage = 20
	}

	cacheKey := anilist.ListMediaCacheKey(
		p.Page,
		p.Search,
		p.PerPage,
		p.Sort,
		p.Status,
		p.Genres,
		p.AverageScoreGreater,
		p.Season,
		p.SeasonYear,
		p.Format,
	)

	cached, ok := anilistListMangaCache.Get(cacheKey)
	if ok {
		return c.RespondWithData(cached)
	}

	ret, err := anilist.ListMangaM(
		p.Page,
		p.Search,
		p.PerPage,
		p.Sort,
		p.Status,
		p.Genres,
		p.AverageScoreGreater,
		p.Season,
		p.SeasonYear,
		p.Format,
		c.App.Logger,
	)
	if err != nil {
		return c.RespondWithError(err)
	}

	anilistListMangaCache.SetT(cacheKey, ret, time.Minute*10)

	return c.RespondWithData(ret)
}

// HandleUpdateMangaProgress will update the progress of the given media entry.
//
// DEVOTE: MyAnimeList is not supported
//
//	POST /v1/manga/update-progress
func HandleUpdateMangaProgress(c *RouteCtx) error {

	type body struct {
		MediaId       int `json:"mediaId"`
		ChapterNumber int `json:"chapterNumber"`
		TotalChapters int `json:"totalChapters"`
	}

	b := new(body)
	if err := c.Fiber.BodyParser(b); err != nil {
		return c.RespondWithError(err)
	}

	// Update the progress on AniList
	err := c.App.AnilistClientWrapper.UpdateMediaListEntryProgress(
		context.Background(),
		&b.MediaId,
		&b.ChapterNumber,
		&b.TotalChapters,
	)
	if err != nil {
		return c.RespondWithError(err)
	}

	_, _ = c.App.RefreshMangaCollection() // Refresh the AniList collection

	return c.RespondWithData(true)
}
