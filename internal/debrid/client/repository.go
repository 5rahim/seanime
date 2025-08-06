package debrid_client

import (
	"context"
	"fmt"
	"path/filepath"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/database/db"
	"seanime/internal/database/models"
	"seanime/internal/debrid/debrid"
	"seanime/internal/debrid/realdebrid"
	"seanime/internal/debrid/torbox"
	"seanime/internal/directstream"
	"seanime/internal/events"
	"seanime/internal/library/playbackmanager"
	"seanime/internal/platforms/platform"
	"seanime/internal/torrents/torrent"
	"seanime/internal/util/result"

	"github.com/rs/zerolog"
	"github.com/samber/mo"
)

var (
	ErrProviderNotSet = fmt.Errorf("debrid: Provider not set")
)

type (
	Repository struct {
		provider               mo.Option[debrid.Provider]
		logger                 *zerolog.Logger
		db                     *db.Database
		settings               *models.DebridSettings
		wsEventManager         events.WSEventManagerInterface
		ctxMap                 *result.Map[string, context.CancelFunc]
		downloadLoopCancelFunc context.CancelFunc
		torrentRepository      *torrent.Repository
		directStreamManager    *directstream.Manager

		playbackManager    *playbackmanager.PlaybackManager
		streamManager      *StreamManager
		completeAnimeCache *anilist.CompleteAnimeCache
		metadataProvider   metadata.Provider
		platform           platform.Platform

		previousStreamOptions mo.Option[*StartStreamOptions]
	}

	NewRepositoryOptions struct {
		Logger         *zerolog.Logger
		WSEventManager events.WSEventManagerInterface
		Database       *db.Database

		TorrentRepository   *torrent.Repository
		PlaybackManager     *playbackmanager.PlaybackManager
		DirectStreamManager *directstream.Manager
		MetadataProvider    metadata.Provider
		Platform            platform.Platform
	}
)

func NewRepository(opts *NewRepositoryOptions) (ret *Repository) {
	ret = &Repository{
		provider:       mo.None[debrid.Provider](),
		logger:         opts.Logger,
		wsEventManager: opts.WSEventManager,
		db:             opts.Database,
		settings: &models.DebridSettings{
			Enabled: false,
		},
		torrentRepository:     opts.TorrentRepository,
		platform:              opts.Platform,
		playbackManager:       opts.PlaybackManager,
		metadataProvider:      opts.MetadataProvider,
		completeAnimeCache:    anilist.NewCompleteAnimeCache(),
		ctxMap:                result.NewResultMap[string, context.CancelFunc](),
		previousStreamOptions: mo.None[*StartStreamOptions](),
		directStreamManager:   opts.DirectStreamManager,
	}

	ret.streamManager = NewStreamManager(ret)

	return
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) startOrStopDownloadLoop() {
	// Cancel the previous download loop if it's running
	if r.downloadLoopCancelFunc != nil {
		r.downloadLoopCancelFunc()
	}

	// Start the download loop if the provider is set and enabled
	if r.settings.Enabled && r.provider.IsPresent() {
		ctx, cancel := context.WithCancel(context.Background())
		r.downloadLoopCancelFunc = cancel
		r.launchDownloadLoop(ctx)
	}
}

// InitializeProvider is called each time the settings change
func (r *Repository) InitializeProvider(settings *models.DebridSettings) error {
	r.settings = settings

	if !settings.Enabled {
		r.provider = mo.None[debrid.Provider]()
		// Stop the download loop if it's running
		r.startOrStopDownloadLoop()
		return nil
	}

	switch settings.Provider {
	case "torbox":
		r.provider = mo.Some(torbox.NewTorBox(r.logger))
	case "realdebrid":
		r.provider = mo.Some(realdebrid.NewRealDebrid(r.logger))
	default:
		r.provider = mo.None[debrid.Provider]()
	}

	if r.provider.IsAbsent() {
		r.logger.Warn().Str("provider", settings.Provider).Msg("debrid: No provider set")
		// Stop the download loop if it's running
		r.startOrStopDownloadLoop()
		return nil
	}

	// Authenticate the provider
	err := r.provider.MustGet().Authenticate(r.settings.ApiKey)
	if err != nil {
		r.logger.Err(err).Msg("debrid: Failed to authenticate")
		r.provider = mo.None[debrid.Provider]()
		// Cancel the download loop if it's running
		if r.downloadLoopCancelFunc != nil {
			r.downloadLoopCancelFunc()
		}
		return err
	}

	// Start the download loop
	r.startOrStopDownloadLoop()

	return nil
}

func (r *Repository) GetProvider() (debrid.Provider, error) {
	p, found := r.provider.Get()
	if !found {
		return nil, ErrProviderNotSet
	}

	return p, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// AddAndQueueTorrent adds a torrent to the debrid service and queues it for automatic download
func (r *Repository) AddAndQueueTorrent(opts debrid.AddTorrentOptions, destination string, mId int) (string, error) {
	provider, err := r.GetProvider()
	if err != nil {
		return "", err
	}

	if !filepath.IsAbs(destination) {
		return "", fmt.Errorf("debrid: Failed to add torrent, destination must be an absolute path")
	}

	// Add the torrent to the debrid service
	torrentItemId, err := provider.AddTorrent(opts)
	if err != nil {
		return "", err
	}

	// Add the torrent item to the database (so it can be downloaded automatically once it's ready)
	// We ignore the error since it's non-critical
	_ = r.db.InsertDebridTorrentItem(&models.DebridTorrentItem{
		TorrentItemID: torrentItemId,
		Destination:   destination,
		Provider:      provider.GetSettings().ID,
		MediaId:       mId,
	})

	return torrentItemId, nil
}

// GetTorrentInfo retrieves information about a torrent.
// This is used for file section for debrid streaming.
// On Real Debrid, this adds the torrent to the user's account.
func (r *Repository) GetTorrentInfo(opts debrid.GetTorrentInfoOptions) (*debrid.TorrentInfo, error) {
	provider, err := r.GetProvider()
	if err != nil {
		return nil, err
	}

	torrentInfo, err := provider.GetTorrentInfo(opts)
	if err != nil {
		return nil, err
	}

	// Remove non-video files
	torrentInfo.Files = debrid.FilterVideoFiles(torrentInfo.Files)

	return torrentInfo, nil
}

func (r *Repository) HasProvider() bool {
	return r.provider.IsPresent()
}

func (r *Repository) GetSettings() *models.DebridSettings {
	return r.settings
}

// CancelDownload cancels the download for the given item ID
func (r *Repository) CancelDownload(itemID string) error {
	cancelFunc, found := r.ctxMap.Get(itemID)
	if !found {
		return fmt.Errorf("no download found for item ID: %s", itemID)
	}

	// Call the cancel function to cancel the download
	if cancelFunc != nil {
		cancelFunc()
	}

	r.ctxMap.Delete(itemID)

	// Notify that the download has been cancelled
	r.wsEventManager.SendEvent(events.DebridDownloadProgress, map[string]interface{}{
		"status": "cancelled",
		"itemID": itemID,
	})

	return nil
}

func (r *Repository) StartStream(ctx context.Context, opts *StartStreamOptions) error {
	return r.streamManager.startStream(ctx, opts)
}

func (r *Repository) GetStreamURL() (string, bool) {
	return r.streamManager.currentStreamUrl, r.streamManager.currentStreamUrl != ""
}

func (r *Repository) CancelStream(opts *CancelStreamOptions) {
	r.streamManager.cancelStream(opts)
}

func (r *Repository) GetPreviousStreamOptions() (*StartStreamOptions, bool) {
	return r.previousStreamOptions.Get()
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
