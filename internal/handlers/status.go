package handlers

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"seanime/internal/database/models"
	"seanime/internal/library/anime"
	"seanime/internal/util"
	"seanime/internal/util/result"
	"slices"
	"strings"
)

// Status is a struct containing the user data, settings, and OS.
// It is used by the client in various places to access necessary information.
type Status struct {
	OS              string           `json:"os"`
	ClientDevice    string           `json:"clientDevice"`
	ClientPlatform  string           `json:"clientPlatform"`
	ClientUserAgent string           `json:"clientUserAgent"`
	User            *anime.User      `json:"user"`
	Settings        *models.Settings `json:"settings"`
	Mal             *models.Mal      `json:"mal"`
	Version         string           `json:"version"`
	ThemeSettings   *models.Theme    `json:"themeSettings"`
	IsOffline       bool             `json:"isOffline"`
	//FeatureFlags          core.FeatureFlags             `json:"featureFlags"`
	MediastreamSettings   *models.MediastreamSettings   `json:"mediastreamSettings"`
	TorrentstreamSettings *models.TorrentstreamSettings `json:"torrentstreamSettings"`
	DebridSettings        *models.DebridSettings        `json:"debridSettings"`
	AnilistClientID       string                        `json:"anilistClientId"`
	Updating              bool                          `json:"updating"`         // If true, a new screen will be displayed
	IsDesktopSidecar      bool                          `json:"isDesktopSidecar"` // The server is running as a desktop sidecar
}

var clientInfoCache = result.NewResultMap[string, util.ClientInfo]()

// NewStatus returns a new Status struct.
// It uses the RouteCtx to get the App instance containing the Database instance.
func NewStatus(c *RouteCtx) *Status {
	var dbAcc *models.Account
	var user *anime.User
	var settings *models.Settings
	var theme *models.Theme
	//var mal *models.Mal

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

	clientInfo, found := clientInfoCache.Get(c.Fiber.Get("User-Agent"))
	if !found {
		clientInfo = util.GetClientInfo(c.Fiber.Get("User-Agent"))
		clientInfoCache.Set(c.Fiber.Get("User-Agent"), clientInfo)
	}

	theme, _ = c.App.Database.GetTheme()

	return &Status{
		OS:                    runtime.GOOS,
		ClientDevice:          clientInfo.Device,
		ClientPlatform:        clientInfo.Platform,
		ClientUserAgent:       c.Fiber.Get("User-Agent"),
		User:                  user,
		Settings:              settings,
		Version:               c.App.Version,
		ThemeSettings:         theme,
		IsOffline:             c.App.Config.Server.Offline,
		MediastreamSettings:   c.App.SecondarySettings.Mediastream,
		TorrentstreamSettings: c.App.SecondarySettings.Torrentstream,
		DebridSettings:        c.App.SecondarySettings.Debrid,
		AnilistClientID:       c.App.Config.Anilist.ClientID,
		Updating:              false,
		IsDesktopSidecar:      c.App.IsDesktopSidecar,
		//FeatureFlags:          c.App.FeatureFlags,
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

func HandleGetLogContent(c *RouteCtx) error {
	if c.App.Config == nil || c.App.Config.Logs.Dir == "" {
		return c.RespondWithData("")
	}

	filename := c.Fiber.AllParams()["*1"]
	fp := strings.ToLower(filepath.ToSlash(filepath.Join(c.App.Config.Logs.Dir, filename)))

	fileContent := ""
	filepath.WalkDir(c.App.Config.Logs.Dir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".log" {
			return nil
		}

		if strings.ToLower(filepath.ToSlash(path)) == fp {
			contentB, err := os.ReadFile(fp)
			if err != nil {
				return err
			}
			fileContent = string(contentB)
		}
		return nil
	})

	return c.RespondWithData(fileContent)
}

var newestLogFilename = ""

// HandleGetLogFilenames
//
//	@summary returns the log filenames.
//	@desc This returns the filenames of all log files in the logs directory.
//	@route /api/v1/logs/filenames [GET]
//	@returns []string
func HandleGetLogFilenames(c *RouteCtx) error {
	if c.App.Config == nil || c.App.Config.Logs.Dir == "" {
		return c.RespondWithData([]string{})
	}

	var filenames []string
	filepath.WalkDir(c.App.Config.Logs.Dir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".log" {
			return nil
		}

		filenames = append(filenames, filepath.Base(path))
		return nil
	})

	// Sort from newest to oldest & store the newest log filename
	if len(filenames) > 0 {
		slices.SortStableFunc(filenames, func(i, j string) int {
			return strings.Compare(j, i)
		})
		for _, filename := range filenames {
			if strings.HasPrefix(strings.ToLower(filename), "seanime-") {
				newestLogFilename = filename
				break
			}
		}
	}

	return c.RespondWithData(filenames)
}

// HandleDeleteLogs
//
//	@summary deletes certain log files.
//	@desc This deletes the log files with the given filenames.
//	@route /api/v1/logs [DELETE]
//	@returns bool
func HandleDeleteLogs(c *RouteCtx) error {
	type body struct {
		Filenames []string `json:"filenames"`
	}

	if c.App.Config == nil || c.App.Config.Logs.Dir == "" {
		return c.RespondWithData(false)
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	filepath.WalkDir(c.App.Config.Logs.Dir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".log" {
			return nil
		}

		for _, filename := range b.Filenames {
			if strings.ToLower(filepath.Base(path)) == strings.ToLower(filename) {
				if strings.ToLower(newestLogFilename) == strings.ToLower(filename) {
					return fmt.Errorf("cannot delete the newest log file")
				}
				if err := os.Remove(path); err != nil {
					return err
				}
			}
		}
		return nil
	})

	return c.RespondWithData(true)
}
