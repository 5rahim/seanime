package handlers

import "seanime/internal/continuity"

// HandleUpdateContinuityWatchHistoryItem
//
//	@summary Updates watch history item.
//	@desc This endpoint is used to update a watch history item.
//	@desc Since this is low priority, we ignore any errors.
//	@route /api/v1/continuity/item [PATCH]
//	@returns bool
func HandleUpdateContinuityWatchHistoryItem(c *RouteCtx) error {
	type body struct {
		Options continuity.UpdateWatchHistoryItemOptions `json:"options"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	err := c.App.ContinuityManager.UpdateWatchHistoryItem(&b.Options)
	if err != nil {
		// Ignore the error
		return c.RespondWithData(false)
	}

	return c.RespondWithData(true)
}

// HandleGetContinuityWatchHistoryItem
//
//	@summary Returns a watch history item.
//	@desc This endpoint is used to retrieve a watch history item.
//	@route /api/v1/continuity/item/{id} [GET]
//	@param id - int - true - "AniList anime media ID"
//	@returns continuity.WatchHistoryItemResponse
func HandleGetContinuityWatchHistoryItem(c *RouteCtx) error {
	id, err := c.Fiber.ParamsInt("id")
	if err != nil {
		return c.RespondWithError(err)
	}

	if !c.App.ContinuityManager.GetSettings().WatchContinuityEnabled {
		return c.RespondWithData(&continuity.WatchHistoryItemResponse{
			Item:  nil,
			Found: false,
		})
	}

	resp := c.App.ContinuityManager.GetWatchHistoryItem(id)
	return c.RespondWithData(resp)
}

// HandleGetContinuityWatchHistory
//
//	@summary Returns the continuity watch history
//	@desc This endpoint is used to retrieve all watch history items.
//	@route /api/v1/continuity/history [GET]
//	@returns continuity.WatchHistory
func HandleGetContinuityWatchHistory(c *RouteCtx) error {
	if !c.App.ContinuityManager.GetSettings().WatchContinuityEnabled {
		ret := make(map[int]*continuity.WatchHistoryItem)
		return c.RespondWithData(ret)
	}

	resp := c.App.ContinuityManager.GetWatchHistory()
	return c.RespondWithData(resp)
}
