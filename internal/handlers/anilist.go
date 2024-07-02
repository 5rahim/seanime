package handlers

import (
	"errors"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/util/result"
	"strconv"
	"time"
)

// HandleGetAnimeCollection
//
//	@summary returns the user's AniList anime collection.
//	@desc Calling GET will return the cached anime collection.
//	@desc The manga collection is also refreshed in the background, and upon completion, a WebSocket event is sent.
//	@desc Calling POST will refetch both the anime and manga collections.
//	@returns anilist.AnimeCollection
//	@route /api/v1/anilist/collection [GET,POST]
func HandleGetAnimeCollection(c *RouteCtx) error {

	bypassCache := c.Fiber.Method() == "POST"

	// Get the user's anilist collection
	animeCollection, err := c.App.GetAnimeCollection(bypassCache)
	if err != nil {
		return c.RespondWithError(err)
	}

	go func() {
		if c.App.Settings == nil {
			return
		}
		if c.App.Settings.Library.EnableManga {
			_, _ = c.App.GetMangaCollection(bypassCache)
			if bypassCache {
				c.App.WSEventManager.SendEvent(events.RefreshedAnilistMangaCollection, nil)
			}
		}
	}()

	return c.RespondWithData(animeCollection)
}

// HandleGetRawAnimeCollection
//
//	@summary returns the user's AniList anime collection without filtering out custom lists.
//	@desc Calling GET will return the cached anime collection.
//	@returns anilist.AnimeCollection
//	@route /api/v1/anilist/collection/raw [GET,POST]
func HandleGetRawAnimeCollection(c *RouteCtx) error {

	bypassCache := c.Fiber.Method() == "POST"

	// Get the user's anilist collection
	animeCollection, err := c.App.GetRawAnimeCollection(bypassCache)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(animeCollection)
}

// HandleEditAnilistListEntry
//
//	@summary updates the user's list entry on Anilist.
//	@desc This is used to edit an entry on AniList.
//	@desc The "type" field is used to determine if the entry is an anime or manga and refreshes the collection accordingly.
//	@desc The client should refetch collection-dependent queries after this mutation.
//	@returns anilist.UpdateMediaListEntry
//	@route /api/v1/anilist/list-entry [POST]
func HandleEditAnilistListEntry(c *RouteCtx) error {

	type body struct {
		MediaId   *int                     `json:"mediaId"`
		Status    *anilist.MediaListStatus `json:"status"`
		Score     *int                     `json:"score"`
		Progress  *int                     `json:"progress"`
		StartDate *anilist.FuzzyDateInput  `json:"startedAt"`
		EndDate   *anilist.FuzzyDateInput  `json:"completedAt"`
		Type      string                   `json:"type"`
	}

	p := new(body)
	if err := c.Fiber.BodyParser(p); err != nil {
		return c.RespondWithError(err)
	}

	ret, err := c.App.AnilistClientWrapper.UpdateMediaListEntry(
		c.Fiber.Context(),
		p.MediaId,
		p.Status,
		p.Score,
		p.Progress,
		p.StartDate,
		p.EndDate,
	)
	if err != nil {
		return c.RespondWithError(err)
	}

	switch p.Type {
	case "anime":
		_, _ = c.App.RefreshAnimeCollection()
	case "manga":
		_, _ = c.App.RefreshMangaCollection()
	default:
		_, _ = c.App.RefreshAnimeCollection()
		_, _ = c.App.RefreshMangaCollection()
	}

	return c.RespondWithData(ret)
}

//----------------------------------------------------------------------------------------------------------------------

var (
	detailsCache = result.NewCache[int, *anilist.MediaDetailsById_Media]()
)

