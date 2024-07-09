package scanner

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/api/anizip"
	"github.com/seanime-app/seanime/internal/test_utils"
	"github.com/seanime-app/seanime/internal/util/limiter"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMediaTreeAnalysis(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	anilistClientWrapper := anilist.TestGetMockAnilistClientWrapper()

	anilistRateLimiter := limiter.NewAnilistLimiter()
	tree := anilist.NewCompleteMediaRelationTree()

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
		// DEVNOTE: This fails because Anizip doesn't include the absolute episode number - edit: no longer fails
		{
			name:                          "Oshi no Ko Season 2",
			mediaId:                       150672, // 86 - Eighty Six Part 2
			absoluteEpisodeNumber:         12,
			expectedRelativeEpisodeNumber: 1,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			mediaF, err := anilistClientWrapper.BaseMediaByID(context.Background(), &tt.mediaId)
			if err != nil {
				t.Fatal("expected media, got not found")
			}
			media := mediaF.GetMedia()

			// +---------------------+
			// |     MediaTree       |
			// +---------------------+

			err = media.FetchMediaTree(
				anilist.FetchMediaTreeAll,
				anilistClientWrapper,
				anilistRateLimiter,
				tree,
				anilist.NewCompleteMediaCache(),
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

func TestMediaTreeAnalysis2(t *testing.T) {

	anilistClientWrapper := anilist.TestGetMockAnilistClientWrapper()
	anilistRateLimiter := limiter.NewAnilistLimiter()
	tree := anilist.NewCompleteMediaRelationTree()

	tests := []struct {
		name    string
		mediaId int
	}{
		{
			name:    "Media Tree Analysis",
			mediaId: 375, // Soreyuke! Uchuu Senkan Yamamoto Yohko
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			media, err := anilistClientWrapper.BaseMediaByID(context.Background(), &tt.mediaId)
			if err != nil {
				t.Fatal("expected media, got error:", err.Error())
			}

			// +---------------------+
			// |     MediaTree       |
			// +---------------------+

			err = media.GetMedia().FetchMediaTree(
				anilist.FetchMediaTreeAll,
				anilistClientWrapper,
				anilistRateLimiter,
				tree,
				anilist.NewCompleteMediaCache(),
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

			t.Log(spew.Sdump(mta))

		})

	}

}
