package handlers

import (
	"errors"
	"github.com/seanime-app/seanime/internal/models"
	"runtime"
	"time"
)

type settingsBody struct {
	Library     models.LibrarySettings     `json:"library"`
	MediaPlayer models.MediaPlayerSettings `json:"mediaPlayer"`
	Torrent     models.TorrentSettings     `json:"torrent"`
	Anilist     models.AnilistSettings     `json:"anilist"`
}

type listSyncSettingsBody struct {
	Automatic bool   `json:"automatic"`
	Origin    string `json:"origin"`
}

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

func HandleSaveSettings(c *RouteCtx) error {

	body := new(settingsBody)

	if err := c.Fiber.BodyParser(body); err != nil {
		return c.RespondWithError(err)
	}

	listSyncSettings := &models.ListSyncSettings{}
	prevSettings, err := c.App.Database.GetSettings()
	if err == nil && prevSettings.ListSync != nil {
		listSyncSettings = prevSettings.ListSync
	}

	settings, err := c.App.Database.UpsertSettings(&models.Settings{
		BaseModel: models.BaseModel{
			ID:        1,
			UpdatedAt: time.Now(),
		},
		Library:     &body.Library,
		MediaPlayer: &body.MediaPlayer,
		Torrent:     &body.Torrent,
		Anilist:     &body.Anilist,
		ListSync:    listSyncSettings,
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

func HandleSaveListSyncSettings(c *RouteCtx) error {

	body := new(listSyncSettingsBody)

	if err := c.Fiber.BodyParser(body); err != nil {
		return c.RespondWithError(err)
	}

	prevSettings, err := c.App.Database.GetSettings()
	if err != nil {
		return c.RespondWithError(err)
	}

	settings, err := c.App.Database.UpsertSettings(&models.Settings{
		BaseModel: models.BaseModel{
			ID:        1,
			UpdatedAt: time.Now(),
		},
		Library:     prevSettings.Library,
		MediaPlayer: prevSettings.MediaPlayer,
		Torrent:     prevSettings.Torrent,
		Anilist:     prevSettings.Anilist,
		ListSync: &models.ListSyncSettings{
			Automatic: body.Automatic,
			Origin:    body.Origin,
		},
	})

	if err != nil {
		return c.RespondWithError(err)
	}

	c.App.WSEventManager.SendEvent("settings", settings)

	status := NewStatus(c)

	return c.RespondWithData(status)
}
