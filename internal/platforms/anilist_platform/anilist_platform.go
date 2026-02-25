package anilist_platform

import (
	"context"
	"errors"
	"seanime/internal/api/anilist"
	"seanime/internal/customsource"
	"seanime/internal/database/db"
	"seanime/internal/extension"
	"seanime/internal/hook"
	"seanime/internal/platforms/platform"
	"seanime/internal/platforms/shared_platform"
	"seanime/internal/util"
	"seanime/internal/util/limiter"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/mo"
)

type (
	AnilistPlatform struct {
		logger                 *zerolog.Logger
		username               mo.Option[string]
		anilistClient          anilist.AnilistClient
		animeCollection        mo.Option[*anilist.AnimeCollection]
		rawAnimeCollection     mo.Option[*anilist.AnimeCollection]
		mangaCollection        mo.Option[*anilist.MangaCollection]
		rawMangaCollection     mo.Option[*anilist.MangaCollection]
		isOffline              bool
		offlinePlatformEnabled bool
		helper                 *shared_platform.PlatformHelper
		db                     *db.Database
		extensionBankRef       *util.Ref[*extension.UnifiedBank]
	}
)

func NewAnilistPlatform(anilistClientRef *util.Ref[anilist.AnilistClient], extensionBankRef *util.Ref[*extension.UnifiedBank], logger *zerolog.Logger, db *db.Database, logoutFunc ...func()) platform.Platform {
	ap := &AnilistPlatform{
		anilistClient:      shared_platform.NewCacheLayer(anilistClientRef, logoutFunc...),
		logger:             logger,
		username:           mo.None[string](),
		animeCollection:    mo.None[*anilist.AnimeCollection](),
		rawAnimeCollection: mo.None[*anilist.AnimeCollection](),
		mangaCollection:    mo.None[*anilist.MangaCollection](),
		rawMangaCollection: mo.None[*anilist.MangaCollection](),
		extensionBankRef:   extensionBankRef,
		helper:             shared_platform.NewPlatformHelper(extensionBankRef, db, logger),
		db:                 db,
	}

	return ap
}

func (ap *AnilistPlatform) ClearCache() {
	ap.helper.ClearCache()
}

func (ap *AnilistPlatform) Close() {
	ap.helper.Close()
}

func (ap *AnilistPlatform) SetUsername(username string) {
	// Set the username for the AnilistPlatform
	if username == "" {
		ap.username = mo.Some[string]("")
		return
	}

	ap.username = mo.Some(username)
}

func (ap *AnilistPlatform) SetAnilistClient(client anilist.AnilistClient) {
	// Set the AnilistClient for the AnilistPlatform
	ap.anilistClient = client
}

func (ap *AnilistPlatform) UpdateEntry(ctx context.Context, mediaID int, status *anilist.MediaListStatus, scoreRaw *int, progress *int, startedAt *anilist.FuzzyDateInput, completedAt *anilist.FuzzyDateInput) error {
	ap.logger.Trace().Msg("anilist platform: Updating entry")

	// Use shared hook handling
	return ap.helper.TriggerUpdateEntryHooks(ctx, mediaID, status, scoreRaw, progress, startedAt, completedAt, func(event *platform.PreUpdateEntryEvent) error {
		// Check if this is a custom source entry (after hooks have been triggered)
		if handled, err := ap.helper.HandleCustomSourceUpdateEntry(ctx, mediaID, event.Status, event.ScoreRaw, event.Progress, event.StartedAt, event.CompletedAt); handled {
			return err
		}

		_, err := ap.anilistClient.UpdateMediaListEntry(ctx, event.MediaID, event.Status, event.ScoreRaw, event.Progress, event.StartedAt, event.CompletedAt)
		return err
	})
}

