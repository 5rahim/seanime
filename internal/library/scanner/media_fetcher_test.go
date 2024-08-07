package scanner

import (
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/api/anizip"
	"github.com/seanime-app/seanime/internal/library/anime"
	"github.com/seanime-app/seanime/internal/test_utils"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/seanime-app/seanime/internal/util/limiter"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewMediaFetcher(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	anilistClientWrapper := anilist.TestGetMockAnilistClientWrapper()
	anizipCache := anizip.NewCache()
	completeMediaCache := anilist.NewCompleteMediaCache()
	anilistRateLimiter := limiter.NewAnilistLimiter()

	dir := "E:/Anime"

	tests := []struct {
		name                   string
		paths                  []string
		enhanced               bool
		disableAnimeCollection bool
	}{
		{
			name: "86 - Eighty Six Part 1 & 2",
			paths: []string{
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 20v2 (1080p) [30072859].mkv",
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 21v2 (1080p) [4B1616A5].mkv",
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 22v2 (1080p) [58BF43B4].mkv",
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 23v2 (1080p) [D94B4894].mkv",
			},
			enhanced:               false,
			disableAnimeCollection: false,
		},
		{
			name: "86 - Eighty Six Part 1 & 2",
			paths: []string{
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 20v2 (1080p) [30072859].mkv",
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 21v2 (1080p) [4B1616A5].mkv",
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 22v2 (1080p) [58BF43B4].mkv",
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 23v2 (1080p) [D94B4894].mkv",
			},
			enhanced:               true,
			disableAnimeCollection: true,
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
			// |    MediaFetcher     |
			// +---------------------+

			mf, err := NewMediaFetcher(&MediaFetcherOptions{
				Enhanced:               tt.enhanced,
				Username:               test_utils.ConfigData.Provider.AnilistUsername,
				AnilistClientWrapper:   anilistClientWrapper,
				LocalFiles:             lfs,
				CompleteMediaCache:     completeMediaCache,
				AnizipCache:            anizipCache,
				Logger:                 util.NewLogger(),
				AnilistRateLimiter:     anilistRateLimiter,
				ScanLogger:             scanLogger,
				DisableAnimeCollection: tt.disableAnimeCollection,
			})
			if err != nil {
				t.Fatal("expected result, got error:", err.Error())
			}

			mc := NewMediaContainer(&MediaContainerOptions{
				AllMedia:   mf.AllMedia,
				ScanLogger: scanLogger,
			})

			for _, m := range mc.NormalizedMedia {
				t.Log(m.GetTitleSafe())
			}

		})

	}

}

func TestNewEnhancedMediaFetcher(t *testing.T) {

	anilistClientWrapper := anilist.TestGetMockAnilistClientWrapper()
	anizipCache := anizip.NewCache()
	completeMediaCache := anilist.NewCompleteMediaCache()
	anilistRateLimiter := limiter.NewAnilistLimiter()

	dir := "E:/Anime"

	tests := []struct {
		name     string
		paths    []string
		enhanced bool
	}{
		{
			name: "86 - Eighty Six Part 1 & 2",
			paths: []string{
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 20v2 (1080p) [30072859].mkv",
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 21v2 (1080p) [4B1616A5].mkv",
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 22v2 (1080p) [58BF43B4].mkv",
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 23v2 (1080p) [D94B4894].mkv",
			},
			enhanced: false,
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
				lf := anime.NewLocalFile(path, dir)
				lfs = append(lfs, lf)
			}

			// +---------------------+
			// |    MediaFetcher     |
			// +---------------------+

			mf, err := NewMediaFetcher(&MediaFetcherOptions{
				Enhanced:             tt.enhanced,
				Username:             "-",
				AnilistClientWrapper: anilistClientWrapper,
				LocalFiles:           lfs,
				CompleteMediaCache:   completeMediaCache,
				AnizipCache:          anizipCache,
				Logger:               util.NewLogger(),
				AnilistRateLimiter:   anilistRateLimiter,
				ScanLogger:           scanLogger,
			})
			if err != nil {
				t.Fatal("expected result, got error:", err.Error())
			}

			mc := NewMediaContainer(&MediaContainerOptions{
				AllMedia:   mf.AllMedia,
				ScanLogger: scanLogger,
			})

			for _, m := range mc.NormalizedMedia {
				t.Log(m.GetTitleSafe())
			}

		})

	}

}

func TestFetchMediaFromLocalFiles(t *testing.T) {

	anilistClientWrapper := anilist.TestGetMockAnilistClientWrapper()
	anizipCache := anizip.NewCache()
	completeMediaCache := anilist.NewCompleteMediaCache()
	anilistRateLimiter := limiter.NewAnilistLimiter()

	tests := []struct {
		name            string
		paths           []string
		expectedMediaId []int
	}{
		{
			name: "86 - Eighty Six Part 1 & 2",
			paths: []string{
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 20v2 (1080p) [30072859].mkv",
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 21v2 (1080p) [4B1616A5].mkv",
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 22v2 (1080p) [58BF43B4].mkv",
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 23v2 (1080p) [D94B4894].mkv",
			},
			expectedMediaId: []int{116589, 131586}, // 86 - Eighty Six Part 1 & 2
		},
	}

	dir := "E:/Anime"

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
				lf := anime.NewLocalFile(path, dir)
				lfs = append(lfs, lf)
			}

			// +--------------------------+
			// | FetchMediaFromLocalFiles |
			// +--------------------------+

			media, ok := FetchMediaFromLocalFiles(
				anilistClientWrapper,
				lfs,
				completeMediaCache,
				anizipCache,
				anilistRateLimiter,
				scanLogger,
			)
			if !ok {
				t.Fatal("could not fetch media from local files")
			}

			ids := lo.Map(media, func(k *anilist.CompleteMedia, _ int) int {
				return k.ID
			})

			// Test if all expected media IDs are present
			for _, id := range tt.expectedMediaId {
				assert.Contains(t, ids, id)
			}

			t.Log("Media IDs:")
			for _, m := range media {
				t.Log(m.GetTitleSafe())
			}

		})
	}

}
