package offline

import (
	"errors"
	"github.com/rs/zerolog"
	"os"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/database/db"
	"seanime/internal/events"
	"seanime/internal/manga"
	"seanime/internal/platforms/platform"
	"seanime/internal/util"
	"seanime/internal/util/filecache"
	"seanime/internal/util/image_downloader"
	"sync"
	"time"
)

type HubInterface interface {
	RetrieveCurrentSnapshot() (ret *Snapshot, ok bool)
	GetCurrentSnapshot() (ret *Snapshot, ok bool)
	UpdateAnimeListStatus(mediaId int, progress int, status anilist.MediaListStatus) (err error)
	UpdateEntryListData(mediaId *int, status *anilist.MediaListStatus, score *int, progress *int, startDate *string, endDate *string, t string) (err error)
	UpdateMangaListStatus(mediaId int, progress int, status anilist.MediaListStatus) (err error)
	SyncListData() error
	CreateSnapshot(opts *NewSnapshotOptions) error
	GetLatestSnapshotEntry() (snapshotEntry *SnapshotEntry, err error)
	GetLatestSnapshot(bypassCache bool) (snapshot *Snapshot, err error)
}

type (
	// Hub is a struct that holds all the offline modules.
	Hub struct {
		platform         platform.Platform // Used to fetch anime and manga data from AniList
		metadataProvider metadata.Provider // Provides metadata for anime and manga entries
		wsEventManager   events.WSEventManagerInterface
		mangaRepository  *manga.Repository
		db               *db.Database
		offlineDb        *database // Stores snapshots
		fileCacher       *filecache.Cacher
		logger           *zerolog.Logger
		assetsHandler    *assetsHandler // Handles downloading metadata assets
		offlineDir       string         // Contains database
		assetDir         string         // Contains assets
		isOffline        bool           // User enabled offline mode

		mu              sync.Mutex
		currentSnapshot *Snapshot

		RefreshAnilistCollectionsFunc func()
	}
)

