package mediaplayer

import (
	"seanime/internal/hook_resolver"
)

// MediaPlayerLocalFileTrackingRequestedEvent is triggered when the playback manager wants to track the progress of a local file.
// Prevent default to stop tracking.
type MediaPlayerLocalFileTrackingRequestedEvent struct {
	hook_resolver.Event
	RefreshDelay int `json:"refreshDelay"` // Refresh the status of the player each x seconds
	MaxRetries   int `json:"maxRetries"`   // Maximum number of retries
}

// MediaPlayerStreamTrackingRequestedEvent is triggered when the playback manager wants to track the progress of a stream.
// Prevent default to stop tracking.
type MediaPlayerStreamTrackingRequestedEvent struct {
	hook_resolver.Event
	RefreshDelay         int `json:"refreshDelay"`         // Refresh the status of the player each x seconds
	MaxRetries           int `json:"maxRetries"`           // Maximum number of retries
	MaxRetriesAfterStart int `json:"maxRetriesAfterStart"` // Maximum number of retries after the player has started
}
