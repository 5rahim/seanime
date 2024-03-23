package anizip

import (
	"regexp"
	"strconv"
)

func (m *Media) GetTitle() string {
	if m == nil {
		return ""
	}
	if len(m.Titles["en"]) > 0 {
		return m.Titles["en"]
	}
	return m.Titles["ro"]
}

func (m *Media) GetMappings() *Mappings {
	if m == nil {
		return &Mappings{}
	}
	return m.Mappings
}

func (m *Media) FindEpisode(ep string) (*Episode, bool) {
	if m.Episodes == nil {
		return nil, false
	}
	episode, found := m.Episodes[ep]
	if !found {
		return nil, false
	}

	return &episode, true
}

func (m *Media) GetMainEpisodeCount() int {
	if m == nil {
		return 0
	}
	return m.EpisodeCount
}

// GetOffset returns the offset of the first episode relative to the absolute episode number.
// e.g, if the first episode's absolute number is 13, then the offset is 12.
func (m *Media) GetOffset() int {
	if m == nil {
		return 0
	}
	firstEp, found := m.FindEpisode("1")
	if !found {
		return 0
	}
	if firstEp.AbsoluteEpisodeNumber == 0 {
		return 0
	}
	return firstEp.AbsoluteEpisodeNumber - 1
}

func (e *Episode) GetTitle() string {
	eng, ok := e.Title["en"]
	if ok {
		return eng
	}
	rom, ok := e.Title["x-jat"]
	if ok {
		return rom
	}
	return ""
}

func ExtractEpisodeInteger(s string) (int, bool) {
	pattern := "[0-9]+"
	regex := regexp.MustCompile(pattern)

	// Find the first match in the input string.
	match := regex.FindString(s)

	if match != "" {
		// Convert the matched string to an integer.
		num, err := strconv.Atoi(match)
		if err != nil {
			return 0, false
		}
		return num, true
	}

	return 0, false
}

func OffsetEpisode(s string, offset int) string {
	pattern := "([0-9]+)"
	regex := regexp.MustCompile(pattern)

	// Replace the first matched integer with the incremented value.
	result := regex.ReplaceAllStringFunc(s, func(matched string) string {
		num, err := strconv.Atoi(matched)
		if err == nil {
			num = num + offset
			return strconv.Itoa(num)
		} else {
			return matched
		}
	})

	return result
}
