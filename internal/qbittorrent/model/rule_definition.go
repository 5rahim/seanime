package qbittorrent_model

type RuleDefinition struct {
	// Whether the rule is enabled
	Enabled bool `json:"enabled"`
	// The substring that the torrent name must contain
	MustContain string `json:"mustContain"`
	// The substring that the torrent name must not contain
	MustNotContain string `json:"mustNotContain"`
	// Enable regex mode in "mustContain" and "mustNotContain"
	UseRegex bool `json:"useRegex"`
	// Episode filter definition
	EpisodeFilter string `json:"episodeFilter"`
	// Enable smart episode filter
	SmartFilter bool `json:"smartFilter"`
	// The list of episode IDs already matched by smart filter
	PreviouslyMatchedEpisodes []string `json:"previouslyMatchedEpisodes"`
	// The feed URLs the rule applied to
	AffectedFeeds []string `json:"affectedFeeds"`
	// Ignore sunsequent rule matches
	IgnoreDays int `json:"ignoreDays"`
	// The rule last match time
	LastMatch string `json:"lastMatch"`
	// Add matched torrent in paused mode
	AddPaused bool `json:"addPaused"`
	// Assign category to the torrent
	AssignedCategory string `json:"assignedCategory"`
	// Save torrent to the given directory
	SavePath string `json:"savePath"`
}
