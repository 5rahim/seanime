package handlers

import (
	"seanime/internal/util"
	"strconv"

	"github.com/labstack/echo/v4"
)

// HandleSyncGetTrackedMediaItems
//
//	@summary gets all tracked media.
//	@route /api/v1/sync/track [GET]
//	@returns []sync.TrackedMediaItem
func (h *Handler) HandleSyncGetTrackedMediaItems(c echo.Context) error {
	tracked := h.App.SyncManager.GetTrackedMediaItems()
	return h.RespondWithData(c, tracked)
}

// HandleSyncAddMedia
//
//	@summary adds one or multiple media to be tracked for offline sync.
//	@route /api/v1/sync/track [POST]
//	@returns bool
func (h *Handler) HandleSyncAddMedia(c echo.Context) error {
	type body struct {
		Media []struct {
			MediaId int    `json:"mediaId"`
			Type    string `json:"type"`
		} `json:"media"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	var err error
	for _, m := range b.Media {
		switch m.Type {
		case "anime":
			err = h.App.SyncManager.AddAnime(m.MediaId)
		case "manga":
			err = h.App.SyncManager.AddManga(m.MediaId)
		}
	}

	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}

// HandleSyncRemoveMedia
//
//	@summary remove media from being tracked for offline sync.
//	@desc This will remove anime from being tracked for offline sync and delete any associated data.
//	@route /api/v1/sync/track [DELETE]
//	@returns bool
func (h *Handler) HandleSyncRemoveMedia(c echo.Context) error {
	type body struct {
		MediaId int    `json:"mediaId"`
		Type    string `json:"type"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	var err error
	switch b.Type {
	case "anime":
		err = h.App.SyncManager.RemoveAnime(b.MediaId)
	case "manga":
		err = h.App.SyncManager.RemoveManga(b.MediaId)
	}

	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}

// HandleSyncGetIsMediaTracked
//
//	@summary checks if media is being tracked for offline sync.
//	@route /api/v1/sync/track/{id}/{type} [GET]
//	@param id - int - true - "AniList anime media ID"
//	@param type - string - true - "Type of media (anime/manga)"
//	@returns bool
func (h *Handler) HandleSyncGetIsMediaTracked(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return h.RespondWithError(c, err)
	}

	kind := c.Param("type")
	tracked := h.App.SyncManager.IsMediaTracked(id, kind)

	return h.RespondWithData(c, tracked)
}

// HandleSyncLocalData
//
//	@summary syncs local data with AniList.
//	@route /api/v1/sync/local [POST]
//	@returns bool
func (h *Handler) HandleSyncLocalData(c echo.Context) error {
	err := h.App.SyncManager.SynchronizeLocal()
	if err != nil {
		return h.RespondWithError(c, err)
	}
	return h.RespondWithData(c, true)
}

// HandleSyncGetQueueState
//
//	@summary gets the current sync queue state.
//	@desc This will return the list of media that are currently queued for syncing.
//	@route /api/v1/sync/queue [GET]
//	@returns sync.QueueState
func (h *Handler) HandleSyncGetQueueState(c echo.Context) error {
	return h.RespondWithData(c, h.App.SyncManager.GetQueue().GetQueueState())
}

// HandleSyncAnilistData
//
//	@summary syncs AniList data with local.
//	@route /api/v1/sync/anilist [POST]
//	@returns bool
func (h *Handler) HandleSyncAnilistData(c echo.Context) error {
	err := h.App.SyncManager.SynchronizeAnilist()
	if err != nil {
		return h.RespondWithError(c, err)
	}
	return h.RespondWithData(c, true)
}

// HandleSyncSetHasLocalChanges
//
//	@summary sets the flag to determine if there are local changes that need to be synced with AniList.
//	@route /api/v1/sync/updated [POST]
//	@returns bool
func (h *Handler) HandleSyncSetHasLocalChanges(c echo.Context) error {
	type body struct {
		Updated bool `json:"updated"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	h.App.SyncManager.SetHasLocalChanges(b.Updated)
	return h.RespondWithData(c, true)
}

// HandleSyncGetHasLocalChanges
//
//	@summary gets the flag to determine if there are local changes that need to be synced with AniList.
//	@route /api/v1/sync/updated [GET]
//	@returns bool
func (h *Handler) HandleSyncGetHasLocalChanges(c echo.Context) error {
	updated := h.App.SyncManager.HasLocalChanges()
	return h.RespondWithData(c, updated)
}

// HandleSyncGetLocalStorageSize
//
//	@summary gets the size of the local storage in a human-readable format.
//	@route /api/v1/sync/storage/size [GET]
//	@returns string
func (h *Handler) HandleSyncGetLocalStorageSize(c echo.Context) error {
	size := h.App.SyncManager.GetLocalStorageSize()
	return h.RespondWithData(c, util.Bytes(uint64(size)))
}
