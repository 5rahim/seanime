package core

import (
	"seanime/internal/library/scanner"
	"seanime/internal/util"
	"sync"
)

// initLibraryWatcher will initialize the library watcher.
//   - Used by AutoScanner
func (a *App) initLibraryWatcher(paths []string) {
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
		LibraryPaths: paths,
	})
	if err != nil {
		a.Logger.Error().Err(err).Msg("app: Failed to watch library files")
		return
	}

	var dirSize uint64 = 0
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	for _, path := range paths {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			ds, _ := util.DirSize(path)
			mu.Lock()
			dirSize += ds
			mu.Unlock()
		}(path)
	}
	wg.Wait()
	a.TotalLibrarySize = dirSize

	a.Logger.Info().Msgf("watcher: Library size: %s", util.Bytes(dirSize))

	// Set the watcher
	a.Watcher = watcher

	// Start watching
	a.Watcher.StartWatching(
		func() {
			// Notify the auto scanner when a file action occurs
			a.AutoScanner.Notify()
		})

}
