package scanner

import (
	"testing"
)

func TestNormalizeTitle(t *testing.T) {
	tests := []struct {
		input        string
		want         string
		wantBase     string
		season       int
		part         int
		wantDenoised string
	}{
		{
			input:        "Attack on Titan",
			want:         "attack on titan",
			wantBase:     "attack on titan",
			season:       -1,
			part:         -1,
			wantDenoised: "attack titan",
		},
		{
			// Season markers are stripped from normalized title for accurate matching
			// Season info is extracted into the Season field instead
			input:        "Attack on Titan Season 2",
			want:         "attack on titan",
			wantBase:     "attack on titan",
			season:       2,
			part:         -1,
			wantDenoised: "attack titan",
		},
		{
			// Season and part markers are stripped from normalized title
			// They're extracted into Season and Part fields
			input:        "Attack on Titan Season 3 Part 2",
			want:         "attack on titan",
			wantBase:     "attack on titan",
			season:       3,
			part:         2,
			wantDenoised: "attack titan",
		},
		{
			// Roman numerals are kept in normalized title for sequel distinction
			// e.g. help distinguish "Overlord II" from "Overlord"
			input:        "Overlord III",
			want:         "overlord iii",
			wantBase:     "overlord",
			season:       3, // ExtractSeasonNumber should extract this
			wantDenoised: "overlord",
		},
		{
			input:        "Steins;Gate",
			want:         "steins gate",
			wantBase:     "steins gate",
			wantDenoised: "steins gate",
		},
		{
			input:        "Kino's Journey",
			want:         "kino journey",
			wantBase:     "kino journey",
			wantDenoised: "kino journey",
		},
		{
			input:        "Persona 4 The Animation",
			want:         "persona 4",
			wantBase:     "persona",
			wantDenoised: "persona",
		},
		{
			input:        "86 - Eighty Six",
			want:         "86 eighty six",
			wantBase:     "86 eighty six",
			wantDenoised: "86 eighty six",
		},
		{
			input:        "ATTACK ON TITAN",
			want:         "attack on titan",
			wantBase:     "attack on titan",
			wantDenoised: "attack titan",
		},
		{
			input:        "The Melancholy of Haruhi Suzumiya",
			want:         "melancholy of haruhi suzumiya",
			wantBase:     "melancholy of haruhi suzumiya",
			wantDenoised: "melancholy haruhi suzumiya",
		},
		{
			input:        "One Piece Episode 1000",
			want:         "one piece 1000",
			wantBase:     "one piece",
			wantDenoised: "one piece",
		},
		{
			input:        "Attack on Titan OAD",
			want:         "attack on titan ova",
			wantBase:     "attack on titan",
			wantDenoised: "attack titan",
		},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := NormalizeTitle(tt.input)
			if got.Normalized != tt.want {
				t.Errorf("NormalizeTitle(%q).Normalized = %q, want %q", tt.input, got.Normalized, tt.want)
			}
			// check base title only if expected is provided (some cases might be tricky with what 'base' implies)
			if tt.wantBase != "" && got.CleanBaseTitle != tt.wantBase {
				t.Errorf("NormalizeTitle(%q).CleanBaseTitle = %q, want %q", tt.input, got.CleanBaseTitle, tt.wantBase)
			}
			// check base title only if expected is provided (some cases might be tricky with what 'base' implies)
			if tt.wantDenoised != "" && got.DenoisedTitle != tt.wantDenoised {
				t.Errorf("NormalizeTitle(%q).DenoisedTitle = %q, want %q", tt.input, got.DenoisedTitle, tt.wantDenoised)
			}
			// Check season extraction if specified
			if tt.season != 0 && got.Season != tt.season {
				t.Errorf("NormalizeTitle(%q).Season = %d, want %d", tt.input, got.Season, tt.season)
			}
			// Check part extraction if specified
			if tt.part != 0 && got.Part != tt.part {
				t.Errorf("NormalizeTitle(%q).Part = %d, want %d", tt.input, got.Part, tt.part)
			}
		})
	}
}
