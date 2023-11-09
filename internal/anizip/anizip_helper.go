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

func (m *Media) FindEpisode(id string) (*Episode, bool) {
	if m.Episodes == nil {
		return nil, false
	}
	episode, found := m.Episodes[id]
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
	// Define a regular expression pattern to match integers in the string.
	pattern := "[0-9]+"

	// Compile the regular expression pattern.
	regex := regexp.MustCompile(pattern)

	// Find the first match in the input string.
	match := regex.FindString(s)

	if match != "" {
		// Convert the matched string to an integer using strconv.Atoi.
		num, err := strconv.Atoi(match)
		if err != nil {
			return 0, false
		}
		return num, true
	}

	// If no match was found, return an error or a default value as needed.
	return 0, false
}

func OffsetEpisode(s string, offset int) string {
	// Define a regular expression pattern to match integers in the string.
	pattern := "([0-9]+)"

	// Compile the regular expression pattern.
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
