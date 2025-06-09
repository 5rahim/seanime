package anilist

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetSeason(t *testing.T) {
	tests := []struct {
		now            time.Time
		expectedSeason MediaSeason
		expectedYear   int
		kind           GetSeasonKind
	}{
		{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), MediaSeasonWinter, 2025, GetSeasonKindCurrent},
		{time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC), MediaSeasonSpring, 2025, GetSeasonKindCurrent},
		{time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC), MediaSeasonSummer, 2025, GetSeasonKindCurrent},
		{time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC), MediaSeasonFall, 2025, GetSeasonKindCurrent},
		{time.Date(2025, 12, 31, 23, 59, 59, 999999999, time.UTC), MediaSeasonWinter, 2025, GetSeasonKindCurrent},
		{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), MediaSeasonSpring, 2025, GetSeasonKindNext},
	}

	for _, test := range tests {
		t.Logf("%s", test.now.Format(time.RFC3339))
		season, year := GetSeason(test.now, GetSeasonKindCurrent)
		assert.Equal(t, test.expectedSeason, season)
		assert.Equal(t, test.expectedYear, year)
	}
}
