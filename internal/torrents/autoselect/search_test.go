package autoselect

import (
	"context"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata_provider"
	"seanime/internal/extension"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/library/anime"
	"seanime/internal/test_utils"
	itorrent "seanime/internal/torrents/torrent"
	"seanime/internal/util"
	"seanime/internal/util/filecache"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// FakeSearchProvider is a fake torrent provider for testing search functionality
type FakeSearchProvider struct {
	SearchResults    map[string][]*hibiketorrent.AnimeTorrent // keyed by resolution
	CanSmartSearch   bool
	SearchCallCount  int
	LastSearchQuery  string
	LastResolution   string
	LastBatchSetting bool
}

func (f *FakeSearchProvider) Search(opts hibiketorrent.AnimeSearchOptions) ([]*hibiketorrent.AnimeTorrent, error) {
	f.SearchCallCount++
	f.LastSearchQuery = opts.Query
	f.LastResolution = ""
	f.LastBatchSetting = false

	// Return default torrents for simple search
	if torrents, ok := f.SearchResults[""]; ok {
		return torrents, nil
	}
	return []*hibiketorrent.AnimeTorrent{}, nil
}

func (f *FakeSearchProvider) SmartSearch(opts hibiketorrent.AnimeSmartSearchOptions) ([]*hibiketorrent.AnimeTorrent, error) {
	f.SearchCallCount++
	f.LastResolution = opts.Resolution
	f.LastBatchSetting = opts.Batch

	// Return torrents matching the resolution
	if torrents, ok := f.SearchResults[opts.Resolution]; ok {
		return torrents, nil
	}
	return []*hibiketorrent.AnimeTorrent{}, nil
}

func (f *FakeSearchProvider) GetTorrentInfoHash(torrent *hibiketorrent.AnimeTorrent) (string, error) {
	return torrent.InfoHash, nil
}

func (f *FakeSearchProvider) GetTorrentMagnetLink(torrent *hibiketorrent.AnimeTorrent) (string, error) {
	return torrent.MagnetLink, nil
}

func (f *FakeSearchProvider) GetLatest() ([]*hibiketorrent.AnimeTorrent, error) {
	return []*hibiketorrent.AnimeTorrent{}, nil
}

func (f *FakeSearchProvider) GetSettings() hibiketorrent.AnimeProviderSettings {
	return hibiketorrent.AnimeProviderSettings{
		CanSmartSearch:     f.CanSmartSearch,
		SmartSearchFilters: nil,
		SupportsAdult:      false,
		Type:               "main",
	}
}

var _ hibiketorrent.AnimeProvider = (*FakeSearchProvider)(nil)

// setupTestAutoSelect creates an AutoSelect instance with a fake provider
func setupTestAutoSelect(t *testing.T, provider *FakeSearchProvider) *AutoSelect {
	logger := util.NewLogger()

	tempDir := t.TempDir()
	filecacher, err := filecache.NewCacher(tempDir)
	require.NoError(t, err)

	extensionBankRef := util.NewRef(extension.NewUnifiedBank())

	// Create fake extension
	ext := extension.NewAnimeTorrentProviderExtension(&extension.Extension{
		ID:   "fake-provider",
		Type: extension.TypeAnimeTorrentProvider,
		Name: "Fake Provider",
	}, provider)

	extensionBankRef.Get().Set("fake-provider", ext)

	metadataProvider := metadata_provider.NewProvider(&metadata_provider.NewProviderImplOptions{
		Logger:           logger,
		FileCacher:       filecacher,
		Database:         nil,
		ExtensionBankRef: extensionBankRef,
	})

	torrentRepository := itorrent.NewRepository(&itorrent.NewRepositoryOptions{
		Logger:              logger,
		MetadataProviderRef: util.NewRef(metadataProvider),
		ExtensionBankRef:    extensionBankRef,
	})

	return New(&NewAutoSelectOptions{
		Logger:            logger,
		TorrentRepository: torrentRepository,
		MetadataProvider:  util.NewRef(metadataProvider),
		Platform:          nil,
		OnEvent:           nil,
	})
}

// createTestMedia creates a mock anime for testing
func createTestMedia(t *testing.T) *anilist.CompleteAnime {
	return &anilist.CompleteAnime{
		ID: 21,
		Title: &anilist.CompleteAnime_Title{
			Romaji:  new("One Piece"),
			English: new("One Piece"),
		},
		Status: new(anilist.MediaStatusReleasing),
		Format: new(anilist.MediaFormatTv),
		StartDate: &anilist.CompleteAnime_StartDate{
			Year: new(1999),
		},
		IsAdult: new(false),
	}
}

func TestSearchFromProvider_SingleResolution(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	media := createTestMedia(t)
	episodeNumber := 1000

	t1080p := []*hibiketorrent.AnimeTorrent{
		{Name: "[SubsPlease] One Piece - 1000 (1080p).mkv", InfoHash: "hash1", Seeders: 100},
		{Name: "[Erai-raws] One Piece - 1000 [1080p].mkv", InfoHash: "hash2", Seeders: 150},
	}

	provider := &FakeSearchProvider{
		SearchResults: map[string][]*hibiketorrent.AnimeTorrent{
			"1080p": t1080p,
		},
		CanSmartSearch: true,
	}

	autoSelect := setupTestAutoSelect(t, provider)

	profile := &anime.AutoSelectProfile{
		Providers:   []string{"fake-provider"},
		Resolutions: []string{"1080p"},
	}

	ctx := context.Background()
	torrents, err := autoSelect.searchFromProvider(ctx, "fake-provider", media, episodeNumber, false, profile)

	assert.NoError(t, err)
	assert.Len(t, torrents, 2)
	assert.Equal(t, "1080p", provider.LastResolution)
	assert.Equal(t, 1, provider.SearchCallCount)
}

func TestSearchFromProvider_MultipleResolutions_FirstSucceeds(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	media := createTestMedia(t)
	episodeNumber := 1000

	t1080p := []*hibiketorrent.AnimeTorrent{
		{Name: "[SubsPlease] One Piece - 1000 (1080p).mkv", InfoHash: "hash1", Seeders: 100},
	}

	provider := &FakeSearchProvider{
		SearchResults: map[string][]*hibiketorrent.AnimeTorrent{
			"1080p": t1080p,
			"720p":  {}, // Empty
		},
		CanSmartSearch: true,
	}

	autoSelect := setupTestAutoSelect(t, provider)

	profile := &anime.AutoSelectProfile{
		Providers:   []string{"fake-provider"},
		Resolutions: []string{"1080p", "720p"}, // 1080p should succeed first
	}

	ctx := context.Background()
	torrents, err := autoSelect.searchFromProvider(ctx, "fake-provider", media, episodeNumber, false, profile)

	assert.NoError(t, err)
	assert.Len(t, torrents, 1)
	assert.Equal(t, "1080p", provider.LastResolution)
	assert.Equal(t, 1, provider.SearchCallCount) // Should only call once since first resolution succeeds
}

func TestSearchFromProvider_MultipleResolutions_Fallback(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	media := createTestMedia(t)
	episodeNumber := 1000

	t720p := []*hibiketorrent.AnimeTorrent{
		{Name: "[SubsPlease] One Piece - 1000 (720p).mkv", InfoHash: "hash1", Seeders: 80},
	}

	provider := &FakeSearchProvider{
		SearchResults: map[string][]*hibiketorrent.AnimeTorrent{
			"1080p": {}, // Empty, should fallback
			"720p":  t720p,
		},
		CanSmartSearch: true,
	}

	autoSelect := setupTestAutoSelect(t, provider)

	profile := &anime.AutoSelectProfile{
		Providers:   []string{"fake-provider"},
		Resolutions: []string{"1080p", "720p"},
	}

	ctx := context.Background()
	torrents, err := autoSelect.searchFromProvider(ctx, "fake-provider", media, episodeNumber, false, profile)

	assert.NoError(t, err)
	assert.Len(t, torrents, 1)
	assert.Equal(t, "720p", provider.LastResolution) // Should use 720p after 1080p fails
	assert.Equal(t, 2, provider.SearchCallCount)     // Should try both resolutions
}

func TestSearchFromProvider_AllResolutionsFail(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	media := createTestMedia(t)
	episodeNumber := 1000

	provider := &FakeSearchProvider{
		SearchResults: map[string][]*hibiketorrent.AnimeTorrent{
			"1080p": {},
			"720p":  {},
			"480p":  {},
		},
		CanSmartSearch: true,
	}

	autoSelect := setupTestAutoSelect(t, provider)

	profile := &anime.AutoSelectProfile{
		Providers:   []string{"fake-provider"},
		Resolutions: []string{"1080p", "720p", "480p"},
	}

	ctx := context.Background()
	torrents, err := autoSelect.searchFromProvider(ctx, "fake-provider", media, episodeNumber, false, profile)

	assert.Error(t, err)
	assert.Nil(t, torrents)
	assert.Contains(t, err.Error(), "no torrents found with any resolution")
	assert.Equal(t, 3, provider.SearchCallCount) // Should try all resolutions
}

func TestSearchFromProvider_NoResolutionsInProfile(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	media := createTestMedia(t)
	episodeNumber := 1000

	tAny := []*hibiketorrent.AnimeTorrent{
		{Name: "[SubsPlease] One Piece - 1000.mkv", InfoHash: "hash1", Seeders: 100},
	}

	provider := &FakeSearchProvider{
		SearchResults: map[string][]*hibiketorrent.AnimeTorrent{
			"": tAny, // Empty resolution key for "any"
		},
		CanSmartSearch: true,
	}

	autoSelect := setupTestAutoSelect(t, provider)

	profile := &anime.AutoSelectProfile{
		Providers:   []string{"fake-provider"},
		Resolutions: []string{}, // No resolutions specified
	}

	ctx := context.Background()
	torrents, err := autoSelect.searchFromProvider(ctx, "fake-provider", media, episodeNumber, false, profile)

	assert.NoError(t, err)
	assert.Len(t, torrents, 1)
	assert.Equal(t, "", provider.LastResolution) // Should search with empty resolution
	assert.Equal(t, 1, provider.SearchCallCount)
}

func TestSearchFromProvider_BatchFallback(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	media := createTestMedia(t)
	episodeNumber := 1

	tSingle := []*hibiketorrent.AnimeTorrent{
		{Name: "[SubsPlease] One Piece - 01 (1080p).mkv", InfoHash: "hash1", Seeders: 100},
	}

	provider := &FakeSearchProvider{
		SearchResults: map[string][]*hibiketorrent.AnimeTorrent{
			"1080p": tSingle,
		},
		CanSmartSearch: true,
	}

	autoSelect := setupTestAutoSelect(t, provider)

	profile := &anime.AutoSelectProfile{
		Providers:   []string{"fake-provider"},
		Resolutions: []string{"1080p"},
	}

	ctx := context.Background()
	torrents, err := autoSelect.searchFromProvider(ctx, "fake-provider", media, episodeNumber, true, profile)

	assert.NoError(t, err)
	assert.NotEmpty(t, torrents)
	// Should try batch first, then fallback to single
	assert.Equal(t, 2, provider.SearchCallCount)
}

func TestSearchFromProviders_MultipleProviders(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	media := createTestMedia(t)
	episodeNumber := 1000

	t1080p := []*hibiketorrent.AnimeTorrent{
		{Name: "[SubsPlease] One Piece - 1000 (1080p).mkv", InfoHash: "hash1", Seeders: 100},
		{Name: "[Erai-raws] One Piece - 1000 [1080p].mkv", InfoHash: "hash2", Seeders: 150},
	}

	provider := &FakeSearchProvider{
		SearchResults: map[string][]*hibiketorrent.AnimeTorrent{
			"1080p": t1080p,
		},
		CanSmartSearch: true,
	}

	autoSelect := setupTestAutoSelect(t, provider)

	profile := &anime.AutoSelectProfile{
		Providers:   []string{"fake-provider"},
		Resolutions: []string{"1080p"},
	}

	ctx := context.Background()
	torrents, err := autoSelect.searchFromProviders(ctx, []string{"fake-provider"}, media, episodeNumber, false, profile)

	assert.NoError(t, err)
	assert.Len(t, torrents, 2)
}

func TestSearchFromProviders_Deduplication(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	media := createTestMedia(t)
	episodeNumber := 1000

	t1080p := []*hibiketorrent.AnimeTorrent{
		{Name: "[SubsPlease] One Piece - 1000 (1080p).mkv", InfoHash: "hash1", Seeders: 100},
		{Name: "[SubsPlease] One Piece - 1000 (1080p).mkv", InfoHash: "hash1", Seeders: 100}, // Duplicate
		{Name: "[Erai-raws] One Piece - 1000 [1080p].mkv", InfoHash: "hash2", Seeders: 150},
	}

	provider := &FakeSearchProvider{
		SearchResults: map[string][]*hibiketorrent.AnimeTorrent{
			"1080p": t1080p,
		},
		CanSmartSearch: true,
	}

	autoSelect := setupTestAutoSelect(t, provider)

	profile := &anime.AutoSelectProfile{
		Providers:   []string{"fake-provider"},
		Resolutions: []string{"1080p"},
	}

	ctx := context.Background()
	torrents, err := autoSelect.searchFromProviders(ctx, []string{"fake-provider"}, media, episodeNumber, false, profile)

	assert.NoError(t, err)
	assert.Len(t, torrents, 2) // Should deduplicate by infohash
}

func TestSearch_Integration(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	media := createTestMedia(t)
	episodeNumber := 1000

	t720p := []*hibiketorrent.AnimeTorrent{
		{Name: "[SubsPlease] One Piece - 1000 (720p).mkv", InfoHash: "hash1", Seeders: 100},
	}

	provider := &FakeSearchProvider{
		SearchResults: map[string][]*hibiketorrent.AnimeTorrent{
			"1080p": {}, // Empty, should fallback to 720p
			"720p":  t720p,
		},
		CanSmartSearch: true,
	}

	autoSelect := setupTestAutoSelect(t, provider)

	profile := &anime.AutoSelectProfile{
		Providers:   []string{"fake-provider"},
		Resolutions: []string{"1080p", "720p"},
	}

	ctx := context.Background()
	torrents, err := autoSelect.search(ctx, media, episodeNumber, profile)

	assert.NoError(t, err)
	assert.Len(t, torrents, 1)
	assert.Equal(t, "hash1", torrents[0].InfoHash)
}

func TestShouldSearchBatch(t *testing.T) {
	logger := zerolog.Nop()
	autoSelect := New(&NewAutoSelectOptions{
		Logger: &logger,
	})

	now := time.Now().UTC()
	threeWeeksAgo := now.AddDate(0, 0, -21)
	exactlyTwoWeeksAgo := now.AddDate(0, 0, -14)
	oneWeekAgo := now.AddDate(0, 0, -7)
	yesterday := now.AddDate(0, 0, -1)
	oldDate := now.AddDate(-5, 0, 0) // 5 years ago

	tests := []struct {
		name     string
		media    *anilist.CompleteAnime
		expected bool
	}{
		{
			name: "Finished anime, ended more than 2 weeks ago",
			media: &anilist.CompleteAnime{
				Status: new(anilist.MediaStatusFinished),
				Format: new(anilist.MediaFormatTv),
				EndDate: &anilist.CompleteAnime_EndDate{
					Year:  new(threeWeeksAgo.Year()),
					Month: new(int(threeWeeksAgo.Month())),
					Day:   new(threeWeeksAgo.Day()),
				},
			},
			expected: true,
		},
		{
			name: "Finished anime, ended exactly 2 weeks ago",
			media: &anilist.CompleteAnime{
				Status: new(anilist.MediaStatusFinished),
				Format: new(anilist.MediaFormatTv),
				EndDate: &anilist.CompleteAnime_EndDate{
					Year:  new(exactlyTwoWeeksAgo.Year()),
					Month: new(int(exactlyTwoWeeksAgo.Month())),
					Day:   new(exactlyTwoWeeksAgo.Day()),
				},
			},
			expected: true,
		},
		{
			name: "Finished anime, ended less than 2 weeks ago",
			media: &anilist.CompleteAnime{
				Status: new(anilist.MediaStatusFinished),
				Format: new(anilist.MediaFormatTv),
				EndDate: &anilist.CompleteAnime_EndDate{
					Year:  new(oneWeekAgo.Year()),
					Month: new(int(oneWeekAgo.Month())),
					Day:   new(oneWeekAgo.Day()),
				},
			},
			expected: false,
		},
		{
			name: "Finished anime, ended yesterday",
			media: &anilist.CompleteAnime{
				Status: new(anilist.MediaStatusFinished),
				Format: new(anilist.MediaFormatTv),
				EndDate: &anilist.CompleteAnime_EndDate{
					Year:  new(yesterday.Year()),
					Month: new(int(yesterday.Month())),
					Day:   new(yesterday.Day()),
				},
			},
			expected: false,
		},
		{
			name: "Finished anime, no end date",
			media: &anilist.CompleteAnime{
				Status: new(anilist.MediaStatusFinished),
				Format: new(anilist.MediaFormatTv),
			},
			expected: true,
		},
		{
			name: "Finished anime, partial end date (no day)",
			media: &anilist.CompleteAnime{
				Status: new(anilist.MediaStatusFinished),
				Format: new(anilist.MediaFormatTv),
				EndDate: &anilist.CompleteAnime_EndDate{
					Year:  new(threeWeeksAgo.Year()),
					Month: new(int(threeWeeksAgo.Month())),
				},
			},
			expected: true,
		},
		{
			name: "Currently airing anime",
			media: &anilist.CompleteAnime{
				Status: new(anilist.MediaStatusReleasing),
				Format: new(anilist.MediaFormatTv),
				EndDate: &anilist.CompleteAnime_EndDate{
					Year:  new(oldDate.Year()),
					Month: new(int(oldDate.Month())),
					Day:   new(oldDate.Day()),
				},
			},
			expected: false,
		},
		{
			name: "Movie, finished",
			media: &anilist.CompleteAnime{
				Status: new(anilist.MediaStatusFinished),
				Format: new(anilist.MediaFormatMovie),
				EndDate: &anilist.CompleteAnime_EndDate{
					Year:  new(oldDate.Year()),
					Month: new(int(oldDate.Month())),
					Day:   new(oldDate.Day()),
				},
			},
			expected: false,
		},
		{
			name: "Old finished anime",
			media: &anilist.CompleteAnime{
				Status: new(anilist.MediaStatusFinished),
				Format: new(anilist.MediaFormatTv),
				EndDate: &anilist.CompleteAnime_EndDate{
					Year:  new(oldDate.Year()),
					Month: new(int(oldDate.Month())),
					Day:   new(oldDate.Day()),
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := autoSelect.shouldSearchBatch(tt.media)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetProvidersToSearch(t *testing.T) {
	logger := util.NewLogger()

	tempDir := t.TempDir()
	filecacher, err := filecache.NewCacher(tempDir)
	require.NoError(t, err)

	extensionBankRef := util.NewRef(extension.NewUnifiedBank())

	// Create fake extensions
	provider1 := &FakeSearchProvider{CanSmartSearch: false}
	ext1 := extension.NewAnimeTorrentProviderExtension(&extension.Extension{
		ID:   "provider1",
		Type: extension.TypeAnimeTorrentProvider,
		Name: "Provider 1",
	}, provider1)

	provider2 := &FakeSearchProvider{CanSmartSearch: false}
	ext2 := extension.NewAnimeTorrentProviderExtension(&extension.Extension{
		ID:   "provider2",
		Type: extension.TypeAnimeTorrentProvider,
		Name: "Provider 2",
	}, provider2)

	extensionBankRef.Get().Set("provider1", ext1)
	extensionBankRef.Get().Set("provider2", ext2)

	metadataProvider := metadata_provider.NewProvider(&metadata_provider.NewProviderImplOptions{
		Logger:           logger,
		FileCacher:       filecacher,
		Database:         nil,
		ExtensionBankRef: extensionBankRef,
	})

	torrentRepository := itorrent.NewRepository(&itorrent.NewRepositoryOptions{
		Logger:              logger,
		MetadataProviderRef: util.NewRef(metadataProvider),
		ExtensionBankRef:    extensionBankRef,
	})

	// Set default provider through settings
	torrentRepository.SetSettings(&itorrent.RepositorySettings{
		DefaultAnimeProvider: "provider1",
	})

	autoSelect := New(&NewAutoSelectOptions{
		Logger:            logger,
		TorrentRepository: torrentRepository,
		MetadataProvider:  util.NewRef(metadataProvider),
		Platform:          nil,
		OnEvent:           nil,
	})

	tests := []struct {
		name     string
		profile  *anime.AutoSelectProfile
		expected []string
	}{
		{
			name: "Use profile providers",
			profile: &anime.AutoSelectProfile{
				Providers: []string{"provider2", "provider1"},
			},
			expected: []string{"provider2", "provider1"},
		},
		{
			name: "Limit to max 3 providers",
			profile: &anime.AutoSelectProfile{
				Providers: []string{"provider1", "provider2", "provider1", "provider2"},
			},
			expected: []string{"provider1", "provider2", "provider1"},
		},
		{
			name:     "No profile - use default",
			profile:  nil,
			expected: []string{"provider1"},
		},
		{
			name: "Empty providers - use default",
			profile: &anime.AutoSelectProfile{
				Providers: []string{},
			},
			expected: []string{"provider1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := autoSelect.getProvidersToSearch(tt.profile)
			assert.Equal(t, tt.expected, result)
		})
	}
}
