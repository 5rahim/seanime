package handlers

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"seanime/internal/constants"
	"seanime/internal/core"
	"seanime/internal/database/models"
	"seanime/internal/user"
	"seanime/internal/util"
	"seanime/internal/util/result"
	"slices"
	"strconv"
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

type MemoryStatsResponse struct {
	Alloc         uint64  `json:"alloc"`         // bytes allocated and not yet freed
	TotalAlloc    uint64  `json:"totalAlloc"`    // bytes allocated (even if freed)
	Sys           uint64  `json:"sys"`           // bytes obtained from system
	Lookups       uint64  `json:"lookups"`       // number of pointer lookups
	Mallocs       uint64  `json:"mallocs"`       // number of mallocs
	Frees         uint64  `json:"frees"`         // number of frees
	HeapAlloc     uint64  `json:"heapAlloc"`     // bytes allocated and not yet freed
	HeapSys       uint64  `json:"heapSys"`       // bytes obtained from system
	HeapIdle      uint64  `json:"heapIdle"`      // bytes in idle spans
	HeapInuse     uint64  `json:"heapInuse"`     // bytes in non-idle span
	HeapReleased  uint64  `json:"heapReleased"`  // bytes released to OS
	HeapObjects   uint64  `json:"heapObjects"`   // total number of allocated objects
	StackInuse    uint64  `json:"stackInuse"`    // bytes used by stack allocator
	StackSys      uint64  `json:"stackSys"`      // bytes obtained from system for stack allocator
	MSpanInuse    uint64  `json:"mSpanInuse"`    // bytes used by mspan structures
	MSpanSys      uint64  `json:"mSpanSys"`      // bytes obtained from system for mspan structures
	MCacheInuse   uint64  `json:"mCacheInuse"`   // bytes used by mcache structures
	MCacheSys     uint64  `json:"mCacheSys"`     // bytes obtained from system for mcache structures
	BuckHashSys   uint64  `json:"buckHashSys"`   // bytes used by the profiling bucket hash table
	GCSys         uint64  `json:"gcSys"`         // bytes used for garbage collection system metadata
	OtherSys      uint64  `json:"otherSys"`      // bytes used for other system allocations
	NextGC        uint64  `json:"nextGC"`        // next collection will happen when HeapAlloc ≥ this amount
	LastGC        uint64  `json:"lastGC"`        // time the last garbage collection finished
	PauseTotalNs  uint64  `json:"pauseTotalNs"`  // cumulative nanoseconds in GC stop-the-world pauses
	PauseNs       uint64  `json:"pauseNs"`       // nanoseconds in recent GC stop-the-world pause
	NumGC         uint32  `json:"numGC"`         // number of completed GC cycles
	NumForcedGC   uint32  `json:"numForcedGC"`   // number of GC cycles that were forced by the application calling the GC function
	GCCPUFraction float64 `json:"gcCPUFraction"` // fraction of this program's available CPU time used by the GC since the program started
	EnableGC      bool    `json:"enableGC"`      // boolean that indicates GC is enabled
	DebugGC       bool    `json:"debugGC"`       // boolean that indicates GC debug mode is enabled
	NumGoroutine  int     `json:"numGoroutine"`  // number of goroutines
}

// HandleGetMemoryStats
//
//	@summary returns current memory statistics.
//	@desc This returns real-time memory usage statistics from the Go runtime.
//	@route /api/v1/memory/stats [GET]
//	@returns handlers.MemoryStatsResponse
func (h *Handler) HandleGetMemoryStats(c echo.Context) error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Force garbage collection to get accurate memory stats
	// runtime.GC()
	runtime.ReadMemStats(&m)

	response := MemoryStatsResponse{
		Alloc:         m.Alloc,
		TotalAlloc:    m.TotalAlloc,
		Sys:           m.Sys,
		Lookups:       m.Lookups,
		Mallocs:       m.Mallocs,
		Frees:         m.Frees,
		HeapAlloc:     m.HeapAlloc,
		HeapSys:       m.HeapSys,
		HeapIdle:      m.HeapIdle,
		HeapInuse:     m.HeapInuse,
		HeapReleased:  m.HeapReleased,
		HeapObjects:   m.HeapObjects,
		StackInuse:    m.StackInuse,
		StackSys:      m.StackSys,
		MSpanInuse:    m.MSpanInuse,
		MSpanSys:      m.MSpanSys,
		MCacheInuse:   m.MCacheInuse,
		MCacheSys:     m.MCacheSys,
		BuckHashSys:   m.BuckHashSys,
		GCSys:         m.GCSys,
		OtherSys:      m.OtherSys,
		NextGC:        m.NextGC,
		LastGC:        m.LastGC,
		PauseTotalNs:  m.PauseTotalNs,
		PauseNs:       m.PauseNs[0], // Most recent pause
		NumGC:         m.NumGC,
		NumForcedGC:   m.NumForcedGC,
		GCCPUFraction: m.GCCPUFraction,
		EnableGC:      m.EnableGC,
		DebugGC:       m.DebugGC,
		NumGoroutine:  runtime.NumGoroutine(),
	}

	return h.RespondWithData(c, response)
}

