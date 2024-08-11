package offline

import (
	"github.com/goccy/go-json"
	"github.com/samber/lo"
	"seanime/internal/api/anilist"
	"seanime/internal/api/anizip"
	"seanime/internal/database/db_bridge"
	"seanime/internal/library/anime"
	"seanime/internal/manga"
	"seanime/internal/util/limiter"
	"slices"
	"time"
)

type (
	NewSnapshotOptions struct {
		AnimeToDownload  []int // MediaIds
		DownloadAssetsOf []int // MediaIds
	}
)

// CreateSnapshot creates a snapshot of the current state of the library and stores it for offline use.
// This is called by the user before going offline.
func (h *Hub) CreateSnapshot(opts *NewSnapshotOptions) error {

	h.logger.Debug().Msg("offline hub: Creating snapshot")

	// Get local files
	lfs, _, err := db_bridge.GetLocalFiles(h.db)
	if err != nil {
		return err
	}

	h.logger.Debug().Msg("offline hub: Retrieved local files")

	// Get user
	dbAcc, err := h.db.GetAccount()
	if err != nil {
		return err
	}
	user, err := anime.NewUser(dbAcc)
	if err != nil {
		return err
	}

	h.logger.Debug().Msg("offline hub: Retrieved user")

	//
	// Collections
	//
	animeCollection, err := h.platform.GetAnimeCollection(false)
	if err != nil {
		h.logger.Error().Err(err).Msg("offline hub: [Snapshot] Failed to get Anilist anime collection")
		return err
	}
	mangaCollection, err := h.platform.GetMangaCollection(false)
	if err != nil {
		h.logger.Error().Err(err).Msg("offline hub: [Snapshot] Failed to get Anilist manga collection")
		return err
	}
	collections := Collections{
		AnimeCollection: animeCollection,
		MangaCollection: mangaCollection,
	}

	h.logger.Debug().Msg("offline hub: Retrieved collections")

	//
	// Anime Entries
	//

	animeEntries := make([]*AnimeEntry, 0)

	lfWrapper := anime.NewLocalFileWrapper(lfs)
	lfEntries := lfWrapper.GetLocalEntries()

	anizipCache := anizip.NewCache()

	rateLimiter := limiter.NewLimiter(1*time.Second, 5)
	for _, lfEntry := range lfEntries {
		if !slices.Contains(opts.AnimeToDownload, lfEntry.GetMediaId()) {
			continue
		}

		// Get the media
		listEntry, ok := animeCollection.GetListEntryFromAnimeId(lfEntry.GetMediaId())
		if !ok {
			h.logger.Error().Err(err).Msgf("offline hub: [Snapshot] Failed to get Anilist media %d", lfEntry.GetMediaId())
			return err
		}

		if listEntry.GetStatus() == nil {
			continue
		}

		h.logger.Debug().Msgf("offline hub: Creating media entry snapshot for media %d", lfEntry.GetMediaId())

		rateLimiter.Wait()
		_mediaEntry, err := anime.NewAnimeEntry(&anime.NewAnimeEntryOptions{
			MediaId:          lfEntry.GetMediaId(),
			LocalFiles:       lfs,
			AnizipCache:      anizipCache,
			AnimeCollection:  animeCollection,
			Platform:         h.platform,
			MetadataProvider: h.metadataProvider,
		})
		if err != nil {
			h.logger.Error().Err(err).Msgf("offline hub: [Snapshot] Failed to create media entry for anime %d", lfEntry.GetMediaId())
			return err
		}

		mediaEpisodes := _mediaEntry.Episodes
		// Note: We don't need the BaseAnime in each episode for the snapshot
		// it's a waste of space
		for _, episode := range mediaEpisodes {
			episode.BaseAnime = nil
		}

		// Create the AnimeEntry
		animeEntry := &AnimeEntry{
			MediaId: lfEntry.GetMediaId(),
			ListData: &ListData{
				Score:       int(*listEntry.GetScore()),
				Status:      *listEntry.GetStatus(),
				Progress:    *listEntry.GetProgress(),
				StartedAt:   anilist.FuzzyDateToString(listEntry.StartedAt),
				CompletedAt: anilist.FuzzyDateToString(listEntry.CompletedAt),
			},
			Media:    listEntry.GetMedia(),
			Episodes: mediaEpisodes,
			//DownloadedAssets: slices.Contains(opts.DownloadAssetsOf, lfEntry.GetMediaId()),
			DownloadedAssets: true,
		}

		// Add the AnimeEntry
		animeEntries = append(animeEntries, animeEntry)

		time.Sleep(1 * time.Second)
	}

	h.logger.Debug().Msg("offline hub: Generated anime entries")

	//
	// Manga Entries
	//

	mangaEntries := make([]*MangaEntry, 0)

	containers, err := h.mangaRepository.GetDownloadedChapterContainers(mangaCollection)
	if err != nil {
		h.logger.Error().Err(err).Msg("offline hub: [Snapshot] Failed to get downloaded manga chapters")
		return err
	}

	uniqContainers := lo.UniqBy(containers, func(c *manga.ChapterContainer) int {
		return c.MediaId
	})

	for _, container := range uniqContainers {
		// Get the media
		listEntry, ok := mangaCollection.GetListEntryFromMediaId(container.MediaId)
		if !ok {
			h.logger.Error().Err(err).Msgf("offline hub: [Snapshot] Failed to get Anilist media %d", container.MediaId)
			return err
		}

		if listEntry.GetStatus() == nil {
			continue
		}

		h.logger.Debug().Msgf("offline hub: Creating media entry snapshot for manga %d", container.MediaId)

		// Get all chapter containers for this media
		// A manga entry can have multiple chapter containers due to different sources
		eContainers := make([]*manga.ChapterContainer, 0)
		for _, c := range containers {
			if c.MediaId == container.MediaId {
				eContainers = append(eContainers, c)
			}
		}

		// Create the MangaEntry
		mangaEntry := &MangaEntry{
			MediaId: container.MediaId,
			ListData: &ListData{
				Score:       int(*listEntry.GetScore()),
				Status:      *listEntry.GetStatus(),
				Progress:    *listEntry.GetProgress(),
				StartedAt:   anilist.FuzzyDateToString(listEntry.StartedAt),
				CompletedAt: anilist.FuzzyDateToString(listEntry.CompletedAt),
			},
			Media:             listEntry.GetMedia(),
			ChapterContainers: eContainers,
			//DownloadedAssets: slices.Contains(opts.DownloadAssetsOf, container.MediaId),
			DownloadedAssets: true,
		}

		// Add the MangaEntry
		mangaEntries = append(mangaEntries, mangaEntry)
	}

	h.logger.Debug().Msg("offline hub: Generated manga entries")

	//
	// DownloadAssets
	//
	assetMap, err := h.assetsHandler.DownloadAssets(animeEntries, mangaEntries, user, opts.DownloadAssetsOf)
	if err != nil {
		h.logger.Error().Err(err).Msg("offline hub: [Snapshot] Failed to download assets")
		return err
	}

	h.logger.Debug().Msg("offline hub: Downloaded assets")

	snapshot := Snapshot{
		User:        user,
		Collections: &collections,
		Entries: &Entries{
			AnimeEntries: animeEntries,
			MangaEntries: mangaEntries,
		},
		AssetMap: assetMap,
	}

	marshaledUser, err := json.Marshal(snapshot.User)
	if err != nil {
		return err
	}

	marshaledCollections, err := json.Marshal(snapshot.Collections)
	if err != nil {
		return err
	}

	marshaledAssetMap, err := json.Marshal(snapshot.AssetMap)
	if err != nil {
		return err
	}

	// Save the snapshot
	snapshotEntry, err := h.offlineDb.InsertSnapshot(marshaledUser, marshaledCollections, marshaledAssetMap)
	if err != nil {
		return err
	}

	h.logger.Info().Msg("offline hub: Saved snapshot")

	// Save the snapshot media entries
	for _, animeEntry := range animeEntries {
		marshaledAnimeEntry, err := json.Marshal(animeEntry)
		if err != nil {
			return err
		}
		_, err = h.offlineDb.InsertSnapshotMediaEntry(snapshotEntry.ID, SnapshotMediaEntryTypeAnime, animeEntry.MediaId, marshaledAnimeEntry)
		if err != nil {
			return err
		}
	}

	for _, mangaEntry := range mangaEntries {
		marshaledMangaEntry, err := json.Marshal(mangaEntry)
		if err != nil {
			return err
		}
		_, err = h.offlineDb.InsertSnapshotMediaEntry(snapshotEntry.ID, SnapshotMediaEntryTypeManga, mangaEntry.MediaId, marshaledMangaEntry)
		if err != nil {
			return err
		}
	}

	h.logger.Info().Msg("offline hub: Saved snapshot media entries")

	return err
}

