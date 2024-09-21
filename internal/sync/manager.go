package sync

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/samber/mo"
	"os"
	"path/filepath"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/database/db"
	"seanime/internal/database/db_bridge"
	"seanime/internal/library/anime"
	"seanime/internal/manga"
)

var (
	ErrAlreadyTracked = fmt.Errorf("sync: Media already tracked")
)

const (
	AnimeType = "anime"
	MangaType = "manga"

	LocalDir  = "local"
	AssetsDir = "assets"
)

type Manager interface {
	SetAnimeCollection(ac *anilist.AnimeCollection)
	SetMangaCollection(mc *anilist.MangaCollection)
	GetLocalAnimeCollection() mo.Option[*anilist.AnimeCollection]
	GetLocalMangaCollection() mo.Option[*anilist.MangaCollection]
	GetQueue() *Syncer
	// AddAnime adds an anime to track.
	// It checks that the anime is currently in the user's anime collection.
	AddAnime(mId int) error
	// RemoveAnime removes the anime from tracking.
	RemoveAnime(mId int) error
	// AddManga adds a manga to track.
	// It checks that the manga is currently in the user's manga collection.
	AddManga(mId int) error
	// RemoveManga removes a manga from tracking.
	RemoveManga(mId int) error
	// Synchronize syncs all currently tracked media.
	// Compares the local database with the user's anime and manga collections and updates the local database accordingly.
	Synchronize() error
}

type (
	ManagerImpl struct {
		db             *db.Database
		localDb        *Database
		localDir       string
		localAssetsDir string

		logger           *zerolog.Logger
		metadataProvider metadata.Provider
		mangaRepository  *manga.Repository

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
	}

	NewManagerOptions struct {
		DataDir          string
		Logger           *zerolog.Logger
		MetadataProvider metadata.Provider
		MangaRepository  *manga.Repository
		Database         *db.Database
	}
)

func NewManager(opts *NewManagerOptions) (Manager, error) {
	localDir := filepath.Join(opts.DataDir, LocalDir)

	_ = os.MkdirAll(localDir, 0755)

	localDb, err := newLocalSyncDatabase(localDir, "local", opts.Logger)
	if err != nil {
		return nil, err
	}

	ret := &ManagerImpl{
		db:                          opts.Database,
		localDb:                     localDb,
		localDir:                    localDir,
		localAssetsDir:              filepath.Join(localDir, AssetsDir),
		logger:                      opts.Logger,
		animeCollection:             mo.None[*anilist.AnimeCollection](),
		mangaCollection:             mo.None[*anilist.MangaCollection](),
		localAnimeCollection:        mo.None[*anilist.AnimeCollection](),
		localMangaCollection:        mo.None[*anilist.MangaCollection](),
		metadataProvider:            opts.MetadataProvider,
		mangaRepository:             opts.MangaRepository,
		downloadedChapterContainers: make([]*manga.ChapterContainer, 0),
		localFiles:                  make([]*anime.LocalFile, 0),
	}

	ret.syncer = NewQueue(ret)

	// Load the local collections
	localAnimeCollection, ok := ret.localDb.GetLocalAnimeCollection()
	if ok {
		ret.localAnimeCollection = mo.Some(localAnimeCollection)
	}
	localMangaCollection, ok := ret.localDb.GetLocalMangaCollection()
	if ok {
		ret.localMangaCollection = mo.Some(localMangaCollection)
	}

	return ret, nil
}

