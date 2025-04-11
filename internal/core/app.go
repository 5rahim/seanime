package core

import (
	"os"
	"runtime"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/constants"
	"seanime/internal/continuity"
	"seanime/internal/database/db"
	"seanime/internal/database/db_bridge"
	"seanime/internal/database/models"
	debrid_client "seanime/internal/debrid/client"
	discordrpc_presence "seanime/internal/discordrpc/presence"
	"seanime/internal/doh"
	"seanime/internal/events"
	"seanime/internal/extension_playground"
	"seanime/internal/extension_repo"
	"seanime/internal/hook"
	"seanime/internal/library/anime"
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
		Settings                *models.Settings
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
		AnilistDataLoaded  bool // Whether the Anilist data from the first request has been fetched
	}
)

// NewApp creates a new server instance
func NewApp(configOpts *ConfigOptions, selfupdater *updater.SelfUpdater) *App {

	logger := util.NewLogger()

	logger.Info().Msgf("app: Seanime %s-%s", constants.Version, constants.VersionName)
	logger.Info().Msgf("app: OS: %s", runtime.GOOS)
	logger.Info().Msgf("app: Arch: %s", runtime.GOARCH)
	logger.Info().Msgf("app: Processor count: %d", runtime.NumCPU())

	hookManager := hook.NewHookManager(hook.NewHookManagerOptions{Logger: logger})
	hook.SetGlobalHookManager(hookManager)
	plugin.GlobalAppContext.SetLogger(logger)

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

	_ = os.MkdirAll(cfg.Logs.Dir, 0755)

	go TrimLogEntries(cfg.Logs.Dir, logger)

	logger.Info().Msgf("app: Data directory: %s", cfg.Data.AppDataDir)

	// Print working directory
	logger.Info().Msgf("app: Working directory: %s", cfg.Data.WorkingDir)

	if configOpts.IsDesktopSidecar {
		logger.Info().Msg("app: Desktop sidecar mode enabled")
	}

	// Initialize the database
	database, err := db.NewDatabase(cfg.Data.AppDataDir, cfg.Database.Name, logger)
	if err != nil {
		logger.Fatal().Err(err).Msgf("app: Failed to initialize database")
	}

	// Add default local file entries if there are none
	if _, _, err = db_bridge.GetLocalFiles(database); err != nil {
		_, err = db_bridge.InsertLocalFiles(database, make([]*anime.LocalFile, 0))
		if err != nil {
			logger.Fatal().Err(err).Msgf("app: Failed to initialize local files in the database")
		}
	}

	database.TrimLocalFileEntries()     // ran in goroutine
	database.TrimScanSummaryEntries()   // ran in goroutine
	database.TrimTorrentstreamHistory() // ran in goroutine

	animeLibraryPaths, _ := database.GetAllLibraryPathsFromSettings()
	plugin.GlobalAppContext.SetModulesPartial(plugin.AppContextModules{
		Database:          database,
		AnimeLibraryPaths: animeLibraryPaths,
	})

	// Get token from stored account or return empty string
	anilistToken := database.GetAnilistToken()

	// Anilist Client Wrapper
	anilistCW := anilist.NewAnilistClient(anilistToken)

	// Websocket Event Manager
	wsEventManager := events.NewWSEventManager(logger)

	if configOpts.IsDesktopSidecar {
		wsEventManager.ExitIfNoConnsAsDesktopSidecar()
	}

	// DoH
	go doh.HandleDoH(cfg.Server.DoHUrl, logger)

	// File Cacher
	fileCacher, err := filecache.NewCacher(cfg.Cache.Dir)
	if err != nil {
		logger.Fatal().Err(err).Msgf("app: Failed to initialize file cacher")
	}

	// Extension Repository
	extensionRepository := extension_repo.NewRepository(&extension_repo.NewRepositoryOptions{
		Logger:         logger,
		ExtensionDir:   cfg.Extensions.Dir,
		WSEventManager: wsEventManager,
		FileCacher:     fileCacher,
		HookManager:    hookManager,
	})
	go LoadExtensions(extensionRepository, logger)

	// Metadata Provider
	metadataProvider := metadata.NewProvider(&metadata.NewProviderImplOptions{
		Logger:     logger,
		FileCacher: fileCacher,
	})

	activeMetadataProvider := metadataProvider

	// Manga Repository
	mangaRepository := manga.NewRepository(&manga.NewRepositoryOptions{
		Logger:         logger,
		FileCacher:     fileCacher,
		CacheDir:       cfg.Cache.Dir,
		ServerURI:      cfg.GetServerURI(),
		WsEventManager: wsEventManager,
		DownloadDir:    cfg.Manga.DownloadDir,
		Database:       database,
	})

	anilistPlatform := anilist_platform.NewAnilistPlatform(anilistCW, logger)

	plugin.GlobalAppContext.SetModulesPartial(plugin.AppContextModules{
		AnilistPlatform: anilistPlatform,
		WSEventManager:  wsEventManager,
	})

	// Platforms
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

	if cfg.Server.Offline {
		activeMetadataProvider = syncManager.GetLocalMetadataProvider()
	}

	localPlatform, err := local_platform.NewLocalPlatform(syncManager, anilistCW, logger)
	if err != nil {
		logger.Fatal().Err(err).Msgf("app: Failed to initialize local platform")
	}

	activePlatform := anilistPlatform
	// If offline mode is enabled, use the local platform
	if cfg.Server.Offline {
		activePlatform = localPlatform
	}

	// Online Stream
	onlinestreamRepository := onlinestream.NewRepository(&onlinestream.NewRepositoryOptions{
		Logger:           logger,
		FileCacher:       fileCacher,
		MetadataProvider: activeMetadataProvider,
		Platform:         activePlatform,
		Database:         database,
	})

	extensionPlaygroundRepository := extension_playground.NewPlaygroundRepository(logger, activePlatform, activeMetadataProvider)

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

	// Perform necessary migrations if the version has changed
	app.runMigrations()

	// Initialize all modules that only need to be initialized once
	app.initModulesOnce()

	// Initialize all setting-dependent modules
	app.InitOrRefreshModules()

	// Load built-in extensions
	app.AddExtensionBankToConsumers()

	// Fetch Anilist collection and set account if not offline
	if !app.IsOffline() {
		app.InitOrRefreshAnilistData()
	}

	// Initialize mediastream settings
	app.InitOrRefreshMediastreamSettings()

	// Initialize torrentstream settings
	app.InitOrRefreshTorrentstreamSettings()

	// Initialize debrid settings
	app.InitOrRefreshDebridSettings()

	// Perform actions that need to be done after the app has been initialized
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
