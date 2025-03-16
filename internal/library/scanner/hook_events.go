package scanner

import (
	"seanime/internal/api/anilist"
	"seanime/internal/hook_resolver"
	"seanime/internal/library/anime"
)

// ScanStartedEvent is triggered when the scanning process begins
type ScanStartedEvent struct {
	hook_resolver.Event
	DirPath       string   `json:"dirPath"`
	OtherDirPaths []string `json:"otherDirPaths"`
	Enhanced      bool     `json:"enhanced"`
	SkipLocked    bool     `json:"skipLocked"`
	SkipIgnored   bool     `json:"skipIgnored"`
}

// ScanFilePathsRetrievedEvent is triggered when the file paths to scan are retrieved
type ScanFilePathsRetrievedEvent struct {
	hook_resolver.Event
	FilePaths []string `json:"filePaths"`
}

// ScanLocalFilesParsedEvent is triggered when the file paths are parsed into local file objects
type ScanLocalFilesParsedEvent struct {
	hook_resolver.Event
	LocalFiles []*anime.LocalFile `json:"localFiles"`
}

// ScanCompletedEvent is triggered when the scanning process finishes
type ScanCompletedEvent struct {
	hook_resolver.Event
	LocalFiles []*anime.LocalFile `json:"localFiles"`
	Duration   int                `json:"duration"` // in milliseconds
}

// ScanMediaFetcherStartedEvent is triggered when the media fetcher begins
type ScanMediaFetcherStartedEvent struct {
	hook_resolver.Event
	Enhanced bool `json:"enhanced"`
}

// ScanMediaFetcherCompletedEvent is triggered when the media fetcher completes
type ScanMediaFetcherCompletedEvent struct {
	hook_resolver.Event
	AllMedia        []*anilist.CompleteAnime `json:"allMedia"`
	UnknownMediaIds []int                    `json:"unknownMediaIds"`
}

// ScanMatchingStartedEvent is triggered when the matching process begins.
// Prevent default to skip the default matching and override the matching.
type ScanMatchingStartedEvent struct {
	hook_resolver.Event
	LocalFiles      []*anime.LocalFile       `json:"localFiles"`
	NormalizedMedia []*anime.NormalizedMedia `json:"normalizedMedia"`
	Algorithm       string                   `json:"algorithm"`
	Threshold       float64                  `json:"threshold"`
}

// ScanLocalFileMatchedEvent is triggered when a local file is matched with media and before the match is analyzed.
// Prevent default to skip the default analysis and override the match.
type ScanLocalFileMatchedEvent struct {
	hook_resolver.Event
	// Can be nil if there's no match
	Match     *anime.NormalizedMedia `json:"match"`
	Found     bool                   `json:"found"`
	LocalFile *anime.LocalFile       `json:"localFile"`
	Score     float64                `json:"score"`
}

// ScanMatchingCompletedEvent is triggered when the matching process completes
type ScanMatchingCompletedEvent struct {
	hook_resolver.Event
	LocalFiles []*anime.LocalFile `json:"localFiles"`
}

// ScanHydrationStartedEvent is triggered when the file hydration process begins
type ScanHydrationStartedEvent struct {
	hook_resolver.Event
	LocalFiles []*anime.LocalFile       `json:"localFiles"`
	AllMedia   []*anime.NormalizedMedia `json:"allMedia"`
}

// ScanLocalFileHydrationStartedEvent is triggered when a local file's metadata is about to be hydrated.
// Prevent default to skip the default hydration and override the hydration.
type ScanLocalFileHydrationStartedEvent struct {
	hook_resolver.Event
	LocalFile *anime.LocalFile       `json:"localFile"`
	Media     *anime.NormalizedMedia `json:"media"`
}

// ScanLocalFileHydratedEvent is triggered when a local file's metadata is hydrated
type ScanLocalFileHydratedEvent struct {
	hook_resolver.Event
	LocalFile *anime.LocalFile `json:"localFile"`
	MediaId   int              `json:"mediaId"`
	Episode   int              `json:"episode"`
}

// ScanHydrationCompletedEvent is triggered when the file hydration process completes
type ScanHydrationCompletedEvent struct {
	hook_resolver.Event
	LocalFiles []*anime.LocalFile `json:"localFiles"`
}
