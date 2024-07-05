package util

import (
	"os"
	"path/filepath"
	"strings"
)

func DirSize(path string) (uint64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return uint64(size), err
}

func IsValidVideoExtension(ext string) bool {
	validExtensions := map[string]struct{}{
		".mp4": {}, ".avi": {}, ".mkv": {}, ".mov": {}, ".flv": {}, ".wmv": {}, ".webm": {},
		".mpeg": {}, ".mpg": {}, ".m4v": {}, ".3gp": {}, ".3g2": {}, ".ogg": {}, ".ogv": {},
		".vob": {}, ".mts": {}, ".m2ts": {}, ".ts": {}, ".f4v": {},
	}
	ext = strings.ToLower(ext)
	_, exists := validExtensions[ext]
	return exists
}
