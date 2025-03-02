package torrentstream

import (
	"seanime/internal/api/anilist"
	"seanime/internal/hook_resolver"
)

// TorrentStreamSendStreamToMediaPlayerEvent is triggered when the torrent stream is about to send a stream to the media player.
// Prevent default to skip the default playback and override the playback.
type TorrentStreamSendStreamToMediaPlayerEvent struct {
	hook_resolver.Event
	WindowTitle  string             `json:"windowTitle"`
	StreamURL    string             `json:"streamURL"`
	Media        *anilist.BaseAnime `json:"media"`
	AniDbEpisode string             `json:"aniDbEpisode"`
	PlaybackType string             `json:"playbackType"`
}
