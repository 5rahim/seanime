package anilist

import (
	"math"
	"time"
)

// winter: 12, 1, 2
// spring: 3, 4, 5
// summer: 6, 7, 8
// fall: 9, 10, 11

type GetSeasonKind int

const (
	GetSeasonKindCurrent GetSeasonKind = iota
	GetSeasonKindNext
	GetSeasonKindPrevious
)

func GetSeason(now time.Time, kind GetSeasonKind) (MediaSeason, int) {
	month := now.Month()
	index := int(math.Floor(float64(month)/12)*4) % 4
	year := now.Year()

	seasons := []MediaSeason{MediaSeasonWinter, MediaSeasonSpring, MediaSeasonSummer, MediaSeasonFall}

	switch kind {
	case GetSeasonKindCurrent:
	case GetSeasonKindNext:
		index = int(math.Floor(float64(month+3)/12)*4) % 4
		if seasons[index] == MediaSeasonWinter {
			year++
		}
	case GetSeasonKindPrevious:
		index = int(math.Floor(float64(month-3)/12)*4) % 4
		if seasons[index] == MediaSeasonFall {
			year--
		}
	}

	return seasons[index], year
}
