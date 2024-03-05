package core

import (
	"fmt"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/animetosho"
	"github.com/seanime-app/seanime/internal/anizip"
	"github.com/seanime-app/seanime/internal/autodownloader"
	_db "github.com/seanime-app/seanime/internal/db"
	"github.com/seanime-app/seanime/internal/entities"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/listsync"
	"github.com/seanime-app/seanime/internal/models"
	"github.com/seanime-app/seanime/internal/mpchc"
	"github.com/seanime-app/seanime/internal/mpv"
	"github.com/seanime-app/seanime/internal/nyaa"
	"github.com/seanime-app/seanime/internal/scanner"
	"github.com/seanime-app/seanime/internal/torrent_client"
	"github.com/seanime-app/seanime/internal/updater"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/seanime-app/seanime/internal/vlc"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type (
	App struct {
		Config                  *Config
		Database                *_db.Database
		Logger                  *zerolog.Logger
		TorrentClientRepository *torrent_client.Repository
		Watcher                 *scanner.Watcher
		AnizipCache             *anizip.Cache // AnizipCache holds fetched AniZip media for 30 minutes. (used by route handlers)
		AnilistClientWrapper    *anilist.ClientWrapper
		NyaaSearchCache         *nyaa.SearchCache
		AnimeToshoSearchCache   *animetosho.SearchCache
		anilistCollection       *anilist.AnimeCollection
		account                 *models.Account
		WSEventManager          *events.WSEventManager
		ListSyncCache           *listsync.Cache
		AutoDownloader          *autodownloader.AutoDownloader
		MediaPlayer             struct {
			VLC   *vlc.VLC
			MpcHc *mpchc.MpcHc
			Mpv   *mpv.Mpv
		}
		Version     string
		Updater     *updater.Updater
		Settings    *models.Settings
		AutoScanner *scanner.AutoScanner
	}

	AppOptions struct {
		Config *ConfigOptions
	}
)

var DefaultAppOptions = AppOptions{
	Config: &DefaultConfig,
}

// NewApp creates a new server instance
func NewApp(options *AppOptions, version string) *App {

	opts := *options

	// Set up a default config if none is provided
	if options.Config == nil {
		opts.Config = &DefaultConfig
	}

	logger := util.NewLogger()

	// Print working directory
	pwd, err := os.Getwd()
	if err != nil {
		logger.Fatal().Err(err).Msg("app: Failed to get working directory")
	}

	logger.Debug().Msgf("app: Working directory: \"%s\"", filepath.ToSlash(pwd))

	// Initialize the config
	// If the config file does not exist, it will be created
	cfg, err := NewConfig(opts.Config)
	if err != nil {
		logger.Fatal().Err(err).Msgf("app: Failed to initialize config")
	}

	logger.Debug().Msgf("app: Loaded config from \"%s\"", cfg.Data.AppDataDir)

	// Initialize the database
	db, err := _db.NewDatabase(cfg.Data.AppDataDir, cfg.Database.Name, logger)
	if err != nil {
		logger.Fatal().Err(err).Msgf("app: Failed to initialize database")
	}

	// Add default local file entries if there are none
	if _, _, err := db.GetLocalFiles(); err != nil {
		_, err := db.InsertLocalFiles(make([]*entities.LocalFile, 0))
		if err != nil {
			logger.Fatal().Err(err).Msgf("app: Failed to initialize local files in the database")
		}
	}

	// Delete old local file entries
	db.CleanUpLocalFiles()
	// Delete old scan summaries
	db.CleanUpScanSummaries()

	logger.Debug().Msgf("app: Connected to database \"%s.db\"", cfg.Database.Name)

	// Get token from stored account or return empty string
	anilistToken := db.GetAnilistToken()

	// Websocket Event Manager
	wsEventManager := events.NewWSEventManager(logger)

	// AniZip Cache
	anizipCache := anizip.NewCache()

	app := &App{
		Config:                  cfg,
		Database:                db,
		AnilistClientWrapper:    anilist.NewClientWrapper(anilistToken),
		AnizipCache:             anizipCache,
		NyaaSearchCache:         nyaa.NewSearchCache(),
		AnimeToshoSearchCache:   animetosho.NewSearchCache(),
		WSEventManager:          wsEventManager,
		ListSyncCache:           listsync.NewCache(),
		Logger:                  logger,
		Version:                 version,
		Updater:                 updater.New(version),
		AutoDownloader:          nil, // Initialized in App.InitModulesOnce
		AutoScanner:             nil, // Initialized in App.InitModulesOnce
		TorrentClientRepository: nil, // Initialized in App.InitOrRefreshModules
	}

	app.InitModulesOnce()
	app.InitOrRefreshModules()

	return app
}

// NewFiberApp creates a new fiber app instance
// and sets up the static file server for the web interface.
func NewFiberApp(app *App) *fiber.App {
	// Create a new fiber app
	fiberApp := fiber.New(fiber.Config{
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
		DisableStartupMessage: true,
	})

	app.Logger.Debug().Msgf("app: Serving web interface from \"%s\"", app.Config.Web.Dir)
	fiberApp.Static("/", app.Config.Web.Dir, fiber.Static{
		Index:    "index.html",
		Compress: true,
	})

	app.Logger.Debug().Msgf("app: Serving web assets from \"%s\"", app.Config.Web.AssetDir)
	fiberApp.Static("/assets", app.Config.Web.AssetDir, fiber.Static{
		Index:    "index.html",
		Compress: false,
	})

	fiberApp.Get("*", func(c *fiber.Ctx) error {
		path := c.OriginalURL()
		if strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/events") {
			return c.Next()
		}
		if !strings.HasSuffix(path, ".html") {
			if strings.Contains(path, "?") {
				// Split the path into the actual path and the query string
				parts := strings.SplitN(path, "?", 2)
				actualPath := parts[0]
				queryString := parts[1]

				// Add ".html" to the actual path
				actualPath += ".html"

				// Reassemble the path with the query string
				path = actualPath + "?" + queryString
			} else {
				path += ".html"
			}
		}
		if path == "/.html" {
			path = "/index.html"
		}
		return c.SendFile("./web" + path)
	})

	return fiberApp
}

// RunServer starts the server
func RunServer(app *App, fiberApp *fiber.App) {
	addr := fmt.Sprintf("%s:%d", app.Config.Server.Host, app.Config.Server.Port)

	// Start the server
	go func() {
		log.Fatal(fiberApp.Listen(addr))
	}()

	pAddr := fmt.Sprintf("http://%s:%d", app.Config.Server.Host, app.Config.Server.Port)
	if app.Config.Server.Host == "" {
		pAddr = fmt.Sprintf(":%d", app.Config.Server.Port)
	}

	app.Logger.Info().Msg("Seanime started at " + pAddr)

}
