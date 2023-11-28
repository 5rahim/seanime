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

	ticker := time.NewTicker(10 * time.Minute)

	go func() {
		for {
			select {
			case <-ticker.C:
				RefreshAnilistCollectionJob(ctx)
			}
		}
	}()

}
