package scanner

import (
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/anizip"
	"github.com/seanime-app/seanime/internal/entities"
	"github.com/seanime-app/seanime/internal/limiter"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewMediaFetcher(t *testing.T) {

	anilistClientWrapper, _, data := anilist.MockAnilistClientWrappers()
	anizipCache := anizip.NewCache()
	baseMediaCache := anilist.NewBaseMediaCache()
	anilistRateLimiter := limiter.NewAnilistLimiter()

	dir := "E:/Anime"

	tests := []struct {
		name                 string
		paths                []string
		enhanced             bool
		username             string
		jwt                  string
		anilistClientWrapper *anilist.ClientWrapper
		useAnilistCollection bool
	}{
		{
			name: "86 - Eighty Six Part 1 & 2",
			paths: []string{
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 20v2 (1080p) [30072859].mkv",
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 21v2 (1080p) [4B1616A5].mkv",
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 22v2 (1080p) [58BF43B4].mkv",
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 23v2 (1080p) [D94B4894].mkv",
			},
			enhanced:             false,
			username:             data.Username,
			jwt:                  data.JWT,
			anilistClientWrapper: anilistClientWrapper,
			useAnilistCollection: true,
		},
		{
			name: "86 - Eighty Six Part 1 & 2",
			paths: []string{
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 20v2 (1080p) [30072859].mkv",
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 21v2 (1080p) [4B1616A5].mkv",
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 22v2 (1080p) [58BF43B4].mkv",
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 23v2 (1080p) [D94B4894].mkv",
			},
			enhanced:             true,
			username:             data.Username,
			jwt:                  data.JWT,
			anilistClientWrapper: anilistClientWrapper,
			useAnilistCollection: false,
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

			var lfs []*entities.LocalFile
			for _, path := range tt.paths {
				lf := entities.NewLocalFile(path, dir)
				lfs = append(lfs, lf)
			}

			// +---------------------+
			// |    MediaFetcher     |
			// +---------------------+

			mf, err := NewMediaFetcher(&MediaFetcherOptions{
				Enhanced:             tt.enhanced,
				Username:             tt.username,
				AnilistClientWrapper: tt.anilistClientWrapper,
				LocalFiles:           lfs,
				BaseMediaCache:       baseMediaCache,
				AnizipCache:          anizipCache,
				Logger:               util.NewLogger(),
				AnilistRateLimiter:   anilistRateLimiter,
				ScanLogger:           scanLogger,
				UseAnilistCollection: tt.useAnilistCollection,
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

	anilistClientWrapper, _, _ := anilist.MockAnilistClientWrappers()
	anizipCache := anizip.NewCache()
	baseMediaCache := anilist.NewBaseMediaCache()
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

			var lfs []*entities.LocalFile
			for _, path := range tt.paths {
				lf := entities.NewLocalFile(path, dir)
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
				BaseMediaCache:       baseMediaCache,
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

	anilistClientWrapper := anilist.MockAnilistClientWrapper()
	anizipCache := anizip.NewCache()
	baseMediaCache := anilist.NewBaseMediaCache()
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

			var lfs []*entities.LocalFile
			for _, path := range tt.paths {
				lf := entities.NewLocalFile(path, dir)
				lfs = append(lfs, lf)
			}

			// +--------------------------+
			// | FetchMediaFromLocalFiles |
			// +--------------------------+

			media, ok := FetchMediaFromLocalFiles(
				anilistClientWrapper,
				lfs,
				baseMediaCache,
				anizipCache,
				anilistRateLimiter,
				scanLogger,
			)
			if !ok {
				t.Fatal("could not fetch media from local files")
			}

			ids := lo.Map(media, func(k *anilist.BaseMedia, _ int) int {
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
