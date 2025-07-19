package core

import (
	"runtime"
	"seanime/internal/api/anilist"
	"seanime/internal/continuity"
	"seanime/internal/database/db"
	"seanime/internal/database/db_bridge"
	"seanime/internal/database/models"
	debrid_client "seanime/internal/debrid/client"
	"seanime/internal/directstream"
	discordrpc_presence "seanime/internal/discordrpc/presence"
	"seanime/internal/events"
	"seanime/internal/library/anime"
	"seanime/internal/library/autodownloader"
	"seanime/internal/library/autoscanner"
	"seanime/internal/library/fillermanager"
	"seanime/internal/library/playbackmanager"
	"seanime/internal/manga"
	"seanime/internal/mediaplayers/iina"
	"seanime/internal/mediaplayers/mediaplayer"
	"seanime/internal/mediaplayers/mpchc"
	"seanime/internal/mediaplayers/mpv"
	"seanime/internal/mediaplayers/vlc"
	"seanime/internal/mediastream"
	"seanime/internal/nakama"
	"seanime/internal/nativeplayer"
	"seanime/internal/notifier"
	"seanime/internal/plugin"
	"seanime/internal/torrent_clients/qbittorrent"
	"seanime/internal/torrent_clients/torrent_client"
	"seanime/internal/torrent_clients/transmission"
	"seanime/internal/torrents/torrent"
	"seanime/internal/torrentstream"
	"seanime/internal/user"

	"github.com/cli/browser"
	"github.com/rs/zerolog"
)

