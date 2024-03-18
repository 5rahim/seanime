package onlinestream

import (
	"context"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/test_utils"
	"github.com/seanime-app/seanime/internal/util"
	"testing"
)

func TestOnlineStream_GetEpisodes(t *testing.T) {
	test_utils.SetTwoLevelDeep()
	test_utils.InitTestProvider(t, test_utils.Anilist())

	anilistClientWrapper := anilist.TestGetMockAnilistClientWrapper()

	os := New(&NewOnlineStreamOptions{
		Logger: util.NewLogger(),
	})

	tests := []struct {
		name    string
		mediaId int
		from    int
		to      int
	}{
		{
			name:    "Cowboy Bebop",
			mediaId: 1,
			from:    1,
			to:      2,
		},
		{
			name:    "One Piece",
			mediaId: 21,
			from:    1075,
			to:      1076,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			mediaF, err := anilistClientWrapper.BaseMediaByID(context.Background(), &tt.mediaId)
			if err != nil {
				t.Fatalf("couldn't get media: %s", err)
			}
			media := mediaF.GetMedia()

			res, found := os.GetEpisodes(tt.mediaId, media.GetAllTitles(), tt.from, tt.to, false)
			if !found {
				t.Fatalf("couldn't find episodes for %+v", tt.mediaId)
			}

			for _, e := range res.ProviderEpisodes {
				t.Logf("Provider: %s, found %d episodes", e.Provider, len(e.Episodes))
				for _, ep := range e.Episodes {
					t.Logf("\t\tEpisode %d has %d server sources", ep.Number, len(ep.ServerSources))
					for _, ss := range ep.ServerSources {
						t.Logf("\t\t\tServer: %s, Quality: %s", ss.Server, ss.Quality)
					}
				}
			}

		})

	}

}
