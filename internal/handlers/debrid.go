package handlers

import (
	"errors"
	hibiketorrent "github.com/5rahim/hibike/pkg/extension/torrent"
	"path/filepath"
	"seanime/internal/api/anilist"
	"seanime/internal/database/models"
	"seanime/internal/debrid/debrid"
	"seanime/internal/events"
)

// HandleGetDebridSettings
//
//	@summary get debrid settings.
//	@desc This returns the debrid settings.
//	@returns models.DebridSettings
//	@route /api/v1/debrid/settings [GET]
func HandleGetDebridSettings(c *RouteCtx) error {
	debridSettings, found := c.App.Database.GetDebridSettings()
	if !found {
		return c.RespondWithError(errors.New("debrid settings not found"))
	}

	return c.RespondWithData(debridSettings)
}

// HandleSaveDebridSettings
//
//	@summary save debrid settings.
//	@desc This saves the debrid settings.
//	@desc The client should refetch the server status.
//	@returns models.DebridSettings
//	@route /api/v1/debrid/settings [PATCH]
func HandleSaveDebridSettings(c *RouteCtx) error {

	type body struct {
		Settings models.DebridSettings `json:"settings"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	settings, err := c.App.Database.UpsertDebridSettings(&b.Settings)
	if err != nil {
		return c.RespondWithError(err)
	}

	c.App.InitOrRefreshDebridSettings()

	return c.RespondWithData(settings)
}

// HandleDebridAddTorrents
//
//	@summary add torrent to debrid.
//	@desc This adds a torrent to the debrid service.
//	@returns bool
//	@route /api/v1/debrid/torrents [POST]
func HandleDebridAddTorrents(c *RouteCtx) error {

	type body struct {
		Torrents    []hibiketorrent.AnimeTorrent `json:"torrents"`
		Media       *anilist.BaseAnime           `json:"media"`
		Destination string                       `json:"destination"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	if !c.App.DebridClientRepository.HasProvider() {
		return c.RespondWithError(errors.New("debrid provider not set"))
	}

	for _, torrent := range b.Torrents {
		// Add the torrent to the debrid service
		_, err := c.App.DebridClientRepository.AddAndQueueTorrent(debrid.AddTorrentOptions{
			MagnetLink: torrent.MagnetLink,
		}, b.Destination, b.Media.ID)
		if err != nil {
			// If there is only one torrent, return the error
			if len(b.Torrents) == 1 {
				return c.RespondWithError(err)
			} else {
				// If there are multiple torrents, send an error toast and continue to the next torrent
				c.App.Logger.Err(err).Msg("Failed to add torrent to debrid")
				c.App.WSEventManager.SendEvent(events.ErrorToast, err.Error())
				continue
			}
		}
	}

	return c.RespondWithData(true)
}

// HandleDebridDownloadTorrent
//
//	@summary download torrent from debrid.
//	@desc Manually downloads a torrent from the debrid service locally.
//	@returns bool
//	@route /api/v1/debrid/torrents/download [POST]
func HandleDebridDownloadTorrent(c *RouteCtx) error {

	type body struct {
		TorrentItem debrid.TorrentItem `json:"torrentItem"`
		Destination string             `json:"destination"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	if !filepath.IsAbs(b.Destination) {
		return c.RespondWithError(errors.New("destination must be an absolute path"))
	}

	// Remove the torrent from the database
	// This is done so that the torrent is not downloaded automatically
	// We ignore the error here because the torrent might not be in the database
	_ = c.App.Database.DeleteDebridTorrentItemByTorrentItemId(b.TorrentItem.ID)

	// Download the torrent locally
	err := c.App.DebridClientRepository.DownloadTorrent(b.TorrentItem, b.Destination)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}

// HandleDebridCancelDownload
//
//	@summary cancel download from debrid.
//	@desc This cancels a download from the debrid service.
//	@returns bool
//	@route /api/v1/debrid/torrents/cancel [POST]
func HandleDebridCancelDownload(c *RouteCtx) error {

	type body struct {
		ItemID string `json:"itemID"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	err := c.App.DebridClientRepository.CancelDownload(b.ItemID)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}

// HandleDebridDeleteTorrent
//
//	@summary remove torrent from debrid.
//	@desc This removes a torrent from the debrid service.
//	@returns bool
//	@route /api/v1/debrid/torrent [DELETE]
func HandleDebridDeleteTorrent(c *RouteCtx) error {

	type body struct {
		TorrentItem debrid.TorrentItem `json:"torrentItem"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	provider, err := c.App.DebridClientRepository.GetProvider()
	if err != nil {
		return c.RespondWithError(err)
	}

	err = provider.DeleteTorrent(b.TorrentItem.ID)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}

// HandleDebridGetTorrents
//
//	@summary get torrents from debrid.
//	@desc This gets the torrents from the debrid service.
//	@returns []debrid.TorrentItem
//	@route /api/v1/debrid/torrents [GET]
func HandleDebridGetTorrents(c *RouteCtx) error {

	provider, err := c.App.DebridClientRepository.GetProvider()
	if err != nil {
		return c.RespondWithError(err)
	}

	torrents, err := provider.GetTorrents()
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(torrents)
}
