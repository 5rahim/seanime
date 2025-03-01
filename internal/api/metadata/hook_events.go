package metadata

import "seanime/internal/hook_resolver"

type AnimeMetadataRequestedEvent struct {
	hook_resolver.Event
	MediaId int `json:"mediaId"`
	// The metadata to be used, nil by default.
	OverrideAnimeMetadata *AnimeMetadata `json:"overrideAnimeMetadata"`
	// When true, the metadata will not be fetched from the provider
	Override *bool `json:"override"`
}

type AnimeMetadataEvent struct {
	hook_resolver.Event
	MediaId       int            `json:"mediaId"`
	AnimeMetadata *AnimeMetadata `json:"animeMetadata"`
}
