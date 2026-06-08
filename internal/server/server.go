package server

import (
	"embed"
	"seanime/internal/core"
	"seanime/internal/cron"
	"seanime/internal/handlers"
	"seanime/internal/updater"
	"time"

	"github.com/rs/zerolog/log"
)

func startApp(embeddedLogo []byte) (*core.App, core.SeanimeFlags, *updater.SelfUpdater) {
	// Print the header
	core.PrintHeader()

	// Get the flags
	flags := core.GetSeanimeFlags()

	selfupdater := updater.NewSelfUpdater()

	// Create the app instance
	app := core.NewApp(&core.ConfigOptions{
		Flags:        flags,
		EmbeddedLogo: embeddedLogo,
	}, selfupdater)

	app.InitLogging(flags.Update)

	return app, flags, selfupdater
}

func startAppLoop(webFS *embed.FS, app *core.App, flags core.SeanimeFlags, selfupdater *updater.SelfUpdater) {
	updateMode := flags.Update

appLoop:
	for {
		switch updateMode {
		case true:

			log.Log().Msg("Running in update mode")

			// Print the header
			core.PrintHeader()

			// Run the self-updater
			err := selfupdater.Run()
			if err != nil {
			}

			log.Log().Msg("Shutting down in 10 seconds...")
			time.Sleep(10 * time.Second)

			break appLoop
		case false:

			// Create the echo app instance
			echoApp := core.NewEchoApp(app, webFS)

			// Initialize the routes
			handlers.InitRoutes(app, echoApp)

			// Run the server
			core.RunEchoServer(app, echoApp)

			// Run the jobs in the background
			cron.RunJobs(app)

			select {
			case <-selfupdater.Started():
				app.Cleanup()
				updateMode = true
				break
			}
		}
		continue
	}
}
