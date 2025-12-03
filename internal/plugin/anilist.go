package plugin

import (
	"context"
	"seanime/internal/api/anilist"
	"seanime/internal/events"
	"seanime/internal/extension"
	"seanime/internal/library/anime"

	"github.com/dop251/goja"
	"github.com/rs/zerolog"
)

type Anilist struct {
	ctx    *AppContextImpl
	ext    *extension.Extension
	logger *zerolog.Logger
}

// BindAnilist binds the anilist API to the Goja runtime.
// Permissions need to be checked by the caller.
// Permissions needed: anilist
func (a *AppContextImpl) BindAnilist(vm *goja.Runtime, logger *zerolog.Logger, ext *extension.Extension) {
	anilistLogger := logger.With().Str("id", ext.ID).Logger()
	al := &Anilist{
		ctx:    a,
		ext:    ext,
		logger: &anilistLogger,
	}
	anilistObj := vm.NewObject()
	_ = anilistObj.Set("refreshAnimeCollection", al.RefreshAnimeCollection)
	_ = anilistObj.Set("refreshMangaCollection", al.RefreshMangaCollection)

	// Bind anilist platform
	anilistPlatformRef, ok := a.anilistPlatformRef.Get()
	if ok {
		_ = anilistObj.Set("updateEntry", func(mediaID int, status *anilist.MediaListStatus, scoreRaw *int, progress *int, startedAt *anilist.FuzzyDateInput, completedAt *anilist.FuzzyDateInput) error {
			return anilistPlatformRef.Get().UpdateEntry(context.Background(), mediaID, status, scoreRaw, progress, startedAt, completedAt)
		})
		_ = anilistObj.Set("updateEntryProgress", func(mediaID int, progress int, totalEpisodes *int) error {
			return anilistPlatformRef.Get().UpdateEntryProgress(context.Background(), mediaID, progress, totalEpisodes)
		})
		_ = anilistObj.Set("updateEntryRepeat", func(mediaID int, repeat int) error {
			return anilistPlatformRef.Get().UpdateEntryRepeat(context.Background(), mediaID, repeat)
		})
		_ = anilistObj.Set("deleteEntry", func(mediaID int, entryId int) error {
			return anilistPlatformRef.Get().DeleteEntry(context.Background(), mediaID, entryId)
		})
		_ = anilistObj.Set("getAnimeCollection", func(bypassCache bool) (*anilist.AnimeCollection, error) {
			return anilistPlatformRef.Get().GetAnimeCollection(context.Background(), bypassCache)
		})
		_ = anilistObj.Set("getRawAnimeCollection", func(bypassCache bool) (*anilist.AnimeCollection, error) {
			return anilistPlatformRef.Get().GetRawAnimeCollection(context.Background(), bypassCache)
		})
		_ = anilistObj.Set("getMangaCollection", func(bypassCache bool) (*anilist.MangaCollection, error) {
			return anilistPlatformRef.Get().GetMangaCollection(context.Background(), bypassCache)
		})
		_ = anilistObj.Set("getRawMangaCollection", func(bypassCache bool) (*anilist.MangaCollection, error) {
			return anilistPlatformRef.Get().GetRawMangaCollection(context.Background(), bypassCache)
		})
		_ = anilistObj.Set("getAnime", func(mediaID int) (*anilist.BaseAnime, error) {
			return anilistPlatformRef.Get().GetAnime(context.Background(), mediaID)
		})
		_ = anilistObj.Set("getManga", func(mediaID int) (*anilist.BaseManga, error) {
			return anilistPlatformRef.Get().GetManga(context.Background(), mediaID)
		})
		_ = anilistObj.Set("getAnimeDetails", func(mediaID int) (*anilist.AnimeDetailsById_Media, error) {
			return anilistPlatformRef.Get().GetAnimeDetails(context.Background(), mediaID)
		})
		_ = anilistObj.Set("getMangaDetails", func(mediaID int) (*anilist.MangaDetailsById_Media, error) {
			return anilistPlatformRef.Get().GetMangaDetails(context.Background(), mediaID)
		})
		_ = anilistObj.Set("getAnimeCollectionWithRelations", func() (*anilist.AnimeCollectionWithRelations, error) {
			return anilistPlatformRef.Get().GetAnimeCollectionWithRelations(context.Background())
		})
		_ = anilistObj.Set("addMediaToCollection", func(mIds []int) error {
			return anilistPlatformRef.Get().AddMediaToCollection(context.Background(), mIds)
		})
		_ = anilistObj.Set("getStudioDetails", func(studioID int) (*anilist.StudioDetails, error) {
			return anilistPlatformRef.Get().GetStudioDetails(context.Background(), studioID)
		})
		_ = anilistObj.Set("listAnime", func(page *int, search *string, perPage *int, sort []*anilist.MediaSort, status []*anilist.MediaStatus, genres []*string, averageScoreGreater *int, season *anilist.MediaSeason, seasonYear *int, format *anilist.MediaFormat, isAdult *bool) (*anilist.ListAnime, error) {
			return anilistPlatformRef.Get().GetAnilistClient().ListAnime(context.Background(), page, search, perPage, sort, status, genres, averageScoreGreater, season, seasonYear, format, isAdult)
		})
		_ = anilistObj.Set("listManga", func(page *int, search *string, perPage *int, sort []*anilist.MediaSort, status []*anilist.MediaStatus, genres []*string, averageScoreGreater *int, startDateGreater *string, startDateLesser *string, format *anilist.MediaFormat, countryOfOrigin *string, isAdult *bool) (*anilist.ListManga, error) {
			return anilistPlatformRef.Get().GetAnilistClient().ListManga(context.Background(), page, search, perPage, sort, status, genres, averageScoreGreater, startDateGreater, startDateLesser, format, countryOfOrigin, isAdult)
		})
		_ = anilistObj.Set("listRecentAnime", func(page *int, perPage *int, airingAtGreater *int, airingAtLesser *int, notYetAired *bool) (*anilist.ListRecentAnime, error) {
			return anilistPlatformRef.Get().GetAnilistClient().ListRecentAnime(context.Background(), page, perPage, airingAtGreater, airingAtLesser, notYetAired)
		})
		_ = anilistObj.Set("clearCache", func() {
			anilistPlatformRef.Get().ClearCache()
			anime.ClearEpisodeCollectionCache()
		})
		_ = anilistObj.Set("customQuery", func(body map[string]interface{}, token string) (interface{}, error) {
			return anilist.CustomQuery(body, a.logger, token)
		})

	}

	_ = vm.Set("$anilist", anilistObj)
}

func (a *Anilist) RefreshAnimeCollection() {
	a.logger.Trace().Msg("plugin: Refreshing anime collection")
	onRefreshAnilistAnimeCollection, ok := a.ctx.onRefreshAnilistAnimeCollection.Get()
	if !ok {
		return
	}

	onRefreshAnilistAnimeCollection()
	wsEventManager, ok := a.ctx.wsEventManager.Get()
	if ok {
		wsEventManager.SendEvent(events.RefreshedAnilistAnimeCollection, nil)
	}
}

func (a *Anilist) RefreshMangaCollection() {
	a.logger.Trace().Msg("plugin: Refreshing manga collection")
	onRefreshAnilistMangaCollection, ok := a.ctx.onRefreshAnilistMangaCollection.Get()
	if !ok {
		return
	}

	onRefreshAnilistMangaCollection()
	wsEventManager, ok := a.ctx.wsEventManager.Get()
	if ok {
		wsEventManager.SendEvent(events.RefreshedAnilistMangaCollection, nil)
	}
}
