package handlers

import (
	"github.com/seanime-app/seanime-server/internal/models"
	"time"
)

type SettingsBody struct {
	Library struct {
		LibraryPath string `json:"libraryPath"`
	} `json:"library"`
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
		Library: &models.LibrarySettings{
			LibraryPath: body.Library.LibraryPath,
		},
	})

	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(settings)
}
