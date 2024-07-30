//go:build (linux || darwin) && !windows

package app

import (
	"embed"
	"fmt"
	"github.com/rs/zerolog/log"
	golog "log"
	"os"
	"path/filepath"
	"seanime/internal/core"
	"seanime/internal/cron"
	"seanime/internal/handlers"
	"seanime/internal/updater"
	"seanime/internal/util"
	"seanime/internal/util/crashlog"
	"time"
)

func StartApp(webFS embed.FS) {

	// Print the header
	core.PrintHeader()

	// Get the flags
	flags := core.GetSeanimeFlags()

	selfupdater := updater.NewSelfUpdater()

	// Create the app instance
	app := core.NewApp(&core.ConfigOptions{
		DataDir: flags.DataDir,
	}, selfupdater)

	// Create log file
	logFilePath := filepath.Join(app.Config.Logs.Dir, fmt.Sprintf("seanime-%s.log", time.Now().Format("2006-01-02_15-04-05")))
	// Open the log file
	logFile, _ := os.OpenFile(
		logFilePath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0664,
	)

	log.Logger = *app.Logger
	golog.SetOutput(app.Logger)
	util.SetupLoggerSignalHandling(logFile)
	crashlog.GlobalCrashLogger.SetLogDir(app.Config.Logs.Dir)

	updateMode := false
	if flags.Update {
		updateMode = true
	} else {

		go func() {
			for {
				util.WriteGlobalLogBufferToFile(logFile)
				time.Sleep(5 * time.Second)
			}
		}()

	}

appLoop:
	for {
		switch updateMode {
		case true:

			fmt.Println("Running in update mode")

			// Print the header
			core.PrintHeader()

			// Run the self-updater
			err := selfupdater.Run()
			if err != nil {
			}

			fmt.Println("Shutting down in 10 seconds...")
			time.Sleep(10 * time.Second)

			break appLoop
		case false:

			// Create the fiber app instance
			fiberApp := core.NewFiberApp(app, &webFS)

			// Initialize the routes
			handlers.InitRoutes(app, fiberApp)

			// Run the server
			core.RunServer(app, fiberApp)

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
