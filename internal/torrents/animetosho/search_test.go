package animetosho

import (
	"context"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/api/anizip"
	"github.com/seanime-app/seanime/internal/test_utils"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestSmartSearch(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	anilistClient := anilist.TestGetMockAnilistClient()

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
			resolution:     "1080",
		},
		{
			name:           "SPY×FAMILY Season 1 Part 2",
			mId:            142838,
			batch:          false,
			episodeNumber:  12,
			absoluteOffset: 12,
			resolution:     "1080",
		},
		{
			name:           "Jujutsu Kaisen Season 2",
			mId:            145064,
			batch:          false,
			episodeNumber:  2,
			absoluteOffset: 24,
			resolution:     "",
		},
		{
			name:           "Violet Evergarden The Movie",
			mId:            103047,
			batch:          false,
			episodeNumber:  1,
			absoluteOffset: 0,
			resolution:     "",
		},
		{
			name:           "Sousou no Frieren",
			mId:            154587,
			batch:          false,
			episodeNumber:  10,
			absoluteOffset: 0,
			resolution:     "1080",
		},
		{
			name:           "Tokubetsu-hen Hibike! Euphonium: Ensemble",
			mId:            150429,
			batch:          false,
			episodeNumber:  1,
			absoluteOffset: 0,
			resolution:     "1080",
		},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {

			mediaRes, err := anilistClient.BaseAnimeByID(context.Background(), &test.mId)

			if assert.NoError(t, err) {

				torrents, err := SearchQuery(&BuildSearchQueryOptions{
					Media:          mediaRes.GetMedia(),
					Batch:          &test.batch,
					EpisodeNumber:  &test.episodeNumber,
					AbsoluteOffset: &test.absoluteOffset,
					Resolution:     &test.resolution,
					Cache:          NewSearchCache(),
					Logger:         util.NewLogger(),
				})

				if assert.NoError(t, err) {
					assert.GreaterOrEqual(t, len(torrents), 1, "expected at least 1 torrent")
					for _, torrent := range torrents {
						t.Log(torrent.Title)
					}
				}

			}

		})

	}
}

func TestSearchByAID(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	anilistClient := anilist.TestGetMockAnilistClient()

	tests := []struct {
		name          string
		mId           int
		episodeNumber int
		resolution    string
	}{
		{
			name:          "Bungou Stray Dogs 5th Season Episode 11",
			mId:           163263,
			episodeNumber: 11,
			resolution:    "1080",
		},
		{
			name:          "SPY×FAMILY Season 1 Part 2",
			mId:           142838,
			episodeNumber: 12,
			resolution:    "1080",
		},
		{
			name:          "Jujutsu Kaisen Season 2",
			mId:           145064,
			episodeNumber: 2,
			resolution:    "",
		},
		{
			name:          "Violet Evergarden The Movie",
			mId:           103047,
			episodeNumber: 1,
			resolution:    "",
		},
		{
			name:          "Sousou no Frieren",
			mId:           154587,
			episodeNumber: 10,
			resolution:    "1080",
		},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {

			mediaRes, err := anilistClient.BaseAnimeByID(context.Background(), &test.mId)
			media := mediaRes.GetMedia()
			anizipMedia, err := anizip.FetchAniZipMedia("anilist", media.ID)
			if err != nil {
				t.Fatal(err)
			}

			if assert.NoError(t, err) {

				torrents, err := SearchByAID(anizipMedia.Mappings.AnidbID, "1080")

				if assert.NoError(t, err) {
					assert.GreaterOrEqual(t, len(torrents), 1, "expected at least 1 torrent")
					for _, torrent := range torrents {
						t.Log(torrent.Title)
					}
				}

			}

		})

	}
}

func TestSearchByEID(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	anilistClient := anilist.TestGetMockAnilistClient()

	tests := []struct {
		name          string
		mId           int
		episodeNumber int
		resolution    string
	}{
		{
			name:          "Dr Stone New World Part 2",
			mId:           162670,
			episodeNumber: 1,
			resolution:    "1080",
		},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {

			mediaRes, err := anilistClient.BaseAnimeByID(context.Background(), &test.mId)
			media := mediaRes.GetMedia()
			anizipMedia, err := anizip.FetchAniZipMedia("anilist", media.ID)
			if err != nil {
				t.Fatal(err)
			}

			anizipEpisode, found := anizipMedia.FindEpisode(strconv.Itoa(test.episodeNumber))
			if !found {
				t.Fatalf("episode %d not found", test.episodeNumber)
			}

			if assert.NoError(t, err) {

				torrents, err := SearchByEID(anizipEpisode.AnidbEid, "1080")

				if assert.NoError(t, err) {
					assert.GreaterOrEqual(t, len(torrents), 1, "expected at least 1 torrent")
					for _, torrent := range torrents {
						t.Log(torrent.Title)
					}
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
