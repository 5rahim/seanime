package comparison

import (
	"regexp"
	"strings"
)

func ValueContainsSeson(val string) bool {
	val = strings.ToLower(val)

	if strings.IndexRune(val, '第') != -1 {
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

func ValueContainsSpecial(val string) bool {
	val = strings.ToLower(val)
	re := regexp.MustCompile(`\b(ova|special|ona)\b`)
	if re.MatchString(val) {
		return true
	}
	return false
}
