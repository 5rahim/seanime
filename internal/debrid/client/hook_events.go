package debrid_client

import (
	"seanime/internal/api/anilist"
	"seanime/internal/hook_resolver"
)

// DebridSendStreamToMediaPlayerEvent is triggered when the debrid client is about to send a stream to the media player.
// Prevent default to skip the playback.
type DebridSendStreamToMediaPlayerEvent struct {
	hook_resolver.Event
	WindowTitle  string             `json:"windowTitle"`
	StreamURL    string             `json:"streamURL"`
	Media        *anilist.BaseAnime `json:"media"`
	AniDbEpisode string             `json:"aniDbEpisode"`
	PlaybackType string             `json:"playbackType"`
}

// DebridLocalDownloadRequestedEvent is triggered when Seanime is about to download a debrid torrent locally.
// Prevent default to skip the default download and override the download.
type DebridLocalDownloadRequestedEvent struct {
	hook_resolver.Event
	TorrentName string `json:"torrentName"`
	Destination string `json:"destination"`
	DownloadUrl string `json:"downloadUrl"`
}
