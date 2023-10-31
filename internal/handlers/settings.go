package handlers

import (
	"github.com/seanime-app/seanime-server/internal/models"
	"time"
)

type SettingsBody struct {
	Library     models.LibrarySettings     `json:"library"`
	MediaPlayer models.MediaPlayerSettings `json:"mediaPlayer"`
	Torrent     models.TorrentSettings     `json:"torrent"`
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
	})

	if err != nil {
		return c.RespondWithError(err)
	}

	c.App.WSEventManager.SendEvent("settings", settings)

	// Refresh the settings dependents
	c.App.InitOrRefreshDependencies()

	return c.RespondWithData(settings)
}
