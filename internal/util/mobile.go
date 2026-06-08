package util

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func IsMobile() bool {
	return runtime.GOOS == "android" || runtime.GOOS == "ios"
}

func IsIOS() bool {
	return runtime.GOOS == "ios"
}

func IsAndroid() bool {
	return runtime.GOOS == "android"
}

func EnsureDocumentsDirectoryVisible() (string, error) {
	if !IsIOS() {
		return "", nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	docsPath := filepath.Join(homeDir, "Documents", "Anime")

	if err := os.MkdirAll(docsPath, 0755); err != nil {
		return "", err
	}

	initFile := filepath.Join(docsPath, ".seanime-keep")
	if _, err := os.Stat(initFile); os.IsNotExist(err) {
		file, err := os.Create(initFile)
		if err != nil {
			return "", err
		}
		file.Close()
	}

	return docsPath, nil
}

// ResolvePhysicalPath resolves a virtual path (e.g. "/Anime") to its physical location on iOS.
// If the path is empty, it returns the Documents/data directory.
// On non-iOS systems, it returns the path unmodified.
func ResolvePhysicalPath(path string) string {
	if !IsIOS() {
		return path
	}
	dataDir := os.Getenv("SEANIME_DATA_DIR")
	if dataDir == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			dataDir = filepath.Join(homeDir, "Documents")
		}
	}
	if dataDir == "" {
		return path
	}

	cleanPath := filepath.ToSlash(filepath.Clean(path))
	if cleanPath == "" || cleanPath == "." || cleanPath == "/" {
		return dataDir
	}

	// If the path already starts with the dataDir prefix, return it as is
	if strings.HasPrefix(cleanPath, filepath.ToSlash(dataDir)) {
		return cleanPath
	}

	// Remove leading slash to join properly
	rel := strings.TrimPrefix(cleanPath, "/")
	return filepath.ToSlash(filepath.Join(dataDir, rel))
}

// ResolveVirtualPath resolves a physical path (e.g. "/var/mobile/.../Documents/Anime") to its virtual representation (e.g. "/Anime") on iOS.
// On non-iOS systems, it returns the path unmodified.
func ResolveVirtualPath(path string) string {
	if !IsIOS() {
		return path
	}
	dataDir := os.Getenv("SEANIME_DATA_DIR")
	if dataDir == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			dataDir = filepath.Join(homeDir, "Documents")
		}
	}
	if dataDir == "" {
		return path
	}

	cleanPath := filepath.ToSlash(filepath.Clean(path))
	cleanDataDir := filepath.ToSlash(filepath.Clean(dataDir))

	if cleanPath == cleanDataDir {
		return "/"
	}

	if strings.HasPrefix(cleanPath, cleanDataDir) {
		rel := strings.TrimPrefix(cleanPath, cleanDataDir)
		if !strings.HasPrefix(rel, "/") {
			rel = "/" + rel
		}
		return filepath.ToSlash(filepath.Clean(rel))
	}

	return path
}
