package core

import (
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/mpchc"
	"github.com/seanime-app/seanime-server/internal/qbittorrent"
	"github.com/seanime-app/seanime-server/internal/scanner"
	"github.com/seanime-app/seanime-server/internal/vlc"
)

// InitOrRefreshDependencies will initialize or refresh App dependencies.
// This function should be called after App.Database is initialized and after settings are updated.
func (a *App) InitOrRefreshDependencies() {

	// Stop watching if already watching
	if a.Watcher != nil {
		a.Watcher.StopWatching()
	}

	settings, err := a.Database.GetSettings()
	if err != nil {
		a.Logger.Warn().Msg("app: Did not initialize dependencies, no settings found")
		return
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
	} else {
		a.Logger.Warn().Msg("app: Did not initialize qBittorrent module, no settings found")
	}

	// Library watcher

	if settings.Library != nil && len(settings.Library.LibraryPath) > 0 {
		a.InitLibraryWatcher(settings.Library.LibraryPath)
	} else {
		a.Logger.Warn().Msg("app: Did not initialize watcher module, no settings found")
	}

	a.Logger.Info().Msg("app: All dependencies initialized")

}

func (a *App) InitLibraryWatcher(path string) {
	// Create a new matcher
	watcher, err := scanner.NewWatcher(&scanner.NewWatcherOptions{
		Logger: a.Logger,
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

func (a *App) UpdateAnilistClientToken(token string) {
	a.AnilistClient = anilist.NewAuthedClient(token)
}
