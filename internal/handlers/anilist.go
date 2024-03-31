package handlers

import (
	"errors"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/util/result"
	"strconv"
	"time"
)

func HandleGetAnilistCollection(c *RouteCtx) error {

	bypassCache := c.Fiber.Method() == "POST"

	// Get the user's anilist collection
	anilistCollection, err := c.App.GetAnilistCollection(bypassCache)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(anilistCollection)

}

// HandleEditAnilistListEntry is used by the Anilist Media Entry Modal
//
// POST /v1/anilist/list-entry
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
		_, _ = c.App.RefreshAnilistCollection()
	case "manga":
		_, _ = c.App.RefreshMangaCollection()
	default:
		_, _ = c.App.RefreshAnilistCollection()
		_, _ = c.App.RefreshMangaCollection()
	}

	return c.RespondWithData(ret)
}

//----------------------------------------------------------------------------------------------------------------------

// HandleGetAnilistMediaDetails
// GET /v1/anilist/media-details/:id
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
		anilistCollection, err := c.App.GetAnilistCollection(false)
		if err != nil {
			return c.RespondWithError(err)
		}

		listEntry, found := anilistCollection.GetListEntryFromMediaId(*p.MediaId)
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
		_, _ = c.App.RefreshAnilistCollection()
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

// HandleAnilistListAnime is used by the "Discover" page
//
//	POST /v1/anilist/list-anime
func HandleAnilistListAnime(c *RouteCtx) error {

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
		c.App.Logger,
	)
	if err != nil {
		return c.RespondWithError(err)
	}

	anilistListMediaCache.SetT(cacheKey, ret, time.Minute*10)

	return c.RespondWithData(ret)
}

// HandleAnilistListRecentAiringAnime is used by the "Schedule" page
//
//	POST /v1/anilist/list-recent-anime
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
