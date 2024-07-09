package scanner

import (
	"context"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/library/anime"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Add more media to this file if needed
// scanner_test_mock_data.json

func TestMatcher_MatchLocalFileWithMedia(t *testing.T) {

	anilistClientWrapper := anilist.TestGetMockAnilistClientWrapper()
	animeCollection, err := anilistClientWrapper.AnimeCollectionWithRelations(context.Background(), nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	allMedia := animeCollection.GetAllMedia()

	dir := "E:/Anime"

	tests := []struct {
		name            string
		paths           []string
		expectedMediaId int
	}{
		{
			// These local files are from "86 - Eighty Six Part 2" but should be matched with "86 - Eighty Six Part 1"
			// because there is no indication for the part. However, the FileHydrator will fix this issue.
			name: "should match with media id 116589",
			paths: []string{
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 20v2 (1080p) [30072859].mkv",
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 21v2 (1080p) [4B1616A5].mkv",
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 22v2 (1080p) [58BF43B4].mkv",
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 23v2 (1080p) [D94B4894].mkv",
			},
			expectedMediaId: 116589, // 86 - Eighty Six Part 1
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			scanLogger, err := NewConsoleScanLogger()
			if err != nil {
				t.Fatal("expected result, got error:", err.Error())
			}

			// +---------------------+
			// |   Local Files       |
			// +---------------------+

			var lfs []*anime.LocalFile
			for _, path := range tt.paths {
				lf := anime.NewLocalFile(path, dir)
				lfs = append(lfs, lf)
			}

			// +---------------------+
			// |   MediaContainer    |
			// +---------------------+

			mc := NewMediaContainer(&MediaContainerOptions{
				AllMedia:   allMedia,
				ScanLogger: scanLogger,
			})

			// +---------------------+
			// |      Matcher        |
			// +---------------------+

			matcher := &Matcher{
				LocalFiles:         lfs,
				MediaContainer:     mc,
				CompleteMediaCache: nil,
				Logger:             util.NewLogger(),
				ScanLogger:         scanLogger,
			}

			err = matcher.MatchLocalFilesWithMedia()

			if assert.NoError(t, err, "Error while matching local files") {
				for _, lf := range lfs {
					if lf.MediaId != tt.expectedMediaId {
						t.Fatalf("expected media id %d, got %d", tt.expectedMediaId, lf.MediaId)
					}
					t.Logf("local file: %s,\nmedia id: %d\n", lf.Name, lf.MediaId)
				}
			}
		})
	}

}
