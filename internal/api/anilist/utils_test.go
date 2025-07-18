package anilist

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetSeason(t *testing.T) {
	tests := []struct {
		now            time.Time
		kind           GetSeasonKind
		expectedSeason MediaSeason
		expectedYear   int
	}{
		{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), GetSeasonKindCurrent, MediaSeasonWinter, 2025},
		{time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC), GetSeasonKindCurrent, MediaSeasonSpring, 2025},
		{time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC), GetSeasonKindCurrent, MediaSeasonSummer, 2025},
		{time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC), GetSeasonKindCurrent, MediaSeasonFall, 2025},
		{time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC), GetSeasonKindNext, MediaSeasonWinter, 2026},
		{time.Date(2025, 12, 31, 23, 59, 59, 999999999, time.UTC), GetSeasonKindCurrent, MediaSeasonWinter, 2025},
		{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), GetSeasonKindNext, MediaSeasonSpring, 2025},
	}

	for _, tt := range tests {
		t.Run(tt.now.Format(time.RFC3339), func(t *testing.T) {
			t.Logf("%s", tt.now.Format(time.RFC3339))
			season, year := GetSeasonInfo(tt.now, tt.kind)
			require.Equal(t, tt.expectedSeason, season, "Expected season %v, got %v", tt.expectedSeason, season)
			require.Equal(t, tt.expectedYear, year, "Expected year %d, got %d", tt.expectedYear, year)
		})
	}
}
