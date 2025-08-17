package handlers

import (
	"seanime/internal/util"
	"strconv"

	"github.com/labstack/echo/v4"
)

// HandleSetOfflineMode
//
//	@summary sets the offline mode.
//	@desc Returns true if the offline mode is active, false otherwise.
//	@route /api/v1/local/offline [POST]
//	@returns bool
func (h *Handler) HandleSetOfflineMode(c echo.Context) error {
	type body struct {
		Enabled bool `json:"enabled"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	h.App.SetOfflineMode(b.Enabled)
	return h.RespondWithData(c, b.Enabled)
}

// HandleLocalGetTrackedMediaItems
//
//	@summary gets all tracked media.
//	@route /api/v1/local/track [GET]
//	@returns []local.TrackedMediaItem
func (h *Handler) HandleLocalGetTrackedMediaItems(c echo.Context) error {
	tracked := h.App.LocalManager.GetTrackedMediaItems()
	return h.RespondWithData(c, tracked)
}

// HandleLocalAddTrackedMedia
//
//	@summary adds one or multiple media to be tracked for offline sync.
//	@route /api/v1/local/track [POST]
//	@returns bool
func (h *Handler) HandleLocalAddTrackedMedia(c echo.Context) error {
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
			err = h.App.LocalManager.TrackAnime(m.MediaId)
		case "manga":
			err = h.App.LocalManager.TrackManga(m.MediaId)
		}
	}

	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}

// HandleLocalRemoveTrackedMedia
//
//	@summary remove media from being tracked for offline sync.
//	@desc This will remove anime from being tracked for offline sync and delete any associated data.
//	@route /api/v1/local/track [DELETE]
//	@returns bool
func (h *Handler) HandleLocalRemoveTrackedMedia(c echo.Context) error {
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
		err = h.App.LocalManager.UntrackAnime(b.MediaId)
	case "manga":
		err = h.App.LocalManager.UntrackManga(b.MediaId)
	}

	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}

// HandleLocalGetIsMediaTracked
//
//	@summary checks if media is being tracked for offline sync.
//	@route /api/v1/local/track/{id}/{type} [GET]
//	@param id - int - true - "AniList anime media ID"
//	@param type - string - true - "Type of media (anime/manga)"
//	@returns bool
func (h *Handler) HandleLocalGetIsMediaTracked(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return h.RespondWithError(c, err)
	}

	kind := c.Param("type")
	tracked := h.App.LocalManager.IsMediaTracked(id, kind)

	return h.RespondWithData(c, tracked)
}

// HandleLocalSyncData
//
//	@summary syncs local data with AniList.
//	@route /api/v1/local/local [POST]
//	@returns bool
func (h *Handler) HandleLocalSyncData(c echo.Context) error {
	// Do not allow syncing if the user is simulated
	if h.App.GetUser().IsSimulated {
		return h.RespondWithData(c, true)
	}
	err := h.App.LocalManager.SynchronizeLocal()
	if err != nil {
		return h.RespondWithError(c, err)
	}

	if h.App.Settings.GetLibrary().AutoSaveCurrentMediaOffline {
		go func() {
			added, _ := h.App.LocalManager.AutoTrackCurrentMedia()
			if added {
				_ = h.App.LocalManager.SynchronizeLocal()
			}
		}()
	}

	return h.RespondWithData(c, true)
}

// HandleLocalGetSyncQueueState
//
//	@summary gets the current sync queue state.
//	@desc This will return the list of media that are currently queued for syncing.
//	@route /api/v1/local/queue [GET]
//	@returns local.QueueState
func (h *Handler) HandleLocalGetSyncQueueState(c echo.Context) error {
	return h.RespondWithData(c, h.App.LocalManager.GetSyncer().GetQueueState())
}

// HandleLocalSyncAnilistData
//
//	@summary syncs AniList data with local.
//	@route /api/v1/local/anilist [POST]
//	@returns bool
func (h *Handler) HandleLocalSyncAnilistData(c echo.Context) error {
	err := h.App.LocalManager.SynchronizeAnilist()
	if err != nil {
		return h.RespondWithError(c, err)
	}
	return h.RespondWithData(c, true)
}

// HandleLocalSetHasLocalChanges
//
//	@summary sets the flag to determine if there are local changes that need to be synced with AniList.
//	@route /api/v1/local/updated [POST]
//	@returns bool
func (h *Handler) HandleLocalSetHasLocalChanges(c echo.Context) error {
	type body struct {
		Updated bool `json:"updated"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	h.App.LocalManager.SetHasLocalChanges(b.Updated)
	return h.RespondWithData(c, true)
}

// HandleLocalGetHasLocalChanges
//
//	@summary gets the flag to determine if there are local changes that need to be synced with AniList.
//	@route /api/v1/local/updated [GET]
//	@returns bool
func (h *Handler) HandleLocalGetHasLocalChanges(c echo.Context) error {
	updated := h.App.LocalManager.HasLocalChanges()
	return h.RespondWithData(c, updated)
}

// HandleLocalGetLocalStorageSize
//
//	@summary gets the size of the local storage in a human-readable format.
//	@route /api/v1/local/storage/size [GET]
//	@returns string
func (h *Handler) HandleLocalGetLocalStorageSize(c echo.Context) error {
	size := h.App.LocalManager.GetLocalStorageSize()
	return h.RespondWithData(c, util.Bytes(uint64(size)))
}

// HandleLocalSyncSimulatedDataToAnilist
//
//	@summary syncs the simulated data to AniList.
//	@route /api/v1/local/sync-simulated-to-anilist [POST]
//	@returns bool
func (h *Handler) HandleLocalSyncSimulatedDataToAnilist(c echo.Context) error {
	err := h.App.LocalManager.SynchronizeSimulatedCollectionToAnilist()
	if err != nil {
		return h.RespondWithError(c, err)
	}
	return h.RespondWithData(c, true)
}
