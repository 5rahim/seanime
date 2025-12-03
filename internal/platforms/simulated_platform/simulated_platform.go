package simulated_platform

import (
	"context"
	"encoding/json"
	"errors"
	"seanime/internal/api/anilist"
	"seanime/internal/database/db"
	"seanime/internal/extension"
	"seanime/internal/hook"
	"seanime/internal/local"
	"seanime/internal/platforms/platform"
	"seanime/internal/platforms/shared_platform"
	"seanime/internal/util"
	"seanime/internal/util/limiter"
	"sync"
	"time"

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
	animeCollection                *anilist.AnimeCollection
	mangaCollection                *anilist.MangaCollection
	mu                             sync.RWMutex
	collectionMu                   sync.RWMutex // used to protect access to collections
	lastAnimeCollectionRefetchTime time.Time    // used to prevent refetching too many times
	lastMangaCollectionRefetchTime time.Time    // used to prevent refetching too many times
	anilistRateLimit               *limiter.Limiter
	helper                         *shared_platform.PlatformHelper
	db                             *db.Database
}

func NewSimulatedPlatform(localManager local.Manager, client *util.Ref[anilist.AnilistClient], extensionBankRef *util.Ref[*extension.UnifiedBank], logger *zerolog.Logger, db *db.Database) (platform.Platform, error) {
	sp := &SimulatedPlatform{
		logger:           logger,
		localManager:     localManager,
		client:           shared_platform.NewCacheLayer(client),
		anilistRateLimit: limiter.NewAnilistLimiter(),
		helper:           shared_platform.NewPlatformHelper(extensionBankRef, db, logger),
		db:               db,
	}

	return sp, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Implementation
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (sp *SimulatedPlatform) SetUsername(username string) {
	// no-op
}

func (sp *SimulatedPlatform) Close() {
	sp.helper.Close()
}

func (sp *SimulatedPlatform) ClearCache() {
	sp.helper.ClearCache()
}

// UpdateEntry updates the entry for the given media ID.
// If the entry doesn't exist, it will be added automatically after determining the media type.
func (sp *SimulatedPlatform) UpdateEntry(ctx context.Context, mediaID int, status *anilist.MediaListStatus, scoreRaw *int, progress *int, startedAt *anilist.FuzzyDateInput, completedAt *anilist.FuzzyDateInput) error {
	sp.logger.Trace().Int("mediaID", mediaID).Msg("simulated platform: Updating entry")

	return sp.helper.TriggerUpdateEntryHooks(ctx, mediaID, status, scoreRaw, progress, startedAt, completedAt, func(event *platform.PreUpdateEntryEvent) error {
		// Check if this is a custom source entry (after hooks have been triggered)
		if handled, err := sp.helper.HandleCustomSourceUpdateEntry(ctx, mediaID, event.Status, event.ScoreRaw, event.Progress, event.StartedAt, event.CompletedAt); handled {
			return err
		}

		sp.mu.Lock()
		defer sp.mu.Unlock()

		// Try anime first
		animeWrapper := sp.GetAnimeCollectionWrapper()
		if _, err := animeWrapper.FindEntry(mediaID); err == nil {
			return animeWrapper.UpdateEntry(mediaID, event.Status, event.ScoreRaw, event.Progress, event.StartedAt, event.CompletedAt)
		}

		// Try manga
		mangaWrapper := sp.GetMangaCollectionWrapper()
		if _, err := mangaWrapper.FindEntry(mediaID); err == nil {
			return mangaWrapper.UpdateEntry(mediaID, event.Status, event.ScoreRaw, event.Progress, event.StartedAt, event.CompletedAt)
		}

		// Entry doesn't exist, determine media type and add it
		defaultStatus := anilist.MediaListStatusPlanning
		if event.Status != nil {
			defaultStatus = *event.Status
		}

		// Try to fetch as anime first
		if _, err := sp.client.BaseAnimeByID(ctx, &mediaID); err == nil {
			// It's an anime, add it to anime collection
			sp.logger.Trace().Int("mediaID", mediaID).Msg("simulated platform: Adding new anime entry")
			if err := animeWrapper.AddEntry(mediaID, defaultStatus); err != nil {
				return err
			}
			// Update with provided values if there are additional updates needed
			if event.Status != &defaultStatus || event.ScoreRaw != nil || event.Progress != nil || event.StartedAt != nil || event.CompletedAt != nil {
				return animeWrapper.UpdateEntry(mediaID, event.Status, event.ScoreRaw, event.Progress, event.StartedAt, event.CompletedAt)
			}
			return nil
		}

		// Try to fetch as manga
		if _, err := sp.client.BaseMangaByID(ctx, &mediaID); err == nil {
			// It's a manga, add it to manga collection
			sp.logger.Trace().Int("mediaID", mediaID).Msg("simulated platform: Adding new manga entry")
			if err := mangaWrapper.AddEntry(mediaID, defaultStatus); err != nil {
				return err
			}
			// Update with provided values if there are additional updates needed
			if event.Status != &defaultStatus || event.ScoreRaw != nil || event.Progress != nil || event.StartedAt != nil || event.CompletedAt != nil {
				return mangaWrapper.UpdateEntry(mediaID, event.Status, event.ScoreRaw, event.Progress, event.StartedAt, event.CompletedAt)
			}
			return nil
		}

		// Media not found in either anime or manga
		return errors.New("media not found on AniList")
	})
}

func (sp *SimulatedPlatform) UpdateEntryProgress(ctx context.Context, mediaID int, progress int, totalEpisodes *int) error {
	sp.logger.Trace().Int("mediaID", mediaID).Int("progress", progress).Msg("simulated platform: Updating entry progress")

	return sp.helper.TriggerUpdateEntryProgressHooks(ctx, mediaID, progress, totalEpisodes, func(event *platform.PreUpdateEntryProgressEvent) error {
		// Check if this is a custom source entry (after hooks have been triggered)
		if handled, err := sp.helper.HandleCustomSourceUpdateEntryProgress(ctx, mediaID, *event.Progress, event.TotalCount); handled {
			return err
		}

		sp.mu.Lock()
		defer sp.mu.Unlock()

		status := anilist.MediaListStatusCurrent
		if event.TotalCount != nil && *event.Progress >= *event.TotalCount {
			status = anilist.MediaListStatusCompleted
			*event.Status = status
		}

		// Try anime first
		animeWrapper := sp.GetAnimeCollectionWrapper()
		if _, err := animeWrapper.FindEntry(mediaID); err == nil {
			return animeWrapper.UpdateEntryProgress(mediaID, *event.Progress, event.TotalCount)
		}

		// Try manga
		mangaWrapper := sp.GetMangaCollectionWrapper()
		if _, err := mangaWrapper.FindEntry(mediaID); err == nil {
			return mangaWrapper.UpdateEntryProgress(mediaID, *event.Progress, event.TotalCount)
		}

		// Entry doesn't exist, determine media type and add it
		// Try to fetch as anime first
		if _, err := sp.client.BaseAnimeByID(ctx, &mediaID); err == nil {
			// It's an anime, add it to anime collection
			sp.logger.Trace().Int("mediaID", mediaID).Msg("simulated platform: Adding new anime entry for progress update")
			if err := animeWrapper.AddEntry(mediaID, status); err != nil {
				return err
			}
			return animeWrapper.UpdateEntryProgress(mediaID, *event.Progress, event.TotalCount)
		}

		// Try to fetch as manga
		if _, err := sp.client.BaseMangaByID(ctx, &mediaID); err == nil {
			// It's a manga, add it to manga collection
			sp.logger.Trace().Int("mediaID", mediaID).Msg("simulated platform: Adding new manga entry for progress update")
			if err := mangaWrapper.AddEntry(mediaID, status); err != nil {
				return err
			}
			return mangaWrapper.UpdateEntryProgress(mediaID, *event.Progress, event.TotalCount)
		}

		// Media not found in either anime or manga
		return errors.New("media not found on AniList")
	})
}

func (sp *SimulatedPlatform) UpdateEntryRepeat(ctx context.Context, mediaID int, repeat int) error {
	sp.logger.Trace().Int("mediaID", mediaID).Int("repeat", repeat).Msg("simulated platform: Updating entry repeat")

	return sp.helper.TriggerUpdateEntryRepeatHooks(ctx, mediaID, repeat, func(event *platform.PreUpdateEntryRepeatEvent) error {
		// Check if this is a custom source entry (after hooks have been triggered)
		if handled, err := sp.helper.HandleCustomSourceUpdateEntryRepeat(ctx, mediaID, *event.Repeat); handled {
			return err
		}

		sp.mu.Lock()
		defer sp.mu.Unlock()

		// Try anime first
		wrapper := sp.GetAnimeCollectionWrapper()
		if entry, err := wrapper.FindEntry(mediaID); err == nil {
			if animeEntry, ok := entry.(*anilist.AnimeCollection_MediaListCollection_Lists_Entries); ok {
				animeEntry.Repeat = event.Repeat
				sp.localManager.SaveSimulatedAnimeCollection(sp.animeCollection)
				return nil
			}
		}

		// Try manga
		wrapper = sp.GetMangaCollectionWrapper()
		if entry, err := wrapper.FindEntry(mediaID); err == nil {
			if mangaEntry, ok := entry.(*anilist.MangaCollection_MediaListCollection_Lists_Entries); ok {
				mangaEntry.Repeat = event.Repeat
				sp.localManager.SaveSimulatedMangaCollection(sp.mangaCollection)
				return nil
			}
		}

		return ErrMediaNotFound
	})
}

func (sp *SimulatedPlatform) DeleteEntry(ctx context.Context, mediaId, entryId int) error {
	sp.logger.Trace().Int("entryId", entryId).Int("mediaId", mediaId).Msg("simulated platform: Deleting entry")

	return sp.helper.TriggerDeleteEntryHooks(ctx, mediaId, entryId, func(event *platform.PreDeleteEntryEvent) error {
		if handled, err := sp.helper.HandleCustomSourceDeleteEntry(ctx, *event.MediaID, *event.EntryID); handled {
			return err
		}

		sp.mu.Lock()
		defer sp.mu.Unlock()

		// Try anime first
		wrapper := sp.GetAnimeCollectionWrapper()
		if _, err := wrapper.FindEntry(*event.EntryID, true); err == nil {
			return wrapper.DeleteEntry(*event.EntryID, true)
		}

		// Try manga
		wrapper = sp.GetMangaCollectionWrapper()
		if _, err := wrapper.FindEntry(*event.EntryID, true); err == nil {
			return wrapper.DeleteEntry(*event.EntryID, true)
		}

		return ErrMediaNotFound
	})
}

func (sp *SimulatedPlatform) GetAnime(ctx context.Context, mediaID int) (*anilist.BaseAnime, error) {
	sp.logger.Trace().Int("mediaID", mediaID).Msg("simulated platform: Getting anime")

	if cachedAnime, ok := sp.helper.GetCachedBaseAnime(mediaID); ok {
		sp.logger.Trace().Msg("simulated platform: Returning anime from cache")
		return sp.helper.TriggerGetAnimeEvent(cachedAnime)
	}

	// Check if this is a custom source entry
	if media, isCustom, err := sp.helper.HandleCustomSourceAnime(ctx, mediaID); isCustom {
		if err != nil {
			return nil, err
		}

		triggeredMedia, err := sp.helper.TriggerGetAnimeEvent(media)
		if err != nil {
			return nil, err
		}

		sp.helper.SetCachedBaseAnime(mediaID, triggeredMedia)

		// Update media data in collection if it exists (simulated platform specific)
		sp.mu.Lock()
		wrapper := sp.GetAnimeCollectionWrapper()
		if _, err := wrapper.FindEntry(mediaID); err == nil {
			_ = wrapper.UpdateMediaData(mediaID, triggeredMedia)
		}
		sp.mu.Unlock()

		return triggeredMedia, nil
	}

	// Get anime from anilist
	resp, err := sp.client.BaseAnimeByID(ctx, &mediaID)
	if err != nil {
		return nil, err
	}
	media := resp.GetMedia()

	triggeredMedia, err := sp.helper.TriggerGetAnimeEvent(media)
	if err != nil {
		return nil, err
	}

	sp.helper.SetCachedBaseAnime(mediaID, triggeredMedia)

	// Update media data in collection if it exists (simulated platform specific)
	sp.mu.Lock()
	wrapper := sp.GetAnimeCollectionWrapper()
	if _, err := wrapper.FindEntry(mediaID); err == nil {
		_ = wrapper.UpdateMediaData(mediaID, triggeredMedia)
	}
	sp.mu.Unlock()

	return triggeredMedia, nil
}

func (sp *SimulatedPlatform) GetAnimeByMalID(ctx context.Context, malID int) (*anilist.BaseAnime, error) {
	sp.logger.Trace().Int("malID", malID).Msg("simulated platform: Getting anime by MAL ID")

	resp, err := sp.client.BaseAnimeByMalID(ctx, &malID)
	if err != nil {
		return nil, err
	}

	media := resp.GetMedia()
	triggeredMedia, err := sp.helper.TriggerGetAnimeEvent(media)
	if err != nil {
		return nil, err
	}

	// Update media data in collection if it exists (simulated platform specific)
	if triggeredMedia != nil {
		sp.mu.Lock()
		wrapper := sp.GetAnimeCollectionWrapper()
		if _, err := wrapper.FindEntry(triggeredMedia.GetID()); err == nil {
			_ = wrapper.UpdateMediaData(triggeredMedia.GetID(), triggeredMedia)
		}
		sp.mu.Unlock()
	}

	return triggeredMedia, nil
}

func (sp *SimulatedPlatform) GetAnimeDetails(ctx context.Context, mediaID int) (*anilist.AnimeDetailsById_Media, error) {
	sp.logger.Trace().Int("mediaID", mediaID).Msg("simulated platform: Getting anime details")

	// Check if this is a custom source entry
	if media, isCustom, err := sp.helper.HandleCustomSourceAnimeDetails(ctx, mediaID); isCustom {
		if err != nil {
			return nil, err
		}
		return sp.helper.TriggerGetAnimeDetailsEvent(media)
	}

	// Get from AniList
	resp, err := sp.client.AnimeDetailsByID(ctx, &mediaID)
	if err != nil {
		return nil, err
	}
	media := resp.GetMedia()

	return sp.helper.TriggerGetAnimeDetailsEvent(media)
}

func (sp *SimulatedPlatform) GetAnimeWithRelations(ctx context.Context, mediaID int) (*anilist.CompleteAnime, error) {
	sp.logger.Trace().Int("mediaID", mediaID).Msg("simulated platform: Getting anime with relations")

	if cachedAnime, ok := sp.helper.GetCachedCompleteAnime(mediaID); ok {
		sp.logger.Trace().Msg("simulated platform: Cache HIT for anime with relations")
		return cachedAnime, nil
	}

	// Check if this is a custom source entry
	if media, isCustom, err := sp.helper.HandleCustomSourceAnimeWithRelations(ctx, mediaID); isCustom {
		if err != nil {
			return nil, err
		}
		sp.helper.SetCachedCompleteAnime(mediaID, media)
		return media, nil
	}

	// Get from AniList
	resp, err := sp.client.CompleteAnimeByID(ctx, &mediaID)
	if err != nil {
		return nil, err
	}
	media := resp.GetMedia()

	sp.helper.SetCachedCompleteAnime(mediaID, media)
	return media, nil
}

func (sp *SimulatedPlatform) GetManga(ctx context.Context, mediaID int) (*anilist.BaseManga, error) {
	sp.logger.Trace().Int("mediaID", mediaID).Msg("simulated platform: Getting manga")

	if cachedManga, ok := sp.helper.GetCachedBaseManga(mediaID); ok {
		sp.logger.Trace().Msg("simulated platform: Returning manga from cache")
		return sp.helper.TriggerGetMangaEvent(cachedManga)
	}

	// Check if this is a custom source entry
	if media, isCustom, err := sp.helper.HandleCustomSourceManga(ctx, mediaID); isCustom {
		if err != nil {
			return nil, err
		}

		triggeredMedia, err := sp.helper.TriggerGetMangaEvent(media)
		if err != nil {
			return nil, err
		}

		sp.helper.SetCachedBaseManga(mediaID, triggeredMedia)

		// Update media data in collection if it exists (simulated platform specific)
		sp.mu.Lock()
		wrapper := sp.GetMangaCollectionWrapper()
		if _, err := wrapper.FindEntry(mediaID); err == nil {
			_ = wrapper.UpdateMediaData(mediaID, triggeredMedia)
		}
		sp.mu.Unlock()

		return triggeredMedia, nil
	}

	// Get manga from anilist
	resp, err := sp.client.BaseMangaByID(ctx, &mediaID)
	if err != nil {
		return nil, err
	}
	media := resp.GetMedia()

	triggeredMedia, err := sp.helper.TriggerGetMangaEvent(media)
	if err != nil {
		return nil, err
	}

	sp.helper.SetCachedBaseManga(mediaID, triggeredMedia)

	// Update media data in collection if it exists (simulated platform specific)
	sp.mu.Lock()
	wrapper := sp.GetMangaCollectionWrapper()
	if _, err := wrapper.FindEntry(mediaID); err == nil {
		_ = wrapper.UpdateMediaData(mediaID, triggeredMedia)
	}
	sp.mu.Unlock()

	return triggeredMedia, nil
}

func (sp *SimulatedPlatform) GetMangaDetails(ctx context.Context, mediaID int) (*anilist.MangaDetailsById_Media, error) {
	sp.logger.Trace().Int("mediaID", mediaID).Msg("simulated platform: Getting manga details")

	// Check if this is a custom source entry
	if media, isCustom, err := sp.helper.HandleCustomSourceMangaDetails(ctx, mediaID); isCustom {
		return media, err
	}

	// Get from AniList
	resp, err := sp.client.MangaDetailsByID(ctx, &mediaID)
	if err != nil {
		return nil, err
	}

	return resp.GetMedia(), nil
}

func (sp *SimulatedPlatform) GetAnimeCollection(ctx context.Context, bypassCache bool) (*anilist.AnimeCollection, error) {
	sp.logger.Trace().Bool("bypassCache", bypassCache).Msg("simulated platform: Getting anime collection")

	if !bypassCache && sp.animeCollection != nil {
		event := new(platform.GetCachedAnimeCollectionEvent)
		event.AnimeCollection = sp.animeCollection
		err := hook.GlobalHookManager.OnGetCachedAnimeCollection().Trigger(event)
		if err != nil {
			return nil, err
		}
		return event.AnimeCollection, nil
	}

	if bypassCache {
		sp.invalidateAnimeCollectionCache()
	}

	collection, err := sp.getOrCreateAnimeCollection()
	if err != nil {
		return nil, err
	}

	// Merge custom source entries if available
	sp.helper.MergeCustomSourceAnimeEntries(collection)

	event := new(platform.GetAnimeCollectionEvent)
	event.AnimeCollection = collection

	err = hook.GlobalHookManager.OnGetAnimeCollection().Trigger(event)
	if err != nil {
		return nil, err
	}

	return event.AnimeCollection, nil
}

func (sp *SimulatedPlatform) GetRawAnimeCollection(ctx context.Context, bypassCache bool) (*anilist.AnimeCollection, error) {
	sp.logger.Trace().Bool("bypassCache", bypassCache).Msg("simulated platform: Getting raw anime collection")

	if !bypassCache && sp.animeCollection != nil {
		event := new(platform.GetCachedRawAnimeCollectionEvent)
		event.AnimeCollection = sp.animeCollection
		err := hook.GlobalHookManager.OnGetCachedRawAnimeCollection().Trigger(event)
		if err != nil {
			return nil, err
		}
		return event.AnimeCollection, nil
	}

	if bypassCache {
		sp.invalidateAnimeCollectionCache()
	}

	collection, err := sp.getOrCreateAnimeCollection()
	if err != nil {
		return nil, err
	}

	// Merge custom source entries if available
	sp.helper.MergeCustomSourceAnimeEntries(collection)

	event := new(platform.GetRawAnimeCollectionEvent)
	event.AnimeCollection = collection

	err = hook.GlobalHookManager.OnGetRawAnimeCollection().Trigger(event)
	if err != nil {
		return nil, err
	}

	return event.AnimeCollection, nil
}

func (sp *SimulatedPlatform) RefreshAnimeCollection(ctx context.Context) (*anilist.AnimeCollection, error) {
	sp.logger.Trace().Msg("simulated platform: Refreshing anime collection")

	sp.invalidateAnimeCollectionCache()
	collection, err := sp.getOrCreateAnimeCollection()
	if err != nil {
		return nil, err
	}

	// Merge custom source entries if available
	sp.helper.MergeCustomSourceAnimeEntries(collection)

	event := new(platform.GetAnimeCollectionEvent)
	event.AnimeCollection = collection

	err = hook.GlobalHookManager.OnGetAnimeCollection().Trigger(event)
	if err != nil {
		return nil, err
	}

	event2 := new(platform.GetRawAnimeCollectionEvent)
	event2.AnimeCollection = collection

	err = hook.GlobalHookManager.OnGetRawAnimeCollection().Trigger(event2)
	if err != nil {
		return nil, err
	}

	return event.AnimeCollection, nil
}

// GetAnimeCollectionWithRelations returns the anime collection (without relations)
func (sp *SimulatedPlatform) GetAnimeCollectionWithRelations(ctx context.Context) (*anilist.AnimeCollectionWithRelations, error) {
	sp.logger.Trace().Msg("simulated platform: Getting anime collection with relations")

	// Use JSON to convert the collection structs
	collection, err := sp.getOrCreateAnimeCollection()
	if err != nil {
		return nil, err
	}

	collectionWithRelations := &anilist.AnimeCollectionWithRelations{}

	marshaled, err := json.Marshal(collection)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(marshaled, collectionWithRelations)
	if err != nil {
		return nil, err
	}

	// For simulated platform, the anime collection will not have relations
	return collectionWithRelations, nil
}

func (sp *SimulatedPlatform) GetMangaCollection(ctx context.Context, bypassCache bool) (*anilist.MangaCollection, error) {
	sp.logger.Trace().Bool("bypassCache", bypassCache).Msg("simulated platform: Getting manga collection")

	if !bypassCache && sp.mangaCollection != nil {
		event := new(platform.GetCachedMangaCollectionEvent)
		event.MangaCollection = sp.mangaCollection
		err := hook.GlobalHookManager.OnGetCachedMangaCollection().Trigger(event)
		if err != nil {
			return nil, err
		}
		return event.MangaCollection, nil
	}

	if bypassCache {
		sp.invalidateMangaCollectionCache()
	}

	collection, err := sp.getOrCreateMangaCollection()
	if err != nil {
		return nil, err
	}

	// Merge custom source entries if available
	sp.helper.MergeCustomSourceMangaEntries(collection)

	event := new(platform.GetMangaCollectionEvent)
	event.MangaCollection = collection

	err = hook.GlobalHookManager.OnGetMangaCollection().Trigger(event)
	if err != nil {
		return nil, err
	}

	return event.MangaCollection, nil
}

func (sp *SimulatedPlatform) GetRawMangaCollection(ctx context.Context, bypassCache bool) (*anilist.MangaCollection, error) {
	sp.logger.Trace().Bool("bypassCache", bypassCache).Msg("simulated platform: Getting raw manga collection")

	if !bypassCache && sp.mangaCollection != nil {
		event := new(platform.GetCachedRawMangaCollectionEvent)
		event.MangaCollection = sp.mangaCollection
		err := hook.GlobalHookManager.OnGetCachedRawMangaCollection().Trigger(event)
		if err != nil {
			return nil, err
		}
		return event.MangaCollection, nil
	}

	if bypassCache {
		sp.invalidateMangaCollectionCache()
	}

	collection, err := sp.getOrCreateMangaCollection()
	if err != nil {
		return nil, err
	}

	// Merge custom source entries if available
	sp.helper.MergeCustomSourceMangaEntries(collection)

	event := new(platform.GetRawMangaCollectionEvent)
	event.MangaCollection = collection

	err = hook.GlobalHookManager.OnGetRawMangaCollection().Trigger(event)
	if err != nil {
		return nil, err
	}

	return event.MangaCollection, nil
}

func (sp *SimulatedPlatform) RefreshMangaCollection(ctx context.Context) (*anilist.MangaCollection, error) {
	sp.logger.Trace().Msg("simulated platform: Refreshing manga collection")

	sp.invalidateMangaCollectionCache()
	collection, err := sp.getOrCreateMangaCollection()
	if err != nil {
		return nil, err
	}

	// Merge custom source entries if available
	sp.helper.MergeCustomSourceMangaEntries(collection)

	event := new(platform.GetMangaCollectionEvent)
	event.MangaCollection = collection

	err = hook.GlobalHookManager.OnGetMangaCollection().Trigger(event)
	if err != nil {
		return nil, err
	}

	event2 := new(platform.GetRawMangaCollectionEvent)
	event2.MangaCollection = collection

	err = hook.GlobalHookManager.OnGetRawMangaCollection().Trigger(event2)
	if err != nil {
		return nil, err
	}

	return event.MangaCollection, nil
}

func (sp *SimulatedPlatform) AddMediaToCollection(ctx context.Context, mIds []int) error {
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

func (sp *SimulatedPlatform) GetStudioDetails(ctx context.Context, studioID int) (*anilist.StudioDetails, error) {
	sp.logger.Trace().Int("studioID", studioID).Msg("simulated platform: Getting studio details")

	ret, err := sp.client.StudioDetails(ctx, &studioID)
	if err != nil {
		return nil, err
	}

	return sp.helper.TriggerGetStudioDetailsEvent(ret)
}

func (sp *SimulatedPlatform) GetAnilistClient() anilist.AnilistClient {
	return sp.client
}

func (sp *SimulatedPlatform) GetViewerStats(ctx context.Context) (*anilist.ViewerStats, error) {
	return nil, errors.New("use a real account to get stats")
}

func (sp *SimulatedPlatform) GetAnimeAiringSchedule(ctx context.Context) (*anilist.AnimeAiringSchedule, error) {
	collection, err := sp.GetAnimeCollection(ctx, false)
	if err != nil {
		return nil, err
	}

	return sp.helper.BuildAnimeAiringSchedule(ctx, collection, sp.client)
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
