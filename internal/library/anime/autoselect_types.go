package anime

type (
	AutoSelectProfileRuleConditionAction string

	AutoSelectPreference string

	AutoSelectProfile struct {
		DbID          uint     `json:"dbId"`
		Providers     []string `json:"providers"`     // Ordered list of preferred providers (max 3)
		ReleaseGroups []string `json:"releaseGroups"` // Preferred groups (e.g., ["SubsPlease", "Erai-raws"])
		Resolutions   []string `json:"resolutions"`   // Preferred resolutions (e.g., ["1080p", "720p"])
		ExcludeTerms  []string `json:"excludeTerms"`  // Can exclude terms like "CamRip"
		// Metadata preferences
		PreferredLanguages      []string             `json:"preferredLanguages"` // Ordered list, e.g. ["jp", "en"]
		PreferredCodecs         []string             `json:"preferredCodecs"`    // Ordered list, e.g. ["HEVC, x265, H.265", "AVC, x264"]
		PreferredSources        []string             `json:"preferredSources"`   // Ordered list, e.g. ["BDRip, BD RIP", "AT-X"]
		MultipleAudioPreference AutoSelectPreference `json:"multipleAudioPreference"`
		MultipleSubsPreference  AutoSelectPreference `json:"multipleSubsPreference"`
		BatchPreference         AutoSelectPreference `json:"batchPreference"`
		BestReleasePreference   AutoSelectPreference `json:"bestReleasePreference"`

		RequireLanguage bool `json:"requireLanguage"` // Reject if no preferred language is found
		RequireCodec    bool `json:"requireCodec"`    // Reject if no preferred codec is found
		RequireSource   bool `json:"requireSource"`   // Reject if no preferred source is found

		// Thresholds
		MinSeeders int    `json:"minSeeders,omitempty"`
		MinSize    string `json:"minSize,omitempty"`
		MaxSize    string `json:"maxSize,omitempty"`
	}
)

const (
	AutoSelectPreferenceNeutral AutoSelectPreference = "neutral"
	AutoSelectPreferencePrefer  AutoSelectPreference = "prefer"
	AutoSelectPreferenceAvoid   AutoSelectPreference = "avoid"
	AutoSelectPreferenceOnly    AutoSelectPreference = "only"
	AutoSelectPreferenceNever   AutoSelectPreference = "never"
)
