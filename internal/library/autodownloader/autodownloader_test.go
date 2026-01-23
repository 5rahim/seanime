package autodownloader

import (
	"seanime/internal/database/models"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/library/anime"
	"seanime/internal/torrent_clients/torrent_client"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestGetRuleProfiles(t *testing.T) {
	ad := &AutoDownloader{}

	tests := []struct {
		name     string
		rule     *anime.AutoDownloaderRule
		expected []uint
		profiles []*anime.AutoDownloaderProfile
	}{
		{
			name: "1 specific profile + 2, 3 globals",
			rule: &anime.AutoDownloaderRule{ProfileID: lo.ToPtr(uint(1))},
			profiles: []*anime.AutoDownloaderProfile{
				{DbID: 1, Global: false},
				{DbID: 2, Global: true},
				{DbID: 3, Global: true},
				{DbID: 4, Global: false},
			},
			expected: []uint{1, 2, 3}, // specific profile 1 + global profiles 2,3
		},
		{
			name:     "only global profiles",
			rule:     &anime.AutoDownloaderRule{ProfileID: nil},
			expected: []uint{2, 3},
			profiles: []*anime.AutoDownloaderProfile{
				{DbID: 1, Global: false},
				{DbID: 2, Global: true},
				{DbID: 3, Global: true},
				{DbID: 4, Global: false},
			},
		},
		{
			name:     "1 specific profile (which is global) + 1, 2 global profiles",
			rule:     &anime.AutoDownloaderRule{ProfileID: lo.ToPtr(uint(2))},
			expected: []uint{2, 3},
			profiles: []*anime.AutoDownloaderProfile{
				{DbID: 1, Global: false},
				{DbID: 2, Global: true},
				{DbID: 3, Global: true},
				{DbID: 4, Global: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ad.getRuleProfiles(tt.rule, tt.profiles)
			resultIDs := lo.Map(result, func(p *anime.AutoDownloaderProfile, _ int) uint { return p.DbID })
			assert.ElementsMatchf(t, tt.expected, resultIDs, "Expected %v, got %v", tt.expected, resultIDs)
		})
	}
}

func TestIsTorrentAlreadyDownloaded(t *testing.T) {
	ad := &AutoDownloader{}

	existingTorrents := []*torrent_client.Torrent{
		{Hash: "hash1"},
		{Hash: "hash2"},
		{Hash: "hash3"},
	}

	tests := []struct {
		name     string
		torrent  *NormalizedTorrent
		expected bool
	}{
		{
			name: "torrent exists",
			torrent: &NormalizedTorrent{
				AnimeTorrent: hibiketorrent.AnimeTorrent{InfoHash: "hash2"},
			},
			expected: true,
		},
		{
			name: "torrent does not exist",
			torrent: &NormalizedTorrent{
				AnimeTorrent: hibiketorrent.AnimeTorrent{InfoHash: "hash999"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ad.isTorrentAlreadyDownloaded(tt.torrent, existingTorrents)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsEpisodeAlreadyHandled(t *testing.T) {
	ad := &AutoDownloader{}

	// Create mock local files
	localFiles := []*anime.LocalFile{
		{MediaId: 1, Metadata: &anime.LocalFileMetadata{Episode: 5, Type: anime.LocalFileTypeMain}},
		{MediaId: 1, Metadata: &anime.LocalFileMetadata{Episode: 6, Type: anime.LocalFileTypeMain}},
		{MediaId: 2, Metadata: &anime.LocalFileMetadata{Episode: 1, Type: anime.LocalFileTypeMain}},
	}
	lfWrapper := anime.NewLocalFileWrapper(localFiles)

	queuedItems := []*models.AutoDownloaderItem{
		{MediaID: 1, Episode: 7},
		{MediaID: 1, Episode: 8},
	}

	tests := []struct {
		name     string
		episode  int
		mediaId  int
		expected bool
	}{
		{
			name:     "episode in local files",
			episode:  5,
			mediaId:  1,
			expected: true,
		},
		{
			name:     "episode in queue",
			episode:  7,
			mediaId:  1,
			expected: true,
		},
		{
			name:     "episode not handled",
			episode:  10,
			mediaId:  1,
			expected: false,
		},
		{
			name:     "different media id",
			episode:  5,
			mediaId:  99,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ad.isEpisodeAlreadyHandled(tt.episode, tt.mediaId, lfWrapper, queuedItems)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateCandidateScore(t *testing.T) {
	ad := &AutoDownloader{}

	tests := []struct {
		name        string
		torrent     *NormalizedTorrent
		profiles    []*anime.AutoDownloaderProfile
		expected    int
		expectedMin int
	}{
		{
			name: "single profile, single condition",
			torrent: &NormalizedTorrent{
				AnimeTorrent: hibiketorrent.AnimeTorrent{Name: "[SubsPlease] One Piece - 1000 (1080p) [ABCD].mkv"},
			},
			profiles: []*anime.AutoDownloaderProfile{
				{
					MinimumScore: 10,
					Conditions: []anime.AutoDownloaderCondition{
						{Term: "1080p", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 15},
					},
				},
			},
			expected:    15,
			expectedMin: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, minScore := ad.calculateCandidateScore(tt.torrent, tt.profiles)
			assert.Equal(t, tt.expected, score)
			assert.Equal(t, tt.expectedMin, minScore)
		})
	}
}

func TestSelectBestCandidate(t *testing.T) {
	ad := &AutoDownloader{}

	tests := []struct {
		name            string
		candidates      []*Candidate
		expectedTorrent string
	}{
		{
			name: "highest score wins",
			candidates: []*Candidate{
				{Score: 10, Torrent: &NormalizedTorrent{AnimeTorrent: hibiketorrent.AnimeTorrent{Name: "a1", Seeders: 50}}},
				{Score: 25, Torrent: &NormalizedTorrent{AnimeTorrent: hibiketorrent.AnimeTorrent{Name: "a2", Seeders: 30}}},
				{Score: 15, Torrent: &NormalizedTorrent{AnimeTorrent: hibiketorrent.AnimeTorrent{Name: "a3", Seeders: 100}}},
			},
			expectedTorrent: "a2",
		},
		{
			name: "same score, more seeders wins",
			candidates: []*Candidate{
				{Score: 20, Torrent: &NormalizedTorrent{AnimeTorrent: hibiketorrent.AnimeTorrent{Name: "a1", Seeders: 50}}},
				{Score: 20, Torrent: &NormalizedTorrent{AnimeTorrent: hibiketorrent.AnimeTorrent{Name: "a2", Seeders: 150}}},
				{Score: 20, Torrent: &NormalizedTorrent{AnimeTorrent: hibiketorrent.AnimeTorrent{Name: "a3", Seeders: 100}}},
			},
			expectedTorrent: "a2",
		},
		{
			name: "single candidate",
			candidates: []*Candidate{
				{Score: 10, Torrent: &NormalizedTorrent{AnimeTorrent: hibiketorrent.AnimeTorrent{Name: "a1", Seeders: 10}}},
			},
			expectedTorrent: "a1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ad.selectBestCandidate(tt.candidates)
			assert.Equal(t, tt.expectedTorrent, result.Torrent.Name, "torrent names should match")
		})
	}
}

func TestReleaseGroupMapping(t *testing.T) {
	ad := &AutoDownloader{}
	profileID := uint(1)

	rules := []*anime.AutoDownloaderRule{
		{
			DbID:          10,
			ReleaseGroups: []string{"SubsPlease"},
			Resolutions:   []string{"1080p"}, // Explicit
			ProfileID:     nil,
		},
		{
			DbID:          20,
			ReleaseGroups: []string{"Erai-Raws"},
			Resolutions:   []string{}, // Should inherit
			ProfileID:     &profileID,
		},
		{
			DbID:          30,
			ReleaseGroups: []string{"SubsPlease"},
			Resolutions:   []string{}, // Should inherit from global
			ProfileID:     nil,
		},
	}

	profiles := []*anime.AutoDownloaderProfile{
		{
			DbID:        1,
			Global:      false,
			Resolutions: []string{"720p"},
		},
		{
			DbID:        2,
			Global:      true,
			Resolutions: []string{"4k"},
		},
	}

	resMap := ad.getReleaseGroupToResolutionsMap(rules, profiles)

	assert.Contains(t, resMap, "SubsPlease")
	assert.ElementsMatch(t, []string{"1080p", "4k"}, resMap["SubsPlease"])
	assert.Contains(t, resMap, "Erai-Raws")
	assert.ElementsMatch(t, []string{"720p", "4k"}, resMap["Erai-Raws"])
}

func TestReleaseGroupInheritanceMap(t *testing.T) {
	ad := &AutoDownloader{}

	tests := []struct {
		name                string
		releaseGroups       []string
		resolutions         []string
		expectedResolutions []string
		specificProfile     *anime.AutoDownloaderProfile
		profiles            []*anime.AutoDownloaderProfile
	}{
		{
			name:            "uses own resolutions",
			releaseGroups:   []string{"SubsPlease"},
			resolutions:     []string{"1080p"},
			specificProfile: nil,
			profiles: []*anime.AutoDownloaderProfile{
				{
					DbID:        1,
					Global:      false,
					Resolutions: []string{"720p"},
				},
			},
			expectedResolutions: []string{"1080p"},
		},
		{
			name:            "should inherit from global profile",
			releaseGroups:   []string{"SubsPlease"},
			resolutions:     []string{}, // not specified, should inherit
			specificProfile: nil,
			profiles: []*anime.AutoDownloaderProfile{
				{
					DbID:         1,
					Global:       true,
					Resolutions:  []string{"1080p", "720p"},
					MinimumScore: 0,
				},
			},
			expectedResolutions: []string{"1080p", "720p"},
		},
		{
			name:          "should ignore global profile",
			releaseGroups: []string{"SubsPlease"},
			resolutions:   []string{"720p"}, // explicitly specified
			profiles: []*anime.AutoDownloaderProfile{
				{
					DbID:         1,
					Global:       false,
					Resolutions:  []string{"1080p", "720p"},
					MinimumScore: 0,
				},
			},
			expectedResolutions: []string{"720p"},
		},
		{
			name:          "should ignore specific profile and global profile",
			releaseGroups: []string{"SubsPlease"},
			resolutions:   []string{"720p"}, // explicitly specified
			specificProfile: &anime.AutoDownloaderProfile{
				DbID:         1,
				Global:       false,
				Resolutions:  []string{"2160p"},
				MinimumScore: 0,
			},
			profiles: []*anime.AutoDownloaderProfile{
				{
					DbID:         1,
					Global:       false,
					Resolutions:  []string{"1080p"},
					MinimumScore: 0,
				},
			},
			expectedResolutions: []string{"720p"},
		},
		{
			name:          "should combine specific profile and global profile",
			releaseGroups: []string{"SubsPlease"},
			resolutions:   []string{}, // explicitly specified
			specificProfile: &anime.AutoDownloaderProfile{
				DbID:         1,
				Global:       false,
				Resolutions:  []string{"2160p"},
				MinimumScore: 0,
			},
			profiles: []*anime.AutoDownloaderProfile{
				{
					DbID:         1,
					Global:       false,
					Resolutions:  []string{"1080p", "720p"},
					MinimumScore: 0,
				},
			},
			expectedResolutions: []string{"2160p", "720p", "1080p"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := &anime.AutoDownloaderRule{
				Resolutions: tt.resolutions,
			}
			profiles := tt.profiles
			if tt.specificProfile != nil {
				profiles = append(profiles, tt.specificProfile)
			}
			res := ad.inheritResolutionsFromProfiles(rule, profiles)
			assert.ElementsMatchf(t, tt.expectedResolutions, res, "Expected %v, got %v", tt.expectedResolutions, res)
		})
	}
}

func TestTorrentScoring(t *testing.T) {
	ad := &AutoDownloader{}

	tests := []struct {
		name          string
		torrent       *NormalizedTorrent
		profiles      []*anime.AutoDownloaderProfile
		expectedScore int
		shouldSucceed bool
	}{
		{
			name: "should return 0 if no conditions match",
			torrent: &NormalizedTorrent{
				AnimeTorrent: hibiketorrent.AnimeTorrent{Name: "[SubsPlease] One Piece - 1000 (1080p) [ABCD].mkv"},
			},
			profiles: []*anime.AutoDownloaderProfile{
				{
					Conditions: []anime.AutoDownloaderCondition{
						{Term: "720p", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 10},
					},
				},
			},
			expectedScore: 0,
			shouldSucceed: true,
		},
		{
			name: "should return 0 if no conditions match and no global profile",
			torrent: &NormalizedTorrent{
				AnimeTorrent: hibiketorrent.AnimeTorrent{Name: "[EMBER] One Piece - 1000 (1080p) [ABCD].mkv"},
			},
			profiles: []*anime.AutoDownloaderProfile{
				{
					Conditions: []anime.AutoDownloaderCondition{
						{Term: "720p", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 10},
						{Term: "1080p", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 20},
					},
					MinimumScore: 10,
					Global:       true,
				},
				{
					Conditions: []anime.AutoDownloaderCondition{
						{Term: "(EMBER)", IsRegex: true, Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: -20},
					},
					MinimumScore: 0,
					Global:       true,
				},
			},
			expectedScore: 0,
			shouldSucceed: false,
		},
		{
			name: "multiple conditions accumulate positive scores",
			torrent: &NormalizedTorrent{
				AnimeTorrent: hibiketorrent.AnimeTorrent{Name: "[SubsPlease] Frieren - 12 (1080p) [x265] [10-bit].mkv"},
			},
			profiles: []*anime.AutoDownloaderProfile{
				{
					Conditions: []anime.AutoDownloaderCondition{
						{Term: "1080p", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 20},
						{Term: "SubsPlease", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 15},
						{Term: "x265", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 10},
						{Term: "10-bit", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 5},
					},
					MinimumScore: 30,
					Global:       true,
				},
			},
			expectedScore: 50,
			shouldSucceed: true,
		},
		{
			name: "regex patterns with mixed positive and negative scores",
			torrent: &NormalizedTorrent{
				AnimeTorrent: hibiketorrent.AnimeTorrent{Name: "[HorribleSubs] Attack on Titan - 25 [720p].mkv"},
			},
			profiles: []*anime.AutoDownloaderProfile{
				{
					Conditions: []anime.AutoDownloaderCondition{
						{Term: "\\[720p\\]", IsRegex: true, Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 10},
						{Term: "HorribleSubs", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 5},
						{Term: "\\[1080p\\]", IsRegex: true, Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 20},
					},
					MinimumScore: 5,
					Global:       true,
				},
			},
			expectedScore: 15,
			shouldSucceed: true,
		},
		{
			name: "multiple profiles combine scores to meet threshold",
			torrent: &NormalizedTorrent{
				AnimeTorrent: hibiketorrent.AnimeTorrent{Name: "[Erai-raws] Demon Slayer - 05 [1080p][Multiple Subtitle].mkv"},
			},
			profiles: []*anime.AutoDownloaderProfile{
				{
					Conditions: []anime.AutoDownloaderCondition{
						{Term: "1080p", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 15},
						{Term: "Multiple Subtitle", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 10},
					},
					MinimumScore: 0,
					Global:       true,
				},
				{
					Conditions: []anime.AutoDownloaderCondition{
						{Term: "Erai-raws", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 8},
					},
					MinimumScore: 25,
					Global:       true,
				},
			},
			expectedScore: 33,
			shouldSucceed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := 0
			minScore := 0
			for _, profile := range tt.profiles {
				score += ad.calculateTorrentScore(tt.torrent, profile)
				if profile.MinimumScore > minScore {
					minScore = profile.MinimumScore
				}
			}
			assert.Equal(t, tt.expectedScore, score, "Expected %d, got %d", tt.expectedScore, score)
			assert.Equal(t, tt.expectedScore >= minScore, tt.shouldSucceed, "Expected %v, got %v", tt.shouldSucceed, tt.expectedScore >= minScore)
		})
	}
}

func TestIsProfileValidChecks(t *testing.T) {
	ad := &AutoDownloader{}

	torrent := &NormalizedTorrent{
		AnimeTorrent: hibiketorrent.AnimeTorrent{
			Name:    "[SubsPlease] One Piece - 1000 (1080p).mkv",
			Seeders: 20,
			Size:    1073741824, // 1GB
		},
	}

	// Case 1: Pass
	p1 := &anime.AutoDownloaderProfile{
		MinSeeders: 10,
		MinSize:    "500MB",
		Conditions: []anime.AutoDownloaderCondition{
			{Term: "1080p", Action: anime.AutoDownloaderProfileRuleFormatActionRequire, ID: "1"},
		},
	}
	assert.True(t, ad.isProfileValidChecks(torrent, p1))

	// Case 2: Fail min seeders
	p2 := &anime.AutoDownloaderProfile{
		MinSeeders: 50,
	}
	assert.False(t, ad.isProfileValidChecks(torrent, p2))

	// Case 3: Fail block term
	p3 := &anime.AutoDownloaderProfile{
		Conditions: []anime.AutoDownloaderCondition{
			{Term: "SubsPlease", Action: anime.AutoDownloaderProfileRuleFormatActionBlock},
		},
	}
	assert.False(t, ad.isProfileValidChecks(torrent, p3))

	// Case 4: Fail require term
	p4 := &anime.AutoDownloaderProfile{
		Conditions: []anime.AutoDownloaderCondition{
			{Term: "HEVC", Action: anime.AutoDownloaderProfileRuleFormatActionRequire, ID: "2"},
		},
	}
	assert.False(t, ad.isProfileValidChecks(torrent, p4))
}

//func TestTorrentSelection(t *testing.T) {
//	test_utils.InitTestProvider(t, test_utils.Anilist())
//
//	// Load anime collection once for all tests
//	anilistClient := anilist.TestGetMockAnilistClient()
//	animeCollection, err := anilistClient.AnimeCollection(context.Background(), nil) // nil = boilerplate
//	require.NoError(t, err)
//
//	tests := []struct {
//		name             string
//		mediaId          int
//		userProgress     int
//		torrents         []*hibiketorrent.AnimeTorrent
//		rule             *anime.AutoDownloaderRule
//		profiles         []*anime.AutoDownloaderProfile
//		expectedResults  int
//		expectedEpisodes []struct {
//			episode     int
//			hash        string
//			score       int
//			shouldMatch func(result *SimulationResult) bool
//		}
//	}{
//		{
//			name:         "Should select SubsPlease 1080p torrents and block BadGroup",
//			mediaId:      154587, // Sousou no Frieren
//			userProgress: 0,
//			torrents: []*hibiketorrent.AnimeTorrent{
//				{
//					Name:       "[SubsPlease] Sousou no Frieren - 01 (1080p) [ABCD1234].mkv",
//					InfoHash:   "hash1",
//					Link:       "https://example.com/1",
//					MagnetLink: "magnet:?xt=urn:btih:hash1",
//					Seeders:    100,
//					Size:       1500000000, // 1.5GB
//				},
//				{
//					Name:       "[SubsPlease] Sousou no Frieren - 01 (720p) [EFGH5678].mkv",
//					InfoHash:   "hash2",
//					Link:       "https://example.com/2",
//					MagnetLink: "magnet:?xt=urn:btih:hash2",
//					Seeders:    50,
//					Size:       800000000, // 800MB
//				},
//				{
//					Name:       "[Erai-raws] Sousou no Frieren - 01 [1080p][Multiple Subtitle].mkv",
//					InfoHash:   "hash3",
//					Link:       "https://example.com/3",
//					MagnetLink: "magnet:?xt=urn:btih:hash3",
//					Seeders:    80,
//					Size:       1400000000, // 1.4GB
//				},
//				{
//					Name:       "[SubsPlease] Sousou no Frieren - 02 (1080p) [IJKL9012].mkv",
//					InfoHash:   "hash4",
//					Link:       "https://example.com/4",
//					MagnetLink: "magnet:?xt=urn:btih:hash4",
//					Seeders:    120,
//					Size:       1600000000, // 1.6GB
//				},
//				{
//					Name:       "[BadGroup] Sousou no Frieren - 01 (1080p) [BAD].mkv",
//					InfoHash:   "hash5",
//					Link:       "https://example.com/5",
//					MagnetLink: "magnet:?xt=urn:btih:hash5",
//					Seeders:    10,
//					Size:       900000000, // 900MB
//				},
//			},
//			rule: &anime.AutoDownloaderRule{
//				DbID:                10,
//				Enabled:             true,
//				MediaId:             154587,
//				Destination:         "/downloads",
//				ProfileID:           lo.ToPtr(uint(2)),
//				ReleaseGroups:       []string{"SubsPlease", "Erai-raws"},
//				Resolutions:         []string{},
//				EpisodeType:         anime.AutoDownloaderRuleEpisodeRecent,
//				ComparisonTitle:     "Sousou no Frieren",
//				TitleComparisonType: anime.AutoDownloaderRuleTitleComparisonLikely,
//				MinSeeders:          0,
//				Providers:           []string{"fake"},
//			},
//			profiles: []*anime.AutoDownloaderProfile{
//				{
//					DbID:   1,
//					Name:   "Global Profile",
//					Global: true,
//					Conditions: []anime.AutoDownloaderCondition{
//						{ID: "1", Term: "1080p", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 20},
//						{ID: "2", Term: "720p", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 10},
//						{ID: "3", Term: "SubsPlease", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 15},
//					},
//					Resolutions:  []string{"1080p", "720p"},
//					MinSeeders:   20,
//					MinimumScore: 0,
//				},
//				{
//					DbID:   2,
//					Name:   "Specific Profile",
//					Global: false,
//					Conditions: []anime.AutoDownloaderCondition{
//						{ID: "4", Term: "BadGroup", Action: anime.AutoDownloaderProfileRuleFormatActionBlock},
//					},
//					Resolutions:  []string{"1080p"},
//					MinSeeders:   15,
//					MinimumScore: 0,
//				},
//			},
//			expectedResults: 2,
//			expectedEpisodes: []struct {
//				episode     int
//				hash        string
//				score       int
//				shouldMatch func(result *SimulationResult) bool
//			}{
//				{
//					episode: 1,
//					hash:    "hash1",
//					score:   35,
//					shouldMatch: func(result *SimulationResult) bool {
//						return assert.Contains(t, result.TorrentName, "SubsPlease") &&
//							assert.Contains(t, result.TorrentName, "1080p") &&
//							assert.Contains(t, result.TorrentName, "- 01")
//					},
//				},
//				{
//					episode: 2,
//					hash:    "hash4",
//					score:   35,
//					shouldMatch: func(result *SimulationResult) bool {
//						return assert.Contains(t, result.TorrentName, "SubsPlease") &&
//							assert.Contains(t, result.TorrentName, "1080p") &&
//							assert.Contains(t, result.TorrentName, "- 02")
//					},
//				},
//			},
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			t.Logf("=== Test: %s ===", tt.name)
//
//			// Log torrents
//			t.Logf("Prepared %d fake torrents:", len(tt.torrents))
//			for i, torrent := range tt.torrents {
//				t.Logf("  [%d] %s (Seeders: %d, Size: %.2f GB)", i+1, torrent.Name, torrent.Seeders, float64(torrent.Size)/1e9)
//			}
//
//			// Create fake provider
//			fake := &Fake{
//				GetLatestResults: tt.torrents,
//				SearchResults:    tt.torrents,
//			}
//
//			// Create AutoDownloader instance
//			ad := fake.New(t)
//			t.Logf("AutoDownloader instance created")
//
//			// Set anime collection
//			ad.SetAnimeCollection(animeCollection)
//			t.Logf("Loaded anime collection with %d list entries", len(animeCollection.GetMediaListCollection().GetLists()))
//
//			// Get anime from collection
//			listEntry, found := animeCollection.GetListEntryFromAnimeId(tt.mediaId)
//			require.True(t, found, "Anime with ID %d should be in the boilerplate collection", tt.mediaId)
//			t.Logf("Found anime: %s (ID: %d)", *listEntry.GetMedia().GetTitle().GetUserPreferred(), listEntry.GetMedia().GetID())
//
//			// Log profiles
//			for _, profile := range tt.profiles {
//				t.Logf("Profile: %s (Global: %v)", profile.Name, profile.Global)
//				if len(profile.Conditions) > 0 {
//					t.Logf("  - Conditions: %d", len(profile.Conditions))
//					for _, cond := range profile.Conditions {
//						if cond.Action == anime.AutoDownloaderProfileRuleFormatActionScore {
//							t.Logf("    * '%s' -> Score: %d", cond.Term, cond.Score)
//						} else {
//							t.Logf("    * '%s' -> Action: %s", cond.Term, cond.Action)
//						}
//					}
//				}
//			}
//
//			// Log rule
//			t.Logf("Rule for media %d:", tt.rule.MediaId)
//			t.Logf("  - Release Groups: %v", tt.rule.ReleaseGroups)
//			t.Logf("  - Episode Type: %s", tt.rule.EpisodeType)
//			t.Logf("  - Title Comparison: %s", tt.rule.TitleComparisonType)
//
//			// Insert profiles and rule into database
//			for _, profile := range tt.profiles {
//				err = db_bridge.InsertAutoDownloaderProfile(ad.database, profile)
//				require.NoError(t, err)
//			}
//			err = db_bridge.InsertAutoDownloaderRule(ad.database, tt.rule)
//			require.NoError(t, err)
//			t.Logf("Inserted rule and %d profile(s) into database", len(tt.profiles))
//
//			// Set user progress
//			listEntry.Progress = lo.ToPtr(tt.userProgress)
//			t.Logf("Set user progress to episode %d", tt.userProgress)
//
//			// Clear any previous simulation results
//			ad.ClearSimulationResults()
//
//			// Run the auto downloader synchronously in simulation mode
//			t.Logf("\n--- Running AutoDownloader ---")
//			ad.RunCheck(true)
//
//			// Get simulation results
//			results := ad.GetSimulationResults()
//			t.Logf("\n--- Simulation Results ---")
//			t.Logf("Total torrents selected: %d", len(results))
//			for i, result := range results {
//				t.Logf("[%d] Episode %d: %s", i+1, result.Episode, result.TorrentName)
//				t.Logf("    Hash: %s, Score: %d, MediaID: %d", result.Hash, result.Score, result.MediaID)
//			}
//
//			// Verify total results
//			assert.Equal(t, tt.expectedResults, len(results), "Should have selected %d torrent(s)", tt.expectedResults)
//
//			// Verify each expected episode
//			for _, expected := range tt.expectedEpisodes {
//				t.Logf("\n--- Verifying Episode %d ---", expected.episode)
//
//				// Find results for this episode
//				episodeResults := lo.Filter(results, func(item *SimulationResult, _ int) bool {
//					return item.Episode == expected.episode
//				})
//
//				require.Equal(t, 1, len(episodeResults), "Should have exactly 1 torrent for episode %d", expected.episode)
//
//				result := episodeResults[0]
//				t.Logf("Selected: %s", result.TorrentName)
//				t.Logf("Hash: %s (expected: %s)", result.Hash, expected.hash)
//				t.Logf("Score: %d (expected: %d)", result.Score, expected.score)
//
//				assert.Equal(t, expected.hash, result.Hash, "Episode %d should have hash %s", expected.episode, expected.hash)
//				assert.Equal(t, expected.score, result.Score, "Episode %d should have score %d", expected.episode, expected.score)
//
//				if expected.shouldMatch != nil {
//					assert.True(t, expected.shouldMatch(result), "Episode %d should match custom criteria", expected.episode)
//				}
//			}
//
//			t.Logf("\n✅ Test passed: %s", tt.name)
//		})
//	}
//}
