package simulated_platform

import (
	"context"
	"errors"
	"seanime/internal/api/anilist"
	"seanime/internal/local"
	"seanime/internal/platforms/platform"
	"seanime/internal/util/limiter"
	"sync"

	"github.com/rs/zerolog"
)

var (
	// ErrMediaNotFound means the media wasn't found in the local collection
	ErrMediaNotFound = errors.New("media not found")
)

// SimulatedPlatform used when the user is not authenticated to AniList.
// It acts as a dummy account using simulated collections stored locally.
type SimulatedPlatform struct {
	logger       *zerolog.Logger
	localManager local.Manager
	client       anilist.AnilistClient // should only receive an unauthenticated client

	// Cache for collections
	animeCollection *anilist.AnimeCollection
	mangaCollection *anilist.MangaCollection
	mu              sync.RWMutex
	collectionMu    sync.RWMutex // used to protect access to collections
}

func NewSimulatedPlatform(localManager local.Manager, client anilist.AnilistClient, logger *zerolog.Logger) (platform.Platform, error) {
	sp := &SimulatedPlatform{
		logger:       logger,
		localManager: localManager,
		client:       client,
	}

	return sp, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Implementation
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (sp *SimulatedPlatform) SetUsername(username string) {
	// no-op
}

func (sp *SimulatedPlatform) SetAnilistClient(client anilist.AnilistClient) {
	sp.client = client // DEVNOTE: Should only be unauthenticated
}

// UpdateEntry updates the entry for the given media ID.
// If the entry doesn't exist, it will be added automatically after determining the media type.
func (sp *SimulatedPlatform) UpdateEntry(mediaID int, status *anilist.MediaListStatus, scoreRaw *int, progress *int, startedAt *anilist.FuzzyDateInput, completedAt *anilist.FuzzyDateInput) error {
	sp.logger.Trace().Int("mediaID", mediaID).Msg("simulated platform: Updating entry")

	sp.mu.Lock()
	defer sp.mu.Unlock()

	// Try anime first
	animeWrapper := sp.GetAnimeCollectionWrapper()
	if _, err := animeWrapper.FindEntry(mediaID); err == nil {
		return animeWrapper.UpdateEntry(mediaID, status, scoreRaw, progress, startedAt, completedAt)
	}

	// Try manga
	mangaWrapper := sp.GetMangaCollectionWrapper()
	if _, err := mangaWrapper.FindEntry(mediaID); err == nil {
		return mangaWrapper.UpdateEntry(mediaID, status, scoreRaw, progress, startedAt, completedAt)
	}

	// Entry doesn't exist, determine media type and add it
	defaultStatus := anilist.MediaListStatusPlanning
	if status != nil {
		defaultStatus = *status
	}

	// Try to fetch as anime first
	if _, err := sp.client.BaseAnimeByID(context.Background(), &mediaID); err == nil {
		// It's an anime, add it to anime collection
		sp.logger.Trace().Int("mediaID", mediaID).Msg("simulated platform: Adding new anime entry")
		if err := animeWrapper.AddEntry(mediaID, defaultStatus); err != nil {
			return err
		}
		// Update with provided values if there are additional updates needed
		if status != &defaultStatus || scoreRaw != nil || progress != nil || startedAt != nil || completedAt != nil {
			return animeWrapper.UpdateEntry(mediaID, status, scoreRaw, progress, startedAt, completedAt)
		}
		return nil
	}

	// Try to fetch as manga
	if _, err := sp.client.BaseMangaByID(context.Background(), &mediaID); err == nil {
		// It's a manga, add it to manga collection
		sp.logger.Trace().Int("mediaID", mediaID).Msg("simulated platform: Adding new manga entry")
		if err := mangaWrapper.AddEntry(mediaID, defaultStatus); err != nil {
			return err
		}
		// Update with provided values if there are additional updates needed
		if status != &defaultStatus || scoreRaw != nil || progress != nil || startedAt != nil || completedAt != nil {
			return mangaWrapper.UpdateEntry(mediaID, status, scoreRaw, progress, startedAt, completedAt)
		}
		return nil
	}

	// Media not found in either anime or manga
	return errors.New("media not found on AniList")
}

func (sp *SimulatedPlatform) UpdateEntryProgress(mediaID int, progress int, totalEpisodes *int) error {
	sp.logger.Trace().Int("mediaID", mediaID).Int("progress", progress).Msg("simulated platform: Updating entry progress")

	sp.mu.Lock()
	defer sp.mu.Unlock()

	status := anilist.MediaListStatusCurrent
	if totalEpisodes != nil && progress >= *totalEpisodes {
		status = anilist.MediaListStatusCompleted
	}

	// Try anime first
	animeWrapper := sp.GetAnimeCollectionWrapper()
	if _, err := animeWrapper.FindEntry(mediaID); err == nil {
		return animeWrapper.UpdateEntryProgress(mediaID, progress, totalEpisodes)
	}

	// Try manga
	mangaWrapper := sp.GetMangaCollectionWrapper()
	if _, err := mangaWrapper.FindEntry(mediaID); err == nil {
		return mangaWrapper.UpdateEntryProgress(mediaID, progress, totalEpisodes)
	}

	// Entry doesn't exist, determine media type and add it
	// Try to fetch as anime first
	if _, err := sp.client.BaseAnimeByID(context.Background(), &mediaID); err == nil {
		// It's an anime, add it to anime collection
		sp.logger.Trace().Int("mediaID", mediaID).Msg("simulated platform: Adding new anime entry for progress update")
		if err := animeWrapper.AddEntry(mediaID, status); err != nil {
			return err
		}
		return animeWrapper.UpdateEntryProgress(mediaID, progress, totalEpisodes)
	}

	// Try to fetch as manga
	if _, err := sp.client.BaseMangaByID(context.Background(), &mediaID); err == nil {
		// It's a manga, add it to manga collection
		sp.logger.Trace().Int("mediaID", mediaID).Msg("simulated platform: Adding new manga entry for progress update")
		if err := mangaWrapper.AddEntry(mediaID, status); err != nil {
			return err
		}
		return mangaWrapper.UpdateEntryProgress(mediaID, progress, totalEpisodes)
	}

	// Media not found in either anime or manga
	return errors.New("media not found on AniList")
}

func (sp *SimulatedPlatform) UpdateEntryRepeat(mediaID int, repeat int) error {
	sp.logger.Trace().Int("mediaID", mediaID).Int("repeat", repeat).Msg("simulated platform: Updating entry repeat")

	sp.mu.Lock()
	defer sp.mu.Unlock()

	// Try anime first
	wrapper := sp.GetAnimeCollectionWrapper()
	if entry, err := wrapper.FindEntry(mediaID); err == nil {
		if animeEntry, ok := entry.(*anilist.AnimeCollection_MediaListCollection_Lists_Entries); ok {
			animeEntry.Repeat = &repeat
			sp.localManager.SaveSimulatedAnimeCollection(sp.animeCollection)
			return nil
		}
	}

	// Try manga
	wrapper = sp.GetMangaCollectionWrapper()
	if entry, err := wrapper.FindEntry(mediaID); err == nil {
		if mangaEntry, ok := entry.(*anilist.MangaCollection_MediaListCollection_Lists_Entries); ok {
			mangaEntry.Repeat = &repeat
			sp.localManager.SaveSimulatedMangaCollection(sp.mangaCollection)
			return nil
		}
	}

	return ErrMediaNotFound
}

func (sp *SimulatedPlatform) DeleteEntry(entryId int) error {
	sp.logger.Trace().Int("entryId", entryId).Msg("simulated platform: Deleting entry")

	sp.mu.Lock()
	defer sp.mu.Unlock()

	// Try anime first
	wrapper := sp.GetAnimeCollectionWrapper()
	if _, err := wrapper.FindEntry(entryId, true); err == nil {
		return wrapper.DeleteEntry(entryId, true)
	}

	// Try manga
	wrapper = sp.GetMangaCollectionWrapper()
	if _, err := wrapper.FindEntry(entryId, true); err == nil {
		return wrapper.DeleteEntry(entryId, true)
	}

	return ErrMediaNotFound
}

func (sp *SimulatedPlatform) GetAnime(mediaID int) (*anilist.BaseAnime, error) {
	sp.logger.Trace().Int("mediaID", mediaID).Msg("simulated platform: Getting anime")

	// Get anime from anilist
	resp, err := sp.client.BaseAnimeByID(context.Background(), &mediaID)
	if err != nil {
		return nil, err
	}

	// Update media data in collection if it exists
	sp.mu.Lock()
	wrapper := sp.GetAnimeCollectionWrapper()
	if _, err := wrapper.FindEntry(mediaID); err == nil {
		_ = wrapper.UpdateMediaData(mediaID, resp.GetMedia())
	}
	sp.mu.Unlock()

	return resp.GetMedia(), nil
}

func (sp *SimulatedPlatform) GetAnimeByMalID(malID int) (*anilist.BaseAnime, error) {
	sp.logger.Trace().Int("malID", malID).Msg("simulated platform: Getting anime by MAL ID")

	resp, err := sp.client.BaseAnimeByMalID(context.Background(), &malID)
	if err != nil {
		return nil, err
	}

	// Update media data in collection if it exists
	if resp.GetMedia() != nil {
		sp.mu.Lock()
		wrapper := sp.GetAnimeCollectionWrapper()
		if _, err := wrapper.FindEntry(resp.GetMedia().GetID()); err == nil {
			_ = wrapper.UpdateMediaData(resp.GetMedia().GetID(), resp.GetMedia())
		}
		sp.mu.Unlock()
	}

	return resp.GetMedia(), nil
}

func (sp *SimulatedPlatform) GetAnimeDetails(mediaID int) (*anilist.AnimeDetailsById_Media, error) {
	sp.logger.Trace().Int("mediaID", mediaID).Msg("simulated platform: Getting anime details")

	resp, err := sp.client.AnimeDetailsByID(context.Background(), &mediaID)
	if err != nil {
		return nil, err
	}

	return resp.GetMedia(), nil
}

func (sp *SimulatedPlatform) GetAnimeWithRelations(mediaID int) (*anilist.CompleteAnime, error) {
	sp.logger.Trace().Int("mediaID", mediaID).Msg("simulated platform: Getting anime with relations")

	resp, err := sp.client.CompleteAnimeByID(context.Background(), &mediaID)
	if err != nil {
		return nil, err
	}

	return resp.GetMedia(), nil
}

func (sp *SimulatedPlatform) GetManga(mediaID int) (*anilist.BaseManga, error) {
	sp.logger.Trace().Int("mediaID", mediaID).Msg("simulated platform: Getting manga")

	// Get manga from anilist
	resp, err := sp.client.BaseMangaByID(context.Background(), &mediaID)
	if err != nil {
		return nil, err
	}

	// Update media data in collection if it exists
	sp.mu.Lock()
	wrapper := sp.GetMangaCollectionWrapper()
	if _, err := wrapper.FindEntry(mediaID); err == nil {
		_ = wrapper.UpdateMediaData(mediaID, resp.GetMedia())
	}
	sp.mu.Unlock()

	return resp.GetMedia(), nil
}

func (sp *SimulatedPlatform) GetMangaDetails(mediaID int) (*anilist.MangaDetailsById_Media, error) {
	sp.logger.Trace().Int("mediaID", mediaID).Msg("simulated platform: Getting manga details")

	resp, err := sp.client.MangaDetailsByID(context.Background(), &mediaID)
	if err != nil {
		return nil, err
	}

	return resp.GetMedia(), nil
}

func (sp *SimulatedPlatform) GetAnimeCollection(bypassCache bool) (*anilist.AnimeCollection, error) {
	sp.logger.Trace().Bool("bypassCache", bypassCache).Msg("simulated platform: Getting anime collection")

	if bypassCache {
		sp.invalidateAnimeCollectionCache()
	}

	return sp.getOrCreateAnimeCollection()
}

func (sp *SimulatedPlatform) GetRawAnimeCollection(bypassCache bool) (*anilist.AnimeCollection, error) {
	return sp.GetAnimeCollection(bypassCache)
}

func (sp *SimulatedPlatform) RefreshAnimeCollection() (*anilist.AnimeCollection, error) {
	sp.logger.Trace().Msg("simulated platform: Refreshing anime collection")

	collection, err := sp.getOrCreateAnimeCollection()
	if err != nil {
		return nil, err
	}

	anilistRateLimit := limiter.NewAnilistLimiter()
	wg := sync.WaitGroup{}
	m := sync.Mutex{}

	// Refresh all media data in the collection
	for _, list := range collection.GetMediaListCollection().GetLists() {
		for _, entry := range list.GetEntries() {
			if entry.GetMedia() != nil {
				mediaID := entry.GetMedia().GetID()
				wg.Add(1)
				go func(mID int, e *anilist.AnimeCollection_MediaListCollection_Lists_Entries) {
					defer wg.Done()
					anilistRateLimit.Wait()
					if updatedMedia, err := sp.GetAnime(mID); err == nil {
						m.Lock()
						e.Media = updatedMedia
						m.Unlock()
					}
				}(mediaID, entry)
			}
		}
	}

	wg.Wait()

	// Save updated collection
	sp.localManager.SaveSimulatedAnimeCollection(collection)

	sp.invalidateAnimeCollectionCache()
	return sp.getOrCreateAnimeCollection()
}

func (sp *SimulatedPlatform) GetAnimeCollectionWithRelations() (*anilist.AnimeCollectionWithRelations, error) {
	sp.logger.Trace().Msg("simulated platform: Getting anime collection with relations")

	// For simulated platform, we return empty collection with relations
	return &anilist.AnimeCollectionWithRelations{
		MediaListCollection: &anilist.AnimeCollectionWithRelations_MediaListCollection{
			Lists: []*anilist.AnimeCollectionWithRelations_MediaListCollection_Lists{},
		},
	}, nil
}

func (sp *SimulatedPlatform) GetMangaCollection(bypassCache bool) (*anilist.MangaCollection, error) {
	sp.logger.Trace().Bool("bypassCache", bypassCache).Msg("simulated platform: Getting manga collection")

	if bypassCache {
		sp.invalidateMangaCollectionCache()
	}

	return sp.getOrCreateMangaCollection()
}

func (sp *SimulatedPlatform) GetRawMangaCollection(bypassCache bool) (*anilist.MangaCollection, error) {
	return sp.GetMangaCollection(bypassCache)
}

func (sp *SimulatedPlatform) RefreshMangaCollection() (*anilist.MangaCollection, error) {
	sp.logger.Trace().Msg("simulated platform: Refreshing manga collection")

	collection, err := sp.getOrCreateMangaCollection()
	if err != nil {
		return nil, err
	}

	anilistRateLimit := limiter.NewAnilistLimiter()
	wg := sync.WaitGroup{}
	m := sync.Mutex{}

	// Refresh all media data in the collection
	for _, list := range collection.GetMediaListCollection().GetLists() {
		for _, entry := range list.GetEntries() {
			if entry.GetMedia() != nil {
				mediaID := entry.GetMedia().GetID()
				wg.Add(1)
				go func(mID int, e *anilist.MangaCollection_MediaListCollection_Lists_Entries) {
					defer wg.Done()
					anilistRateLimit.Wait()
					if updatedMedia, err := sp.GetManga(mID); err == nil {
						m.Lock()
						e.Media = updatedMedia
						m.Unlock()
					}
				}(mediaID, entry)
			}
		}
	}

	wg.Wait()

	// Save updated collection
	sp.localManager.SaveSimulatedMangaCollection(collection)

	sp.invalidateMangaCollectionCache()
	return sp.getOrCreateMangaCollection()
}

func (sp *SimulatedPlatform) AddMediaToCollection(mIds []int) error {
	sp.logger.Trace().Interface("mediaIDs", mIds).Msg("simulated platform: Adding media to collection")

	sp.mu.Lock()
	defer sp.mu.Unlock()

	// DEVNOTE: We assume it's anime for now since it's only been used for anime
	wrapper := sp.GetAnimeCollectionWrapper()
	for _, mediaID := range mIds {
		// Try to add as anime first, if it fails, ignore
		_ = wrapper.AddEntry(mediaID, anilist.MediaListStatusPlanning)
	}

	return nil
}

func (sp *SimulatedPlatform) GetStudioDetails(studioID int) (*anilist.StudioDetails, error) {
	return sp.client.StudioDetails(context.Background(), &studioID)
}

func (sp *SimulatedPlatform) GetAnilistClient() anilist.AnilistClient {
	return sp.client
}

func (sp *SimulatedPlatform) GetViewerStats() (*anilist.ViewerStats, error) {
	return nil, errors.New("use a real account to get stats")
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Helper Methods
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (sp *SimulatedPlatform) getOrCreateAnimeCollection() (*anilist.AnimeCollection, error) {
	sp.collectionMu.RLock()
	if sp.animeCollection != nil {
		defer sp.collectionMu.RUnlock()
		return sp.animeCollection, nil
	}
	sp.collectionMu.RUnlock()

	sp.collectionMu.Lock()
	defer sp.collectionMu.Unlock()

	// Double-check after acquiring write lock
	if sp.animeCollection != nil {
		return sp.animeCollection, nil
	}

	// Try to load from database
	if collection := sp.localManager.GetSimulatedAnimeCollection(); collection.IsPresent() {
		sp.animeCollection = collection.MustGet()
		return sp.animeCollection, nil
	}

	// Create empty collection
	sp.animeCollection = &anilist.AnimeCollection{
		MediaListCollection: &anilist.AnimeCollection_MediaListCollection{
			Lists: []*anilist.AnimeCollection_MediaListCollection_Lists{},
		},
	}

	// Save empty collection
	sp.localManager.SaveSimulatedAnimeCollection(sp.animeCollection)

	return sp.animeCollection, nil
}

func (sp *SimulatedPlatform) getOrCreateMangaCollection() (*anilist.MangaCollection, error) {
	sp.collectionMu.RLock()
	if sp.mangaCollection != nil {
		defer sp.collectionMu.RUnlock()
		return sp.mangaCollection, nil
	}
	sp.collectionMu.RUnlock()

	sp.collectionMu.Lock()
	defer sp.collectionMu.Unlock()

	// Double-check after acquiring write lock
	if sp.mangaCollection != nil {
		return sp.mangaCollection, nil
	}

	// Try to load from database
	if collection := sp.localManager.GetSimulatedMangaCollection(); collection.IsPresent() {
		sp.mangaCollection = collection.MustGet()
		return sp.mangaCollection, nil
	}

	// Create empty collection
	sp.mangaCollection = &anilist.MangaCollection{
		MediaListCollection: &anilist.MangaCollection_MediaListCollection{
			Lists: []*anilist.MangaCollection_MediaListCollection_Lists{},
		},
	}

	// Save empty collection
	sp.localManager.SaveSimulatedMangaCollection(sp.mangaCollection)

	return sp.mangaCollection, nil
}

func (sp *SimulatedPlatform) invalidateAnimeCollectionCache() {
	sp.collectionMu.Lock()
	defer sp.collectionMu.Unlock()
	sp.animeCollection = nil
}

func (sp *SimulatedPlatform) invalidateMangaCollectionCache() {
	sp.collectionMu.Lock()
	defer sp.collectionMu.Unlock()
	sp.mangaCollection = nil
}
