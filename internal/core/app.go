package core

import (
	"os"
	"runtime"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/constants"
	"seanime/internal/continuity"
	"seanime/internal/database/db"
	"seanime/internal/database/models"
	debrid_client "seanime/internal/debrid/client"
	discordrpc_presence "seanime/internal/discordrpc/presence"
	"seanime/internal/doh"
	"seanime/internal/events"
	"seanime/internal/extension_playground"
	"seanime/internal/extension_repo"
	"seanime/internal/hook"
	"seanime/internal/library/autodownloader"
	"seanime/internal/library/autoscanner"
	"seanime/internal/library/fillermanager"
	"seanime/internal/library/playbackmanager"
	"seanime/internal/library/scanner"
	"seanime/internal/manga"
	"seanime/internal/mediaplayers/mediaplayer"
	"seanime/internal/mediaplayers/mpchc"
	"seanime/internal/mediaplayers/mpv"
	"seanime/internal/mediaplayers/vlc"
	"seanime/internal/mediastream"
	"seanime/internal/onlinestream"
	"seanime/internal/platforms/anilist_platform"
	"seanime/internal/platforms/local_platform"
	"seanime/internal/platforms/platform"
	"seanime/internal/plugin"
	"seanime/internal/report"
	sync2 "seanime/internal/sync"
	"seanime/internal/torrent_clients/torrent_client"
	"seanime/internal/torrents/torrent"
	"seanime/internal/torrentstream"
	"seanime/internal/updater"
	"seanime/internal/util"
	"seanime/internal/util/filecache"
	"sync"

	"github.com/rs/zerolog"
)

type (
	App struct {
		Config                        *Config
		Database                      *db.Database
		Logger                        *zerolog.Logger
		TorrentClientRepository       *torrent_client.Repository
		TorrentRepository             *torrent.Repository
		DebridClientRepository        *debrid_client.Repository
		Watcher                       *scanner.Watcher
		AnilistClient                 anilist.AnilistClient
		AnilistPlatform               platform.Platform
		LocalPlatform                 platform.Platform
		SyncManager                   sync2.Manager
		FillerManager                 *fillermanager.FillerManager
		WSEventManager                *events.WSEventManager
		AutoDownloader                *autodownloader.AutoDownloader
		ExtensionRepository           *extension_repo.Repository
		ExtensionPlaygroundRepository *extension_playground.PlaygroundRepository
		MediaPlayer                   struct {
			VLC   *vlc.VLC
			MpcHc *mpchc.MpcHc
			Mpv   *mpv.Mpv
		}
		MediaPlayerRepository   *mediaplayer.Repository
		Version                 string
		Updater                 *updater.Updater
		AutoScanner             *autoscanner.AutoScanner
		PlaybackManager         *playbackmanager.PlaybackManager
		FileCacher              *filecache.Cacher
		OnlinestreamRepository  *onlinestream.Repository
		MangaRepository         *manga.Repository
		MetadataProvider        metadata.Provider
		DiscordPresence         *discordrpc_presence.Presence
		MangaDownloader         *manga.Downloader
		ContinuityManager       *continuity.Manager
		Cleanups                []func()
		OnFlushLogs             func()
		MediastreamRepository   *mediastream.Repository
		TorrentstreamRepository *torrentstream.Repository
		FeatureFlags            FeatureFlags
		Settings                *models.Settings
		SecondarySettings       struct {
			Mediastream   *models.MediastreamSettings
			Torrentstream *models.TorrentstreamSettings
			Debrid        *models.DebridSettings
		} // Struct for other settings sent to client
		SelfUpdater        *updater.SelfUpdater
		ReportRepository   *report.Repository
		TotalLibrarySize   uint64 // Initialized in modules.go
		LibraryDir         string
		IsDesktopSidecar   bool
		animeCollection    *anilist.AnimeCollection
		rawAnimeCollection *anilist.AnimeCollection // (retains custom lists)
		mangaCollection    *anilist.MangaCollection
		rawMangaCollection *anilist.MangaCollection // (retains custom lists)
		account            *models.Account
		previousVersion    string
		moduleMu           sync.Mutex
		HookManager        hook.Manager
		ServerReady        bool // Whether the Anilist data from the first request has been fetched
	}
)

