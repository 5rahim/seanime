package scanner

import (
	"testing"
)

func TestNormalizeTitle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     string
		wantBase string
		season   int
		part     int
	}{
		{
			name:     "Basic title",
			input:    "Attack on Titan",
			want:     "attack on titan",
			wantBase: "attack on titan",
			season:   -1,
			part:     -1,
		},
		{
			// Season markers are stripped from normalized title for accurate matching
			// Season info is extracted into the Season field instead
			name:     "Title with season",
			input:    "Attack on Titan Season 2",
			want:     "attack on titan",
			wantBase: "attack on titan",
			season:   2,
			part:     -1,
		},
		{
			// Season and part markers are stripped from normalized title
			// They're extracted into Season and Part fields
			name:     "Title with part",
			input:    "Attack on Titan Season 3 Part 2",
			want:     "attack on titan",
			wantBase: "attack on titan",
			season:   3,
			part:     2,
		},
		{
			// Roman numerals are kept in normalized title for sequel distinction
			// e.g. help distinguish "Overlord II" from "Overlord"
			name:     "Roman numeral season",
			input:    "Overlord III",
			want:     "overlord iii",
			wantBase: "overlord iii",
			season:   3, // ExtractSeasonNumber should extract this
		},
		{
			name:     "Special characters",
			input:    "Steins;Gate",
			want:     "steins gate",
			wantBase: "steins gate",
		},
		{
			name:     "Smart quotes",
			input:    "Kino's Journey",
			want:     "kinos journey",
			wantBase: "kinos journey",
		},
		{
			name:     "The Animation suffix",
			input:    "Persona 4 The Animation",
			want:     "persona 4",
			wantBase: "persona 4",
		},
		{
			name:     "Case sensitivity",
			input:    "ATTACK ON TITAN",
			want:     "attack on titan",
			wantBase: "attack on titan",
		},
		{
			name:     "With 'The'",
			input:    "The Melancholy of Haruhi Suzumiya",
			want:     "melancholy of haruhi suzumiya",
			wantBase: "melancholy of haruhi suzumiya",
		},
		{
			name:     "With 'Episode'",
			input:    "One Piece Episode 1000",
			want:     "one piece 1000",
			wantBase: "one piece 1000",
		},
		{
			name:     "OAD/OVA",
			input:    "Attack on Titan OAD",
			want:     "attack on titan ova",
			wantBase: "attack on titan ova",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeTitle(tt.input)
			if got.Normalized != tt.want {
				t.Errorf("NormalizeTitle(%q).Normalized = %q, want %q", tt.input, got.Normalized, tt.want)
			}
			// check base title only if expected is provided (some cases might be tricky with what 'base' implies)
			if tt.wantBase != "" && got.CleanBaseTitle != tt.wantBase {
				t.Errorf("NormalizeTitle(%q).CleanBaseTitle = %q, want %q", tt.input, got.CleanBaseTitle, tt.wantBase)
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