// initModulesOnce will initialize modules that need to persist.
// This function is called once after the App instance is created.
// The settings of these modules will be set/refreshed in InitOrRefreshModules.
func (a *App) initModulesOnce() {

	a.LocalManager.SetRefreshAnilistCollectionsFunc(func() {
		_, _ = a.RefreshAnimeCollection()
		_, _ = a.RefreshMangaCollection()
	})

	plugin.GlobalAppContext.SetModulesPartial(plugin.AppContextModules{
		OnRefreshAnilistAnimeCollection: func() {
			_, _ = a.RefreshAnimeCollection()
		},
		OnRefreshAnilistMangaCollection: func() {
			_, _ = a.RefreshMangaCollection()
		},
	})

	// +---------------------+
	// |     Discord RPC     |
	// +---------------------+

	a.DiscordPresence = discordrpc_presence.New(nil, a.Logger)
	a.AddCleanupFunction(func() {
		a.DiscordPresence.Close()
	})

	plugin.GlobalAppContext.SetModulesPartial(plugin.AppContextModules{
		DiscordPresence: a.DiscordPresence,
	})

	// +---------------------+
	// |       Filler        |
	// +---------------------+

	a.FillerManager = fillermanager.New(&fillermanager.NewFillerManagerOptions{
		DB:     a.Database,
		Logger: a.Logger,
	})

	plugin.GlobalAppContext.SetModulesPartial(plugin.AppContextModules{
		FillerManager: a.FillerManager,
	})

	// +---------------------+
	// |     Continuity      |
	// +---------------------+

	a.ContinuityManager = continuity.NewManager(&continuity.NewManagerOptions{
		FileCacher: a.FileCacher,
		Logger:     a.Logger,
		Database:   a.Database,
	})

	// +---------------------+
	// |   Playback Manager  |
	// +---------------------+

	// Playback Manager
	a.PlaybackManager = playbackmanager.New(&playbackmanager.NewPlaybackManagerOptions{
		Logger:            a.Logger,
		WSEventManager:    a.WSEventManager,
		Platform:          a.AnilistPlatform,
		MetadataProvider:  a.MetadataProvider,
		Database:          a.Database,
		DiscordPresence:   a.DiscordPresence,
		IsOffline:         a.IsOffline(),
		ContinuityManager: a.ContinuityManager,
		RefreshAnimeCollectionFunc: func() {
			_, _ = a.RefreshAnimeCollection()
		},
	})

	// +---------------------+
	// |  Torrent Repository |
	// +---------------------+

	a.TorrentRepository = torrent.NewRepository(&torrent.NewRepositoryOptions{
		Logger:           a.Logger,
		MetadataProvider: a.MetadataProvider,
	})

	// +---------------------+
	// |  Debrid Client Repo |
	// +---------------------+

	a.DebridClientRepository = debrid_client.NewRepository(&debrid_client.NewRepositoryOptions{
		Logger:            a.Logger,
		WSEventManager:    a.WSEventManager,
		Database:          a.Database,
		MetadataProvider:  a.MetadataProvider,
		Platform:          a.AnilistPlatform,
		PlaybackManager:   a.PlaybackManager,
		TorrentRepository: a.TorrentRepository,
	})

	// +---------------------+
	// |   Auto Downloader   |
	// +---------------------+

	a.AutoDownloader = autodownloader.New(&autodownloader.NewAutoDownloaderOptions{
		Logger:                  a.Logger,
		TorrentClientRepository: a.TorrentClientRepository,
		TorrentRepository:       a.TorrentRepository,
		Database:                a.Database,
		WSEventManager:          a.WSEventManager,
		MetadataProvider:        a.MetadataProvider,
		DebridClientRepository:  a.DebridClientRepository,
		IsOffline:               a.IsOffline(),
	})

	// This is run in a goroutine
	a.AutoDownloader.Start()

	// +---------------------+
	// |   Auto Scanner      |
	// +---------------------+

	a.AutoScanner = autoscanner.New(&autoscanner.NewAutoScannerOptions{
		Database:         a.Database,
		Platform:         a.AnilistPlatform,
		Logger:           a.Logger,
		WSEventManager:   a.WSEventManager,
		Enabled:          false, // Will be set in InitOrRefreshModules
		AutoDownloader:   a.AutoDownloader,
		MetadataProvider: a.MetadataProvider,
		LogsDir:          a.Config.Logs.Dir,
	})

	// This is run in a goroutine
	a.AutoScanner.Start()

	// +---------------------+
	// |  Manga Downloader   |
	// +---------------------+

	a.MangaDownloader = manga.NewDownloader(&manga.NewDownloaderOptions{
		Database:       a.Database,
		Logger:         a.Logger,
		WSEventManager: a.WSEventManager,
		DownloadDir:    a.Config.Manga.DownloadDir,
		Repository:     a.MangaRepository,
		IsOffline:      a.IsOffline(),
	})

	a.MangaDownloader.Start()

	// +---------------------+
	// |    Media Stream     |
	// +---------------------+

	a.MediastreamRepository = mediastream.NewRepository(&mediastream.NewRepositoryOptions{
		Logger:         a.Logger,
		WSEventManager: a.WSEventManager,
		FileCacher:     a.FileCacher,
	})

	a.AddCleanupFunction(func() {
		a.MediastreamRepository.OnCleanup()
	})

	// +---------------------+
	// |    Native Player    |
	// +---------------------+

	a.NativePlayer = nativeplayer.New(nativeplayer.NewNativePlayerOptions{
		WsEventManager: a.WSEventManager,
		Logger:         a.Logger,
	})

	// +---------------------+
	// |   Direct Stream     |
	// +---------------------+

	a.DirectStreamManager = directstream.NewManager(directstream.NewManagerOptions{
		Logger:            a.Logger,
		WSEventManager:    a.WSEventManager,
		ContinuityManager: a.ContinuityManager,
		MetadataProvider:  a.MetadataProvider,
		DiscordPresence:   a.DiscordPresence,
		Platform:          a.AnilistPlatform,
		RefreshAnimeCollectionFunc: func() {
			_, _ = a.RefreshAnimeCollection()
		},
		IsOffline:    a.IsOffline(),
		NativePlayer: a.NativePlayer,
	})

	// +---------------------+
	// |   Torrent Stream    |
	// +---------------------+

	a.TorrentstreamRepository = torrentstream.NewRepository(&torrentstream.NewRepositoryOptions{
		Logger:              a.Logger,
		BaseAnimeCache:      anilist.NewBaseAnimeCache(),
		CompleteAnimeCache:  anilist.NewCompleteAnimeCache(),
		MetadataProvider:    a.MetadataProvider,
		TorrentRepository:   a.TorrentRepository,
		Platform:            a.AnilistPlatform,
		PlaybackManager:     a.PlaybackManager,
		WSEventManager:      a.WSEventManager,
		Database:            a.Database,
		DirectStreamManager: a.DirectStreamManager,
		NativePlayer:        a.NativePlayer,
	})

	plugin.GlobalAppContext.SetModulesPartial(plugin.AppContextModules{
		MediaPlayerRepository: a.MediaPlayerRepository,
		PlaybackManager:       a.PlaybackManager,
		MangaRepository:       a.MangaRepository,
	})

	// +---------------------+
	// |       Nakama        |
	// +---------------------+

	a.NakamaManager = nakama.NewManager(&nakama.NewManagerOptions{
		Logger:                  a.Logger,
		WSEventManager:          a.WSEventManager,
		PlaybackManager:         a.PlaybackManager,
		TorrentstreamRepository: a.TorrentstreamRepository,
		DebridClientRepository:  a.DebridClientRepository,
		Platform:                a.AnilistPlatform,
		ServerHost:              a.Config.Server.Host,
		ServerPort:              a.Config.Server.Port,
	})

}

