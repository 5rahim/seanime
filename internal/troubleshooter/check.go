package troubleshooter

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// IsExecutable checks if a given path points to an executable file or if a command exists in PATH
func IsExecutable(name string) (string, error) {
	// If name contains any path separators, treat it as a path
	if strings.Contains(name, string(os.PathSeparator)) {
		path, err := filepath.Abs(name)
		if err != nil {
			return "", err
		}
		return checkExecutable(path)
	}

	// Otherwise, search in PATH
	return findInPath(name)
}

// findInPath searches for an executable in the system's PATH
func findInPath(name string) (string, error) {
	// On Windows, also check for .exe extension if not provided
	if runtime.GOOS == "windows" && !strings.HasSuffix(strings.ToLower(name), ".exe") {
		name += ".exe"
	}

	// Get system PATH
	pathEnv := os.Getenv("PATH")
	paths := strings.Split(pathEnv, string(os.PathListSeparator))

	// Search each directory in PATH
	for _, dir := range paths {
		if dir == "" {
			continue
		}
		path := filepath.Join(dir, name)
		fullPath, err := checkExecutable(path)
		if err == nil {
			return fullPath, nil
		}
	}

	return "", errors.New("executable not found in PATH")
}

// checkExecutable verifies if a given path points to an executable file
func checkExecutable(path string) (string, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	if fileInfo.IsDir() {
		return "", errors.New("path points to a directory")
	}

	// On Windows, just check if the file exists (as Windows uses file extensions)
	if runtime.GOOS == "windows" {
		return path, nil
	}

	// On Unix-like systems, check if the file is executable
	if fileInfo.Mode()&0111 != 0 {
		return path, nil
	}

	return "", errors.New("file is not executable")
}
