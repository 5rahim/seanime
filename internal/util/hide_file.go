//go:build !windows

package util

import (
	"os"
	"path/filepath"
	"strings"
)

func HideFile(path string) (string, error) {
	filename := filepath.Base(path)
	if strings.HasPrefix(filename, ".") {
		return path, nil
	}

	newPath := filepath.Join(filepath.Dir(path), "."+filename)
	err := os.Rename(path, newPath)
	if err != nil {
		return "", err
	}

	return newPath, nil
}
