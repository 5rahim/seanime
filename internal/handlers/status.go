package handlers

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"seanime/internal/constants"
	"seanime/internal/core"
	"seanime/internal/database/models"
	"seanime/internal/user"
	"seanime/internal/util"
	"seanime/internal/util/result"
	"slices"
	"strings"
	"time"

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
	User                  *user.User                    `json:"user"`
	Settings              *models.Settings              `json:"settings"`
	Version               string                        `json:"version"`
	VersionName           string                        `json:"versionName"`
	ThemeSettings         *models.Theme                 `json:"themeSettings"`
	IsOffline             bool                          `json:"isOffline"`
	MediastreamSettings   *models.MediastreamSettings   `json:"mediastreamSettings"`
	TorrentstreamSettings *models.TorrentstreamSettings `json:"torrentstreamSettings"`
	DebridSettings        *models.DebridSettings        `json:"debridSettings"`
	AnilistClientID       string                        `json:"anilistClientId"`
	Updating              bool                          `json:"updating"`         // If true, a new screen will be displayed
	IsDesktopSidecar      bool                          `json:"isDesktopSidecar"` // The server is running as a desktop sidecar
	FeatureFlags          core.FeatureFlags             `json:"featureFlags"`
	ServerReady           bool                          `json:"serverReady"`
	ServerHasPassword     bool                          `json:"serverHasPassword"`
}

var clientInfoCache = result.NewResultMap[string, util.ClientInfo]()

// NewStatus returns a new Status struct.
// It uses the RouteCtx to get the App instance containing the Database instance.
func (h *Handler) NewStatus(c echo.Context) *Status {
	var dbAcc *models.Account
	var currentUser *user.User
	var settings *models.Settings
	var theme *models.Theme
	//var mal *models.Mal

	// Get the user from the database (if logged in)
	if dbAcc, _ = h.App.Database.GetAccount(); dbAcc != nil {
		currentUser, _ = user.NewUser(dbAcc)
		if currentUser != nil {
			currentUser.Token = "HIDDEN"
		}
	} else {
		// If the user is not logged in, create a simulated user
		currentUser = user.NewSimulatedUser()
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

	status := &Status{
		OS:                    runtime.GOOS,
		ClientDevice:          clientInfo.Device,
		ClientPlatform:        clientInfo.Platform,
		DataDir:               h.App.Config.Data.AppDataDir,
		ClientUserAgent:       c.Request().UserAgent(),
		User:                  currentUser,
		Settings:              settings,
		Version:               h.App.Version,
		VersionName:           constants.VersionName,
		ThemeSettings:         theme,
		IsOffline:             h.App.Config.Server.Offline,
		MediastreamSettings:   h.App.SecondarySettings.Mediastream,
		TorrentstreamSettings: h.App.SecondarySettings.Torrentstream,
		DebridSettings:        h.App.SecondarySettings.Debrid,
		AnilistClientID:       h.App.Config.Anilist.ClientID,
		Updating:              false,
		IsDesktopSidecar:      h.App.IsDesktopSidecar,
		FeatureFlags:          h.App.FeatureFlags,
		ServerReady:           h.App.ServerReady,
		ServerHasPassword:     h.App.Config.Server.Password != "",
	}

	if c.Get("unauthenticated") != nil && c.Get("unauthenticated").(bool) {
		// If the user is unauthenticated, return a status with no user data
		status.OS = ""
		status.DataDir = ""
		status.User = user.NewSimulatedUser()
		status.ThemeSettings = nil
		status.MediastreamSettings = nil
		status.TorrentstreamSettings = nil
		status.Settings = &models.Settings{}
		status.DebridSettings = nil
		status.FeatureFlags = core.FeatureFlags{}
	}

	return status
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

	filename := c.Param("*")
	if filepath.Base(filename) != filename {
		h.App.Logger.Error().Msg("handlers: Invalid filename")
		return h.RespondWithError(c, fmt.Errorf("invalid filename"))
	}

	fp := filepath.Join(h.App.Config.Logs.Dir, filename)

	if filepath.Ext(fp) != ".log" {
		h.App.Logger.Error().Msg("handlers: Unsupported file extension")
		return h.RespondWithError(c, fmt.Errorf("unsupported file extension"))
	}

	if _, err := os.Stat(fp); err != nil {
		h.App.Logger.Error().Err(err).Msg("handlers: Stat error")
		return h.RespondWithError(c, err)
	}

	contentB, err := os.ReadFile(fp)
	if err != nil {
		h.App.Logger.Error().Err(err).Msg("handlers: Failed to read log file")
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, string(contentB))
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

// HandleGetLatestLogContent
//
//	@summary returns the content of the latest server log file.
//	@desc This returns the content of the most recent seanime- log file after flushing logs.
//	@route /api/v1/logs/latest [GET]
//	@returns string
func (h *Handler) HandleGetLatestLogContent(c echo.Context) error {
	if h.App.Config == nil || h.App.Config.Logs.Dir == "" {
		return h.RespondWithData(c, "")
	}

	// Flush logs first
	if h.App.OnFlushLogs != nil {
		h.App.OnFlushLogs()
		// Small delay to ensure logs are written
		time.Sleep(100 * time.Millisecond)
	}

	dirEntries, err := os.ReadDir(h.App.Config.Logs.Dir)
	if err != nil {
		h.App.Logger.Error().Err(err).Msg("handlers: Failed to read log directory")
		return h.RespondWithError(c, err)
	}

	var logFiles []string
	for _, entry := range dirEntries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) != ".log" || !strings.HasPrefix(strings.ToLower(name), "seanime-") {
			continue
		}
		logFiles = append(logFiles, filepath.Join(h.App.Config.Logs.Dir, name))
	}

	if len(logFiles) == 0 {
		h.App.Logger.Warn().Msg("handlers: No log files found")
		return h.RespondWithData(c, "")
	}

	// Sort files in descending order based on filename
	slices.SortFunc(logFiles, func(a, b string) int {
		return strings.Compare(filepath.Base(b), filepath.Base(a))
	})

	latestLogFile := logFiles[0]

	contentB, err := os.ReadFile(latestLogFile)
	if err != nil {
		h.App.Logger.Error().Err(err).Msg("handlers: Failed to read latest log file")
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, string(contentB))
}

// HandleGetAnnouncements
//
//	@summary returns the server announcements.
//	@desc This returns the announcements for the server.
//	@route /api/v1/announcements [POST]
//	@returns []updater.Announcement
func (h *Handler) HandleGetAnnouncements(c echo.Context) error {
	type body struct {
		Platform string `json:"platform"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	settings, _ := h.App.Database.GetSettings()

	announcements := h.App.Updater.GetAnnouncements(h.App.Version, b.Platform, settings)

	return h.RespondWithData(c, announcements)

}
