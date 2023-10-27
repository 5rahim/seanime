package comparison

import (
	"regexp"
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

func ValueContainsSpecial(val string) bool {
	regexes := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(^|(?P<show>.*?)[ _.\-(]+)(SP|OAV|OVA|OAD|ONA) ?(?P<ep>\d{1,2})(-(?P<ep2>[0-9]{1,3}))? ?(?P<title>.*)$`),
		regexp.MustCompile(`(?i)[-._( ](OVA|ONA)[-._) ]`),
		regexp.MustCompile(`(?i)[-._ ](S|SP)(?P<season>(0|00))([Ee]\d)`),
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
		regexp.MustCompile(`(?i)(^|(?P<show>.*?)[ _.\-(]+)(SP|OAV|OVA|OAD|ONA) ?(?P<ep>\d{1,2})(-(?P<ep2>[0-9]{1,3}))? ?(?P<title>.*)$`),
		regexp.MustCompile(`(?i)[-._( ](OVA|ONA)[-._) ]`),
		regexp.MustCompile(`(?i)[-._ ](S|SP)(?P<season>(0|00))([Ee]\d)`),
		regexp.MustCompile(`(?i)(^|(?P<show>.*?)[ _.\-(]+)(OP|NCOP) ?(?P<ep>\d{1,2}[a-z]?)? ?([ _.\-)]+(?P<title>.*))?`),
		regexp.MustCompile(`(?i)(^|(?P<show>.*?)[ _.\-(]+)(ED|NCED) ?(?P<ep>\d{1,2}[a-z]?)? ?([ _.\-)]+(?P<title>.*))?`),
		regexp.MustCompile(`(?i)(^|(?P<show>.*?)[ _.\-(]+)(TRAILER|PROMO|PV|T) ?(?P<ep>\d{1,2}) ?([ _.\-)]+(?P<title>.*))?`),
		regexp.MustCompile(`(?i)(^|(?P<show>.*?)[ _.\-(]+)(O|OTHERS?)(?P<ep>\d{1,2}) ?[ _.\-)]+(?P<title>.*)`),
	}

	for _, regex := range regexes {
		if regex.MatchString(val) {
			return true
		}
	}

	return false
}
