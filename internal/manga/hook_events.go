package manga

import (
	"seanime/internal/api/anilist"
	"seanime/internal/hook_resolver"
)

// MangaEntryRequestedEvent is triggered when a manga entry is requested.
// Prevent default to skip the default behavior and return the modified entry.
// If the modified entry is nil, an error will be returned.
type MangaEntryRequestedEvent struct {
	hook_resolver.Event
	MediaId         int                      `json:"mediaId"`
	MangaCollection *anilist.MangaCollection `json:"mangaCollection"`
	// Empty entry object, will be used if the hook prevents the default behavior
	Entry *Entry `json:"entry"`
}

// MangaEntryEvent is triggered when the manga entry is being returned.
type MangaEntryEvent struct {
	hook_resolver.Event
	Entry *Entry `json:"entry"`
}

// MangaLibraryCollectionRequestedEvent is triggered when the manga library collection is being requested.
type MangaLibraryCollectionRequestedEvent struct {
	hook_resolver.Event
	MangaCollection *anilist.MangaCollection `json:"mangaCollection"`
}

// MangaLibraryCollectionEvent is triggered when the manga library collection is being returned.
type MangaLibraryCollectionEvent struct {
	hook_resolver.Event
	LibraryCollection *Collection `json:"libraryCollection"`
}
