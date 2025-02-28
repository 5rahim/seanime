package plugin

import (
	"seanime/internal/events"
	"seanime/internal/extension"

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
	anilist := &Anilist{
		ctx:    a,
		ext:    ext,
		logger: &anilistLogger,
	}
	anilistObj := vm.NewObject()
	_ = anilistObj.Set("refreshAnimeCollection", anilist.RefreshAnimeCollection)
	_ = anilistObj.Set("refreshMangaCollection", anilist.RefreshMangaCollection)

	// Bind anilist platform
	anilistPlatform, ok := a.anilistPlatform.Get()
	if ok {
		_ = anilistObj.Set("updateEntry", anilistPlatform.UpdateEntry)
		_ = anilistObj.Set("updateEntryProgress", anilistPlatform.UpdateEntryProgress)
		_ = anilistObj.Set("updateEntryRepeat", anilistPlatform.UpdateEntryRepeat)
		_ = anilistObj.Set("deleteEntry", anilistPlatform.DeleteEntry)
		_ = anilistObj.Set("getAnimeCollection", anilistPlatform.GetAnimeCollection)
		_ = anilistObj.Set("getRawAnimeCollection", anilistPlatform.GetRawAnimeCollection)
		_ = anilistObj.Set("getMangaCollection", anilistPlatform.GetMangaCollection)
		_ = anilistObj.Set("getRawMangaCollection", anilistPlatform.GetRawMangaCollection)
		_ = anilistObj.Set("getAnime", anilistPlatform.GetAnime)
		_ = anilistObj.Set("getManga", anilistPlatform.GetManga)
		_ = anilistObj.Set("getAnimeDetails", anilistPlatform.GetAnimeDetails)
		_ = anilistObj.Set("getMangaDetails", anilistPlatform.GetMangaDetails)
		_ = anilistObj.Set("getAnimeCollectionWithRelations", anilistPlatform.GetAnimeCollectionWithRelations)
		_ = anilistObj.Set("addMediaToCollection", anilistPlatform.AddMediaToCollection)
		_ = anilistObj.Set("getStudioDetails", anilistPlatform.GetStudioDetails)

		anilistClient := anilistPlatform.GetAnilistClient()
		_ = anilistObj.Set("listAnime", anilistClient.ListAnime)
		_ = anilistObj.Set("listManga", anilistClient.ListManga)
		_ = anilistObj.Set("listRecentAnime", anilistClient.ListRecentAnime)

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
