package debrid_client

import (
	"seanime/internal/api/anilist"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/hook_resolver"
)

// DebridAutoSelectTorrentsFetchedEvent is triggered when the torrents are fetched for auto select.
// The torrents are sorted by seeders from highest to lowest.
// This event is triggered before the top 3 torrents are analyzed.
type DebridAutoSelectTorrentsFetchedEvent struct {
	hook_resolver.Event
	Torrents []*hibiketorrent.AnimeTorrent
}

// DebridSkipStreamCheckEvent is triggered when the debrid client is about to skip the stream check.
// Prevent default to enable the stream check.
type DebridSkipStreamCheckEvent struct {
	hook_resolver.Event
	StreamURL  string `json:"streamURL"`
	Retries    int    `json:"retries"`
	RetryDelay int    `json:"retryDelay"` // in seconds
}

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
