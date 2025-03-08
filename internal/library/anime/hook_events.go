package anime

import (
	"seanime/internal/api/anilist"
	"seanime/internal/hook_resolver"
)

/////////////////////////////
// Anime Library Events
/////////////////////////////

// AnimeEntryRequestedEvent is triggered when a new media entry is being created.
type AnimeEntryRequestedEvent struct {
	hook_resolver.Event
	MediaId         int                      `json:"mediaId"`
	LocalFiles      []*LocalFile             `json:"localFiles"`
	AnimeCollection *anilist.AnimeCollection `json:"animeCollection"`
}

// AnimeEntryEvent is triggered when the media entry is being returned.
type AnimeEntryEvent struct {
	hook_resolver.Event
	Entry *Entry `json:"entry"`
}

// AnimeEntryFillerHydrationEvent is triggered when the filler data is being added to the media entry.
// Prevent default to avoid adding the filler data.
type AnimeEntryFillerHydrationEvent struct {
	hook_resolver.Event
	Entry *Entry `json:"entry"`
}

// AnimeEntryLibraryDataRequestedEvent is triggered when the app requests the library data for a media entry.
type AnimeEntryLibraryDataRequestedEvent struct {
	hook_resolver.Event
	EntryLocalFiles []*LocalFile `json:"entryLocalFiles"`
	MediaId         int          `json:"mediaId"`
	CurrentProgress int          `json:"currentProgress"`
}

// AnimeEntryLibraryDataEvent is triggered when the library data is being added to the media entry.
type AnimeEntryLibraryDataEvent struct {
	hook_resolver.Event
	EntryLibraryData *EntryLibraryData `json:"entryLibraryData"`
}

// AnimeEntryManualMatchBeforeSaveEvent is triggered when the user manually matches local files to a media entry.
type AnimeEntryManualMatchBeforeSaveEvent struct {
	hook_resolver.Event
	// The media ID chosen by the user
	MediaId int `json:"mediaId"`
	// The paths of the local files that are being matched
	Paths []string `json:"paths"`
	// The local files that are being matched
	MatchedLocalFiles []*LocalFile `json:"matchedLocalFiles"`
}

// MissingEpisodesRequestedEvent is triggered when the user requests the missing episodes for a media entry.
type MissingEpisodesRequestedEvent struct {
	hook_resolver.Event
	AnimeCollection  *anilist.AnimeCollection `json:"animeCollection"`
	LocalFiles       []*LocalFile             `json:"localFiles"`
	SilencedMediaIds []int                    `json:"silencedMediaIds"`
}

type MissingEpisodesEvent struct {
	hook_resolver.Event
	MissingEpisodes *MissingEpisodes `json:"missingEpisodes"`
}

/////////////////////////////
// Anime Collection Events
/////////////////////////////

type AnimeLibraryCollectionEvent struct {
	hook_resolver.Event
	LibraryCollection *LibraryCollection `json:"libraryCollection"`
}

type AnimeLibraryStreamCollectionEvent struct {
	hook_resolver.Event
	StreamCollection *StreamCollection `json:"streamCollection"`
}

type AnimeLibraryCollectionRequestedEvent struct {
	hook_resolver.Event
	AnimeCollection *anilist.AnimeCollection `json:"animeCollection"`
	LocalFiles      []*LocalFile             `json:"localFiles"`
}

type AnimeLibraryStreamCollectionRequestedEvent struct {
	hook_resolver.Event
	AnimeCollection   *anilist.AnimeCollection `json:"animeCollection"`
	LibraryCollection *LibraryCollection       `json:"libraryCollection"`
}
