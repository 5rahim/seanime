package shared_platform

import (
	"context"
	"errors"
	"seanime/internal/api/anilist"
	"seanime/internal/customsource"
	"seanime/internal/database/db"
	"seanime/internal/extension"
	"seanime/internal/hook"
	"seanime/internal/platforms/platform"
	"seanime/internal/util"
	"seanime/internal/util/result"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

type PlatformHelper struct {
	logger              *zerolog.Logger
	customSourceManager *customsource.Manager
	baseAnimeCache      *result.BoundedCache[int, *anilist.BaseAnime]
	baseMangaCache      *result.BoundedCache[int, *anilist.BaseManga]
	completeAnimeCache  *result.BoundedCache[int, *anilist.CompleteAnime]
	extensionBankRef    *util.Ref[*extension.UnifiedBank]
}

func NewPlatformHelper(extensionBankRef *util.Ref[*extension.UnifiedBank], db *db.Database, logger *zerolog.Logger) *PlatformHelper {
	helper := &PlatformHelper{
		logger:              logger,
		baseAnimeCache:      result.NewBoundedCache[int, *anilist.BaseAnime](50),
		baseMangaCache:      result.NewBoundedCache[int, *anilist.BaseManga](50),
		completeAnimeCache:  result.NewBoundedCache[int, *anilist.CompleteAnime](10),
		extensionBankRef:    extensionBankRef,
		customSourceManager: customsource.NewManager(extensionBankRef, db, logger),
	}

	return helper
}

func (h *PlatformHelper) Close() {
	if h.customSourceManager != nil {
		h.customSourceManager.Close()
	}
}

func (h *PlatformHelper) ClearCache() {
	h.baseAnimeCache.Clear()
	h.baseMangaCache.Clear()
	h.completeAnimeCache.Clear()
}

func (h *PlatformHelper) GetCustomSourceManager() *customsource.Manager {
	return h.customSourceManager
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Custom Source
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (h *PlatformHelper) HandleCustomSourceAnime(ctx context.Context, mediaID int) (*anilist.BaseAnime, bool, error) {
	if h.customSourceManager == nil {
		return nil, false, nil
	}

	if customSource, localId, isCustom, hasExtension := h.customSourceManager.GetProviderFromId(mediaID); isCustom {
		if !hasExtension {
			return nil, true, errors.New("custom source does not exist or identifier has changed")
		}
		ret, err := customSource.GetProvider().GetAnime(ctx, []int{localId})
		if err != nil {
			return nil, true, err
		}
		if len(ret) == 0 {
			return nil, true, errors.New("no anime found")
		}
		media := ret[0]
		customsource.NormalizeMedia(customSource.GetExtensionIdentifier(), customSource.GetID(), media)
		return media, true, nil
	}

	return nil, false, nil
}

func (h *PlatformHelper) HandleCustomSourceAnimeDetails(ctx context.Context, mediaID int) (*anilist.AnimeDetailsById_Media, bool, error) {
	if h.customSourceManager == nil {
		return nil, false, nil
	}

	if customSource, localId, isCustom, hasExtension := h.customSourceManager.GetProviderFromId(mediaID); isCustom {
		if !hasExtension {
			return nil, true, errors.New("custom source does not exist or identifier has changed")
		}
		ret, err := customSource.GetProvider().GetAnimeDetails(ctx, localId)
		if err != nil {
			return nil, true, err
		}
		customsource.NormalizeMedia(customSource.GetExtensionIdentifier(), customSource.GetID(), ret)
		return ret, true, nil
	}

	return nil, false, nil
}

func (h *PlatformHelper) HandleCustomSourceAnimeWithRelations(ctx context.Context, mediaID int) (*anilist.CompleteAnime, bool, error) {
	if h.customSourceManager == nil {
		return nil, false, nil
	}

	if customSource, localId, isCustom, hasExtension := h.customSourceManager.GetProviderFromId(mediaID); isCustom {
		if !hasExtension {
			return nil, true, errors.New("custom source does not exist or identifier has changed")
		}
		ret, err := customSource.GetProvider().GetAnimeWithRelations(ctx, localId)
		if err != nil {
			return nil, true, err
		}
		customsource.NormalizeMedia(customSource.GetExtensionIdentifier(), customSource.GetID(), ret)
		return ret, true, nil
	}

	return nil, false, nil
}

func (h *PlatformHelper) HandleCustomSourceManga(ctx context.Context, mediaID int) (*anilist.BaseManga, bool, error) {
	if h.customSourceManager == nil {
		return nil, false, nil
	}

	if customSource, localId, isCustom, hasExtension := h.customSourceManager.GetProviderFromId(mediaID); isCustom {
		if !hasExtension {
			return nil, true, errors.New("custom source does not exist or identifier has changed")
		}
		ret, err := customSource.GetProvider().GetManga(ctx, []int{localId})
		if err != nil {
			return nil, true, err
		}
		if len(ret) == 0 {
			return nil, true, errors.New("no manga found")
		}
		media := ret[0]
		customsource.NormalizeMedia(customSource.GetExtensionIdentifier(), customSource.GetID(), media)
		return media, true, nil
	}

	return nil, false, nil
}

func (h *PlatformHelper) HandleCustomSourceMangaDetails(ctx context.Context, mediaID int) (*anilist.MangaDetailsById_Media, bool, error) {
	if h.customSourceManager == nil {
		return nil, false, nil
	}

	if customSource, localId, isCustom, hasExtension := h.customSourceManager.GetProviderFromId(mediaID); isCustom {
		if !hasExtension {
			return nil, true, errors.New("custom source does not exist or identifier has changed")
		}
		ret, err := customSource.GetProvider().GetMangaDetails(ctx, localId)
		if err != nil {
			return nil, true, err
		}
		customsource.NormalizeMedia(customSource.GetExtensionIdentifier(), customSource.GetID(), ret)
		return ret, true, nil
	}

	return nil, false, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Cache
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (h *PlatformHelper) GetCachedBaseAnime(mediaID int) (*anilist.BaseAnime, bool) {
	return h.baseAnimeCache.Get(mediaID)
}

func (h *PlatformHelper) SetCachedBaseAnime(mediaID int, anime *anilist.BaseAnime) {
	h.baseAnimeCache.SetT(mediaID, anime, time.Minute*30)
}

func (h *PlatformHelper) GetCachedBaseManga(mediaID int) (*anilist.BaseManga, bool) {
	return h.baseMangaCache.Get(mediaID)
}

func (h *PlatformHelper) SetCachedBaseManga(mediaID int, manga *anilist.BaseManga) {
	h.baseMangaCache.SetT(mediaID, manga, time.Minute*30)
}

func (h *PlatformHelper) GetCachedCompleteAnime(mediaID int) (*anilist.CompleteAnime, bool) {
	return h.completeAnimeCache.Get(mediaID)
}

func (h *PlatformHelper) SetCachedCompleteAnime(mediaID int, anime *anilist.CompleteAnime) {
	h.completeAnimeCache.SetT(mediaID, anime, 4*time.Hour)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Hook Events
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (h *PlatformHelper) TriggerGetAnimeEvent(anime *anilist.BaseAnime) (*anilist.BaseAnime, error) {
	event := new(platform.GetAnimeEvent)
	event.Anime = anime
	err := hook.GlobalHookManager.OnGetAnime().Trigger(event)
	if err != nil {
		return nil, err
	}
	return event.Anime, nil
}

func (h *PlatformHelper) TriggerGetAnimeDetailsEvent(anime *anilist.AnimeDetailsById_Media) (*anilist.AnimeDetailsById_Media, error) {
	event := new(platform.GetAnimeDetailsEvent)
	event.Anime = anime
	err := hook.GlobalHookManager.OnGetAnimeDetails().Trigger(event)
	if err != nil {
		return nil, err
	}
	return event.Anime, nil
}

func (h *PlatformHelper) TriggerGetMangaEvent(manga *anilist.BaseManga) (*anilist.BaseManga, error) {
	event := new(platform.GetMangaEvent)
	event.Manga = manga
	err := hook.GlobalHookManager.OnGetManga().Trigger(event)
	if err != nil {
		return nil, err
	}
	return event.Manga, nil
}

func (h *PlatformHelper) TriggerGetStudioDetailsEvent(studio *anilist.StudioDetails) (*anilist.StudioDetails, error) {
	event := new(platform.GetStudioDetailsEvent)
	event.Studio = studio
	err := hook.GlobalHookManager.OnGetStudioDetails().Trigger(event)
	if err != nil {
		return nil, err
	}
	return event.Studio, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Custom Source
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (h *PlatformHelper) MergeCustomSourceAnimeEntries(collection *anilist.AnimeCollection) {
	if h.customSourceManager != nil {
		h.customSourceManager.MergeAnimeEntries(collection)
	}
}

func (h *PlatformHelper) MergeCustomSourceMangaEntries(collection *anilist.MangaCollection) {
	if h.customSourceManager != nil {
		h.customSourceManager.MergeMangaEntries(collection)
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Anime Airing Schedule
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (h *PlatformHelper) BuildAnimeAiringSchedule(ctx context.Context, collection *anilist.AnimeCollection, client anilist.AnilistClient) (*anilist.AnimeAiringSchedule, error) {
	mediaIds := make([]*int, 0)
	for _, list := range collection.MediaListCollection.Lists {
		for _, entry := range list.Entries {
			if customsource.IsExtensionId(entry.GetMedia().GetID()) {
				continue
			}
			mediaIds = append(mediaIds, &[]int{entry.GetMedia().GetID()}[0])
		}
	}

	var ret *anilist.AnimeAiringSchedule

	now := time.Now()
	currentSeason, currentSeasonYear := anilist.GetSeasonInfo(now, anilist.GetSeasonKindCurrent)
	previousSeason, previousSeasonYear := anilist.GetSeasonInfo(now, anilist.GetSeasonKindPrevious)
	nextSeason, nextSeasonYear := anilist.GetSeasonInfo(now, anilist.GetSeasonKindNext)

	var err error
	ret, err = client.AnimeAiringSchedule(ctx, mediaIds, &currentSeason, &currentSeasonYear, &previousSeason, &previousSeasonYear, &nextSeason, &nextSeasonYear)
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
			if customsource.IsExtensionId(entry.Media.GetID()) {
				continue
			}
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
		retB, err := client.AnimeAiringScheduleRaw(ctx, missingIds)
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

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Update
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (h *PlatformHelper) HandleCustomSourceUpdateEntry(ctx context.Context, mediaID int, status *anilist.MediaListStatus, scoreRaw *int, progress *int, startedAt *anilist.FuzzyDateInput, completedAt *anilist.FuzzyDateInput) (bool, error) {
	if h.customSourceManager != nil && customsource.IsExtensionId(mediaID) {
		err := h.customSourceManager.UpdateEntry(ctx, mediaID, status, scoreRaw, progress, startedAt, completedAt)
		return true, err
	}
	return false, nil
}

func (h *PlatformHelper) HandleCustomSourceUpdateEntryProgress(ctx context.Context, mediaID int, progress int, totalCount *int) (bool, error) {
	if h.customSourceManager != nil && customsource.IsExtensionId(mediaID) {
		err := h.customSourceManager.UpdateEntryProgress(ctx, mediaID, progress, totalCount)
		return true, err
	}
	return false, nil
}

func (h *PlatformHelper) HandleCustomSourceUpdateEntryRepeat(ctx context.Context, mediaID int, repeat int) (bool, error) {
	if h.customSourceManager != nil && customsource.IsExtensionId(mediaID) {
		err := h.customSourceManager.UpdateEntryRepeat(ctx, mediaID, repeat)
		return true, err
	}
	return false, nil
}

func (h *PlatformHelper) HandleCustomSourceDeleteEntry(ctx context.Context, mediaID int, entryId int) (bool, error) {
	if h.customSourceManager != nil && customsource.IsExtensionId(mediaID) {
		err := h.customSourceManager.DeleteEntry(ctx, mediaID, entryId)
		return true, err
	}
	return false, nil
}

func (h *PlatformHelper) TriggerUpdateEntryHooks(ctx context.Context, mediaID int, status *anilist.MediaListStatus, scoreRaw *int, progress *int, startedAt *anilist.FuzzyDateInput, completedAt *anilist.FuzzyDateInput, updateFunc func(event *platform.PreUpdateEntryEvent) error) error {
	// Trigger pre-update hook
	event := new(platform.PreUpdateEntryEvent)
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

	// Execute the update
	err = updateFunc(event)
	if err != nil {
		return err
	}

	// Trigger post-update hook
	postEvent := new(platform.PostUpdateEntryEvent)
	postEvent.MediaID = &mediaID
	err = hook.GlobalHookManager.OnPostUpdateEntry().Trigger(postEvent)
	return err
}

// TriggerUpdateEntryProgressHooks triggers pre and post update entry progress hooks
func (h *PlatformHelper) TriggerUpdateEntryProgressHooks(ctx context.Context, mediaID int, progress int, totalCount *int, updateFunc func(event *platform.PreUpdateEntryProgressEvent) error) error {
	// Trigger pre-update hook
	event := new(platform.PreUpdateEntryProgressEvent)
	event.MediaID = &mediaID
	event.Progress = &progress
	event.TotalCount = totalCount
	currentStatus := anilist.MediaListStatusCurrent
	event.Status = &currentStatus

	_ = hook.GlobalHookManager.OnPreUpdateEntryProgress().Trigger(event)

	if event.DefaultPrevented {
		return nil
	}

	// Execute the update
	err := updateFunc(event)
	if err != nil {
		return err
	}

	// Trigger post-update hook
	postEvent := new(platform.PostUpdateEntryProgressEvent)
	postEvent.MediaID = &mediaID
	_ = hook.GlobalHookManager.OnPostUpdateEntryProgress().Trigger(postEvent)
	return err
}

func (h *PlatformHelper) TriggerUpdateEntryRepeatHooks(ctx context.Context, mediaID int, repeat int, updateFunc func(event *platform.PreUpdateEntryRepeatEvent) error) error {
	// Trigger pre-update hook
	event := new(platform.PreUpdateEntryRepeatEvent)
	event.MediaID = &mediaID
	event.Repeat = &repeat

	err := hook.GlobalHookManager.OnPreUpdateEntryRepeat().Trigger(event)
	if err != nil {
		return err
	}

	if event.DefaultPrevented {
		return nil
	}

	// Execute the update
	err = updateFunc(event)
	if err != nil {
		return err
	}

	// Trigger post-update hook
	postEvent := new(platform.PostUpdateEntryRepeatEvent)
	postEvent.MediaID = &mediaID
	err = hook.GlobalHookManager.OnPostUpdateEntryRepeat().Trigger(postEvent)
	return err
}

func (h *PlatformHelper) TriggerDeleteEntryHooks(ctx context.Context, mediaID int, entryId int, deleteFunc func(event *platform.PreDeleteEntryEvent) error) error {
	// Trigger pre-delete hook
	event := new(platform.PreDeleteEntryEvent)
	event.MediaID = &mediaID
	event.EntryID = &entryId

	err := hook.GlobalHookManager.OnPreDeleteEntry().Trigger(event)
	if err != nil {
		return err
	}

	if event.DefaultPrevented {
		return nil
	}

	// Execute the deletion
	err = deleteFunc(event)
	if err != nil {
		return err
	}

	// Trigger post-delete hook
	postEvent := new(platform.PostDeleteEntryEvent)
	postEvent.MediaID = &mediaID
	postEvent.EntryID = &entryId
	err = hook.GlobalHookManager.OnPostDeleteEntry().Trigger(postEvent)
	return err
}

func (h *PlatformHelper) FilterOutCustomAnimeLists(lists []*anilist.AnimeCollection_MediaListCollection_Lists) []*anilist.AnimeCollection_MediaListCollection_Lists {
	return lo.Filter(lists, func(list *anilist.AnimeCollection_MediaListCollection_Lists, _ int) bool {
		return list.Status != nil
	})
}

func (h *PlatformHelper) FilterOutCustomMangaLists(lists []*anilist.MangaCollection_MediaListCollection_Lists) []*anilist.MangaCollection_MediaListCollection_Lists {
	return lo.Filter(lists, func(list *anilist.MangaCollection_MediaListCollection_Lists, _ int) bool {
		return list.Status != nil
	})
}

func (h *PlatformHelper) RemoveNovelsFromMangaCollection(collection *anilist.MangaCollection) {
	for _, list := range collection.MediaListCollection.Lists {
		// Filter out novel entries
		list.Entries = lo.Filter(list.Entries, func(e *anilist.MangaCollection_MediaListCollection_Lists_Entries, _ int) bool {
			if e.GetMedia().GetFormat() == nil {
				return true
			}
			return *e.GetMedia().GetFormat() != anilist.MediaFormatNovel
		})
	}
}
