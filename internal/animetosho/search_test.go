package animetosho

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/test_utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSearchQuery(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	anilistClientWrapper := anilist.TestGetMockAnilistClientWrapper()

	tests := []struct {
		name           string
		mId            int
		batch          bool
		episodeNumber  int
		absoluteOffset int
		resolution     string
	}{
		{
			name:           "Bungou Stray Dogs 5th Season Episode 11",
			mId:            163263,
			batch:          false,
			episodeNumber:  11,
			absoluteOffset: 45,
			resolution:     "1080p",
		},
		{
			name:           "SPY×FAMILY Season 1 Part 2",
			mId:            142838,
			batch:          false,
			episodeNumber:  12,
			absoluteOffset: 12,
			resolution:     "1080p",
		},
		{
			name:           "Jujutsu Kaisen Season 2",
			mId:            145064,
			batch:          false,
			episodeNumber:  2,
			absoluteOffset: 24,
			resolution:     "",
		},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {

			mediaRes, err := anilistClientWrapper.BaseMediaByID(context.Background(), &test.mId)

			if assert.NoError(t, err) {

				torrents, err := SearchQuery(&BuildSearchQueryOptions{
					Media:          mediaRes.GetMedia(),
					Batch:          &test.batch,
					EpisodeNumber:  &test.episodeNumber,
					AbsoluteOffset: &test.absoluteOffset,
					Resolution:     &test.resolution,
					Cache:          NewSearchCache(),
				})

				if assert.NoError(t, err) {
					assert.GreaterOrEqual(t, len(torrents), 1, "expected at least 1 torrent")
					spew.Dump(torrents)
				}

			}

		})

	}
}

func TestSearch2(t *testing.T) {
	torrents, err := Search("Kusuriya no Hitorigoto 05")
	if assert.NoError(t, err) {
		assert.GreaterOrEqual(t, len(torrents), 1, "expected at least 1 torrent")
	}
}
