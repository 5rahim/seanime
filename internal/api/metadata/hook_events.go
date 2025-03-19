package metadata

import "seanime/internal/hook_resolver"

// AnimeMetadataRequestedEvent is triggered when anime metadata is requested and right before the metadata is processed.
// This event is followed by [AnimeMetadataEvent] which is triggered when the metadata is available.
// Prevent default to skip the default behavior and return the modified metadata.
// If the modified metadata is nil, an error will be returned.
type AnimeMetadataRequestedEvent struct {
	hook_resolver.Event
	MediaId int `json:"mediaId"`
	// Empty metadata object, will be used if the hook prevents the default behavior
	AnimeMetadata *AnimeMetadata `json:"animeMetadata"`
}

// AnimeMetadataEvent is triggered when anime metadata is available and is about to be returned.
// Anime metadata can be requested in many places, ranging from displaying the anime entry to starting a torrent stream.
// This event is triggered after [AnimeMetadataRequestedEvent].
// If the modified metadata is nil, an error will be returned.
type AnimeMetadataEvent struct {
	hook_resolver.Event
	MediaId       int            `json:"mediaId"`
	AnimeMetadata *AnimeMetadata `json:"animeMetadata"`
}

// AnimeEpisodeMetadataRequestedEvent is triggered when anime episode metadata is requested.
// Prevent default to skip the default behavior and return the overridden metadata.
// This event is triggered before [AnimeEpisodeMetadataEvent].
// If the modified episode metadata is nil, an empty EpisodeMetadata object will be returned.
type AnimeEpisodeMetadataRequestedEvent struct {
	hook_resolver.Event
	// Empty metadata object, will be used if the hook prevents the default behavior
	EpisodeMetadata *EpisodeMetadata `json:"animeEpisodeMetadata"`
	EpisodeNumber   int              `json:"episodeNumber"`
	MediaId         int              `json:"mediaId"`
}

// AnimeEpisodeMetadataEvent is triggered when anime episode metadata is available and is about to be returned.
// In the current implementation, episode metadata is requested for display purposes. It is used to get a more complete metadata object since the original AnimeMetadata object is not complete.
// This event is triggered after [AnimeEpisodeMetadataRequestedEvent].
// If the modified episode metadata is nil, an empty EpisodeMetadata object will be returned.
type AnimeEpisodeMetadataEvent struct {
	hook_resolver.Event
	EpisodeMetadata *EpisodeMetadata `json:"animeEpisodeMetadata"`
	EpisodeNumber   int              `json:"episodeNumber"`
	MediaId         int              `json:"mediaId"`
}
