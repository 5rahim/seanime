package local

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/database/db"
	"seanime/internal/database/db_bridge"
	"seanime/internal/events"
	"seanime/internal/library/anime"
	"seanime/internal/manga"
	"seanime/internal/platforms/platform"

	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

var (
	ErrAlreadyTracked = fmt.Errorf("local manager: Media already tracked")
)

const (
	AnimeType = "anime"
	MangaType = "manga"
)

type Manager interface {
	// SetAnimeCollection updates the online anime collection in the manager.
	SetAnimeCollection(ac *anilist.AnimeCollection)
	// SetMangaCollection updates the online manga collection in the manager.
	SetMangaCollection(mc *anilist.MangaCollection)
	// GetLocalAnimeCollection returns the local anime collection stored in the local database.
	GetLocalAnimeCollection() mo.Option[*anilist.AnimeCollection]
	// GetLocalMangaCollection returns the local manga collection stored in the local database.
	GetLocalMangaCollection() mo.Option[*anilist.MangaCollection]
	// UpdateLocalAnimeCollection updates the local anime collection using the online data.
	UpdateLocalAnimeCollection(ac *anilist.AnimeCollection)
	// UpdateLocalMangaCollection updates the local manga collection using the online data.
	UpdateLocalMangaCollection(mc *anilist.MangaCollection)
	// GetOfflineMetadataProvider returns the offline metadata provider.
	GetOfflineMetadataProvider() metadata.Provider
	// GetSyncer returns the syncer (used to synchronize the anime and manga snapshots in the local database).
	GetSyncer() *Syncer
	// TrackAnime adds an anime to track for offline use.
	// It checks that the anime is currently in the user's anime collection.
	TrackAnime(mId int) error
	// UntrackAnime removes the anime from tracking.
	UntrackAnime(mId int) error
	// TrackManga adds a manga to track for offline use.
	// It checks that the manga is currently in the user's manga collection.
	TrackManga(mId int) error
	// UntrackManga removes a manga from tracking.
	UntrackManga(mId int) error
	// IsMediaTracked checks if the media is tracked in the local database.
	IsMediaTracked(aId int, kind string) bool
	// GetTrackedMediaItems returns all tracked media items.
	GetTrackedMediaItems() []*TrackedMediaItem
	// SynchronizeLocal syncs all currently tracked media.
	// Compares the local database with the user's anime and manga collections and updates the local database accordingly.
	SynchronizeLocal() error
	// SynchronizeAnilist syncs the user's AniList data with data stored in the local database.
	SynchronizeAnilist() error
	// SetRefreshAnilistCollectionsFunc sets the function to call to refresh the online AniList collections.
	SetRefreshAnilistCollectionsFunc(func())
	// HasLocalChanges checks if there are any local changes that need to be uploaded or ignored.
	HasLocalChanges() bool
	// SetHasLocalChanges sets the flag to determine if there are local changes that need to be uploaded or ignored.
	SetHasLocalChanges(bool)
	// GetLocalStorageSize returns the size of the local storage in bytes.
	GetLocalStorageSize() int64
	// GetSimulatedAnimeCollection returns the simulated anime collection for unauthenticated users.
	GetSimulatedAnimeCollection() mo.Option[*anilist.AnimeCollection]
	// GetSimulatedMangaCollection returns the simulated manga collection for unauthenticated users.
	GetSimulatedMangaCollection() mo.Option[*anilist.MangaCollection]
	// SaveSimulatedAnimeCollection sets the simulated anime collection for unauthenticated users.
	SaveSimulatedAnimeCollection(ac *anilist.AnimeCollection)
	// SaveSimulatedMangaCollection sets the simulated manga collection for unauthenticated users.
	SaveSimulatedMangaCollection(mc *anilist.MangaCollection)
	// SynchronizeSimulatedCollectionToAnilist synchronizes the simulated anime and manga collections to the user's AniList account.
	SynchronizeSimulatedCollectionToAnilist() error
	// SynchronizeAnilistToSimulatedCollection synchronizes the user's AniList account to the simulated anime and manga collections.
	SynchronizeAnilistToSimulatedCollection() error

	SetOffline(bool)
}

