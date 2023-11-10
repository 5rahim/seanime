package handlers

import "github.com/seanime-app/seanime-server/internal/anilist"

func HandleGetAnilistCollection(c *RouteCtx) error {

	// Get the user's anilist collection
	anilistCollection, err := c.App.GetAnilistCollection()
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

	ret, err := c.App.AnilistClient.UpdateMediaListEntry(
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