// HandleGetAnilistMediaDetails
//
//	@summary returns more details about an AniList anime entry.
//	@desc This fetches more fields omitted from the base queries.
//	@param id - int - true - "The AniList anime ID"
//	@returns anilist.MediaDetailsById_Media
//	@route /api/v1/anilist/media-details/{id} [GET]
func HandleGetAnilistMediaDetails(c *RouteCtx) error {

	mId, err := strconv.Atoi(c.Fiber.Params("id"))
	if err != nil {
		return c.RespondWithError(err)
	}

	if details, ok := detailsCache.Get(mId); ok {
		return c.RespondWithData(details)
	}
	details, err := c.App.AnilistClientWrapper.MediaDetailsByID(c.Fiber.Context(), &mId)
	if err != nil {
		return c.RespondWithError(err)
	}
	detailsCache.Set(mId, details.GetMedia())

	return c.RespondWithData(details.GetMedia())
}

//----------------------------------------------------------------------------------------------------------------------

var studioDetailsMap = result.NewResultMap[int, *anilist.StudioDetails]()

// HandleGetAnilistStudioDetails
//
//	@summary returns details about a studio.
//	@desc This fetches media produced by the studio.
//	@param id - int - true - "The AniList studio ID"
//	@returns anilist.StudioDetails
//	@route /api/v1/anilist/studio-details/{id} [GET]
func HandleGetAnilistStudioDetails(c *RouteCtx) error {

	mId, err := strconv.Atoi(c.Fiber.Params("id"))
	if err != nil {
		return c.RespondWithError(err)
	}

	if details, ok := studioDetailsMap.Get(mId); ok {
		return c.RespondWithData(details)
	}
	details, err := c.App.AnilistClientWrapper.StudioDetails(c.Fiber.Context(), &mId)
	if err != nil {
		return c.RespondWithError(err)
	}

	go func() {
		if details != nil {
			studioDetailsMap.Set(mId, details)
		}
	}()

	return c.RespondWithData(details)
}

//----------------------------------------------------------------------------------------------------------------------

// HandleDeleteAnilistListEntry
//
//	@summary deletes an entry from the user's AniList list.
//	@desc This is used to delete an entry on AniList.
//	@desc The "type" field is used to determine if the entry is an anime or manga and refreshes the collection accordingly.
//	@desc The client should refetch collection-dependent queries after this mutation.
//	@route /api/v1/anilist/list-entry [DELETE]
//	@returns anilist.DeleteEntry
func HandleDeleteAnilistListEntry(c *RouteCtx) error {

	type body struct {
		MediaId *int    `json:"mediaId"`
		Type    *string `json:"type"`
	}

	p := new(body)
	if err := c.Fiber.BodyParser(p); err != nil {
		return c.RespondWithError(err)
	}

	if p.Type == nil || p.MediaId == nil {
		return c.RespondWithError(errors.New("missing parameters"))
	}

	var listEntryID int

	switch *p.Type {
	case "anime":
		// Get the list entry ID
		animeCollection, err := c.App.GetAnimeCollection(false)
		if err != nil {
			return c.RespondWithError(err)
		}

		listEntry, found := animeCollection.GetListEntryFromMediaId(*p.MediaId)
		if !found {
			return c.RespondWithError(errors.New("list entry not found"))
		}
		listEntryID = listEntry.ID
	case "manga":
		// Get the list entry ID
		mangaCollection, err := c.App.GetMangaCollection(false)
		if err != nil {
			return c.RespondWithError(err)
		}

		listEntry, found := mangaCollection.GetListEntryFromMediaId(*p.MediaId)
		if !found {
			return c.RespondWithError(errors.New("list entry not found"))
		}
		listEntryID = listEntry.ID
	}

	// Delete the list entry
	ret, err := c.App.AnilistClientWrapper.DeleteEntry(
		c.Fiber.Context(),
		&listEntryID,
	)
	if err != nil {
		return c.RespondWithError(err)
	}

	switch *p.Type {
	case "anime":
		_, _ = c.App.RefreshAnimeCollection()
	case "manga":
		_, _ = c.App.RefreshMangaCollection()
	}

	return c.RespondWithData(ret)
}

//----------------------------------------------------------------------------------------------------------------------

var (
	anilistListMediaCache       = result.NewCache[string, *anilist.ListMedia]()
	anilistListRecentMediaCache = result.NewCache[string, *anilist.ListRecentMedia]() // holds 1 value
)

