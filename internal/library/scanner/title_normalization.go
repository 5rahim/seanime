package scanner

import (
	"regexp"
	"seanime/internal/library/anime"
	"seanime/internal/util/comparison"
	"strconv"
	"strings"
	"unicode"
)

// Noise words that should be weighted less
var noiseWords = map[string]struct{}{
	"the": {}, "a": {}, "an": {}, "of": {}, "to": {}, "in": {}, "for": {},
	"on": {}, "with": {}, "at": {}, "by": {}, "from": {}, "as": {}, "is": {},
	"it": {}, "that": {}, "this": {}, "be": {}, "are": {}, "was": {}, "were": {},
	// japanese particles/common words
	"no": {}, "wa": {}, "wo": {}, "ga": {}, "ni": {}, "de": {}, "ka": {},
	"mo": {}, "ya": {}, "e": {}, "he": {},
	// common anime title words
	"anime": {}, "ova": {}, "ona": {}, "oad": {}, "tv": {}, "movie": {},
	"nc": {}, "nced": {}, "ncop": {},
	"extras": {}, "ending": {}, "opening": {}, "preview": {},
	"special": {}, "specials": {}, "sp": {}, "finale": {},
	"season": {}, "uncensored": {}, "censored": {}, "bluray": {},
}

var ordinalToNumber = map[string]int{
	"first": 1, "1st": 1,
	"second": 2, "2nd": 2,
	"third": 3, "3rd": 3,
	"fourth": 4, "4th": 4,
	"fifth": 5, "5th": 5,
	"sixth": 6, "6th": 6,
	"seventh": 7, "7th": 7,
	"eighth": 8, "8th": 8,
	"ninth": 9, "9th": 9,
	"tenth": 10, "10th": 10,
}

type fileFormatType int

const (
	fileFormatUnknown fileFormatType = iota
	fileFormatOVA
	fileFormatSpecial
	fileFormatMovie
	fileFormatNC
)

var fileOVARegex = regexp.MustCompile(`(?i)(?:\b|_|\d)(?:OVA|OAD|OAV)\s*\d*(?:\b|_)`)
var fileSpecialRegex = regexp.MustCompile(`(?i)(?:\b|_)(?:SP|Specials?)\s*\d*(?:\b|_)`)
var fileMovieRegex = regexp.MustCompile(`(?i)(?:\b|_)(?:Movie|Film|Gekijouban|Gekijō|Gekijyou)(?:\b|_)`)
var extrasFolderRegex = regexp.MustCompile(`(?i)(?:^|[/\\])(Extras?|Specials?)(?:[/\\]|$)`)

// Roman numerals
// Note: skip I and X bc they are ambiguous
var romanToNumber = map[string]string{
	"ii": "2", "iii": "3", "iv": "4", "v": "5",
	"vi": "6", "vii": "7", "viii": "8", "ix": "9",
	"xi": "11", "xii": "12", "xiii": "13",
}

// isSeparatorChar returns true for characters that should be normalized to spaces
func isSeparatorChar(r rune) bool {
	switch r {
	case '_', '.', '-', ':', ';', ',', '|':
		return true
	}
	return false
}

func isAlphanumOrSpace(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == ' '
}

// collapseWhitespace collapses multiple whitespace characters into single spaces
// and trims leading/trailing whitespace. avoids allocating an intermediate slice.
func collapseWhitespace(s string) string {
	if s == "" {
		return s
	}

	var sb strings.Builder
	sb.Grow(len(s)) // estimate capacity

	inSpace := true // start as true to trim leading whitespace
	for _, r := range s {
		isSpace := r == ' ' || r == '\t' || r == '\n' || r == '\r'
		if isSpace {
			if !inSpace {
				sb.WriteByte(' ')
				inSpace = true
			}
		} else {
			sb.WriteRune(r)
			inSpace = false
		}
	}

	result := sb.String()
	// trim trailing space if present
	if len(result) > 0 && result[len(result)-1] == ' ' {
		result = result[:len(result)-1]
	}
	return result
}

