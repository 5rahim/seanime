package torrentstream

import (
	"errors"
	"github.com/anacrolix/torrent"
	"github.com/rs/zerolog"
	"github.com/samber/mo"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/api/anizip"
	"github.com/seanime-app/seanime/internal/api/metadata"
	"github.com/seanime-app/seanime/internal/database/models"
	"github.com/seanime-app/seanime/internal/library/playbackmanager"
	"github.com/seanime-app/seanime/internal/mediaplayers/mediaplayer"
	"github.com/seanime-app/seanime/internal/torrents/animetosho"
	"github.com/seanime-app/seanime/internal/torrents/nyaa"
	"github.com/seanime-app/seanime/internal/util"
	"os"
	"path/filepath"
)

type (
	Repository struct {
		client        *Client
		serverManager *ServerManager
		playback      *playback

		anizipCache          *anizip.Cache
		baseMediaCache       *anilist.BaseMediaCache
		animeCollection      *anilist.AnimeCollection
		anilistClientWrapper anilist.ClientWrapperInterface

		nyaaSearchCache       *nyaa.SearchCache
		animeToshoSearchCache *animetosho.SearchCache
		metadataProvider      *metadata.Provider

		playbackManager       *playbackmanager.PlaybackManager
		mediaPlayerRepository *mediaplayer.Repository
		settings              mo.Option[Settings] // None by default, set and refreshed by SetSettings
		logger                *zerolog.Logger
	}

	playback struct {
		currentFile    mo.Option[*torrent.File]
		currentTorrent mo.Option[*torrent.Torrent]
	}

	Settings struct {
		models.TorrentstreamSettings
	}

	NewRepositoryOptions struct {
		Logger                *zerolog.Logger
		MediaPlayerRepository *mediaplayer.Repository
		AnizipCache           *anizip.Cache
		BaseMediaCache        *anilist.BaseMediaCache
		AnimeCollection       *anilist.AnimeCollection
		AnilistClientWrapper  anilist.ClientWrapperInterface
		NyaaSearchCache       *nyaa.SearchCache
		AnimeToshoSearchCache *animetosho.SearchCache
		MetadataProvider      *metadata.Provider
		PlaybackManager       *playbackmanager.PlaybackManager
	}
)

// NewRepository creates a new injectable Repository instance
func NewRepository(opts *NewRepositoryOptions) *Repository {
	ret := &Repository{
		logger:                opts.Logger,
		mediaPlayerRepository: opts.MediaPlayerRepository,
		playback: &playback{
			currentFile: mo.None[*torrent.File](),
		},
		anizipCache:           opts.AnizipCache,
		baseMediaCache:        opts.BaseMediaCache,
		animeCollection:       opts.AnimeCollection,
		anilistClientWrapper:  opts.AnilistClientWrapper,
		nyaaSearchCache:       opts.NyaaSearchCache,
		animeToshoSearchCache: opts.AnimeToshoSearchCache,
		metadataProvider:      opts.MetadataProvider,
		playbackManager:       opts.PlaybackManager,
	}
	ret.client = NewClient(ret)
	ret.serverManager = NewServerManager(ret)
	return ret
}

func (r *Repository) SetMediaPlayerRepository(mediaPlayerRepository *mediaplayer.Repository) {
	r.mediaPlayerRepository = mediaPlayerRepository
}

func (r *Repository) SetAnimeCollection(ac *anilist.AnimeCollection) {
	r.animeCollection = ac
}

// InitModules sets the settings for the torrentstream module
// It should be called before any other method, to ensure the module is active
func (r *Repository) InitModules(settings *models.TorrentstreamSettings, host string) (err error) {
	r.client.Close()

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
	err = r.client.InitializeClient()
	if err != nil {
		return err
	}

	// Initialize the streaming server
	r.serverManager.InitializeServer()

	r.logger.Info().Msg("torrentstream: Module initialized")
	return nil
}

func (r *Repository) FailIfNoSettings() error {
	if r.settings.IsAbsent() {
		return errors.New("torrentstream: no settings provided, the module is dormant")
	}
	return nil
}

// Shutdown cleans up the resources used by the module, including closing the client and server
func (r *Repository) Shutdown() {
	if r.settings.IsAbsent() {
		return
	}
	r.client.Close()
	r.serverManager.StopServer()
	_ = os.RemoveAll(r.GetDownloadDir())
}

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
