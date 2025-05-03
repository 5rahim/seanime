package anizip

import "seanime/internal/hook_resolver"

// AnizipMediaRequestedEvent is triggered when the AniZip media is requested.
// Prevent default to skip the default behavior and return your own data.
type AnizipMediaRequestedEvent struct {
	hook_resolver.Event
	From string `json:"from"`
	Id   int    `json:"id"`
	// Empty data object, will be used if the hook prevents the default behavior
	Media *Media `json:"media"`
}

// AnizipMediaEvent is triggered after processing AnizipMedia.
type AnizipMediaEvent struct {
	hook_resolver.Event
	Media *Media `json:"media"`
}
