package anime

// DEVNOTE: The structs are defined in this file because they are imported by both the autodownloader package and the db package.
// Defining them in the autodownloader package would create a circular dependency because the db package imports these structs.

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

	// AutoDownloaderRule is a rule that is used to automatically download media.
	// The structs are sent to the client, thus adding `dbId` to facilitate mutations.
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
		AdditionalTerms     []string                              `json:"additionalTerms"`
	}
)
