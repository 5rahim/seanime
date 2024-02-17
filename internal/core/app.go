package core

import (
	"fmt"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/anizip"
	"github.com/seanime-app/seanime/internal/autodownloader"
	"github.com/seanime-app/seanime/internal/constants"
	_db "github.com/seanime-app/seanime/internal/db"
	"github.com/seanime-app/seanime/internal/entities"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/listsync"
	"github.com/seanime-app/seanime/internal/models"
	"github.com/seanime-app/seanime/internal/mpchc"
	"github.com/seanime-app/seanime/internal/mpv"
	"github.com/seanime-app/seanime/internal/nyaa"
	"github.com/seanime-app/seanime/internal/qbittorrent"
	"github.com/seanime-app/seanime/internal/scanner"
	"github.com/seanime-app/seanime/internal/updater"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/seanime-app/seanime/internal/vlc"
	"log"
	"strings"
)

type (
	App struct {
		Config               *Config
		Database             *_db.Database
		Logger               *zerolog.Logger
		QBittorrent          *qbittorrent.Client
		Watcher              *scanner.Watcher
		AnizipCache          *anizip.Cache // AnizipCache holds fetched AniZip media for 30 minutes. (used by route handlers)
		AnilistClientWrapper *anilist.ClientWrapper
		NyaaSearchCache      *nyaa.SearchCache
		anilistCollection    *anilist.AnimeCollection
		account              *models.Account
		WSEventManager       *events.WSEventManager
		ListSyncCache        *listsync.Cache // DEVNOTE: Shelved
		AutoDownloader       *autodownloader.AutoDownloader
		MediaPlayer          struct {
			VLC   *vlc.VLC
			MpcHc *mpchc.MpcHc
			Mpv   *mpv.Mpv
		}
		Version string
		Updater *updater.Updater
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

	// Initialize the config
	// If the config file does not exist, it will be created
	cfg, err := NewConfig(opts.Config)
	if err != nil {
		logger.Fatal().Err(err).Msgf("app: Failed to initialize config")
	}

	logger.Info().Msgf("app: Loaded config from \"%s\"", cfg.Data.AppDataDir)

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

	logger.Info().Msgf("app: Connected to database \"%s.db\"", cfg.Database.Name)

	// Get token from stored account or return empty string
	anilistToken := db.GetAnilistToken()

	// Websocket Event Manager
	wsEventManager := events.NewWSEventManager(logger)

	// AniZip Cache
	anizipCache := anizip.NewCache()

	// Auto downloader
	nAutoDownloader := autodownloader.NewAutoDownloader(&autodownloader.NewAutoDownloaderOptions{
		Logger:            logger,
		QbittorrentClient: nil, // Will be set in app.InitOrRefreshModules
		AnilistCollection: nil, // Will be set and refreshed in app.RefreshAnilistCollection
		Database:          db,
		WSEventManager:    wsEventManager,
		AniZipCache:       anizipCache,
	})

	app := &App{
		Config:               cfg,
		Database:             db,
		AnilistClientWrapper: anilist.NewClientWrapper(anilistToken),
		AnizipCache:          anizipCache,
		NyaaSearchCache:      nyaa.NewSearchCache(),
		WSEventManager:       wsEventManager,
		ListSyncCache:        listsync.NewCache(),
		AutoDownloader:       nAutoDownloader,
		Logger:               logger,
		Version:              version,
		Updater:              updater.New(version),
	}

	app.InitOrRefreshModules()

	// Initialize the AutoDownloader
	app.initAutoDownloader()

	return app
}

func NewFiberApp(app *App) *fiber.App {
	// Create a new fiber app
	fiberApp := fiber.New(fiber.Config{
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
		DisableStartupMessage: true,
	})

	if constants.DevelopmentWebBuild {
		fiberApp.Static("/", "./seanime-web/web")
	} else {
		fiberApp.Static("/", "./web")
	}

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
		if constants.DevelopmentWebBuild {
			return c.SendFile("./seanime-web/web" + path)
		} else {
			return c.SendFile("./web" + path)
		}
	})

	return fiberApp
}

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
