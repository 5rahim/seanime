package comparison

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	// ValueContainsSeason regex
	seasonOrdinalRegex = regexp.MustCompile(`\d(st|nd|rd|th) [Ss].*`)

	// ExtractSeasonNumber regexes
	seasonExplicitRegex   = regexp.MustCompile(`season\s*(\d+)`)
	seasonFormatRegex     = regexp.MustCompile(`\bs0?(\d{1,2})(?:e\d|$|\s|\.)`)
	seasonOrdinalNumRegex = regexp.MustCompile(`(\d+)(?:st|nd|rd|th)\s+season`)
	romanPattern1Regex    = regexp.MustCompile(`[\s.](i{1,3}|iv|vi?i?i?|ix|x)(?:\s|$|[:,.]|['\'])`)
	romanPattern2Regex    = regexp.MustCompile(`[\s.](i{1,3}|iv|vi?i?i?|ix|x)[.\s]*(?:s\d|e\d|part)`)
	seasonTrailingNumRe   = regexp.MustCompile(`(?:^|\s)(\d{1,2})\s*$`)
	seasonPartCourRegex   = regexp.MustCompile(`(?:part|cour|specials?|sp|movie|ova|ona|oad)\s*\d{1,2}\s*$`)
	seasonJapaneseRegex   = regexp.MustCompile(`(?:第)?(\d+)\s*期`)
	// Written-out ordinal + season: "Second Season", "Third Season"
	seasonWordOrdinalRegex = regexp.MustCompile(`(?i)\b(first|second|third|fourth|fifth|sixth|seventh|eighth|ninth|tenth)\s+season\b`)

	// ValueContainsSpecial regexes
	specialRegex1 = regexp.MustCompile(`(?i)(^|(?P<show>.*?)[ _.\-(]+)(SP|OAV|OVA|OAD|ONA) ?(?P<ep>\d{1,2})(-(?P<ep2>[0-9]{1,3}))? ?(?P<title>.*)$`)
	specialRegex2 = regexp.MustCompile(`(?i)[-._( ](OVA|ONA)[-._) ]`)
	specialRegex3 = regexp.MustCompile(`(?i)[-._ ](S|SP)(?P<season>(0|00))([Ee]\d)`)
	specialRegex4 = regexp.MustCompile(`[-._({[ ]?(OVA|ONA|OAV|OAD)[])}\-._ ]?`)

	// ValueContainsIgnoredKeywords regex
	ignoredKeywordsRegex = regexp.MustCompile(`(?i)^\s?[({\[]?\s?(EXTRAS?|OVAS?|OTHERS?|SPECIALS|MOVIES|SEASONS|NC)\s?[])}]?\s?$`)

	// ValueContainsBatchKeywords regex
	batchKeywordsRegex = regexp.MustCompile(`(?i)[({\[]?\s?(EXTRAS|OVAS|OTHERS|SPECIALS|MOVIES|SEASONS|BATCH|COMPLETE|COMPLETE SERIES)\s?[])}]?\s?`)

	// ValueContainsNC regexes
	ncRegex1 = regexp.MustCompile(`(?i)(^|(?P<show>.*?)[ _.\-(]+)\b(OP|NCOP|OPED)\b ?(?P<ep>\d{1,2}[a-z]?)? ?([ _.\-)]+(?P<title>.*))?`)
	ncRegex2 = regexp.MustCompile(`(?i)(^|(?P<show>.*?)[ _.\-(]+)\b(ED|NCED)\b ?(?P<ep>\d{1,2}[a-z]?)? ?([ _.\-)]+(?P<title>.*))?`)
	ncRegex3 = regexp.MustCompile(`(?i)(^|(?P<show>.*?)[ _.\-(]+)\b(TRAILER|PROMO|PV)\b ?(?P<ep>\d{1,2}) ?([ _.\-)]+(?P<title>.*))?`)
	ncRegex4 = regexp.MustCompile(`(?i)(^|(?P<show>.*?)[ _.\-(]+)\b(OTHERS?)\b(?P<ep>\d{1,2}) ?[ _.\-)]+(?P<title>.*)`)
	ncRegex5 = regexp.MustCompile(`(?i)(^|(?P<show>.*?)[ _.\-(]+)\b(CM|COMMERCIAL|AD)\b ?(?P<ep>\d{1,2}) ?([ _.\-)]+(?P<title>.*))?`)
	ncRegex6 = regexp.MustCompile(`(?i)(^|(?P<show>.*?)[ _.\-(]+)\b(CREDITLESS|NCOP|NCED|OP|ED)\b ?(?P<ep>\d{1,2}[a-z]?)? ?([ _.\-)]+(?P<title>.*))?`)
	ncRegex7 = regexp.MustCompile(`(?i)- ?(Opening|Ending)`)

	// Roman numeral mapping
	romanToNum = map[string]int{
		"ii": 2, "iii": 3, "iv": 4, "v": 5,
		"vi": 6, "vii": 7, "viii": 8, "ix": 9,
	}

	// Written-out ordinal word mapping
	ordinalWordToNum = map[string]int{
		"first": 1, "second": 2, "third": 3, "fourth": 4, "fifth": 5,
		"sixth": 6, "seventh": 7, "eighth": 8, "ninth": 9, "tenth": 10,
	}

	IgnoredFilenames = map[string]struct{}{
		"extra": {}, "extras": {}, "ova": {}, "ovas": {}, "ona": {}, "onas": {}, "oad": {}, "oads": {}, "other": {}, "others": {}, "special": {}, "specials": {}, "movie": {}, "movies": {}, "season": {}, "seasons": {}, "batch": {},
		"complete": {}, "complete series": {}, "nc": {}, "music": {}, "mv": {}, "trailer": {}, "promo": {}, "pv": {}, "commercial": {}, "ad": {}, "opening": {}, "ending": {},
		"op": {}, "ed": {}, "ncop": {}, "nced": {}, "creditless": {},
	}
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

	if seasonOrdinalRegex.MatchString(val) {
		return true
	}

	return false
}

