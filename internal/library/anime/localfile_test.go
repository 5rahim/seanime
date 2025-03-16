package anime_test

import (
	"runtime"
	"seanime/internal/library/anime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLocalFile(t *testing.T) {

	tests := []struct {
		path              string
		libraryPath       string
		expectedNbFolders int
		expectedFilename  string
		os                string
	}{
		{
			path:              "E:\\Anime\\Bungou Stray Dogs 5th Season\\[SubsPlease] Bungou Stray Dogs - 61 (1080p) [F609B947].mkv",
			libraryPath:       "E:\\Anime",
			expectedFilename:  "[SubsPlease] Bungou Stray Dogs - 61 (1080p) [F609B947].mkv",
			expectedNbFolders: 1,
			os:                "windows",
		},
		{
			path:              "E:\\Anime\\Bungou Stray Dogs 5th Season\\[SubsPlease] Bungou Stray Dogs - 61 (1080p) [F609B947].mkv",
			libraryPath:       "E:/ANIME",
			expectedFilename:  "[SubsPlease] Bungou Stray Dogs - 61 (1080p) [F609B947].mkv",
			expectedNbFolders: 1,
			os:                "windows",
		},
		{
			path:              "/mnt/Anime/Bungou Stray Dogs/Bungou Stray Dogs 5th Season/[SubsPlease] Bungou Stray Dogs - 61 (1080p) [F609B947].mkv",
			libraryPath:       "/mnt/Anime",
			expectedFilename:  "[SubsPlease] Bungou Stray Dogs - 61 (1080p) [F609B947].mkv",
			expectedNbFolders: 2,
			os:                "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if tt.os != "" {
				if tt.os != runtime.GOOS {
					t.Skipf("skipping test for %s", tt.path)
				}
			}

			lf := anime.NewLocalFile(tt.path, tt.libraryPath)

			if assert.NotNil(t, lf) {
				assert.Equal(t, tt.expectedNbFolders, len(lf.ParsedFolderData))
				assert.Equal(t, tt.expectedFilename, lf.Name)
			}
		})
	}
}