// HandleGetMemoryProfile
//
//	@summary generates and returns a memory profile.
//	@desc This generates a memory profile that can be analyzed with go tool pprof.
//	@desc Query parameters: heap=true for heap profile, allocs=true for alloc profile.
//	@route /api/v1/memory/profile [GET]
//	@returns nil
func (h *Handler) HandleGetMemoryProfile(c echo.Context) error {
	// Parse query parameters
	heap := c.QueryParam("heap") == "true"
	allocs := c.QueryParam("allocs") == "true"

	// Default to heap profile if no specific type requested
	if !heap && !allocs {
		heap = true
	}

	// Set response headers for file download
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	var filename string
	var profile *pprof.Profile
	var err error

	if heap {
		filename = fmt.Sprintf("seanime-heap-profile-%s.pprof", timestamp)
		profile = pprof.Lookup("heap")
	} else if allocs {
		filename = fmt.Sprintf("seanime-allocs-profile-%s.pprof", timestamp)
		profile = pprof.Lookup("allocs")
	}

	if profile == nil {
		h.App.Logger.Error().Msg("handlers: Failed to lookup memory profile")
		return h.RespondWithError(c, fmt.Errorf("failed to lookup memory profile"))
	}

	c.Response().Header().Set("Content-Type", "application/octet-stream")
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// // Force garbage collection before profiling for more accurate results
	// runtime.GC()

	// Write profile to response
	if err = profile.WriteTo(c.Response().Writer, 0); err != nil {
		h.App.Logger.Error().Err(err).Msg("handlers: Failed to write memory profile")
		return h.RespondWithError(c, err)
	}

	return nil
}

// HandleGetGoRoutineProfile
//
//	@summary generates and returns a goroutine profile.
//	@desc This generates a goroutine profile showing all running goroutines and their stack traces.
//	@route /api/v1/memory/goroutine [GET]
//	@returns nil
func (h *Handler) HandleGetGoRoutineProfile(c echo.Context) error {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("seanime-goroutine-profile-%s.pprof", timestamp)

	profile := pprof.Lookup("goroutine")
	if profile == nil {
		h.App.Logger.Error().Msg("handlers: Failed to lookup goroutine profile")
		return h.RespondWithError(c, fmt.Errorf("failed to lookup goroutine profile"))
	}

	c.Response().Header().Set("Content-Type", "application/octet-stream")
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	if err := profile.WriteTo(c.Response().Writer, 0); err != nil {
		h.App.Logger.Error().Err(err).Msg("handlers: Failed to write goroutine profile")
		return h.RespondWithError(c, err)
	}

	return nil
}

// HandleGetCPUProfile
//
//	@summary generates and returns a CPU profile.
//	@desc This generates a CPU profile for the specified duration (default 30 seconds).
//	@desc Query parameter: duration=30 for duration in seconds.
//	@route /api/v1/memory/cpu [GET]
//	@returns nil
func (h *Handler) HandleGetCPUProfile(c echo.Context) error {
	// Parse duration from query parameter (default to 30 seconds)
	durationStr := c.QueryParam("duration")
	duration := 30 * time.Second
	if durationStr != "" {
		if d, err := strconv.Atoi(durationStr); err == nil && d > 0 && d <= 300 { // Max 5 minutes
			duration = time.Duration(d) * time.Second
		}
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("seanime-cpu-profile-%s.pprof", timestamp)

	c.Response().Header().Set("Content-Type", "application/octet-stream")
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// Start CPU profiling
	if err := pprof.StartCPUProfile(c.Response().Writer); err != nil {
		h.App.Logger.Error().Err(err).Msg("handlers: Failed to start CPU profile")
		return h.RespondWithError(c, err)
	}

	// Profile for the specified duration
	h.App.Logger.Info().Msgf("handlers: Starting CPU profile for %v", duration)
	time.Sleep(duration)

	// Stop CPU profiling
	pprof.StopCPUProfile()
	h.App.Logger.Info().Msg("handlers: CPU profile completed")

	return nil
}

// HandleForceGC
//
//	@summary forces garbage collection and returns memory stats.
//	@desc This forces a garbage collection cycle and returns the updated memory statistics.
//	@route /api/v1/memory/gc [POST]
//	@returns handlers.MemoryStatsResponse
func (h *Handler) HandleForceGC(c echo.Context) error {
	h.App.Logger.Info().Msg("handlers: Forcing garbage collection")

	// Force garbage collection
	runtime.GC()
	runtime.GC() // Run twice to ensure cleanup

	// Get updated memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	response := MemoryStatsResponse{
		Alloc:         m.Alloc,
		TotalAlloc:    m.TotalAlloc,
		Sys:           m.Sys,
		Lookups:       m.Lookups,
		Mallocs:       m.Mallocs,
		Frees:         m.Frees,
		HeapAlloc:     m.HeapAlloc,
		HeapSys:       m.HeapSys,
		HeapIdle:      m.HeapIdle,
		HeapInuse:     m.HeapInuse,
		HeapReleased:  m.HeapReleased,
		HeapObjects:   m.HeapObjects,
		StackInuse:    m.StackInuse,
		StackSys:      m.StackSys,
		MSpanInuse:    m.MSpanInuse,
		MSpanSys:      m.MSpanSys,
		MCacheInuse:   m.MCacheInuse,
		MCacheSys:     m.MCacheSys,
		BuckHashSys:   m.BuckHashSys,
		GCSys:         m.GCSys,
		OtherSys:      m.OtherSys,
		NextGC:        m.NextGC,
		LastGC:        m.LastGC,
		PauseTotalNs:  m.PauseTotalNs,
		PauseNs:       m.PauseNs[0],
		NumGC:         m.NumGC,
		NumForcedGC:   m.NumForcedGC,
		GCCPUFraction: m.GCCPUFraction,
		EnableGC:      m.EnableGC,
		DebugGC:       m.DebugGC,
		NumGoroutine:  runtime.NumGoroutine(),
	}

	h.App.Logger.Info().Msgf("handlers: GC completed, heap size: %d bytes", response.HeapAlloc)

	return h.RespondWithData(c, response)
}
