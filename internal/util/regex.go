package util

import "regexp"

// MatchesRegex checks if a string matches a regex pattern
func MatchesRegex(str, pattern string) (bool, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}
	return re.MatchString(str), nil
}
