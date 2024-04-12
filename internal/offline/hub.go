package offline

import (
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

// func (h *Hub) UpdateAnimeListStatus

// func (h *Hub) UpdateMangaListStatus