// Season patterns
var (
	// "Season 2", "S2", "S02", "2nd Season", etc.
	seasonPatternExplicit = regexp.MustCompile(`(?i)\b(?:season|s|series)\s*0*(\d+)\b`)
	seasonPatternOrdinal  = regexp.MustCompile(`(?i)\b(\d+)(?:st|nd|rd|th)\s*(?:part|season|series)\b`)
	seasonPatternSuffix   = regexp.MustCompile(`(?i)\b(\d+)\s*(?:期|シーズン)\b`)
	// Written-out ordinal + season: "Second Season", "Third Season", etc.
	seasonPatternWordOrdinal = regexp.MustCompile(`(?i)\b(first|second|third|fourth|fifth|sixth|seventh|eighth|ninth|tenth)\s+(?:season)\b`)

	// Part patterns, e.g. "Part 2", "Part II", "Cour 2", "2nd Part"
	partPatternExplicit = regexp.MustCompile(`(?i)\b(?:part|cour)\s*0*(\d+)\b`)
	partPatternOrdinal  = regexp.MustCompile(`(?i)\b(\d+)(?:st|nd|rd|th)\s*(?:part|cour)\b`)
	partPatternRoman    = regexp.MustCompile(`(?i)\b(?:part|cour)\s+(i{1,3}|iv|vi?i?i?|ix|x)\b`)

	// Year patterns
	yearParenRegex      = regexp.MustCompile(`\((\d{4})\)`)
	yearStandaloneRegex = regexp.MustCompile(`\b(19\d{2}|20\d{2})\b`)
)

// NormalizedTitle holds the normalized form and extracted metadata
type NormalizedTitle struct {
	Original       string   `json:"original"`
	Normalized     string   `json:"normalized"`
	CleanBaseTitle string   `json:"cleanBaseTitle"` // Title with roman numerals, season/part/year stripped
	DenoisedTitle  string   `json:"denoisedTitle"`  // Title with noise words removed
	Tokens         []string `json:"tokens"`
	Season         int      `json:"season"`
	Part           int      `json:"part"`
	Year           int      `json:"year"`
	IsMain         bool     `json:"isMain"`  // Whether this title is a main title (romaji,english)
	IsExtra        bool     `json:"isExtra"` // Whether this title is an extra/fallback title (like a root folder title)
}

// NormalizeTitle creates a normalized version of a title for matching
func NormalizeTitle(title string) *NormalizedTitle {
	if title == "" {
		return &NormalizedTitle{}
	}

	result := &NormalizedTitle{
		Original: title,
		Season:   -1,
		Part:     -1,
		Year:     -1,
	}

	// Extract metadata
	result.Season = comparison.ExtractSeasonNumber(title)
	// Fallback: try written-out ordinal seasons ("Second Season", etc.)
	if result.Season == -1 {
		result.Season = extractWordOrdinalSeason(title)
	}
	result.Part = ExtractPartNumber(title)
	result.Year = ExtractYear(title)

	// Normalize the full title
	normalizedFull := normalizeString(title)
	result.Normalized = normalizedFull

	result.CleanBaseTitle, result.DenoisedTitle = computeCleanBaseTitle(normalizedFull)

	// Tokenize
	result.Tokens = tokenize(normalizedFull)

	return result
}

