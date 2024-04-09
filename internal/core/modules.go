package core

import (
	"context"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/discordrpc/presence"
	"github.com/seanime-app/seanime/internal/library/autodownloader"
	"github.com/seanime-app/seanime/internal/library/autoscanner"
	"github.com/seanime-app/seanime/internal/library/playbackmanager"
	"github.com/seanime-app/seanime/internal/library/scanner"
	"github.com/seanime-app/seanime/internal/manga"
	"github.com/seanime-app/seanime/internal/mediaplayers/mediaplayer"
	"github.com/seanime-app/seanime/internal/mediaplayers/mpchc"
	"github.com/seanime-app/seanime/internal/mediaplayers/mpv"
	"github.com/seanime-app/seanime/internal/mediaplayers/vlc"
	"github.com/seanime-app/seanime/internal/torrents/qbittorrent"
	"github.com/seanime-app/seanime/internal/torrents/torrent_client"
	"github.com/seanime-app/seanime/internal/torrents/transmission"
)

// initModulesOnce will initialize modules that need to persist.
// This function is called once after the App instance is created.
// The settings of these modules will be set/refreshed in InitOrRefreshModules.
func (a *App) initModulesOnce() {

	// Initialize Discord RPC
	// Settings are set in InitOrRefreshModules
	a.DiscordPresence = discordrpc_presence.New(nil, a.Logger)
	a.Cleanups = append(a.Cleanups, func() {
		a.DiscordPresence.Close()
	})

	// Progress manager
	a.PlaybackManager = playbackmanager.New(&playbackmanager.NewPlaybackManagerOptions{
		Logger:               a.Logger,
		WSEventManager:       a.WSEventManager,
		AnilistClientWrapper: a.AnilistClientWrapper,
		Database:             a.Database,
		AnilistCollection:    nil, // Will be set and refreshed in app.RefreshAnilistCollection
		RefreshAnilistCollectionFunc: func() {
			_, _ = a.RefreshAnilistCollection()
		},
		DiscordPresence: a.DiscordPresence,
	})

	// Auto downloader
	a.AutoDownloader = autodownloader.New(&autodownloader.NewAutoDownloaderOptions{
		Logger:                  a.Logger,
		TorrentClientRepository: a.TorrentClientRepository,
		AnilistCollection:       nil, // Will be set and refreshed in app.RefreshAnilistCollection
		Database:                a.Database,
		WSEventManager:          a.WSEventManager,
		AnizipCache:             a.AnizipCache,
	})

	a.AutoDownloader.Start()

	// Auto scanner
	a.AutoScanner = autoscanner.New(&autoscanner.NewAutoScannerOptions{
		Database:             a.Database,
		Enabled:              false,
		AutoDownloader:       a.AutoDownloader,
		AnilistClientWrapper: a.AnilistClientWrapper,
		Logger:               a.Logger,
		WSEventManager:       a.WSEventManager,
	})

	a.AutoScanner.Start()

	// Manga Downloader
	a.MangaDownloader = manga.NewDownloader(&manga.NewDownloaderOptions{
		Database:       a.Database,
		Logger:         a.Logger,
		WSEventManager: a.WSEventManager,
		DownloadDir:    a.Config.Manga.BackupDir,
		Repository:     a.MangaRepository,
	})

	a.MangaDownloader.Start()

}

// InitOrRefreshModules will initialize or refresh modules that depend on settings.
// This function is called:
//   - After the App instance is created
//   - After settings are updated.
func (a *App) InitOrRefreshModules() {
	if a.cancelContext != nil {
		a.Logger.Debug().Msg("app: Avoided concurrent refresh")
		return
	}

	var ctx context.Context
	ctx, a.cancelContext = context.WithCancel(context.Background())
	defer func() {
		ctx.Done()
		a.cancelContext = nil
	}()

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

	// +---------------------+
	// |   Module settings   |
	// +---------------------+
	// Refresh settings of modules that were initialized in initModulesOnce

	// Refresh updater settings
	if settings.Library != nil && a.Updater != nil {
		a.Updater.SetEnabled(!settings.Library.DisableUpdateCheck)
	}

	// Refresh auto scanner settings
	if settings.Library != nil && a.AutoScanner != nil {
		a.AutoScanner.SetEnabled(settings.Library.AutoScan)
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
		a.MediaPlayRepository = mediaplayer.NewRepository(&mediaplayer.NewRepositoryOptions{
			Logger:         a.Logger,
			Default:        settings.MediaPlayer.Default,
			VLC:            a.MediaPlayer.VLC,
			MpcHc:          a.MediaPlayer.MpcHc,
			Mpv:            a.MediaPlayer.Mpv,
			WSEventManager: a.WSEventManager,
		})

		a.PlaybackManager.SetMediaPlayerRepository(a.MediaPlayRepository)
	} else {
		a.Logger.Warn().Msg("app: Did not initialize media player module, no settings found")
	}

	// +---------------------+
	// |   Torrent Client    |
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

		// Set Repository
		a.TorrentClientRepository = torrent_client.NewRepository(&torrent_client.NewRepositoryOptions{
			Logger:            a.Logger,
			QbittorrentClient: qbit,
			Transmission:      trans,
			Provider:          settings.Torrent.Default,
		})

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
		a.initLibraryWatcher(settings.Library.LibraryPath)
	}

	// +---------------------+
	// |       Discord       |
	// +---------------------+

	if settings.Discord != nil && a.DiscordPresence != nil {
		a.DiscordPresence.SetSettings(settings.Discord)
	}

	// +---------------------+
	// |       AniList       |
	// +---------------------+

	a.initAnilistData()

	a.Logger.Info().Msg("app: Initialized modules")

}

// initLibraryWatcher will initialize the library watcher.
//   - Used by AutoScanner
func (a *App) initLibraryWatcher(path string) {
	// Create a new matcher
	watcher, err := scanner.NewWatcher(&scanner.NewWatcherOptions{
		Logger:         a.Logger,
		WSEventManager: a.WSEventManager,
	})
	if err != nil {
		a.Logger.Error().Err(err).Msg("app: Failed to initialize watcher")
		return
	}

	// Initialize library file watcher
	err = watcher.InitLibraryFileWatcher(&scanner.WatchLibraryFilesOptions{
		LibraryPath: path,
	})
	if err != nil {
		a.Logger.Error().Err(err).Msg("app: Failed to watch library files")
		return
	}

	// Set the watcher
	a.Watcher = watcher

	// Start watching
	a.Watcher.StartWatching(
		func() {
			// Notify the auto scanner when a file action occurs
			a.AutoScanner.Notify()
		})

}

// initAnilistData will initialize the Anilist anime collection and the account.
// This function should be called after App.Database is initialized and after settings are updated.
func (a *App) initAnilistData() {
	a.Logger.Debug().Msg("app: Initializing Anilist data")

	acc, err := a.Database.GetAccount()
	if err != nil {
		return
	}

	if acc.Token == "" || acc.Username == "" {
		return
	}

	// Set account
	a.account = acc

	_, err = a.RefreshAnilistCollection()
	if err != nil {
		a.Logger.Error().Err(err).Msg("app: Failed to fetch Anilist collection")
		return
	}

	a.Logger.Info().Msg("app: Fetched Anilist collection")

}

// UpdateAnilistClientToken will update the Anilist Client Wrapper token.
// This function should be called when a user logs in
func (a *App) UpdateAnilistClientToken(token string) {
	a.AnilistClientWrapper = anilist.NewClientWrapper(token)
	a.PlaybackManager.SetAnilistClientWrapper(a.AnilistClientWrapper) // Update Anilist Client Wrapper in Playback Manager
}
