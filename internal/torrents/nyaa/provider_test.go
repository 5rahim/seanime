package nyaa

import (
	"seanime/internal/api/anilist"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/platforms/anilist_platform"
	"seanime/internal/util"
	"seanime/internal/util/limiter"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSearch(t *testing.T) {

	nyaaProvider := NewProvider(util.NewLogger(), categoryAnime)

	torrents, err := nyaaProvider.Search(hibiketorrent.AnimeSearchOptions{
		Query: "One Piece",
	})
	require.NoError(t, err)

	for _, torrent := range torrents {
		t.Log(torrent.Name)
	}
}

func TestSmartSearch(t *testing.T) {

	anilistLimiter := limiter.NewAnilistLimiter()
	anilistClient := anilist.TestGetMockAnilistClient()
	logger := util.NewLogger()
	anilistPlatform := anilist_platform.NewAnilistPlatform(anilistClient, logger)

	nyaaProvider := NewProvider(util.NewLogger(), categoryAnime)

	tests := []struct {
		name           string
		mId            int
		batch          bool
		episodeNumber  int
		absoluteOffset int
		resolution     string
		scrapeMagnet   bool
	}{
		{
			name:           "Bungou Stray Dogs 5th Season Episode 11",
			mId:            163263,
			batch:          false,
			episodeNumber:  11,
			absoluteOffset: 45,
			resolution:     "1080",
			scrapeMagnet:   true,
		},
		{
			name:           "SPY×FAMILY Season 1 Part 2",
			mId:            142838,
			batch:          false,
			episodeNumber:  12,
			absoluteOffset: 12,
			resolution:     "1080",
			scrapeMagnet:   false,
		},
		{
			name:           "Jujutsu Kaisen Season 2",
			mId:            145064,
			batch:          false,
			episodeNumber:  2,
			absoluteOffset: 24,
			resolution:     "1080",
			scrapeMagnet:   false,
		},
		{
			name:           "Violet Evergarden The Movie",
			mId:            103047,
			batch:          true,
			episodeNumber:  1,
			absoluteOffset: 0,
			resolution:     "720",
			scrapeMagnet:   false,
		},
		{
			name:           "Sousou no Frieren",
			mId:            154587,
			batch:          false,
			episodeNumber:  10,
			absoluteOffset: 0,
			resolution:     "1080",
			scrapeMagnet:   false,
		},
		{
			name:           "Tokubetsu-hen Hibike! Euphonium: Ensemble",
			mId:            150429,
			batch:          false,
			episodeNumber:  1,
			absoluteOffset: 0,
			resolution:     "1080",
			scrapeMagnet:   false,
		},
	}

	for _, tt := range tests {

		anilistLimiter.Wait()

		t.Run(tt.name, func(t *testing.T) {

			media, err := anilistPlatform.GetAnime(t.Context(), tt.mId)
			require.NoError(t, err)
			require.NotNil(t, media)

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

			torrents, err := nyaaProvider.SmartSearch(hibiketorrent.AnimeSmartSearchOptions{
				Media:         queryMedia,
				Query:         "",
				Batch:         tt.batch,
				EpisodeNumber: tt.episodeNumber,
				Resolution:    tt.resolution,
				AnidbAID:      0,     // Not supported
				AnidbEID:      0,     // Not supported
				BestReleases:  false, // Not supported
			})
			require.NoError(t, err, "error searching nyaa")

			for _, torrent := range torrents {

				scrapedMagnet := ""
				if tt.scrapeMagnet {
					magn, err := nyaaProvider.GetTorrentMagnetLink(torrent)
					if err == nil {
						scrapedMagnet = magn
					}
				}

				t.Log(torrent.Name)
				t.Logf("\tMagnet: %s", torrent.MagnetLink)
				if scrapedMagnet != "" {
					t.Logf("\tMagnet (Scraped): %s", scrapedMagnet)
				}
				t.Logf("\tEpisodeNumber: %d", torrent.EpisodeNumber)
				t.Logf("\tResolution: %s", torrent.Resolution)
				t.Logf("\tIsBatch: %v", torrent.IsBatch)
				t.Logf("\tConfirmed: %v", torrent.Confirmed)
			}

		})

	}

}
