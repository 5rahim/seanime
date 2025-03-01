package scanner

import (
	"seanime/internal/api/anilist"
	"seanime/internal/hook_resolver"
	"seanime/internal/library/anime"
)

// ScanStartedEvent is triggered when a scan operation begins
type ScanStartedEvent struct {
	hook_resolver.Event
	DirPath       string   `json:"dirPath"`
	OtherDirPaths []string `json:"otherDirPaths"`
	Enhanced      bool     `json:"enhanced"`
	SkipLocked    bool     `json:"skipLocked"`
	SkipIgnored   bool     `json:"skipIgnored"`
}

// ScanCompletedEvent is triggered when a scan operation finishes
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

// ScanMatchingStartedEvent is triggered when the matching process begins
type ScanMatchingStartedEvent struct {
	hook_resolver.Event
	LocalFiles      []*anime.LocalFile       `json:"localFiles"`
	NormalizedMedia []*anime.NormalizedMedia `json:"normalizedMedia"`
	Algorithm       string                   `json:"algorithm"`
	Threshold       float64                  `json:"threshold"`
}

// ScanLocalFileMatchedEvent is triggered when a local file is matched with media and before the match is analyzed
type ScanLocalFileMatchedEvent struct {
	hook_resolver.Event
	// Can be nil if there's no match
	Match     *anime.NormalizedMedia `json:"match"`
	Found     bool                   `json:"found"`
	LocalFile *anime.LocalFile       `json:"localFile"`
	Score     float64                `json:"score"`
	// When true, match analysis will be skipped
	// Set this to true if you plan to override the values so that the default analysis is not performed
	Override *bool `json:"override"`
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

type ScanLocalFileHydrationStartedEvent struct {
	hook_resolver.Event
	LocalFile *anime.LocalFile       `json:"localFile"`
	Media     *anime.NormalizedMedia `json:"media"`
	// When true, the default hydration will be skipped
	Override *bool `json:"override"`
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