type (
	NewHubOptions struct {
		Platform                    platform.Platform
		WSEventManager              events.WSEventManagerInterface
		MetadataProvider            metadata.Provider
		MangaRepository             *manga.Repository
		Database                    *db.Database
		FileCacher                  *filecache.Cacher
		Logger                      *zerolog.Logger
		OfflineDir                  string
		AssetDir                    string
		IsOffline                   bool
		RefreshAnimeCollectionsFunc func()
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

	//if opts.IsOffline {
	//
	//	if !offlineDb.HasSnapshots() {
	//		opts.Logger.Fatal().Msg("offline hub: No snapshots found")
	//	}
	//
	//	opts.Logger.Info().Msg("offline hub: Offline mode enabled")
	//}

	imgDownloader := image_downloader.NewImageDownloader(opts.AssetDir, opts.Logger)

	return &Hub{
		platform:                      opts.Platform,
		wsEventManager:                opts.WSEventManager,
		metadataProvider:              opts.MetadataProvider,
		mangaRepository:               opts.MangaRepository,
		db:                            opts.Database,
		offlineDb:                     offlineDb,
		fileCacher:                    opts.FileCacher,
		logger:                        opts.Logger,
		offlineDir:                    opts.OfflineDir,
		assetDir:                      opts.AssetDir,
		isOffline:                     opts.IsOffline,
		assetsHandler:                 newAssetsHandler(opts.Logger, imgDownloader),
		RefreshAnilistCollectionsFunc: opts.RefreshAnimeCollectionsFunc,
	}
}

func (h *Hub) RetrieveCurrentSnapshot() (ret *Snapshot, ok bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.currentSnapshot == nil {
		// Refresh current snapshot
		ret, err := h.GetLatestSnapshot(true)
		if err != nil {
			return nil, false
		}
		h.currentSnapshot = ret
	}
	return h.currentSnapshot, true
}

func (h *Hub) GetCurrentSnapshot() (ret *Snapshot, ok bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

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
	h.mu.Lock()
	defer h.mu.Unlock()

	h.logger.Debug().Int("progress", progress).Msg("offline hub: Updating anime list status")

	if h.currentSnapshot == nil {
		// Refresh current snapshot
		ret, err := h.GetLatestSnapshot(true)
		if err != nil {
			return errors.New("snapshot not found")
		}
		h.currentSnapshot = ret
	}

	var snapshotEntry *SnapshotMediaEntry
	snapshotEntry, err = h.offlineDb.GetSnapshotMediaEntry(mediaId, h.currentSnapshot.DbId)
	if err != nil {
		return err
	}

	animeEntry := snapshotEntry.GetAnimeEntry()
	if animeEntry == nil {
		return errors.New("anime entry not found")
	}

	animeEntry.ListData.Progress = progress
	animeEntry.ListData.Status = status

	snapshotEntry.Value = animeEntry.Marshal()

	_, err = h.offlineDb.UpdateSnapshotMediaEntryT(snapshotEntry)
	if err != nil {
		return err
	}

	// Refresh current snapshot
	ret, err := h.GetLatestSnapshot(true)
	if err != nil {
		return err
	}

	h.currentSnapshot = ret

	h.logger.Info().Msg("offline hub: Updated anime list status")

	return
}

func (h *Hub) UpdateEntryListData(
	mediaId *int,
	status *anilist.MediaListStatus,
	score *int,
	progress *int,
	startDate *string,
	endDate *string,
	t string,
) (err error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.logger.Debug().Int("mediaId", *mediaId).Msg("offline hub: Updating anime list data")

	if h.currentSnapshot == nil {
		// Refresh current snapshot
		ret, err := h.GetLatestSnapshot(true)
		if err != nil {
			return errors.New("snapshot not found")
		}
		h.currentSnapshot = ret
	}

	var snapshotEntry *SnapshotMediaEntry
	snapshotEntry, err = h.offlineDb.GetSnapshotMediaEntry(*mediaId, h.currentSnapshot.DbId)
	if err != nil {
		return err
	}

	switch t {
	case "anime":
		entry := snapshotEntry.GetAnimeEntry()
		if entry == nil {
			return errors.New("entry not found")
		}

		if progress != nil {
			entry.ListData.Progress = *progress
		}
		if status != nil {
			entry.ListData.Status = *status
		}
		if score != nil {
			entry.ListData.Score = *score
		}
		if startDate != nil {
			entry.ListData.StartedAt = *startDate
		}
		if endDate != nil {
			entry.ListData.CompletedAt = *endDate
		}

		snapshotEntry.Value = entry.Marshal()

		_, err = h.offlineDb.UpdateSnapshotMediaEntryT(snapshotEntry)
		if err != nil {
			return err
		}
	case "manga":
		entry := snapshotEntry.GetMangaEntry()
		if entry == nil {
			return errors.New("entry not found")
		}
		if progress != nil {
			entry.ListData.Progress = *progress
		}
		if status != nil {
			entry.ListData.Status = *status
		}
		if score != nil {
			entry.ListData.Score = *score
		}
		if startDate != nil {
			entry.ListData.StartedAt = *startDate
		}
		if endDate != nil {
			entry.ListData.CompletedAt = *endDate
		}

		snapshotEntry.Value = entry.Marshal()

		_, err = h.offlineDb.UpdateSnapshotMediaEntryT(snapshotEntry)
		if err != nil {
			return err
		}
	}

	// Refresh current snapshot
	ret, err := h.GetLatestSnapshot(true)
	if err != nil {
		return err
	}

	h.currentSnapshot = ret

	h.logger.Info().Msg("offline hub: Updated anime list data")

	return
}

func (h *Hub) UpdateMangaListStatus(
	mediaId int,
	progress int,
	status anilist.MediaListStatus,
) (err error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.logger.Debug().Int("progress", progress).Msg("offline hub: Updating manga list status")

	if h.currentSnapshot == nil {
		// Refresh current snapshot
		ret, err := h.GetLatestSnapshot(true)
		if err != nil {
			return errors.New("snapshot not found")
		}
		h.currentSnapshot = ret
	}

	var snapshotEntry *SnapshotMediaEntry
	snapshotEntry, err = h.offlineDb.GetSnapshotMediaEntry(mediaId, h.currentSnapshot.DbId)
	if err != nil {
		return err
	}

	mangaEntry := snapshotEntry.GetMangaEntry()
	if mangaEntry == nil {
		return errors.New("manga entry not found")
	}

	mangaEntry.ListData.Progress = progress
	mangaEntry.ListData.Status = status

	snapshotEntry.Value = mangaEntry.Marshal()

	_, err = h.offlineDb.UpdateSnapshotMediaEntryT(snapshotEntry)
	if err != nil {
		return err
	}

	// Refresh current snapshot
	ret, err := h.GetLatestSnapshot(true)
	if err != nil {
		return err
	}

	h.currentSnapshot = ret

	h.logger.Info().Msg("offline hub: Updated manga list status")

	return

}

// SyncListData updates the user's AniList collection once they come back online
func (h *Hub) SyncListData() error {
	defer util.HandlePanicInModuleThen("offline/SyncListData", func() {})

	if h.isOffline {
		return nil
	}

	snapshotItem, err := h.offlineDb.GetLatestSnapshot()
	if err != nil {
		return errors.New("no snapshot found")
	}

	if snapshotItem == nil {
		return errors.New("no snapshot found")
	}

	if snapshotItem.Synced {
		return errors.New("data already synced")
	}

	if !snapshotItem.Used {
		return errors.New("snapshot not used")
	}

	snapshotEntries, err := h.offlineDb.GetSnapshotMediaEntries(snapshotItem.ID)
	if err != nil {
		h.logger.Error().Err(err).Msg("offline hub: Failed to retrieve offline updates")
		return err
	}

	updatedSnapshotEntries := make([]*SnapshotMediaEntry, 0)
	for _, se := range snapshotEntries {
		if se.CreatedAt == se.UpdatedAt {
			continue
		}
		updatedSnapshotEntries = append(updatedSnapshotEntries, se)
	}

	//snapshotItem.Synced = true
	//_, _ = h.offlineDb.UpdateSnapshotT(snapshotItem)

	if len(updatedSnapshotEntries) == 0 {
		return nil
	}

	h.logger.Info().Msg("offline hub: Syncing list data")

	var errI error
	for _, se := range updatedSnapshotEntries {

		var listData *ListData

		switch se.Type {
		case "anime":
			listData = se.GetAnimeEntry().ListData
		case "manga":
			listData = se.GetMangaEntry().ListData
		}

		if listData == nil {
			continue
		}

		//listData.Score = listData.Score * 10

		var startDate *anilist.FuzzyDateInput
		var endDate *anilist.FuzzyDateInput
		if listData.StartedAt != "" {
			parsedDate, err := time.Parse(time.RFC3339, listData.StartedAt)
			if err == nil {
				year := parsedDate.Year()
				month := int(parsedDate.Month())
				day := parsedDate.Day()
				startDate = &anilist.FuzzyDateInput{
					Year:  &year,
					Month: &month,
					Day:   &day,
				}
			}
		}
		if listData.CompletedAt != "" {
			parsedDate, err := time.Parse(time.RFC3339, listData.CompletedAt)
			if err == nil {
				year := parsedDate.Year()
				month := int(parsedDate.Month())
				day := parsedDate.Day()
				endDate = &anilist.FuzzyDateInput{
					Year:  &year,
					Month: &month,
					Day:   &day,
				}
			}
		}

		errI = h.platform.UpdateEntry(
			se.MediaId,
			&listData.Status,
			&listData.Score,
			&listData.Progress,
			startDate,
			endDate,
		)

	}

	if errI != nil {
		h.logger.Error().Err(err).Msg("offline hub: Failed to sync some data. Please try again.")
		return err
	}

	_ = h.offlineDb.DeleteSnapshot(snapshotItem.ID)

	h.RefreshAnilistCollectionsFunc()

	h.wsEventManager.SendEvent(events.RefreshedAnilistAnimeCollection, nil)
	h.wsEventManager.SendEvent(events.RefreshedAnilistMangaCollection, nil)

	return nil
}