// HandleAnilistListAnime
//
//	@summary returns a list of anime based on the search parameters.
//	@desc This is used by the "Discover" and "Advanced Search".
//	@route /api/v1/anilist/list-anime [POST]
//	@returns anilist.ListMedia
func HandleAnilistListAnime(c *RouteCtx) error {

	type body struct {
		Page                *int                   `json:"page,omitempty"`
		Search              *string                `json:"search,omitempty"`
		PerPage             *int                   `json:"perPage,omitempty"`
		Sort                []*anilist.MediaSort   `json:"sort,omitempty"`
		Status              []*anilist.MediaStatus `json:"status,omitempty"`
		Genres              []*string              `json:"genres,omitempty"`
		AverageScoreGreater *int                   `json:"averageScore_greater,omitempty"`
		Season              *anilist.MediaSeason   `json:"season,omitempty"`
		SeasonYear          *int                   `json:"seasonYear,omitempty"`
		Format              *anilist.MediaFormat   `json:"format,omitempty"`
		IsAdult             *bool                  `json:"isAdult,omitempty"`
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
		p.Season,
		p.SeasonYear,
		p.Format,
		&isAdult,
	)

	cached, ok := anilistListMediaCache.Get(cacheKey)
	if ok {
		return c.RespondWithData(cached)
	}

	ret, err := anilist.ListMediaM(
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
		&isAdult,
		c.App.Logger,
	)
	if err != nil {
		return c.RespondWithError(err)
	}

	if ret != nil {
		anilistListMediaCache.SetT(cacheKey, ret, time.Minute*10)
	}

	return c.RespondWithData(ret)
}

// HandleAnilistListRecentAiringAnime
//
//	@summary returns a list of recently aired anime.
//	@desc This is used by the "Schedule" page to display recently aired anime.
//	@route /api/v1/anilist/list-recent-anime [POST]
//	@returns anilist.ListRecentMedia
func HandleAnilistListRecentAiringAnime(c *RouteCtx) error {

	type body struct {
		Page            *int    `json:"page,omitempty"`
		Search          *string `json:"search,omitempty"`
		PerPage         *int    `json:"perPage,omitempty"`
		AiringAtGreater *int    `json:"airingAt_greater,omitempty"`
		AiringAtLesser  *int    `json:"airingAt_lesser,omitempty"`
	}

	p := new(body)
	if err := c.Fiber.BodyParser(p); err != nil {
		return c.RespondWithError(err)
	}

	if p.Page == nil || p.PerPage == nil {
		*p.Page = 1
		*p.PerPage = 50
	}

	cacheKey := "recent_airing_anime"

	cached, ok := anilistListRecentMediaCache.Get(cacheKey)
	if ok {
		return c.RespondWithData(cached)
	}

	ret, err := anilist.ListRecentAiringMediaM(
		p.Page,
		p.Search,
		p.PerPage,
		p.AiringAtGreater,
		p.AiringAtLesser,
		c.App.Logger,
	)
	if err != nil {
		return c.RespondWithError(err)
	}

	anilistListRecentMediaCache.SetT(cacheKey, ret, time.Minute*10)

	return c.RespondWithData(ret)
}

var anilistStatsCache = result.NewCache[int, *anilist.Stats]()

// HandleGetAniListStats
//
//	@summary returns the anilist stats.
//	@desc This returns the AniList stats for the user.
//	@route /api/v1/anilist/stats [GET]
//	@returns anilist.Stats
func HandleGetAniListStats(c *RouteCtx) error {
	cached, ok := anilistStatsCache.Get(0)
	if ok {
		return c.RespondWithData(cached)
	}

	ret, err := anilist.GetStats(
		c.Fiber.Context(),
		c.App.AnilistClientWrapper,
	)
	if err != nil {
		return c.RespondWithError(err)
	}

	anilistStatsCache.SetT(0, ret, time.Hour*1)

	return c.RespondWithData(ret)
}
