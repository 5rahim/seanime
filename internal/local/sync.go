package local

import (
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/events"
	"seanime/internal/library/anime"
	"seanime/internal/manga"
	"seanime/internal/util"
	"seanime/internal/util/result"
	"sync"

	"github.com/samber/lo"
)

// DEVNOTE: The synchronization process is split into 3 parts:
// 1. ManagerImpl.synchronize removes outdated tracked anime & manga, runs Syncer.runDiffs and adds changed tracked anime & manga to the queue.
// 2. The Syncer processes the queue, calling Syncer.synchronizeAnime and Syncer.synchronizeManga for each job.
// 3. Syncer.synchronizeCollections creates a local collection that mirrors the remote collection, containing only the tracked anime & manga. Only called when the queue is emptied.

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type (
	// Syncer will synchronize the anime and manga snapshots in the local database.
	// Anytime Manager.Synchronize is called, tracked anime and manga will be added to the queue.
	// The queue will synchronize one anime and one manga every X minutes, until it's empty.
	//
	// Synchronization can fail due to network issues. When it does, the anime or manga will be added to the failed queue.
	Syncer struct {
		animeJobQueue chan AnimeTask
		mangaJobQueue chan MangaTask

		failedAnimeQueue *result.Cache[int, *anilist.AnimeListEntry]
		failedMangaQueue *result.Cache[int, *anilist.MangaListEntry]

		trackedAnimeMap map[int]*TrackedMedia
		trackedMangaMap map[int]*TrackedMedia

		manager *ManagerImpl
		mu      sync.RWMutex

		shouldUpdateLocalCollections bool
		doneUpdatingLocalCollections chan struct{}

		queueState   QueueState
		queueStateMu sync.RWMutex
	}

	QueueState struct {
		AnimeTasks map[int]*QueueMediaTask `json:"animeTasks"`
		MangaTasks map[int]*QueueMediaTask `json:"mangaTasks"`
	}

	QueueMediaTask struct {
		MediaId int    `json:"mediaId"`
		Image   string `json:"image"`
		Title   string `json:"title"`
		Type    string `json:"type"`
	}
	AnimeTask struct {
		Diff *AnimeDiffResult
	}
	MangaTask struct {
		Diff *MangaDiffResult
	}
)

func NewQueue(manager *ManagerImpl) *Syncer {
	ret := &Syncer{
		animeJobQueue:                make(chan AnimeTask, 100),
		mangaJobQueue:                make(chan MangaTask, 100),
		failedAnimeQueue:             result.NewCache[int, *anilist.AnimeListEntry](),
		failedMangaQueue:             result.NewCache[int, *anilist.MangaListEntry](),
		shouldUpdateLocalCollections: false,
		doneUpdatingLocalCollections: make(chan struct{}, 1),
		manager:                      manager,
		mu:                           sync.RWMutex{},
		queueState: QueueState{
			AnimeTasks: make(map[int]*QueueMediaTask),
			MangaTasks: make(map[int]*QueueMediaTask),
		},
		queueStateMu: sync.RWMutex{},
	}

	go ret.processAnimeJobs()
	go ret.processMangaJobs()

	return ret
}

func (q *Syncer) processAnimeJobs() {
	for job := range q.animeJobQueue {

		q.queueStateMu.Lock()
		q.queueState.AnimeTasks[job.Diff.AnimeEntry.Media.ID] = &QueueMediaTask{
			MediaId: job.Diff.AnimeEntry.Media.ID,
			Image:   job.Diff.AnimeEntry.Media.GetCoverImageSafe(),
			Title:   job.Diff.AnimeEntry.Media.GetPreferredTitle(),
			Type:    "anime",
		}
		q.SendQueueStateToClient()
		q.queueStateMu.Unlock()

		q.shouldUpdateLocalCollections = true
		q.synchronizeAnime(job.Diff)

		q.queueStateMu.Lock()
		delete(q.queueState.AnimeTasks, job.Diff.AnimeEntry.Media.ID)
		q.SendQueueStateToClient()
		q.queueStateMu.Unlock()

		q.checkAndUpdateLocalCollections()
	}
}

