package anilist_platform

import (
	"context"
	"errors"
	"seanime/internal/api/anilist"
	"seanime/internal/hook"
	"seanime/internal/platforms/platform"
	"seanime/internal/util/limiter"
	"seanime/internal/util/result"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/lo"
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
		baseAnimeCache         *result.BoundedCache[int, *anilist.BaseAnime]
	}
)

func NewAnilistPlatform(anilistClient anilist.AnilistClient, logger *zerolog.Logger) platform.Platform {
	ap := &AnilistPlatform{
		anilistClient:      anilistClient,
		logger:             logger,
		username:           mo.None[string](),
		animeCollection:    mo.None[*anilist.AnimeCollection](),
		rawAnimeCollection: mo.None[*anilist.AnimeCollection](),
		mangaCollection:    mo.None[*anilist.MangaCollection](),
		rawMangaCollection: mo.None[*anilist.MangaCollection](),
		baseAnimeCache:     result.NewBoundedCache[int, *anilist.BaseAnime](50),
	}

	return ap
}

func (ap *AnilistPlatform) clearCache() {
	ap.baseAnimeCache.Clear()
}

func (ap *AnilistPlatform) SetUsername(username string) {
	// Set the username for the AnilistPlatform
	if username == "" {
		ap.username = mo.Some[string]("")
		return
	}

	ap.username = mo.Some(username)
	return
}

func (ap *AnilistPlatform) SetAnilistClient(client anilist.AnilistClient) {
	// Set the AnilistClient for the AnilistPlatform
	ap.anilistClient = client
}

func (ap *AnilistPlatform) UpdateEntry(ctx context.Context, mediaID int, status *anilist.MediaListStatus, scoreRaw *int, progress *int, startedAt *anilist.FuzzyDateInput, completedAt *anilist.FuzzyDateInput) error {
	ap.logger.Trace().Msg("anilist platform: Updating entry")

	event := new(PreUpdateEntryEvent)
	event.MediaID = &mediaID
	event.Status = status
	event.ScoreRaw = scoreRaw
	event.Progress = progress
	event.StartedAt = startedAt
	event.CompletedAt = completedAt

	err := hook.GlobalHookManager.OnPreUpdateEntry().Trigger(event)
	if err != nil {
		return err
	}

	if event.DefaultPrevented {
		return nil
	}

	_, err = ap.anilistClient.UpdateMediaListEntry(ctx, event.MediaID, event.Status, event.ScoreRaw, event.Progress, event.StartedAt, event.CompletedAt)
	if err != nil {
		return err
	}

	postEvent := new(PostUpdateEntryEvent)
	postEvent.MediaID = &mediaID

	err = hook.GlobalHookManager.OnPostUpdateEntry().Trigger(postEvent)

	return nil
}