// HandleNewDatabaseEntries initializes essential database collections.
// It creates an empty local files collection if one does not already exist.
func HandleNewDatabaseEntries(database *db.Database, logger *zerolog.Logger) {

	// Create initial empty local files collection if none exists
	if _, _, err := db_bridge.GetLocalFiles(database); err != nil {
		_, err := db_bridge.InsertLocalFiles(database, make([]*anime.LocalFile, 0))
		if err != nil {
			logger.Fatal().Err(err).Msgf("app: Failed to initialize local files in the database")
		}
	}

}

// InitOrRefreshModules will initialize or refresh modules that depend on settings.
// This function is called:
//   - After the App instance is created
//   - After settings are updated.
func (a *App) InitOrRefreshModules() {
	a.moduleMu.Lock()
	defer a.moduleMu.Unlock()

	a.Logger.Debug().Msgf("app: Refreshing modules")

	// Stop watching if already watching
	if a.Watcher != nil {
		a.Watcher.StopWatching()
	}

	// If Discord presence is already initialized, close it
	if a.DiscordPresence != nil {
		a.DiscordPresence.Close()
	}

	// Get settings from database
	settings, err := a.Database.GetSettings()
	if err != nil || settings == nil {
		a.Logger.Warn().Msg("app: Did not initialize modules, no settings found")
		return
	}

	a.Settings = settings // Store settings instance in app
	if settings.Library != nil {
		a.LibraryDir = settings.GetLibrary().LibraryPath
	}

	// +---------------------+
	// |   Module settings   |
	// +---------------------+
	// Refresh settings of modules that were initialized in initModulesOnce

	notifier.GlobalNotifier.SetSettings(a.Config.Data.AppDataDir, a.Settings.GetNotifications(), a.Logger)

	// Refresh updater settings
	if settings.Library != nil {
		plugin.GlobalAppContext.SetModulesPartial(plugin.AppContextModules{
			AnimeLibraryPaths: a.Database.AllLibraryPathsFromSettings(settings),
		})

		if a.Updater != nil {
			a.Updater.SetEnabled(!settings.Library.DisableUpdateCheck)
		}

		// Refresh auto scanner settings
		if a.AutoScanner != nil {
			a.AutoScanner.SetSettings(*settings.Library)
		}

		// Torrent Repository
		a.TorrentRepository.SetSettings(&torrent.RepositorySettings{
			DefaultAnimeProvider: settings.Library.TorrentProvider,
		})
	}

	if settings.MediaPlayer != nil {
		a.MediaPlayer.VLC = &vlc.VLC{
			Host:     settings.MediaPlayer.Host,
			Port:     settings.MediaPlayer.VlcPort,
			Password: settings.MediaPlayer.VlcPassword,
			Path:     settings.MediaPlayer.VlcPath,
			Logger:   a.Logger,
		}
		a.MediaPlayer.MpcHc = &mpchc.MpcHc{
			Host:   settings.MediaPlayer.Host,
			Port:   settings.MediaPlayer.MpcPort,
			Path:   settings.MediaPlayer.MpcPath,
			Logger: a.Logger,
		}
		a.MediaPlayer.Mpv = mpv.New(a.Logger, settings.MediaPlayer.MpvSocket, settings.MediaPlayer.MpvPath, settings.MediaPlayer.MpvArgs)
		a.MediaPlayer.Iina = iina.New(a.Logger, settings.MediaPlayer.IinaSocket, settings.MediaPlayer.IinaPath, settings.MediaPlayer.IinaArgs)

		// Set media player repository
		a.MediaPlayerRepository = mediaplayer.NewRepository(&mediaplayer.NewRepositoryOptions{
			Logger:            a.Logger,
			Default:           settings.MediaPlayer.Default,
			VLC:               a.MediaPlayer.VLC,
			MpcHc:             a.MediaPlayer.MpcHc,
			Mpv:               a.MediaPlayer.Mpv, // Socket
			Iina:              a.MediaPlayer.Iina,
			WSEventManager:    a.WSEventManager,
			ContinuityManager: a.ContinuityManager,
		})

		a.PlaybackManager.SetMediaPlayerRepository(a.MediaPlayerRepository)
		a.PlaybackManager.SetSettings(&playbackmanager.Settings{
			AutoPlayNextEpisode: a.Settings.GetLibrary().AutoPlayNextEpisode,
		})

		a.DirectStreamManager.SetSettings(&directstream.Settings{
			AutoPlayNextEpisode: a.Settings.GetLibrary().AutoPlayNextEpisode,
			AutoUpdateProgress:  a.Settings.GetLibrary().AutoUpdateProgress,
		})

		a.TorrentstreamRepository.SetMediaPlayerRepository(a.MediaPlayerRepository)
	} else {
		a.Logger.Warn().Msg("app: Did not initialize media player module, no settings found")
	}

	// +---------------------+
	// |       Torrents      |
	// +---------------------+

	if settings.Torrent != nil {
		// Init qBittorrent
		qbit := qbittorrent.NewClient(&qbittorrent.NewClientOptions{
			Logger:   a.Logger,
			Username: settings.Torrent.QBittorrentUsername,
			Password: settings.Torrent.QBittorrentPassword,
			Port:     settings.Torrent.QBittorrentPort,
			Host:     settings.Torrent.QBittorrentHost,
			Path:     settings.Torrent.QBittorrentPath,
			Tags:     settings.Torrent.QBittorrentTags,
		})
		// Login to qBittorrent
		go func() {
			if settings.Torrent.Default == "qbittorrent" {
				err = qbit.Login()
				if err != nil {
					a.Logger.Error().Err(err).Msg("app: Failed to login to qBittorrent")
				} else {
					a.Logger.Info().Msg("app: Logged in to qBittorrent")
				}
			}
		}()
		// Init Transmission
		trans, err := transmission.New(&transmission.NewTransmissionOptions{
			Logger:   a.Logger,
			Username: settings.Torrent.TransmissionUsername,
			Password: settings.Torrent.TransmissionPassword,
			Port:     settings.Torrent.TransmissionPort,
			Host:     settings.Torrent.TransmissionHost,
			Path:     settings.Torrent.TransmissionPath,
		})
		if err != nil && settings.Torrent.TransmissionUsername != "" && settings.Torrent.TransmissionPassword != "" { // Only log error if username and password are set
			a.Logger.Error().Err(err).Msg("app: Failed to initialize transmission client")
		}

		// Shutdown torrent client first
		if a.TorrentClientRepository != nil {
			a.TorrentClientRepository.Shutdown()
		}

		// Torrent Client Repository
		a.TorrentClientRepository = torrent_client.NewRepository(&torrent_client.NewRepositoryOptions{
			Logger:            a.Logger,
			QbittorrentClient: qbit,
			Transmission:      trans,
			TorrentRepository: a.TorrentRepository,
			Provider:          settings.Torrent.Default,
			MetadataProvider:  a.MetadataProvider,
		})

		a.TorrentClientRepository.InitActiveTorrentCount(settings.Torrent.ShowActiveTorrentCount, a.WSEventManager)

		// Set AutoDownloader qBittorrent client
		a.AutoDownloader.SetTorrentClientRepository(a.TorrentClientRepository)

		plugin.GlobalAppContext.SetModulesPartial(plugin.AppContextModules{
			TorrentClientRepository: a.TorrentClientRepository,
			AutoDownloader:          a.AutoDownloader,
		})
	} else {
		a.Logger.Warn().Msg("app: Did not initialize torrent client module, no settings found")
	}

	// +---------------------+
	// |   AutoDownloader    |
	// +---------------------+

	// Update Auto Downloader - This runs in a goroutine
	if settings.AutoDownloader != nil {
		a.AutoDownloader.SetSettings(settings.AutoDownloader, settings.Library.TorrentProvider)
	}

	// +---------------------+
	// |   Library Watcher   |
	// +---------------------+

	// Initialize library watcher
	if settings.Library != nil && len(settings.Library.LibraryPath) > 0 {
		go func() {
			a.initLibraryWatcher(settings.Library.GetLibraryPaths())
		}()
	}

	// +---------------------+
	// |       Discord       |
	// +---------------------+

	if settings.Discord != nil && a.DiscordPresence != nil {
		a.DiscordPresence.SetSettings(settings.Discord)
	}

	// +---------------------+
	// |     Continuity      |
	// +---------------------+

	if settings.Library != nil {
		a.ContinuityManager.SetSettings(&continuity.Settings{
			WatchContinuityEnabled: settings.Library.EnableWatchContinuity,
		})
	}

	if settings.Manga != nil {
		a.MangaRepository.SetSettings(settings)
	}

	// +---------------------+
	// |       Nakama        |
	// +---------------------+

	if settings.Nakama != nil {
		a.NakamaManager.SetSettings(settings.Nakama)
	}

	runtime.GC()

	a.Logger.Info().Msg("app: Refreshed modules")

}

