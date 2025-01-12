package handlersv2

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"seanime/internal/core"
	"seanime/internal/database/models"
	"seanime/internal/library/anime"
	"seanime/internal/util"
	"seanime/internal/util/result"
	"slices"
	"strings"

	"github.com/labstack/echo/v4"
)

// Status is a struct containing the user data, settings, and OS.
// It is used by the client in various places to access necessary information.
type Status struct {
	OS                    string                        `json:"os"`
	ClientDevice          string                        `json:"clientDevice"`
	ClientPlatform        string                        `json:"clientPlatform"`
	ClientUserAgent       string                        `json:"clientUserAgent"`
	DataDir               string                        `json:"dataDir"`
	User                  *anime.User                   `json:"user"`
	Settings              *models.Settings              `json:"settings"`
	Version               string                        `json:"version"`
	ThemeSettings         *models.Theme                 `json:"themeSettings"`
	IsOffline             bool                          `json:"isOffline"`
	MediastreamSettings   *models.MediastreamSettings   `json:"mediastreamSettings"`
	TorrentstreamSettings *models.TorrentstreamSettings `json:"torrentstreamSettings"`
	DebridSettings        *models.DebridSettings        `json:"debridSettings"`
	AnilistClientID       string                        `json:"anilistClientId"`
	Updating              bool                          `json:"updating"`         // If true, a new screen will be displayed
	IsDesktopSidecar      bool                          `json:"isDesktopSidecar"` // The server is running as a desktop sidecar
	FeatureFlags          core.FeatureFlags             `json:"featureFlags"`
}

var clientInfoCache = result.NewResultMap[string, util.ClientInfo]()

// NewStatus returns a new Status struct.
// It uses the RouteCtx to get the App instance containing the Database instance.
func (h *Handler) NewStatus(c echo.Context) *Status {
	var dbAcc *models.Account
	var user *anime.User
	var settings *models.Settings
	var theme *models.Theme
	//var mal *models.Mal

	if dbAcc, _ = h.App.Database.GetAccount(); dbAcc != nil {
		user, _ = anime.NewUser(dbAcc)
		if user != nil {
			user.Token = "HIDDEN"
		}
	}

	if settings, _ = h.App.Database.GetSettings(); settings != nil {
		if settings.ID == 0 || settings.Library == nil || settings.Torrent == nil || settings.MediaPlayer == nil {
			settings = nil
		}
	}

	clientInfo, found := clientInfoCache.Get(c.Request().UserAgent())
	if !found {
		clientInfo = util.GetClientInfo(c.Request().UserAgent())
		clientInfoCache.Set(c.Request().UserAgent(), clientInfo)
	}

	theme, _ = h.App.Database.GetTheme()

	return &Status{
		OS:                    runtime.GOOS,
		ClientDevice:          clientInfo.Device,
		ClientPlatform:        clientInfo.Platform,
		DataDir:               h.App.Config.Data.AppDataDir,
		ClientUserAgent:       c.Request().UserAgent(),
		User:                  user,
		Settings:              settings,
		Version:               h.App.Version,
		ThemeSettings:         theme,
		IsOffline:             h.App.Config.Server.Offline,
		MediastreamSettings:   h.App.SecondarySettings.Mediastream,
		TorrentstreamSettings: h.App.SecondarySettings.Torrentstream,
		DebridSettings:        h.App.SecondarySettings.Debrid,
		AnilistClientID:       h.App.Config.Anilist.ClientID,
		Updating:              false,
		IsDesktopSidecar:      h.App.IsDesktopSidecar,
		FeatureFlags:          h.App.FeatureFlags,
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
func (h *Handler) HandleGetStatus(c echo.Context) error {

	status := h.NewStatus(c)

	return h.RespondWithData(c, status)

}

func (h *Handler) HandleGetLogContent(c echo.Context) error {
	if h.App.Config == nil || h.App.Config.Logs.Dir == "" {
		return h.RespondWithData(c, "")
	}

	filename := c.Param("*1")
	fp := util.NormalizePath(filepath.Join(h.App.Config.Logs.Dir, filename))

	fileContent := ""
	filepath.WalkDir(h.App.Config.Logs.Dir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".log" {
			return nil
		}

		if util.NormalizePath(path) == fp {
			contentB, err := os.ReadFile(fp)
			if err != nil {
				return err
			}
			fileContent = string(contentB)
		}
		return nil
	})

	return h.RespondWithData(c, fileContent)
}

var newestLogFilename = ""

// HandleGetLogFilenames
//
//	@summary returns the log filenames.
//	@desc This returns the filenames of all log files in the logs directory.
//	@route /api/v1/logs/filenames [GET]
//	@returns []string
func (h *Handler) HandleGetLogFilenames(c echo.Context) error {
	if h.App.Config == nil || h.App.Config.Logs.Dir == "" {
		return h.RespondWithData(c, []string{})
	}

	var filenames []string
	filepath.WalkDir(h.App.Config.Logs.Dir, func(path string, d fs.DirEntry, err error) error {
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

	return h.RespondWithData(c, filenames)
}

// HandleDeleteLogs
//
//	@summary deletes certain log files.
//	@desc This deletes the log files with the given filenames.
//	@route /api/v1/logs [DELETE]
//	@returns bool
func (h *Handler) HandleDeleteLogs(c echo.Context) error {
	type body struct {
		Filenames []string `json:"filenames"`
	}

	if h.App.Config == nil || h.App.Config.Logs.Dir == "" {
		return h.RespondWithData(c, false)
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	filepath.WalkDir(h.App.Config.Logs.Dir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".log" {
			return nil
		}

		for _, filename := range b.Filenames {
			if util.NormalizePath(filepath.Base(path)) == util.NormalizePath(filename) {
				if util.NormalizePath(newestLogFilename) == util.NormalizePath(filename) {
					return fmt.Errorf("cannot delete the newest log file")
				}
				if err := os.Remove(path); err != nil {
					return err
				}
			}
		}
		return nil
	})

	return h.RespondWithData(c, true)
}