func normalizeString(s string) string {
	s = strings.ToLower(s)

	// Macrons to double vowels
	s = strings.ReplaceAll(s, "ō", "ou")
	s = strings.ReplaceAll(s, "ū", "uu")

	// Character replacements
	s = strings.ReplaceAll(s, "@", "a")
	s = strings.ReplaceAll(s, "×", " x ")
	s = strings.ReplaceAll(s, "꞉", ":")
	s = strings.ReplaceAll(s, "＊", " * ")

	s = replaceWord(s, "the animation", "")
	s = replaceWord(s, "the", "")
	s = replaceWord(s, "episode", "")
	s = replaceWord(s, "oad", "ova")
	s = replaceWord(s, "oav", "ova")
	s = replaceWord(s, "specials", "sp")
	s = replaceWord(s, "special", "sp")
	s = strings.ReplaceAll(s, "(tv)", "")
	s = replaceWord(s, "&", "and")

	// Replace possessive 's to prevent them from adhering to words
	s = strings.ReplaceAll(s, "'s", " ")
	s = strings.ReplaceAll(s, "’s", " ")
	s = strings.ReplaceAll(s, "`s", " ")

	// Replace smart quotes and apostrophes
	s = strings.ReplaceAll(s, "'", "")
	s = strings.ReplaceAll(s, "’", "")
	s = strings.ReplaceAll(s, "`", "")
	s = strings.ReplaceAll(s, "\"", "")
	s = strings.ReplaceAll(s, "“", "")
	s = strings.ReplaceAll(s, "”", "")

	// Normalize separators to spaces
	// normalize separators and non-alphanumeric characters
	var sb strings.Builder
	sb.Grow(len(s))
	prevWasSpace := false
	for _, r := range s {
		if isSeparatorChar(r) || !isAlphanumOrSpace(r) {
			// convert to space but avoid consecutive spaces
			if !prevWasSpace {
				sb.WriteByte(' ')
				prevWasSpace = true
			}
		} else {
			sb.WriteRune(r)
			prevWasSpace = r == ' '
		}
	}
	s = sb.String()

	// Remove season markers entirely from normalized title
	// Season/part numbers are extracted separately for scoring
	// We don't want "Title S2" to match "Other Title 2" just because of the bare "2"
	s = seasonPatternExplicit.ReplaceAllString(s, " ")    // "Season X", "SX", "S0X"
	s = seasonPatternOrdinal.ReplaceAllString(s, " ")     // "2nd Season", "3rd Season"
	s = seasonPatternWordOrdinal.ReplaceAllString(s, " ") // "Second Season", "Third Season"
	s = seasonPatternSuffix.ReplaceAllString(s, " ")      // Japanese "2期", "シーズン"
	// Remove part markers entirely
	s = partPatternExplicit.ReplaceAllString(s, " ") // "Part X", "Cour X"
	s = partPatternOrdinal.ReplaceAllString(s, " ")  // "2nd Part"
	s = partPatternRoman.ReplaceAllString(s, " ")    // "Part II"

	// Collapse whitespace
	s = collapseWhitespace(s)

	// Devnote: intentionally keep Roman numerals (II, III, etc.) in the normalized title
	// They help distinguish sequels like "Overlord II" from "Overlord"
	// Season extraction handles them separately for scoring

	return s
}

// replaceWord replaces all occurrences of old with new in s but only if old is a whole word.
// It matches case-insensitively since s is expected to be lower-cased, but the implementation relies on exact match of old.
// Assumes s is already lower-cased if old is lower-cased.
func replaceWord(s string, oldStr, newStr string) string {
	if s == "" || oldStr == "" {
		return s
	}

	var sb strings.Builder
	// estimate the size to be roughly the same
	sb.Grow(len(s))

	start := 0
	oldLen := len(oldStr)

	for {
		idx := strings.Index(s[start:], oldStr)
		if idx == -1 {
			sb.WriteString(s[start:])
			break
		}

		absIdx := start + idx

		// Check boundaries
		isStartBoundary := absIdx == 0 || !isAlphanumeric(rune(s[absIdx-1]))
		isEndBoundary := absIdx+oldLen == len(s) || !isAlphanumeric(rune(s[absIdx+oldLen]))

		if isStartBoundary && isEndBoundary {
			sb.WriteString(s[start:absIdx])
			sb.WriteString(newStr)
			start = absIdx + oldLen
		} else {
			sb.WriteString(s[start : absIdx+1]) // advance by 1 to avoid infinite loop on same match
			start = absIdx + 1
		}
	}

	return sb.String()
}

func isAlphanumeric(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z')
}

// tokenize splits a normalized string into tokens
func tokenize(s string) []string {
	if s == "" {
		return nil
	}

	// count tokens first to preallocate exact size
	count := 0
	inField := false
	for _, r := range s {
		isSpace := r == ' ' || r == '\t' || r == '\n' || r == '\r'
		if isSpace {
			inField = false
		} else if !inField {
			count++
			inField = true
		}
	}

	if count == 0 {
		return nil
	}

	// allocate and fill
	result := make([]string, 0, count)
	inField = false
	fieldStart := 0
	for i, r := range s {
		isSpace := r == ' ' || r == '\t' || r == '\n' || r == '\r'
		if isSpace {
			if inField {
				result = append(result, s[fieldStart:i])
				inField = false
			}
		} else {
			if !inField {
				fieldStart = i
				inField = true
			}
		}
	}
	if inField {
		result = append(result, s[fieldStart:])
	}
	return result
}

