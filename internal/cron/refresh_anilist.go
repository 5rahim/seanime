package cron

import (
	"github.com/seanime-app/seanime-server/internal/events"
)

func RefreshAnilistCollectionJob(c *JobCtx) {
	// Refresh the Anilist Collection
	anilistCollection, err := c.App.GetAnilistCollection(true)
	if err != nil {
		return
	}

	c.App.WSEventManager.SendEvent(events.RefreshedAnilistCollection, anilistCollection)
}
