package autodownloader

import (
	"context"
	"seanime/internal/api/anilist"
	"seanime/internal/database/db_bridge"
	"seanime/internal/database/models"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/library/anime"
	"seanime/internal/test_utils"
	"seanime/internal/torrent_clients/torrent_client"
	"seanime/internal/util"
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsConstraintsMatch(t *testing.T) {
	ad := &AutoDownloader{}

	tests := []struct {
		name     string
		torrent  *NormalizedTorrent
		rule     *anime.AutoDownloaderRule
		expected bool
	}{
		{
			name:     "Min seeders pass",
			torrent:  &NormalizedTorrent{AnimeTorrent: &hibiketorrent.AnimeTorrent{Seeders: 10}},
			rule:     &anime.AutoDownloaderRule{MinSeeders: 5},
			expected: true,
		},
		{
			name:     "Min seeders pass (no data)",
			torrent:  &NormalizedTorrent{AnimeTorrent: &hibiketorrent.AnimeTorrent{Seeders: -1}},
			rule:     &anime.AutoDownloaderRule{MinSeeders: 5},
			expected: true,
		},
		{
			name:     "Min seeders fail",
			torrent:  &NormalizedTorrent{AnimeTorrent: &hibiketorrent.AnimeTorrent{Seeders: 2}},
			rule:     &anime.AutoDownloaderRule{MinSeeders: 5},
			expected: false,
		},
		{
			name:     "Min size pass",
			torrent:  &NormalizedTorrent{AnimeTorrent: &hibiketorrent.AnimeTorrent{Size: 2048}}, // 2KB
			rule:     &anime.AutoDownloaderRule{MinSize: "1KB"},
			expected: true,
		},
		{
			name:     "Min size pass (no data)",
			torrent:  &NormalizedTorrent{AnimeTorrent: &hibiketorrent.AnimeTorrent{Size: 0}},
			rule:     &anime.AutoDownloaderRule{MinSize: "1KB"},
			expected: true,
		},
		{
			name:     "Min size fail",
			torrent:  &NormalizedTorrent{AnimeTorrent: &hibiketorrent.AnimeTorrent{Size: 512}}, // 0.5KB
			rule:     &anime.AutoDownloaderRule{MinSize: "1KB"},
			expected: false,
		},
		{
			name:     "Max size pass",
			torrent:  &NormalizedTorrent{AnimeTorrent: &hibiketorrent.AnimeTorrent{Size: 1024}}, // 1KB
			rule:     &anime.AutoDownloaderRule{MaxSize: "2KB"},
			expected: true,
		},
		{
			name:     "Max size fail",
			torrent:  &NormalizedTorrent{AnimeTorrent: &hibiketorrent.AnimeTorrent{Size: 3072}}, // 3KB
			rule:     &anime.AutoDownloaderRule{MaxSize: "2KB"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ad.isConstraintsMatch(tt.torrent, tt.rule); got != tt.expected {
				t.Errorf("isConstraintsMatch() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsExcludedTermsMatch(t *testing.T) {
	ad := &AutoDownloader{}

	tests := []struct {
		name     string
		torrent  string
		rule     *anime.AutoDownloaderRule
		expected bool
	}{
		{
			name:     "No excluded terms",
			torrent:  "One Piece - 1000",
			rule:     &anime.AutoDownloaderRule{ExcludeTerms: []string{}},
			expected: true,
		},
		{
			name:     "Contains excluded term",
			torrent:  "One Piece - 1000 - French",
			rule:     &anime.AutoDownloaderRule{ExcludeTerms: []string{"French"}},
			expected: false,
		},
		{
			name:     "Does not contain excluded term",
			torrent:  "One Piece - 1000 - English",
			rule:     &anime.AutoDownloaderRule{ExcludeTerms: []string{"French"}},
			expected: true,
		},
		{
			name:     "Case insensitive check",
			torrent:  "One Piece - 1000 - french",
			rule:     &anime.AutoDownloaderRule{ExcludeTerms: []string{"French"}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ad.isExcludedTermsMatch(tt.torrent, tt.rule); got != tt.expected {
				t.Errorf("isExcludedTermsMatch() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsResolutionMatch(t *testing.T) {
	ad := &AutoDownloader{}

	tests := []struct {
		name        string
		quality     string
		resolutions []string
		expected    bool
	}{
		{
			name:        "Match Exact",
			quality:     "1080p",
			resolutions: []string{"1080p"},
			expected:    true,
		},
		{
			name:        "Match List",
			quality:     "720p",
			resolutions: []string{"1080p", "720p"},
			expected:    true,
		},
		{
			name:        "No Match",
			quality:     "480p",
			resolutions: []string{"1080p", "720p"},
			expected:    false,
		},
		{
			name:        "Empty Resolutions (Match All)",
			quality:     "480p",
			resolutions: []string{},
			expected:    true,
		},
		{
			name:        "Mixed Case",
			quality:     "1080P",
			resolutions: []string{"1080p"},
			expected:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ad.isResolutionMatch(tt.quality, tt.resolutions)
			assert.Equal(t, tt.expected, result)
		})
	}
}

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
			rule: &anime.AutoDownloaderRule{ProfileID: new(uint(1))},
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
			rule:     &anime.AutoDownloaderRule{ProfileID: new(uint(2))},
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
				AnimeTorrent: &hibiketorrent.AnimeTorrent{InfoHash: "hash2"},
			},
			expected: true,
		},
		{
			name: "torrent does not exist",
			torrent: &NormalizedTorrent{
				AnimeTorrent: &hibiketorrent.AnimeTorrent{InfoHash: "hash999"},
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

	// Create fake local files
	localFiles := []*anime.LocalFile{
		{MediaId: 1, Metadata: &anime.LocalFileMetadata{Episode: 5, Type: anime.LocalFileTypeMain}},
		{MediaId: 1, Metadata: &anime.LocalFileMetadata{Episode: 6, Type: anime.LocalFileTypeMain}},
		{MediaId: 2, Metadata: &anime.LocalFileMetadata{Episode: 1, Type: anime.LocalFileTypeMain}},
		{MediaId: 21, Metadata: &anime.LocalFileMetadata{Episode: 1, Type: anime.LocalFileTypeMain}},
	}
	lfWrapper := anime.NewLocalFileWrapper(localFiles)

	// fake queued items
	queuedItems := []*models.AutoDownloaderItem{
		{RuleID: 1, MediaID: 1, Episode: 7},
		{RuleID: 1, MediaID: 1, Episode: 8},
		{RuleID: 2, MediaID: 1, Episode: 8},
		{RuleID: 1, MediaID: 10, Episode: 42, IsDelayed: true},
		{RuleID: 3, MediaID: 20, Episode: 13},
		{RuleID: 5, MediaID: 20, Episode: 1},
	}

	tests := []struct {
		name     string
		episode  int
		offset   int
		ruleID   uint
		mediaId  int
		expected bool
	}{
		{
			name:     "episode in local files",
			episode:  5,
			ruleID:   1,
			mediaId:  1,
			expected: true,
		},
		{
			name:     "episode in queue for same rule",
			episode:  7,
			ruleID:   1,
			mediaId:  1,
			expected: true,
		},
		{
			name:     "episode in queue but different rule",
			episode:  8,
			ruleID:   3,
			mediaId:  1,
			expected: false,
		},
		{
			name:     "episode not handled",
			episode:  10,
			ruleID:   1,
			mediaId:  1,
			expected: false,
		},
		{
			name:     "different media id",
			episode:  5,
			ruleID:   1,
			mediaId:  99,
			expected: false,
		},
		{
			name:     "same episode but delayed",
			episode:  42,
			ruleID:   1,
			mediaId:  10,
			expected: false,
		},
		{
			name:     "absolute offset handled",
			episode:  13, // -> 1
			offset:   12,
			ruleID:   1,
			mediaId:  2,
			expected: true,
		},
		{
			name:     "absolute offset handled - found in queue",
			episode:  13, // the one found in the queue has the same episode number
			offset:   12,
			ruleID:   3,
			mediaId:  20,
			expected: true,
		},
		{
			name:     "absolute offset handled - not found in queue",
			episode:  14, // -> 2
			offset:   12,
			ruleID:   3,
			mediaId:  20,
			expected: false,
		},
		{
			name:     "absolute offset handled - found in local files",
			episode:  13, // the local file has episode 1
			offset:   12,
			ruleID:   4,
			mediaId:  21,
			expected: true,
		},
		{
			name:     "absolute offset handled - non-offset episode in queue",
			episode:  13, // the queue has episode 1
			offset:   12,
			ruleID:   5,
			mediaId:  20,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ad.isEpisodeAlreadyHandled(tt.episode, tt.offset, tt.ruleID, tt.mediaId, lfWrapper, queuedItems)
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
				AnimeTorrent: &hibiketorrent.AnimeTorrent{Name: "[SubsPlease] One Piece - 1000 (1080p) [ABCD].mkv"},
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
				{Score: 10, Torrent: &NormalizedTorrent{AnimeTorrent: &hibiketorrent.AnimeTorrent{Name: "a1", Seeders: 50}}},
				{Score: 25, Torrent: &NormalizedTorrent{AnimeTorrent: &hibiketorrent.AnimeTorrent{Name: "a2", Seeders: 30}}},
				{Score: 15, Torrent: &NormalizedTorrent{AnimeTorrent: &hibiketorrent.AnimeTorrent{Name: "a3", Seeders: 100}}},
			},
			expectedTorrent: "a2",
		},
		{
			name: "same score, more seeders wins",
			candidates: []*Candidate{
				{Score: 20, Torrent: &NormalizedTorrent{AnimeTorrent: &hibiketorrent.AnimeTorrent{Name: "a1", Seeders: 50}}},
				{Score: 20, Torrent: &NormalizedTorrent{AnimeTorrent: &hibiketorrent.AnimeTorrent{Name: "a2", Seeders: 150}}},
				{Score: 20, Torrent: &NormalizedTorrent{AnimeTorrent: &hibiketorrent.AnimeTorrent{Name: "a3", Seeders: 100}}},
			},
			expectedTorrent: "a2",
		},
		{
			name: "single candidate",
			candidates: []*Candidate{
				{Score: 10, Torrent: &NormalizedTorrent{AnimeTorrent: &hibiketorrent.AnimeTorrent{Name: "a1", Seeders: 10}}},
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
			name: "preferred_fansub_match",
			torrent: &NormalizedTorrent{
				AnimeTorrent: &hibiketorrent.AnimeTorrent{Name: "[SubsPlease] Shangri-La Frontier S02 - 14 (1080p).mkv"},
			},
			profiles: []*anime.AutoDownloaderProfile{
				{
					Conditions: []anime.AutoDownloaderCondition{
						{Term: "SubsPlease, Erai-raws", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 50},
						{Term: "1080p", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 20},
					},
					MinimumScore: 60,
				},
			},
			expectedScore: 70,
			shouldSucceed: true,
		},
		{
			name: "reject_batch_uploads",
			torrent: &NormalizedTorrent{
				AnimeTorrent: &hibiketorrent.AnimeTorrent{Name: "[ASW] Solo Leveling - S02E01 [1080p HEVC x265 10bit] (Batch)"},
			},
			profiles: []*anime.AutoDownloaderProfile{
				{
					Conditions: []anime.AutoDownloaderCondition{
						{Term: "1080p", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 30},
						{Term: "Batch, Dual-Audio, Dub", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: -100},
					},
					MinimumScore: 10,
				},
			},
			expectedScore: -70,
			shouldSucceed: false,
		},
		{
			name: "regex_media_source_detection",
			torrent: &NormalizedTorrent{
				AnimeTorrent: &hibiketorrent.AnimeTorrent{Name: "[EMBER] Frieren - 01 [BDRip] [1080p] [HEVC]"},
			},
			profiles: []*anime.AutoDownloaderProfile{
				{
					Conditions: []anime.AutoDownloaderCondition{
						{Term: `(BD|Blu-?ray|BDRip)`, IsRegex: true, Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 100},
						{Term: "HEVC, x265", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 20},
					},
					MinimumScore: 110,
				},
			},
			expectedScore: 120,
			shouldSucceed: true,
		},
		{
			name: "cumulative_global_and_group_profiles",
			torrent: &NormalizedTorrent{
				AnimeTorrent: &hibiketorrent.AnimeTorrent{Name: "[Nippon-Yasan] Monogatari Series - Off & Monster Season - 01 (1080p).mkv"},
			},
			profiles: []*anime.AutoDownloaderProfile{
				{
					Name: "Global Format",
					Conditions: []anime.AutoDownloaderCondition{
						{Term: "1080p", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 20},
					},
					MinimumScore: 0,
					Global:       true,
				},
				{
					Name: "Encoder Preference",
					Conditions: []anime.AutoDownloaderCondition{
						{Term: "Nippon-Yasan, MTBB", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 40},
					},
					MinimumScore: 50,
				},
			},
			expectedScore: 60,
			shouldSucceed: true,
		},
		{
			name: "version_revision_handling",
			torrent: &NormalizedTorrent{
				AnimeTorrent: &hibiketorrent.AnimeTorrent{Name: "[SubsPlease] Dandadan - 01v2 (1080p).mkv"},
			},
			profiles: []*anime.AutoDownloaderProfile{
				{
					Conditions: []anime.AutoDownloaderCondition{
						{Term: `v[2-9]`, IsRegex: true, Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 15},
						{Term: "1080p", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 10},
					},
					MinimumScore: 20,
				},
			},
			expectedScore: 25,
			shouldSucceed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, minScore := ad.calculateCandidateScore(tt.torrent, tt.profiles)
			assert.Equal(t, tt.expectedScore, score, "Expected %d, got %d", tt.expectedScore, score)
			assert.Equal(t, tt.expectedScore >= minScore, tt.shouldSucceed, "Expected %v, got %v", tt.shouldSucceed, tt.expectedScore >= minScore)
		})
	}
}

func TestTorrentRanking(t *testing.T) {
	ad := &AutoDownloader{}

	tests := []struct {
		name           string
		scenarioDesc   string
		candidates     []*NormalizedTorrent
		profiles       []*anime.AutoDownloaderProfile
		expectedWinner string
	}{
		{
			name:         "high_fidelity_archive",
			scenarioDesc: "Prefer BDRip with lossless audio over high-seed Web-DL",
			candidates: []*NormalizedTorrent{
				{AnimeTorrent: &hibiketorrent.AnimeTorrent{Name: "[SubsPlease] Sousou no Frieren - 01 (1080p) [Web-DL].mkv", Seeders: 3500, Size: 1400000000}},
				{AnimeTorrent: &hibiketorrent.AnimeTorrent{Name: "[Erai-raws] Sousou no Frieren - 01 [1080p].mkv", Seeders: 1200, Size: 1600000000}},
				{AnimeTorrent: &hibiketorrent.AnimeTorrent{Name: "[CoalGirls] Sousou no Frieren - 01 [BDRip 1920x1080 HEVC FLAC].mkv", Seeders: 85, Size: 5200000000}},
			},
			profiles: []*anime.AutoDownloaderProfile{
				{
					Name: "Archival Quality",
					Conditions: []anime.AutoDownloaderCondition{
						{Term: "FLAC", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 100},
						{Term: "BDRip", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 50},
						{Term: "WebDL, WEBRIP, WEB RIP, Web-DL", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: -20},
					},
					MinimumScore: 10,
				},
			},
			expectedWinner: "[CoalGirls] Sousou no Frieren - 01 [BDRip 1920x1080 HEVC FLAC].mkv",
		},
		{
			name:         "codec_compatibility",
			scenarioDesc: "Filter out HEVC for legacy playback hardware",
			candidates: []*NormalizedTorrent{
				{AnimeTorrent: &hibiketorrent.AnimeTorrent{Name: "[Judas] Kaiju No. 8 - 01 [1080p][HEVC x265 10bit].mkv", Seeders: 2200}},
				{AnimeTorrent: &hibiketorrent.AnimeTorrent{Name: "[SubsPlease] Kaiju No. 8 - 01 (1080p).mkv", Seeders: 1800}},
			},
			profiles: []*anime.AutoDownloaderProfile{
				{
					Name: "H264 Only",
					Conditions: []anime.AutoDownloaderCondition{
						{Term: "1080p", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 10},
						{Term: "HEVC, x265, 10bit, 10-bit", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: -500},
					},
					MinimumScore: 0,
				},
			},
			expectedWinner: "[SubsPlease] Kaiju No. 8 - 01 (1080p).mkv",
		},
		{
			name:         "multi_audio_preference",
			scenarioDesc: "Prioritize Dual Audio releases for local library",
			candidates: []*NormalizedTorrent{
				{AnimeTorrent: &hibiketorrent.AnimeTorrent{Name: "[SubsPlease] Dungeon Meshi - 01 (1080p).mkv", Seeders: 2800}},
				{AnimeTorrent: &hibiketorrent.AnimeTorrent{Name: "[EMBER] Dungeon Meshi - 01 (1080p) [Dual Audio].mkv", Seeders: 210}},
			},
			profiles: []*anime.AutoDownloaderProfile{
				{
					Name: "English Dub",
					Conditions: []anime.AutoDownloaderCondition{
						{Term: "Dual Audio", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 150},
						{Term: "1080p", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 10},
					},
					MinimumScore: 50,
				},
			},
			expectedWinner: "[EMBER] Dungeon Meshi - 01 (1080p) [Dual Audio].mkv",
		},
		{
			name:         "release_revision_v2",
			scenarioDesc: "Ensure v2/corrected releases are preferred over initial uploads",
			candidates: []*NormalizedTorrent{
				{AnimeTorrent: &hibiketorrent.AnimeTorrent{Name: "[Group] Metallic Rouge - 01 [1080p].mkv", Seeders: 900}},
				{AnimeTorrent: &hibiketorrent.AnimeTorrent{Name: "[Group] Metallic Rouge - 01v2 [1080p].mkv", Seeders: 340}},
			},
			profiles: []*anime.AutoDownloaderProfile{
				{
					Name: "Latest Revision",
					Conditions: []anime.AutoDownloaderCondition{
						{Term: "v2", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 100},
					},
					MinimumScore: 0,
				},
			},
			expectedWinner: "[Group] Metallic Rouge - 01v2 [1080p].mkv",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var candidates []*Candidate

			for _, torrent := range tt.candidates {
				score, reqMinScore := ad.calculateCandidateScore(torrent, tt.profiles)
				if score >= reqMinScore {
					candidates = append(candidates, &Candidate{
						Torrent: torrent,
						Score:   score,
					})
				}
			}

			if tt.expectedWinner == "" {
				assert.Empty(t, candidates)
			} else {
				if assert.NotEmpty(t, candidates) {
					winner := ad.selectBestCandidate(candidates)
					assert.Equal(t, tt.expectedWinner, winner.Torrent.Name)
				}
			}
		})
	}
}

func TestIsProfileValidChecks(t *testing.T) {
	ad := &AutoDownloader{}

	tests := []struct {
		name    string
		torrent *NormalizedTorrent
		profile *anime.AutoDownloaderProfile
		isValid bool
	}{
		{
			name: "valid_standard_release",
			torrent: &NormalizedTorrent{
				AnimeTorrent: &hibiketorrent.AnimeTorrent{
					Name:    "[SubsPlease] One Piece - 1100 (1080p).mkv",
					Seeders: 1500,
					Size:    1400000000,
				},
			},
			profile: &anime.AutoDownloaderProfile{
				MinSeeders: 100,
				MinSize:    "500MB",
				Conditions: []anime.AutoDownloaderCondition{
					{Term: "1080p", Action: anime.AutoDownloaderProfileRuleFormatActionRequire},
				},
			},
			isValid: true,
		},
		{
			name: "fail_low_seeder_count",
			torrent: &NormalizedTorrent{
				AnimeTorrent: &hibiketorrent.AnimeTorrent{
					Name:    "[OldGroup] Classic Movie (1080p).mkv",
					Seeders: 2,
					Size:    4000000000,
				},
			},
			profile: &anime.AutoDownloaderProfile{
				MinSeeders: 10,
			},
			isValid: false,
		},
		{
			name: "fail_blocked_release_group",
			torrent: &NormalizedTorrent{
				AnimeTorrent: &hibiketorrent.AnimeTorrent{
					Name:    "[SubsPlease] Bleach - 01 (1080p).mkv",
					Seeders: 500,
				},
			},
			profile: &anime.AutoDownloaderProfile{
				Conditions: []anime.AutoDownloaderCondition{
					{Term: "SubsPlease", Action: anime.AutoDownloaderProfileRuleFormatActionBlock},
				},
			},
			isValid: false,
		},
		{
			name: "fail_missing_required_codec",
			torrent: &NormalizedTorrent{
				AnimeTorrent: &hibiketorrent.AnimeTorrent{
					Name:    "[Erai-raws] Danmachi S5 - 01 (1080p).mkv",
					Seeders: 800,
				},
			},
			profile: &anime.AutoDownloaderProfile{
				Conditions: []anime.AutoDownloaderCondition{
					{Term: "AV1", Action: anime.AutoDownloaderProfileRuleFormatActionRequire},
				},
			},
			isValid: false,
		},
		{
			name: "fail_undersized_release",
			torrent: &NormalizedTorrent{
				AnimeTorrent: &hibiketorrent.AnimeTorrent{
					Name: "[Micro] Chainsaw Man - 01 (1080p).mkv",
					Size: 150000000,
				},
			},
			profile: &anime.AutoDownloaderProfile{
				MinSize: "500MB",
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ad.isProfileValidChecks(tt.torrent, tt.profile)
			assert.Equal(t, tt.isValid, result)
		})
	}
}

func TestIntegration(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	anilistClient := anilist.TestGetMockAnilistClient()
	animeCollection, err := anilistClient.AnimeCollection(context.Background(), nil)
	require.NoError(t, err)

	tests := []struct {
		name             string
		mediaId          int
		userProgress     int
		torrents         []*hibiketorrent.AnimeTorrent
		rule             *anime.AutoDownloaderRule
		profiles         []*anime.AutoDownloaderProfile
		expectedResults  int
		expectedEpisodes []struct {
			episode int
			hash    string
			score   int
		}
	}{
		{
			name:         "seasonal_catchup_with_quality_filtering",
			mediaId:      154587, // Sousou no Frieren
			userProgress: 0,
			torrents: []*hibiketorrent.AnimeTorrent{
				{
					Name:     "[SubsPlease] Sousou no Frieren - 01 (1080p) [V0].mkv",
					InfoHash: "hash_sp_01",
					Seeders:  1200,
					Size:     1400000000,
				},
				{
					Name:     "[SubsPlease] Sousou no Frieren - 01 (720p).mkv",
					InfoHash: "hash_sp_01_720",
					Seeders:  450,
					Size:     800000000,
				},
				{
					Name:     "[Erai-raws] Sousou no Frieren - 01 [1080p].mkv",
					InfoHash: "hash_erai_01",
					Seeders:  900,
					Size:     1350000000,
				},
				{
					Name:     "[SubsPlease] Sousou no Frieren - 02 (1080p).mkv",
					InfoHash: "hash_sp_02",
					Seeders:  1100,
					Size:     1450000000,
				},
				{
					Name:     "[LowQuality] Sousou no Frieren - 01 (1080p).mkv",
					InfoHash: "hash_bad_01",
					Seeders:  10,
					Size:     700000000,
				},
			},
			rule: &anime.AutoDownloaderRule{
				DbID:                10,
				Enabled:             true,
				MediaId:             154587,
				Destination:         "/media/anime/frieren",
				ProfileID:           new(uint(2)),
				ReleaseGroups:       []string{"SubsPlease", "Erai-raws"},
				EpisodeType:         anime.AutoDownloaderRuleEpisodeRecent,
				ComparisonTitle:     "Sousou no Frieren",
				TitleComparisonType: anime.AutoDownloaderRuleTitleComparisonLikely,
				Providers:           []string{"fake", "inexistant"}, // "inexistant" should be filtered out
			},
			profiles: []*anime.AutoDownloaderProfile{
				{
					DbID:   1,
					Name:   "Global Defaults",
					Global: true,
					Conditions: []anime.AutoDownloaderCondition{
						{Term: "1080p", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 25},
						{Term: "720p", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 10},
						{Term: "SubsPlease", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 10},
					},
					Resolutions:  []string{"1080p", "720p"},
					MinSeeders:   20,
					MinimumScore: 0,
				},
				{
					DbID:   2,
					Name:   "Strict Filtering",
					Global: false,
					Conditions: []anime.AutoDownloaderCondition{
						{Term: "LowQuality", Action: anime.AutoDownloaderProfileRuleFormatActionBlock},
					},
					Resolutions:  []string{"1080p"},
					MinSeeders:   50,
					MinimumScore: 10,
				},
			},
			expectedResults: 2,
			expectedEpisodes: []struct {
				episode int
				hash    string
				score   int
			}{
				{episode: 1, hash: "hash_sp_01", score: 35},
				{episode: 2, hash: "hash_sp_02", score: 35},
			},
		},
		{
			name:         "do_not_download_watched_episodes",
			mediaId:      154587, // Sousou no Frieren
			userProgress: 1,
			torrents: []*hibiketorrent.AnimeTorrent{
				{
					Name:     "[SubsPlease] Sousou no Frieren - 01 (1080p) [V0].mkv",
					InfoHash: "hash_sp_01",
					Seeders:  1200,
					Size:     1400000000,
				},
				{
					Name:     "[Erai-raws] Sousou no Frieren - 01 [1080p].mkv",
					InfoHash: "hash_erai_01",
					Seeders:  900,
					Size:     1350000000,
				},
				{
					Name:     "[SubsPlease] Sousou no Frieren - 02 (1080p).mkv",
					InfoHash: "hash_sp_02",
					Seeders:  1100,
					Size:     1450000000,
				},
				{
					Name:     "[LowQuality] Sousou no Frieren - 01 (1080p).mkv",
					InfoHash: "hash_bad_01",
					Seeders:  10,
					Size:     700000000,
				},
			},
			rule: &anime.AutoDownloaderRule{
				DbID:                10,
				Enabled:             true,
				MediaId:             154587,
				Destination:         "/media/anime/frieren",
				ProfileID:           new(uint(2)),
				ReleaseGroups:       []string{"SubsPlease", "Erai-raws"},
				EpisodeType:         anime.AutoDownloaderRuleEpisodeRecent,
				ComparisonTitle:     "Sousou no Frieren",
				TitleComparisonType: anime.AutoDownloaderRuleTitleComparisonLikely,
				Providers:           []string{"fake", "inexistant"}, // "inexistant" should be filtered out
			},
			profiles: []*anime.AutoDownloaderProfile{
				{
					DbID:   1,
					Name:   "Global Defaults",
					Global: true,
					Conditions: []anime.AutoDownloaderCondition{
						{Term: "1080p", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 25},
						{Term: "720p", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 10},
						{Term: "SubsPlease", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 10},
					},
					Resolutions:  []string{"1080p", "720p"},
					MinSeeders:   20,
					MinimumScore: 0,
				},
				{
					DbID:   2,
					Name:   "Strict Filtering",
					Global: false,
					Conditions: []anime.AutoDownloaderCondition{
						{Term: "LowQuality", Action: anime.AutoDownloaderProfileRuleFormatActionBlock},
					},
					Resolutions:  []string{"1080p"},
					MinSeeders:   50,
					MinimumScore: 10,
				},
			},
			expectedResults: 1,
			expectedEpisodes: []struct {
				episode int
				hash    string
				score   int
			}{
				{episode: 2, hash: "hash_sp_02", score: 35},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new fake
			fake := &Fake{
				GetLatestResults: tt.torrents,
				SearchResults:    tt.torrents,
			}
			ad := fake.New(t)
			ad.SetAnimeCollection(animeCollection)

			// Add local files to the database
			_, err = fake.Database.InsertLocalFiles(&models.LocalFiles{Value: []byte("[]")})
			require.NoError(t, err)

			// Set user progress
			listEntry, found := animeCollection.GetListEntryFromAnimeId(tt.mediaId)
			require.True(t, found)
			listEntry.Progress = new(tt.userProgress)

			// Add profiles and rule to the database
			for _, profile := range tt.profiles {
				err = db_bridge.InsertAutoDownloaderProfile(ad.database, profile)
				require.NoError(t, err)
			}
			err = db_bridge.InsertAutoDownloaderRule(ad.database, tt.rule)
			require.NoError(t, err)

			ad.ClearSimulationResults()
			// Run simulation
			ad.RunCheck(t.Context(), true)

			results := ad.GetSimulationResults()
			assert.Equal(t, tt.expectedResults, len(results))

			// Print the results
			t.Logf("Results: %d\n", len(results))
			for _, result := range results {
				t.Logf("TorrentName: %s\n", result.TorrentName)
				t.Logf("\tEpisode %d: Hash=%s, Score=%d\n", result.Episode, result.Hash, result.Score)
			}

			for _, expected := range tt.expectedEpisodes {
				episodeResults := lo.Filter(results, func(item *SimulationResult, _ int) bool {
					return item.Episode == expected.episode
				})

				if assert.Len(t, episodeResults, 1) {
					res := episodeResults[0]
					assert.Equal(t, expected.hash, res.Hash)
					assert.Equal(t, expected.score, res.Score)
				}
			}
		})
	}
}

func TestDelayIntegration(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	anilistClient := anilist.TestGetMockAnilistClient()
	animeCollection, err := anilistClient.AnimeCollection(context.Background(), nil)
	require.NoError(t, err)

	mediaId := 154587 // Sousou no Frieren

	tests := []struct {
		name                 string
		torrents             []*hibiketorrent.AnimeTorrent
		existingItems        []*models.AutoDownloaderItem
		profile              *anime.AutoDownloaderProfile
		expectedQueued       int
		expectedDownloaded   int
		checkDelayedItemFunc func(t *testing.T, items []*models.AutoDownloaderItem, simResults []*SimulationResult)
	}{
		{
			name: "Standard",
			torrents: []*hibiketorrent.AnimeTorrent{
				{Name: "[SubsPlease] Sousou no Frieren - 01 (1080p).mkv", InfoHash: "hash1", Seeders: 100},
			},
			profile: &anime.AutoDownloaderProfile{
				Conditions: []anime.AutoDownloaderCondition{{Term: "1080p", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 10}},
			},
			expectedQueued: 1,
			checkDelayedItemFunc: func(t *testing.T, items []*models.AutoDownloaderItem, _ []*SimulationResult) {
				require.Len(t, items, 1)
				assert.False(t, items[0].IsDelayed)  // MUST be true
				assert.False(t, items[0].Downloaded) // will always be false
				assert.Equal(t, "hash1", items[0].Hash)
			},
		},
		{
			name: "Queue item for delay",
			torrents: []*hibiketorrent.AnimeTorrent{
				{Name: "[SubsPlease] Sousou no Frieren - 01 (1080p).mkv", InfoHash: "hash1", Seeders: 100},
			},
			profile: &anime.AutoDownloaderProfile{
				Conditions:     []anime.AutoDownloaderCondition{{Term: "1080p", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 10}},
				DelayMinutes:   10,
				SkipDelayScore: 50,
			},
			expectedQueued: 1,
			checkDelayedItemFunc: func(t *testing.T, items []*models.AutoDownloaderItem, _ []*SimulationResult) {
				require.Len(t, items, 1)
				assert.True(t, items[0].IsDelayed)   // MUST be true
				assert.False(t, items[0].Downloaded) // will always be false
				assert.Equal(t, "hash1", items[0].Hash)
			},
		},
		{
			name: "Skip delay on high score",
			torrents: []*hibiketorrent.AnimeTorrent{
				{Name: "[SubsPlease] Sousou no Frieren - 01 (1080p).mkv", InfoHash: "hash1", Seeders: 100},
			},
			profile: &anime.AutoDownloaderProfile{
				Conditions:     []anime.AutoDownloaderCondition{{Term: "1080p", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 100}},
				DelayMinutes:   10,
				SkipDelayScore: 50,
			},
			expectedDownloaded: 1,
			checkDelayedItemFunc: func(t *testing.T, items []*models.AutoDownloaderItem, _ []*SimulationResult) {
				require.Len(t, items, 1)
				assert.False(t, items[0].IsDelayed)
				assert.False(t, items[0].Downloaded) // will always be false
			},
		},
		{
			name: "Download expired delayed item",
			torrents: []*hibiketorrent.AnimeTorrent{
				{Name: "[SubsPlease] Sousou no Frieren - 01 (1080p).mkv", InfoHash: "hash1", Seeders: 100},
			},
			existingItems: []*models.AutoDownloaderItem{
				{
					RuleID:      1,
					MediaID:     mediaId,
					Episode:     1,
					Hash:        "hash1",
					IsDelayed:   true,
					DelayUntil:  time.Now().Add(-1 * time.Minute), // Expired
					TorrentData: mustMarshalTorrent(&NormalizedTorrent{AnimeTorrent: &hibiketorrent.AnimeTorrent{Name: "[SubsPlease] Sousou no Frieren - 01 (1080p).mkv", InfoHash: "hash1"}}),
				},
			},
			profile: &anime.AutoDownloaderProfile{
				Conditions:   []anime.AutoDownloaderCondition{{Term: "1080p", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 10}},
				DelayMinutes: 10,
			},
			expectedDownloaded: 1,
			checkDelayedItemFunc: func(t *testing.T, items []*models.AutoDownloaderItem, _ []*SimulationResult) {
				require.Len(t, items, 1)
				util.Spew(items)
				assert.False(t, items[0].IsDelayed)
			},
		},
		{
			name: "Upgrade delayed item",
			torrents: []*hibiketorrent.AnimeTorrent{
				{Name: "[BetterGroup] Sousou no Frieren - 01 (1080p).mkv", InfoHash: "hash_better", Seeders: 100}, // Score 20
			},
			existingItems: []*models.AutoDownloaderItem{
				{
					RuleID:      1,
					MediaID:     mediaId,
					Episode:     1,
					Hash:        "hash_bad",
					Score:       10,
					IsDelayed:   true,
					DelayUntil:  time.Now().Add(5 * time.Minute), // Not expired
					TorrentData: mustMarshalTorrent(&NormalizedTorrent{AnimeTorrent: &hibiketorrent.AnimeTorrent{Name: "Name", InfoHash: "hash_bad"}}),
				},
			},
			profile: &anime.AutoDownloaderProfile{
				Conditions: []anime.AutoDownloaderCondition{
					{Term: "BetterGroup", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 20},
					{Term: "1080p", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 0},
				},
				DelayMinutes:   10,
				SkipDelayScore: 50,
			},
			expectedQueued: 1, // Updated but still queued
			checkDelayedItemFunc: func(t *testing.T, items []*models.AutoDownloaderItem, _ []*SimulationResult) {
				require.Len(t, items, 1)
				assert.True(t, items[0].IsDelayed)
				assert.Equal(t, "hash_better", items[0].Hash) // Updated hash
				assert.Equal(t, 20, items[0].Score)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &Fake{GetLatestResults: tt.torrents, SearchResults: tt.torrents}
			ad := fake.New(t)
			ad.SetAnimeCollection(animeCollection)

			// Setup DB
			_, _ = fake.Database.InsertLocalFiles(&models.LocalFiles{Value: []byte("[]")})
			if tt.existingItems != nil {
				for _, item := range tt.existingItems {
					_ = ad.database.InsertAutoDownloaderItem(item)
				}
			}

			// Setup Rule/Profile
			_ = db_bridge.InsertAutoDownloaderProfile(ad.database, tt.profile)
			_ = db_bridge.InsertAutoDownloaderRule(ad.database, &anime.AutoDownloaderRule{
				DbID: 1, Enabled: true, MediaId: mediaId, ProfileID: new(uint(1)),
				EpisodeType: anime.AutoDownloaderRuleEpisodeRecent, ComparisonTitle: "Sousou no Frieren", TitleComparisonType: anime.AutoDownloaderRuleTitleComparisonLikely,
			})

			ad.ClearSimulationResults()
			// devnote: We don't run in simulation mode so the items are added to the DB
			// they won't be downloaded since DownloadAutomatically=false
			ad.RunCheck(t.Context(), false)

			results := ad.GetSimulationResults()

			// Devnote: Assertions on internal state or simulation results won't reflect delayed downloads exactly since runCheck calls selectAndDownload then downloadDelayed
			// We check the DB state instead
			items, _ := ad.database.GetAutoDownloaderItems()
			if tt.checkDelayedItemFunc != nil {
				tt.checkDelayedItemFunc(t, items, results)
			}
		})
	}
}

func mustMarshalTorrent(t *NormalizedTorrent) []byte {
	b, _ := json.Marshal(t)
	return b
}