// stripRomanNumerals removes standalone roman numeral words from a normalized title string.
// e.g. "overlord iii" → "overlord", "attack on titan" → "attack on titan"
func stripRomanNumerals(s string) string {
	tokens := strings.Fields(s)
	result := tokens[:0]
	for _, t := range tokens {
		if _, ok := romanToNumber[t]; !ok {
			result = append(result, t)
		}
	}
	return strings.Join(result, " ")
}

// formatWords are the tokens we strip when building CleanBaseTitle.
var formatWords = map[string]struct{}{
	"ova": {}, "ona": {}, "oad": {}, "oav": {}, "sp": {}, "special": {}, "specials": {},
	"movie": {}, "film": {}, "tv": {}, "nc": {}, "nced": {}, "ncop": {},
	"extras": {}, "opening": {}, "ending": {}, "preview": {}, "finale": {},
}

// computeCleanBaseTitle strips roman numerals, format words and standalone numbers from a normalized title.
// e.g. "persona 4" → "persona", "attack on titan ova" → "attack on titan"
var standaloneNumberRegex = regexp.MustCompile(`\b\d+\b$`)

func computeCleanBaseTitle(normalized string) (string, string) {
	s := standaloneNumberRegex.ReplaceAllString(normalized, " ")
	tokens := strings.Fields(s)
	result := tokens[:0]
	denoisedResult := make([]string, 0, len(tokens))
	for _, t := range tokens {
		if _, isFmt := formatWords[t]; isFmt {
			continue
		}
		if _, isRoman := romanToNumber[t]; isRoman {
			continue
		}
		_, isNoise := noiseWords[t]
		if !isNoise {
			denoisedResult = append(denoisedResult, t)
		}
		result = append(result, t)
	}
	return strings.Join(result, " "), strings.Join(denoisedResult, " ")
}

// ExtractPartNumber extracts the part number from a title string
func ExtractPartNumber(val string) int {
	val = strings.ToLower(val)

	// Check explicit patterns first: "Part 2", "Cour 2"
	matches := partPatternExplicit.FindStringSubmatch(val)
	if len(matches) > 1 {
		if num, err := strconv.Atoi(matches[1]); err == nil {
			return num
		}
	}

	// Check ordinal patterns, e.g. "2nd Part", "2nd Cour"
	matches = partPatternOrdinal.FindStringSubmatch(val)
	if len(matches) > 1 {
		if num, err := strconv.Atoi(matches[1]); err == nil {
			return num
		}
	}

	// Check roman numeral patterns, e.g. "Part II", not I or X since they're ambiguous
	matches = partPatternRoman.FindStringSubmatch(val)
	if len(matches) > 1 {
		romanNum := strings.ToLower(matches[1])
		if romanNum == "i" || romanNum == "x" {
			return -1
		}
		if numStr, ok := romanToNumber[romanNum]; ok {
			if num, err := strconv.Atoi(numStr); err == nil {
				return num
			}
		}
	}

	return -1
}

// extractWordOrdinalSeason extracts a season number from written-out ordinal words
// e.g. "Second Season" -> 2, "Third Season" -> 3
func extractWordOrdinalSeason(val string) int {
	matches := seasonPatternWordOrdinal.FindStringSubmatch(val)
	if len(matches) > 1 {
		ordinalWord := strings.ToLower(matches[1])
		if num, ok := ordinalToNumber[ordinalWord]; ok {
			return num
		}
	}
	return -1
}

// ExtractYear extracts a year from a title string
func ExtractYear(val string) int {
	// Match years in parentheses first, e.g. "(2024)"
	matches := yearParenRegex.FindStringSubmatch(val)
	if len(matches) > 1 {
		if year, err := strconv.Atoi(matches[1]); err == nil && year >= 1900 && year <= 2100 {
			return year
		}
	}

	// Match standalone years, look for 4-digit numbers that could be years
	matches = yearStandaloneRegex.FindStringSubmatch(val)
	if len(matches) > 1 {
		if year, err := strconv.Atoi(matches[1]); err == nil {
			return year
		}
	}

	return -1
}

// GetSignificantTokens returns tokens that are not noise words
func GetSignificantTokens(tokens []string) []string {
	result := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if _, isNoise := noiseWords[token]; !isNoise && len(token) > 1 {
			result = append(result, token)
		}
	}
	return result
}

