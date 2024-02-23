package util

import (
	"regexp"
	"strings"
)

func ExtractSeasonNumber(title string) (int, string) {
	title = strings.ToLower(title)

	rgx := regexp.MustCompile(`((?P<a>\d+)(st|nd|rd|th)?\s*(season))|((season)\s*(?P<b>\d+))`)

	matches := rgx.FindStringSubmatch(title)
	if len(matches) < 1 {
		return 0, title
	}
	m := matches[rgx.SubexpIndex("a")]
	if m == "" {
		m = matches[rgx.SubexpIndex("b")]
	}
	if m == "" {
		return 0, title
	}
	ret, ok := StringToInt(m)
	if !ok {
		return 0, title
	}

	cTitle := strings.TrimSpace(rgx.ReplaceAllString(title, ""))

	return ret, cTitle
}

func ExtractPartNumber(title string) (int, string) {
	title = strings.ToLower(title)

	rgx := regexp.MustCompile(`((?P<a>\d+)(st|nd|rd|th)?\s*(cour|part))|((cour|part)\s*(?P<b>\d+))`)

	matches := rgx.FindStringSubmatch(title)
	if len(matches) < 1 {
		return 0, title
	}
	m := matches[rgx.SubexpIndex("a")]
	if m == "" {
		m = matches[rgx.SubexpIndex("b")]
	}
	if m == "" {
		return 0, title
	}
	ret, ok := StringToInt(m)
	if !ok {
		return 0, title
	}

	cTitle := strings.TrimSpace(rgx.ReplaceAllString(title, ""))

	return ret, cTitle

}
