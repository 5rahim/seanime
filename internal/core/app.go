package core

import (
	"os"
	"path/filepath"
	"runtime"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata_provider"
	"seanime/internal/constants"
	"seanime/internal/continuity"
	"seanime/internal/database/db"
	"seanime/internal/database/models"
	debrid_client "seanime/internal/debrid/client"
	"seanime/internal/directstream"
	discordrpc_presence "seanime/internal/discordrpc/presence"
	"seanime/internal/doh"
	"seanime/internal/events"
	"seanime/internal/extension"
	"seanime/internal/extension_playground"
	"seanime/internal/extension_repo"
	"seanime/internal/hook"
	"seanime/internal/library/autodownloader"
	"seanime/internal/library/autoscanner"
	"seanime/internal/library/fillermanager"
	"seanime/internal/library/playbackmanager"
	"seanime/internal/library/scanner"
	"seanime/internal/library_explorer"
	"seanime/internal/local"
	"seanime/internal/manga"
	"seanime/internal/mediaplayers/iina"
	"seanime/internal/mediaplayers/mediaplayer"
	"seanime/internal/mediaplayers/mpchc"
	"seanime/internal/mediaplayers/mpv"
	"seanime/internal/mediaplayers/vlc"
	"seanime/internal/mediastream"
	"seanime/internal/nakama"
	"seanime/internal/nativeplayer"
	"seanime/internal/onlinestream"
	"seanime/internal/platforms/anilist_platform"
	"seanime/internal/platforms/offline_platform"
	"seanime/internal/platforms/platform"
	"seanime/internal/platforms/simulated_platform"
	"seanime/internal/playlist"
	"seanime/internal/plugin"
	"seanime/internal/report"
	"seanime/internal/torrent_clients/torrent_client"
	"seanime/internal/torrents/torrent"
	"seanime/internal/torrentstream"
	"seanime/internal/updater"
	"seanime/internal/user"
	"seanime/internal/util"
	"seanime/internal/util/filecache"
	"seanime/internal/util/result"
	"seanime/internal/videocore"
	"sync"
	"sync/atomic"

	"github.com/rs/zerolog"
)

type (
	App struct {
		// Core
		Config   *Config
		Database *db.Database
		Logger   *zerolog.Logger

		// Torrent and debrid services
		TorrentClientRepository *torrent_client.Repository
		TorrentRepository       *torrent.Repository
		DebridClientRepository  *debrid_client.Repository

		// File system monitoring
		Watcher *scanner.Watcher

		// API clients and providers
		AnilistClientRef    *util.Ref[anilist.AnilistClient]
		AnilistPlatformRef  *util.Ref[platform.Platform]
		OfflinePlatformRef  *util.Ref[platform.Platform]
		MetadataProviderRef *util.Ref[metadata_provider.Provider]

		// Library
		FillerManager   *fillermanager.FillerManager
		AutoDownloader  *autodownloader.AutoDownloader
		AutoScanner     *autoscanner.AutoScanner
		PlaybackManager *playbackmanager.PlaybackManager

		// Real-time communication
		WSEventManager *events.WSEventManager

		// Extensions
		ExtensionRepository           *extension_repo.Repository
		ExtensionBankRef              *util.Ref[*extension.UnifiedBank]
		ExtensionPlaygroundRepository *extension_playground.PlaygroundRepository

		// Streaming
		DirectStreamManager     *directstream.Manager
		OnlinestreamRepository  *onlinestream.Repository
		MediastreamRepository   *mediastream.Repository
		TorrentstreamRepository *torrentstream.Repository

		// Players
		NativePlayer *nativeplayer.NativePlayer
		VideoCore    *videocore.VideoCore
		MediaPlayer  struct {
			VLC   *vlc.VLC
			MpcHc *mpchc.MpcHc
			Mpv   *mpv.Mpv
			Iina  *iina.Iina
		}
		MediaPlayerRepository *mediaplayer.Repository

		// Manga services
		MangaRepository *manga.Repository
		MangaDownloader *manga.Downloader

		// Offline and local account
		LocalManager local.Manager

		// Utilities
		FileCacher       *filecache.Cacher
		Updater          *updater.Updater
		SelfUpdater      *updater.SelfUpdater
		ReportRepository *report.Repository

		// Integrations
		DiscordPresence *discordrpc_presence.Presence

		// Continuity and sync
		ContinuityManager *continuity.Manager

		// Lifecycle management
		Cleanups                        []func()
		OnRefreshAnilistCollectionFuncs *result.Map[string, func()]
		OnFlushLogs                     func()

		// Configuration and feature flags
		FeatureFlags      FeatureFlags
		FeatureManager    *FeatureManager
		Settings          *models.Settings
		SecondarySettings struct {
			Mediastream   *models.MediastreamSettings
			Torrentstream *models.TorrentstreamSettings
			Debrid        *models.DebridSettings
		}

		// Metadata
		Version          string
		TotalLibrarySize uint64
		LibraryDir       string
		AnilistCacheDir  string
		IsDesktopSidecar bool
		Flags            SeanimeFlags

		// Internal state
		user               *user.User
		previousVersion    string
		moduleMu           sync.Mutex
		ServerReady        bool
		isOfflineRef       *util.Ref[bool]
		ServerPasswordHash string
		logoutInProgress   atomic.Bool

		// Plugin system
		HookManager hook.Manager

		// Features
		PlaylistManager *playlist.Manager
		LibraryExplorer *library_explorer.LibraryExplorer
		NakamaManager   *nakama.Manager

		// Show this version's tour on the frontend
		// Hydrated by migrations.go when there's a version change
		ShowTour string
	}
)

