package torrent

import (
	"context"
	"seanime/internal/api/metadata"
	"seanime/internal/api/metadata_provider"
	"seanime/internal/extension"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/hook"
	"seanime/internal/hook_resolver"
	"seanime/internal/testmocks"
	"seanime/internal/util"
	"sync"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func useTestHookManager(t *testing.T) hook.Manager {
	t.Helper()

	prev := hook.GlobalHookManager
	hm := hook.NewHookManager(hook.NewHookManagerOptions{Logger: util.NewLogger()})
	hook.SetGlobalHookManager(hm)
	t.Cleanup(func() {
		hook.SetGlobalHookManager(prev)
	})

	return hm
}

func TestRepositoryProviderSelection(t *testing.T) {
	mainProvider := newStubAnimeProvider(hibiketorrent.AnimeProviderSettings{Type: hibiketorrent.AnimeProviderTypeMain})
	fallbackProvider := newStubAnimeProvider(hibiketorrent.AnimeProviderSettings{Type: hibiketorrent.AnimeProviderTypeMain})
	specialProvider := newStubAnimeProvider(hibiketorrent.AnimeProviderSettings{Type: hibiketorrent.AnimeProviderTypeSpecial})

	repo := newTorrentRepositoryForTests(map[string]*stubAnimeProvider{
		"main":     mainProvider,
		"fallback": fallbackProvider,
		"special":  specialProvider,
	}, testmocks.NewFakeMetadataProviderBuilder().Build())
	repo.SetSettings(&RepositorySettings{DefaultAnimeProvider: "fallback", AutoSelectProvider: "special"})

	ext, ok := repo.GetDefaultAnimeProviderExtension()
	require.True(t, ok)
	require.Equal(t, "fallback", ext.GetID())

	ext, ok = repo.GetAnimeProviderExtensionOrFirst("missing")
	require.True(t, ok)
	require.Equal(t, "fallback", ext.GetID())

	ext, ok = repo.GetAutoSelectProviderExtension()
	require.True(t, ok)
	require.Equal(t, "special", ext.GetID())

	filtered := repo.GetAnimeProviderExtensionsBy(func(ext extension.AnimeTorrentProviderExtension) bool {
		return ext.GetProvider().GetSettings().Type == hibiketorrent.AnimeProviderTypeSpecial
	})
	require.Len(t, filtered, 1)
	require.Equal(t, "special", filtered[0].GetID())
}

func TestRepositoryProviderSelectionWithoutProviders(t *testing.T) {
	repo := newTorrentRepositoryForTests(nil, testmocks.NewFakeMetadataProviderBuilder().Build())

	ext, ok := repo.GetDefaultAnimeProviderExtension()
	require.False(t, ok)
	require.Nil(t, ext)

	ext, ok = repo.GetAnimeProviderExtensionOrFirst("missing")
	require.False(t, ok)
	require.Nil(t, ext)
}

func TestSearchAnimeSimpleFallbackDedupAndSorting(t *testing.T) {
	metadataCache.Clear()
	provider := newStubAnimeProvider(hibiketorrent.AnimeProviderSettings{Type: hibiketorrent.AnimeProviderTypeMain})
	provider.searchResults = []*hibiketorrent.AnimeTorrent{
		{Name: "[SubsPlease] Example Show - 02 (1080p).mkv", InfoHash: "hash-2", Seeders: 10},
		{Name: "[Best] Example Show - 01-12 (1080p).mkv", InfoHash: "hash-best", Seeders: 1, IsBestRelease: true},
		{Name: "[Duplicate] Example Show - 02 (1080p).mkv", InfoHash: "hash-2", Seeders: 99},
		{Name: "[SubsPlease] Example Show - 01 (720p).mkv", Seeders: 20},
	}

	fakeMetadata := testmocks.NewFakeMetadataProviderBuilder().Build()
	repo := newTorrentRepositoryForTests(map[string]*stubAnimeProvider{"main": provider}, fakeMetadata)
	repo.SetSettings(&RepositorySettings{DefaultAnimeProvider: "main"})

	media := testmocks.NewBaseAnimeBuilder(100, "Example Show").WithEpisodes(12).Build()

	result, err := repo.SearchAnime(context.Background(), AnimeSearchOptions{
		Provider: "missing-provider",
		Type:     AnimeSearchTypeSimple,
		Media:    media,
		Query:    "Example Show",
	})

	require.NoError(t, err)
	require.Equal(t, 1, provider.searchCallsCount())
	require.Equal(t, 0, provider.smartCallsCount())
	require.Equal(t, 1, fakeMetadata.MetadataCalls(media.ID))
	require.Len(t, result.Torrents, 3)
	require.Empty(t, result.Previews)
	require.Equal(t, "hash-best", result.Torrents[0].InfoHash)
	require.Equal(t, "[SubsPlease] Example Show - 01 (720p).mkv", result.Torrents[1].InfoHash)
	require.Equal(t, "hash-2", result.Torrents[2].InfoHash)
	require.Equal(t, 10, result.Torrents[2].Seeders)
	require.Len(t, result.TorrentMetadata, 3)
	require.Contains(t, result.TorrentMetadata, "hash-best")
	require.Contains(t, result.TorrentMetadata, "hash-2")
	require.Contains(t, result.TorrentMetadata, "[SubsPlease] Example Show - 01 (720p).mkv")
	require.Nil(t, result.AnimeMetadata)

	lastSearch := provider.lastSearchOptions()
	require.Equal(t, "Example Show", lastSearch.Query)
	require.Equal(t, media.ID, lastSearch.Media.ID)
	require.Equal(t, media.GetTotalEpisodeCount(), lastSearch.Media.EpisodeCount)
}

func TestSearchAnimeUsesRequestedHookOverride(t *testing.T) {
	metadataCache.Clear()
	hm := useTestHookManager(t)
	provider := newStubAnimeProvider(hibiketorrent.AnimeProviderSettings{Type: hibiketorrent.AnimeProviderTypeMain})
	provider.searchResults = []*hibiketorrent.AnimeTorrent{{
		Name:     "[Provider] Example Show - 01 (1080p).mkv",
		InfoHash: "provider-hash",
		Seeders:  10,
	}}

	hm.OnTorrentSearchRequested().BindFunc(func(e hook_resolver.Resolver) error {
		event := e.(*TorrentSearchRequestedEvent)
		event.SearchData = &SearchData{
			Torrents: []*hibiketorrent.AnimeTorrent{{
				Name:     "[Hook] Example Show - 01 (1080p).mkv",
				InfoHash: "hook-hash",
				Seeders:  99,
			}},
		}
		event.PreventDefault()
		return event.Next()
	})

	repo := newTorrentRepositoryForTests(map[string]*stubAnimeProvider{"main": provider}, testmocks.NewFakeMetadataProviderBuilder().Build())
	repo.SetSettings(&RepositorySettings{DefaultAnimeProvider: "main"})

	media := testmocks.NewBaseAnimeBuilder(101, "Example Show").WithEpisodes(12).Build()

	result, err := repo.SearchAnime(context.Background(), AnimeSearchOptions{
		Provider: "main",
		Type:     AnimeSearchTypeSimple,
		Media:    media,
		Query:    "Example Show",
	})

	require.NoError(t, err)
	require.Equal(t, 0, provider.searchCallsCount())
	require.Len(t, result.Torrents, 1)
	require.Equal(t, "hook-hash", result.Torrents[0].InfoHash)
}

func TestSearchAnimeAppliesSearchHookBeforeCaching(t *testing.T) {
	metadataCache.Clear()
	hm := useTestHookManager(t)
	provider := newStubAnimeProvider(hibiketorrent.AnimeProviderSettings{Type: hibiketorrent.AnimeProviderTypeMain})
	provider.searchResults = []*hibiketorrent.AnimeTorrent{{
		Name:     "[Provider] Example Show - 01 (1080p).mkv",
		InfoHash: "provider-hash",
		Seeders:  10,
	}}

	hookCalls := 0
	hm.OnTorrentSearch().BindFunc(func(e hook_resolver.Resolver) error {
		event := e.(*TorrentSearchEvent)
		hookCalls++
		event.SearchData.Torrents = append(event.SearchData.Torrents, &hibiketorrent.AnimeTorrent{
			Name:     "[Hook] Example Show Batch (1080p).mkv",
			InfoHash: "hook-added",
			Seeders:  50,
		})
		return event.Next()
	})

	repo := newTorrentRepositoryForTests(map[string]*stubAnimeProvider{"main": provider}, testmocks.NewFakeMetadataProviderBuilder().Build())
	repo.SetSettings(&RepositorySettings{DefaultAnimeProvider: "main"})

	media := testmocks.NewBaseAnimeBuilder(102, "Example Show").WithEpisodes(12).Build()
	searchOpts := AnimeSearchOptions{
		Provider: "main",
		Type:     AnimeSearchTypeSimple,
		Media:    media,
		Query:    "Example Show",
	}

	result, err := repo.SearchAnime(context.Background(), searchOpts)
	require.NoError(t, err)
	require.Len(t, result.Torrents, 2)
	require.Equal(t, 1, provider.searchCallsCount())
	require.Equal(t, 1, hookCalls)

	cached, err := repo.SearchAnime(context.Background(), searchOpts)
	require.NoError(t, err)
	require.Len(t, cached.Torrents, 2)
	require.Equal(t, "hook-added", cached.Torrents[0].InfoHash)
	require.Equal(t, 1, provider.searchCallsCount())
	require.Equal(t, 1, hookCalls)
}

func TestSearchAnimeSmartSearchesRequestedProviders(t *testing.T) {
	metadataCache.Clear()
	// extra providers are only searched when they are explicitly requested
	mainProvider := newStubAnimeProvider(hibiketorrent.AnimeProviderSettings{
		Type:           hibiketorrent.AnimeProviderTypeMain,
		CanSmartSearch: true,
	})
	mainProvider.smartResults = []*hibiketorrent.AnimeTorrent{{
		Name:          "[Main] Example Show - 05 (1080p).mkv",
		InfoHash:      "main-hash",
		Seeders:       20,
		EpisodeNumber: 5,
	}}

	specialSimpleProvider := newStubAnimeProvider(hibiketorrent.AnimeProviderSettings{
		Type:           hibiketorrent.AnimeProviderTypeSpecial,
		CanSmartSearch: false,
	})
	specialSimpleProvider.searchResults = []*hibiketorrent.AnimeTorrent{{
		Name:     "[SpecialSimple] Example Show Batch (1080p).mkv",
		InfoHash: "special-simple-hash",
		Seeders:  5,
		IsBatch:  true,
	}}

	specialSmartProvider := newStubAnimeProvider(hibiketorrent.AnimeProviderSettings{
		Type:           hibiketorrent.AnimeProviderTypeSpecial,
		CanSmartSearch: true,
	})
	specialSmartProvider.smartResults = []*hibiketorrent.AnimeTorrent{{
		Name:     "[SpecialSmart] Example Show Batch (1080p).mkv",
		InfoHash: "special-smart-hash",
		Seeders:  8,
		IsBatch:  true,
	}}
	unusedSpecialProvider := newStubAnimeProvider(hibiketorrent.AnimeProviderSettings{
		Type:           hibiketorrent.AnimeProviderTypeSpecial,
		CanSmartSearch: true,
	})
	unusedSpecialProvider.smartResults = []*hibiketorrent.AnimeTorrent{{
		Name:     "[UnusedSpecial] Example Show Batch (1080p).mkv",
		InfoHash: "unused-special-hash",
		Seeders:  30,
		IsBatch:  true,
	}}

	fakeMetadata := testmocks.NewFakeMetadataProviderBuilder().WithAnimeMetadata(200, &metadata.AnimeMetadata{
		Episodes: map[string]*metadata.EpisodeMetadata{
			"1": {Episode: "1", AbsoluteEpisodeNumber: 13},
			"5": {Episode: "5", AnidbEid: 505},
		},
		Mappings: &metadata.AnimeMappings{AnidbId: 1001},
	}).Build()

	repo := newTorrentRepositoryForTests(map[string]*stubAnimeProvider{
		"main":           mainProvider,
		"special-simple": specialSimpleProvider,
		"special-smart":  specialSmartProvider,
		"special-unused": unusedSpecialProvider,
	}, fakeMetadata)

	media := testmocks.NewBaseAnimeBuilder(200, "Example Show").WithEpisodes(24).Build()

	result, err := repo.SearchAnime(context.Background(), AnimeSearchOptions{
		Provider:                "main,special-simple,special-smart",
		Type:                    AnimeSearchTypeSmart,
		Media:                   media,
		Query:                   "Example Show",
		Batch:                   true,
		EpisodeNumber:           5,
		IncludeSpecialProviders: true,
		SkipPreviews:            true,
	})

	require.NoError(t, err)
	require.Equal(t, 1, mainProvider.smartCallsCount())
	require.Equal(t, 0, mainProvider.searchCallsCount())
	require.Equal(t, 1, specialSimpleProvider.searchCallsCount())
	require.Equal(t, 0, specialSimpleProvider.smartCallsCount())
	require.Equal(t, 1, specialSmartProvider.smartCallsCount())
	require.Equal(t, 0, unusedSpecialProvider.smartCallsCount())
	require.Equal(t, 1, fakeMetadata.MetadataCalls(media.ID))
	require.ElementsMatch(t, []string{"special-simple", "special-smart"}, result.IncludedSpecialProviders)
	require.Len(t, result.Torrents, 3)
	require.Equal(t, "main-hash", result.Torrents[0].InfoHash)
	require.Equal(t, "special-smart-hash", result.Torrents[1].InfoHash)
	require.Equal(t, "special-simple-hash", result.Torrents[2].InfoHash)
	require.NotNil(t, result.AnimeMetadata)
	require.Empty(t, result.Previews)

	lastSmart := mainProvider.lastSmartOptions()
	require.Equal(t, 1001, lastSmart.AnidbAID)
	require.Equal(t, 505, lastSmart.AnidbEID)
	require.Equal(t, 12, lastSmart.Media.AbsoluteSeasonOffset)
	require.Equal(t, 5, lastSmart.EpisodeNumber)
	require.Equal(t, "Example Show", lastSmart.Query)
}

func TestSearchAnimeIgnoresIncludeSpecialProviders(t *testing.T) {
	metadataCache.Clear()
	// legacy includeSpecialProviders stays accepted but no longer adds providers
	mainProvider := newStubAnimeProvider(hibiketorrent.AnimeProviderSettings{
		Type:           hibiketorrent.AnimeProviderTypeMain,
		CanSmartSearch: true,
	})
	mainProvider.smartResults = []*hibiketorrent.AnimeTorrent{{
		Name:          "[Main] Example Show - 05 (1080p).mkv",
		InfoHash:      "main-hash",
		Seeders:       20,
		EpisodeNumber: 5,
	}}

	specialProvider := newStubAnimeProvider(hibiketorrent.AnimeProviderSettings{
		Type:           hibiketorrent.AnimeProviderTypeSpecial,
		CanSmartSearch: true,
	})
	specialProvider.smartResults = []*hibiketorrent.AnimeTorrent{{
		Name:     "[Special] Example Show Batch (1080p).mkv",
		InfoHash: "special-hash",
		Seeders:  8,
		IsBatch:  true,
	}}

	repo := newTorrentRepositoryForTests(map[string]*stubAnimeProvider{
		"main":    mainProvider,
		"special": specialProvider,
	}, testmocks.NewFakeMetadataProviderBuilder().Build())

	media := testmocks.NewBaseAnimeBuilder(201, "Example Show").WithEpisodes(24).Build()

	result, err := repo.SearchAnime(context.Background(), AnimeSearchOptions{
		Provider:                "main",
		Type:                    AnimeSearchTypeSmart,
		Media:                   media,
		Query:                   "Example Show",
		Batch:                   true,
		EpisodeNumber:           5,
		IncludeSpecialProviders: true,
		SkipPreviews:            true,
	})

	require.NoError(t, err)
	require.Equal(t, 1, mainProvider.smartCallsCount())
	require.Equal(t, 0, specialProvider.smartCallsCount())
	require.Empty(t, result.IncludedSpecialProviders)
	require.Len(t, result.Torrents, 1)
}

func TestSearchAnimeErrorsWhenNoProviderExists(t *testing.T) {
	metadataCache.Clear()
	repo := newTorrentRepositoryForTests(nil, testmocks.NewFakeMetadataProviderBuilder().Build())
	media := testmocks.NewBaseAnime(300, "Missing Provider")

	result, err := repo.SearchAnime(context.Background(), AnimeSearchOptions{
		Provider: "missing",
		Type:     AnimeSearchTypeSimple,
		Media:    media,
		Query:    "Missing Provider",
	})

	require.Nil(t, result)
	require.EqualError(t, err, "torrent provider not found")
}

type stubAnimeProvider struct {
	settings      hibiketorrent.AnimeProviderSettings
	searchResults []*hibiketorrent.AnimeTorrent
	smartResults  []*hibiketorrent.AnimeTorrent
	searchErr     error
	smartErr      error

	mu            sync.Mutex
	searchCalls   int
	smartCalls    int
	lastSearchOpt hibiketorrent.AnimeSearchOptions
	lastSmartOpt  hibiketorrent.AnimeSmartSearchOptions
}

func newStubAnimeProvider(settings hibiketorrent.AnimeProviderSettings) *stubAnimeProvider {
	return &stubAnimeProvider{settings: settings}
}

func (s *stubAnimeProvider) Search(opts hibiketorrent.AnimeSearchOptions) ([]*hibiketorrent.AnimeTorrent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.searchCalls++
	s.lastSearchOpt = opts
	return cloneTorrents(s.searchResults), s.searchErr
}

func (s *stubAnimeProvider) SmartSearch(opts hibiketorrent.AnimeSmartSearchOptions) ([]*hibiketorrent.AnimeTorrent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.smartCalls++
	s.lastSmartOpt = opts
	return cloneTorrents(s.smartResults), s.smartErr
}

func (s *stubAnimeProvider) GetTorrentInfoHash(torrent *hibiketorrent.AnimeTorrent) (string, error) {
	if torrent == nil {
		return "", nil
	}
	return torrent.InfoHash, nil
}

func (s *stubAnimeProvider) GetTorrentMagnetLink(*hibiketorrent.AnimeTorrent) (string, error) {
	return "magnet:?xt=urn:btih:test", nil
}

func (s *stubAnimeProvider) GetLatest() ([]*hibiketorrent.AnimeTorrent, error) {
	return nil, nil
}

func (s *stubAnimeProvider) GetSettings() hibiketorrent.AnimeProviderSettings {
	return s.settings
}

func (s *stubAnimeProvider) searchCallsCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.searchCalls
}

