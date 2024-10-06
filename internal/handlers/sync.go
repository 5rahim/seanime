package handlers

// HandleSyncGetTrackedMediaItems
//
//	@summary gets all tracked media.
//	@route /api/v1/sync/track [GET]
//	@returns []sync.TrackedMediaItem
func HandleSyncGetTrackedMediaItems(c *RouteCtx) error {
	tracked := c.App.SyncManager.GetTrackedMediaItems()
	return c.RespondWithData(tracked)
}

// HandleSyncAddMedia
//
//	@summary adds a media to be tracked for offline sync.
//	@route /api/v1/sync/track [POST]
//	@returns bool
func HandleSyncAddMedia(c *RouteCtx) error {
	type body struct {
		MediaId int    `json:"mediaId"`
		Type    string `json:"type"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	var err error
	switch b.Type {
	case "anime":
		err = c.App.SyncManager.AddAnime(b.MediaId)
	case "manga":
		err = c.App.SyncManager.AddManga(b.MediaId)
	}

	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}

// HandleSyncRemoveMedia
//
//	@summary remove media from being tracked for offline sync.
//	@desc This will remove anime from being tracked for offline sync and delete any associated data.
//	@route /api/v1/sync/track [DELETE]
//	@returns bool
func HandleSyncRemoveMedia(c *RouteCtx) error {
	type body struct {
		MediaId int    `json:"mediaId"`
		Type    string `json:"type"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	var err error
	switch b.Type {
	case "anime":
		err = c.App.SyncManager.RemoveAnime(b.MediaId)
	case "manga":
		err = c.App.SyncManager.RemoveManga(b.MediaId)
	}

	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}

// HandleSyncGetIsMediaTracked
//
//	@summary checks if media is being tracked for offline sync.
//	@route /api/v1/sync/track/{id}/{type} [GET]
//	@param id - int - true - "AniList anime media ID"
//	@param type - string - true - "Type of media (anime/manga)"
//	@returns bool
func HandleSyncGetIsMediaTracked(c *RouteCtx) error {
	id, err := c.Fiber.ParamsInt("id")
	if err != nil {
		return c.RespondWithError(err)
	}

	kind := c.Fiber.Params("type")
	tracked := c.App.SyncManager.IsMediaTracked(id, kind)

	return c.RespondWithData(tracked)
}

// HandleSyncLocalData
//
//	@summary syncs local data with AniList.
//	@route /api/v1/sync/local [POST]
//	@returns bool
func HandleSyncLocalData(c *RouteCtx) error {
	err := c.App.SyncManager.SynchronizeLocal()
	if err != nil {
		return c.RespondWithError(err)
	}
	return c.RespondWithData(true)
}

// HandleSyncGetQueueState
//
//	@summary gets the current sync queue state.
//	@desc This will return the list of media that are currently queued for syncing.
//	@route /api/v1/sync/queue [GET]
//	@returns sync.QueueState
func HandleSyncGetQueueState(c *RouteCtx) error {
	return c.RespondWithData(c.App.SyncManager.GetQueue().GetQueueState())
}

// HandleSyncAnilistData
//
//	@summary syncs AniList data with local.
//	@route /api/v1/sync/anilist [POST]
//	@returns bool
func HandleSyncAnilistData(c *RouteCtx) error {
	err := c.App.SyncManager.SynchronizeAnilist()
	if err != nil {
		return c.RespondWithError(err)
	}
	return c.RespondWithData(true)
}

// HandleSyncSetHasLocalChanges
//
//	@summary sets the flag to determine if there are local changes that need to be synced with AniList.
//	@route /api/v1/sync/updated [POST]
//	@returns bool
func HandleSyncSetHasLocalChanges(c *RouteCtx) error {
	type body struct {
		Updated bool `json:"updated"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	c.App.SyncManager.SetHasLocalChanges(b.Updated)
	return c.RespondWithData(true)
}

// HandleSyncGetHasLocalChanges
//
//	@summary gets the flag to determine if there are local changes that need to be synced with AniList.
//	@route /api/v1/sync/updated [GET]
//	@returns bool
func HandleSyncGetHasLocalChanges(c *RouteCtx) error {
	updated := c.App.SyncManager.HasLocalChanges()
	return c.RespondWithData(updated)
}
