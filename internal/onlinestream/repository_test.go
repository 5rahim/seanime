package onlinestream

import (
	"seanime/internal/api/anilist"
	"seanime/internal/extension"
	hibikeonlinestream "seanime/internal/extension/hibike/onlinestream"
	"seanime/internal/util"
	"seanime/internal/util/filecache"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type testProvider struct {
	serverCalls int
}

func (p *testProvider) Search(hibikeonlinestream.SearchOptions) ([]*hibikeonlinestream.SearchResult, error) {
	return nil, nil
}

func (p *testProvider) FindEpisodes(string) ([]*hibikeonlinestream.EpisodeDetails, error) {
	return nil, nil
}

func (p *testProvider) FindEpisodeServer(_ *hibikeonlinestream.EpisodeDetails, server string) (*hibikeonlinestream.EpisodeServer, error) {
	p.serverCalls++
	return &hibikeonlinestream.EpisodeServer{
		Server: server,
		VideoSources: []*hibikeonlinestream.VideoSource{
			{URL: "https://example.com/new.m3u8", Quality: "auto", Type: "m3u8"},
		},
	}, nil
}

func (p *testProvider) GetSettings() hibikeonlinestream.Settings {
	return hibikeonlinestream.Settings{EpisodeServers: []string{"default"}}
}

func TestEpisodeSourceRefreshReplacesOnlySourceCache(t *testing.T) {
	require.Equal(t, 15*time.Minute, episodeSourceCacheTTL)
	require.Equal(t, 24*time.Hour, episodeListCacheTTL)

	logger := util.NewLogger()
	cacher, err := filecache.NewCacher(t.TempDir())
	require.NoError(t, err)

	provider := &testProvider{}
	bank := extension.NewUnifiedBank()
	bank.Set("test", extension.NewOnlinestreamProviderExtension(&extension.Extension{
		ID:   "test",
		Name: "Test",
		Type: extension.TypeOnlinestreamProvider,
	}, provider))

	repository := &Repository{
		logger:           logger,
		fileCacher:       cacher,
		extensionBankRef: util.NewRef(bank),
	}
	media := &anilist.BaseAnime{ID: 1}
	episodeDetails := []*hibikeonlinestream.EpisodeDetails{{ID: "episode-1", Number: 1}}
	listKey := "1$test$false"
	sourceKey := "1$test$1$false"

	require.NoError(t, cacher.Set(repository.getFcEpisodeListBucket("test", 1), listKey, episodeDetails))
	require.NoError(t, cacher.Set(repository.getFcEpisodeDataBucket("test", 1), sourceKey, &episodeData{
		Number: 1,
		Servers: []*hibikeonlinestream.EpisodeServer{{
			Server: "default",
			VideoSources: []*hibikeonlinestream.VideoSource{
				{URL: "https://example.com/old.m3u8", Quality: "auto", Type: "m3u8"},
			},
		}},
	}))

	cached, err := repository.getEpisodeContainer("test", media, 1, 1, false, 2026, false)
	require.NoError(t, err)
	require.Equal(t, "https://example.com/old.m3u8", cached.Episodes[0].Servers[0].VideoSources[0].URL)
	require.Zero(t, provider.serverCalls)

	refreshed, err := repository.getEpisodeContainer("test", media, 1, 1, false, 2026, true)
	require.NoError(t, err)
	require.Equal(t, "https://example.com/new.m3u8", refreshed.Episodes[0].Servers[0].VideoSources[0].URL)
	require.Equal(t, 1, provider.serverCalls)

	var storedSource *episodeData
	found, err := cacher.Get(repository.getFcEpisodeDataBucket("test", 1), sourceKey, &storedSource)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, "https://example.com/new.m3u8", storedSource.Servers[0].VideoSources[0].URL)

	var storedList []*hibikeonlinestream.EpisodeDetails
	found, err = cacher.Get(repository.getFcEpisodeListBucket("test", 1), listKey, &storedList)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, episodeDetails, storedList)
	sourceBucket := repository.getFcEpisodeDataBucket("test", 1)
	listBucket := repository.getFcEpisodeListBucket("test", 1)
	require.NotEqual(t, sourceBucket.Name(), listBucket.Name())
}

func TestSubtitleFromProviderKeepsDefaultFlag(t *testing.T) {
	subtitle := subtitleFromProvider(&hibikeonlinestream.VideoSubtitle{
		URL:       "https://example.com/subtitle.vtt",
		Language:  "en",
		IsDefault: true,
	})

	require.Equal(t, "https://example.com/subtitle.vtt", subtitle.URL)
	require.Equal(t, "en", subtitle.Language)
	require.True(t, subtitle.IsDefault)
}
