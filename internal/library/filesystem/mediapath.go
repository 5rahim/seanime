package filesystem

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"seanime/internal/util"
	"sort"
	"strings"
)

type SeparatedFilePath struct {
	Filename   string
	Dirnames   []string
	PrefixPath string
}

// SeparateFilePath separates a path into a filename and a slice of dirnames while ignoring the prefix.
func SeparateFilePath(path string, prefixPath string) *SeparatedFilePath {
	path = filepath.ToSlash(path)
	prefixPath = filepath.ToSlash(prefixPath)
	cleaned := path
	if strings.HasPrefix(strings.ToLower(path), strings.ToLower(prefixPath)) {
		cleaned = path[len(prefixPath):] // Remove prefix
	}
	fp := filepath.Base(filepath.ToSlash(path))
	parentsPath := filepath.Dir(filepath.ToSlash(cleaned))
	if parentsPath == "." || parentsPath == "/" || parentsPath == ".." {
		parentsPath = ""
	}

	return &SeparatedFilePath{
		Filename:   fp,
		Dirnames:   strings.Split(parentsPath, "/"),
		PrefixPath: prefixPath,
	}
}

// SeparateFilePathS separates a path into a filename and a slice of dirnames while ignoring the prefix.
// Unlike [SeparateFilePath], it will check multiple prefixes.
//
// Example:
//
//	path = "/path/to/file.mkv"
//	potentialPrefixes = []string{"/path/to", "/path"}
//	fp, dirs := SeparateFilePathS(path, potentialPrefixes)
//	fmt.Println(fp) // file.mkv
//	fmt.Println(dirs) // [to]
func SeparateFilePathS(path string, potentialPrefixes []string) *SeparatedFilePath {
	// Sort prefix paths by length in descending order
	sort.Slice(potentialPrefixes, func(i, j int) bool {
		return len(potentialPrefixes[i]) > len(potentialPrefixes[j])
	})

	// Check each prefix path, and remove the first match from the path
	prefixPath := ""
	for _, p := range potentialPrefixes {
		// Normalize the paths for comparison only
		if strings.HasPrefix(util.NormalizePath(path), util.NormalizePath(p)) {
			// Remove the prefix from the path
			path = path[len(p):]
			prefixPath = p
			break
		}
	}

	filename := filepath.ToSlash(filepath.Base(path))
	parentsPath := filepath.ToSlash(filepath.Dir(filepath.ToSlash(path)))

	dirs := make([]string, 0)
	for _, dir := range strings.Split(parentsPath, "/") {
		if dir != "" {
			dirs = append(dirs, dir)
		}
	}

	return &SeparatedFilePath{
		Filename:   filename,
		Dirnames:   dirs,
		PrefixPath: prefixPath,
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

		currentPath := oCurrentPath

		// Normalize current path
		resolvedPath, err := filepath.EvalSymlinks(oCurrentPath)
		if err == nil {
			currentPath = resolvedPath
		}

		if visited[currentPath] {
			return nil
		}
		visited[currentPath] = true

		return filepath.WalkDir(currentPath, func(path string, d fs.DirEntry, err error) error {

			if err != nil {
				return nil
			}

			// If it's a symlink directory, resolve and walk the symlink
			info, err := os.Lstat(path)
			if err != nil {
				return nil
			}

			if info.Mode()&os.ModeSymlink != 0 {
				linkPath, err := os.Readlink(path)
				if err != nil {
					return nil
				}

				// Resolve the symlink to an absolute path
				if !filepath.IsAbs(linkPath) {
					linkPath = filepath.Join(filepath.Dir(path), linkPath)
				}

				// Only follow the symlink if we can access it
				if _, err := os.Stat(linkPath); err == nil {
					return walkDir(linkPath)
				}
				return nil
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
