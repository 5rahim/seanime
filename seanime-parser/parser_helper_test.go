package seanime_parser

import "testing"

func TestExtractSeasonAndEpisode(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		expectedSeason    string
		expectedSeparator string
		expectedEpisode   string
		expectedOK        bool
	}{
		{"valid input1", "S1E2", "1", "E", "2", true},
		{"valid input2", "S01E02", "01", "E", "02", true},
		{"valid input2", "S01E02v2", "01", "E", "02v2", true},
		{"valid input3", "s2e3", "2", "e", "3", true},
		{"valid input4", "03x04", "03", "x", "04", true},
		{"invalid input5", "1x04", "1", "x", "04", true},
		{"invalid input1", "abc", "", "", "", false},
		{"invalid input2", "3x4", "", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			season, separator, episode, ok := extractSeasonAndEpisode(tt.input)
			if season != tt.expectedSeason || separator != tt.expectedSeparator || episode != tt.expectedEpisode || ok != tt.expectedOK {
				t.Errorf("got %v %v %v %v, want %v %v %v %v", season, separator, episode, ok, tt.expectedSeason, tt.expectedSeparator, tt.expectedEpisode, tt.expectedOK)
			}
		})
	}
}
