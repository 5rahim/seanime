package core

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

func TrimLogEntries(dir string, logger *zerolog.Logger) {
	// Get all log files in the directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		logger.Error().Err(err).Msg("core: Failed to read log directory")
		return
	}

	// Get the total size of all log entries
	var totalSize int64
	for _, file := range entries {
		if file.IsDir() {
			continue
		}
		info, err := file.Info()
		if err != nil {
			continue
		}
		totalSize += info.Size()
	}

	var files []os.FileInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, info)
	}

	var serverLogFiles []os.FileInfo
	var scanLogFiles []os.FileInfo

	for _, file := range files {
		if strings.HasPrefix(file.Name(), "seanime-") {
			serverLogFiles = append(serverLogFiles, file)
		} else if strings.Contains(file.Name(), "-scan") {
			scanLogFiles = append(scanLogFiles, file)
		}
	}

	for _, _files := range [][]os.FileInfo{serverLogFiles, scanLogFiles} {
		files := _files
		if len(files) <= 1 {
			continue
		}

		// Sort from newest to oldest
		sort.Slice(files, func(i, j int) bool {
			return files[i].ModTime().After(files[j].ModTime())
		})

		// Delete all log files older than 14 days
		deleted := 0
		for i := 1; i < len(files); i++ {
			if time.Since(files[i].ModTime()) > 14*24*time.Hour {
				err := os.Remove(filepath.Join(dir, files[i].Name()))
				if err != nil {
					continue
				}
				deleted++
			}
		}
		if deleted > 0 {
			logger.Info().Msgf("app: Deleted %d log files older than 14 days", deleted)
		}
	}

}
