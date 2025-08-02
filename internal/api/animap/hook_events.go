package animap

import "seanime/internal/hook_resolver"

// AnimapMediaRequestedEvent is triggered when the Animap media is requested.
// Prevent default to skip the default behavior and return your own data.
type AnimapMediaRequestedEvent struct {
	hook_resolver.Event
	From string `json:"from"`
	Id   int    `json:"id"`
	// Empty data object, will be used if the hook prevents the default behavior
	Media *Anime `json:"media"`
}

// AnimapMediaEvent is triggered after processing AnimapMedia.
type AnimapMediaEvent struct {
	hook_resolver.Event
	Media *Anime `json:"media"`
}
