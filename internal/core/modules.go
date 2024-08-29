package core

import (
	"github.com/cli/browser"
	"runtime"
	"seanime/internal/api/anilist"
	"seanime/internal/database/models"
	"seanime/internal/discordrpc/presence"
	"seanime/internal/library/autodownloader"
	"seanime/internal/library/autoscanner"
	"seanime/internal/library/fillermanager"
	"seanime/internal/library/playbackmanager"
	"seanime/internal/manga"
	"seanime/internal/mediaplayers/mediaplayer"
	"seanime/internal/mediaplayers/mpchc"
	"seanime/internal/mediaplayers/mpv"
	"seanime/internal/mediaplayers/vlc"
	"seanime/internal/mediastream"
	"seanime/internal/notifier"
	"seanime/internal/offline"
	"seanime/internal/torrent_clients/qbittorrent"
	"seanime/internal/torrent_clients/torrent_client"
	"seanime/internal/torrent_clients/transmission"
	"seanime/internal/torrents/torrent"
	"seanime/internal/torrentstream"
)

// initModulesOnce will initialize modules that need to persist.
// This function is called once after the App instance is created.
// The settings of these modules will be set/refreshed in InitOrRefreshModules.
func (a *App) initModulesOnce() {

	// +---------------------+
	// |     Offline Hub     |
	// +---------------------+

	// Will exit if offline mode is enabled and no snapshots are found
	a.OfflineHub = offline.NewHub(&offline.NewHubOptions{
		Platform:         a.AnilistPlatform,
		MetadataProvider: a.MetadataProvider,
		MangaRepository:  a.MangaRepository,
		WSEventManager:   a.WSEventManager,
		Database:         a.Database,
		FileCacher:       a.FileCacher,
		Logger:           a.Logger,
		OfflineDir:       a.Config.Offline.Dir,
		AssetDir:         a.Config.Offline.AssetDir,
		IsOffline:        a.Config.Server.Offline,
		RefreshAnimeCollectionsFunc: func() {
			_, _ = a.RefreshAnimeCollection()
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

	// +---------------------+
	// |       Filler        |
	// +---------------------+

	a.FillerManager = fillermanager.New(&fillermanager.NewFillerManagerOptions{
		DB:     a.Database,
		Logger: a.Logger,
	})

	// +---------------------+
	// |   Playback Manager  |
	// +---------------------+

	// Playback Manager
	a.PlaybackManager = playbackmanager.New(&playbackmanager.NewPlaybackManagerOptions{
		Logger:          a.Logger,
		WSEventManager:  a.WSEventManager,
		Platform:        a.AnilistPlatform,
		Database:        a.Database,
		DiscordPresence: a.DiscordPresence,
		IsOffline:       a.IsOffline(),
		OfflineHub:      a.OfflineHub,
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
	// |   Auto Downloader   |
	// +---------------------+

	a.AutoDownloader = autodownloader.New(&autodownloader.NewAutoDownloaderOptions{
		Logger:                  a.Logger,
		TorrentClientRepository: a.TorrentClientRepository,
		TorrentRepository:       a.TorrentRepository,
		Database:                a.Database,
		WSEventManager:          a.WSEventManager,
		AnizipCache:             a.AnizipCache,
	})

	if !a.IsOffline() {
		// This is run in a goroutine
		a.AutoDownloader.Start()
	}

	// +---------------------+
	// |   Auto Scanner      |
	// +---------------------+

	a.AutoScanner = autoscanner.New(&autoscanner.NewAutoScannerOptions{
		Database:       a.Database,
		Enabled:        false, // Will be set in InitOrRefreshModules
		AutoDownloader: a.AutoDownloader,
		Platform:       a.AnilistPlatform,
		Logger:         a.Logger,
		WSEventManager: a.WSEventManager,
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
	})

	if !a.IsOffline() {
		// This is run in a goroutine
		a.MangaDownloader.Start()
	}

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
	// |   Torrent Stream    |
	// +---------------------+

	a.TorrentstreamRepository = torrentstream.NewRepository(&torrentstream.NewRepositoryOptions{
		Logger:             a.Logger,
		AnizipCache:        a.AnizipCache,
		BaseAnimeCache:     anilist.NewBaseAnimeCache(),
		CompleteAnimeCache: anilist.NewCompleteAnimeCache(),
		MetadataProvider:   a.MetadataProvider,
		TorrentRepository:  a.TorrentRepository,
		Platform:           a.AnilistPlatform,
		PlaybackManager:    a.PlaybackManager,
		WSEventManager:     a.WSEventManager,
		Database:           a.Database,
	})

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
	if err != nil {
		a.Logger.Warn().Msg("app: Did not initialize modules, no settings found")
		return
	}

	a.Settings = settings // Store settings instance in app
	if settings != nil && settings.Library != nil {
		a.LibraryDir = settings.Library.LibraryPath
	}

	// +---------------------+
	// |   Module settings   |
	// +---------------------+
	// Refresh settings of modules that were initialized in initModulesOnce

	notifier.GlobalNotifier.SetSettings(a.Config.Data.AppDataDir, a.Settings.Notifications, a.Logger)

	// Refresh updater settings
	if settings.Library != nil && a.Updater != nil {
		a.Updater.SetEnabled(!settings.Library.DisableUpdateCheck)
	}

	// Refresh auto scanner settings
	if settings.Library != nil && a.AutoScanner != nil {

		a.AutoScanner.SetEnabled(settings.Library.AutoScan)

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
		a.MediaPlayer.Mpv = mpv.New(a.Logger, settings.MediaPlayer.MpvSocket, settings.MediaPlayer.MpvPath)

		// Set media player repository
		a.MediaPlayerRepository = mediaplayer.NewRepository(&mediaplayer.NewRepositoryOptions{
			Logger:         a.Logger,
			Default:        settings.MediaPlayer.Default,
			VLC:            a.MediaPlayer.VLC,
			MpcHc:          a.MediaPlayer.MpcHc,
			Mpv:            a.MediaPlayer.Mpv, // Socket
			WSEventManager: a.WSEventManager,
		})

		a.PlaybackManager.SetMediaPlayerRepository(a.MediaPlayerRepository)
		a.PlaybackManager.SetSettings(&playbackmanager.Settings{
			AutoPlayNextEpisode: a.Settings.Library.AutoPlayNextEpisode,
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
		})
		_ = qbit.Login()
		// Init Transmission
		trans, err := transmission.New(&transmission.NewTransmissionOptions{
			Logger:   a.Logger,
			Username: settings.Torrent.TransmissionUsername,
			Password: settings.Torrent.TransmissionPassword,
			Port:     settings.Torrent.TransmissionPort,
			Path:     settings.Torrent.TransmissionPath,
		})
		if err != nil && settings.Torrent.TransmissionUsername != "" && settings.Torrent.TransmissionPassword != "" { // Only log error if username and password are set
			a.Logger.Error().Err(err).Msg("app: Failed to initialize transmission client")
		}

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
		})

		a.TorrentClientRepository.InitActiveTorrentCount(settings.Torrent.ShowActiveTorrentCount, a.WSEventManager)

		// Set AutoDownloader qBittorrent client
		a.AutoDownloader.SetTorrentClientRepository(a.TorrentClientRepository)
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
			a.initLibraryWatcher(settings.Library.LibraryPath)
		}()
	}

	// +---------------------+
	// |       Discord       |
	// +---------------------+

	if settings.Discord != nil && a.DiscordPresence != nil {
		a.DiscordPresence.SetSettings(settings.Discord, "")
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
			Enabled:                        false,
			AutoSelect:                     true,
			AddToLibrary:                   false,
			StreamingServerHost:            "0.0.0.0",
			StreamingServerPort:            43214,
			TorrentClientPort:              43213,
			FallbackToTorrentStreamingView: false,
		})
		if err != nil {
			a.Logger.Error().Err(err).Msg("app: Failed to initialize mediastream module")
			return
		}
	}

	err := a.TorrentstreamRepository.InitModules(settings, a.Config.Server.Host)
	if err != nil && settings.Enabled {
		a.Logger.Error().Err(err).Msg("app: Failed to initialize Torrent streaming module")
	}

	a.Cleanups = append(a.Cleanups, func() {
		a.TorrentstreamRepository.Shutdown()
	})

	// Set torrent streaming settings in secondary settings
	// so the client can use them
	a.SecondarySettings.Torrentstream = settings
}

