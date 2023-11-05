package handlers

import (
	"github.com/seanime-app/seanime-server/internal/entities"
	"github.com/seanime-app/seanime-server/internal/models"
	"runtime"
)

type Status struct {
	OS       string           `json:"os"`
	User     *entities.User   `json:"user"`
	Settings *models.Settings `json:"settings"`
}

func NewStatus(c *RouteCtx) *Status {
	dbAcc, err := c.App.Database.GetAccount()
	if err != nil {
		dbAcc = nil
	}
	user, err := entities.NewUser(dbAcc)
	if err != nil {
		user = nil
	}

	settings, err := c.App.Database.GetSettings()
	if err != nil {
		settings = nil
	}
	if settings.ID == 0 || settings.Library == nil || settings.Torrent == nil || settings.MediaPlayer == nil {
		settings = nil
	}
	return &Status{
		OS:       runtime.GOOS,
		User:     user,
		Settings: settings,
	}
}

func HandleStatus(c *RouteCtx) error {

	status := NewStatus(c)

	return c.RespondWithData(status)

}
