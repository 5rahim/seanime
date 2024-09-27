package seanime_parser

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSeasonAndEpisode(t *testing.T) {

	tests := []struct {
		input    string
		seasons  *[]string
		episodes *[]string
		debug    bool
	}{
		//{"[Seanime] S01 E02 - An episode.mkv", &[]string{"01"}, &[]string{"02"}, true},

		// Single season
		{"[Seanime] Jujutsu Kaisen 2nd Season - 20 [720p][AV1 10bit][AAC][Multi-Sub] (Weekly).mkv", &[]string{"2"}, &[]string{"20"}, false},
		{"[Seanime] Jujutsu Kaisen Season 01.mkv", &[]string{"01"}, nil, false},
		{"[Seanime] Jujutsu Kaisen S1.mkv", &[]string{"1"}, nil, false},
		{"[Seanime] Jujutsu Kaisen 1st Season.mkv", &[]string{"1"}, nil, false},
		{"[Seanime] Jujutsu Kaisen First Season.mkv", &[]string{"1"}, nil, false},
		{"[Seanime] Jujutsu Kaisen S01v2.mkv", &[]string{"01v2"}, nil, false},

		{"Jujutsu Kaisen 2nd Season", &[]string{"2"}, nil, false},
		{"Jujutsu Kaisen Season 01", &[]string{"01"}, nil, false},
		{"Jujutsu Kaisen S1", &[]string{"1"}, nil, false},
		{"Jujutsu Kaisen 1st Season", &[]string{"1"}, nil, false},
		{"Jujutsu Kaisen First Season", &[]string{"1"}, nil, false},
		{"Jujutsu Kaisen S01v2", &[]string{"01v2"}, nil, false},

		// Season 1 Episode 2
		{"[Seanime] S01E02 - An episode.mkv", &[]string{"01"}, &[]string{"02"}, false},
		{"[Seanime] S01EP02 - An episode.mkv", &[]string{"01"}, &[]string{"02"}, false},
		{"[Seanime] Jujutsu Kaisen 01x02.mkv", &[]string{"01"}, &[]string{"02"}, false},
		{"[Seanime] Jujutsu Kaisen S01E02.mkv", &[]string{"01"}, &[]string{"02"}, false},
		{"[Seanime] Jujutsu Kaisen S1- 02.mkv", &[]string{"1"}, &[]string{"02"}, false},
		{"[Seanime] Jujutsu Kaisen S1-02.mkv", &[]string{"1"}, &[]string{"02"}, false},
		{"[Seanime] Jujutsu Kaisen S1 - 02.mkv", &[]string{"1"}, &[]string{"02"}, false},
		{"[Seanime] Jujutsu Kaisen Season 01 - 02.mkv", &[]string{"01"}, &[]string{"02"}, false},
		{"[Seanime] Jujutsu Kaisen S1 - 02.5.mkv", &[]string{"1"}, &[]string{"02.5"}, true},

		{"[Seanime] Jujutsu Kaisen Season 01 - 12.mkv", &[]string{"01"}, &[]string{"12"}, false},

		// Season 1 to 3
		{"[Seanime] Jujutsu Kaisen Seasons 1 ~ 3.mkv", &[]string{"1", "3"}, nil, false},
		{"[Seanime] Jujutsu Kaisen Seasons 01-03.mkv", &[]string{"01", "03"}, nil, false},
		{"[Seanime] Jujutsu Kaisen Season 01-03.mkv", &[]string{"01", "03"}, nil, false},
		{"[Seanime] Jujutsu Kaisen S01-03.mkv", &[]string{"01", "03"}, nil, false},
		{"[Seanime] Jujutsu Kaisen S1-3.mkv", &[]string{"1", "3"}, nil, false},

		// Multiple seasons
		{"[Seanime] Jujutsu Kaisen S1 + S2 + S3.mkv", &[]string{"1", "2", "3"}, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := newParser(tt.input)
			p.parse()
			if tt.debug {
				t.Log(p.tokenManager.tokens.Sdump())
			}

			if tt.seasons != nil {
				assertMetadataExists(t, p, metadataSeason, *tt.seasons)
			} else {
				assertMetadataDoesNotExist(t, p, metadataSeason)
			}

			if tt.episodes != nil {
				assertMetadataExists(t, p, metadataEpisodeNumber, *tt.episodes)
			} else {
				assertMetadataDoesNotExist(t, p, metadataEpisodeNumber)
			}
		})
	}

}

func TestPart(t *testing.T) {

	tests := []struct {
		input string
		parts *[]string
		debug bool
	}{
		{"[Judas] Spy x Family (Season 1 Part 2) [1080p][HEVC x265 10bit][Dual-Audio][Multi-Subs] (Batch)", &[]string{"2"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := newParser(tt.input)
			p.parse()
			if tt.debug {
				t.Log(p.tokenManager.tokens.Sdump())
			}

			if tt.parts != nil {
				assertMetadataExists(t, p, metadataPart, *tt.parts)
			} else {
				assertMetadataDoesNotExist(t, p, metadataPart)
			}
		})
	}

}

func TestEpisodeAlt(t *testing.T) {

	tests := []struct {
		input       string
		episodes    *[]string
		episodeAlts *[]string
		debug       bool
	}{
		{"[Seanime] Jujutsu Kaisen 2nd Season - 01 (14) [720p][AV1 10bit][AAC][Multi-Sub] (Weekly).mkv", &[]string{"01"}, &[]string{"14"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := newParser(tt.input)
			p.parse()
			if tt.debug {
				t.Log(p.tokenManager.tokens.Sdump())
			}

			if tt.episodeAlts != nil {
				assertMetadataExists(t, p, metadataEpisodeNumberAlt, *tt.episodeAlts)
			} else {
				assertMetadataDoesNotExist(t, p, metadataEpisodeNumberAlt)
			}

			if tt.episodes != nil {
				assertMetadataExists(t, p, metadataEpisodeNumber, *tt.episodes)
			} else {
				assertMetadataDoesNotExist(t, p, metadataEpisodeNumber)
			}
		})
	}

}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func assertMetadataExists(t *testing.T, p *parser, metadata metadataCategory, expected []string) {
	found, tkns := p.tokenManager.tokens.findWithMetadataCategory(metadata)

	if expected[0] == "" {
		assert.False(t, found)
		return
	}

	if assert.True(t, found) {
		if assert.Len(t, tkns, len(expected)) {
			for i, tkn := range tkns {
				assert.Equal(t, expected[i], tkn.getValue())
			}
		}
	}
}

func assertMetadataDoesNotExist(t *testing.T, p *parser, metadata metadataCategory) {
	found, _ := p.tokenManager.tokens.findWithMetadataCategory(metadata)
	assert.False(t, found)
}
