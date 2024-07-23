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
	"os"
	"path/filepath"
	"seanime/internal/core"
	"seanime/internal/cron"
	"seanime/internal/handlers"
	"seanime/internal/icon"
	"seanime/internal/updater"
	"seanime/internal/util"
	"time"

	"github.com/gonutz/w32/v2"
)

// hideConsole will hide the terminal window if the app was not started with the -H=windowsgui flag.
// NOTE: This will only minimize the terminal window on Windows 11 if the 'default terminal app' is set to 'Windows Terminal' or 'Let Windows choose' instead of 'Windows Console Host'
func hideConsole() {
	console := w32.GetConsoleWindow()
	if console == 0 {
		return // no console attached
	}
	// If this application is the process that created the console window, then
	// this program was not compiled with the -H=windowsgui flag and on start-up
	// it created a console along with the main application window. In this case
	// hide the console window.
	// See
	// http://stackoverflow.com/questions/9009333/how-to-check-if-the-program-is-run-from-a-console
	_, consoleProcID := w32.GetWindowThreadProcessId(console)
	if w32.GetCurrentProcessId() == consoleProcID {
		w32.ShowWindow(console, w32.SW_HIDE)
	}
}

func setupSignalHandling(file *os.File) {
	if file == nil {
		return
	}

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

	hideConsole()

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

	setupSignalHandling(logFile)

	// Flush log buffer to the log file when the app exits
	defer func() {
		if r := recover(); r != nil {
			log.Error().Msgf("Recovered from panic: %v", r)
			util.WriteGlobalLogBufferToFile(logFile)
			_ = logFile.Close()
			os.Exit(1)
		} else {
			// Ensure buffer is flushed on normal exit
			util.WriteGlobalLogBufferToFile(logFile)
			_ = logFile.Close()
		}
	}()

	go func() {
		for {
			util.WriteGlobalLogBufferToFile(logFile)
			time.Sleep(5 * time.Second)
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
		mWeb := systray.AddMenuItem(app.Config.GetServerURI("127.0.0.1"), "Open web interface")
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
