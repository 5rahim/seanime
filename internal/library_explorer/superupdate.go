package library_explorer

import (
	"os"
	"path/filepath"
	"seanime/internal/database/db_bridge"
	"seanime/internal/library/anime"
	"sync"

	"github.com/samber/lo"
)

type SuperUpdateFileOptions struct {
	// Path to the file
	Path string `json:"path"`
	// New name of the file
	NewName string `json:"newName,omitempty"`
	// Metadata of the file
	Metadata *anime.LocalFileMetadata `json:"metadata,omitempty"`
}

func (l *LibraryExplorer) SuperUpdateFiles(opts []*SuperUpdateFileOptions) error {

	const MaxConcurrentUpdates = 10
	sem := make(chan struct{}, MaxConcurrentUpdates)

	l.logger.Debug().
		Int("count", len(opts)).
		Msg("library explorer: Updating files")

	wg := sync.WaitGroup{}
	wg.Add(len(opts))

	lfs, lfsId, err := db_bridge.GetLocalFiles(l.database)
	if err != nil {
		return err
	}

	settings, err := l.database.GetSettings()
	if err != nil {
		return err
	}

	for _, opt := range opts {
		go func(opt *SuperUpdateFileOptions) {
			sem <- struct{}{}
			defer func() { <-sem }()
			defer wg.Done()
			_ = l.superUpdateFile(opt, lfs, lfsId, settings.GetLibrary().GetLibraryPaths())
		}(opt)
	}

	wg.Wait()

	// Save the local files
	_, err = db_bridge.SaveLocalFiles(l.database, lfsId, lfs)
	if err != nil {
		return err
	}

	l.fileTree = nil

	return nil
}

func (l *LibraryExplorer) superUpdateFile(opt *SuperUpdateFileOptions, lfs []*anime.LocalFile, lfsId uint, libraryPaths []string) error {

	l.logger.Debug().
		Any("path", opt.Path).
		Msg("library explorer: Updating file")

	lf, found := lo.Find(lfs, func(i *anime.LocalFile) bool {
		return i.HasSamePath(opt.Path)
	})

	if opt.NewName != "" {
		newPath := filepath.Join(filepath.Dir(opt.Path), opt.NewName)
		// Update the file name
		// If the local file exists, update the name
		if found {
			lf.Name = opt.NewName
			// Update the parsed info
			newLf := anime.NewLocalFileS(newPath, libraryPaths)
			lf.ParsedData = newLf.ParsedData
			lf.ParsedFolderData = newLf.ParsedFolderData
			lf.Path = newPath
		}

		// Rename the real file name
		err := os.Rename(opt.Path, newPath)
		if err != nil {
			return err
		}
	}

	if opt.Metadata != nil {
		l.logger.Debug().
			Any("path", opt.Path).
			Any("metadata", opt.Metadata).
			Msg("library explorer: Updating file metadata")
		if found {
			lf.Metadata = opt.Metadata
			lf.Locked = true
			lf.Ignored = false
		}
	}

	return nil
}
