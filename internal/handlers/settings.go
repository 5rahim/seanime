package handlers

import (
	"github.com/seanime-app/seanime-server/internal/models"
	"time"
)

type SettingsBody struct {
	Library     models.LibrarySettings     `json:"library"`
	MediaPlayer models.MediaPlayerSettings `json:"mediaPlayer"`
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
	})

	if err != nil {
		return c.RespondWithError(err)
	}

	// Refresh the settings dependents
	c.App.InitSettingsDependents()

	return c.RespondWithData(settings)
}