func (h *Hub) GetLatestSnapshotEntry() (snapshotEntry *SnapshotEntry, err error) {
	return h.offlineDb.GetLatestSnapshot()
}

func (h *Hub) GetLatestSnapshot(bypassCache bool) (snapshot *Snapshot, err error) {

	if h.currentSnapshot != nil && !bypassCache {
		return h.currentSnapshot, nil
	}

	h.logger.Debug().Msg("offline hub: Getting latest snapshot")

	snapshot = &Snapshot{
		User: &anime.User{},
		Entries: &Entries{
			AnimeEntries: make([]*AnimeEntry, 0),
			MangaEntries: make([]*MangaEntry, 0),
		},
		Collections: &Collections{},
		AssetMap:    new(AssetMapImageMap),
	}

	snapshotEntry, err := h.offlineDb.GetLatestSnapshot()
	if err != nil {
		h.logger.Error().Err(err).Msg("offline hub: Failed to get latest snapshot")
		return nil, err
	}

	if snapshotEntry == nil {
		h.logger.Info().Msg("offline hub: No snapshot found")
		return nil, nil
	}

	snapshotEntry.Synced = false
	if h.isOffline {
		snapshotEntry.Used = true
	}
	go func() {
		_, _ = h.offlineDb.UpdateSnapshotT(snapshotEntry)
	}()

	snapshot.DbId = snapshotEntry.ID

	// Get the user
	err = json.Unmarshal(snapshotEntry.User, &snapshot.User)
	if err != nil {
		h.logger.Error().Err(err).Msg("offline hub: Failed to unmarshal user")
		return nil, err
	}

	// Get the collections
	err = json.Unmarshal(snapshotEntry.Collections, &snapshot.Collections)
	if err != nil {
		h.logger.Error().Err(err).Msg("offline hub: Failed to unmarshal collections")
		return nil, err
	}

	// Get the asset map
	err = json.Unmarshal(snapshotEntry.AssetMap, &snapshot.AssetMap)
	if err != nil {
		h.logger.Error().Err(err).Msg("offline hub: Failed to unmarshal asset map")
		return nil, err
	}

	// Get the snapshot media entries
	mediaEntries, err := h.offlineDb.GetSnapshotMediaEntries(snapshotEntry.ID)
	if err != nil {
		h.logger.Error().Err(err).Msg("offline hub: Failed to get snapshot media entries")
		return nil, err
	}

	for _, mediaEntry := range mediaEntries {
		switch SnapshotMediaEntryType(mediaEntry.Type) {
		case SnapshotMediaEntryTypeAnime:
			animeEntry := &AnimeEntry{}
			err = json.Unmarshal(mediaEntry.Value, animeEntry)
			if err != nil {
				h.logger.Error().Err(err).Msg("offline hub: Failed to unmarshal anime entry")
				return nil, err
			}
			snapshot.Entries.AnimeEntries = append(snapshot.Entries.AnimeEntries, animeEntry)
		case SnapshotMediaEntryTypeManga:
			mangaEntry := &MangaEntry{}
			err = json.Unmarshal(mediaEntry.Value, mangaEntry)
			if err != nil {
				h.logger.Error().Err(err).Msg("offline hub: Failed to unmarshal manga entry")
				return nil, err
			}
			snapshot.Entries.MangaEntries = append(snapshot.Entries.MangaEntries, mangaEntry)
		}
	}

	h.logger.Info().Msg("offline hub: Retrieved latest snapshot")

	h.currentSnapshot = snapshot

	return snapshot, nil
}