func (s *stubAnimeProvider) smartCallsCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.smartCalls
}

func (s *stubAnimeProvider) lastSearchOptions() hibiketorrent.AnimeSearchOptions {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastSearchOpt
}

func (s *stubAnimeProvider) lastSmartOptions() hibiketorrent.AnimeSmartSearchOptions {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastSmartOpt
}

func newTorrentRepositoryForTests(providers map[string]*stubAnimeProvider, metadataProvider metadata_provider.Provider) *Repository {
	bank := extension.NewUnifiedBank()
	for id, provider := range providers {
		bank.Set(id, extension.NewAnimeTorrentProviderExtension(&extension.Extension{
			ID:          id,
			Name:        id,
			Version:     "1.0.0",
			ManifestURI: "builtin",
			Language:    extension.LanguageGo,
			Type:        extension.TypeAnimeTorrentProvider,
		}, provider))
	}

	return NewRepository(&NewRepositoryOptions{
		Logger:              new(zerolog.Nop()),
		MetadataProviderRef: util.NewRef[metadata_provider.Provider](metadataProvider),
		ExtensionBankRef:    util.NewRef(bank),
	})
}

func cloneTorrents(in []*hibiketorrent.AnimeTorrent) []*hibiketorrent.AnimeTorrent {
	if in == nil {
		return nil
	}
	out := make([]*hibiketorrent.AnimeTorrent, 0, len(in))
	for _, torrent := range in {
		if torrent == nil {
			out = append(out, nil)
			continue
		}
		out = append(out, new(*torrent))
	}
	return out
}

