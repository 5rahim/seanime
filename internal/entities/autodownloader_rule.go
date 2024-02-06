package entities

const (
	AutoDownloaderRuleTitleComparisonContains AutoDownloaderRuleTitleComparisonType = "contains"
	AutoDownloaderRuleTitleComparisonLikely   AutoDownloaderRuleTitleComparisonType = "likely"
)

const (
	AutoDownloaderRuleEpisodeUnwatched AutoDownloaderRuleEpisodeType = "unwatched"
	AutoDownloaderRuleEpisodeAll       AutoDownloaderRuleEpisodeType = "all"
	AutoDownloaderRuleEpisodeSelected  AutoDownloaderRuleEpisodeType = "selected"
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
