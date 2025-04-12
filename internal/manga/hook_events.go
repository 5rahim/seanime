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

// MangaDownloadedChapterContainersRequestedEvent is triggered when the manga downloaded chapter containers are being requested.
// Prevent default to skip the default behavior and return the modified chapter containers.
// If the modified chapter containers are nil, an error will be returned.
type MangaDownloadedChapterContainersRequestedEvent struct {
	hook_resolver.Event
	MangaCollection *anilist.MangaCollection `json:"mangaCollection"`
	// Empty chapter containers object, will be used if the hook prevents the default behavior
	ChapterContainers []*ChapterContainer `json:"chapterContainers"`
}

// MangaDownloadedChapterContainersEvent is triggered when the manga downloaded chapter containers are being returned.
type MangaDownloadedChapterContainersEvent struct {
	hook_resolver.Event
	ChapterContainers []*ChapterContainer `json:"chapterContainers"`
}

// MangaLatestChapterNumbersMapEvent is triggered when the manga latest chapter numbers map is being returned.
type MangaLatestChapterNumbersMapEvent struct {
	hook_resolver.Event
	LatestChapterNumbersMap map[int][]MangaLatestChapterNumberItem `json:"latestChapterNumbersMap"`
}

// MangaDownloadMapEvent is triggered when the manga download map has been updated.
// This map is used to tell the client which chapters have been downloaded.
type MangaDownloadMapEvent struct {
	hook_resolver.Event
	MediaMap *MediaMap `json:"mediaMap"`
}

// MangaChapterContainerRequestedEvent is triggered when the manga chapter container is being requested.
// This event happens before the chapter container is fetched from the cache or provider.
// Prevent default to skip the default behavior and return the modified chapter container.
// If the modified chapter container is nil, an error will be returned.
type MangaChapterContainerRequestedEvent struct {
	hook_resolver.Event
	Provider string    `json:"provider"`
	MediaId  int       `json:"mediaId"`
	Titles   []*string `json:"titles"`
	Year     int       `json:"year"`
	// Empty chapter container object, will be used if the hook prevents the default behavior
	ChapterContainer *ChapterContainer `json:"chapterContainer"`
}

// MangaChapterContainerEvent is triggered when the manga chapter container is being returned.
// This event happens after the chapter container is fetched from the cache or provider.
type MangaChapterContainerEvent struct {
	hook_resolver.Event
	ChapterContainer *ChapterContainer `json:"chapterContainer"`
}