func ExtractSeasonNumber(val string) int {
	val = strings.ToLower(val)

	// "season X" pattern
	matches := seasonExplicitRegex.FindStringSubmatch(val)
	if len(matches) > 1 {
		season, err := strconv.Atoi(matches[1])
		if err == nil {
			return season
		}
	}

	// "SXX" or "S0X" format
	matches = seasonFormatRegex.FindStringSubmatch(val)
	if len(matches) > 1 {
		season, err := strconv.Atoi(matches[1])
		if err == nil && season > 0 && season < 20 {
			return season
		}
	}

	// Ordinal + season (e.g., "2nd season")
	matches = seasonOrdinalNumRegex.FindStringSubmatch(val)
	if len(matches) > 1 {
		season, err := strconv.Atoi(matches[1])
		if err == nil {
			return season
		}
	}

	// Roman numerals at end of title or before common markers (e.g., "Overlord II", "Title III")
	romanPatterns := []*regexp.Regexp{romanPattern1Regex, romanPattern2Regex}
	for _, re := range romanPatterns {
		matches = re.FindStringSubmatch(val)
		if len(matches) > 1 {
			romanNum := strings.ToLower(matches[1])
			if num, ok := romanToNum[romanNum]; ok {
				return num
			}
		}
	}

	// Number at the end of title (e.g., "Konosuba 2", only 2-10 range)
	// Exclude numbers preceded by "part" or "cour" as those indicate parts, not seasons
	matches = seasonTrailingNumRe.FindStringSubmatch(val)
	if len(matches) > 1 {
		// check if preceded by "part" or "cour"
		if !seasonPartCourRegex.MatchString(val) {
			season, err := strconv.Atoi(matches[1])
			if err == nil && season >= 2 && season <= 10 {
				return season
			}
		}
	}

	// Japanese season indicators (e.g., "2期")
	matches = seasonJapaneseRegex.FindStringSubmatch(val)
	if len(matches) > 1 {
		season, err := strconv.Atoi(matches[1])
		if err == nil {
			return season
		}
	}

	// Written-out ordinal seasons (e.g., "Second Season", "Third Season")
	matches = seasonWordOrdinalRegex.FindStringSubmatch(val)
	if len(matches) > 1 {
		ordinalWord := strings.ToLower(matches[1])
		if num, ok := ordinalWordToNum[ordinalWord]; ok {
			return num
		}
	}

	return -1
}

func ValueContainsSpecial(val string) bool {
	regexes := []*regexp.Regexp{
		specialRegex1,
		specialRegex2,
		specialRegex3,
		specialRegex4,
	}

	for _, regex := range regexes {
		if regex.MatchString(val) {
			return true
		}
	}

	return false
}

func ValueContainsIgnoredKeywords(val string) bool {
	return ignoredKeywordsRegex.MatchString(val)
}

func ValueContainsBatchKeywords(val string) bool {
	return batchKeywordsRegex.MatchString(val)
}

func ValueContainsNC(val string) bool {
	regexes := []*regexp.Regexp{
		ncRegex1,
		ncRegex2,
		ncRegex3,
		ncRegex4,
		ncRegex5,
		ncRegex6,
		ncRegex7,
	}

	for _, regex := range regexes {
		if regex.MatchString(val) {
			return true
		}
	}

	return false
}

// ExtractNCType parses a filename to determine the NC (non-content) type and returns the AniDB episode prefix.
// Returns the prefix (e.g. "OP", "ED") and true if found
func ExtractNCType(val string) (string, bool) {
	// OP-type regexes OP|NCOP|OPED, CREDITLESS|NCOP|NCED|OP|ED
	for _, re := range []*regexp.Regexp{ncRegex1, ncRegex6} {
		matches := re.FindStringSubmatch(val)
		if len(matches) > 0 {
			keyword := strings.ToUpper(matches[3])
			switch keyword {
			case "OP", "NCOP", "OPED", "CREDITLESS":
				if keyword == "CREDITLESS" {
					// can't determine OP vs ED from "CREDITLESS" alone, skip
					continue
				}
				return "OP", true
			case "ED", "NCED":
				return "ED", true
			}
		}
	}

	// ED|NCED
	if ncRegex2.MatchString(val) {
		return "ED", true
	}

	if matches := ncRegex7.FindStringSubmatch(val); len(matches) > 0 {
		keyword := strings.ToLower(matches[1])
		if keyword == "opening" {
			return "OP", true
		}
		if keyword == "ending" {
			return "ED", true
		}
	}

	return "", false
}
