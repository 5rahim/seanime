package mediaplayer

import (
	"seanime/internal/hook_resolver"
)

// MediaPlayerLocalFileTrackingRequestedEvent is triggered when the playback manager wants to track the progress of a local file.
// Prevent default to stop tracking.
type MediaPlayerLocalFileTrackingRequestedEvent struct {
	hook_resolver.Event
	// StartRefreshDelay is the number of seconds to wait before attempting to get the status
	StartRefreshDelay int `json:"startRefreshDelay"`
	// RefreshDelay is the number of seconds to wait before we refresh the status of the player after getting it for the first time
	RefreshDelay int `json:"refreshDelay"`
	// MaxRetries is the maximum number of retries
	MaxRetries int `json:"maxRetries"`
}

// MediaPlayerStreamTrackingRequestedEvent is triggered when the playback manager wants to track the progress of a stream.
// Prevent default to stop tracking.
type MediaPlayerStreamTrackingRequestedEvent struct {
	hook_resolver.Event
	// StartRefreshDelay is the number of seconds to wait before attempting to get the status
	StartRefreshDelay int `json:"startRefreshDelay"`
	// RefreshDelay is the number of seconds to wait before we refresh the status of the player after getting it for the first time
	RefreshDelay int `json:"refreshDelay"`
	// MaxRetries is the maximum number of retries
	MaxRetries int `json:"maxRetries"`
	// MaxRetriesAfterStart is the maximum number of retries after the player has started
	MaxRetriesAfterStart int `json:"maxRetriesAfterStart"`
}
