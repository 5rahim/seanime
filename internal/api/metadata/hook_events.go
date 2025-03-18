package metadata

import "seanime/internal/hook_resolver"

// AnimeMetadataRequestedEvent is triggered when anime metadata is requested.
// Prevent default to skip the default behavior and return the overridden metadata.
type AnimeMetadataRequestedEvent struct {
	hook_resolver.Event
	MediaId int `json:"mediaId"`
	// Empty metadata object, will be used if the hook prevents the default behavior
	AnimeMetadata *AnimeMetadata `json:"animeMetadata"`
}

// AnimeMetadataEvent is triggered when anime metadata is available.
type AnimeMetadataEvent struct {
	hook_resolver.Event
	MediaId       int            `json:"mediaId"`
	AnimeMetadata *AnimeMetadata `json:"animeMetadata"`
}

// AnimeEpisodeMetadataRequestedEvent is triggered when anime episode metadata is requested.
// Prevent default to skip the default behavior and return the overridden metadata.
type AnimeEpisodeMetadataRequestedEvent struct {
	hook_resolver.Event
	// Empty metadata object, will be used if the hook prevents the default behavior
	EpisodeMetadata *EpisodeMetadata `json:"animeEpisodeMetadata"`
	EpisodeNumber   int              `json:"episodeNumber"`
	MediaId         int              `json:"mediaId"`
}

// AnimeEpisodeMetadataEvent is triggered when anime episode metadata is available and is about to be returned.
type AnimeEpisodeMetadataEvent struct {
	hook_resolver.Event
	EpisodeMetadata *EpisodeMetadata `json:"animeEpisodeMetadata"`
	EpisodeNumber   int              `json:"episodeNumber"`
	MediaId         int              `json:"mediaId"`
}
