package handlers

import (
	"errors"
	"github.com/seanime-app/seanime/internal/models"
	"runtime"
	"time"
)

type SettingsBody struct {
	Library     models.LibrarySettings     `json:"library"`
	MediaPlayer models.MediaPlayerSettings `json:"mediaPlayer"`
	Torrent     models.TorrentSettings     `json:"torrent"`
	Anilist     models.AnilistSettings     `json:"anilist"`
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

	body := new(SettingsBody)

	if err := c.Fiber.BodyParser(body); err != nil {
		return c.RespondWithError(err)
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
	})

	if err != nil {
		return c.RespondWithError(err)
	}

	c.App.WSEventManager.SendEvent("settings", settings)

	status := NewStatus(c)

	// Refresh the settings dependents
	c.App.InitOrRefreshModules()

	return c.RespondWithData(status)
}