// NewApp creates a new server instance
func NewApp(configOpts *ConfigOptions, selfupdater *updater.SelfUpdater) *App {

	var app *App

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

	// Compute SHA-256 hash of the server password
	serverPasswordHash := ""
	if cfg.Server.Password != "" {
		serverPasswordHash = util.HashSHA256Hex(cfg.Server.Password)
	}

	// Create logs directory if it doesn't exist
	_ = os.MkdirAll(cfg.Logs.Dir, 0755)

	// Start background process to trim log files
	go TrimLogEntries(cfg.Logs.Dir, logger)

	logger.Info().Msgf("app: Data directory: %s", cfg.Data.AppDataDir)
	logger.Info().Msgf("app: Working directory: %s", cfg.Data.WorkingDir)

	// Log if running in desktop sidecar mode
	if configOpts.Flags.IsDesktopSidecar {
		logger.Info().Msg("app: Desktop sidecar mode enabled")
	}

	// Initialize database connection
	database, err := db.NewDatabase(cfg.Data.AppDataDir, cfg.Database.Name, logger)
	if err != nil {
		logger.Fatal().Err(err).Msgf("app: Failed to initialize database")
	}

	HandleNewDatabaseEntries(database, logger)

	// Clean up old database entries using the cleanup manager to prevent concurrent access issues
	database.RunDatabaseCleanup() // Remove old entries from all tables sequentially

	// Get anime library paths for plugin context
	animeLibraryPaths, _ := database.GetAllLibraryPathsFromSettings()
	plugin.GlobalAppContext.SetModulesPartial(plugin.AppContextModules{
		Database:          database,
		AnimeLibraryPaths: &animeLibraryPaths,
	})

	// Get Anilist token from database if available
	anilistToken := database.GetAnilistToken()

	anilistCacheDir := filepath.Join(cfg.Cache.Dir, "anilist")

	// Initialize Anilist API client with the token
	// If the token is empty, the client will not be authenticated
	anilistCW := anilist.NewAnilistClient(anilistToken, anilistCacheDir)
	anilistCWRef := util.NewRef[anilist.AnilistClient](anilistCW)

	// Initialize WebSocket event manager for real-time communication
	wsEventManager := events.NewWSEventManager(logger)

	// Exit if no WebSocket connections in desktop sidecar mode
	if configOpts.Flags.IsDesktopSidecar {
		wsEventManager.ExitIfNoConnsAsDesktopSidecar()
	}

	// Initialize DNS-over-HTTPS service in background
	go doh.HandleDoH(cfg.Server.DoHUrl, logger)

	// Initialize file cache system for media and metadata
	fileCacher, err := filecache.NewCacher(cfg.Cache.Dir)
	if err != nil {
		logger.Fatal().Err(err).Msgf("app: Failed to initialize file cacher")
	}

	// Initialize the extension bank that will be shared across modules
	extensionBankRef := util.NewRef(extension.NewUnifiedBank())

	// Initialize extension repository
	extensionRepository := extension_repo.NewRepository(&extension_repo.NewRepositoryOptions{
		Logger:           logger,
		ExtensionDir:     cfg.Extensions.Dir,
		WSEventManager:   wsEventManager,
		FileCacher:       fileCacher,
		HookManager:      hookManager,
		ExtensionBankRef: extensionBankRef,
	})

	// Initialize metadata provider for media information
	metadataProvider := metadata_provider.NewProvider(&metadata_provider.NewProviderImplOptions{
		Logger:           logger,
		FileCacher:       fileCacher,
		Database:         database,
		ExtensionBankRef: extensionBankRef,
	})

	// Set initial metadata provider (will change if offline mode is enabled)
	activeMetadataProvider := metadataProvider

	// Initialize manga repository
	mangaRepository := manga.NewRepository(&manga.NewRepositoryOptions{
		Logger:           logger,
		FileCacher:       fileCacher,
		CacheDir:         cfg.Cache.Dir,
		ServerURI:        cfg.GetServerURI(),
		WsEventManager:   wsEventManager,
		DownloadDir:      cfg.Manga.DownloadDir,
		Database:         database,
		ExtensionBankRef: extensionBankRef,
	})

	// Initialize Anilist platform
	anilistPlatform := anilist_platform.NewAnilistPlatform(anilistCWRef, extensionBankRef, logger, database, func() {
		if app != nil {
			app.LogoutFromAnilist()
		}
	})

	activePlatformRef := util.NewRef[platform.Platform](anilistPlatform)
	metadataProviderRef := util.NewRef[metadata_provider.Provider](activeMetadataProvider)

	// Initialize sync manager for offline/online synchronization
	localManager, err := local.NewManager(&local.NewManagerOptions{
		LocalDir:            cfg.Offline.Dir,
		AssetDir:            cfg.Offline.AssetDir,
		Logger:              logger,
		MetadataProviderRef: metadataProviderRef,
		MangaRepository:     mangaRepository,
		Database:            database,
		WSEventManager:      wsEventManager,
		IsOffline:           cfg.Server.Offline,
		AnilistPlatformRef:  activePlatformRef,
	})
	if err != nil {
		logger.Fatal().Err(err).Msgf("app: Failed to initialize sync manager")
	}

	// Use local metadata provider if in offline mode
	if cfg.Server.Offline {
		activeMetadataProvider = localManager.GetOfflineMetadataProvider()
	}

	// Initialize local platform for offline operations
	offlinePlatform, err := offline_platform.NewOfflinePlatform(localManager, anilistCWRef, logger)
	if err != nil {
		logger.Fatal().Err(err).Msgf("app: Failed to initialize local platform")
	}

	// Initialize simulated platform for unauthenticated operations
	simulatedPlatform, err := simulated_platform.NewSimulatedPlatform(localManager, anilistCWRef, extensionBankRef, logger, database)
	if err != nil {
		logger.Fatal().Err(err).Msgf("app: Failed to initialize simulated platform")
	}

	// Change active platform if offline mode is enabled
	if cfg.Server.Offline {
		logger.Warn().Msg("app: Offline mode is active, using offline platform")
		activePlatformRef.Set(offlinePlatform)
	} else if !anilistCWRef.Get().IsAuthenticated() {
		logger.Warn().Msg("app: Anilist client is not authenticated, using simulated platform")
		activePlatformRef.Set(simulatedPlatform)
	}

	isOfflineRef := util.NewRef(cfg.Server.Offline)
	offlinePlatformRef := util.NewRef[platform.Platform](offlinePlatform)

	// Update plugin context with new modules
	plugin.GlobalAppContext.SetModulesPartial(plugin.AppContextModules{
		IsOfflineRef:        isOfflineRef,
		AnilistPlatformRef:  activePlatformRef,
		WSEventManager:      wsEventManager,
		MetadataProviderRef: metadataProviderRef,
	})

	// Initialize online streaming repository
	onlinestreamRepository := onlinestream.NewRepository(&onlinestream.NewRepositoryOptions{
		Logger:              logger,
		FileCacher:          fileCacher,
		MetadataProviderRef: metadataProviderRef,
		PlatformRef:         activePlatformRef,
		Database:            database,
		ExtensionBankRef:    extensionBankRef,
	})

	// Initialize extension playground for testing extensions
	extensionPlaygroundRepository := extension_playground.NewPlaygroundRepository(logger, activePlatformRef, metadataProviderRef)

	// Create the main app instance with initialized components
	app = &App{
		Config:                        cfg,
		Flags:                         configOpts.Flags,
		FeatureManager:                NewFeatureManager(logger, configOpts.Flags),
		Database:                      database,
		AnilistClientRef:              anilistCWRef,
		AnilistPlatformRef:            activePlatformRef,
		OfflinePlatformRef:            offlinePlatformRef,
		LocalManager:                  localManager,
		WSEventManager:                wsEventManager,
		AnilistCacheDir:               anilistCacheDir,
		Logger:                        logger,
		Version:                       constants.Version,
		Updater:                       updater.New(constants.Version, logger, wsEventManager),
		FileCacher:                    fileCacher,
		OnlinestreamRepository:        onlinestreamRepository,
		MetadataProviderRef:           metadataProviderRef,
		MangaRepository:               mangaRepository,
		ExtensionRepository:           extensionRepository,
		ExtensionBankRef:              extensionBankRef,
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
		DirectStreamManager:           nil, // Initialized in App.initModulesOnce
		NativePlayer:                  nil, // Initialized in App.initModulesOnce
		VideoCore:                     nil, // Initialized in App.initModulesOnce
		NakamaManager:                 nil, // Initialized in App.initModulesOnce
		LibraryExplorer:               nil, // Initialized in App.initModulesOnce
		TorrentClientRepository:       nil, // Initialized in App.InitOrRefreshModules
		MediaPlayerRepository:         nil, // Initialized in App.InitOrRefreshModules
		DiscordPresence:               nil, // Initialized in App.InitOrRefreshModules
		previousVersion:               previousVersion,
		FeatureFlags:                  NewFeatureFlags(cfg, logger),
		IsDesktopSidecar:              configOpts.Flags.IsDesktopSidecar,
		SecondarySettings: struct {
			Mediastream   *models.MediastreamSettings
			Torrentstream *models.TorrentstreamSettings
			Debrid        *models.DebridSettings
		}{Mediastream: nil, Torrentstream: nil},
		SelfUpdater:                     selfupdater,
		moduleMu:                        sync.Mutex{},
		OnRefreshAnilistCollectionFuncs: result.NewMap[string, func()](),
		HookManager:                     hookManager,
		isOfflineRef:                    isOfflineRef,
		ServerPasswordHash:              serverPasswordHash,
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

	if !app.IsOffline() {
		go app.Updater.FetchAnnouncements()
	}

	// Initialize all modules that depend on settings
	app.InitOrRefreshModules()

	// Load custom source extensions before fetching AniList data
	LoadCustomSourceExtensions(extensionRepository)

	// Initialize Anilist data if not in offline mode
	if !app.IsOffline() {
		app.InitOrRefreshAnilistData()
	} else {
		app.ServerReady = true
	}

	// Load the other extensions asynchronously
	go LoadExtensions(extensionRepository, logger, cfg)

	// Initialize mediastream settings (for streaming media)
	app.InitOrRefreshMediastreamSettings()

	// Initialize torrentstream settings (for torrent streaming)
	app.InitOrRefreshTorrentstreamSettings()

	// Initialize debrid settings (for debrid services)
	app.InitOrRefreshDebridSettings()

	// Register Nakama manager cleanup
	app.AddCleanupFunction(app.NakamaManager.Cleanup)

	// Run one-time initialization actions
	app.performActionsOnce()

	return app
}

func (a *App) IsOffline() bool {
	return a.isOfflineRef.Get()
}

func (a *App) IsOfflineRef() *util.Ref[bool] {
	return a.isOfflineRef
}

func (a *App) AddCleanupFunction(f func()) {
	a.Cleanups = append(a.Cleanups, f)
}
func (a *App) AddOnRefreshAnilistCollectionFunc(key string, f func()) {
	if key == "" {
		return
	}
	a.OnRefreshAnilistCollectionFuncs.Set(key, f)
}

func (a *App) Cleanup() {
	for _, f := range a.Cleanups {
		f()
	}
}