// getSignificantTokensInto filters tokens that are not noise words into the provided slice.
// This avoids allocations when the caller can reuse a slice.
func getSignificantTokensInto(tokens []string, dst []string) []string {
	for _, token := range tokens {
		if _, isNoise := noiseWords[token]; !isNoise && len(token) > 1 {
			dst = append(dst, token)
		}
	}
	return dst
}

func IsNoiseWord(word string) bool {
	_, isNoise := noiseWords[strings.ToLower(word)]
	return isNoise
}

// TokenMatchRatio calculates the ratio of matching tokens between two token sets.
func TokenMatchRatio(tokensA, tokensB []string) float64 {
	if len(tokensA) == 0 || len(tokensB) == 0 {
		return 0.0
	}

	// Create a set of tokensB for O(1) lookup
	setB := getTokenSet()
	defer putTokenSet(setB)
	for _, t := range tokensB {
		setB[t] = struct{}{}
	}

	// Count matches
	matches := 0
	for _, t := range tokensA {
		if _, found := setB[t]; found {
			matches++
		}
	}

	// Return ratio based on the smaller set (more lenient for subset matching)
	minLen := len(tokensA)
	if len(tokensB) < minLen {
		minLen = len(tokensB)
	}

	return float64(matches) / float64(minLen)
}

// WeightedTokenMatchRatio calculates match ratio with noise words weighted less.
func WeightedTokenMatchRatio(tokensA, tokensB []string) float64 {
	if len(tokensA) == 0 || len(tokensB) == 0 {
		return 0.0
	}

	setB := getTokenSet()
	defer putTokenSet(setB)
	for _, t := range tokensB {
		setB[t] = struct{}{}
	}

	totalWeight := 0.0
	matchedWeight := 0.0

	for _, t := range tokensA {
		weight := 1.0
		if IsNoiseWord(t) {
			weight = 0.3 // Noise words contribute less
		} else if isYearToken(t) {
			weight = 0.5 // Years contribute less
		}
		totalWeight += weight

		if _, found := setB[t]; found {
			matchedWeight += weight
		}
	}

	if totalWeight == 0 {
		return 0.0
	}

	return matchedWeight / totalWeight
}

func isYearToken(token string) bool {
	if len(token) != 4 {
		return false
	}
	// simple check for 19xx or 20xx
	if (strings.HasPrefix(token, "19") || strings.HasPrefix(token, "20")) &&
		unicode.IsDigit(rune(token[2])) && unicode.IsDigit(rune(token[3])) {
		return true
	}
	return false
}

// ContainsAllTokens returns true if all tokens from subset are in superset
func ContainsAllTokens(subset, superset []string) bool {
	if len(subset) == 0 {
		return true
	}
	if len(superset) == 0 {
		return false
	}

	setSuper := getTokenSet()
	defer putTokenSet(setSuper)
	for _, t := range superset {
		setSuper[t] = struct{}{}
	}

	for _, t := range subset {
		if _, found := setSuper[t]; !found {
			return false
		}
	}
	return true
}

func RemoveNonAlphanumeric(s string) string {
	var result strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// HasStrongMatch returns true if there is a match between tokens that is not a year token or a noise word.
func HasStrongMatch(tokensA, tokensB []string) bool {
	if len(tokensA) == 0 || len(tokensB) == 0 {
		return false
	}

	setB := getTokenSet()
	defer putTokenSet(setB)
	for _, t := range tokensB {
		setB[t] = struct{}{}
	}

	for _, t := range tokensA {
		if IsNoiseWord(t) || isYearToken(t) {
			continue
		}
		if _, found := setB[t]; found {
			return true
		}
	}
	return false
}

// getFileFormatType detects the content format type from the filename and folder path.
// Only the filename bc folder names are too unreliable
func getFileFormatType(lf *anime.LocalFile) fileFormatType {
	name := lf.Name
	path := lf.Path

	if comparison.ValueContainsNC(name) {
		return fileFormatNC
	}
	if fileOVARegex.MatchString(name) {
		return fileFormatOVA
	}
	if fileSpecialRegex.MatchString(name) {
		return fileFormatSpecial
	}
	if fileMovieRegex.MatchString(name) {
		return fileFormatMovie
	}
	if extrasFolderRegex.MatchString(path) {
		return fileFormatSpecial
	}

	return fileFormatUnknown
}
