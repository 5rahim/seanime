package core

import (
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/api/anizip"
	"github.com/seanime-app/seanime/internal/api/listsync"
	"github.com/seanime-app/seanime/internal/api/metadata"
	"github.com/seanime-app/seanime/internal/constants"
	"github.com/seanime-app/seanime/internal/database/db"
	"github.com/seanime-app/seanime/internal/database/models"
	"github.com/seanime-app/seanime/internal/discordrpc/presence"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/library/anime"
	"github.com/seanime-app/seanime/internal/library/autodownloader"
	"github.com/seanime-app/seanime/internal/library/autoscanner"
	"github.com/seanime-app/seanime/internal/library/playbackmanager"
	"github.com/seanime-app/seanime/internal/library/scanner"
	"github.com/seanime-app/seanime/internal/manga"
	"github.com/seanime-app/seanime/internal/mediaplayers/mediaplayer"
	"github.com/seanime-app/seanime/internal/mediaplayers/mpchc"
	"github.com/seanime-app/seanime/internal/mediaplayers/mpv"
	"github.com/seanime-app/seanime/internal/mediaplayers/vlc"
	"github.com/seanime-app/seanime/internal/mediastream"
	"github.com/seanime-app/seanime/internal/offline"
	"github.com/seanime-app/seanime/internal/onlinestream"
	"github.com/seanime-app/seanime/internal/torrents/animetosho"
	"github.com/seanime-app/seanime/internal/torrents/nyaa"
	"github.com/seanime-app/seanime/internal/torrents/torrent_client"
	"github.com/seanime-app/seanime/internal/torrentstream"
	"github.com/seanime-app/seanime/internal/updater"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/seanime-app/seanime/internal/util/filecache"
	"log"
	"runtime"
	"strings"
)

type (
	App struct {
		Config                  *Config
		Database                *db.Database
		Logger                  *zerolog.Logger
		TorrentClientRepository *torrent_client.Repository
		Watcher                 *scanner.Watcher
		AnizipCache             *anizip.Cache // AnizipCache holds fetched AniZip media for 30 minutes. (used by route handlers)
		AnilistClientWrapper    anilist.ClientWrapperInterface
		NyaaSearchCache         *nyaa.SearchCache
		AnimeToshoSearchCache   *animetosho.SearchCache
		anilistCollection       *anilist.AnimeCollection // should be retrieved with Get funcs in routes
		mangaCollection         *anilist.MangaCollection // should be retrieved with Get funcs in routes
		account                 *models.Account          // should be retrieved with Get funcs in routes
		WSEventManager          *events.WSEventManager
		ListSyncCache           *listsync.Cache
		AutoDownloader          *autodownloader.AutoDownloader
		MediaPlayer             struct {
			VLC   *vlc.VLC
			MpcHc *mpchc.MpcHc
			Mpv   *mpv.Mpv
		}
		MediaPlayerRepository   *mediaplayer.Repository
		Version                 string
		Updater                 *updater.Updater
		Settings                *models.Settings
		AutoScanner             *autoscanner.AutoScanner
		PlaybackManager         *playbackmanager.PlaybackManager
		FileCacher              *filecache.Cacher
		Onlinestream            *onlinestream.OnlineStream
		MangaRepository         *manga.Repository
		MetadataProvider        *metadata.Provider
		WD                      string // Working directory
		DiscordPresence         *discordrpc_presence.Presence
		MangaDownloader         *manga.Downloader
		Cleanups                []func()
		cancelContext           func()
		previousVersion         string
		OfflineHub              *offline.Hub
		MediastreamRepository   *mediastream.Repository
		TorrentstreamRepository *torrentstream.Repository
		FeatureFlags            FeatureFlags
		SecondarySettings       struct {
			Mediastream   *models.MediastreamSettings
			Torrentstream *models.TorrentstreamSettings
		}
	}
)

