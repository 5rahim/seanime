package metadata

import "seanime/internal/hook_resolver"

// AnimeMetadataRequestedEvent is triggered when anime metadata is requested.
// Prevent default to skip the default behavior and return the overridden metadata.
type AnimeMetadataRequestedEvent struct {
	hook_resolver.Event
	MediaId int `json:"mediaId"`
	// Empty metadata object, will be used if the hook prevents the default behavior
	AnimeMetadata *AnimeMetadata `json:"overrideAnimeMetadata"`
}

// AnimeMetadataEvent is triggered when anime metadata is available.
type AnimeMetadataEvent struct {
	hook_resolver.Event
	MediaId       int            `json:"mediaId"`
	AnimeMetadata *AnimeMetadata `json:"animeMetadata"`
}
