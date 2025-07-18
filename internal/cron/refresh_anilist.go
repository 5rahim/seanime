package cron

import (
	"seanime/internal/events"
)

func RefreshAnilistDataJob(c *JobCtx) {
	defer func() {
		if r := recover(); r != nil {
		}
	}()

	if c.App.Settings == nil || c.App.Settings.Library == nil {
		return
	}

	// Refresh the Anilist Collection
	animeCollection, _ := c.App.RefreshAnimeCollection()

	if c.App.Settings.GetLibrary().EnableManga {
		mangaCollection, _ := c.App.RefreshMangaCollection()
		c.App.WSEventManager.SendEvent(events.RefreshedAnilistMangaCollection, mangaCollection)
	}

	c.App.WSEventManager.SendEvent(events.RefreshedAnilistAnimeCollection, animeCollection)
}

func SyncLocalDataJob(c *JobCtx) {
	defer func() {
		if r := recover(); r != nil {
		}
	}()

	if c.App.Settings == nil || c.App.Settings.Library == nil || !c.App.Settings.Library.AutoSyncOfflineLocalData {
		return
	}

	// Only synchronize local data if the user is not simulated
	if !c.App.GetUser().IsSimulated {
		_ = c.App.LocalManager.SynchronizeLocal()
	}
}