// InitOrRefreshAnilistData will initialize the Anilist anime collection and the account.
// This function should be called after App.Database is initialized and after settings are updated.
func (a *App) InitOrRefreshAnilistData() {
	a.Logger.Debug().Msg("app: Fetching Anilist data")

	acc, err := a.Database.GetAccount()
	if err != nil {
		return
	}

	if acc.Token == "" || acc.Username == "" {
		return
	}

	// Set username to Anilist platform
	a.AnilistPlatform.SetUsername(acc.Username)

	// Set account
	a.account = acc
	a.Logger.Info().Msg("app: Authenticated to AniList as " + acc.Username)

	_, err = a.RefreshAnimeCollection()
	if err != nil {
		a.Logger.Error().Err(err).Msg("app: Failed to fetch Anilist collection")
		return
	}

	go func(username string) {
		a.DiscordPresence.SetUsername(username)
	}(a.account.Username)

	a.Logger.Info().Msg("app: Fetched Anilist data")
}

func (a *App) performActionsOnce() {

	go func() {
		if a.Settings == nil || a.Settings.Library == nil {
			return
		}

		if a.Settings.Library.OpenWebURLOnStart {
			// Open the web URL
			err := browser.OpenURL(a.Config.GetServerURI("127.0.0.1"))
			if err != nil {
				a.Logger.Warn().Err(err).Msg("app: Failed to open web URL, please open it manually in your browser")
			} else {
				a.Logger.Info().Msg("app: Opened web URL")
			}
		}

		if a.Settings.Library.RefreshLibraryOnStart {
			go func() {
				a.Logger.Debug().Msg("app: Refreshing library")
				a.AutoScanner.RunNow()
				a.Logger.Info().Msg("app: Refreshed library")
			}()
		}

		if a.Settings.Library.OpenTorrentClientOnStart && a.TorrentClientRepository != nil {
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
