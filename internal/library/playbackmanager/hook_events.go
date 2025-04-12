package playbackmanager

import (
	"seanime/internal/api/anilist"
	"seanime/internal/hook_resolver"
	"seanime/internal/library/anime"
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

// PlaybackBeforeTrackingEvent is triggered just before the playback tracking starts.
// Prevent default to skip playback tracking.
type PlaybackBeforeTrackingEvent struct {
	hook_resolver.Event
	IsStream bool `json:"isStream"`
}

// PlaybackLocalFileDetailsRequestedEvent is triggered when the local files details for a specific path are requested.
// This event is triggered right after the media player loads an episode.
// The playback manager uses the local files details to track the progress, propose next episodes, etc.
// In the current implementation, the details are fetched by selecting the local file from the database and making requests to retrieve the media and anime list entry.
// Prevent default to skip the default fetching and override the details.
type PlaybackLocalFileDetailsRequestedEvent struct {
	hook_resolver.Event
	Path string `json:"path"`
	// List of all local files
	LocalFiles []*anime.LocalFile `json:"localFiles"`
	// Empty anime list entry
	AnimeListEntry *anilist.AnimeListEntry `json:"animeListEntry"`
	// Empty local file
	LocalFile *anime.LocalFile `json:"localFile"`
	// Empty local file wrapper entry
	LocalFileWrapperEntry *anime.LocalFileWrapperEntry `json:"localFileWrapperEntry"`
}

// PlaybackStreamDetailsRequestedEvent is triggered when the stream details are requested.
// Prevent default to skip the default fetching and override the details.
// In the current implementation, the details are fetched by selecting the anime from the anime collection. If nothing is found, the stream is still tracked.
type PlaybackStreamDetailsRequestedEvent struct {
	hook_resolver.Event
	AnimeCollection *anilist.AnimeCollection `json:"animeCollection"`
	MediaId         int                      `json:"mediaId"`
	// Empty anime list entry
	AnimeListEntry *anilist.AnimeListEntry `json:"animeListEntry"`
}