func (ap *AnilistPlatform) UpdateEntryProgress(ctx context.Context, mediaID int, progress int, totalCount *int) error {
	ap.logger.Trace().Msg("anilist platform: Updating entry progress")

	event := new(PreUpdateEntryProgressEvent)
	event.MediaID = &mediaID
	event.Progress = &progress
	event.TotalCount = totalCount
	event.Status = lo.ToPtr(anilist.MediaListStatusCurrent)

	err := hook.GlobalHookManager.OnPreUpdateEntryProgress().Trigger(event)
	if err != nil {
		return err
	}

	if event.DefaultPrevented {
		return nil
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
	if realTotalCount > 0 && progress >= realTotalCount {
		*event.Status = anilist.MediaListStatusCompleted
	}

	if realTotalCount > 0 && progress > realTotalCount {
		*event.Progress = realTotalCount
	}

	_, err = ap.anilistClient.UpdateMediaListEntryProgress(
		ctx,
		event.MediaID,
		event.Progress,
		event.Status,
	)
	if err != nil {
		return err
	}

	postEvent := new(PostUpdateEntryProgressEvent)
	postEvent.MediaID = &mediaID

	err = hook.GlobalHookManager.OnPostUpdateEntryProgress().Trigger(postEvent)
	if err != nil {
		return err
	}

	return nil
}

func (ap *AnilistPlatform) UpdateEntryRepeat(ctx context.Context, mediaID int, repeat int) error {
	ap.logger.Trace().Msg("anilist platform: Updating entry repeat")

	event := new(PreUpdateEntryRepeatEvent)
	event.MediaID = &mediaID
	event.Repeat = &repeat

	err := hook.GlobalHookManager.OnPreUpdateEntryRepeat().Trigger(event)
	if err != nil {
		return err
	}

	if event.DefaultPrevented {
		return nil
	}

	_, err = ap.anilistClient.UpdateMediaListEntryRepeat(ctx, event.MediaID, event.Repeat)
	if err != nil {
		return err
	}

	postEvent := new(PostUpdateEntryRepeatEvent)
	postEvent.MediaID = &mediaID

	err = hook.GlobalHookManager.OnPostUpdateEntryRepeat().Trigger(postEvent)
	if err != nil {
		return err
	}

	return nil
}

func (ap *AnilistPlatform) DeleteEntry(ctx context.Context, mediaID int) error {
	ap.logger.Trace().Msg("anilist platform: Deleting entry")
	_, err := ap.anilistClient.DeleteEntry(ctx, &mediaID)
	if err != nil {
		return err
	}
	return nil
}

func (ap *AnilistPlatform) GetAnime(ctx context.Context, mediaID int) (*anilist.BaseAnime, error) {
	ap.logger.Trace().Msg("anilist platform: Fetching anime")

	//if cachedAnime, ok := ap.baseAnimeCache.Get(mediaID); ok {
	//	ap.logger.Trace().Msg("anilist platform: Returning anime from cache")
	//	event := new(GetAnimeEvent)
	//	event.Anime = cachedAnime
	//	err := hook.GlobalHookManager.OnGetAnime().Trigger(event)
	//	if err != nil {
	//		return nil, err
	//	}
	//	return event.Anime, nil
	//}

	ret, err := ap.anilistClient.BaseAnimeByID(ctx, &mediaID)
	if err != nil {

		return nil, err
	}

	media := ret.GetMedia()

	event := new(GetAnimeEvent)
	event.Anime = media

	err = hook.GlobalHookManager.OnGetAnime().Trigger(event)
	if err != nil {
		return nil, err
	}

	//ap.baseAnimeCache.SetT(mediaID, event.Anime, time.Minute*30)

	return event.Anime, nil
}

func (ap *AnilistPlatform) GetAnimeByMalID(ctx context.Context, malID int) (*anilist.BaseAnime, error) {
	ap.logger.Trace().Msg("anilist platform: Fetching anime by MAL ID")
	ret, err := ap.anilistClient.BaseAnimeByMalID(ctx, &malID)
	if err != nil {
		return nil, err
	}

	media := ret.GetMedia()

	event := new(GetAnimeEvent)
	event.Anime = media

	err = hook.GlobalHookManager.OnGetAnime().Trigger(event)
	if err != nil {
		return nil, err
	}

	return event.Anime, nil
}

func (ap *AnilistPlatform) GetAnimeDetails(ctx context.Context, mediaID int) (*anilist.AnimeDetailsById_Media, error) {
	ap.logger.Trace().Msg("anilist platform: Fetching anime details")
	ret, err := ap.anilistClient.AnimeDetailsByID(ctx, &mediaID)
	if err != nil {
		return nil, err
	}

	media := ret.GetMedia()

	event := new(GetAnimeDetailsEvent)
	event.Anime = media

	err = hook.GlobalHookManager.OnGetAnimeDetails().Trigger(event)
	if err != nil {
		return nil, err
	}

	return event.Anime, nil
}

func (ap *AnilistPlatform) GetAnimeWithRelations(ctx context.Context, mediaID int) (*anilist.CompleteAnime, error) {
	ap.logger.Trace().Msg("anilist platform: Fetching anime with relations")
	ret, err := ap.anilistClient.CompleteAnimeByID(ctx, &mediaID)
	if err != nil {
		return nil, err
	}
	return ret.GetMedia(), nil
}

func (ap *AnilistPlatform) GetManga(ctx context.Context, mediaID int) (*anilist.BaseManga, error) {
	ap.logger.Trace().Msg("anilist platform: Fetching manga")
	ret, err := ap.anilistClient.BaseMangaByID(ctx, &mediaID)
	if err != nil {
		return nil, err
	}

	media := ret.GetMedia()

	event := new(GetMangaEvent)
	event.Manga = media

	err = hook.GlobalHookManager.OnGetManga().Trigger(event)
	if err != nil {
		return nil, err
	}

	return event.Manga, nil
}

func (ap *AnilistPlatform) GetMangaDetails(ctx context.Context, mediaID int) (*anilist.MangaDetailsById_Media, error) {
	ap.logger.Trace().Msg("anilist platform: Fetching manga details")
	ret, err := ap.anilistClient.MangaDetailsByID(ctx, &mediaID)
	if err != nil {
		return nil, err
	}
	return ret.GetMedia(), nil
}

func (ap *AnilistPlatform) GetAnimeCollection(ctx context.Context, bypassCache bool) (*anilist.AnimeCollection, error) {
	if !bypassCache && ap.animeCollection.IsPresent() {
		event := new(GetCachedAnimeCollectionEvent)
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

	event := new(GetAnimeCollectionEvent)
	event.AnimeCollection = ap.animeCollection.MustGet()

	err = hook.GlobalHookManager.OnGetAnimeCollection().Trigger(event)
	if err != nil {
		return nil, err
	}

	return event.AnimeCollection, nil
}

func (ap *AnilistPlatform) GetRawAnimeCollection(ctx context.Context, bypassCache bool) (*anilist.AnimeCollection, error) {
	if !bypassCache && ap.rawAnimeCollection.IsPresent() {
		event := new(GetCachedRawAnimeCollectionEvent)
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

	event := new(GetRawAnimeCollectionEvent)
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

	event := new(GetAnimeCollectionEvent)
	event.AnimeCollection = ap.animeCollection.MustGet()

	err = hook.GlobalHookManager.OnGetAnimeCollection().Trigger(event)
	if err != nil {
		return nil, err
	}

	event2 := new(GetRawAnimeCollectionEvent)
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

	// Save the raw collection to App (retains the lists with no status)
	collectionCopy := *collection
	ap.rawAnimeCollection = mo.Some(&collectionCopy)
	listCollectionCopy := *collection.MediaListCollection
	ap.rawAnimeCollection.MustGet().MediaListCollection = &listCollectionCopy
	listsCopy := make([]*anilist.AnimeCollection_MediaListCollection_Lists, len(collection.MediaListCollection.Lists))
	copy(listsCopy, collection.MediaListCollection.Lists)
	ap.rawAnimeCollection.MustGet().MediaListCollection.Lists = listsCopy

	// Remove lists with no status (custom lists)
	collection.MediaListCollection.Lists = lo.Filter(collection.MediaListCollection.Lists, func(list *anilist.AnimeCollection_MediaListCollection_Lists, _ int) bool {
		return list.Status != nil
	})

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
		event := new(GetCachedMangaCollectionEvent)
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

	event := new(GetMangaCollectionEvent)
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
		event := new(GetCachedRawMangaCollectionEvent)
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

	event := new(GetRawMangaCollectionEvent)
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

	event := new(GetMangaCollectionEvent)
	event.MangaCollection = ap.mangaCollection.MustGet()

	err = hook.GlobalHookManager.OnGetMangaCollection().Trigger(event)
	if err != nil {
		return nil, err
	}

	event2 := new(GetRawMangaCollectionEvent)
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

	// Save the raw collection to App (retains the lists with no status)
	collectionCopy := *collection
	ap.rawMangaCollection = mo.Some(&collectionCopy)
	listCollectionCopy := *collection.MediaListCollection
	ap.rawMangaCollection.MustGet().MediaListCollection = &listCollectionCopy
	listsCopy := make([]*anilist.MangaCollection_MediaListCollection_Lists, len(collection.MediaListCollection.Lists))
	copy(listsCopy, collection.MediaListCollection.Lists)
	ap.rawMangaCollection.MustGet().MediaListCollection.Lists = listsCopy

	// Remove lists with no status (custom lists)
	collection.MediaListCollection.Lists = lo.Filter(collection.MediaListCollection.Lists, func(list *anilist.MangaCollection_MediaListCollection_Lists, _ int) bool {
		return list.Status != nil
	})

	// Remove Novels from both collections
	for _, list := range collection.MediaListCollection.Lists {
		for _, entry := range list.Entries {
			if entry.GetMedia().GetFormat() != nil && *entry.GetMedia().GetFormat() == anilist.MediaFormatNovel {
				list.Entries = lo.Filter(list.Entries, func(e *anilist.MangaCollection_MediaListCollection_Lists_Entries, _ int) bool {
					return *e.GetMedia().GetFormat() != anilist.MediaFormatNovel
				})
			}
		}
	}
	for _, list := range ap.rawMangaCollection.MustGet().MediaListCollection.Lists {
		for _, entry := range list.Entries {
			if entry.GetMedia().GetFormat() != nil && *entry.GetMedia().GetFormat() == anilist.MediaFormatNovel {
				list.Entries = lo.Filter(list.Entries, func(e *anilist.MangaCollection_MediaListCollection_Lists_Entries, _ int) bool {
					return *e.GetMedia().GetFormat() != anilist.MediaFormatNovel
				})
			}
		}
	}

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
			_, err := ap.anilistClient.UpdateMediaListEntry(
				ctx,
				&id,
				lo.ToPtr(anilist.MediaListStatusPlanning),
				lo.ToPtr(0),
				lo.ToPtr(0),
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

	event := new(GetStudioDetailsEvent)
	event.Studio = ret

	err = hook.GlobalHookManager.OnGetStudioDetails().Trigger(event)
	if err != nil {
		return nil, err
	}

	return event.Studio, nil
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

	mediaIds := make([]*int, 0)
	for _, list := range collection.MediaListCollection.Lists {
		for _, entry := range list.Entries {
			mediaIds = append(mediaIds, &[]int{entry.GetMedia().GetID()}[0])
		}
	}

	var ret *anilist.AnimeAiringSchedule

	now := time.Now()
	currentSeason, currentSeasonYear := anilist.GetSeasonInfo(now, anilist.GetSeasonKindCurrent)
	previousSeason, previousSeasonYear := anilist.GetSeasonInfo(now, anilist.GetSeasonKindPrevious)
	nextSeason, nextSeasonYear := anilist.GetSeasonInfo(now, anilist.GetSeasonKindNext)

	ret, err = ap.anilistClient.AnimeAiringSchedule(ctx, mediaIds, &currentSeason, &currentSeasonYear, &previousSeason, &previousSeasonYear, &nextSeason, &nextSeasonYear)
	if err != nil {
		return nil, err
	}

	type animeScheduleMedia interface {
		GetMedia() []*anilist.AnimeSchedule
	}

	foundIds := make(map[int]struct{})
	addIds := func(n animeScheduleMedia) {
		for _, m := range n.GetMedia() {
			if m == nil {
				continue
			}
			foundIds[m.GetID()] = struct{}{}
		}
	}
	addIds(ret.GetOngoing())
	addIds(ret.GetOngoingNext())
	addIds(ret.GetPreceding())
	addIds(ret.GetUpcoming())
	addIds(ret.GetUpcomingNext())

	missingIds := make([]*int, 0)
	for _, list := range collection.MediaListCollection.Lists {
		for _, entry := range list.Entries {
			if _, found := foundIds[entry.GetMedia().GetID()]; found {
				continue
			}
			endDate := entry.GetMedia().GetEndDate()
			// Ignore if ended more than 2 months ago
			if endDate == nil || endDate.GetYear() == nil || endDate.GetMonth() == nil {
				missingIds = append(missingIds, &[]int{entry.GetMedia().GetID()}[0])
				continue
			}
			endTime := time.Date(*endDate.GetYear(), time.Month(*endDate.GetMonth()), 1, 0, 0, 0, 0, time.UTC)
			if endTime.Before(now.AddDate(0, -2, 0)) {
				continue
			}
			missingIds = append(missingIds, &[]int{entry.GetMedia().GetID()}[0])
		}
	}

	if len(missingIds) > 0 {
		retB, err := ap.anilistClient.AnimeAiringScheduleRaw(ctx, missingIds)
		if err != nil {
			return nil, err
		}
		if len(retB.GetPage().GetMedia()) > 0 {
			// Add to ongoing next
			for _, m := range retB.Page.GetMedia() {
				if ret.OngoingNext == nil {
					ret.OngoingNext = &anilist.AnimeAiringSchedule_OngoingNext{
						Media: make([]*anilist.AnimeSchedule, 0),
					}
				}
				if m == nil {
					continue
				}

				ret.OngoingNext.Media = append(ret.OngoingNext.Media, m)
			}
		}
	}

	return ret, nil
}
