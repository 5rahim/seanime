package handlers

import (
	"github.com/seanime-app/seanime/internal/database/models"
	"github.com/seanime-app/seanime/internal/library/entities"
	"runtime"
)

// Status is a struct containing the user data, settings, and OS.
// It is used by the client in various places to access necessary information.
type Status struct {
	OS            string           `json:"os"`
	User          *entities.User   `json:"user"`
	Settings      *models.Settings `json:"settings"`
	Mal           *models.Mal      `json:"mal"`
	Version       string           `json:"version"`
	ThemeSettings *models.Theme    `json:"themeSettings"`
}

// NewStatus returns a new Status struct.
// It uses the RouteCtx to get the App instance containing the Database instance.
func NewStatus(c *RouteCtx) *Status {
	var dbAcc *models.Account
	var user *entities.User
	var settings *models.Settings
	var theme *models.Theme
	var mal *models.Mal

	if dbAcc, _ = c.App.Database.GetAccount(); dbAcc != nil {
		user, _ = entities.NewUser(dbAcc)
	}

	if settings, _ = c.App.Database.GetSettings(); settings != nil {
		if settings.ID == 0 || settings.Library == nil || settings.Torrent == nil || settings.MediaPlayer == nil {
			settings = nil
		}
	}

	theme, _ = c.App.Database.GetTheme()
	mal, _ = c.App.Database.GetMalInfo()
	return &Status{
		OS:            runtime.GOOS,
		User:          user,
		Settings:      settings,
		Mal:           mal,
		Version:       c.App.Version,
		ThemeSettings: theme,
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
