package core

import (
	"github.com/dustin/go-humanize"
	"github.com/seanime-app/seanime/internal/library/scanner"
	"github.com/seanime-app/seanime/internal/util"
)

// initLibraryWatcher will initialize the library watcher.
//   - Used by AutoScanner
func (a *App) initLibraryWatcher(path string) {
	// Create a new watcher
	watcher, err := scanner.NewWatcher(&scanner.NewWatcherOptions{
		Logger:         a.Logger,
		WSEventManager: a.WSEventManager,
	})
	if err != nil {
		a.Logger.Error().Err(err).Msg("app: Failed to initialize watcher")
		return
	}

	// Initialize library file watcher
	err = watcher.InitLibraryFileWatcher(&scanner.WatchLibraryFilesOptions{
		LibraryPath: path,
	})
	if err != nil {
		a.Logger.Error().Err(err).Msg("app: Failed to watch library files")
		return
	}

	dirSize, _ := util.DirSize(path)
	a.TotalLibrarySize = dirSize

	a.Logger.Info().Msgf("app: Library size: %s", humanize.Bytes(dirSize))

	// Set the watcher
	a.Watcher = watcher

	// Start watching
	a.Watcher.StartWatching(
		func() {
			// Notify the auto scanner when a file action occurs
			a.AutoScanner.Notify()
		})

}
