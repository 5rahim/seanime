package scanner

import "seanime/internal/library/anime"

type Config struct {
	Matching  MatchingConfig  `json:"matching"`
	Hydration HydrationConfig `json:"hydration"`
	Logs      LogsConfig      `json:"logs"`
	//Ignore    []string        `json:"ignore"`
}

type LogsConfig struct {
	Verbose bool `json:"verbose"`
}

type MatchingConfig struct {
	Rules []*MatchingRule `json:"rules"`
}

// MatchingRule defines a rule for matching a filename or folder name to a Media ID
//
//	Example:
//		{
//			"pattern": "(?i)(.*)/(Mob Psycho)/(Season 1)"
//			"mediaId": 12345
//		}
type MatchingRule struct {
	Pattern string `json:"pattern"`
	// The Media ID to force match to
	MediaID int `json:"mediaId"`
}

type HydrationConfig struct {
	Rules []*HydrationRule `json:"rules"`
}

// HydrationRule defines a rule for attaching metadata to local files.
//
//	Example:
//		"hydration": [{
//			"mediaId": 12345,
//			"files": [
//				{
//					"pattern": "Mob Psycho - (\d+) - (.*)$",
//					"episode": "calc($1-12)",
//					"aniDbEpisode": "$1",
//					"type": "main",
//				},
//				{
//					"pattern": "Mob Psycho - NCOP(\d+)$",
//					"episode": "$1",
//					"aniDbEpisode": "$1",
//					"type": "nc",
//				}
//			]
//		}]
type HydrationRule struct {
	// Regex pattern for the path
	Pattern string `json:"pattern"`
	// The Media ID
	MediaID int `json:"mediaId"`
	// Files represents a collection of files associated with a specific hydration rule.
	Files []*HydrationFileRule `json:"files"`
}

type HydrationFileRule struct {
	Filename     string              `json:"filename"`
	IsRegex      bool                `json:"isRegex"`
	Episode      string              `json:"episode"`
	AniDbEpisode string              `json:"aniDbEpisode"`
	Type         anime.LocalFileType `json:"type,omitempty"`
}
