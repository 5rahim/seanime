package handlers

import (
	"github.com/seanime-app/seanime/internal/entities"
	"github.com/seanime-app/seanime/internal/models"
	"runtime"
)

// Status is a struct containing the user data, settings, and OS.
// It is used by the client in various places to access necessary information.
type Status struct {
	OS       string           `json:"os"`
	User     *entities.User   `json:"user"`
	Settings *models.Settings `json:"settings"`
	Mal      *models.Mal      `json:"mal"`
	Version  string           `json:"version"`
}

// NewStatus returns a new Status struct.
// It uses the RouteCtx to get the App instance containing the Database instance.
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

// HandleStatus is a route handler that returns the app status.
//
//	GET /v1/status
//
// It is called on every page load to get the most up-to-date data.
// It is also called right after updating the settings.
func HandleStatus(c *RouteCtx) error {

	status := NewStatus(c)

	return c.RespondWithData(status)

}