// InitOrRefreshMediastreamSettings will initialize or refresh the mediastream settings.
// It is called after the App instance is created and after settings are updated.
func (a *App) InitOrRefreshMediastreamSettings() {

	var settings *models.MediastreamSettings
	var found bool
	settings, found = a.Database.GetMediastreamSettings()
	if !found {

		var err error
		settings, err = a.Database.UpsertMediastreamSettings(&models.MediastreamSettings{
			BaseModel: models.BaseModel{
				ID: 1,
			},
			TranscodeEnabled:    false,
			TranscodeHwAccel:    "cpu",
			TranscodePreset:     "fast",
			PreTranscodeEnabled: false,
		})
		if err != nil {
			a.Logger.Error().Err(err).Msg("app: Failed to initialize mediastream module")
			return
		}
	}

	a.MediastreamRepository.InitializeModules(settings, a.Config.Cache.Dir, a.Config.Cache.TranscodeDir)

	// Cleanup cache
	go func() {
		if settings.TranscodeEnabled {
			// If transcoding is enabled, trim files
			_ = a.FileCacher.TrimMediastreamVideoFiles()
		} else {
			// If transcoding is disabled, clear all files
			_ = a.FileCacher.ClearMediastreamVideoFiles()
		}
	}()

	a.SecondarySettings.Mediastream = settings
}

