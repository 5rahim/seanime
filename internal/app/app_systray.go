//go:build windows && !(linux || darwin)

package app

import (
	golog "log"
	"os/signal"
	"syscall"

	"embed"
	"fmt"
	"fyne.io/systray"
	"github.com/cli/browser"
	"github.com/rs/zerolog/log"
	"github.com/seanime-app/seanime/internal/core"
	"github.com/seanime-app/seanime/internal/cron"
	"github.com/seanime-app/seanime/internal/handlers"
	"github.com/seanime-app/seanime/internal/icon"
	"github.com/seanime-app/seanime/internal/updater"
	"github.com/seanime-app/seanime/internal/util"
	"os"
	"path/filepath"
	"time"
)

func setupSignalHandling(file *os.File) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Trace().Msgf("Received signal: %s", sig)
		// Flush log buffer to the log file when the app exits
		util.WriteGlobalLogBufferToFile(file)
		_ = file.Close()
		os.Exit(0)
	}()
}

func StartApp(webFS embed.FS) {
	onExit := func() {}

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
	// Delete if log file already exists
	_ = os.Remove(logFilePath)
	// Open the log file
	logFile, err := os.OpenFile(
		logFilePath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0664,
	)
	if err != nil {
		return
	}

	log.Logger = *app.Logger
	golog.SetOutput(app.Logger)

	setupSignalHandling(logFile)

	// Flush log buffer to the log file when the app exits
	defer func() {
		if r := recover(); r != nil {
			log.Error().Msgf("Recovered from panic: %v", r)
			util.WriteGlobalLogBufferToFile(logFile)
			_ = logFile.Close()
			os.Exit(1) // Exit with an error code
		} else {
			// Ensure buffer is flushed on normal exit
			util.WriteGlobalLogBufferToFile(logFile)
			_ = logFile.Close()
		}
	}()

	// Blocks until systray.Quit() is called
	systray.Run(onReady(webFS, app, flags, selfupdater), onExit)
}

func addQuitItem() {
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit Seanime", "Quit the whole app")
	mQuit.Enable()
	go func() {
		<-mQuit.ClickedCh
		log.Trace().Msg("systray: Quitting system tray")
		systray.Quit()
		log.Trace().Msg("systray: Quit system tray")
	}()
}

func onReady(webFS embed.FS, app *core.App, flags core.SeanimeFlags, selfupdater *updater.SelfUpdater) func() {
	return func() {
		systray.SetTemplateIcon(icon.Data, icon.Data)
		systray.SetTitle("Seanime")
		systray.SetTooltip("Seanime")
		log.Trace().Msg("systray: App is ready")

		// Menu items
		mWeb := systray.AddMenuItem("Open Web Interface", "Open web interface")
		mOpenLibrary := systray.AddMenuItem("Open Anime Library", "Open anime library")
		mOpenDataDir := systray.AddMenuItem("Open Data Directory", "Open data directory")
		mOpenLogsDir := systray.AddMenuItem("Open Logs Directory", "Open logs directory")

		addQuitItem()

		go func() {
			// Close the systray when the app exits
			defer systray.Quit()

			updateMode := false
			if flags.Update {
				updateMode = true
			}

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
		}()

		go func() {
			for {
				select {
				case <-mWeb.ClickedCh:
					_ = browser.OpenURL(app.Config.GetServerURI("127.0.0.1"))
				case <-mOpenLibrary.ClickedCh:
					handlers.OpenDirInExplorer(app.LibraryDir)
				case <-mOpenDataDir.ClickedCh:
					handlers.OpenDirInExplorer(app.Config.Data.AppDataDir)
				case <-mOpenLogsDir.ClickedCh:
					handlers.OpenDirInExplorer(app.Config.Logs.Dir)
				}
			}
		}()
	}
}
