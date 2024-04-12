package offline

import (
	"errors"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/api/metadata"
	"github.com/seanime-app/seanime/internal/database/db"
	"github.com/seanime-app/seanime/internal/manga"
	"github.com/seanime-app/seanime/internal/util/filecache"
	"github.com/seanime-app/seanime/internal/util/image_downloader"
	"os"
)

type (
	// Hub is a struct that holds all the offline modules.
	Hub struct {
		anilistClientWrapper anilist.ClientWrapperInterface // Used to fetch anime and manga data from AniList
		metadataProvider     *metadata.Provider             // Provides metadata for anime and manga entries
		mangaRepository      *manga.Repository
		db                   *db.Database
		offlineDb            *database // Stores snapshots
		fileCacher           *filecache.Cacher
		logger               *zerolog.Logger
		assetsHandler        *assetsHandler // Handles downloading metadata assets
		offlineDir           string         // Contains database
		assetDir             string         // Contains assets
		isOffline            bool           // User enabled offline mode

		currentSnapshot *Snapshot
	}
)

type (
	NewHubOptions struct {
		AnilistClientWrapper anilist.ClientWrapperInterface
		MetadataProvider     *metadata.Provider
		MangaRepository      *manga.Repository
		Db                   *db.Database
		FileCacher           *filecache.Cacher
		Logger               *zerolog.Logger
		OfflineDir           string
		AssetDir             string
		IsOffline            bool
	}
)

// NewHub creates a new offline hub.
func NewHub(opts *NewHubOptions) *Hub {

	_ = os.MkdirAll(opts.OfflineDir, 0755)
	_ = os.MkdirAll(opts.AssetDir, 0755)

	offlineDb, err := newDatabase(opts.OfflineDir, "seanime-offline", opts.Logger, opts.IsOffline)
	if err != nil {
		opts.Logger.Fatal().Err(err).Msg("offline hub: Failed to instantiate offline database")
	}

	if opts.IsOffline {

		if !offlineDb.HasSnapshots() {
			opts.Logger.Fatal().Msg("offline hub: No snapshots found")
		}

		opts.Logger.Info().Msg("offline hub: Offline mode enabled")
	}

	imgDownloader := image_downloader.NewImageDownloader(opts.AssetDir, opts.Logger)

	return &Hub{
		anilistClientWrapper: opts.AnilistClientWrapper,
		metadataProvider:     opts.MetadataProvider,
		mangaRepository:      opts.MangaRepository,
		db:                   opts.Db,
		offlineDb:            offlineDb,
		fileCacher:           opts.FileCacher,
		logger:               opts.Logger,
		offlineDir:           opts.OfflineDir,
		assetDir:             opts.AssetDir,
		isOffline:            opts.IsOffline,
		assetsHandler:        newAssetsHandler(opts.Logger, imgDownloader),
	}
}

func (h *Hub) GetCurrentSnapshot() (ret *Snapshot, ok bool) {
	if h.currentSnapshot == nil {
		return nil, false
	}
	return h.currentSnapshot, true
}

func (h *Hub) UpdateAnimeListStatus(
	mediaId int,
	progress int,
	status anilist.MediaListStatus,
) (err error) {

	if h.currentSnapshot == nil {
		return errors.New("snapshot not found")
	}

	snapshotEntry, err := h.offlineDb.GetSnapshotMediaEntry(mediaId, h.currentSnapshot.DbId)
	if err != nil {
		return err
	}

	animeEntry := snapshotEntry.GetAnimeEntry()
	if animeEntry == nil {
		return errors.New("anime entry not found")
	}

	animeEntry.ListData.Progress = progress
	animeEntry.ListData.Status = status

	_, err = h.offlineDb.UpdateSnapshotMediaEntry(mediaId, snapshotEntry.ID, animeEntry.Marshal())

	// Refresh current snapshot
	ret, err := h.GetLatestSnapshot(true)
	if err != nil {
		return err
	}

	h.currentSnapshot = ret

	return
}

func (h *Hub) UpdateMangaListStatus(
	mediaId int,
	progress int,
	status anilist.MediaListStatus,
) (err error) {

	if h.currentSnapshot == nil {
		return errors.New("snapshot not found")
	}

	snapshotEntry, err := h.offlineDb.GetSnapshotMediaEntry(mediaId, h.currentSnapshot.DbId)
	if err != nil {
		return err
	}

	mangaEntry := snapshotEntry.GetMangaEntry()
	if mangaEntry == nil {
		return errors.New("manga entry not found")
	}

	mangaEntry.ListData.Progress = progress
	mangaEntry.ListData.Status = status

	_, err = h.offlineDb.UpdateSnapshotMediaEntry(mediaId, snapshotEntry.ID, mangaEntry.Marshal())

	// Refresh current snapshot
	ret, err := h.GetLatestSnapshot(true)
	if err != nil {
		return err
	}

	h.currentSnapshot = ret

	return

}
