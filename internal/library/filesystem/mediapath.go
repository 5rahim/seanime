package filesystem

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"seanime/internal/util"
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

// GetMediaFilePathsFromDir returns a slice of strings containing the paths of all the media files in a directory.
// DEPRECATED: Use GetMediaFilePathsFromDirS instead.
func GetMediaFilePathsFromDir(dirPath string) ([]string, error) {
	filePaths := make([]string, 0)

	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		ext := strings.ToLower(filepath.Ext(path))

		if !d.IsDir() && util.IsValidVideoExtension(ext) {
			filePaths = append(filePaths, path)
		}
		return nil
	})

	if err != nil {
		return nil, errors.New("could not traverse the local directory")
	}

	return filePaths, nil
}

// GetMediaFilePathsFromDirS returns a slice of strings containing the paths of all the video files in a directory.
// Unlike GetMediaFilePathsFromDir, it follows symlinks.
func GetMediaFilePathsFromDirS(oDirPath string) ([]string, error) {
	filePaths := make([]string, 0)
	visited := make(map[string]bool)

	// Normalize the initial directory path
	dirPath, err := filepath.Abs(oDirPath)
	if err != nil {
		return nil, fmt.Errorf("could not resolve path: %w", err)
	}

	var walkDir func(string) error
	walkDir = func(oCurrentPath string) error {
		// Normalize current path
		currentPath, err := filepath.EvalSymlinks(oCurrentPath)
		if err != nil {
			return fmt.Errorf("could not evaluate symlink: %w", err)
		}

		if visited[currentPath] {
			return nil
		}
		visited[currentPath] = true

		return filepath.WalkDir(currentPath, func(path string, d fs.DirEntry, err error) error {

			if err != nil {
				return err
			}

			// If it's a symlink directory, resolve and walk the symlink
			info, err := os.Lstat(path)
			if err != nil {
				return fmt.Errorf("could not get file info: %w", err)
			}

			if info.Mode()&os.ModeSymlink != 0 {
				linkPath, err := os.Readlink(path)
				if err != nil {
					return fmt.Errorf("could not read symlink: %w", err)
				}

				// Resolve the symlink to an absolute path
				if !filepath.IsAbs(linkPath) {
					linkPath = filepath.Join(filepath.Dir(path), linkPath)
				}

				return walkDir(linkPath)
			}

			if d.IsDir() {
				return nil
			}

			ext := strings.ToLower(filepath.Ext(path))
			if util.IsValidMediaFile(path) && util.IsValidVideoExtension(ext) {
				filePaths = append(filePaths, path)
			}
			return nil
		})
	}

	if err = walkDir(dirPath); err != nil {
		return nil, fmt.Errorf("could not traverse directory %s: %w", dirPath, err)
	}

	return filePaths, nil
}

//----------------------------------------------------------------------------------------------------------------------

func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !errors.Is(err, os.ErrNotExist)
}
