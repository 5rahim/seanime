package util

import (
	"os"
	"path/filepath"
	"runtime"
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
		".vob": {}, ".mts": {}, ".m2ts": {}, ".ts": {}, ".f4v": {}, ".ogm": {}, ".rm": {},
		".rmvb": {}, ".drc": {}, ".yuv": {}, ".asf": {}, ".amv": {}, ".m2v": {}, ".mpe": {},
		".mpv": {}, ".mp2": {}, ".svi": {}, ".mxf": {}, ".roq": {}, ".nsv": {}, ".f4p": {},
		".f4a": {}, ".f4b": {},
	}
	ext = strings.ToLower(ext)
	_, exists := validExtensions[ext]
	return exists
}

func IsSubdirectory(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return rel != "." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}

func IsSubdirectoryOfAny(dirs []string, child string) bool {
	for _, dir := range dirs {
		if IsSubdirectory(dir, child) {
			return true
		}
	}
	return false
}

func IsSameDir(dir1, dir2 string) bool {
	if runtime.GOOS == "windows" {
		dir1 = strings.ToLower(dir1)
		dir2 = strings.ToLower(dir2)
	}

	absDir1, err := filepath.Abs(dir1)
	if err != nil {
		return false
	}
	absDir2, err := filepath.Abs(dir2)
	if err != nil {
		return false
	}
	return absDir1 == absDir2
}
