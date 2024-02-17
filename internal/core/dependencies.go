package core

import (
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/mpchc"
	"github.com/seanime-app/seanime/internal/mpv"
	"github.com/seanime-app/seanime/internal/qbittorrent"
	"github.com/seanime-app/seanime/internal/scanner"
	"github.com/seanime-app/seanime/internal/vlc"
)

// InitOrRefreshModules will initialize or refresh modules that use settings.
// This function should be called after App.Database is initialized and after settings are updated.
func (a *App) InitOrRefreshModules() {

	// Stop watching if already watching
	if a.Watcher != nil {
		a.Watcher.StopWatching()
	}

	// Get settings from database
	settings, err := a.Database.GetSettings()
	if err != nil {
		a.Logger.Warn().Msg("app: Did not initialize modules, no settings found")
		return
	}

	// Update updater
	if settings.Library != nil && a.Updater != nil {
		a.Updater.CheckForUpdate = !settings.Library.DisableUpdateCheck
	}

	// Update VLC/MPC-HC

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
	} else {
		a.Logger.Warn().Msg("app: Did not initialize media player module, no settings found")
	}

	// Update qBittorrent

	if settings.Torrent != nil {
		a.QBittorrent = qbittorrent.NewClient(&qbittorrent.NewClientOptions{
			Logger:   a.Logger,
			Username: settings.Torrent.QBittorrentUsername,
			Password: settings.Torrent.QBittorrentPassword,
			Port:     settings.Torrent.QBittorrentPort,
			Host:     settings.Torrent.QBittorrentHost,
			Path:     settings.Torrent.QBittorrentPath,
		})
		a.AutoDownloader.QbittorrentClient = a.QBittorrent
	} else {
		a.Logger.Warn().Msg("app: Did not initialize qBittorrent module, no settings found")
	}

	// Update Auto Downloader
	if settings.AutoDownloader != nil {
		go a.AutoDownloader.SetSettings(settings.AutoDownloader)
	}

	// Initialize library watcher
	if settings.Library != nil && len(settings.Library.LibraryPath) > 0 {
		a.initLibraryWatcher(settings.Library.LibraryPath)
	} else {
		a.Logger.Warn().Msg("app: Did not initialize watcher module, no settings found")
	}

	// Save account and Anilist collection
	a.initAnilistData()

	a.Logger.Info().Msg("app: Initialized modules")

}

func (a *App) initAutoDownloader() {
	go a.AutoDownloader.Start()
}

// InitLibraryWatcher will initialize the library watcher.
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
	a.Watcher.StartWatching()

}

// initAnilistData will initialize the Anilist anime collection dependency and the account.
// This function should be called after App.Database is initialized and after settings are updated.
func (a *App) initAnilistData() {

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

func (a *App) UpdateAnilistClientToken(token string) {
	a.AnilistClientWrapper = anilist.NewClientWrapper(token)
}
