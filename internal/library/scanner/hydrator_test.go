package scanner

import (
	"context"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/api/anizip"
	"github.com/seanime-app/seanime/internal/library/anime"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/seanime-app/seanime/internal/util/limiter"
	"testing"
)

func TestFileHydrator_HydrateMetadata(t *testing.T) {

	completeMediaCache := anilist.NewCompleteMediaCache()
	anizipCache := anizip.NewCache()
	anilistRateLimiter := limiter.NewAnilistLimiter()
	logger := util.NewLogger()

	anilistClientWrapper := anilist.TestGetMockAnilistClientWrapper()
	animeCollection, err := anilistClientWrapper.AnimeCollectionWithRelations(context.Background(), nil)
	if err != nil {
		t.Fatal("expected result, got error:", err.Error())
	}

	allMedia := animeCollection.GetAllMedia()

	tests := []struct {
		name            string
		paths           []string
		expectedMediaId int
	}{
		{
			name: "should be hydrated with id 131586",
			paths: []string{
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 20v2 (1080p) [30072859].mkv",
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 21v2 (1080p) [4B1616A5].mkv",
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 22v2 (1080p) [58BF43B4].mkv",
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 23v2 (1080p) [D94B4894].mkv",
			},
			expectedMediaId: 131586, // 86 - Eighty Six Part 2
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
				lf := anime.NewLocalFile(path, "E:/Anime")
				lfs = append(lfs, lf)
			}

			// +---------------------+
			// |   MediaContainer    |
			// +---------------------+

			mc := NewMediaContainer(&MediaContainerOptions{
				AllMedia:   allMedia,
				ScanLogger: scanLogger,
			})

			for _, nm := range mc.NormalizedMedia {
				t.Logf("media id: %d, title: %s", nm.ID, nm.GetTitleSafe())
			}

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
			if err != nil {
				t.Fatal("expected result, got error:", err.Error())
			}

			// +---------------------+
			// |    FileHydrator     |
			// +---------------------+

			fh := &FileHydrator{
				LocalFiles:           lfs,
				AllMedia:             mc.NormalizedMedia,
				CompleteMediaCache:   completeMediaCache,
				AnizipCache:          anizipCache,
				AnilistClientWrapper: anilistClientWrapper,
				AnilistRateLimiter:   anilistRateLimiter,
				Logger:               logger,
				ScanLogger:           scanLogger,
			}

			fh.HydrateMetadata()

			for _, lf := range fh.LocalFiles {
				if lf.MediaId != tt.expectedMediaId {
					t.Fatalf("expected media id %d, got %d", tt.expectedMediaId, lf.MediaId)
				}

				t.Logf("local file: %s,\nmedia id: %d\n", lf.Name, lf.MediaId)
			}

		})
	}

}
