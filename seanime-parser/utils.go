package seanime_parser

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

func isNumber(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func isDigitsOnly(s string) bool {
	rgx := regexp.MustCompile(`^\d+$`)
	return rgx.MatchString(s)
}

// isNumberLike checks if the provided string matches a specific pattern.
// It returns true if the string matches the pattern, otherwise false.
// The pattern is defined as follows:
//   - The string must start with one or more digits. (e.g. 123)
//   - The string may have an optional 'x' followed by one or two digits. (e.g. 03x04)
//   - The string may have an optional 'v' followed by a single digit. (e.g. 03v2)
//   - The string may end with an optional single quote ('). (e.g. 04')
func isNumberLike(s string) bool {
	rgx := regexp.MustCompile(`^(?i)\d+((x(\d{1,2}))|(v\d)|[abc])?(')?$`)
	return rgx.MatchString(s)
}

// isNumberOrLike checks if the provided string is a number, or follows a specific pattern.
// It returns true if the string is a number, or matches the specified pattern, otherwise false.
// The function relies on the helper functions isNumber and isNumberLike to determine if the string is a number
// or matches the pattern.
func isNumberOrLike(s string) bool {
	return isNumber(s) || isNumberLike(s)
}

func isNumberZeroPadded(s string) bool {
	if len(s) < 2 {
		return false
	}
	if !isNumber(s) && !isNumberLike(s) {
		return false
	}
	if strings.HasPrefix(s, "0") {
		return true
	}
	return false
}

// isOrdinalNumber returns true if the provided string is an ordinal number, otherwise false.
// It checks if the string is present in a pre-defined list of ordinal numbers in lowercase.
func isOrdinalNumber(s string) bool {
	t := []string{"first", "second", "third", "fourth", "fifth", "sixth", "seventh", "eighth", "ninth"}
	for _, n := range t {
		if strings.ToLower(s) == n {
			return true
		}
	}
	rgx := regexp.MustCompile(`^(?i)\d+(st|nd|rd|th)$`)
	return rgx.MatchString(s)
}

// getNumberFromOrdinal returns the corresponding numeric value of the ordinal string provided.
// If the provided string does not match any of the supported ordinals, it returns 0.
// Example usage: getNumberFromOrdinal("5th") => 5
func getNumberFromOrdinal(s string) (int, bool) {
	ordinals := map[string]int{
		"1st": 1, "first": 1,
		"2nd": 2, "second": 2,
		"3rd": 3, "third": 3,
		"4th": 4, "fourth": 4,
		"5th": 5, "fifth": 5,
		"6th": 6, "sixth": 6,
		"7th": 7, "seventh": 7,
		"8th": 8, "eighth": 8,
		"9th": 9, "ninth": 9,
		"10th": 10, "tenth": 10,
	}

	lowerStr := strings.ToLower(s)
	num, ok := ordinals[lowerStr]
	return num, ok
}

// isCRC32 checks if the given string represents a valid CRC32.
// It returns true if the string is a valid CRC32, otherwise it returns false.
func isCRC32(s string) bool {
	return len(s) == 8 && isHexadecimalString(s)
}

// isHexadecimalString checks if the given string represents a valid hexadecimal string.
// It returns true if the string is a valid hexadecimal string, otherwise it returns false.
func isHexadecimalString(s string) bool {
	_, err := strconv.ParseInt(s, 16, 64)
	return err == nil
}

// isResolution checks if the given string represents a valid resolution.
// It returns true if the string is a valid resolution, otherwise it returns false.
func isResolution(s string) bool {
	found, _ := regexp.Match(`\d{3,4}([pP]|[×xX\\u00D7]\d{3,4})$`, []byte(s))
	return found
}

const yearMin = 1900
const yearMax = 2050

// isYearNumber checks if the given string represents a valid year number within the range of yearMin and yearMax.
// It returns true if the string is a valid year number, otherwise it returns false.
func isYearNumber(str string) bool {
	n, err := strconv.Atoi(str)
	if err != nil {
		return false
	}

	if n >= yearMin && n <= yearMax {
		return true
	}

	return false
}

func stringToInt(str string) int {
	dotIndex := strings.IndexByte(str, '.')
	if dotIndex != -1 {
		str = str[:dotIndex]
	}
	i, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return i
}

func isLatinRune(r rune) bool {
	return unicode.In(r, unicode.Latin)
}

// findNumberInString searches for the first occurrence of a digit in the given string and returns its index.
// If no digit is found, it returns -1.
func findNumberInString(str string) int {
	for _, c := range str {
		if unicode.IsDigit(c) {
			return strings.IndexRune(str, c)
		}
	}
	return -1
}

func mergeValues(start string, values []string) string {
	merged := start
	for _, v := range values {
		merged += v
	}
	return merged
}
