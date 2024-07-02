package scanner

import (
	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/events"
	"os"
	"path/filepath"
)

// Watcher is a custom file system event watcher
type Watcher struct {
	Watcher        *fsnotify.Watcher
	Logger         *zerolog.Logger
	WSEventManager events.WSEventManagerInterface
	TotalSize      string
}

type NewWatcherOptions struct {
	Logger         *zerolog.Logger
	WSEventManager events.WSEventManagerInterface
}

// NewWatcher creates a new Watcher instance for monitoring a directory and its subdirectories
func NewWatcher(opts *NewWatcherOptions) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &Watcher{
		Watcher:        watcher,
		Logger:         opts.Logger,
		WSEventManager: opts.WSEventManager,
	}, nil
}

//----------------------------------------------------------------------------------------------------------------------

type WatchLibraryFilesOptions struct {
	LibraryPath string
}

// InitLibraryFileWatcher starts watching the specified directory and its subdirectories for file system events
func (w *Watcher) InitLibraryFileWatcher(opts *WatchLibraryFilesOptions) error {
	// Define a function to add directories and their subdirectories to the watcher
	watchDir := func(dir string) error {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return w.Watcher.Add(path)
			}
			return nil
		})
		return err
	}

	// Add the initial directory and its subdirectories to the watcher
	if err := watchDir(opts.LibraryPath); err != nil {
		return err
	}

	w.Logger.Info().Msgf("watcher: Watching directory: \"%s\"", opts.LibraryPath)

	return nil
}

func (w *Watcher) StartWatching(
	onFileAction func(),
) {
	// Start a goroutine to handle file system events
	go func() {
		for {
			select {
			case event, ok := <-w.Watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					w.Logger.Debug().Msgf("File created: %s", event.Name)
					w.WSEventManager.SendEvent(events.LibraryWatcherFileAdded, event.Name)
					onFileAction()
				}
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					w.Logger.Debug().Msgf("File removed: %s", event.Name)
					w.WSEventManager.SendEvent(events.LibraryWatcherFileRemoved, event.Name)
					onFileAction()
				}

			case err, ok := <-w.Watcher.Errors:
				if !ok {
					return
				}
				w.Logger.Warn().Err(err).Msgf("watcher: Error while watching directory")
			}
		}
	}()
}

func (w *Watcher) StopWatching() {
	err := w.Watcher.Close()
	if err == nil {
		w.Logger.Trace().Err(err).Msgf("watcher: Watcher stopped")
	}
}