func (ap *AnilistPlatform) UpdateEntryProgress(ctx context.Context, mediaID int, progress int, totalCount *int) error {
	ap.logger.Trace().Msg("anilist platform: Updating entry progress")

	// Use shared hook handling
	return ap.helper.TriggerUpdateEntryProgressHooks(ctx, mediaID, progress, totalCount, func(event *platform.PreUpdateEntryProgressEvent) error {
		// Check if this is a custom source entry (after hooks have been triggered)
		if handled, err := ap.helper.HandleCustomSourceUpdateEntryProgress(ctx, mediaID, *event.Progress, event.TotalCount); handled {
			return err
		}

		realTotalCount := 0
		if totalCount != nil && *totalCount > 0 {
			realTotalCount = *totalCount
		}

		// Check if the anime is in the repeating list
		// If it is, set the status to repeating
		if ap.rawAnimeCollection.IsPresent() {
			for _, list := range ap.rawAnimeCollection.MustGet().MediaListCollection.Lists {
				if list.Status != nil && *list.Status == anilist.MediaListStatusRepeating {
					if list.Entries != nil {
						for _, entry := range list.Entries {
							if entry.GetMedia().GetID() == mediaID {
								*event.Status = anilist.MediaListStatusRepeating
								break
							}
						}
					}
				}
			}
		}
		if realTotalCount > 0 && *event.Progress >= realTotalCount {
			*event.Status = anilist.MediaListStatusCompleted
		}

		if realTotalCount > 0 && *event.Progress > realTotalCount {
			*event.Progress = realTotalCount
		}

		_, err := ap.anilistClient.UpdateMediaListEntryProgress(
			ctx,
			event.MediaID,
			event.Progress,
			event.Status,
		)
		return err
	})
}

func (ap *AnilistPlatform) UpdateEntryRepeat(ctx context.Context, mediaID int, repeat int) error {
	ap.logger.Trace().Msg("anilist platform: Updating entry repeat")

	// Use shared hook handling
	return ap.helper.TriggerUpdateEntryRepeatHooks(ctx, mediaID, repeat, func(event *platform.PreUpdateEntryRepeatEvent) error {
		// Check if this is a custom source entry (after hooks have been triggered)
		if handled, err := ap.helper.HandleCustomSourceUpdateEntryRepeat(ctx, mediaID, *event.Repeat); handled {
			return err
		}

		_, err := ap.anilistClient.UpdateMediaListEntryRepeat(ctx, event.MediaID, event.Repeat)
		return err
	})
}

func (ap *AnilistPlatform) DeleteEntry(ctx context.Context, mediaID, entryId int) error {
	ap.logger.Trace().Msg("anilist platform: Deleting entry")

	return ap.helper.TriggerDeleteEntryHooks(ctx, mediaID, entryId, func(event *platform.PreDeleteEntryEvent) error {
		if handled, err := ap.helper.HandleCustomSourceDeleteEntry(ctx, *event.MediaID, *event.EntryID); handled {
			return err
		}

		_, err := ap.anilistClient.DeleteEntry(ctx, event.EntryID)
		if err != nil {
			return err
		}
		return nil
	})
}

func (ap *AnilistPlatform) GetAnime(ctx context.Context, mediaID int) (*anilist.BaseAnime, error) {
	ap.logger.Trace().Int("mediaId", mediaID).Msg("anilist platform: Fetching anime")

	if cachedAnime, ok := ap.helper.GetCachedBaseAnime(mediaID); ok {
		ap.logger.Trace().Msg("anilist platform: Returning anime from cache")
		return ap.helper.TriggerGetAnimeEvent(cachedAnime)
	}

	// Check if this is a custom source entry
	if media, isCustom, err := ap.helper.HandleCustomSourceAnime(ctx, mediaID); isCustom {
		if err != nil {
			return nil, err
		}

		triggeredMedia, err := ap.helper.TriggerGetAnimeEvent(media)
		if err != nil {
			return nil, err
		}

		ap.helper.SetCachedBaseAnime(mediaID, triggeredMedia)
		return triggeredMedia, nil
	}

	// Get from AniList
	ret, err := ap.anilistClient.BaseAnimeByID(ctx, &mediaID)
	if err != nil {
		return nil, err
	}

	media := ret.GetMedia()
	triggeredMedia, err := ap.helper.TriggerGetAnimeEvent(media)
	if err != nil {
		return nil, err
	}

	ap.helper.SetCachedBaseAnime(mediaID, triggeredMedia)
	return triggeredMedia, nil
}

