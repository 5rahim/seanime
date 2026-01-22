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

const (
	// AutoDownloaderRulePrioritySeeders if multiple torrents from different groups, prefer the one with the most seeders.
	AutoDownloaderRulePrioritySeeders AutoDownloaderRulePriorityType = "seeders"
	// AutoDownloaderRulePriorityReleaseGroup if multiple torrents from different groups, prefer the one from the top release group.
	AutoDownloaderRulePriorityReleaseGroup AutoDownloaderRulePriorityType = "releaseGroup"
	// AutoDownloaderRulePrioritySize if multiple torrents from different groups, prefer the one with the largest size.
	// Is skipped if provider doesn't return accurate sizes.
	AutoDownloaderRulePrioritySize AutoDownloaderRulePriorityType = "size"
)

const (
	// FormatActionScore: Adds/Subtracts from the total score.
	AutoDownloaderProfileRuleFormatActionScore AutoDownloaderProfileRuleFormatAction = "score"
	// AutoDownloaderProfileRuleFormatActionBlock: Immediately rejects the torrent if found (Hard Filter).
	AutoDownloaderProfileRuleFormatActionBlock AutoDownloaderProfileRuleFormatAction = "block"
	// AutoDownloaderProfileRuleFormatActionRequire: Immediately rejects the torrent if NOT found (Hard Filter).
	AutoDownloaderProfileRuleFormatActionRequire AutoDownloaderProfileRuleFormatAction = "require"
)

type (
	AutoDownloaderRuleTitleComparisonType string
	AutoDownloaderRuleEpisodeType         string
	AutoDownloaderProfileRuleFormatAction string
	AutoDownloaderRulePriorityType        string

	AutoDownloaderRule struct {
		DbID        uint   `json:"dbId"`
		Enabled     bool   `json:"enabled"`
		MediaId     int    `json:"mediaId"`
		Destination string `json:"destination"`

		// ProfileID links to a specific strategy
		// This runs IN ADDITION to any profile marked "Global".
		ProfileID *uint `json:"profileId,omitempty"`

		// Local overrides
		// e.g., If Resolutions is set here, it ignores the Profile's resolutions.
		ReleaseGroups  []string                      `json:"releaseGroups,omitempty"`
		Resolutions    []string                      `json:"resolutions,omitempty"`
		EpisodeNumbers []int                         `json:"episodeNumbers,omitempty"`
		EpisodeType    AutoDownloaderRuleEpisodeType `json:"episodeType"`

		// Text Filters
		ComparisonTitle     string                                `json:"comparisonTitle"`
		TitleComparisonType AutoDownloaderRuleTitleComparisonType `json:"titleComparisonType"`

		AdditionalTerms []string `json:"additionalTerms"`
		ExcludeTerms    []string `json:"excludeTerms"`

		// Constraints
		MinSeeders                        int   `json:"minSeeders"`
		MinSize                           int64 `json:"minSize"`
		MaxSize                           int64 `json:"maxSize"`
		CustomEpisodeNumberAbsoluteOffset int   `json:"customEpisodeNumberAbsoluteOffset,omitempty"`
		// Providers (extension IDs) If set, only torrents from these providers are considered.
		// Overrides default provider if set.
		Providers []string `json:"providers"`
	}

	AutoDownloaderProfile struct {
		DbID uint   `json:"dbId"`
		Name string `json:"name"`

		// Global If true, this profile is applied to all rules.
		Global bool `json:"global"`

		// Ordered list (e.g., ["1080p", "720p"]).
		Resolutions []string `json:"resolutions"`

		Conditions []AutoDownloaderCondition `json:"conditions"`

		// Thresholds
		MinimumScore int   `json:"minimumScore"`
		MinSeeders   int   `json:"minSeeders,omitempty"`
		MinSize      int64 `json:"minSize,omitempty"`
		MaxSize      int64 `json:"maxSize,omitempty"`
		// Providers (extension IDs) If set, only torrents from these providers are considered.
		Providers []string `json:"providers"`
	}

	AutoDownloaderCondition struct {
		DbID    uint                                  `json:"dbId"`
		Term    string                                `json:"term"`
		IsRegex bool                                  `json:"isRegex"`
		Action  AutoDownloaderProfileRuleFormatAction `json:"action"`
		Score   int                                   `json:"score"` // Only used if Action == "score"
	}

	//// AutoDownloaderProfile is a template that can be quickly applied to a rule or can be applied globally.
	//AutoDownloaderProfile struct {
	//	DbID                uint                                  `json:"dbId"` // Will be set when fetched from the database
	//	Name                string                                `json:"name"`
	//	Global       bool                                  `json:"applyGlobally"`
	//	ReleaseGroups       []string                              `json:"releaseGroups,omitempty"`
	//	Resolutions         []string                              `json:"resolutions,omitempty"`
	//	TitleComparisonType AutoDownloaderRuleTitleComparisonType `json:"titleComparisonType,omitempty"`
	//	AdditionalTerms     []string                              `json:"additionalTerms,omitempty"`
	//	ExcludeTerms        []string                              `json:"excludeTerms,omitempty"`
	//	Providers           []string                              `json:"providers,omitempty"`
	//	MinSeeders          int                                   `json:"minSeeders,omitempty"`
	//	MinSize             int64                                 `json:"minSize,omitempty"`
	//	MaxSize             int64                                 `json:"maxSize,omitempty"`
	//}
)
