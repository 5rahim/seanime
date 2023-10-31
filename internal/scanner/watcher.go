package scanner

import (
	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog"
	"log"
	"os"
	"path/filepath"
)

// Watcher is a custom file system event watcher
type Watcher struct {
	watcher *fsnotify.Watcher
	logger  *zerolog.Logger
}

type NewWatcherOptions struct {
	Logger *zerolog.Logger
}

// NewWatcher creates a new Watcher instance for monitoring a directory and its subdirectories
func NewWatcher(opts *NewWatcherOptions) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &Watcher{
		watcher: watcher,
		logger:  opts.Logger,
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
				return w.watcher.Add(path)
			}
			return nil
		})
		return err
	}

	// Add the initial directory and its subdirectories to the watcher
	if err := watchDir(opts.LibraryPath); err != nil {
		return err
	}

	w.logger.Info().Msgf("watcher: Watching directory: \"%s\"", opts.LibraryPath)

	return nil
}

func (w *Watcher) StartWatching() {
	// Start a goroutine to handle file system events
	go func() {
		for {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Printf("File modified: %s", event.Name)
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					log.Printf("File created: %s", event.Name)
				}
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					log.Printf("File removed: %s", event.Name)
				}
				// You can add more event types as needed

			case err, ok := <-w.watcher.Errors:
				if !ok {
					return
				}
				w.logger.Warn().Err(err).Msgf("watcher: Error while watching")
			}
		}
	}()
}

func (w *Watcher) StopWatching() {
	err := w.watcher.Close()
	if err == nil {
		w.logger.Debug().Err(err).Msgf("watcher: Watcher is closed")
	}
}