func (ap *AnilistPlatform) GetAnimeByMalID(ctx context.Context, malID int) (*anilist.BaseAnime, error) {
	ap.logger.Trace().Msg("anilist platform: Fetching anime by MAL ID")
	ret, err := ap.anilistClient.BaseAnimeByMalID(ctx, &malID)
	if err != nil {
		return nil, err
	}

	media := ret.GetMedia()
	return ap.helper.TriggerGetAnimeEvent(media)
}

func (ap *AnilistPlatform) GetAnimeDetails(ctx context.Context, mediaID int) (*anilist.AnimeDetailsById_Media, error) {
	ap.logger.Trace().Int("mediaId", mediaID).Msg("anilist platform: Fetching anime details")

	// Check if this is a custom source entry
	if media, isCustom, err := ap.helper.HandleCustomSourceAnimeDetails(ctx, mediaID); isCustom {
		if err != nil {
			return nil, err
		}
		return ap.helper.TriggerGetAnimeDetailsEvent(media)
	}

	// Get from AniList
	ret, err := ap.anilistClient.AnimeDetailsByID(ctx, &mediaID)
	if err != nil {
		return nil, err
	}

	media := ret.GetMedia()
	return ap.helper.TriggerGetAnimeDetailsEvent(media)
}

func (ap *AnilistPlatform) GetAnimeWithRelations(ctx context.Context, mediaID int) (*anilist.CompleteAnime, error) {
	ap.logger.Trace().Int("mediaId", mediaID).Msg("anilist platform: Fetching anime with relations")

	if cachedAnime, ok := ap.helper.GetCachedCompleteAnime(mediaID); ok {
		ap.logger.Trace().Msg("anilist platform: Cache HIT for anime with relations")
		return cachedAnime, nil
	}

	// Check if this is a custom source entry
	if media, isCustom, err := ap.helper.HandleCustomSourceAnimeWithRelations(ctx, mediaID); isCustom {
		if err != nil {
			return nil, err
		}
		ap.helper.SetCachedCompleteAnime(mediaID, media)
		return media, nil
	}

	// Get from AniList
	ret, err := ap.anilistClient.CompleteAnimeByID(ctx, &mediaID)
	if err != nil {
		return nil, err
	}
	media := ret.GetMedia()

	ap.helper.SetCachedCompleteAnime(mediaID, media)
	return media, nil
}

func (ap *AnilistPlatform) GetManga(ctx context.Context, mediaID int) (*anilist.BaseManga, error) {
	ap.logger.Trace().Int("mediaId", mediaID).Msg("anilist platform: Fetching manga")

	if cachedManga, ok := ap.helper.GetCachedBaseManga(mediaID); ok {
		ap.logger.Trace().Msg("anilist platform: Returning manga from cache")
		return ap.helper.TriggerGetMangaEvent(cachedManga)
	}

	// Check if this is a custom source entry
	if media, isCustom, err := ap.helper.HandleCustomSourceManga(ctx, mediaID); isCustom {
		if err != nil {
			return nil, err
		}

		triggeredMedia, err := ap.helper.TriggerGetMangaEvent(media)
		if err != nil {
			return nil, err
		}

		ap.helper.SetCachedBaseManga(mediaID, triggeredMedia)
		return triggeredMedia, nil
	}

	// Get from AniList
	ret, err := ap.anilistClient.BaseMangaByID(ctx, &mediaID)
	if err != nil {
		return nil, err
	}

	media := ret.GetMedia()
	triggeredMedia, err := ap.helper.TriggerGetMangaEvent(media)
	if err != nil {
		return nil, err
	}

	ap.helper.SetCachedBaseManga(mediaID, triggeredMedia)
	return triggeredMedia, nil
}

