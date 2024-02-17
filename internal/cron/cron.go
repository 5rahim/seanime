package cron

import (
	"github.com/seanime-app/seanime/internal/core"
	"time"
)

type JobCtx struct {
	App *core.App
}

func RunJobs(app *core.App) {

	ctx := &JobCtx{
		App: app,
	}

	refreshAnilistTicker := time.NewTicker(10 * time.Minute)
	refetchReleaseTicker := time.NewTicker(1 * time.Hour)

	go func() {
		for {
			select {
			case <-refreshAnilistTicker.C:
				RefreshAnilistCollectionJob(ctx)
			case <-refetchReleaseTicker.C:
				app.Updater.ShouldRefetchReleases()
			}
		}
	}()

}