func (m *ManagerImpl) GetQueue() *Syncer {
	return m.syncer
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

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

// AddAnime adds an anime to track.
// It checks that the anime is currently in the user's anime collection.
// The anime should have local files, or else ManagerImpl.Synchronize will remove it from tracking.
func (m *ManagerImpl) AddAnime(mId int) error {

	m.logger.Trace().Msgf("sync: Adding anime %d to local database", mId)

	s := &TrackedMedia{
		MediaId: mId,
		Type:    AnimeType,
	}

	// Check if the anime is in the user's anime collection
	if m.animeCollection.IsAbsent() {
		return fmt.Errorf("sync: Anime collection not set")
	}

	if _, found := m.animeCollection.MustGet().GetListEntryFromAnimeId(mId); !found {
		return fmt.Errorf("sync: Anime %d not found in user's anime collection", mId)
	}

	if _, found := m.localDb.GetTrackedMedia(mId, AnimeType); found {
		return ErrAlreadyTracked
	}

	err := m.localDb.gormdb.Create(s).Error
	if err != nil {
		return fmt.Errorf("sync: Failed to add anime %d to local database: %w", mId, err)
	}

	return nil
}

func (m *ManagerImpl) RemoveAnime(mId int) error {

	m.logger.Trace().Msgf("sync: Removing anime %d from local database", mId)

	if _, found := m.localDb.GetTrackedMedia(mId, AnimeType); !found {
		return fmt.Errorf("sync: Anime %d not in local database", mId)
	}

	err := m.removeAnime(mId)
	if err != nil {
		return err
	}

	return nil
}

//----------------------------------------------------------------------------------------------------------------------------------------------------

// AddManga adds a manga to track.
// It checks that the manga is currently in the user's manga collection.
// The manga should have downloaded chapter containers, or else ManagerImpl.Synchronize will remove it from tracking.
func (m *ManagerImpl) AddManga(mId int) error {

	m.logger.Trace().Msgf("sync: Adding manga %d to local database", mId)

	s := &TrackedMedia{
		MediaId: mId,
		Type:    MangaType,
	}

	// Check if the manga is in the user's manga collection
	if m.mangaCollection.IsAbsent() {
		return fmt.Errorf("sync: Manga collection not set")
	}

	if _, found := m.mangaCollection.MustGet().GetListEntryFromMangaId(mId); !found {
		return fmt.Errorf("sync: Manga %d not found in user's manga collection", mId)
	}

	if _, found := m.localDb.GetTrackedMedia(mId, MangaType); found {
		return ErrAlreadyTracked
	}

	err := m.localDb.gormdb.Create(s).Error
	if err != nil {
		return fmt.Errorf("sync: Failed to add manga %d to local database: %w", mId, err)
	}

	return nil
}

func (m *ManagerImpl) RemoveManga(mId int) error {

	m.logger.Trace().Msgf("sync: Removing manga %d from local database", mId)

	if _, found := m.localDb.GetTrackedMedia(mId, MangaType); !found {
		return fmt.Errorf("sync: Manga %d not in local database", mId)
	}

	err := m.removeManga(mId)
	if err != nil {
		return err
	}

	return nil
}

//----------------------------------------------------------------------------------------------------------------------------------------------------

// Synchronize should be called after updates to the user's anime and manga collections.
//
//   - After adding/removing an anime or manga to track
//   - After the user's anime and manga collections have been updated (e.g. after a user's anime and manga list has been updated)
//
// It will add media list entries from the user's collection to the Syncer only if the media is tracked.
//   - The Syncer will then synchronize the anime & manga with the local database if needed
//
// It will remove any anime & manga from the local database that are not in the user's collection anymore.
// It will then update the ManagerImpl.localAnimeCollection and ManagerImpl.localMangaCollection
func (m *ManagerImpl) Synchronize() error {

	lfs, _, err := db_bridge.GetLocalFiles(m.db)
	if err != nil {
		return fmt.Errorf("sync: Couldn't start syncing, failed to get local files: %w", err)
	}

	// Check if the anime and manga collections are set
	if m.animeCollection.IsAbsent() {
		return fmt.Errorf("sync: Couldn't start syncing, anime collection not set")
	}

	if m.mangaCollection.IsAbsent() {
		return fmt.Errorf("sync: Couldn't start syncing, manga collection not set")
	}

	mangaChapterContainers, err := m.mangaRepository.GetDownloadedChapterContainers(m.mangaCollection.MustGet())
	if err != nil {
		return fmt.Errorf("sync: Couldn't start syncing, failed to get downloaded chapter containers: %w", err)
	}

	return m.synchronize(lfs, mangaChapterContainers)
}

func (m *ManagerImpl) synchronize(lfs []*anime.LocalFile, mangaChapterContainers []*manga.ChapterContainer) error {

	m.logger.Trace().Msg("sync: Synchronizing local database with user's anime and manga collections")

	m.localFiles = lfs
	m.downloadedChapterContainers = mangaChapterContainers

	// Check if the anime and manga collections are set
	if m.animeCollection.IsAbsent() {
		return fmt.Errorf("sync: Anime collection not set")
	}

	if m.mangaCollection.IsAbsent() {
		return fmt.Errorf("sync: Manga collection not set")
	}

	// Remove anime and manga from the local database that are not in the user's anime and manga collections
	trackedAnime, _ := m.localDb.GetAllTrackedMediaByType(AnimeType)
	trackedManga, _ := m.localDb.GetAllTrackedMediaByType(MangaType)

	trackedAnimeMap := make(map[int]*TrackedMedia)
	for _, item := range trackedAnime {
		trackedAnimeMap[item.MediaId] = item
	}

	trackedMangaMap := make(map[int]*TrackedMedia)
	for _, m := range trackedManga {
		trackedMangaMap[m.MediaId] = m
	}

	for _, item := range trackedAnime {
		// If the anime is not in the user's anime collection, remove it from the local database
		if _, found := m.animeCollection.MustGet().GetListEntryFromAnimeId(item.MediaId); !found {
			err := m.removeAnime(item.MediaId)
			if err != nil {
				return fmt.Errorf("sync: Failed to remove anime %d from local database: %w", item.MediaId, err)
			}
		}
	}

	for _, item := range trackedManga {
		// If the manga is not in the user's manga collection, remove it from the local database
		if _, found := m.mangaCollection.MustGet().GetListEntryFromMangaId(item.MediaId); !found {
			err := m.removeManga(item.MediaId)
			if err != nil {
				return fmt.Errorf("sync: Failed to remove manga %d from local database: %w", item.MediaId, err)
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

func (m *ManagerImpl) removeAnime(aId int) error {
	m.logger.Trace().Msgf("sync: Removing anime %d from local database", aId)
	// Remove the tracked anime
	err := m.localDb.RemoveTrackedMedia(aId, AnimeType)
	if err != nil {
		return fmt.Errorf("sync: Failed to remove anime %d from local database: %w", aId, err)
	}
	// Remove the anime snapshot
	_ = m.localDb.RemoveAnimeSnapshot(aId)
	// Remove the images
	_ = m.removeMediaImages(aId)
	return nil
}

func (m *ManagerImpl) removeManga(mId int) error {
	m.logger.Trace().Msgf("sync: Removing manga %d from local database", mId)
	// Remove the tracked manga
	err := m.localDb.RemoveTrackedMedia(mId, MangaType)
	if err != nil {
		return fmt.Errorf("sync: Failed to remove manga %d from local database: %w", mId, err)
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
	m.logger.Trace().Msgf("sync: Removing images for media %d", mediaId)
	path := filepath.Join(m.localAssetsDir, fmt.Sprintf("%d", mediaId))
	_ = os.RemoveAll(path)
	//if err != nil {
	//	return fmt.Errorf("sync: Failed to remove images for media %d: %w", mediaId, err)
	//}
	return nil
}