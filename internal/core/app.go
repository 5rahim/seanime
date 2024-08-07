package core

import (
	"github.com/rs/zerolog"
	"runtime"
	"seanime/internal/api/anilist"
	"seanime/internal/api/anizip"
	"seanime/internal/api/metadata"
	"seanime/internal/constants"
	"seanime/internal/database/db"
	"seanime/internal/database/db_bridge"
	"seanime/internal/database/models"
	"seanime/internal/discordrpc/presence"
	"seanime/internal/events"
	"seanime/internal/extension_repo"
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
	"seanime/internal/offline"
	"seanime/internal/onlinestream"
	"seanime/internal/platforms/anilist_platform"
	"seanime/internal/platforms/platform"
	"seanime/internal/torrent_clients/torrent_client"
	"seanime/internal/torrents/torrent"
	"seanime/internal/torrentstream"
	"seanime/internal/updater"
	"seanime/internal/util"
	"seanime/internal/util/filecache"
	"sync"
)

type (
	App struct {
		Config                  *Config
		Database                *db.Database
		Logger                  *zerolog.Logger
		TorrentClientRepository *torrent_client.Repository
		TorrentRepository       *torrent.Repository
		Watcher                 *scanner.Watcher
		AnizipCache             *anizip.Cache // AnizipCache holds fetched AniZip media for 30 minutes. (used by route handlers)
		AnilistClient           anilist.AnilistClient
		AnilistPlatform         platform.Platform
		FillerManager           *fillermanager.FillerManager
		WSEventManager          *events.WSEventManager
		AutoDownloader          *autodownloader.AutoDownloader
		ExtensionRepository     *extension_repo.Repository
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
		OnlinestreamRepository  *onlinestream.Repository
		MangaRepository         *manga.Repository
		MetadataProvider        *metadata.Provider
		DiscordPresence         *discordrpc_presence.Presence
		MangaDownloader         *manga.Downloader
		Cleanups                []func()
		OfflineHub              *offline.Hub
		MediastreamRepository   *mediastream.Repository
		TorrentstreamRepository *torrentstream.Repository
		FeatureFlags            FeatureFlags
		SecondarySettings       struct {
			Mediastream   *models.MediastreamSettings
			Torrentstream *models.TorrentstreamSettings
		} // Struct for other settings sent to client
		SelfUpdater        *updater.SelfUpdater
		TotalLibrarySize   uint64 // Initialized in modules.go
		LibraryDir         string
		animeCollection    *anilist.AnimeCollection // TODO: Rename to animeCollection
		rawAnimeCollection *anilist.AnimeCollection // (retains custom lists)
		mangaCollection    *anilist.MangaCollection
		rawMangaCollection *anilist.MangaCollection // (retains custom lists)
		account            *models.Account
		previousVersion    string
		moduleMu           sync.Mutex
	}
)

// NewApp creates a new server instance
func NewApp(configOpts *ConfigOptions, selfupdater *updater.SelfUpdater) *App {

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
	if _, _, err = db_bridge.GetLocalFiles(database); err != nil {
		_, err = db_bridge.InsertLocalFiles(database, make([]*anime.LocalFile, 0))
		if err != nil {
			logger.Fatal().Err(err).Msgf("app: Failed to initialize local files in the database")
		}
	}

	database.TrimLocalFileEntries()   // ran in goroutine
	database.TrimScanSummaryEntries() // ran in goroutine

	// Get token from stored account or return empty string
	anilistToken := database.GetAnilistToken()

	// Anilist Client Wrapper
	anilistCW := anilist.NewAnilistClient(anilistToken)

	// Websocket Event Manager
	wsEventManager := events.NewWSEventManager(logger)

	// Anilist Platform
	anilistPlatform := anilist_platform.NewAnilistPlatform(anilistCW, logger)

	// AniZip Cache
	anizipCache := anizip.NewCache()

	// File Cacher
	fileCacher, err := filecache.NewCacher(cfg.Cache.Dir)
	if err != nil {
		logger.Fatal().Err(err).Msgf("app: Failed to initialize file cacher")
	}

	// Online Stream
	onlineStream := onlinestream.NewRepository(&onlinestream.NewRepositoryOptions{
		Logger:      logger,
		FileCacher:  fileCacher,
		AnizipCache: anizipCache,
		Platform:    anilistPlatform,
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

	// Extension Repository
	extensionRepository := extension_repo.NewRepository(&extension_repo.NewRepositoryOptions{
		Logger:         logger,
		ExtensionDir:   cfg.Extensions.Dir,
		WSEventManager: wsEventManager,
	})

	app := &App{
		Config:                  cfg,
		Database:                database,
		AnilistClient:           anilistCW,
		AnilistPlatform:         anilistPlatform,
		AnizipCache:             anizipCache,
		WSEventManager:          wsEventManager,
		Logger:                  logger,
		Version:                 constants.Version,
		Updater:                 updater.New(constants.Version, logger),
		FileCacher:              fileCacher,
		OnlinestreamRepository:  onlineStream,
		MetadataProvider:        metadataProvider,
		MangaRepository:         mangaRepository,
		ExtensionRepository:     extensionRepository,
		TorrentRepository:       nil, // Initialized in App.initModulesOnce
		FillerManager:           nil, // Initialized in App.initModulesOnce
		MangaDownloader:         nil, // Initialized in App.initModulesOnce
		PlaybackManager:         nil, // Initialized in App.initModulesOnce
		AutoDownloader:          nil, // Initialized in App.initModulesOnce
		AutoScanner:             nil, // Initialized in App.initModulesOnce
		MediastreamRepository:   nil, // Initialized in App.initModulesOnce
		TorrentstreamRepository: nil, // Initialized in App.initModulesOnce
		OfflineHub:              nil, // Initialized in App.initModulesOnce
		TorrentClientRepository: nil, // Initialized in App.InitOrRefreshModules
		MediaPlayerRepository:   nil, // Initialized in App.InitOrRefreshModules
		DiscordPresence:         nil, // Initialized in App.InitOrRefreshModules
		previousVersion:         previousVersion,
		FeatureFlags:            NewFeatureFlags(cfg, logger),
		SecondarySettings: struct {
			Mediastream   *models.MediastreamSettings
			Torrentstream *models.TorrentstreamSettings
		}{Mediastream: nil, Torrentstream: nil},
		SelfUpdater: selfupdater,
		moduleMu:    sync.Mutex{},
	}

	// Perform necessary migrations if the version has changed
	app.runMigrations()

	// Initialize all modules that only need to be initialized once
	app.initModulesOnce()

	// Initialize all setting-dependent modules
	app.InitOrRefreshModules()

	// Load built-in extensions
	app.LoadBuiltInExtensions()
	// Load external extensions
	app.LoadOrRefreshExternalExtensions()

	// Fetch Anilist collection and set account if not offline
	if !app.IsOffline() {
		app.InitOrRefreshAnilistData()
	}

	// Initialize mediastream settings
	app.InitOrRefreshMediastreamSettings()

	// Initialize torrentstream settings
	app.InitOrRefreshTorrentstreamSettings()

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
