package handlers

import (
	"errors"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/offline"
)

var creatingOfflineSnapshot = false

// HandleCreateOfflineSnapshot
//
//	POST /api/offline/snapshot
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
		}

		c.App.WSEventManager.SendEvent(events.InfoToast, "Offline snapshot created successfully")
		c.App.WSEventManager.SendEvent(events.OfflineSnapshotCreated, true)
	}()

	return c.RespondWithData(true)
}

// HandleGetOfflineSnapshot
//
//	GET /api/offline/snapshot
func HandleGetOfflineSnapshot(c *RouteCtx) error {
	snapshot, _ := c.App.OfflineHub.GetLatestSnapshot(false)
	return c.RespondWithData(snapshot)
}

// HandleGetOfflineSnapshotEntry
//
//	GET /api/offline/snapshot-entry
func HandleGetOfflineSnapshotEntry(c *RouteCtx) error {
	entry, _ := c.App.OfflineHub.GetLatestSnapshotEntry()
	if entry != nil {
		entry.Collections = nil
	}
	return c.RespondWithData(entry)
}