// InitOrRefreshTorrentstreamSettings will initialize or refresh the mediastream settings.
// It is called after the App instance is created and after settings are updated.
func (a *App) InitOrRefreshTorrentstreamSettings() {

	var settings *models.TorrentstreamSettings
	var found bool
	settings, found = a.Database.GetTorrentstreamSettings()
	if !found {

		var err error
		settings, err = a.Database.UpsertTorrentstreamSettings(&models.TorrentstreamSettings{
			BaseModel: models.BaseModel{
				ID: 1,
			},
			Enabled:             false,
			AutoSelect:          true,
			PreferredResolution: "",
			DisableIPV6:         false,
			DownloadDir:         "",
			AddToLibrary:        false,
			TorrentClientHost:   "",
			TorrentClientPort:   43213,
			StreamingServerHost: "0.0.0.0",
			StreamingServerPort: 43214,
			IncludeInLibrary:    false,
			StreamUrlAddress:    "",
			SlowSeeding:         false,
		})
		if err != nil {
			a.Logger.Error().Err(err).Msg("app: Failed to initialize mediastream module")
			return
		}
	}

	err := a.TorrentstreamRepository.InitModules(settings, a.Config.Server.Host, a.Config.Server.Port)
	if err != nil && settings.Enabled {
		a.Logger.Error().Err(err).Msg("app: Failed to initialize Torrent streaming module")
		//_, _ = a.Database.UpsertTorrentstreamSettings(&models.TorrentstreamSettings{
		//	BaseModel: models.BaseModel{
		//		ID: 1,
		//	},
		//	Enabled: false,
		//})
	}

	a.Cleanups = append(a.Cleanups, func() {
		a.TorrentstreamRepository.Shutdown()
	})

	// Set torrent streaming settings in secondary settings
	// so the client can use them
	a.SecondarySettings.Torrentstream = settings
}

