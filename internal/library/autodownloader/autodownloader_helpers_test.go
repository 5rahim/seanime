package autodownloader

import (
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/library/anime"
	"testing"

	"github.com/stretchr/testify/assert"
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
			torrent:  &NormalizedTorrent{AnimeTorrent: hibiketorrent.AnimeTorrent{Seeders: 10}},
			rule:     &anime.AutoDownloaderRule{MinSeeders: 5},
			expected: true,
		},
		{
			name:     "Min seeders pass (no data)",
			torrent:  &NormalizedTorrent{AnimeTorrent: hibiketorrent.AnimeTorrent{Seeders: -1}},
			rule:     &anime.AutoDownloaderRule{MinSeeders: 5},
			expected: true,
		},
		{
			name:     "Min seeders fail",
			torrent:  &NormalizedTorrent{AnimeTorrent: hibiketorrent.AnimeTorrent{Seeders: 2}},
			rule:     &anime.AutoDownloaderRule{MinSeeders: 5},
			expected: false,
		},
		{
			name:     "Min size pass",
			torrent:  &NormalizedTorrent{AnimeTorrent: hibiketorrent.AnimeTorrent{Size: 2048}}, // 2KB
			rule:     &anime.AutoDownloaderRule{MinSize: "1KB"},
			expected: true,
		},
		{
			name:     "Min size pass (no data)",
			torrent:  &NormalizedTorrent{AnimeTorrent: hibiketorrent.AnimeTorrent{Size: 0}},
			rule:     &anime.AutoDownloaderRule{MinSize: "1KB"},
			expected: true,
		},
		{
			name:     "Min size fail",
			torrent:  &NormalizedTorrent{AnimeTorrent: hibiketorrent.AnimeTorrent{Size: 512}}, // 0.5KB
			rule:     &anime.AutoDownloaderRule{MinSize: "1KB"},
			expected: false,
		},
		{
			name:     "Max size pass",
			torrent:  &NormalizedTorrent{AnimeTorrent: hibiketorrent.AnimeTorrent{Size: 1024}}, // 1KB
			rule:     &anime.AutoDownloaderRule{MaxSize: "2KB"},
			expected: true,
		},
		{
			name:     "Max size fail",
			torrent:  &NormalizedTorrent{AnimeTorrent: hibiketorrent.AnimeTorrent{Size: 3072}}, // 3KB
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

func TestStringToBytes(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		hasError bool
	}{
		{"1GB", 1073741824, false},
		{"1 GB", 1073741824, false},
		{"1.5 GB", 1610612736, false},
		{"1 GiB", 1073741824, false},
		{"500MB", 524288000, false},
		{"500 MiB", 524288000, false},
		{"100KB", 102400, false},
		{"1024 B", 1024, false},
		{"", 0, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			val, err := stringToBytes(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, val)
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