// NewApp creates a new server instance
func NewApp(configOpts *ConfigOptions, selfupdater *updater.SelfUpdater) *App {

	// Initialize logger with predefined format
	logger := util.NewLogger()

	// Log application version, OS, architecture and system info
	logger.Info().Msgf("app: Seanime %s-%s", constants.Version, constants.VersionName)
	logger.Info().Msgf("app: OS: %s", runtime.GOOS)
	logger.Info().Msgf("app: Arch: %s", runtime.GOARCH)
	logger.Info().Msgf("app: Processor count: %d", runtime.NumCPU())

	// Initialize hook manager for plugin event system
	hookManager := hook.NewHookManager(hook.NewHookManagerOptions{Logger: logger})
	hook.SetGlobalHookManager(hookManager)
	plugin.GlobalAppContext.SetLogger(logger)

	// Store current version to detect version changes
	previousVersion := constants.Version

	// Add callback to track version changes
	configOpts.OnVersionChange = append(configOpts.OnVersionChange, func(oldVersion string, newVersion string) {
		logger.Info().Str("prev", oldVersion).Str("current", newVersion).Msg("app: Version change detected")
		previousVersion = oldVersion
	})

	// Initialize configuration with provided options
	// Creates config directory if it doesn't exist
	cfg, err := NewConfig(configOpts, logger)
	if err != nil {
		logger.Fatal().Err(err).Msgf("app: Failed to initialize config")
	}

	// Create logs directory if it doesn't exist
	_ = os.MkdirAll(cfg.Logs.Dir, 0755)

	// Start background process to trim log files
	go TrimLogEntries(cfg.Logs.Dir, logger)

	logger.Info().Msgf("app: Data directory: %s", cfg.Data.AppDataDir)
	logger.Info().Msgf("app: Working directory: %s", cfg.Data.WorkingDir)

	// Log if running in desktop sidecar mode
	if configOpts.IsDesktopSidecar {
		logger.Info().Msg("app: Desktop sidecar mode enabled")
	}

	// Initialize database connection
	database, err := db.NewDatabase(cfg.Data.AppDataDir, cfg.Database.Name, logger)
	if err != nil {
		logger.Fatal().Err(err).Msgf("app: Failed to initialize database")
	}

	HandleNewDatabaseEntries(database, logger)

	// Clean up old database entries in background goroutines
	database.TrimLocalFileEntries()     // Remove old local file entries
	database.TrimScanSummaryEntries()   // Remove old scan summaries
	database.TrimTorrentstreamHistory() // Remove old torrent stream history

	// Get anime library paths for plugin context
	animeLibraryPaths, _ := database.GetAllLibraryPathsFromSettings()
	plugin.GlobalAppContext.SetModulesPartial(plugin.AppContextModules{
		Database:          database,
		AnimeLibraryPaths: &animeLibraryPaths,
	})

	// Get Anilist token from database if available
	anilistToken := database.GetAnilistToken()

	// Initialize Anilist API client
	anilistCW := anilist.NewAnilistClient(anilistToken)

	// Initialize WebSocket event manager for real-time communication
	wsEventManager := events.NewWSEventManager(logger)

	// Exit if no WebSocket connections in desktop sidecar mode
	if configOpts.IsDesktopSidecar {
		wsEventManager.ExitIfNoConnsAsDesktopSidecar()
	}

	// Initialize DNS-over-HTTPS service in background
	go doh.HandleDoH(cfg.Server.DoHUrl, logger)

	// Initialize file cache system for media and metadata
	fileCacher, err := filecache.NewCacher(cfg.Cache.Dir)
	if err != nil {
		logger.Fatal().Err(err).Msgf("app: Failed to initialize file cacher")
	}

	// Initialize extension repository
	extensionRepository := extension_repo.NewRepository(&extension_repo.NewRepositoryOptions{
		Logger:         logger,
		ExtensionDir:   cfg.Extensions.Dir,
		WSEventManager: wsEventManager,
		FileCacher:     fileCacher,
		HookManager:    hookManager,
	})
	// Load extensions in background
	go LoadExtensions(extensionRepository, logger)

	// Initialize metadata provider for media information
	metadataProvider := metadata.NewProvider(&metadata.NewProviderImplOptions{
		Logger:     logger,
		FileCacher: fileCacher,
	})

	// Set initial metadata provider (will change if offline mode is enabled)
	activeMetadataProvider := metadataProvider

	// Initialize manga repository
	mangaRepository := manga.NewRepository(&manga.NewRepositoryOptions{
		Logger:         logger,
		FileCacher:     fileCacher,
		CacheDir:       cfg.Cache.Dir,
		ServerURI:      cfg.GetServerURI(),
		WsEventManager: wsEventManager,
		DownloadDir:    cfg.Manga.DownloadDir,
		Database:       database,
	})

	// Initialize Anilist platform
	anilistPlatform := anilist_platform.NewAnilistPlatform(anilistCW, logger)

	// Update plugin context with new modules
	plugin.GlobalAppContext.SetModulesPartial(plugin.AppContextModules{
		AnilistPlatform:  anilistPlatform,
		WSEventManager:   wsEventManager,
		MetadataProvider: metadataProvider,
	})

	// Initialize sync manager for offline/online synchronization
	syncManager, err := sync2.NewManager(&sync2.NewManagerOptions{
		LocalDir:         cfg.Offline.Dir,
		AssetDir:         cfg.Offline.AssetDir,
		Logger:           logger,
		MetadataProvider: metadataProvider,
		MangaRepository:  mangaRepository,
		Database:         database,
		WSEventManager:   wsEventManager,
		IsOffline:        cfg.Server.Offline,
		AnilistPlatform:  anilistPlatform,
	})
	if err != nil {
		logger.Fatal().Err(err).Msgf("app: Failed to initialize sync manager")
	}

	// Use local metadata provider if in offline mode
	if cfg.Server.Offline {
		activeMetadataProvider = syncManager.GetLocalMetadataProvider()
	}

	// Initialize local platform for offline operation
	localPlatform, err := local_platform.NewLocalPlatform(syncManager, anilistCW, logger)
	if err != nil {
		logger.Fatal().Err(err).Msgf("app: Failed to initialize local platform")
	}

	// Change active platform if offline mode is enabled
	activePlatform := anilistPlatform
	if cfg.Server.Offline {
		activePlatform = localPlatform
	}

	// Initialize online streaming repository
	onlinestreamRepository := onlinestream.NewRepository(&onlinestream.NewRepositoryOptions{
		Logger:           logger,
		FileCacher:       fileCacher,
		MetadataProvider: activeMetadataProvider,
		Platform:         activePlatform,
		Database:         database,
	})

	// Initialize extension playground for testing extensions
	extensionPlaygroundRepository := extension_playground.NewPlaygroundRepository(logger, activePlatform, activeMetadataProvider)

	// Create the main app instance with initialized components
	app := &App{
		Config:                        cfg,
		Database:                      database,
		AnilistClient:                 anilistCW,
		AnilistPlatform:               activePlatform,
		LocalPlatform:                 localPlatform,
		SyncManager:                   syncManager,
		WSEventManager:                wsEventManager,
		Logger:                        logger,
		Version:                       constants.Version,
		Updater:                       updater.New(constants.Version, logger, wsEventManager),
		FileCacher:                    fileCacher,
		OnlinestreamRepository:        onlinestreamRepository,
		MetadataProvider:              activeMetadataProvider,
		MangaRepository:               mangaRepository,
		ExtensionRepository:           extensionRepository,
		ExtensionPlaygroundRepository: extensionPlaygroundRepository,
		ReportRepository:              report.NewRepository(logger),
		TorrentRepository:             nil, // Initialized in App.initModulesOnce
		FillerManager:                 nil, // Initialized in App.initModulesOnce
		MangaDownloader:               nil, // Initialized in App.initModulesOnce
		PlaybackManager:               nil, // Initialized in App.initModulesOnce
		AutoDownloader:                nil, // Initialized in App.initModulesOnce
		AutoScanner:                   nil, // Initialized in App.initModulesOnce
		MediastreamRepository:         nil, // Initialized in App.initModulesOnce
		TorrentstreamRepository:       nil, // Initialized in App.initModulesOnce
		ContinuityManager:             nil, // Initialized in App.initModulesOnce
		DebridClientRepository:        nil, // Initialized in App.initModulesOnce
		TorrentClientRepository:       nil, // Initialized in App.InitOrRefreshModules
		MediaPlayerRepository:         nil, // Initialized in App.InitOrRefreshModules
		DiscordPresence:               nil, // Initialized in App.InitOrRefreshModules
		previousVersion:               previousVersion,
		FeatureFlags:                  NewFeatureFlags(cfg, logger),
		IsDesktopSidecar:              configOpts.IsDesktopSidecar,
		SecondarySettings: struct {
			Mediastream   *models.MediastreamSettings
			Torrentstream *models.TorrentstreamSettings
			Debrid        *models.DebridSettings
		}{Mediastream: nil, Torrentstream: nil},
		SelfUpdater: selfupdater,
		moduleMu:    sync.Mutex{},
		HookManager: hookManager,
	}

	// Run database migrations if version has changed
	app.runMigrations()

	// Initialize modules that only need to be initialized once
	app.initModulesOnce()

	plugin.GlobalAppContext.SetModulesPartial(plugin.AppContextModules{
		ContinuityManager:       app.ContinuityManager,
		AutoScanner:             app.AutoScanner,
		AutoDownloader:          app.AutoDownloader,
		FileCacher:              app.FileCacher,
		OnlinestreamRepository:  app.OnlinestreamRepository,
		MediastreamRepository:   app.MediastreamRepository,
		TorrentstreamRepository: app.TorrentstreamRepository,
	})

	// Initialize all modules that depend on settings
	app.InitOrRefreshModules()

	// Load built-in extensions into extension consumers
	app.AddExtensionBankToConsumers()

	// Initialize Anilist data if not in offline mode
	if !app.IsOffline() {
		app.InitOrRefreshAnilistData()
	} else {
		app.ServerReady = true
	}

	// Initialize mediastream settings (for streaming media)
	app.InitOrRefreshMediastreamSettings()

	// Initialize torrentstream settings (for torrent streaming)
	app.InitOrRefreshTorrentstreamSettings()

	// Initialize debrid settings (for debrid services)
	app.InitOrRefreshDebridSettings()

	// Run one-time initialization actions
	app.performActionsOnce()

	return app
}

func (a *App) IsOffline() bool {
	if a.Config == nil {
		return false
	}

	return a.Config.Server.Offline
}

func (a *App) AddCleanupFunction(f func()) {
	a.Cleanups = append(a.Cleanups, f)
}

func (a *App) Cleanup() {
	for _, f := range a.Cleanups {
		f()
	}
}
