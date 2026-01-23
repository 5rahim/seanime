package autodownloader

import (
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/library/anime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetReleaseGroupToResolutionsMap(t *testing.T) {
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
			Resolutions:   []string{}, // Should inherit from Global
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

	// Check SubsPlease
	// Rule 10 adds 1080p.
	// Rule 30 adds 4k (Global profile inheritance).
	assert.Contains(t, resMap, "SubsPlease")
	assert.ElementsMatch(t, []string{"1080p", "4k"}, resMap["SubsPlease"])

	// Check Erai-Raws
	// Rule 20 adds 720p (Specific Profile) AND 4k (Global Profile).
	// Because logical inheritance usually implies combining available resolutions if none specified.
	assert.Contains(t, resMap, "Erai-Raws")
	assert.ElementsMatch(t, []string{"720p", "4k"}, resMap["Erai-Raws"])
}

func TestCalculateTorrentScore(t *testing.T) {
	ad := &AutoDownloader{}

	torrent := &NormalizedTorrent{
		AnimeTorrent: hibiketorrent.AnimeTorrent{Name: "[SubsPlease] One Piece - 1000 (1080p) [ABCD].mkv"},
	}

	profile := &anime.AutoDownloaderProfile{
		Conditions: []anime.AutoDownloaderCondition{
			{Term: "1080p", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 10},
			{Term: "x265, HEVC", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 20},
			{Term: "SubsPlease", Action: anime.AutoDownloaderProfileRuleFormatActionScore, Score: 5},
		},
	}

	score := ad.calculateTorrentScore(torrent, profile)
	// Expect 10 (1080p) + 5 (SubsPlease) = 15
	assert.Equal(t, 15, score)
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

func TestTorrentFollowsRule_Resolutions(t *testing.T) {
}
