package cron

import (
	"github.com/seanime-app/seanime/internal/events"
)

func RefreshAnilistCollectionJob(c *JobCtx) {
	defer func() {
		if r := recover(); r != nil {
		}
	}()

	if c.App.Settings == nil || c.App.Settings.Library == nil {
		return
	}

	// Refresh the Anilist Collection
	anilistCollection, err := c.App.GetAnilistCollection(true)
	if err != nil {
		return
	}

	if c.App.Settings.Library.EnableManga {
		mangaCollection, err := c.App.GetMangaCollection(true)
		if err != nil {
			return
		}
		c.App.WSEventManager.SendEvent(events.RefreshedAnilistMangaCollection, mangaCollection)
	}

	c.App.WSEventManager.SendEvent(events.RefreshedAnilistCollection, anilistCollection)
}