type (
	ManagerImpl struct {
		db             *db.Database
		localDb        *Database
		localDir       string
		localAssetsDir string
		isOffline      bool

		logger                  *zerolog.Logger
		metadataProvider        metadata.Provider
		mangaRepository         *manga.Repository
		wsEventManager          events.WSEventManagerInterface
		offlineMetadataProvider metadata.Provider
		anilistPlatform         platform.Platform

		syncer *Syncer

		// Anime collection stored in the local database, without modifications
		localAnimeCollection mo.Option[*anilist.AnimeCollection]
		// Manga collection stored in the local database, without modifications
		localMangaCollection mo.Option[*anilist.MangaCollection]

		// Anime collection from the user's AniList account, changed by ManagerImpl.SetAnimeCollection
		animeCollection mo.Option[*anilist.AnimeCollection]
		// Manga collection from the user's AniList account, changed by ManagerImpl.SetMangaCollection
		mangaCollection mo.Option[*anilist.MangaCollection]

		// Downloaded chapter containers, set by ManagerImpl.Synchronize, accessed by the synchronization Syncer
		downloadedChapterContainers []*manga.ChapterContainer
		// Local files, set by ManagerImpl.Synchronize, accessed by the synchronization Syncer
		localFiles []*anime.LocalFile

		RefreshAnilistCollectionsFunc func()
	}
	TrackedMediaItem struct {
		MediaId    int                     `json:"mediaId"`
		Type       string                  `json:"type"`
		AnimeEntry *anilist.AnimeListEntry `json:"animeEntry,omitempty"`
		MangaEntry *anilist.MangaListEntry `json:"mangaEntry,omitempty"`
	}

	NewManagerOptions struct {
		LocalDir         string
		AssetDir         string
		Logger           *zerolog.Logger
		MetadataProvider metadata.Provider
		MangaRepository  *manga.Repository
		Database         *db.Database
		WSEventManager   events.WSEventManagerInterface
		AnilistPlatform  platform.Platform
		IsOffline        bool
	}
)

func NewManager(opts *NewManagerOptions) (Manager, error) {

	_ = os.MkdirAll(opts.LocalDir, os.ModePerm)

	localDb, err := newLocalSyncDatabase(opts.LocalDir, "local", opts.Logger)
	if err != nil {
		return nil, err
	}

	ret := &ManagerImpl{
		db:                            opts.Database,
		localDb:                       localDb,
		localDir:                      opts.LocalDir,
		localAssetsDir:                opts.AssetDir,
		logger:                        opts.Logger,
		animeCollection:               mo.None[*anilist.AnimeCollection](),
		mangaCollection:               mo.None[*anilist.MangaCollection](),
		localAnimeCollection:          mo.None[*anilist.AnimeCollection](),
		localMangaCollection:          mo.None[*anilist.MangaCollection](),
		metadataProvider:              opts.MetadataProvider,
		mangaRepository:               opts.MangaRepository,
		downloadedChapterContainers:   make([]*manga.ChapterContainer, 0),
		localFiles:                    make([]*anime.LocalFile, 0),
		wsEventManager:                opts.WSEventManager,
		isOffline:                     opts.IsOffline,
		anilistPlatform:               opts.AnilistPlatform,
		RefreshAnilistCollectionsFunc: func() {},
	}

	ret.syncer = NewQueue(ret)
	ret.offlineMetadataProvider = NewOfflineMetadataProvider(ret)

	// Load the local collections
	ret.loadLocalAnimeCollection()
	ret.loadLocalMangaCollection()

	_ = ret.localDb.GetSettings()

	return ret, nil
}

func (m *ManagerImpl) SetRefreshAnilistCollectionsFunc(f func()) {
	m.RefreshAnilistCollectionsFunc = f
}

func (m *ManagerImpl) GetSyncer() *Syncer {
	return m.syncer
}

func (m *ManagerImpl) GetOfflineMetadataProvider() metadata.Provider {
	return m.offlineMetadataProvider
}

func (m *ManagerImpl) SetOffline(enabled bool) {
	m.isOffline = enabled
}

func (m *ManagerImpl) HasLocalChanges() bool {
	s := m.localDb.GetSettings()
	return s.Updated
}

