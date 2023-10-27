package filesystem

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
)

type SeparatedFilePath struct {
	Filename string
	Dirnames []string
}

// SeparateFilePath separates a path into a Filename and a slice of Dirnames.
func SeparateFilePath(path string, prefixPath string) *SeparatedFilePath {
	path = filepath.ToSlash(path)
	prefixPath = filepath.ToSlash(prefixPath)
	cleaned := path
	if strings.HasPrefix(strings.ToLower(path), strings.ToLower(prefixPath)) {
		cleaned = path[len(prefixPath):] // Remove prefix
	}
	fp := filepath.ToSlash(filepath.Base(path))
	parentsPath := filepath.ToSlash(filepath.Dir(cleaned))

	return &SeparatedFilePath{
		Filename: fp,
		Dirnames: strings.Split(parentsPath, "/"),
	}
}

////////////////////////////////////////////////////

// GetVideoFilePathsFromDir returns a slice of strings containing the paths of all the video files in a directory.
func GetVideoFilePathsFromDir(dirPath string) ([]string, error) {
	filePaths := make([]string, 0)

	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		ext := filepath.Ext(path)

		if !d.IsDir() && (ext == ".mkv" || ext == ".mp4") {
			filePaths = append(filePaths, path)
		}
		return nil
	})

	if err != nil {
		return nil, errors.New("could not find the local directory")
	}

	return filePaths, nil
}
