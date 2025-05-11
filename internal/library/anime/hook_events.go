package anime

import (
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/hook_resolver"
)

/////////////////////////////
// Anime Library Events
/////////////////////////////

// AnimeEntryRequestedEvent is triggered when an anime entry is requested.
// Prevent default to skip the default behavior and return the modified entry.
// This event is triggered before [AnimeEntryEvent].
// If the modified entry is nil, an error will be returned.
type AnimeEntryRequestedEvent struct {
	hook_resolver.Event
	MediaId         int                      `json:"mediaId"`
	LocalFiles      []*LocalFile             `json:"localFiles"`
	AnimeCollection *anilist.AnimeCollection `json:"animeCollection"`
	// Empty entry object, will be used if the hook prevents the default behavior
	Entry *Entry `json:"entry"`
}

// AnimeEntryEvent is triggered when the media entry is being returned.
// This event is triggered after [AnimeEntryRequestedEvent].
type AnimeEntryEvent struct {
	hook_resolver.Event
	Entry *Entry `json:"entry"`
}

// AnimeEntryFillerHydrationEvent is triggered when the filler data is being added to the media entry.
// This event is triggered after [AnimeEntryEvent].
// Prevent default to skip the filler data.
type AnimeEntryFillerHydrationEvent struct {
	hook_resolver.Event
	Entry *Entry `json:"entry"`
}

// AnimeEntryLibraryDataRequestedEvent is triggered when the app requests the library data for a media entry.
// This is triggered before [AnimeEntryLibraryDataEvent].
type AnimeEntryLibraryDataRequestedEvent struct {
	hook_resolver.Event
	EntryLocalFiles []*LocalFile `json:"entryLocalFiles"`
	MediaId         int          `json:"mediaId"`
	CurrentProgress int          `json:"currentProgress"`
}

// AnimeEntryLibraryDataEvent is triggered when the library data is being added to the media entry.
// This is triggered after [AnimeEntryLibraryDataRequestedEvent].
type AnimeEntryLibraryDataEvent struct {
	hook_resolver.Event
	EntryLibraryData *EntryLibraryData `json:"entryLibraryData"`
}

// AnimeEntryManualMatchBeforeSaveEvent is triggered when the user manually matches local files to a media entry.
// Prevent default to skip saving the local files.
type AnimeEntryManualMatchBeforeSaveEvent struct {
	hook_resolver.Event
	// The media ID chosen by the user
	MediaId int `json:"mediaId"`
	// The paths of the local files that are being matched
	Paths []string `json:"paths"`
	// The local files that are being matched
	MatchedLocalFiles []*LocalFile `json:"matchedLocalFiles"`
}

// MissingEpisodesRequestedEvent is triggered when the user requests the missing episodes for the entire library.
// Prevent default to skip the default process and return the modified missing episodes.
type MissingEpisodesRequestedEvent struct {
	hook_resolver.Event
	AnimeCollection  *anilist.AnimeCollection `json:"animeCollection"`
	LocalFiles       []*LocalFile             `json:"localFiles"`
	SilencedMediaIds []int                    `json:"silencedMediaIds"`
	// Empty missing episodes object, will be used if the hook prevents the default behavior
	MissingEpisodes *MissingEpisodes `json:"missingEpisodes"`
}

// MissingEpisodesEvent is triggered when the missing episodes are being returned.
type MissingEpisodesEvent struct {
	hook_resolver.Event
	MissingEpisodes *MissingEpisodes `json:"missingEpisodes"`
}

/////////////////////////////
// Anime Collection Events
/////////////////////////////

// AnimeLibraryCollectionRequestedEvent is triggered when the user requests the library collection.
// Prevent default to skip the default process and return the modified library collection.
// If the modified library collection is nil, an error will be returned.
type AnimeLibraryCollectionRequestedEvent struct {
	hook_resolver.Event
	AnimeCollection *anilist.AnimeCollection `json:"animeCollection"`
	LocalFiles      []*LocalFile             `json:"localFiles"`
	// Empty library collection object, will be used if the hook prevents the default behavior
	LibraryCollection *LibraryCollection `json:"libraryCollection"`
}

// AnimeLibraryCollectionEvent is triggered when the user requests the library collection.
type AnimeLibraryCollectionEvent struct {
	hook_resolver.Event
	LibraryCollection *LibraryCollection `json:"libraryCollection"`
}

// AnimeLibraryStreamCollectionRequestedEvent is triggered when the user requests the library stream collection.
// This is called when the user enables "Include in library" for either debrid/online/torrent streamings.
type AnimeLibraryStreamCollectionRequestedEvent struct {
	hook_resolver.Event
	AnimeCollection   *anilist.AnimeCollection `json:"animeCollection"`
	LibraryCollection *LibraryCollection       `json:"libraryCollection"`
}

// AnimeLibraryStreamCollectionEvent is triggered when the library stream collection is being returned.
type AnimeLibraryStreamCollectionEvent struct {
	hook_resolver.Event
	StreamCollection *StreamCollection `json:"streamCollection"`
}

////////////////////////////////////////

// AnimeEntryDownloadInfoRequestedEvent is triggered when the app requests the download info for a media entry.
// This is triggered before [AnimeEntryDownloadInfoEvent].
type AnimeEntryDownloadInfoRequestedEvent struct {
	hook_resolver.Event
	LocalFiles    []*LocalFile `json:"localFiles"`
	AnimeMetadata *metadata.AnimeMetadata
	Media         *anilist.BaseAnime
	Progress      *int
	Status        *anilist.MediaListStatus
	// Empty download info object, will be used if the hook prevents the default behavior
	EntryDownloadInfo *EntryDownloadInfo `json:"entryDownloadInfo"`
}

// AnimeEntryDownloadInfoEvent is triggered when the download info is being returned.
type AnimeEntryDownloadInfoEvent struct {
	hook_resolver.Event
	EntryDownloadInfo *EntryDownloadInfo `json:"entryDownloadInfo"`
}

/////////////////////////////////////

// AnimeEpisodeCollectionRequestedEvent is triggered when the episode collection is being requested.
// Prevent default to skip the default behavior and return your own data.
type AnimeEpisodeCollectionRequestedEvent struct {
	hook_resolver.Event
	Media    *anilist.BaseAnime      `json:"media"`
	Metadata *metadata.AnimeMetadata `json:"metadata"`
	// Empty episode collection object, will be used if the hook prevents the default behavior
	EpisodeCollection *EpisodeCollection `json:"episodeCollection"`
}

// AnimeEpisodeCollectionEvent is triggered when the episode collection is being returned.
type AnimeEpisodeCollectionEvent struct {
	hook_resolver.Event
	EpisodeCollection *EpisodeCollection `json:"episodeCollection"`
}