func (ap *AnilistPlatform) GetMangaDetails(ctx context.Context, mediaID int) (*anilist.MangaDetailsById_Media, error) {
	ap.logger.Trace().Msg("anilist platform: Fetching manga details")

	// Check if this is a custom source entry
	if media, isCustom, err := ap.helper.HandleCustomSourceMangaDetails(ctx, mediaID); isCustom {
		return media, err
	}

	// Get from AniList
	ret, err := ap.anilistClient.MangaDetailsByID(ctx, &mediaID)
	if err != nil {
		return nil, err
	}

	return ret.GetMedia(), nil
}

func (ap *AnilistPlatform) GetAnimeCollection(ctx context.Context, bypassCache bool) (*anilist.AnimeCollection, error) {
	if !bypassCache && ap.animeCollection.IsPresent() {
		event := new(platform.GetCachedAnimeCollectionEvent)
		event.AnimeCollection = ap.animeCollection.MustGet()
		err := hook.GlobalHookManager.OnGetCachedAnimeCollection().Trigger(event)
		if err != nil {
			return nil, err
		}
		return event.AnimeCollection, nil
	}

	if ap.username.IsAbsent() {
		return nil, nil
	}

	err := ap.refreshAnimeCollection(ctx)
	if err != nil {
		return nil, err
	}

	event := new(platform.GetAnimeCollectionEvent)
	event.AnimeCollection = ap.animeCollection.MustGet()

	err = hook.GlobalHookManager.OnGetAnimeCollection().Trigger(event)
	if err != nil {
		return nil, err
	}

	return event.AnimeCollection, nil
}

func (ap *AnilistPlatform) GetRawAnimeCollection(ctx context.Context, bypassCache bool) (*anilist.AnimeCollection, error) {
	if !bypassCache && ap.rawAnimeCollection.IsPresent() {
		event := new(platform.GetCachedRawAnimeCollectionEvent)
		event.AnimeCollection = ap.rawAnimeCollection.MustGet()
		err := hook.GlobalHookManager.OnGetCachedRawAnimeCollection().Trigger(event)
		if err != nil {
			return nil, err
		}
		return event.AnimeCollection, nil
	}

	if ap.username.IsAbsent() {
		return nil, nil
	}

	err := ap.refreshAnimeCollection(ctx)
	if err != nil {
		return nil, err
	}

	event := new(platform.GetRawAnimeCollectionEvent)
	event.AnimeCollection = ap.rawAnimeCollection.MustGet()

	err = hook.GlobalHookManager.OnGetRawAnimeCollection().Trigger(event)
	if err != nil {
		return nil, err
	}

	return event.AnimeCollection, nil
}

func (ap *AnilistPlatform) RefreshAnimeCollection(ctx context.Context) (*anilist.AnimeCollection, error) {
	if ap.username.IsAbsent() {
		return nil, nil
	}

	err := ap.refreshAnimeCollection(ctx)
	if err != nil {
		return nil, err
	}

	event := new(platform.GetAnimeCollectionEvent)
	event.AnimeCollection = ap.animeCollection.MustGet()

	err = hook.GlobalHookManager.OnGetAnimeCollection().Trigger(event)
	if err != nil {
		return nil, err
	}

	event2 := new(platform.GetRawAnimeCollectionEvent)
	event2.AnimeCollection = ap.rawAnimeCollection.MustGet()

	err = hook.GlobalHookManager.OnGetRawAnimeCollection().Trigger(event2)
	if err != nil {
		return nil, err
	}

	return event.AnimeCollection, nil
}

func (ap *AnilistPlatform) refreshAnimeCollection(ctx context.Context) error {
	if ap.username.IsAbsent() {
		return errors.New("anilist: Username is not set")
	}

	// Else, get the collection from Anilist
	collection, err := ap.anilistClient.AnimeCollection(ctx, ap.username.ToPointer())
	if err != nil {
		return err
	}

	// Merge the custom entries into the collection
	ap.helper.MergeCustomSourceAnimeEntries(collection)

	// Save the raw collection to App (retains the lists with no status)
	ap.rawAnimeCollection = mo.Some(new(*collection))
	ap.rawAnimeCollection.MustGet().MediaListCollection = new(*collection.MediaListCollection)
	listsCopy := make([]*anilist.AnimeCollection_MediaListCollection_Lists, len(collection.MediaListCollection.Lists))
	copy(listsCopy, collection.MediaListCollection.Lists)
	ap.rawAnimeCollection.MustGet().MediaListCollection.Lists = listsCopy

	// Remove lists with no status (custom lists)
	collection.MediaListCollection.Lists = ap.helper.FilterOutCustomAnimeLists(collection.MediaListCollection.Lists)

	// Save the collection to App
	ap.animeCollection = mo.Some(collection)

	return nil
}

