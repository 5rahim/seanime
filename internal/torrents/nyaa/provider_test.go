package nyaa

import (
	hibiketorrent "github.com/5rahim/hibike/pkg/extension/torrent"
	"github.com/stretchr/testify/require"
	"seanime/internal/api/anilist"
	"seanime/internal/platform"
	"seanime/internal/util"
	"seanime/internal/util/limiter"
	"testing"
)

func TestSearch(t *testing.T) {

	nyaaProvider := NewProvider(util.NewLogger())

	torrents, err := nyaaProvider.Search(hibiketorrent.SearchOptions{
		Query: "One Piece",
	})
	require.NoError(t, err)

	for _, torrent := range torrents {
		t.Log(torrent.Name)
	}
}

func TestBuildSearchQuery(t *testing.T) {

	anilistLimiter := limiter.NewAnilistLimiter()
	anilistClient := anilist.TestGetMockAnilistClient()
	anilistPlatform := platform.NewAnilistPlatform(anilistClient, util.NewLogger())

	nyaaProvider := NewProvider(util.NewLogger())

	tests := []struct {
		name           string
		mediaId        int
		batch          bool
		episodeNumber  int
		absoluteOffset int
		resolution     string
		title          *string
	}{
		{
			name:           "ReZero kara Hajimeru Isekai Seikatsu 2nd Season",
			batch:          false,
			mediaId:        108632,
			episodeNumber:  1,
			absoluteOffset: 24,
			resolution:     "",
			title:          nil,
		},
	}

	for _, tt := range tests {

		anilistLimiter.Wait()

		t.Run(tt.name, func(t *testing.T) {

			media, err := anilistPlatform.GetAnime(tt.mediaId)
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

			torrents, err := nyaaProvider.SmartSearch(hibiketorrent.SmartSearchOptions{
				Media:         queryMedia,
				Query:         "",
				Batch:         tt.batch,
				EpisodeNumber: tt.episodeNumber,
				Resolution:    "",
				AniDbAID:      0,
				AniDbEID:      0,
				BestReleases:  false,
			})
			require.NoError(t, err, "error searching nyaa")

			for _, torrent := range torrents {
				t.Log(torrent.Name)
				t.Logf("\tMagnet: %s", torrent.MagnetLink)
				t.Logf("\tEpisodeNumber: %d", torrent.EpisodeNumber)
				t.Logf("\tResolution: %s", torrent.Resolution)
				t.Logf("\tIsBatch: %v", torrent.IsBatch)
				t.Logf("\tConfirmed: %v", torrent.Confirmed)
			}

		})

	}

}
