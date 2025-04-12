package scanner

import (
	"seanime/internal/api/anilist"
	"seanime/internal/hook_resolver"
	"seanime/internal/library/anime"
)

// ScanStartedEvent is triggered when the scanning process begins.
// Prevent default to skip the rest of the scanning process and return the local files.
type ScanStartedEvent struct {
	hook_resolver.Event
	// The main directory to scan
	LibraryPath string `json:"libraryPath"`
	// Other directories to scan
	OtherLibraryPaths []string `json:"otherLibraryPaths"`
	// Whether to use enhanced scanning,
	// Enhanced scanning will fetch media from AniList based on the local files' titles,
	// and use the metadata to match the local files.
	Enhanced bool `json:"enhanced"`
	// Whether to skip locked files
	SkipLocked bool `json:"skipLocked"`
	// Whether to skip ignored files
	SkipIgnored bool `json:"skipIgnored"`
	// All previously scanned local files
	LocalFiles []*anime.LocalFile `json:"localFiles"`
}

// ScanFilePathsRetrievedEvent is triggered when the file paths to scan are retrieved.
// The event includes file paths from all directories to scan.
// The event includes file paths of local files that will be skipped.
type ScanFilePathsRetrievedEvent struct {
	hook_resolver.Event
	FilePaths []string `json:"filePaths"`
}

// ScanLocalFilesParsedEvent is triggered right after the file paths are parsed into local file objects.
// The event does not include local files that are skipped.
type ScanLocalFilesParsedEvent struct {
	hook_resolver.Event
	LocalFiles []*anime.LocalFile `json:"localFiles"`
}

// ScanCompletedEvent is triggered when the scanning process finishes.
// The event includes all the local files (skipped and scanned) to be inserted as a new entry.
// Right after this event, the local files will be inserted as a new entry.
type ScanCompletedEvent struct {
	hook_resolver.Event
	LocalFiles []*anime.LocalFile `json:"localFiles"`
	Duration   int                `json:"duration"` // in milliseconds
}

// ScanMediaFetcherStartedEvent is triggered right before Seanime starts fetching media to be matched against the local files.
type ScanMediaFetcherStartedEvent struct {
	hook_resolver.Event
	// Whether to use enhanced scanning.
	// Enhanced scanning will fetch media from AniList based on the local files' titles,
	// and use the metadata to match the local files.
	Enhanced bool `json:"enhanced"`
}

// ScanMediaFetcherCompletedEvent is triggered when the media fetcher completes.
// The event includes all the media fetched from AniList.
// The event includes the media IDs that are not in the user's collection.
type ScanMediaFetcherCompletedEvent struct {
	hook_resolver.Event
	// All media fetched from AniList, to be matched against the local files.
	AllMedia []*anilist.CompleteAnime `json:"allMedia"`
	// Media IDs that are not in the user's collection.
	UnknownMediaIds []int `json:"unknownMediaIds"`
}

// ScanMatchingStartedEvent is triggered when the matching process begins.
// Prevent default to skip the default matching, in which case modified local files will be used.
type ScanMatchingStartedEvent struct {
	hook_resolver.Event
	// Local files to be matched.
	// If default is prevented, these local files will be used.
	LocalFiles []*anime.LocalFile `json:"localFiles"`
	// Media to be matched against the local files.
	NormalizedMedia []*anime.NormalizedMedia `json:"normalizedMedia"`
	// Matching algorithm.
	Algorithm string `json:"algorithm"`
	// Matching threshold.
	Threshold float64 `json:"threshold"`
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

// ScanMatchingCompletedEvent is triggered when the matching process completes.
type ScanMatchingCompletedEvent struct {
	hook_resolver.Event
	LocalFiles []*anime.LocalFile `json:"localFiles"`
}

// ScanHydrationStartedEvent is triggered when the file hydration process begins.
// Prevent default to skip the rest of the hydration process, in which case the event's local files will be used.
type ScanHydrationStartedEvent struct {
	hook_resolver.Event
	// Local files to be hydrated.
	LocalFiles []*anime.LocalFile `json:"localFiles"`
	// Media to be hydrated.
	AllMedia []*anime.NormalizedMedia `json:"allMedia"`
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
