package cron

import (
	"seanime/internal/core"
	"time"
)

type JobCtx struct {
	App *core.App
}

func RunJobs(app *core.App) {

	// Run the jobs only if the server is online
	ctx := &JobCtx{
		App: app,
	}

	refreshAnilistTicker := time.NewTicker(10 * time.Minute)
	refreshLocalDataTicker := time.NewTicker(30 * time.Minute)
	refetchReleaseTicker := time.NewTicker(1 * time.Hour)
	refetchAnnouncementsTicker := time.NewTicker(10 * time.Minute)

	go func() {
		for {
			select {
			case <-refreshAnilistTicker.C:
				if *app.IsOffline() {
					continue
				}
				RefreshAnilistDataJob(ctx)
				if app.LocalManager != nil &&
					!app.GetUser().IsSimulated &&
					app.Settings != nil &&
					app.Settings.Library != nil &&
					app.Settings.Library.AutoSyncToLocalAccount {
					_ = app.LocalManager.SynchronizeAnilistToSimulatedCollection()
				}
			}
		}
	}()

	go func() {
		for {
			select {
			case <-refreshLocalDataTicker.C:
				if *app.IsOffline() {
					continue
				}
				SyncLocalDataJob(ctx)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-refetchReleaseTicker.C:
				if *app.IsOffline() {
					continue
				}
				app.Updater.ShouldRefetchReleases()
			}
		}
	}()

	go func() {
		for {
			select {
			case <-refetchAnnouncementsTicker.C:
				if *app.IsOffline() {
					continue
				}
				app.Updater.FetchAnnouncements()
			}
		}
	}()

}
