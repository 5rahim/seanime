package handlers

import (
	"github.com/seanime-app/seanime/internal/entities"
	"github.com/seanime-app/seanime/internal/models"
	"runtime"
)

// Status is a struct containing the user data, settings, and OS.
// It is used by the client to authenticate the user and get settings.
type Status struct {
	OS       string           `json:"os"`
	User     *entities.User   `json:"user"`
	Settings *models.Settings `json:"settings"`
	Mal      *models.Mal      `json:"mal"`
	Version  string           `json:"version"`
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

	mal, err := c.App.Database.GetMalInfo()
	if err != nil {
		mal = nil
	}
	return &Status{
		OS:       runtime.GOOS,
		User:     user,
		Settings: settings,
		Mal:      mal,
		Version:  c.App.Version,
	}
}

func HandleStatus(c *RouteCtx) error {

	status := NewStatus(c)

	return c.RespondWithData(status)

}