func (m *ManagerImpl) SetHasLocalChanges(b bool) {
	s := m.localDb.GetSettings()
	if s.Updated == b {
		return
	}
	s.Updated = b
	_ = m.localDb.SaveSettings(s)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (m *ManagerImpl) loadLocalAnimeCollection() {
	collection, ok := m.localDb.GetLocalAnimeCollection()
	if !ok {
		m.localAnimeCollection = mo.None[*anilist.AnimeCollection]()
	}
	m.localAnimeCollection = mo.Some(collection)
}

func (m *ManagerImpl) loadLocalMangaCollection() {
	collection, ok := m.localDb.GetLocalMangaCollection()
	if !ok {
		m.localMangaCollection = mo.None[*anilist.MangaCollection]()
	}
	m.localMangaCollection = mo.Some(collection)
}

func (m *ManagerImpl) SetAnimeCollection(ac *anilist.AnimeCollection) {
	if ac == nil {
		m.animeCollection = mo.None[*anilist.AnimeCollection]()
	} else {
		m.animeCollection = mo.Some[*anilist.AnimeCollection](ac)
	}
}

func (m *ManagerImpl) SetMangaCollection(mc *anilist.MangaCollection) {
	if mc == nil {
		m.mangaCollection = mo.None[*anilist.MangaCollection]()
	} else {
		m.mangaCollection = mo.Some[*anilist.MangaCollection](mc)
	}
}

func (m *ManagerImpl) GetLocalAnimeCollection() mo.Option[*anilist.AnimeCollection] {
	return m.localAnimeCollection
}

func (m *ManagerImpl) GetLocalMangaCollection() mo.Option[*anilist.MangaCollection] {
	return m.localMangaCollection
}

func (m *ManagerImpl) UpdateLocalAnimeCollection(ac *anilist.AnimeCollection) {
	_ = m.localDb.SaveAnimeCollection(ac)
	m.loadLocalAnimeCollection()
}

func (m *ManagerImpl) UpdateLocalMangaCollection(mc *anilist.MangaCollection) {
	_ = m.localDb.SaveMangaCollection(mc)
	m.loadLocalMangaCollection()
}

// TrackAnime adds an anime to track.
// It checks that the anime is currently in the user's anime collection.
// The anime should have local files, or else ManagerImpl.Synchronize will remove it from tracking.
func (m *ManagerImpl) TrackAnime(mId int) error {

	m.logger.Trace().Msgf("local manager: Adding anime %d to local database", mId)

	s := &TrackedMedia{
		MediaId: mId,
		Type:    AnimeType,
	}

	// Check if the anime is in the user's anime collection
	if m.animeCollection.IsAbsent() {
		m.logger.Error().Msg("local manager: Anime collection not set")
		return fmt.Errorf("anime collection not set")
	}

	if _, found := m.animeCollection.MustGet().GetListEntryFromAnimeId(mId); !found {
		m.logger.Error().Msgf("local manager: Anime %d not found in user's anime collection", mId)
		return fmt.Errorf("anime is not in AniList collection")
	}

	if _, found := m.localDb.GetTrackedMedia(mId, AnimeType); found {
		return ErrAlreadyTracked
	}

	err := m.localDb.gormdb.Create(s).Error
	if err != nil {
		m.logger.Error().Msgf("local manager: Failed to add anime %d to local database: %w", mId, err)
		return fmt.Errorf("failed to add anime %d to local database: %w", mId, err)
	}

	return nil
}

func (m *ManagerImpl) UntrackAnime(mId int) error {

	m.logger.Trace().Msgf("local manager: Removing anime %d from local database", mId)

	if _, found := m.localDb.GetTrackedMedia(mId, AnimeType); !found {
		m.logger.Error().Msgf("local manager: Anime %d not in local database", mId)
		return fmt.Errorf("anime is not in local database")
	}

	err := m.removeAnime(mId)
	if err != nil {
		return err
	}

	m.GetSyncer().refreshCollections()

	return nil
}

//----------------------------------------------------------------------------------------------------------------------------------------------------

// TrackManga adds a manga to track.
// It checks that the manga is currently in the user's manga collection.
// The manga should have downloaded chapter containers, or else ManagerImpl.Synchronize will remove it from tracking.
func (m *ManagerImpl) TrackManga(mId int) error {

	m.logger.Trace().Msgf("local manager: Adding manga %d to local database", mId)

	s := &TrackedMedia{
		MediaId: mId,
		Type:    MangaType,
	}

	// Check if the manga is in the user's manga collection
	if m.mangaCollection.IsAbsent() {
		m.logger.Error().Msg("local manager: Manga collection not set")
		return fmt.Errorf("manga collection not set")
	}

	if _, found := m.mangaCollection.MustGet().GetListEntryFromMangaId(mId); !found {
		m.logger.Error().Msgf("local manager: Manga %d not found in user's manga collection", mId)
		return fmt.Errorf("manga is not in AniList collection")
	}

	if _, found := m.localDb.GetTrackedMedia(mId, MangaType); found {
		return ErrAlreadyTracked
	}

	err := m.localDb.gormdb.Create(s).Error
	if err != nil {
		m.logger.Error().Msgf("local manager: Failed to add manga %d to local database: %w", mId, err)
		return fmt.Errorf("failed to add manga %d to local database: %w", mId, err)
	}

	return nil
}

func (m *ManagerImpl) UntrackManga(mId int) error {

	m.logger.Trace().Msgf("local manager: Removing manga %d from local database", mId)

	if _, found := m.localDb.GetTrackedMedia(mId, MangaType); !found {
		m.logger.Error().Msgf("local manager: Manga %d not in local database", mId)
		return fmt.Errorf("manga is not in local database")
	}

	err := m.removeManga(mId)
	if err != nil {
		return err
	}

	m.GetSyncer().refreshCollections()

	return nil
}

//----------------------------------------------------------------------------------------------------------------------------------------------------

func (m *ManagerImpl) IsMediaTracked(aId int, kind string) bool {
	_, found := m.localDb.GetTrackedMedia(aId, kind)
	return found
}

//----------------------------------------------------------------------------------------------------------------------------------------------------

func (m *ManagerImpl) GetTrackedMediaItems() (ret []*TrackedMediaItem) {
	trackedMedia, ok := m.localDb.GetAllTrackedMedia()
	if !ok {
		return
	}

	if m.animeCollection.IsAbsent() || m.mangaCollection.IsAbsent() {
		return
	}

	for _, item := range trackedMedia {
		if item.Type == AnimeType {
			if localAnimeCollection, found := m.localAnimeCollection.Get(); found {
				if e, found := localAnimeCollection.GetListEntryFromAnimeId(item.MediaId); found {
					ret = append(ret, &TrackedMediaItem{
						MediaId:    item.MediaId,
						Type:       item.Type,
						AnimeEntry: e,
					})
					continue
				}
				if e, found := m.animeCollection.MustGet().GetListEntryFromAnimeId(item.MediaId); found {
					ret = append(ret, &TrackedMediaItem{
						MediaId:    item.MediaId,
						Type:       item.Type,
						AnimeEntry: e,
					})
					continue
				}
			}
		} else if item.Type == MangaType {
			if localMangaCollection, found := m.localMangaCollection.Get(); found {
				if e, found := localMangaCollection.GetListEntryFromMangaId(item.MediaId); found {
					ret = append(ret, &TrackedMediaItem{
						MediaId:    item.MediaId,
						Type:       item.Type,
						MangaEntry: e,
					})
					continue
				}
			}
			if e, found := m.mangaCollection.MustGet().GetListEntryFromMangaId(item.MediaId); found {
				ret = append(ret, &TrackedMediaItem{
					MediaId:    item.MediaId,
					Type:       item.Type,
					MangaEntry: e,
				})
				continue
			}
		}
	}

	return
}

//----------------------------------------------------------------------------------------------------------------------------------------------------

// SynchronizeLocal should be called after updates to the user's anime and manga collections.
//
//   - After adding/removing an anime or manga to track
//   - After the user's anime and manga collections have been updated (e.g. after a user's anime and manga list has been updated)
//
// It will add media list entries from the user's collection to the Syncer only if the media is tracked.
//   - The Syncer will then synchronize the anime & manga with the local database if needed
//
// It will remove any anime & manga from the local database that are not in the user's collection anymore.
// It will then update the ManagerImpl.localAnimeCollection and ManagerImpl.localMangaCollection
func (m *ManagerImpl) SynchronizeLocal() error {

	localStorageSizeCache = 0

	m.loadLocalAnimeCollection()
	m.loadLocalMangaCollection()

	settings := m.localDb.GetSettings()
	if settings.Updated {
		return fmt.Errorf("cannot sync, upload or ignore local changes before syncing")
	}

	lfs, _, err := db_bridge.GetLocalFiles(m.db)
	if err != nil {
		return fmt.Errorf("local manager: Couldn't start syncing, failed to get local files: %w", err)
	}

	// Check if the anime and manga collections are set
	if m.animeCollection.IsAbsent() {
		return fmt.Errorf("local manager: Couldn't start syncing, anime collection not set")
	}

	if m.mangaCollection.IsAbsent() {
		return fmt.Errorf("local manager: Couldn't start syncing, manga collection not set")
	}

	mangaChapterContainers, err := m.mangaRepository.GetDownloadedChapterContainers(m.mangaCollection.MustGet())
	if err != nil {
		return fmt.Errorf("local manager: Couldn't start syncing, failed to get downloaded chapter containers: %w", err)
	}

	return m.synchronize(lfs, mangaChapterContainers)
}

func (m *ManagerImpl) synchronize(lfs []*anime.LocalFile, mangaChapterContainers []*manga.ChapterContainer) error {

	m.logger.Trace().Msg("local manager: Synchronizing local database with user's anime and manga collections")

	m.localFiles = lfs
	m.downloadedChapterContainers = mangaChapterContainers

	// Check if the anime and manga collections are set
	if m.animeCollection.IsAbsent() {
		return fmt.Errorf("local manager: Anime collection not set")
	}

	if m.mangaCollection.IsAbsent() {
		return fmt.Errorf("local manager: Manga collection not set")
	}

	trackedAnimeMap, trackedMangaMap := m.loadTrackedMedia()

	// Remove anime and manga from the local database that are not in the user's anime and manga collections
	for _, item := range trackedAnimeMap {
		// If the anime is not in the user's anime collection, remove it from the local database
		if _, found := m.animeCollection.MustGet().GetListEntryFromAnimeId(item.MediaId); !found {
			err := m.removeAnime(item.MediaId)
			if err != nil {
				return fmt.Errorf("local manager: Failed to remove anime %d from local database: %w", item.MediaId, err)
			}
		}
	}
	for _, item := range trackedMangaMap {
		// If the manga is not in the user's manga collection, remove it from the local database
		if _, found := m.mangaCollection.MustGet().GetListEntryFromMangaId(item.MediaId); !found {
			err := m.removeManga(item.MediaId)
			if err != nil {
				return fmt.Errorf("local manager: Failed to remove manga %d from local database: %w", item.MediaId, err)
			}
		}
	}

	// Get snapshots for all tracked anime and manga
	animeSnapshots, _ := m.localDb.GetAnimeSnapshots()
	mangaSnapshots, _ := m.localDb.GetMangaSnapshots()

	// Create a map of the snapshots
	animeSnapshotMap := make(map[int]*AnimeSnapshot)
	for _, snapshot := range animeSnapshots {
		animeSnapshotMap[snapshot.MediaId] = snapshot
	}

	mangaSnapshotMap := make(map[int]*MangaSnapshot)
	for _, snapshot := range mangaSnapshots {
		mangaSnapshotMap[snapshot.MediaId] = snapshot
	}

	m.syncer.runDiffs(trackedAnimeMap, animeSnapshotMap, trackedMangaMap, mangaSnapshotMap, m.localFiles, m.downloadedChapterContainers)

	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (m *ManagerImpl) SynchronizeAnilist() error {
	if m.animeCollection.IsAbsent() {
		return fmt.Errorf("local manager: Anime collection not set")
	}

	if m.mangaCollection.IsAbsent() {
		return fmt.Errorf("local manager: Manga collection not set")
	}

	m.loadLocalAnimeCollection()
	m.loadLocalMangaCollection()

	if localAnimeCollection, ok := m.localAnimeCollection.Get(); ok {
		for _, list := range localAnimeCollection.MediaListCollection.Lists {
			if list.GetStatus() == nil || list.GetEntries() == nil {
				continue
			}
			for _, entry := range list.GetEntries() {
				if entry.GetStatus() == nil {
					continue
				}

				// Get the entry from AniList
				var originalEntry *anilist.AnimeListEntry
				if e, found := m.animeCollection.MustGet().GetListEntryFromAnimeId(entry.GetMedia().GetID()); found {
					originalEntry = e
				}
				if originalEntry == nil {
					continue
				}

				key1 := GetAnimeListDataKey(entry)
				key2 := GetAnimeListDataKey(originalEntry)

				// If the entry is the same, skip
				if key1 == key2 {
					continue
				}

				var startDate *anilist.FuzzyDateInput
				if entry.GetStartedAt() != nil {
					startDate = &anilist.FuzzyDateInput{
						Year:  entry.GetStartedAt().GetYear(),
						Month: entry.GetStartedAt().GetMonth(),
						Day:   entry.GetStartedAt().GetDay(),
					}
				}

				var endDate *anilist.FuzzyDateInput
				if entry.GetCompletedAt() != nil {
					endDate = &anilist.FuzzyDateInput{
						Year:  entry.GetCompletedAt().GetYear(),
						Month: entry.GetCompletedAt().GetMonth(),
						Day:   entry.GetCompletedAt().GetDay(),
					}
				}

				var score *int
				if entry.GetScore() != nil {
					score = lo.ToPtr(int(*entry.GetScore()))
				}

				_ = m.anilistPlatform.UpdateEntry(
					context.Background(),
					entry.GetMedia().GetID(),
					entry.GetStatus(),
					score,
					entry.GetProgress(),
					startDate,
					endDate,
				)
			}
		}
	}

	if localMangaCollection, ok := m.localMangaCollection.Get(); ok {
		for _, list := range localMangaCollection.MediaListCollection.Lists {
			if list.GetStatus() == nil || list.GetEntries() == nil {
				continue
			}
			for _, entry := range list.GetEntries() {
				if entry.GetStatus() == nil {
					continue
				}

				// Get the entry from AniList
				var originalEntry *anilist.MangaListEntry
				if e, found := m.mangaCollection.MustGet().GetListEntryFromMangaId(entry.GetMedia().GetID()); found {
					originalEntry = e
				}
				if originalEntry == nil {
					continue
				}

				key1 := GetMangaListDataKey(entry)
				key2 := GetMangaListDataKey(originalEntry)

				// If the entry is the same, skip
				if key1 == key2 {
					continue
				}

				var startDate *anilist.FuzzyDateInput
				if entry.GetStartedAt() != nil {
					startDate = &anilist.FuzzyDateInput{
						Year:  entry.GetStartedAt().GetYear(),
						Month: entry.GetStartedAt().GetMonth(),
						Day:   entry.GetStartedAt().GetDay(),
					}
				}

				var endDate *anilist.FuzzyDateInput
				if entry.GetCompletedAt() != nil {
					endDate = &anilist.FuzzyDateInput{
						Year:  entry.GetCompletedAt().GetYear(),
						Month: entry.GetCompletedAt().GetMonth(),
						Day:   entry.GetCompletedAt().GetDay(),
					}
				}

				var score *int
				if entry.GetScore() != nil {
					score = lo.ToPtr(int(*entry.GetScore()))
				}

				_ = m.anilistPlatform.UpdateEntry(
					context.Background(),
					entry.GetMedia().GetID(),
					entry.GetStatus(),
					score,
					entry.GetProgress(),
					startDate,
					endDate,
				)
			}
		}
	}

	m.RefreshAnilistCollectionsFunc()

	m.wsEventManager.SendEvent(events.RefreshedAnilistAnimeCollection, nil)
	m.wsEventManager.SendEvent(events.RefreshedAnilistMangaCollection, nil)

	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (m *ManagerImpl) loadTrackedMedia() (trackedAnimeMap map[int]*TrackedMedia, trackedMangaMap map[int]*TrackedMedia) {
	trackedAnime, _ := m.localDb.GetAllTrackedMediaByType(AnimeType)
	trackedManga, _ := m.localDb.GetAllTrackedMediaByType(MangaType)

	trackedAnimeMap = make(map[int]*TrackedMedia)
	for _, item := range trackedAnime {
		trackedAnimeMap[item.MediaId] = item
	}

	trackedMangaMap = make(map[int]*TrackedMedia)
	for _, m := range trackedManga {
		trackedMangaMap[m.MediaId] = m
	}

	m.GetSyncer().trackedMangaMap = trackedMangaMap
	m.GetSyncer().trackedAnimeMap = trackedAnimeMap

	return trackedAnimeMap, trackedMangaMap
}

func (m *ManagerImpl) removeAnime(aId int) error {
	m.logger.Trace().Msgf("local manager: Removing anime %d from local database", aId)
	// Remove the tracked anime
	err := m.localDb.RemoveTrackedMedia(aId, AnimeType)
	if err != nil {
		return fmt.Errorf("local manager: Failed to remove anime %d from local database: %w", aId, err)
	}
	// Remove the anime snapshot
	_ = m.localDb.RemoveAnimeSnapshot(aId)
	// Remove the images
	_ = m.removeMediaImages(aId)
	return nil
}

func (m *ManagerImpl) removeManga(mId int) error {
	m.logger.Trace().Msgf("local manager: Removing manga %d from local database", mId)
	// Remove the tracked manga
	err := m.localDb.RemoveTrackedMedia(mId, MangaType)
	if err != nil {
		return fmt.Errorf("local manager: Failed to remove manga %d from local database: %w", mId, err)
	}
	// Remove the manga snapshot
	_ = m.localDb.RemoveMangaSnapshot(mId)
	// Remove the images
	_ = m.removeMediaImages(mId)
	return nil
}

// removeMediaImages removes the images for the media with the given ID.
//   - The images are stored in the local assets' directory.
//   - e.g. datadir/local/assets/{mediaId}/*
func (m *ManagerImpl) removeMediaImages(mediaId int) error {
	m.logger.Trace().Msgf("local manager: Removing images for media %d", mediaId)
	path := filepath.Join(m.localAssetsDir, fmt.Sprintf("%d", mediaId))
	_ = os.RemoveAll(path)
	//if err != nil {
	//	return fmt.Errorf("local manager: Failed to remove images for media %d: %w", mediaId, err)
	//}
	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Avoids recalculating the size of the cache directory every time it is requested
var localStorageSizeCache int64

func (m *ManagerImpl) GetLocalStorageSize() int64 {

	if localStorageSizeCache != 0 {
		return localStorageSizeCache
	}

	var size int64
	_ = filepath.Walk(m.localDir, func(path string, info os.FileInfo, err error) error {
		if info != nil {
			size += info.Size()
		}
		return nil
	})

	localStorageSizeCache = size

	return size
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (m *ManagerImpl) GetSimulatedAnimeCollection() mo.Option[*anilist.AnimeCollection] {
	ac, ok := m.localDb.GetSimulatedAnimeCollection()
	if !ok {
		return mo.None[*anilist.AnimeCollection]()
	}
	return mo.Some(ac)
}

func (m *ManagerImpl) GetSimulatedMangaCollection() mo.Option[*anilist.MangaCollection] {
	mc, ok := m.localDb.GetSimulatedMangaCollection()
	if !ok {
		return mo.None[*anilist.MangaCollection]()
	}
	return mo.Some(mc)
}

func (m *ManagerImpl) SaveSimulatedAnimeCollection(ac *anilist.AnimeCollection) {
	//// Remove airing dates from each entry
	//for _, list := range ac.MediaListCollection.Lists {
	//	for _, entry := range list.Entries {
	//		entry.GetMedia().NextAiringEpisode = nil
	//	}
	//}
	_ = m.localDb.SaveSimulatedAnimeCollection(ac)
}

func (m *ManagerImpl) SaveSimulatedMangaCollection(mc *anilist.MangaCollection) {
	_ = m.localDb.SaveSimulatedMangaCollection(mc)
}

func (m *ManagerImpl) SynchronizeAnilistToSimulatedCollection() error {
	if animeCollection, ok := m.animeCollection.Get(); ok {
		m.SaveSimulatedAnimeCollection(animeCollection)
	}

	if mangaCollection, ok := m.mangaCollection.Get(); ok {
		m.SaveSimulatedMangaCollection(mangaCollection)
	}

	return nil
}

func (m *ManagerImpl) SynchronizeSimulatedCollectionToAnilist() error {
	if localAnimeCollection, ok := m.localDb.GetSimulatedAnimeCollection(); ok {
		for _, list := range localAnimeCollection.MediaListCollection.Lists {
			if list.GetStatus() == nil || list.GetEntries() == nil {
				continue
			}
			for _, entry := range list.GetEntries() {
				if entry.GetStatus() == nil {
					continue
				}

				// Get the entry from AniList
				var originalEntry *anilist.AnimeListEntry
				if e, found := m.animeCollection.MustGet().GetListEntryFromAnimeId(entry.GetMedia().GetID()); found {
					originalEntry = e
				}
				if originalEntry == nil {
					continue
				}

				key1 := GetAnimeListDataKey(entry)
				key2 := GetAnimeListDataKey(originalEntry)

				// If the entry is the same, skip
				if key1 == key2 {
					continue
				}

				var startDate *anilist.FuzzyDateInput
				if entry.GetStartedAt() != nil {
					startDate = &anilist.FuzzyDateInput{
						Year:  entry.GetStartedAt().GetYear(),
						Month: entry.GetStartedAt().GetMonth(),
						Day:   entry.GetStartedAt().GetDay(),
					}
				}

				var endDate *anilist.FuzzyDateInput
				if entry.GetCompletedAt() != nil {
					endDate = &anilist.FuzzyDateInput{
						Year:  entry.GetCompletedAt().GetYear(),
						Month: entry.GetCompletedAt().GetMonth(),
						Day:   entry.GetCompletedAt().GetDay(),
					}
				}

				var score *int
				if entry.GetScore() != nil {
					score = lo.ToPtr(int(*entry.GetScore()))
				} else {
					score = lo.ToPtr(0)
				}

				_ = m.anilistPlatform.UpdateEntry(
					context.Background(),
					entry.GetMedia().GetID(),
					entry.GetStatus(),
					score,
					entry.GetProgress(),
					startDate,
					endDate,
				)
			}
		}
	}

	if localMangaCollection, ok := m.localDb.GetSimulatedMangaCollection(); ok {
		for _, list := range localMangaCollection.MediaListCollection.Lists {
			if list.GetStatus() == nil || list.GetEntries() == nil {
				continue
			}
			for _, entry := range list.GetEntries() {
				if entry.GetStatus() == nil {
					continue
				}

				// Get the entry from AniList
				var originalEntry *anilist.MangaListEntry
				if e, found := m.mangaCollection.MustGet().GetListEntryFromMangaId(entry.GetMedia().GetID()); found {
					originalEntry = e
				}
				if originalEntry == nil {
					continue
				}

				key1 := GetMangaListDataKey(entry)
				key2 := GetMangaListDataKey(originalEntry)

				// If the entry is the same, skip
				if key1 == key2 {
					continue
				}

				var startDate *anilist.FuzzyDateInput
				if entry.GetStartedAt() != nil {
					startDate = &anilist.FuzzyDateInput{
						Year:  entry.GetStartedAt().GetYear(),
						Month: entry.GetStartedAt().GetMonth(),
						Day:   entry.GetStartedAt().GetDay(),
					}
				}

				var endDate *anilist.FuzzyDateInput
				if entry.GetCompletedAt() != nil {
					endDate = &anilist.FuzzyDateInput{
						Year:  entry.GetCompletedAt().GetYear(),
						Month: entry.GetCompletedAt().GetMonth(),
						Day:   entry.GetCompletedAt().GetDay(),
					}
				}

				var score *int
				if entry.GetScore() != nil {
					score = lo.ToPtr(int(*entry.GetScore()))
				} else {
					score = lo.ToPtr(0)
				}

				_ = m.anilistPlatform.UpdateEntry(
					context.Background(),
					entry.GetMedia().GetID(),
					entry.GetStatus(),
					score,
					entry.GetProgress(),
					startDate,
					endDate,
				)
			}
		}
	}

	m.RefreshAnilistCollectionsFunc()

	m.wsEventManager.SendEvent(events.RefreshedAnilistAnimeCollection, nil)
	m.wsEventManager.SendEvent(events.RefreshedAnilistMangaCollection, nil)

	return nil
}
