package onlinestream

import (
	"context"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/test_utils"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/seanime-app/seanime/internal/util/filecache"
	"path/filepath"
	"testing"
)

func TestOnlineStream_GetEpisodes(t *testing.T) {
	test_utils.SetTwoLevelDeep()
	test_utils.InitTestProvider(t, test_utils.Anilist())

	anilistClientWrapper := anilist.TestGetMockAnilistClientWrapper()

	fileCacher, _ := filecache.NewCacher(filepath.Join(test_utils.ConfigData.Path.DataDir, "cache"))

	os := New(&NewOnlineStreamOptions{
		Logger:     util.NewLogger(),
		FileCacher: fileCacher,
	})

	tests := []struct {
		name     string
		mediaId  int
		from     int
		to       int
		provider Provider
	}{
		{
			name:     "Cowboy Bebop",
			mediaId:  1,
			from:     1,
			to:       2,
			provider: ProviderGogoanime,
		},
		{
			name:     "Cowboy Bebop",
			mediaId:  1,
			from:     1,
			to:       2,
			provider: ProviderGogoanime,
		},
		{
			name:     "One Piece",
			mediaId:  21,
			from:     1075,
			to:       1076,
			provider: ProviderZoro,
		},
		{
			name:     "Dungeon Meshi",
			mediaId:  153518,
			from:     1,
			to:       1,
			provider: ProviderZoro,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			mediaF, err := anilistClientWrapper.BaseMediaByID(context.Background(), &tt.mediaId)
			if err != nil {
				t.Fatalf("couldn't get media: %s", err)
			}
			media := mediaF.GetMedia()

			res, found := os.getEpisodeContainer(tt.provider, tt.mediaId, media.GetAllTitles(), tt.from, tt.to, false)
			if !found {
				t.Fatalf("couldn't find episodes for %+v", tt.mediaId)
			}

			for _, e := range res.ProviderEpisodes {
				t.Logf("Provider: %s, found %d episodes", e.Provider, len(e.ExtractedEpisodes))
				for _, ep := range e.ExtractedEpisodes {
					t.Logf("\t\tEpisode %d has %d server sources", ep.Number, len(ep.ServerSources))
					for _, ss := range ep.ServerSources {
						t.Logf("\t\t\tServer: %s", ss.Server)
						for _, vs := range ss.VideoSources {
							t.Logf("\t\t\t\tVideo Source: %s, Type: %s", vs.Quality, vs.Type)
						}
					}
				}
			}

		})

	}

}