func (q *Syncer) processMangaJobs() {
	for job := range q.mangaJobQueue {

		q.queueStateMu.Lock()
		q.queueState.MangaTasks[job.Diff.MangaEntry.Media.ID] = &QueueMediaTask{
			MediaId: job.Diff.MangaEntry.Media.ID,
			Image:   job.Diff.MangaEntry.Media.GetCoverImageSafe(),
			Title:   job.Diff.MangaEntry.Media.GetPreferredTitle(),
			Type:    "manga",
		}
		q.SendQueueStateToClient()
		q.queueStateMu.Unlock()

		q.shouldUpdateLocalCollections = true
		q.synchronizeManga(job.Diff)

		q.queueStateMu.Lock()
		delete(q.queueState.MangaTasks, job.Diff.MangaEntry.Media.ID)
		q.SendQueueStateToClient()
		q.queueStateMu.Unlock()

		q.checkAndUpdateLocalCollections()
	}
}

// checkAndUpdateLocalCollections will synchronize the local collections once the job queue is emptied.
func (q *Syncer) checkAndUpdateLocalCollections() {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Check if we need to update the local collections
	if q.shouldUpdateLocalCollections {
		// Check if both queues are empty
		if len(q.animeJobQueue) == 0 && len(q.mangaJobQueue) == 0 {
			// Update the local collections
			err := q.synchronizeCollections()
			if err != nil {
				q.manager.logger.Error().Err(err).Msg("local manager: Failed to synchronize collections")
			}
			q.SendQueueStateToClient()
			q.manager.wsEventManager.SendEvent(events.SyncLocalFinished, nil)
			q.shouldUpdateLocalCollections = false
			select {
			case q.doneUpdatingLocalCollections <- struct{}{}:
			default:
			}
		}
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (q *Syncer) GetQueueState() QueueState {
	return q.queueState
}

func (q *Syncer) SendQueueStateToClient() {
	q.manager.wsEventManager.SendEvent(events.SyncLocalQueueState, q.GetQueueState())
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// synchronizeCollections should be called after the tracked anime & manga snapshots have been updated.
// The ManagerImpl.animeCollection and ManagerImpl.mangaCollection should be set & up-to-date.
// Instead of modifying the local collections directly, we create new collections that mirror the remote collections, but with up-to-date data.
func (q *Syncer) synchronizeCollections() (err error) {
	defer util.HandlePanicInModuleWithError("sync/synchronizeCollections", &err)

	q.manager.loadTrackedMedia()

	// DEVNOTE: "_" prefix = original/remote collection
	// We shouldn't modify the remote collection, so making sure we get new pointers

	q.manager.logger.Trace().Msg("local manager: Synchronizing local collections")

	_animeCollection := q.manager.animeCollection.MustGet()
	_mangaCollection := q.manager.mangaCollection.MustGet()

	// Get up-to-date snapshots
	animeSnapshots, _ := q.manager.localDb.GetAnimeSnapshots()
	mangaSnapshots, _ := q.manager.localDb.GetMangaSnapshots()

	animeSnapshotMap := make(map[int]*AnimeSnapshot)
	for _, snapshot := range animeSnapshots {
		animeSnapshotMap[snapshot.MediaId] = snapshot
	}

	mangaSnapshotMap := make(map[int]*MangaSnapshot)
	for _, snapshot := range mangaSnapshots {
		mangaSnapshotMap[snapshot.MediaId] = snapshot
	}

	localAnimeCollection := &anilist.AnimeCollection{
		MediaListCollection: &anilist.AnimeCollection_MediaListCollection{
			Lists: []*anilist.AnimeCollection_MediaListCollection_Lists{},
		},
	}

	localMangaCollection := &anilist.MangaCollection{
		MediaListCollection: &anilist.MangaCollection_MediaListCollection{
			Lists: []*anilist.MangaCollection_MediaListCollection_Lists{},
		},
	}

	// Re-create all anime collection lists, without entries
	for _, _animeList := range _animeCollection.MediaListCollection.GetLists() {
		if _animeList.GetStatus() == nil {
			continue
		}
		list := &anilist.AnimeCollection_MediaListCollection_Lists{
			Status:       ToNewPointer(_animeList.Status),
			Name:         ToNewPointer(_animeList.Name),
			IsCustomList: ToNewPointer(_animeList.IsCustomList),
			Entries:      []*anilist.AnimeListEntry{},
		}
		localAnimeCollection.MediaListCollection.Lists = append(localAnimeCollection.MediaListCollection.Lists, list)
	}

	// Re-create all manga collection lists, without entries
	for _, _mangaList := range _mangaCollection.MediaListCollection.GetLists() {
		if _mangaList.GetStatus() == nil {
			continue
		}
		list := &anilist.MangaCollection_MediaListCollection_Lists{
			Status:       ToNewPointer(_mangaList.Status),
			Name:         ToNewPointer(_mangaList.Name),
			IsCustomList: ToNewPointer(_mangaList.IsCustomList),
			Entries:      []*anilist.MangaListEntry{},
		}
		localMangaCollection.MediaListCollection.Lists = append(localMangaCollection.MediaListCollection.Lists, list)
	}

	//visited := make(map[int]struct{})

	if len(animeSnapshots) > 0 {
		// Create local anime collection
		for _, _animeList := range _animeCollection.MediaListCollection.GetLists() {
			if _animeList.GetStatus() == nil {
				continue
			}
			for _, _animeEntry := range _animeList.GetEntries() {
				// Check if the anime is tracked
				_, found := q.trackedAnimeMap[_animeEntry.GetMedia().GetID()]
				if !found {
					continue
				}
				// Get the anime snapshot
				snapshot, found := animeSnapshotMap[_animeEntry.GetMedia().GetID()]
				if !found {
					continue
				}

				// Add the anime to the right list
				for _, list := range localAnimeCollection.MediaListCollection.GetLists() {
					if list.GetStatus() == nil {
						continue
					}

					if *list.GetStatus() != *_animeList.GetStatus() {
						continue
					}

					editedAnime := BaseAnimeDeepCopy(_animeEntry.GetMedia())
					editedAnime.BannerImage = FormatAssetUrl(snapshot.MediaId, snapshot.BannerImagePath)
					editedAnime.CoverImage = &anilist.BaseAnime_CoverImage{
						ExtraLarge: FormatAssetUrl(snapshot.MediaId, snapshot.CoverImagePath),
						Large:      FormatAssetUrl(snapshot.MediaId, snapshot.CoverImagePath),
						Medium:     FormatAssetUrl(snapshot.MediaId, snapshot.CoverImagePath),
						Color:      FormatAssetUrl(snapshot.MediaId, snapshot.CoverImagePath),
					}

					var startedAt *anilist.AnimeCollection_MediaListCollection_Lists_Entries_StartedAt
					if _animeEntry.GetStartedAt() != nil {
						startedAt = &anilist.AnimeCollection_MediaListCollection_Lists_Entries_StartedAt{
							Year:  ToNewPointer(_animeEntry.GetStartedAt().GetYear()),
							Month: ToNewPointer(_animeEntry.GetStartedAt().GetMonth()),
							Day:   ToNewPointer(_animeEntry.GetStartedAt().GetDay()),
						}
					}

					var completedAt *anilist.AnimeCollection_MediaListCollection_Lists_Entries_CompletedAt
					if _animeEntry.GetCompletedAt() != nil {
						completedAt = &anilist.AnimeCollection_MediaListCollection_Lists_Entries_CompletedAt{
							Year:  ToNewPointer(_animeEntry.GetCompletedAt().GetYear()),
							Month: ToNewPointer(_animeEntry.GetCompletedAt().GetMonth()),
							Day:   ToNewPointer(_animeEntry.GetCompletedAt().GetDay()),
						}
					}

					entry := &anilist.AnimeListEntry{
						ID:          _animeEntry.GetID(),
						Score:       ToNewPointer(_animeEntry.GetScore()),
						Progress:    ToNewPointer(_animeEntry.GetProgress()),
						Status:      ToNewPointer(_animeEntry.GetStatus()),
						Notes:       ToNewPointer(_animeEntry.GetNotes()),
						Repeat:      ToNewPointer(_animeEntry.GetRepeat()),
						Private:     ToNewPointer(_animeEntry.GetPrivate()),
						StartedAt:   startedAt,
						CompletedAt: completedAt,
						Media:       editedAnime,
					}
					list.Entries = append(list.Entries, entry)
					break
				}

			}
		}
	}

	if len(mangaSnapshots) > 0 {
		// Create local manga collection
		for _, _mangaList := range _mangaCollection.MediaListCollection.GetLists() {
			if _mangaList.GetStatus() == nil {
				continue
			}
			for _, _mangaEntry := range _mangaList.GetEntries() {
				// Check if the manga is tracked
				_, found := q.trackedMangaMap[_mangaEntry.GetMedia().GetID()]
				if !found {
					continue
				}
				// Get the manga snapshot
				snapshot, found := mangaSnapshotMap[_mangaEntry.GetMedia().GetID()]
				if !found {
					continue
				}

				// Add the manga to the right list
				for _, list := range localMangaCollection.MediaListCollection.GetLists() {
					if list.GetStatus() == nil {
						continue
					}

					if *list.GetStatus() != *_mangaList.GetStatus() {
						continue
					}

					editedManga := BaseMangaDeepCopy(_mangaEntry.GetMedia())
					editedManga.BannerImage = FormatAssetUrl(snapshot.MediaId, snapshot.BannerImagePath)
					editedManga.CoverImage = &anilist.BaseManga_CoverImage{
						ExtraLarge: FormatAssetUrl(snapshot.MediaId, snapshot.CoverImagePath),
						Large:      FormatAssetUrl(snapshot.MediaId, snapshot.CoverImagePath),
						Medium:     FormatAssetUrl(snapshot.MediaId, snapshot.CoverImagePath),
						Color:      FormatAssetUrl(snapshot.MediaId, snapshot.CoverImagePath),
					}

					var startedAt *anilist.MangaCollection_MediaListCollection_Lists_Entries_StartedAt
					if _mangaEntry.GetStartedAt() != nil {
						startedAt = &anilist.MangaCollection_MediaListCollection_Lists_Entries_StartedAt{
							Year:  ToNewPointer(_mangaEntry.GetStartedAt().GetYear()),
							Month: ToNewPointer(_mangaEntry.GetStartedAt().GetMonth()),
							Day:   ToNewPointer(_mangaEntry.GetStartedAt().GetDay()),
						}
					}

					var completedAt *anilist.MangaCollection_MediaListCollection_Lists_Entries_CompletedAt
					if _mangaEntry.GetCompletedAt() != nil {
						completedAt = &anilist.MangaCollection_MediaListCollection_Lists_Entries_CompletedAt{
							Year:  ToNewPointer(_mangaEntry.GetCompletedAt().GetYear()),
							Month: ToNewPointer(_mangaEntry.GetCompletedAt().GetMonth()),
							Day:   ToNewPointer(_mangaEntry.GetCompletedAt().GetDay()),
						}
					}

					entry := &anilist.MangaListEntry{
						ID:          _mangaEntry.GetID(),
						Score:       ToNewPointer(_mangaEntry.GetScore()),
						Progress:    ToNewPointer(_mangaEntry.GetProgress()),
						Status:      ToNewPointer(_mangaEntry.GetStatus()),
						Notes:       ToNewPointer(_mangaEntry.GetNotes()),
						Repeat:      ToNewPointer(_mangaEntry.GetRepeat()),
						Private:     ToNewPointer(_mangaEntry.GetPrivate()),
						StartedAt:   startedAt,
						CompletedAt: completedAt,
						Media:       editedManga,
					}
					list.Entries = append(list.Entries, entry)
					break
				}

			}
		}
	}

	// Save the local collections
	err = q.manager.localDb.SaveAnimeCollection(localAnimeCollection)
	if err != nil {
		return err
	}

	err = q.manager.localDb.SaveMangaCollection(localMangaCollection)
	if err != nil {
		return err
	}

	q.manager.loadLocalAnimeCollection()
	q.manager.loadLocalMangaCollection()

	q.manager.logger.Debug().Msg("local manager: Synchronized local collections")

	return nil
}

//----------------------------------------------------------------------------------------------------------------------------------------------------

func (q *Syncer) sendAnimeToFailedQueue(entry *anilist.AnimeListEntry) {
	q.failedAnimeQueue.Set(entry.Media.ID, entry)
}

func (q *Syncer) sendMangaToFailedQueue(entry *anilist.MangaListEntry) {
	q.failedMangaQueue.Set(entry.Media.ID, entry)
}

//----------------------------------------------------------------------------------------------------------------------------------------------------

func (q *Syncer) refreshCollections() {

	q.manager.logger.Trace().Msg("local manager: Refreshing collections")

	if len(q.animeJobQueue) > 0 || len(q.mangaJobQueue) > 0 {
		q.manager.logger.Trace().Msg("local manager: Skipping refreshCollections, job queues are not empty")
		return
	}

	q.shouldUpdateLocalCollections = true
	q.checkAndUpdateLocalCollections()
}

// runDiffs runs the diffing process to find outdated anime & manga.
// The diffs are then added to the job queues for synchronization.
func (q *Syncer) runDiffs(
	trackedAnimeMap map[int]*TrackedMedia,
	trackedAnimeSnapshotMap map[int]*AnimeSnapshot,
	trackedMangaMap map[int]*TrackedMedia,
	trackedMangaSnapshotMap map[int]*MangaSnapshot,
	localFiles []*anime.LocalFile,
	downloadedChapterContainers []*manga.ChapterContainer,
) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.manager.logger.Trace().Msg("local manager: Running diffs")

	if q.manager.animeCollection.IsAbsent() {
		q.manager.logger.Error().Msg("local manager: Cannot get diffs, anime collection is absent")
		return
	}

	if q.manager.mangaCollection.IsAbsent() {
		q.manager.logger.Error().Msg("local manager: Cannot get diffs, manga collection is absent")
		return
	}

	if len(q.animeJobQueue) > 0 || len(q.mangaJobQueue) > 0 {
		q.manager.logger.Trace().Msg("local manager: Skipping diffs, job queues are not empty")
		return
	}

	diff := &Diff{
		Logger: q.manager.logger,
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	var animeDiffs map[int]*AnimeDiffResult

	go func() {
		animeDiffs = diff.GetAnimeDiffs(GetAnimeDiffOptions{
			Collection:      q.manager.animeCollection.MustGet(),
			LocalCollection: q.manager.localAnimeCollection,
			LocalFiles:      localFiles,
			TrackedAnime:    trackedAnimeMap,
			Snapshots:       trackedAnimeSnapshotMap,
		})
		wg.Done()
		//q.manager.logger.Trace().Msg("local manager: Finished getting anime diffs")
	}()

	var mangaDiffs map[int]*MangaDiffResult

	go func() {
		mangaDiffs = diff.GetMangaDiffs(GetMangaDiffOptions{
			Collection:                  q.manager.mangaCollection.MustGet(),
			LocalCollection:             q.manager.localMangaCollection,
			DownloadedChapterContainers: downloadedChapterContainers,
			TrackedManga:                trackedMangaMap,
			Snapshots:                   trackedMangaSnapshotMap,
		})
		wg.Done()
		//q.manager.logger.Trace().Msg("local manager: Finished getting manga diffs")
	}()

	wg.Wait()

	// Add the diffs to be synced asynchronously
	go func() {
		q.manager.logger.Trace().Int("animeJobs", len(animeDiffs)).Int("mangaJobs", len(mangaDiffs)).Msg("local manager: Adding diffs to the job queues")

		for _, i := range animeDiffs {
			q.animeJobQueue <- AnimeTask{Diff: i}
		}
		for _, i := range mangaDiffs {
			q.mangaJobQueue <- MangaTask{Diff: i}
		}

		if len(animeDiffs) == 0 && len(mangaDiffs) == 0 {
			q.manager.logger.Trace().Msg("local manager: No diffs found")
			//q.refreshCollections()
		}
	}()

	// Done
	q.manager.logger.Trace().Msg("local manager: Done running diffs")
}

//----------------------------------------------------------------------------------------------------------------------------------------------------

// synchronizeAnime creates or updates the anime snapshot in the local database.
// The anime should be tracked.
//   - If the anime has no local files, it will be removed entirely from the local database.
//   - If the anime has local files, we create or update the snapshot.
func (q *Syncer) synchronizeAnime(diff *AnimeDiffResult) {
	defer util.HandlePanicInModuleThen("sync/synchronizeAnime", func() {})

	entry := diff.AnimeEntry

	if entry == nil {
		return
	}

	q.manager.logger.Trace().Msgf("local manager: Starting synchronization of anime %d, diff type: %+v", entry.Media.ID, diff.DiffType)

	lfs := lo.Filter(q.manager.localFiles, func(f *anime.LocalFile, _ int) bool {
		return f.MediaId == entry.Media.ID
	})

	// If the anime (which is tracked) has no local files, remove it entirely from the local database
	if len(lfs) == 0 {
		q.manager.logger.Warn().Msgf("local manager: No local files found for anime %d, removing from the local database", entry.Media.ID)
		_ = q.manager.removeAnime(entry.Media.ID)
		return
	}

	var animeMetadata *metadata.AnimeMetadata
	var metadataWrapper metadata.AnimeMetadataWrapper
	if diff.DiffType == DiffTypeMissing || diff.DiffType == DiffTypeMetadata {
		// Get the anime metadata
		var err error
		animeMetadata, err = q.manager.metadataProvider.GetAnimeMetadata(metadata.AnilistPlatform, entry.Media.ID)
		if err != nil {
			q.sendAnimeToFailedQueue(entry)
			q.manager.logger.Error().Err(err).Msgf("local manager: Failed to get metadata for anime %d", entry.Media.ID)
			return
		}

		metadataWrapper = q.manager.metadataProvider.GetAnimeMetadataWrapper(diff.AnimeEntry.Media, animeMetadata)
	}

	//
	// The snapshot is missing
	//
	if diff.DiffType == DiffTypeMissing && animeMetadata != nil {
		bannerImage, coverImage, episodeImagePaths, ok := DownloadAnimeImages(q.manager.logger, q.manager.localAssetsDir, entry, animeMetadata, metadataWrapper, lfs)
		if !ok {
			q.sendAnimeToFailedQueue(entry)
			return
		}

		// Create a new snapshot
		snapshot := &AnimeSnapshot{
			MediaId:           entry.GetMedia().GetID(),
			AnimeMetadata:     LocalAnimeMetadata(*animeMetadata),
			BannerImagePath:   bannerImage,
			CoverImagePath:    coverImage,
			EpisodeImagePaths: episodeImagePaths,
			ReferenceKey:      GetAnimeReferenceKey(entry.GetMedia(), q.manager.localFiles),
		}

		// Save the snapshot
		err := q.manager.localDb.SaveAnimeSnapshot(snapshot)
		if err != nil {
			q.sendAnimeToFailedQueue(entry)
			q.manager.logger.Error().Err(err).Msgf("local manager: Failed to save anime snapshot for anime %d", entry.GetMedia().GetID())
		}
		return
	}

	//
	// The snapshot metadata is outdated (local files have changed)
	// Update the anime metadata & download the new episode images if needed
	//
	if diff.DiffType == DiffTypeMetadata && diff.AnimeSnapshot != nil && animeMetadata != nil {

		snapshot := *diff.AnimeSnapshot
		snapshot.AnimeMetadata = LocalAnimeMetadata(*animeMetadata)
		snapshot.ReferenceKey = GetAnimeReferenceKey(entry.GetMedia(), q.manager.localFiles)

		// Get the current episode image URLs
		currentEpisodeImageUrls := make(map[string]string)
		for episodeNum, episode := range animeMetadata.Episodes {
			if episode.Image == "" {
				continue
			}
			currentEpisodeImageUrls[episodeNum] = episode.Image
		}

		// Get the episode image URLs that we need to download (i.e. the ones that are not in the snapshot)
		episodeImageUrlsToDownload := make(map[string]string)
		// For each current episode image URL, check if the key (episode number) is in the snapshot
		for episodeNum, episodeImageUrl := range currentEpisodeImageUrls {
			if _, found := snapshot.EpisodeImagePaths[episodeNum]; !found {
				episodeImageUrlsToDownload[episodeNum] = episodeImageUrl
			}
		}

		// Download the episode images if needed
		if len(episodeImageUrlsToDownload) > 0 {
			// Download only the episode images that we need to download
			episodeImagePaths, ok := DownloadAnimeEpisodeImages(q.manager.logger, q.manager.localAssetsDir, entry.GetMedia().GetID(), episodeImageUrlsToDownload)
			if !ok {
				// DownloadAnimeEpisodeImages will log the error
				q.sendAnimeToFailedQueue(entry)
				return
			}
			// Update the snapshot by adding the new episode images
			for episodeNum, episodeImagePath := range episodeImagePaths {
				snapshot.EpisodeImagePaths[episodeNum] = episodeImagePath
			}
		}

		// Save the snapshot
		err := q.manager.localDb.SaveAnimeSnapshot(&snapshot)
		if err != nil {
			q.sendAnimeToFailedQueue(entry)
			q.manager.logger.Error().Err(err).Msgf("local manager: Failed to save anime snapshot for anime %d", entry.GetMedia().GetID())
		}
		return
	}

	// The snapshot is up-to-date
	return
}

// synchronizeManga creates or updates the manga snapshot in the local database.
// We know that the manga is tracked.
//   - If the manga has no chapter containers, it will be removed entirely from the local database.
//   - If the manga has chapter containers, we create or update the snapshot.
func (q *Syncer) synchronizeManga(diff *MangaDiffResult) {
	defer util.HandlePanicInModuleThen("sync/synchronizeManga", func() {})

	entry := diff.MangaEntry

	if entry == nil {
		return
	}

	q.manager.logger.Trace().Msgf("local manager: Starting synchronization of manga %d, diff type: %+v", entry.GetMedia().GetID(), diff.DiffType)

	if q.manager.mangaCollection.IsAbsent() {
		return
	}

	eContainers := make([]*manga.ChapterContainer, 0)

	// Get the manga
	listEntry, ok := q.manager.mangaCollection.MustGet().GetListEntryFromMangaId(entry.GetMedia().GetID())
	if !ok {
		q.manager.logger.Error().Msgf("local manager: Failed to get manga")
		return
	}

	if listEntry.GetStatus() == nil {
		return
	}

	// Get all chapter containers for this manga
	// A manga entry can have multiple chapter containers due to different sources
	for _, c := range q.manager.downloadedChapterContainers {
		if c.MediaId == entry.GetMedia().GetID() {
			eContainers = append(eContainers, c)
		}
	}

	// If there are no chapter containers (they may have been deleted), remove the manga from the local database
	if len(eContainers) == 0 {
		_ = q.manager.removeManga(entry.GetMedia().GetID())
		return
	}

	if diff.DiffType == DiffTypeMissing {
		bannerImage, coverImage, ok := DownloadMangaImages(q.manager.logger, q.manager.localAssetsDir, entry)
		if !ok {
			q.sendMangaToFailedQueue(entry)
			return
		}

		// Create a new snapshot
		snapshot := &MangaSnapshot{
			MediaId:           entry.GetMedia().GetID(),
			ChapterContainers: eContainers,
			BannerImagePath:   bannerImage,
			CoverImagePath:    coverImage,
			ReferenceKey:      GetMangaReferenceKey(entry.GetMedia(), eContainers),
		}

		// Save the snapshot
		err := q.manager.localDb.SaveMangaSnapshot(snapshot)
		if err != nil {
			q.sendMangaToFailedQueue(entry)
			q.manager.logger.Error().Err(err).Msgf("local manager: Failed to save manga snapshot for manga %d", entry.GetMedia().GetID())
		}
		return
	}

	if diff.DiffType == DiffTypeMetadata && diff.MangaSnapshot != nil {
		snapshot := *diff.MangaSnapshot

		// Update the snapshot
		snapshot.ChapterContainers = eContainers
		snapshot.ReferenceKey = GetMangaReferenceKey(entry.GetMedia(), eContainers)

		// Save the snapshot
		err := q.manager.localDb.SaveMangaSnapshot(&snapshot)
		if err != nil {
			q.sendMangaToFailedQueue(entry)
			q.manager.logger.Error().Err(err).Msgf("local manager: Failed to save manga snapshot for manga %d", entry.GetMedia().GetID())
		}
		return
	}

	// The snapshot is up-to-date
	return
}
