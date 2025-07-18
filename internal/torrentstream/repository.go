package torrentstream

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/database/db"
	"seanime/internal/database/models"
	"seanime/internal/directstream"
	"seanime/internal/events"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/library/anime"
	"seanime/internal/library/playbackmanager"
	"seanime/internal/mediaplayers/mediaplayer"
	"seanime/internal/nativeplayer"
	"seanime/internal/platforms/platform"
	"seanime/internal/torrents/torrent"
	"seanime/internal/util"
	"seanime/internal/util/result"

	"github.com/rs/zerolog"
	"github.com/samber/mo"
)

type (
	Repository struct {
		client   *Client
		handler  *handler
		playback playback
		settings mo.Option[Settings] // None by default, set and refreshed by [SetSettings]

		selectionHistoryMap *result.Map[int, *hibiketorrent.AnimeTorrent] // Key: AniList media ID

		// Injected dependencies
		torrentRepository               *torrent.Repository
		baseAnimeCache                  *anilist.BaseAnimeCache
		completeAnimeCache              *anilist.CompleteAnimeCache
		platform                        platform.Platform
		wsEventManager                  events.WSEventManagerInterface
		metadataProvider                metadata.Provider
		playbackManager                 *playbackmanager.PlaybackManager
		mediaPlayerRepository           *mediaplayer.Repository
		mediaPlayerRepositorySubscriber *mediaplayer.RepositorySubscriber
		nativePlayerSubscriber          *nativeplayer.Subscriber
		directStreamManager             *directstream.Manager
		nativePlayer                    *nativeplayer.NativePlayer
		logger                          *zerolog.Logger
		db                              *db.Database

		onEpisodeCollectionChanged func(ec *anime.EpisodeCollection)

		previousStreamOptions mo.Option[*StartStreamOptions]
	}

	Settings struct {
		models.TorrentstreamSettings
		Host string
		Port int
	}

	NewRepositoryOptions struct {
		Logger              *zerolog.Logger
		TorrentRepository   *torrent.Repository
		BaseAnimeCache      *anilist.BaseAnimeCache
		CompleteAnimeCache  *anilist.CompleteAnimeCache
		Platform            platform.Platform
		MetadataProvider    metadata.Provider
		PlaybackManager     *playbackmanager.PlaybackManager
		WSEventManager      events.WSEventManagerInterface
		Database            *db.Database
		DirectStreamManager *directstream.Manager
		NativePlayer        *nativeplayer.NativePlayer
	}
)

// NewRepository creates a new injectable Repository instance
func NewRepository(opts *NewRepositoryOptions) *Repository {
	ret := &Repository{
		client:                          nil,
		handler:                         nil,
		settings:                        mo.Option[Settings]{},
		selectionHistoryMap:             result.NewResultMap[int, *hibiketorrent.AnimeTorrent](),
		torrentRepository:               opts.TorrentRepository,
		baseAnimeCache:                  opts.BaseAnimeCache,
		completeAnimeCache:              opts.CompleteAnimeCache,
		platform:                        opts.Platform,
		wsEventManager:                  opts.WSEventManager,
		metadataProvider:                opts.MetadataProvider,
		playbackManager:                 opts.PlaybackManager,
		mediaPlayerRepository:           nil,
		mediaPlayerRepositorySubscriber: nil,
		logger:                          opts.Logger,
		db:                              opts.Database,
		directStreamManager:             opts.DirectStreamManager,
		nativePlayer:                    opts.NativePlayer,
		previousStreamOptions:           mo.None[*StartStreamOptions](),
	}
	ret.client = NewClient(ret)
	ret.handler = newHandler(ret)
	return ret
}

func (r *Repository) IsEnabled() bool {
	return r.settings.IsPresent() && r.settings.MustGet().Enabled && r.client != nil
}

func (r *Repository) GetPreviousStreamOptions() (*StartStreamOptions, bool) {
	return r.previousStreamOptions.OrElse(nil), r.previousStreamOptions.IsPresent()
}

// SetMediaPlayerRepository sets the mediaplayer repository and listens to events.
// This MUST be called after instantiating the repository and will run even if the module is disabled.
//
// // Note: This is also used for Debrid streaming
func (r *Repository) SetMediaPlayerRepository(mediaPlayerRepository *mediaplayer.Repository) {
	r.mediaPlayerRepository = mediaPlayerRepository
	r.listenToMediaPlayerEvents()
}

// InitModules sets the settings for the torrentstream module.
// It should be called before any other method, to ensure the module is active.
func (r *Repository) InitModules(settings *models.TorrentstreamSettings, host string, port int) (err error) {
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
		s.StreamingServerHost = "127.0.0.1"
	}

	// Set the settings
	r.settings = mo.Some(Settings{
		TorrentstreamSettings: s,
		Host:                  host,
		Port:                  port,
	})

	// Initialize the torrent client
	err = r.client.initializeClient()
	if err != nil {
		return err
	}

	// Start listening to native player events
	r.listenToNativePlayerEvents()

	r.logger.Info().Msg("torrentstream: Module initialized")
	return nil
}

func (r *Repository) HTTPStreamHandler() http.Handler {
	return r.handler
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
	r.logger.Debug().Msg("torrentstream: Shutting down module")
	r.client.Shutdown()
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
