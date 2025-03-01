package manga

import (
	"seanime/internal/api/anilist"
	"seanime/internal/hook_resolver"
)

// MangaEntryRequestedEvent is triggered when a new media entry is being created.
type MangaEntryRequestedEvent struct {
	hook_resolver.Event
	MediaId         int                      `json:"mediaId"`
	MangaCollection *anilist.MangaCollection `json:"mangaCollection"`
}

// MangaEntryEvent is triggered when the media entry is being returned.
type MangaEntryEvent struct {
	hook_resolver.Event
	Entry *Entry `json:"entry"`
}

type MangaLibraryCollectionEvent struct {
	hook_resolver.Event
	LibraryCollection *Collection `json:"libraryCollection"`
}

type MangaLibraryCollectionRequestedEvent struct {
	hook_resolver.Event
	MangaCollection *anilist.MangaCollection `json:"mangaCollection"`
}
