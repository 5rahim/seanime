package scanner

import (
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/library/anime"
	"seanime/internal/platforms/anilist_platform"
	"seanime/internal/util"
	"seanime/internal/util/limiter"
	"testing"
)

func TestScanLogger(t *testing.T) {

	anilistClient := anilist.TestGetMockAnilistClient()
	logger := util.NewLogger()
	anilistPlatform := anilist_platform.NewAnilistPlatform(anilistClient, logger)
	animeCollection, err := anilistPlatform.GetAnimeCollectionWithRelations(t.Context())
	if err != nil {
		t.Fatal(err.Error())
	}
	allMedia := animeCollection.GetAllAnime()
	metadataProvider := metadata.GetMockProvider(t)
	completeAnimeCache := anilist.NewCompleteAnimeCache()
	anilistRateLimiter := limiter.NewAnilistLimiter()

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

			scanLogger, err := NewScanLogger("./logs")
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
				CompleteAnimeCache: completeAnimeCache,
				Logger:             util.NewLogger(),
				ScanLogger:         scanLogger,
				ScanSummaryLogger:  nil,
			}

			err = matcher.MatchLocalFilesWithMedia()
			if err != nil {
				t.Fatal("expected result, got error:", err.Error())
			}

			// +---------------------+
			// |   FileHydrator      |
			// +---------------------+

			fh := FileHydrator{
				LocalFiles:         lfs,
				AllMedia:           mc.NormalizedMedia,
				CompleteAnimeCache: completeAnimeCache,
				Platform:           anilistPlatform,
				MetadataProvider:   metadataProvider,
				AnilistRateLimiter: anilistRateLimiter,
				Logger:             logger,
				ScanLogger:         scanLogger,
				ScanSummaryLogger:  nil,
				ForceMediaId:       0,
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
