//go:build windows && !nosystray

package server

import (
	"embed"
	"fmt"
	"fyne.io/systray"
	"github.com/cli/browser"
	"github.com/gonutz/w32/v2"
	"github.com/rs/zerolog/log"
	"seanime/internal/constants"
	"seanime/internal/core"
	"seanime/internal/handlers"
	"seanime/internal/icon"
	"seanime/internal/updater"
)

func StartServer(webFS embed.FS, embeddedLogo []byte) {
	onExit := func() {}
	hideConsole()

	app, flags, selfupdater := startApp(embeddedLogo)

	// Blocks until systray.Quit() is called
	systray.Run(onReady(&webFS, app, flags, selfupdater), onExit)
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

func onReady(webFS *embed.FS, app *core.App, flags core.SeanimeFlags, selfupdater *updater.SelfUpdater) func() {
	return func() {
		systray.SetTemplateIcon(icon.Data, icon.Data)
		systray.SetTitle(fmt.Sprintf("Seanime v%s", constants.Version))
		systray.SetTooltip(fmt.Sprintf("Seanime v%s", constants.Version))
		log.Trace().Msg("systray: App is ready")

		// Menu items
		systray.AddMenuItem("Seanime v"+constants.Version, "Seanime version")
		mWeb := systray.AddMenuItem(app.Config.GetServerURI("127.0.0.1"), "Open web interface")
		mOpenLibrary := systray.AddMenuItem("Open Anime Library", "Open anime library")
		mOpenDataDir := systray.AddMenuItem("Open Data Directory", "Open data directory")
		mOpenLogsDir := systray.AddMenuItem("Open Log Directory", "Open log directory")

		addQuitItem()

		go func() {
			// Close the systray when the app exits
			defer systray.Quit()

			startAppLoop(webFS, app, flags, selfupdater)
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
