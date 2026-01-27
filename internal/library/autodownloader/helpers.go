package autodownloader

import (
	"seanime/internal/library/anime"
	"strings"

	"github.com/samber/lo"
)

func GetReleaseGroupToResolutionsMap(rules []*anime.AutoDownloaderRule) map[string][]string {
	groups := make(map[string][]string)
	for _, rule := range rules {
		for _, releaseGroup := range rule.ReleaseGroups {
			if _, ok := groups[strings.ToLower(releaseGroup)]; ok {
				groups[strings.ToLower(releaseGroup)] = append(groups[strings.ToLower(releaseGroup)], rule.Resolutions...)
			} else {
				groups[strings.ToLower(releaseGroup)] = rule.Resolutions
			}
		}
	}
	for k := range groups {
		groups[k] = lo.Uniq(lo.Map(groups[k], func(item string, _ int) string {
			return strings.TrimSpace(strings.ToLower(item))
		}))
	}
	return groups
}
