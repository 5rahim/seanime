package torrentstream

import (
	"errors"
	"github.com/rs/zerolog"
	"github.com/samber/mo"
	"os"
	"path/filepath"
	"seanime/internal/api/anilist"
	"seanime/internal/api/anizip"
	"seanime/internal/api/metadata"
	"seanime/internal/database/models"
	"seanime/internal/events"
	"seanime/internal/library/playbackmanager"
	"seanime/internal/mediaplayers/mediaplayer"
	"seanime/internal/platform"
	"seanime/internal/util"
)

type (
	Repository struct {
		client                   *Client
		serverManager            *serverManager
		playback                 playback
		settings                 mo.Option[Settings]           // None by default, set and refreshed by [SetSettings]
		currentEpisodeCollection mo.Option[*EpisodeCollection] // Refreshed in [list.go] when the user opens the streaming page for a media

		// Injected dependencies
		anizipCache                     *anizip.Cache
		baseAnimeCache                  *anilist.BaseAnimeCache
		completeAnimeCache              *anilist.CompleteAnimeCache
		platform                        platform.Platform
		wsEventManager                  events.WSEventManagerInterface
		metadataProvider                *metadata.Provider
		playbackManager                 *playbackmanager.PlaybackManager
		mediaPlayerRepository           *mediaplayer.Repository
		mediaPlayerRepositorySubscriber *mediaplayer.RepositorySubscriber
		logger                          *zerolog.Logger
	}

	Settings struct {
		models.TorrentstreamSettings
	}

	NewRepositoryOptions struct {
		Logger             *zerolog.Logger
		AnizipCache        *anizip.Cache
		BaseAnimeCache     *anilist.BaseAnimeCache
		CompleteAnimeCache *anilist.CompleteAnimeCache
		Platform           platform.Platform
		MetadataProvider   *metadata.Provider
		PlaybackManager    *playbackmanager.PlaybackManager
		WSEventManager     events.WSEventManagerInterface
	}
)

// NewRepository creates a new injectable Repository instance
func NewRepository(opts *NewRepositoryOptions) *Repository {
	ret := &Repository{
		logger:             opts.Logger,
		anizipCache:        opts.AnizipCache,
		baseAnimeCache:     opts.BaseAnimeCache,
		completeAnimeCache: opts.CompleteAnimeCache,
		platform:           opts.Platform,
		metadataProvider:   opts.MetadataProvider,
		playbackManager:    opts.PlaybackManager,
		wsEventManager:     opts.WSEventManager,
	}
	ret.client = NewClient(ret)
	ret.serverManager = newServerManager(ret)
	return ret
}

// setEpisodeCollection sets the current episode collection in the repository.
func (r *Repository) setEpisodeCollection(ec *EpisodeCollection) {
	if ec == nil {
		r.currentEpisodeCollection = mo.None[*EpisodeCollection]()
		return
	}

	r.currentEpisodeCollection = mo.Some(ec)

	// Notify the playback manager
	if r.playbackManager != nil {
		r.playbackManager.SetStreamEpisodeCollection(ec.Episodes)
	}
}

// SetMediaPlayerRepository sets the mediaplayer repository and listens to events.
// This MUST be called after instantiating the repository.
func (r *Repository) SetMediaPlayerRepository(mediaPlayerRepository *mediaplayer.Repository) {
	r.mediaPlayerRepository = mediaPlayerRepository
	r.listenToMediaPlayerEvents()
}

// InitModules sets the settings for the torrentstream module.
// It should be called before any other method, to ensure the module is active.
func (r *Repository) InitModules(settings *models.TorrentstreamSettings, host string) (err error) {
	r.client.Shutdown()

	defer util.HandlePanicInModuleWithError("torrentstream/InitModules", &err)

	if settings == nil {
		r.logger.Error().Msg("torrentstream: Cannot initialize module, no settings provided")
		r.settings = mo.None[Settings]()
		return errors.New("torrentstream: Cannot initialize module, no settings provided")
	}

	s := *settings

	if s.Enabled == false {
		r.logger.Info().Msg("torrentstream: Module is disabled")
		r.Shutdown()
		r.settings = mo.None[Settings]()
		return nil
	}

	// Set default download directory, which is a temporary directory
	if s.DownloadDir == "" {
		s.DownloadDir = r.getDefaultDownloadPath()
		_ = os.MkdirAll(s.DownloadDir, os.ModePerm) // Create the directory if it doesn't exist
	}

	// DEVNOTE: Commented code below causes error log after initializing the client
	//// Empty the download directory
	//_ = os.RemoveAll(s.DownloadDir)

	if s.StreamingServerPort == 0 {
		s.StreamingServerPort = 43214
	}
	if s.TorrentClientPort == 0 {
		s.TorrentClientPort = 43213
	}
	if s.StreamingServerHost == "" {
		s.StreamingServerHost = "0.0.0.0"
	}

	// Set the settings
	r.settings = mo.Some(Settings{
		TorrentstreamSettings: s,
	})

	// Initialize the torrent client
	err = r.client.initializeClient()
	if err != nil {
		return err
	}

	// Initialize the streaming server
	r.serverManager.initializeServer()

	r.logger.Info().Msg("torrentstream: Module initialized")
	return nil
}

func (r *Repository) FailIfNoSettings() error {
	if r.settings.IsAbsent() {
		return errors.New("torrentstream: no settings provided, the module is dormant")
	}
	return nil
}

// Shutdown closes the torrent client and streaming server
// TEST-ONLY
func (r *Repository) Shutdown() {
	if r.settings.IsAbsent() {
		return
	}
	r.logger.Debug().Msg("torrentstream: Shutting down module")
	r.client.Shutdown()
	r.serverManager.stopServer()
}

//// Cleanup shuts down the module and removes the download directory
//func (r *Repository) Cleanup() {
//	if r.settings.IsAbsent() {
//		return
//	}
//	r.client.Close()
//
//	// Remove the download directory
//	downloadDir := r.GetDownloadDir()
//	_ = os.RemoveAll(downloadDir)
//}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) GetDownloadDir() string {
	if r.settings.IsAbsent() {
		return r.getDefaultDownloadPath()
	}
	if r.settings.MustGet().DownloadDir == "" {
		return r.getDefaultDownloadPath()
	}
	return r.settings.MustGet().DownloadDir
}

func (r *Repository) getDefaultDownloadPath() string {
	tempDir := os.TempDir()
	downloadDirPath := filepath.Join(tempDir, "seanime", "torrentstream")
	return downloadDirPath
}