// NewApp creates a new server instance
func NewApp(configOpts *ConfigOptions) *App {
	logger := util.NewLogger()

	logger.Info().Msgf("app: Seanime %s-%s", constants.Version, constants.VersionName)
	logger.Info().Msgf("app: OS: %s", runtime.GOOS)
	logger.Info().Msgf("app: Arch: %s", runtime.GOARCH)
	logger.Info().Msgf("app: Processor count: %d", runtime.NumCPU())

	previousVersion := constants.Version

	configOpts.OnVersionChange = append(configOpts.OnVersionChange, func(oldVersion string, newVersion string) {
		logger.Info().Str("prev", oldVersion).Str("current", newVersion).Msg("app: Version change detected")
		previousVersion = oldVersion
	})

	// Initialize the config
	// If the config dir does not exist, it will be created
	cfg, err := NewConfig(configOpts, logger)
	if err != nil {
		logger.Fatal().Err(err).Msgf("app: Failed to initialize config")
	}

	logger.Info().Msgf("app: Data directory: %s", cfg.Data.AppDataDir)

	// Print working directory
	logger.Info().Msgf("app: Working directory: %s", cfg.Data.WorkingDir)

	// Initialize the database
	database, err := db.NewDatabase(cfg.Data.AppDataDir, cfg.Database.Name, logger)
	if err != nil {
		logger.Fatal().Err(err).Msgf("app: Failed to initialize database")
	}

	// Add default local file entries if there are none
	if _, _, err = database.GetLocalFiles(); err != nil {
		_, err = database.InsertLocalFiles(make([]*anime.LocalFile, 0))
		if err != nil {
			logger.Fatal().Err(err).Msgf("app: Failed to initialize local files in the database")
		}
	}

	database.TrimLocalFileEntries()
	database.TrimScanSummaryEntries()

	// Get token from stored account or return empty string
	anilistToken := database.GetAnilistToken()

	// Anilist Client Wrapper
	anilistCW := anilist.NewClientWrapper(anilistToken)

	// Websocket Event Manager
	wsEventManager := events.NewWSEventManager(logger)

	// AniZip Cache
	anizipCache := anizip.NewCache()

	// File Cacher
	fileCacher, err := filecache.NewCacher(cfg.Cache.Dir)
	if err != nil {
		logger.Fatal().Err(err).Msgf("app: Failed to initialize file cacher")
	}

	// Online Stream
	onlineStream := onlinestream.New(&onlinestream.NewOnlineStreamOptions{
		Logger:               logger,
		FileCacher:           fileCacher,
		AnizipCache:          anizipCache,
		AnilistClientWrapper: anilistCW,
	})

	// Metadata Provider
	metadataProvider := metadata.NewProvider(&metadata.NewProviderOptions{
		Logger:     logger,
		FileCacher: fileCacher,
	})

	// Manga Repository
	mangaRepository := manga.NewRepository(&manga.NewRepositoryOptions{
		Logger:         logger,
		FileCacher:     fileCacher,
		BackupDir:      cfg.Manga.DownloadDir,
		ServerURI:      cfg.GetServerURI(),
		WsEventManager: wsEventManager,
		DownloadDir:    cfg.Manga.DownloadDir,
	})

	// Offline Hub
	// Will exit if offline mode is enabled and no snapshots are found
	offlineHub := offline.NewHub(&offline.NewHubOptions{
		AnilistClientWrapper: anilistCW,
		MetadataProvider:     metadataProvider,
		MangaRepository:      mangaRepository,
		WSEventManager:       wsEventManager,
		Database:             database,
		FileCacher:           fileCacher,
		Logger:               logger,
		OfflineDir:           cfg.Offline.Dir,
		AssetDir:             cfg.Offline.AssetDir,
		IsOffline:            cfg.Server.Offline,
	})

	app := &App{
		Config:                  cfg,
		Database:                database,
		AnilistClientWrapper:    anilistCW,
		AnizipCache:             anizipCache,
		NyaaSearchCache:         nyaa.NewSearchCache(),
		AnimeToshoSearchCache:   animetosho.NewSearchCache(),
		WSEventManager:          wsEventManager,
		ListSyncCache:           listsync.NewCache(),
		Logger:                  logger,
		Version:                 constants.Version,
		Updater:                 updater.New(constants.Version, logger),
		FileCacher:              fileCacher,
		Onlinestream:            onlineStream,
		MetadataProvider:        metadataProvider,
		MangaRepository:         mangaRepository,
		MangaDownloader:         nil, // Initialized in App.initModulesOnce
		PlaybackManager:         nil, // Initialized in App.initModulesOnce
		AutoDownloader:          nil, // Initialized in App.initModulesOnce
		AutoScanner:             nil, // Initialized in App.initModulesOnce
		TorrentClientRepository: nil, // Initialized in App.InitOrRefreshModules
		MediaPlayerRepository:   nil, // Initialized in App.InitOrRefreshModules
		DiscordPresence:         nil, // Initialized in App.InitOrRefreshModules
		MediastreamRepository:   nil, // Initialized in App.initModulesOnce
		TorrentstreamRepository: nil, // Initialized in App.initModulesOnce
		WD:                      cfg.Data.WorkingDir,
		previousVersion:         previousVersion,
		OfflineHub:              offlineHub,
		FeatureFlags:            NewFeatureFlags(cfg, logger),
		SecondarySettings: struct {
			Mediastream   *models.MediastreamSettings
			Torrentstream *models.TorrentstreamSettings
		}{Mediastream: nil, Torrentstream: nil},
	}

	app.runMigrations()
	app.initModulesOnce()
	app.InitOrRefreshModules()

	app.InitOrRefreshMediastreamSettings()

	app.InitOrRefreshTorrentstreamSettings()

	app.launchModulesOnce()

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

	app.Logger.Info().Msgf("app: Web interface path: %s", app.Config.Web.Dir)
	fiberApp.Static("/", app.Config.Web.Dir, fiber.Static{
		Index:    "index.html",
		Compress: true,
	})

	app.Logger.Info().Msgf("app: Web assets path: %s", app.Config.Web.AssetDir)
	fiberApp.Static("/assets", app.Config.Web.AssetDir, fiber.Static{
		Index:    "index.html",
		Compress: false,
	})

	if app.Config.Manga.DownloadDir != "" {
		app.Logger.Info().Msgf("app: Manga downloads path: %s", app.Config.Manga.DownloadDir)
		fiberApp.Static("/manga-downloads", app.Config.Manga.DownloadDir, fiber.Static{
			Index:    "index.html",
			Compress: false,
		})
	}

	if app.IsOffline() {
		app.Logger.Info().Msgf("app: Offline assets path: %s", app.Config.Offline.AssetDir)
		fiberApp.Static("/offline-assets", app.Config.Offline.AssetDir, fiber.Static{
			Index:    "index.html",
			Compress: false,
		})
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
		if path == "/.html" {
			path = "/index.html"
		}
		return c.SendFile(app.Config.Web.Dir + path)
	})

	return fiberApp
}

// RunServer starts the server
func RunServer(app *App, fiberApp *fiber.App) {
	app.Logger.Info().Msgf("app: Server Address: %s", app.Config.GetServerAddr())

	// DEVNOTE: Crashes self-update loop
	//app.Cleanups = append(app.Cleanups, func() {
	//	_ = fiberApp.ShutdownWithTimeout(time.Millisecond)
	//})

	// Start the server
	go func() {
		log.Fatal(fiberApp.Listen(app.Config.GetServerAddr()))
	}()

	app.Logger.Info().Msg("app: Seanime started at " + app.Config.GetServerURI())
}

func (a *App) Cleanup() {
	for _, f := range a.Cleanups {
		f()
	}
}
