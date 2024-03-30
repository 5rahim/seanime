package cron

import (
	"github.com/seanime-app/seanime/internal/events"
)

func RefreshAnilistCollectionJob(c *JobCtx) {
	// Refresh the Anilist Collection
	anilistCollection, err := c.App.GetAnilistCollection(true)
	if err != nil {
		return
	}

	if c.App.Config.Manga.Enabled {
		mangaCollection, err := c.App.GetMangaCollection(true)
		if err != nil {
			return
		}
		c.App.WSEventManager.SendEvent(events.RefreshedAnilistMangaCollection, mangaCollection)
	}

	c.App.WSEventManager.SendEvent(events.RefreshedAnilistCollection, anilistCollection)
}
