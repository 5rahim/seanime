package handlers

import (
	"errors"
	"github.com/seanime-app/seanime/internal/anilist"
	"strconv"
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

func HandleEditAnilistListEntry(c *RouteCtx) error {

	type body struct {
		MediaId   *int                     `json:"mediaId"`
		Status    *anilist.MediaListStatus `json:"status"`
		Score     *int                     `json:"score"`
		Progress  *int                     `json:"progress"`
		StartDate *anilist.FuzzyDateInput  `json:"startedAt"`
		EndDate   *anilist.FuzzyDateInput  `json:"completedAt"`
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

	// Refresh the anilist collection
	_, _ = c.App.RefreshAnilistCollection()

	return c.RespondWithData(ret)
}

//----------------------------------------------------------------------------------------------------------------------

// HandleEditAnilistListEntryProgress
// DEPRECATED
func HandleEditAnilistListEntryProgress(c *RouteCtx) error {

	type body struct {
		MediaId  *int `json:"mediaId"`
		Progress *int `json:"progress"`
		Episodes *int `json:"episodes"`
	}

	p := new(body)
	if err := c.Fiber.BodyParser(p); err != nil {
		return c.RespondWithError(err)
	}

	// Update the progress
	err := c.App.AnilistClientWrapper.UpdateMediaListEntryProgress(
		c.Fiber.Context(),
		p.MediaId,
		p.Progress,
		p.Episodes,
	)
	if err != nil {
		return c.RespondWithError(err)
	}

	// Refresh the anilist collection
	_, _ = c.App.RefreshAnilistCollection()

	return c.RespondWithData(true)
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
		MediaId *int `json:"mediaId"`
	}

	p := new(body)
	if err := c.Fiber.BodyParser(p); err != nil {
		return c.RespondWithError(err)
	}

	// Get the list entry ID
	anilistCollection, err := c.App.GetAnilistCollection(false)
	if err != nil {
		return c.RespondWithError(err)
	}
	listEntry, found := anilistCollection.GetListEntryFromMediaId(*p.MediaId)
	if !found {
		return c.RespondWithError(errors.New("list entry not found"))

	}

	// Delete the list entry
	ret, err := c.App.AnilistClientWrapper.DeleteEntry(
		c.Fiber.Context(),
		&listEntry.ID,
	)
	if err != nil {
		return c.RespondWithError(err)
	}

	// Refresh the anilist collection
	_, _ = c.App.RefreshAnilistCollection()

	return c.RespondWithData(ret)
}
