package animetosho

import (
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/platforms/anilist_platform"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSmartSearch(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	anilistClient := anilist.TestGetMockAnilistClient()
	logger := util.NewLogger()
	anilistPlatform := anilist_platform.NewAnilistPlatform(anilistClient, logger)

	toshoPlatform := NewProvider(util.NewLogger())

	metadataProvider := metadata.GetMockProvider(t)

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
			batch:          true,
			episodeNumber:  1,
			absoluteOffset: 0,
			resolution:     "720",
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

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			media, err := anilistPlatform.GetAnime(t.Context(), tt.mId)
			animeMetadata, err := metadataProvider.GetAnimeMetadata(metadata.AnilistPlatform, tt.mId)
			require.NoError(t, err)

			queryMedia := hibiketorrent.Media{
				ID:                   media.GetID(),
				IDMal:                media.GetIDMal(),
				Status:               string(*media.GetStatus()),
				Format:               string(*media.GetFormat()),
				EnglishTitle:         media.GetTitle().GetEnglish(),
				RomajiTitle:          media.GetRomajiTitleSafe(),
				EpisodeCount:         media.GetTotalEpisodeCount(),
				AbsoluteSeasonOffset: tt.absoluteOffset,
				Synonyms:             media.GetSynonymsContainingSeason(),
				IsAdult:              *media.GetIsAdult(),
				StartDate: &hibiketorrent.FuzzyDate{
					Year:  *media.GetStartDate().GetYear(),
					Month: media.GetStartDate().GetMonth(),
					Day:   media.GetStartDate().GetDay(),
				},
			}

			if assert.NoError(t, err) {

				episodeMetadata, ok := animeMetadata.FindEpisode(strconv.Itoa(tt.episodeNumber))
				require.True(t, ok)

				torrents, err := toshoPlatform.SmartSearch(hibiketorrent.AnimeSmartSearchOptions{
					Media:         queryMedia,
					Query:         "",
					Batch:         tt.batch,
					EpisodeNumber: tt.episodeNumber,
					Resolution:    tt.resolution,
					AnidbAID:      animeMetadata.Mappings.AnidbId,
					AnidbEID:      episodeMetadata.AnidbEid,
					BestReleases:  false,
				})

				require.NoError(t, err)
				require.GreaterOrEqual(t, len(torrents), 1, "expected at least 1 torrent")

				for _, torrent := range torrents {
					t.Log(torrent.Name)
					t.Logf("\tLink: %s", torrent.Link)
					t.Logf("\tMagnet: %s", torrent.MagnetLink)
					t.Logf("\tEpisodeNumber: %d", torrent.EpisodeNumber)
					t.Logf("\tResolution: %s", torrent.Resolution)
					t.Logf("\tIsBatch: %v", torrent.IsBatch)
					t.Logf("\tConfirmed: %v", torrent.Confirmed)
				}

			}

		})

	}
}

func TestSearch2(t *testing.T) {

	toshoPlatform := NewProvider(util.NewLogger())
	torrents, err := toshoPlatform.Search(hibiketorrent.AnimeSearchOptions{
		Media: hibiketorrent.Media{},
		Query: "Kusuriya no Hitorigoto 05",
	})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(torrents), 1, "expected at least 1 torrent")

	for _, torrent := range torrents {
		t.Log(torrent.Name)
		t.Logf("\tLink: %s", torrent.Link)
		t.Logf("\tMagnet: %s", torrent.MagnetLink)
		t.Logf("\tEpisodeNumber: %d", torrent.EpisodeNumber)
		t.Logf("\tResolution: %s", torrent.Resolution)
		t.Logf("\tIsBatch: %v", torrent.IsBatch)
		t.Logf("\tConfirmed: %v", torrent.Confirmed)
	}
}

func TestGetLatest(t *testing.T) {

	toshoPlatform := NewProvider(util.NewLogger())
	torrents, err := toshoPlatform.GetLatest()
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(torrents), 1, "expected at least 1 torrent")

	for _, torrent := range torrents {
		t.Log(torrent.Name)
		t.Logf("\tLink: %s", torrent.Link)
		t.Logf("\tMagnet: %s", torrent.MagnetLink)
		t.Logf("\tEpisodeNumber: %d", torrent.EpisodeNumber)
		t.Logf("\tResolution: %s", torrent.Resolution)
		t.Logf("\tIsBatch: %v", torrent.IsBatch)
		t.Logf("\tConfirmed: %v", torrent.Confirmed)
	}
}
