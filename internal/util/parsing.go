package util

import (
	"regexp"
	"strconv"
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

// StringToBytes converts a size string (e.g. "1.5GB", "200MB", "1GiB") to bytes.
// Supports B, KB, MB, GB, TB, KiB, MiB, GiB, TiB.
// All units are treated as binary (1024-based)
func StringToBytes(s string) (int64, error) {
	s = strings.TrimSpace(strings.ToUpper(s))
	if s == "" {
		return 0, nil
	}

	if strings.Contains(s, "IB") {
		s = strings.ReplaceAll(s, "IB", "B")
	}

	var multiplier int64 = 1
	var numStr string

	if strings.HasSuffix(s, "TB") {
		multiplier = 1024 * 1024 * 1024 * 1024
		numStr = strings.TrimSuffix(s, "TB")
	} else if strings.HasSuffix(s, "GB") {
		multiplier = 1024 * 1024 * 1024
		numStr = strings.TrimSuffix(s, "GB")
	} else if strings.HasSuffix(s, "MB") {
		multiplier = 1024 * 1024
		numStr = strings.TrimSuffix(s, "MB")
	} else if strings.HasSuffix(s, "KB") {
		multiplier = 1024
		numStr = strings.TrimSuffix(s, "KB")
	} else if strings.HasSuffix(s, "B") {
		numStr = strings.TrimSuffix(s, "B")
	} else {
		numStr = s // Assume raw or default to simple parse attempt
	}

	numStr = strings.TrimSpace(numStr)
	val, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, err
	}

	return int64(val * float64(multiplier)), nil
}

// NormalizeResolution normalizes a resolution string to a standard format
func NormalizeResolution(val string) string {
	val = strings.TrimSpace(val)
	valLower := strings.ToLower(val)

	if strings.Contains(valLower, "4k") || strings.Contains(valLower, "2160") {
		return "2160p"
	}
	if strings.Contains(valLower, "2k") || strings.Contains(valLower, "1440") {
		return "1440p"
	}
	if strings.Contains(valLower, "1080") {
		return "1080p"
	}
	if strings.Contains(valLower, "720") {
		return "720p"
	}
	if strings.Contains(valLower, "540") {
		return "540p"
	}
	if strings.Contains(valLower, "480") {
		return "480p"
	}
	if strings.Contains(valLower, "360") {
		return "360p"
	}
	if strings.Contains(valLower, "240") {
		return "240p"
	}
	if strings.Contains(valLower, "144") {
		return "144p"
	}

	return val // Return original if no standard resolution found
}

// ExtractResolutionInt extracts the resolution from a string and returns it as an integer.
// This is used for comparing resolutions.
// If the resolution is not found, it returns 0.
func ExtractResolutionInt(val string) int {
	val = strings.ToLower(val)

	if strings.Contains(val, "4k") || strings.Contains(val, "2160") {
		return 2160
	}
	if strings.Contains(val, "2k") || strings.Contains(val, "1440") {
		return 1440
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
