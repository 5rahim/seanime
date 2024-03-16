package seanime_parser

import (
	"regexp"
	"strconv"
	"strings"
)

func extractSeasonAndEpisode(input string) (season string, separator string, episode string, ok bool) {
	re := regexp.MustCompile(`(?i)^(s?)(\d+)(e|x|ep)((\d+)('|([vV]\d{1,2}))?)$`)

	captures := re.FindStringSubmatch(input)

	if captures == nil {
		return "", "", "", false
	}

	season = strings.TrimSpace(captures[2])
	separator = strings.TrimSpace(captures[3])
	episode = strings.TrimSpace(captures[4])
	ok = true

	// Make sure numbers are zero-padded to avoid capturing strings like "1x3"
	if strings.ToLower(separator) == "x" {
		//nSeason, _ := strconv.Atoi(season)
		nEpisode, _ := strconv.Atoi(episode)
		//if nSeason < 10 && !strings.HasPrefix(season, "0") {
		//	return "", "", "", false
		//}
		if season == "0" && nEpisode > 500 { // avoid 0x539
			return "", "", "", false
		}
		if nEpisode < 10 && !strings.HasPrefix(episode, "0") {
			return "", "", "", false
		}
	}

	return
}
