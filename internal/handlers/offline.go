package handlers

import (
	"errors"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/offline"
)

var creatingOfflineSnapshot = false

// HandleCreateOfflineSnapshot
//
//	@summary creates an offline snapshot.
//	@desc This will create an offline snapshot of the given anime media ids and downloaded manga chapters.
//	@desc It sends a websocket event when the snapshot is created, telling the client to refetch the offline snapshot.
//	@desc This is a non-blocking operation.
//	@route /api/offline/snapshot [POST]
//	@returns bool
func HandleCreateOfflineSnapshot(c *RouteCtx) error {

	type body struct {
		AnimeMediaIds []int `json:"animeMediaIds"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	if creatingOfflineSnapshot {
		return c.RespondWithError(errors.New("snapshot creation is already in progress"))
	}

	go func() {
		creatingOfflineSnapshot = true
		defer func() { creatingOfflineSnapshot = false }()

		err := c.App.OfflineHub.CreateSnapshot(&offline.NewSnapshotOptions{
			AnimeToDownload:  b.AnimeMediaIds,
			DownloadAssetsOf: b.AnimeMediaIds,
		})

		if err != nil {
			c.App.WSEventManager.SendEvent(events.ErrorToast, err.Error())
		} else {
			c.App.WSEventManager.SendEvent(events.SuccessToast, "Offline snapshot created successfully")
			c.App.WSEventManager.SendEvent(events.OfflineSnapshotCreated, true)
		}
	}()

	return c.RespondWithData(true)
}

// HandleGetOfflineSnapshot
//
//	@summary retrieves the offline snapshot.
//	@desc This will return the latest offline snapshot. (Offline only)
//	@route /api/offline/snapshot [GET]
//	@returns offline.Snapshot
func HandleGetOfflineSnapshot(c *RouteCtx) error {
	snapshot, _ := c.App.OfflineHub.GetLatestSnapshot(false)
	return c.RespondWithData(snapshot)
}

// HandleGetOfflineSnapshotEntry
//
//	@summary retrieves an offline snapshot entry.
//	@desc This will return the latest offline snapshot entry so the client can display the data.
//	@route /api/offline/snapshot-entry [GET]
//	@returns offline.SnapshotEntry
func HandleGetOfflineSnapshotEntry(c *RouteCtx) error {
	entry, _ := c.App.OfflineHub.GetLatestSnapshotEntry()
	if entry != nil {
		entry.Collections = nil
	}
	return c.RespondWithData(entry)
}

// HandleUpdateOfflineEntryListData
//
//	@summary updates data for an offline entry list.
//	@desc This will update the offline entry list data. (Offline only)
//	@route /api/offline/snapshot-entry [PATCH]
//	@returns bool
func HandleUpdateOfflineEntryListData(c *RouteCtx) error {

	type body struct {
		MediaId   *int                     `json:"mediaId"`
		Status    *anilist.MediaListStatus `json:"status"`
		Score     *int                     `json:"score"`
		Progress  *int                     `json:"progress"`
		StartDate *string                  `json:"startDate"`
		EndDate   *string                  `json:"endDate"`
		Type      string                   `json:"type"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	err := c.App.OfflineHub.UpdateEntryListData(
		b.MediaId,
		b.Status,
		b.Score,
		b.Progress,
		b.StartDate,
		b.EndDate,
		b.Type,
	)
	if err != nil {
		return c.RespondWithError(err)
	}
	return c.RespondWithData(true)
}

// HandleSyncOfflineData
//
//	@summary synchronizes offline data with AniList when the user is back online.
//	@route /api/offline/sync [POST]
//	@returns bool
func HandleSyncOfflineData(c *RouteCtx) error {
	err := c.App.OfflineHub.SyncListData()
	if err != nil {
		return c.RespondWithError(err)
	}
	return c.RespondWithData(true)
}
