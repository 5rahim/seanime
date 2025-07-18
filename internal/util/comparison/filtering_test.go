package comparison

import (
	"testing"
)

func TestValueContainsSeason(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Contains 'season' in lowercase",
			input:    "JJK season 2",
			expected: true,
		},
		{
			name:     "Contains 'season' in uppercase",
			input:    "JJK SEASON 2",
			expected: true,
		},
		{
			name:     "Contains '2nd S' in lowercase",
			input:    "Spy x Family 2nd Season",
			expected: true,
		},
		{
			name:     "Contains '2nd S' in uppercase",
			input:    "Spy x Family 2ND SEASON",
			expected: true,
		},
		{
			name:     "Does not contain 'season' or '1st S'",
			input:    "This is a test",
			expected: false,
		},
		{
			name:     "Contains special characters",
			input:    "JJK season 2 (OVA)",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ValueContainsSeason(test.input)
			if result != test.expected {
				t.Errorf("ValueContainsSeason() with args %v, expected %v, but got %v.", test.input, test.expected, result)
			}
		})
	}
}

func TestExtractSeasonNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "Contains 'season' followed by a number",
			input:    "JJK season 2",
			expected: 2,
		},
		{
			name:     "Contains a number followed by 'st', 'nd', 'rd', or 'th', followed by 's' or 'S'",
			input:    "Spy x Family 2nd S",
			expected: 2,
		},
		{
			name:     "Does not contain 'season' or '1st S'",
			input:    "This is a test",
			expected: -1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ExtractSeasonNumber(test.input)
			if result != test.expected {
				t.Errorf("ExtractSeasonNumber() with args %v, expected %v, but got %v.", test.input, test.expected, result)
			}
		})
	}
}

func TestExtractResolutionInt(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "Contains '4K' in uppercase",
			input:    "4K",
			expected: 2160,
		},
		{
			name:     "Contains '4k' in lowercase",
			input:    "4k",
			expected: 2160,
		},
		{
			name:     "Contains '2160'",
			input:    "2160",
			expected: 2160,
		},
		{
			name:     "Contains '1080'",
			input:    "1080",
			expected: 1080,
		},
		{
			name:     "Contains '720'",
			input:    "720",
			expected: 720,
		},
		{
			name:     "Contains '480'",
			input:    "480",
			expected: 480,
		},
		{
			name:     "Does not contain a resolution",
			input:    "This is a test",
			expected: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ExtractResolutionInt(test.input)
			if result != test.expected {
				t.Errorf("ExtractResolutionInt() with args %v, expected %v, but got %v.", test.input, test.expected, result)
			}
		})
	}
}

func TestValueContainsSpecial(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Contains 'OVA' in uppercase",
			input:    "JJK OVA",
			expected: true,
		},
		{
			name:     "Contains 'ova' in lowercase",
			input:    "JJK ova",
			expected: false,
		},
		{
			name:     "Does not contain special keywords",
			input:    "This is a test",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ValueContainsSpecial(test.input)
			if result != test.expected {
				t.Errorf("ValueContainsSpecial() with args %v, expected %v, but got %v.", test.input, test.expected, result)
			}
		})
	}
}

func TestValueContainsIgnoredKeywords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Contains 'EXTRAS' in uppercase",
			input:    "EXTRAS",
			expected: true,
		},
		{
			name:     "Contains 'extras' in lowercase",
			input:    "extras",
			expected: true,
		},
		{
			name:     "Does not contain ignored keywords",
			input:    "This is a test",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ValueContainsIgnoredKeywords(test.input)
			if result != test.expected {
				t.Errorf("ValueContainsIgnoredKeywords() with args %v, expected %v, but got %v.", test.input, test.expected, result)
			}
		})
	}
}

func TestValueContainsBatchKeywords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Contains 'BATCH' in uppercase",
			input:    "BATCH",
			expected: true,
		},
		{
			name:     "Contains 'batch' in lowercase",
			input:    "batch",
			expected: true,
		},
		{
			name:     "Does not contain batch keywords",
			input:    "This is a test",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ValueContainsBatchKeywords(test.input)
			if result != test.expected {
				t.Errorf("ValueContainsBatchKeywords() with args %v, expected %v, but got %v.", test.input, test.expected, result)
			}
		})
	}
}

func TestValueContainsNC(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{
			input:    "NCOP",
			expected: true,
		},
		{
			input:    "ncop",
			expected: true,
		},
		{
			input:    "One Piece - 1000 - NCOP",
			expected: true,
		},
		{
			input:    "One Piece ED 2",
			expected: true,
		},
		{
			input:    "This is a test",
			expected: false,
		}, {
			input:    "This is a test",
			expected: false,
		},
		{
			input:    "Himouto.Umaru.chan.S01E02.1080p.BluRay.Opus2.0.x265-smol",
			expected: false,
		},
		{
			input:    "Himouto.Umaru.chan.S01E02.1080p.BluRay.x265-smol",
			expected: false,
		},
		{
			input:    "One Piece - 1000 - Operation something something",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := ValueContainsNC(test.input)
			if result != test.expected {
				t.Errorf("ValueContainsNC() with args %v, expected %v, but got %v.", test.input, test.expected, result)
			}
		})
	}
}

//func TestLikelyNC(t *testing.T) {
//	tests := []struct {
//		name     string
//		input    string
//		expected bool
//	}{
//		{
//			name:     "Does not contain NC keywords 1",
//			input:    "Himouto.Umaru.chan.S01E02.1080p.BluRay.Opus2.0.x265-smol",
//			expected: false,
//		},
//		{
//			name:     "Does not contain NC keywords 2",
//			input:    "Himouto.Umaru.chan.S01E02.1080p.BluRay.x265-smol",
//			expected: false,
//		},
//		{
//			name:     "Contains NC keywords 1",
//			input:    "Himouto.Umaru.chan.S00E02.1080p.BluRay.x265-smol",
//			expected: true,
//		},
//		{
//			name:     "Contains NC keywords 2",
//			input:    "Himouto.Umaru.chan.OP02.1080p.BluRay.x265-smol",
//			expected: true,
//		},
//	}
//
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			metadata := habari.Parse(test.input)
//			var episode string
//			var season string
//
//			if len(metadata.SeasonNumber) > 0 {
//				if len(metadata.SeasonNumber) == 1 {
//					season = metadata.SeasonNumber[0]
//				}
//			}
//
//			if len(metadata.EpisodeNumber) > 0 {
//				if len(metadata.EpisodeNumber) == 1 {
//					episode = metadata.EpisodeNumber[0]
//				}
//			}
//
//			result := LikelyNC(test.input, season, episode)
//			if result != test.expected {
//				t.Errorf("ValueContainsNC() with args %v, expected %v, but got %v.", test.input, test.expected, result)
//			}
//		})
//	}
//}
