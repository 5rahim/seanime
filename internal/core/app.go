package core

import (
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
	"github.com/seanime-app/seanime/internal/library/fillermanager"
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
	"runtime"
	"sync"
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
		FillerManager           *fillermanager.FillerManager
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
		}
		SelfUpdater        *updater.SelfUpdater
		TotalLibrarySize   uint64                   // Initialized in modules.go
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