func (a *App) InitOrRefreshDebridSettings() {

	settings, found := a.Database.GetDebridSettings()
	if !found {

		var err error
		settings, err = a.Database.UpsertDebridSettings(&models.DebridSettings{
			BaseModel: models.BaseModel{
				ID: 1,
			},
			Enabled:                      false,
			Provider:                     "",
			ApiKey:                       "",
			IncludeDebridStreamInLibrary: false,
			StreamAutoSelect:             false,
			StreamPreferredResolution:    "",
		})
		if err != nil {
			a.Logger.Error().Err(err).Msg("app: Failed to initialize debrid module")
			return
		}
	}

	a.SecondarySettings.Debrid = settings

	err := a.DebridClientRepository.InitializeProvider(settings)
	if err != nil {
		a.Logger.Error().Err(err).Msg("app: Failed to initialize debrid provider")
		return
	}
}

// InitOrRefreshAnilistData will initialize the Anilist anime collection and the account.
// This function should be called after App.Database is initialized and after settings are updated.
func (a *App) InitOrRefreshAnilistData() {
	a.Logger.Debug().Msg("app: Fetching Anilist data")

	var currUser *user.User
	acc, err := a.Database.GetAccount()
	if err != nil || acc.Username == "" {
		a.ServerReady = true
		currUser = user.NewSimulatedUser() // Create a simulated user if no account is found
	} else {
		currUser, err = user.NewUser(acc)
		if err != nil {
			a.Logger.Error().Err(err).Msg("app: Failed to create user from account")
			return
		}
	}

	a.user = currUser

	// Set username to Anilist platform
	a.AnilistPlatform.SetUsername(currUser.Viewer.Name)

	a.Logger.Info().Msg("app: Authenticated to AniList")

	go func() {
		_, err = a.RefreshAnimeCollection()
		if err != nil {
			a.Logger.Error().Err(err).Msg("app: Failed to fetch Anilist anime collection")
		}

		a.ServerReady = true
		a.WSEventManager.SendEvent(events.ServerReady, nil)

		_, err = a.RefreshMangaCollection()
		if err != nil {
			a.Logger.Error().Err(err).Msg("app: Failed to fetch Anilist manga collection")
		}
	}()

	go func(username string) {
		a.DiscordPresence.SetUsername(username)
	}(currUser.Viewer.Name)

	a.Logger.Info().Msg("app: Fetched Anilist data")
}

func (a *App) performActionsOnce() {

	go func() {
		if a.Settings == nil || a.Settings.Library == nil {
			return
		}

		if a.Settings.GetLibrary().OpenWebURLOnStart {
			// Open the web URL
			err := browser.OpenURL(a.Config.GetServerURI("127.0.0.1"))
			if err != nil {
				a.Logger.Warn().Err(err).Msg("app: Failed to open web URL, please open it manually in your browser")
			} else {
				a.Logger.Info().Msg("app: Opened web URL")
			}
		}

		if a.Settings.GetLibrary().RefreshLibraryOnStart {
			go func() {
				a.Logger.Debug().Msg("app: Refreshing library")
				a.AutoScanner.RunNow()
				a.Logger.Info().Msg("app: Refreshed library")
			}()
		}

		if a.Settings.GetLibrary().OpenTorrentClientOnStart && a.TorrentClientRepository != nil {
			// Start the torrent client
			ok := a.TorrentClientRepository.Start()
			if !ok {
				a.Logger.Warn().Msg("app: Failed to open torrent client")
			} else {
				a.Logger.Info().Msg("app: Started torrent client")
			}

		}
	}()

}
