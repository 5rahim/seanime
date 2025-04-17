package onlinestream

import (
	"context"
	"path/filepath"
	"seanime/internal/api/anilist"
	onlinestream_providers "seanime/internal/onlinestream/providers"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"seanime/internal/util/filecache"
	"testing"
)

func TestOnlineStream_GetEpisodes(t *testing.T) {
	t.Skip("TODO: Fix this test by loading built-in extensions")
	test_utils.SetTwoLevelDeep()
	test_utils.InitTestProvider(t, test_utils.Anilist())

	tempDir := t.TempDir()

	anilistClient := anilist.TestGetMockAnilistClient()

	//fileCacher, _ := filecache.NewCacher(filepath.Join(test_utils.ConfigData.Path.DataDir, "cache"))
	fileCacher, _ := filecache.NewCacher(filepath.Join(tempDir, "cache"))

	os := NewRepository(&NewRepositoryOptions{
		Logger:     util.NewLogger(),
		FileCacher: fileCacher,
	})

	tests := []struct {
		name     string
		mediaId  int
		from     int
		to       int
		provider string
		dubbed   bool
	}{
		{
			name:     "Cowboy Bebop",
			mediaId:  1,
			from:     1,
			to:       2,
			provider: onlinestream_providers.GogoanimeProvider,
			dubbed:   false,
		},
		{
			name:     "Cowboy Bebop",
			mediaId:  1,
			from:     1,
			to:       2,
			provider: onlinestream_providers.ZoroProvider,
			dubbed:   false,
		},
		{
			name:     "One Piece",
			mediaId:  21,
			from:     1075,
			to:       1076,
			provider: onlinestream_providers.ZoroProvider,
			dubbed:   false,
		},
		{
			name:     "Dungeon Meshi",
			mediaId:  153518,
			from:     1,
			to:       1,
			provider: onlinestream_providers.ZoroProvider,
			dubbed:   false,
		},
		{
			name:     "Omoi, Omoware, Furi, Furare",
			mediaId:  109125,
			from:     1,
			to:       1,
			provider: onlinestream_providers.ZoroProvider,
			dubbed:   false,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			mediaF, err := anilistClient.BaseAnimeByID(context.Background(), &tt.mediaId)
			if err != nil {
				t.Fatalf("couldn't get media: %s", err)
			}
			media := mediaF.GetMedia()

			ec, err := os.getEpisodeContainer(tt.provider, media, tt.from, tt.to, tt.dubbed, 0)
			if err != nil {
				t.Fatalf("couldn't find episodes, %s", err)
			}

			t.Logf("Provider: %s, found %d episodes for the anime", ec.Provider, len(ec.ProviderEpisodeList))
			// Episode Data
			for _, ep := range ec.Episodes {
				t.Logf("\t\tEpisode %d has %d servers", ep.Number, len(ep.Servers))
				for _, s := range ep.Servers {
					t.Logf("\t\t\tServer: %s", s.Server)
					for _, vs := range s.VideoSources {
						t.Logf("\t\t\t\tVideo Source: %s, Type: %s", vs.Quality, vs.Type)
					}
				}
			}

		})

	}

}