func (ap *AnilistPlatform) GetAnimeCollectionWithRelations(ctx context.Context) (*anilist.AnimeCollectionWithRelations, error) {
	ap.logger.Trace().Msg("anilist platform: Fetching anime collection with relations")

	if ap.username.IsAbsent() {
		return nil, nil
	}

	ret, err := ap.anilistClient.AnimeCollectionWithRelations(ctx, ap.username.ToPointer())
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (ap *AnilistPlatform) GetMangaCollection(ctx context.Context, bypassCache bool) (*anilist.MangaCollection, error) {

	if !bypassCache && ap.mangaCollection.IsPresent() {
		event := new(platform.GetCachedMangaCollectionEvent)
		event.MangaCollection = ap.mangaCollection.MustGet()
		err := hook.GlobalHookManager.OnGetCachedMangaCollection().Trigger(event)
		if err != nil {
			return nil, err
		}
		return event.MangaCollection, nil
	}

	if ap.username.IsAbsent() {
		return nil, nil
	}

	err := ap.refreshMangaCollection(ctx)
	if err != nil {
		return nil, err
	}

	event := new(platform.GetMangaCollectionEvent)
	event.MangaCollection = ap.mangaCollection.MustGet()

	err = hook.GlobalHookManager.OnGetMangaCollection().Trigger(event)
	if err != nil {
		return nil, err
	}

	return event.MangaCollection, nil
}

func (ap *AnilistPlatform) GetRawMangaCollection(ctx context.Context, bypassCache bool) (*anilist.MangaCollection, error) {
	ap.logger.Trace().Msg("anilist platform: Fetching raw manga collection")

	if !bypassCache && ap.rawMangaCollection.IsPresent() {
		ap.logger.Trace().Msg("anilist platform: Returning raw manga collection from cache")
		event := new(platform.GetCachedRawMangaCollectionEvent)
		event.MangaCollection = ap.rawMangaCollection.MustGet()
		err := hook.GlobalHookManager.OnGetCachedRawMangaCollection().Trigger(event)
		if err != nil {
			return nil, err
		}
		return event.MangaCollection, nil
	}

	if ap.username.IsAbsent() {
		return nil, nil
	}

	err := ap.refreshMangaCollection(ctx)
	if err != nil {
		return nil, err
	}

	event := new(platform.GetRawMangaCollectionEvent)
	event.MangaCollection = ap.rawMangaCollection.MustGet()

	err = hook.GlobalHookManager.OnGetRawMangaCollection().Trigger(event)
	if err != nil {
		return nil, err
	}

	return event.MangaCollection, nil
}

func (ap *AnilistPlatform) RefreshMangaCollection(ctx context.Context) (*anilist.MangaCollection, error) {
	if ap.username.IsAbsent() {
		return nil, nil
	}

	err := ap.refreshMangaCollection(ctx)
	if err != nil {
		return nil, err
	}

	event := new(platform.GetMangaCollectionEvent)
	event.MangaCollection = ap.mangaCollection.MustGet()

	err = hook.GlobalHookManager.OnGetMangaCollection().Trigger(event)
	if err != nil {
		return nil, err
	}

	event2 := new(platform.GetRawMangaCollectionEvent)
	event2.MangaCollection = ap.rawMangaCollection.MustGet()

	err = hook.GlobalHookManager.OnGetRawMangaCollection().Trigger(event2)
	if err != nil {
		return nil, err
	}

	return event.MangaCollection, nil
}

func (ap *AnilistPlatform) refreshMangaCollection(ctx context.Context) error {
	if ap.username.IsAbsent() {
		return errors.New("anilist: Username is not set")
	}

	collection, err := ap.anilistClient.MangaCollection(ctx, ap.username.ToPointer())
	if err != nil {
		return err
	}

	// Merge the custom entries into the collection
	ap.helper.MergeCustomSourceMangaEntries(collection)

	// Save the raw collection to App (retains the lists with no status)
	ap.rawMangaCollection = mo.Some(new(*collection))
	ap.rawMangaCollection.MustGet().MediaListCollection = new(*collection.MediaListCollection)
	listsCopy := make([]*anilist.MangaCollection_MediaListCollection_Lists, len(collection.MediaListCollection.Lists))
	copy(listsCopy, collection.MediaListCollection.Lists)
	ap.rawMangaCollection.MustGet().MediaListCollection.Lists = listsCopy

	// Remove lists with no status (custom lists)
	collection.MediaListCollection.Lists = ap.helper.FilterOutCustomMangaLists(collection.MediaListCollection.Lists)

	// Remove Novels from both collections
	ap.helper.RemoveNovelsFromMangaCollection(collection)
	ap.helper.RemoveNovelsFromMangaCollection(ap.rawMangaCollection.MustGet())

	// Save the collection to App
	ap.mangaCollection = mo.Some(collection)

	return nil
}

func (ap *AnilistPlatform) AddMediaToCollection(ctx context.Context, mIds []int) error {
	ap.logger.Trace().Msg("anilist platform: Adding media to collection")
	if len(mIds) == 0 {
		ap.logger.Debug().Msg("anilist: No media added to planning list")
		return nil
	}

	rateLimiter := limiter.NewLimiter(1*time.Second, 1) // 1 request per second

	wg := sync.WaitGroup{}
	for _, _id := range mIds {
		wg.Add(1)
		go func(id int) {
			rateLimiter.Wait()
			defer wg.Done()

			if customsource.IsExtensionId(id) {
				_, err := ap.helper.HandleCustomSourceUpdateEntry(ctx,
					id,
					new(anilist.MediaListStatusPlanning),
					new(0),
					new(0),
					nil,
					nil,
				)
				if err != nil {
					ap.logger.Error().Msg("anilist: An error occurred while adding media to planning list: " + err.Error())
				}
				return
			}

			_, err := ap.anilistClient.UpdateMediaListEntry(
				ctx,
				&id,
				new(anilist.MediaListStatusPlanning),
				new(0),
				new(0),
				nil,
				nil,
			)
			if err != nil {
				ap.logger.Error().Msg("anilist: An error occurred while adding media to planning list: " + err.Error())
			}
		}(_id)
	}
	wg.Wait()

	ap.logger.Debug().Any("count", len(mIds)).Msg("anilist: Media added to planning list")
	return nil
}

func (ap *AnilistPlatform) GetStudioDetails(ctx context.Context, studioID int) (*anilist.StudioDetails, error) {
	ap.logger.Trace().Msg("anilist platform: Fetching studio details")
	ret, err := ap.anilistClient.StudioDetails(ctx, &studioID)
	if err != nil {
		return nil, err
	}

	return ap.helper.TriggerGetStudioDetailsEvent(ret)
}

func (ap *AnilistPlatform) GetAnilistClient() anilist.AnilistClient {
	return ap.anilistClient
}

func (ap *AnilistPlatform) GetViewerStats(ctx context.Context) (*anilist.ViewerStats, error) {
	if ap.username.IsAbsent() {
		return nil, errors.New("anilist: Username is not set")
	}

	ap.logger.Trace().Msg("anilist platform: Fetching viewer stats")
	ret, err := ap.anilistClient.ViewerStats(ctx)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (ap *AnilistPlatform) GetAnimeAiringSchedule(ctx context.Context) (*anilist.AnimeAiringSchedule, error) {
	if ap.username.IsAbsent() {
		return nil, errors.New("anilist: Username is not set")
	}

	collection, err := ap.GetAnimeCollection(ctx, false)
	if err != nil {
		return nil, err
	}

	return ap.helper.BuildAnimeAiringSchedule(ctx, collection, ap.anilistClient)
}
