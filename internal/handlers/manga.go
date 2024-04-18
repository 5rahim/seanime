package handlers

import (
	"context"
	"errors"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/api/mal"
	"github.com/seanime-app/seanime/internal/core"
	"github.com/seanime-app/seanime/internal/events"
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

// HandleGetAnilistMangaCollection
//
//	@summary returns the user's AniList manga collection.
//	@route /api/v1/manga/anilist/collection [GET]
//	@returns anilist.MangaCollection
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

// HandleGetMangaCollection
//
//	@summary returns the user's main manga collection.
//	@desc This is an object that contains all the user's manga entries in a structured format.
//	@route /api/v1/manga/collection [GET]
//	@returns manga.Collection
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
//	@summary returns a manga entry for the given AniList manga id.
//	@desc This is used by the manga media entry pages to get all the data about the anime. It includes metadata and AniList list data.
//	@route /api/v1/manga/entry/{id} [GET]
//	@param id - int - true - "AniList manga media ID"
//	@returns manga.Entry
func HandleGetMangaEntry(c *RouteCtx) error {

	if err := checkMangaFlag(c.App); err != nil {
		return c.RespondWithError(err)
	}

	id, err := c.Fiber.ParamsInt("id")
	if err != nil {
		return c.RespondWithError(err)
	}

	anilistCollection, err := c.App.GetMangaCollection(false)
	if err != nil {
		return c.RespondWithError(err)
	}

	entry, err := manga.NewEntry(&manga.NewEntryOptions{
		MediaId:              id,
		Logger:               c.App.Logger,
		FileCacher:           c.App.FileCacher,
		AnilistClientWrapper: c.App.AnilistClientWrapper,
		MangaCollection:      anilistCollection,
	})
	if err != nil {
		return c.RespondWithError(err)
	}

	if entry != nil {
		baseMangaCache.SetT(entry.MediaId, entry.Media, 1*time.Hour)
	}

	return c.RespondWithData(entry)
}

// HandleGetMangaEntryDetails
//
//	@summary returns more details about an AniList manga entry.
//	@desc This fetches more fields omitted from the base queries.
//	@route /api/v1/manga/entry/{id}/details [GET]
//	@param id - int - true - "AniList manga media ID"
//	@returns anilist.MangaDetailsById_Media
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

// HandleEmptyMangaEntryCache
//
//	@summary empties the cache for a manga entry.
//	@desc This will empty the cache for a manga entry (chapter lists and pages), allowing the client to fetch fresh data.
//	@desc HandleGetMangaEntryChapters should be called after this to fetch the new chapter list.
//	@desc Returns 'true' if the operation was successful.
//	@route /api/v1/manga/entry/cache [DELETE]
//	@returns bool
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

// HandleGetMangaEntryChapters
//
//	@summary returns the chapters for a manga entry based on the provider.
//	@route /api/v1/manga/chapters [POST]
//	@returns manga.ChapterContainer
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

	container, err := c.App.MangaRepository.GetMangaChapterContainer(b.Provider, b.MediaId, titles)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(container)
}

// HandleGetMangaEntryPages
//
//	@summary returns the pages for a manga entry based on the provider and chapter id.
//	@desc This will return the pages for a manga chapter.
//	@desc If the app is offline and the chapter is not downloaded, it will return an error.
//	@desc If the app is online and the chapter is not downloaded, it will return the pages from the provider.
//	@desc If the chapter is downloaded, it will return the appropriate struct.
//	@desc If 'double page' is requested, it will fetch image sizes and include the dimensions in the response.
//	@route /api/v1/manga/pages [POST]
//	@returns manga.PageContainer
func HandleGetMangaEntryPages(c *RouteCtx) error {

	if err := checkMangaFlag(c.App); err != nil {
		return c.RespondWithError(err)
	}

	type body struct {
		MediaId    int                      `json:"mediaId"`
		Provider   manga_providers.Provider `json:"provider"`
		ChapterId  string                   `json:"chapterId"`
		DoublePage bool                     `json:"doublePage"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	container, err := c.App.MangaRepository.GetMangaPageContainer(b.Provider, b.MediaId, b.ChapterId, b.DoublePage, c.App.IsOffline())
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(container)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var (
	anilistListMangaCache = result.NewCache[string, *anilist.ListManga]()
)

// HandleAnilistListManga
//
//	@summary returns a list of manga based on the search parameters.
//	@desc This is used by "Advanced Search" and search function.
//	@route /api/v1/manga/anilist/list [POST]
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
		AverageScoreGreater *int                   `json:"averageScore_greater,omitempty"`
		Year                *int                   `json:"year,omitempty"`
		IsAdult             *bool                  `json:"isAdult,omitempty"`
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

	isAdult := false
	if p.IsAdult != nil {
		isAdult = *p.IsAdult && c.App.Settings.Anilist.EnableAdultContent
	}

	cacheKey := anilist.ListMediaCacheKey(
		p.Page,
		p.Search,
		p.PerPage,
		p.Sort,
		p.Status,
		p.Genres,
		p.AverageScoreGreater,
		nil,
		p.Year,
		p.Format,
		&isAdult,
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
		p.Year,
		p.Format,
		&isAdult,
		c.App.Logger,
	)
	if err != nil {
		return c.RespondWithError(err)
	}

	if ret != nil {
		anilistListMangaCache.SetT(cacheKey, ret, time.Minute*10)
	}

	return c.RespondWithData(ret)
}

// HandleUpdateMangaProgress
//
//	@summary updates the progress of a manga entry.
//	@desc Note: MyAnimeList is not supported
//	@route /api/v1/manga/update-progress [POST]
func HandleUpdateMangaProgress(c *RouteCtx) error {

	type body struct {
		MediaId       int `json:"mediaId"`
		MalId         int `json:"malId"`
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

	go func() {
		// Update the progress on MAL if an account is linked
		malInfo, _ := c.App.Database.GetMalInfo()
		if malInfo != nil && malInfo.AccessToken != "" && b.MalId > 0 {

			// Verify MAL auth
			malInfo, err = mal.VerifyMALAuth(malInfo, c.App.Database, c.App.Logger)
			if err != nil {
				c.App.WSEventManager.SendEvent(events.WarningToast, "Failed to update progress on MyAnimeList")
				return
			}

			client := mal.NewWrapper(malInfo.AccessToken, c.App.Logger)
			err = client.UpdateMangaProgress(&mal.MangaListProgressParams{
				NumChaptersRead: &b.ChapterNumber,
			}, b.MalId)
			if err != nil {
				c.App.WSEventManager.SendEvent(events.WarningToast, "Failed to update progress on MyAnimeList")
			}
		}
	}()

	return c.RespondWithData(true)
}
