package scanner

import (
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/anizip"
	"github.com/seanime-app/seanime/internal/limiter"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMediaTreeAnalysis(t *testing.T) {

	allMedia := getMockedAllMedia(t)

	anilistRateLimiter := limiter.NewAnilistLimiter()
	tree := anilist.NewBaseMediaRelationTree()

	tests := []struct {
		name                          string
		mediaId                       int
		absoluteEpisodeNumber         int
		expectedRelativeEpisodeNumber int
	}{
		{
			name:                          "Media Tree Analysis for 86 - Eighty Six Part 2",
			mediaId:                       131586, // 86 - Eighty Six Part 2
			absoluteEpisodeNumber:         23,
			expectedRelativeEpisodeNumber: 12,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			media, found := lo.Find(allMedia, func(m *anilist.BaseMedia) bool {
				return m.ID == tt.mediaId
			})
			if !found || media == nil {
				t.Fatal("expected media, got not found")
			}

			// +---------------------+
			// |     MediaTree       |
			// +---------------------+

			err := media.FetchMediaTree(
				anilist.FetchMediaTreeAll,
				anilist.NewAuthedClient(""),
				anilistRateLimiter,
				tree,
				anilist.NewBaseMediaCache(),
			)

			if err != nil {
				t.Fatal("expected media tree, got error:", err.Error())
			}

			// +---------------------+
			// |  MediaTreeAnalysis  |
			// +---------------------+

			mta, err := NewMediaTreeAnalysis(&MediaTreeAnalysisOptions{
				tree:        tree,
				anizipCache: anizip.NewCache(),
				rateLimiter: limiter.NewLimiter(time.Minute, 25),
			})
			if err != nil {
				t.Fatal("expected media tree analysis, got error:", err.Error())
			}

			// +---------------------+
			// |  Relative Episode   |
			// +---------------------+

			relEp, _, ok := mta.getRelativeEpisodeNumber(tt.absoluteEpisodeNumber)

			if assert.Truef(t, ok, "expected relative episode number %v for absolute episode number %v, nothing found", tt.expectedRelativeEpisodeNumber, tt.absoluteEpisodeNumber) {

				assert.Equal(t, tt.expectedRelativeEpisodeNumber, relEp)

			}

		})

	}

}
