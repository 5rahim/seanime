package util

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func DownloadDir() (string, error) {
	return userDir("Downloads")
}

func DesktopDir() (string, error) {
	return userDir("Desktop")
}

func DocumentsDir() (string, error) {
	return userDir("Documents")
}

// userDir returns the path to the specified user directory (Desktop or Documents).
func userDir(dirType string) (string, error) {
	var dir string
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "windows":
		dir = filepath.Join(home, dirType)

	case "darwin":
		dir = filepath.Join(home, dirType)

	case "linux":
		// Linux: Use $XDG_DESKTOP_DIR / $XDG_DOCUMENTS_DIR / $XDG_DOWNLOAD_DIR if set, otherwise default
		envVar := ""
		if dirType == "Desktop" {
			envVar = os.Getenv("XDG_DESKTOP_DIR")
		} else if dirType == "Documents" {
			envVar = os.Getenv("XDG_DOCUMENTS_DIR")
		} else if dirType == "Downloads" {
			envVar = os.Getenv("XDG_DOWNLOAD_DIR")
		}

		if envVar != "" {
			dir = envVar
		} else {
			dir = filepath.Join(home, dirType)
		}

	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return dir, nil
}
