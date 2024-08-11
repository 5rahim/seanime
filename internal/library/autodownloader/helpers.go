package autodownloader

import (
	"seanime/internal/library/anime"
	"strings"
)

func GetUniqueReleaseGroups(rules []*anime.AutoDownloaderRule) []string {
	uniqueReleaseGroups := make(map[string]string)
	for _, rule := range rules {
		for _, releaseGroup := range rule.ReleaseGroups {
			// make it case-insensitive
			uniqueReleaseGroups[strings.ToLower(releaseGroup)] = releaseGroup
		}
	}
	var result []string
	for k := range uniqueReleaseGroups {
		result = append(result, k)
	}
	return result
}
