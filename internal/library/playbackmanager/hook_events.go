package playbackmanager

import (
	"seanime/internal/api/anilist"
	"seanime/internal/hook_resolver"
)

// LocalFilePlaybackRequestedEvent is triggered when a local file is requested to be played.
// Prevent default to skip the default playback and override the playback.
type LocalFilePlaybackRequestedEvent struct {
	hook_resolver.Event
	Path string `json:"path"`
}

// StreamPlaybackRequestedEvent is triggered when a stream is requested to be played.
// Prevent default to skip the default playback and override the playback.
type StreamPlaybackRequestedEvent struct {
	hook_resolver.Event
	WindowTitle  string             `json:"windowTitle"`
	Payload      string             `json:"payload"`
	Media        *anilist.BaseAnime `json:"media"`
	AniDbEpisode string             `json:"aniDbEpisode"`
}

// PrePlaybackTrackingEvent is triggered just before the playback tracking starts.
// Prevent default to skip the default playback tracking.
type PrePlaybackTrackingEvent struct {
	hook_resolver.Event
	IsStream bool `json:"isStream"`
}
