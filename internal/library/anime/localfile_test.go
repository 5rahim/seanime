package anime_test

import (
	"github.com/stretchr/testify/assert"
	"seanime/internal/library/anime"
	"testing"
)

func TestNewLocalFile(t *testing.T) {

	tests := []struct {
		path              string
		libraryPath       string
		expectedNbFolders int
		expectedFilename  string
	}{
		{
			path:              "E:\\Anime\\Bungou Stray Dogs 5th Season\\[SubsPlease] Bungou Stray Dogs - 61 (1080p) [F609B947].mkv",
			libraryPath:       "E:\\Anime",
			expectedFilename:  "[SubsPlease] Bungou Stray Dogs - 61 (1080p) [F609B947].mkv",
			expectedNbFolders: 1,
		},
		{
			path:              "E:\\Anime\\Bungou Stray Dogs 5th Season\\[SubsPlease] Bungou Stray Dogs - 61 (1080p) [F609B947].mkv",
			libraryPath:       "E:/ANIME",
			expectedFilename:  "[SubsPlease] Bungou Stray Dogs - 61 (1080p) [F609B947].mkv",
			expectedNbFolders: 1,
		},
		{
			path:              "/mnt/Anime/Bungou Stray Dogs/Bungou Stray Dogs 5th Season/[SubsPlease] Bungou Stray Dogs - 61 (1080p) [F609B947].mkv",
			libraryPath:       "/mnt/Anime",
			expectedFilename:  "[SubsPlease] Bungou Stray Dogs - 61 (1080p) [F609B947].mkv",
			expectedNbFolders: 2,
		},
	}

	for _, tt := range tests {

		lf := anime.NewLocalFile(tt.path, tt.libraryPath)

		if assert.NotNil(t, lf) {
			assert.Equal(t, tt.expectedNbFolders, len(lf.ParsedFolderData))
			assert.Equal(t, tt.expectedFilename, lf.Name)
			assert.Empty(t, lf.Metadata)
		}

	}

}
