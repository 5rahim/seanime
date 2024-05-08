package handlers

import (
	"github.com/seanime-app/seanime/internal/core"
	"github.com/seanime-app/seanime/internal/database/models"
	"github.com/seanime-app/seanime/internal/library/anime"
	"runtime"
)

// Status is a struct containing the user data, settings, and OS.
// It is used by the client in various places to access necessary information.
type Status struct {
	OS            string            `json:"os"`
	User          *anime.User       `json:"user"`
	Settings      *models.Settings  `json:"settings"`
	Mal           *models.Mal       `json:"mal"`
	Version       string            `json:"version"`
	ThemeSettings *models.Theme     `json:"themeSettings"`
	IsOffline     bool              `json:"isOffline"`
	FeatureFlags  core.FeatureFlags `json:"featureFlags"`
}

// NewStatus returns a new Status struct.
// It uses the RouteCtx to get the App instance containing the Database instance.
func NewStatus(c *RouteCtx) *Status {
	var dbAcc *models.Account
	var user *anime.User
	var settings *models.Settings
	var theme *models.Theme
	var mal *models.Mal

	if dbAcc, _ = c.App.Database.GetAccount(); dbAcc != nil {
		user, _ = anime.NewUser(dbAcc)
		if user != nil {
			user.Token = "HIDDEN"
		}
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
		IsOffline:     c.App.Config.Server.Offline,
		FeatureFlags:  c.App.FeatureFlags,
	}
}

// HandleGetStatus
//
//	@summary returns the server status.
//	@desc The server status includes app info, auth info and settings.
//	@desc The client uses this to set the UI.
//	@desc It is called on every page load to get the most up-to-date data.
//	@desc It should be called right after updating the settings.
//	@route /api/v1/status [GET]
//	@returns handlers.Status
func HandleGetStatus(c *RouteCtx) error {

	status := NewStatus(c)

	return c.RespondWithData(status)

}