func TestSearchAnimeSkipPreviewsCacheRegression(t *testing.T) {
	metadataCache.Clear()

	mainProvider := newStubAnimeProvider(hibiketorrent.AnimeProviderSettings{
		Type:           hibiketorrent.AnimeProviderTypeMain,
		CanSmartSearch: true,
	})
	mainProvider.smartResults = []*hibiketorrent.AnimeTorrent{{
		Name:          "[Main] Example Show - 05 (1080p).mkv",
		InfoHash:      "main-hash",
		Seeders:       20,
		EpisodeNumber: 5,
	}}

	fakeMetadata := testmocks.NewFakeMetadataProviderBuilder().WithAnimeMetadata(200, &metadata.AnimeMetadata{
		Episodes: map[string]*metadata.EpisodeMetadata{
			"5": {Episode: "5", AnidbEid: 505},
		},
		Mappings: &metadata.AnimeMappings{AnidbId: 1001},
	}).Build()

	repo := newTorrentRepositoryForTests(map[string]*stubAnimeProvider{
		"main": mainProvider,
	}, fakeMetadata)
	repo.SetSettings(&RepositorySettings{DefaultAnimeProvider: "main"})

	media := testmocks.NewBaseAnimeBuilder(200, "Example Show").WithEpisodes(24).Build()

	// first search with SkipPreviews = true (e.g. autoselect)
	result1, err := repo.SearchAnime(context.Background(), AnimeSearchOptions{
		Provider:      "main",
		Type:          AnimeSearchTypeSmart,
		Media:         media,
		Query:         "Example Show",
		EpisodeNumber: 5,
		SkipPreviews:  true,
	})
	require.NoError(t, err)
	require.Empty(t, result1.Previews)

	// second search with SkipPreviews = false
	result2, err := repo.SearchAnime(context.Background(), AnimeSearchOptions{
		Provider:      "main",
		Type:          AnimeSearchTypeSmart,
		Media:         media,
		Query:         "Example Show",
		EpisodeNumber: 5,
		SkipPreviews:  false,
	})
	require.NoError(t, err)
	require.NotEmpty(t, result2.Previews)
}

func TestSearchAnimeEmptyTypeNoCacheCollision(t *testing.T) {
	metadataCache.Clear()

	mainProvider := newStubAnimeProvider(hibiketorrent.AnimeProviderSettings{
		Type:           hibiketorrent.AnimeProviderTypeMain,
		CanSmartSearch: false,
	})

	repo := newTorrentRepositoryForTests(map[string]*stubAnimeProvider{
		"main": mainProvider,
	}, testmocks.NewFakeMetadataProviderBuilder().Build())
	repo.SetSettings(&RepositorySettings{DefaultAnimeProvider: "main"})

	mediaA := testmocks.NewBaseAnimeBuilder(301, "Show A").Build()

	_, err := repo.SearchAnime(context.Background(), AnimeSearchOptions{
		Provider: "main",
		Type:     "", // no type
		Media:    mediaA,
		Query:    "Show A",
	})
	require.NoError(t, err)

	// verify that nothing was cached under the empty key ""
	cache := getAnimeSearchCache(repo.animeProviderSearchCaches, "main")
	_, found := cache.Get("")
	require.False(t, found)
}
