package scanner

import (
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/anizip"
	"github.com/seanime-app/seanime/internal/entities"
	"github.com/seanime-app/seanime/internal/limiter"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFetchMediaFromLocalFiles(t *testing.T) {

	anilistClient := anilist.MockGetAnilistClient()
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

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			// +---------------------+
			// |   Local Files       |
			// +---------------------+

			var lfs []*entities.LocalFile
			for _, path := range tt.paths {
				lf := entities.NewLocalFile(path, "E:/Anime")
				lfs = append(lfs, lf)
			}

			media, ok := FetchMediaFromLocalFiles(anilistClient, lfs, baseMediaCache, anizipCache, anilistRateLimiter)
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
