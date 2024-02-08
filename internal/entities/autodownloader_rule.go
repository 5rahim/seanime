package entities

import "strings"

const (
	AutoDownloaderRuleTitleComparisonContains AutoDownloaderRuleTitleComparisonType = "contains"
	AutoDownloaderRuleTitleComparisonLikely   AutoDownloaderRuleTitleComparisonType = "likely"
)

const (
	AutoDownloaderRuleEpisodeRecent   AutoDownloaderRuleEpisodeType = "recent"
	AutoDownloaderRuleEpisodeSelected AutoDownloaderRuleEpisodeType = "selected"
)

type (
	AutoDownloaderRuleTitleComparisonType string
	AutoDownloaderRuleEpisodeType         string

	AutoDownloaderRule struct {
		DbID                uint                                  `json:"dbId"` // Will be set when fetched from the database
		Enabled             bool                                  `json:"enabled"`
		MediaId             int                                   `json:"mediaId"`
		ReleaseGroups       []string                              `json:"releaseGroups"`
		Resolutions         []string                              `json:"resolutions"`
		ComparisonTitle     string                                `json:"comparisonTitle"`
		TitleComparisonType AutoDownloaderRuleTitleComparisonType `json:"titleComparisonType"`
		EpisodeType         AutoDownloaderRuleEpisodeType         `json:"episodeType"`
		EpisodeNumbers      []int                                 `json:"episodeNumbers,omitempty"`
		Destination         string                                `json:"destination"`
	}
)

func (r *AutoDownloaderRule) IsQualityMatch(quality string) bool {
	for _, q := range r.Resolutions {
		qualityWithoutP, _ := strings.CutSuffix(q, "p")
		qWithoutP := strings.TrimSuffix(q, "p")
		if quality == q || qualityWithoutP == qWithoutP {
			return true
		}
		if strings.Contains(quality, qWithoutP) { // e.g. 1080 in 1920x1080
			return true
		}
	}
	return false
}
