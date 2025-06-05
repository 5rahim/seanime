package comparison

import (
	"regexp"
	"strconv"
	"strings"
)

func ValueContainsSeason(val string) bool {
	val = strings.ToLower(val)

	if strings.IndexRune(val, '第') != -1 {
		return false
	}
	if ValueContainsSpecial(val) {
		return false
	}

	if strings.Contains(val, "season") {
		return true
	}

	re := regexp.MustCompile(`\d(st|nd|rd|th) [Ss].*`)
	if re.MatchString(val) {
		return true
	}

	return false
}

func ExtractSeasonNumber(val string) int {
	val = strings.ToLower(val)

	// Check for the word "season" followed by a number
	re := regexp.MustCompile(`season (\d+)`)
	matches := re.FindStringSubmatch(val)
	if len(matches) > 1 {
		season, err := strconv.Atoi(matches[1])
		if err == nil {
			return season
		}
	}

	// Check for a number followed by "st", "nd", "rd", or "th", followed by "s" or "S"
	re = regexp.MustCompile(`(\d+)(st|nd|rd|th) [sS]`)
	matches = re.FindStringSubmatch(val)
	if len(matches) > 1 {
		season, err := strconv.Atoi(matches[1])
		if err == nil {
			return season
		}
	}

	// No season number found
	return -1
}

// ExtractResolutionInt extracts the resolution from a string and returns it as an integer.
// This is used for comparing resolutions.
// If the resolution is not found, it returns 0.
func ExtractResolutionInt(val string) int {
	val = strings.ToLower(val)

	if strings.Contains(strings.ToUpper(val), "4K") {
		return 2160
	}
	if strings.Contains(val, "2160") {
		return 2160
	}
	if strings.Contains(val, "1080") {
		return 1080
	}
	if strings.Contains(val, "720") {
		return 720
	}
	if strings.Contains(val, "540") {
		return 540
	}
	if strings.Contains(val, "480") {
		return 480
	}

	re := regexp.MustCompile(`^\d{3,4}([pP])$`)
	matches := re.FindStringSubmatch(val)
	if len(matches) > 1 {
		res, err := strconv.Atoi(matches[1])
		if err != nil {
			return 0
		}
		return res
	}

	return 0
}

func ValueContainsSpecial(val string) bool {
	regexes := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(^|(?P<show>.*?)[ _.\-(]+)(SP|OAV|OVA|OAD|ONA) ?(?P<ep>\d{1,2})(-(?P<ep2>[0-9]{1,3}))? ?(?P<title>.*)$`),
		regexp.MustCompile(`(?i)[-._( ](OVA|ONA)[-._) ]`),
		regexp.MustCompile(`(?i)[-._ ](S|SP)(?P<season>(0|00))([Ee]\d)`),
		regexp.MustCompile(`[({\[]?(OVA|ONA|OAV|OAD|SP|SPECIAL)[])}]?`),
	}

	for _, regex := range regexes {
		if regex.MatchString(val) {
			return true
		}
	}

	return false
}

func ValueContainsIgnoredKeywords(val string) bool {
	regexes := []*regexp.Regexp{
		regexp.MustCompile(`(?i)^\s?[({\[]?\s?(EXTRAS?|OVAS?|OTHERS?|SPECIALS|MOVIES|SEASONS|NC)\s?[])}]?\s?$`),
	}

	for _, regex := range regexes {
		if regex.MatchString(val) {
			return true
		}
	}

	return false
}
func ValueContainsBatchKeywords(val string) bool {
	regexes := []*regexp.Regexp{
		regexp.MustCompile(`(?i)[({\[]?\s?(EXTRAS|OVAS|OTHERS|SPECIALS|MOVIES|SEASONS|BATCH|COMPLETE|COMPLETE SERIES)\s?[])}]?\s?`),
	}

	for _, regex := range regexes {
		if regex.MatchString(val) {
			return true
		}
	}

	return false
}

func ValueContainsNC(val string) bool {
	regexes := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(^|(?P<show>.*?)[ _.\-(]+)\b(OP|NCOP|OPED)\b ?(?P<ep>\d{1,2}[a-z]?)? ?([ _.\-)]+(?P<title>.*))?`),
		regexp.MustCompile(`(?i)(^|(?P<show>.*?)[ _.\-(]+)\b(ED|NCED)\b ?(?P<ep>\d{1,2}[a-z]?)? ?([ _.\-)]+(?P<title>.*))?`),
		regexp.MustCompile(`(?i)(^|(?P<show>.*?)[ _.\-(]+)\b(TRAILER|PROMO|PV)\b ?(?P<ep>\d{1,2}) ?([ _.\-)]+(?P<title>.*))?`),
		regexp.MustCompile(`(?i)(^|(?P<show>.*?)[ _.\-(]+)\b(OTHERS?)\b(?P<ep>\d{1,2}) ?[ _.\-)]+(?P<title>.*)`),
		regexp.MustCompile(`(?i)(^|(?P<show>.*?)[ _.\-(]+)\b(CM|COMMERCIAL|AD)\b ?(?P<ep>\d{1,2}) ?([ _.\-)]+(?P<title>.*))?`),
		regexp.MustCompile(`(?i)(^|(?P<show>.*?)[ _.\-(]+)\b(CREDITLESS|NCOP|NCED|OP|ED)\b ?(?P<ep>\d{1,2}[a-z]?)? ?([ _.\-)]+(?P<title>.*))?`),
	}

	for _, regex := range regexes {
		if regex.MatchString(val) {
			return true
		}
	}

	return false
}
