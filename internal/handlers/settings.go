package handlers

import (
	"errors"
	"github.com/seanime-app/seanime/internal/database/models"
	"runtime"
	"time"
)

type settingsBody struct {
	Library     models.LibrarySettings     `json:"library"`
	MediaPlayer models.MediaPlayerSettings `json:"mediaPlayer"`
	Torrent     models.TorrentSettings     `json:"torrent"`
	Anilist     models.AnilistSettings     `json:"anilist"`
	Discord     models.DiscordSettings     `json:"discord"`
}

// HandleGetSettings returns the app settings.
//
//	GET /v1/settings
func HandleGetSettings(c *RouteCtx) error {

	settings, err := c.App.Database.GetSettings()
	if err != nil {
		return c.RespondWithError(err)
	}
	if settings.ID == 0 {
		return c.RespondWithError(errors.New(runtime.GOOS))
	}

	return c.RespondWithData(settings)
}

// HandleSaveSettings updates the app settings.
// It returns a new Status containing the updated settings.
//
//	POST /v1/settings
func HandleSaveSettings(c *RouteCtx) error {

	body := new(settingsBody)

	if err := c.Fiber.BodyParser(body); err != nil {
		return c.RespondWithError(err)
	}

	listSyncSettings := &models.ListSyncSettings{}
	autoDownloaderSettings := &models.AutoDownloaderSettings{}
	prevSettings, err := c.App.Database.GetSettings()
	if err == nil && prevSettings.ListSync != nil {
		listSyncSettings = prevSettings.ListSync
	}
	if err == nil && prevSettings.AutoDownloader != nil {
		autoDownloaderSettings = prevSettings.AutoDownloader
	}

	settings, err := c.App.Database.UpsertSettings(&models.Settings{
		BaseModel: models.BaseModel{
			ID:        1,
			UpdatedAt: time.Now(),
		},
		Library:        &body.Library,
		MediaPlayer:    &body.MediaPlayer,
		Torrent:        &body.Torrent,
		Anilist:        &body.Anilist,
		Discord:        &body.Discord,
		ListSync:       listSyncSettings,
		AutoDownloader: autoDownloaderSettings,
	})

	if err != nil {
		return c.RespondWithError(err)
	}

	c.App.WSEventManager.SendEvent("settings", settings)

	status := NewStatus(c)

	// Refresh modules that depend on the settings
	c.App.InitOrRefreshModules()

	return c.RespondWithData(status)
}

// HandleSaveListSyncSettings
// This will also delete the cached listsync.ListSync instance.
// It returns true if the settings were saved successfully.
//
//	PATCH /v1/settings/list-sync
func HandleSaveListSyncSettings(c *RouteCtx) error {

	body := new(models.ListSyncSettings)

	if err := c.Fiber.BodyParser(body); err != nil {
		return c.RespondWithError(err)
	}

	prevSettings, err := c.App.Database.GetSettings()
	if err != nil {
		return c.RespondWithError(err)
	}

	_, err = c.App.Database.UpsertSettings(&models.Settings{
		BaseModel: models.BaseModel{
			ID:        1,
			UpdatedAt: time.Now(),
		},
		Library:        prevSettings.Library,
		MediaPlayer:    prevSettings.MediaPlayer,
		Torrent:        prevSettings.Torrent,
		Anilist:        prevSettings.Anilist,
		AutoDownloader: prevSettings.AutoDownloader,
		Discord:        prevSettings.Discord,
		ListSync: &models.ListSyncSettings{
			Automatic: body.Automatic,
			Origin:    body.Origin,
		},
	})

	if err != nil {
		return c.RespondWithError(err)
	}

	c.App.ListSyncCache.Delete(0)

	// DEVNOTE: Refetch server status from client

	return c.RespondWithData(true)
}

// HandleSaveAutoDownloaderSettings
// It returns true if the settings were saved successfully.
//
//	PATCH /v1/settings/auto-downloader
func HandleSaveAutoDownloaderSettings(c *RouteCtx) error {

	body := new(models.AutoDownloaderSettings)

	if err := c.Fiber.BodyParser(body); err != nil {
		return c.RespondWithError(err)
	}

	prevSettings, err := c.App.Database.GetSettings()
	if err != nil {
		return c.RespondWithError(err)
	}

	// Validation
	if body.Interval < 2 {
		return c.RespondWithError(errors.New("interval must be at least 2 minutes"))
	}

	autoDownloaderSettings := &models.AutoDownloaderSettings{
		Provider:              prevSettings.Library.TorrentProvider,
		Interval:              body.Interval,
		Enabled:               body.Enabled,
		DownloadAutomatically: body.DownloadAutomatically,
	}

	_, err = c.App.Database.UpsertSettings(&models.Settings{
		BaseModel: models.BaseModel{
			ID:        1,
			UpdatedAt: time.Now(),
		},
		Library:        prevSettings.Library,
		MediaPlayer:    prevSettings.MediaPlayer,
		Torrent:        prevSettings.Torrent,
		Anilist:        prevSettings.Anilist,
		ListSync:       prevSettings.ListSync,
		Discord:        prevSettings.Discord,
		AutoDownloader: autoDownloaderSettings,
	})
	if err != nil {
		return c.RespondWithError(err)
	}

	// Update Auto Downloader - This runs in a goroutine
	c.App.AutoDownloader.SetSettings(autoDownloaderSettings, prevSettings.Library.TorrentProvider)

	return c.RespondWithData(true)
}
