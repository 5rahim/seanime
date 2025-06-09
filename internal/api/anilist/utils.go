package anilist

import (
	"time"
)

type GetSeasonKind int

const (
	GetSeasonKindCurrent GetSeasonKind = iota
	GetSeasonKindNext
	GetSeasonKindPrevious
)

func GetSeasonInfo(now time.Time, kind GetSeasonKind) (MediaSeason, int) {
	month, year := now.Month(), now.Year()

	getSeasonIndex := func(m time.Month) int {
		switch {
		case m >= 3 && m <= 5: // spring: 3, 4, 5
			return 1
		case m >= 6 && m <= 8: // summer: 6, 7, 8
			return 2
		case m >= 9 && m <= 11: // fall: 9, 10, 11
			return 3
		default: // winter: 12, 1, 2
			return 0
		}
	}

	seasons := []MediaSeason{MediaSeasonWinter, MediaSeasonSpring, MediaSeasonSummer, MediaSeasonFall}
	var index int

	switch kind {
	case GetSeasonKindCurrent:
		index = getSeasonIndex(month)

	case GetSeasonKindNext:
		nextMonth := month + 3
		nextYear := year
		if nextMonth > 12 {
			nextMonth -= 12
			nextYear++
		}
		index = getSeasonIndex(nextMonth)
		year = nextYear

	case GetSeasonKindPrevious:
		prevMonth := month - 3
		prevYear := year
		if prevMonth <= 0 {
			prevMonth += 12
			prevYear--
		}
		index = getSeasonIndex(prevMonth)
		year = prevYear
	}

	return seasons[index], year
}
