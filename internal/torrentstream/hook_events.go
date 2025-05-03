package torrentstream

import (
	"seanime/internal/api/anilist"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/hook_resolver"
)

// TorrentStreamAutoSelectTorrentsFetchedEvent is triggered when the torrents are fetched for auto select.
// The torrents are sorted by seeders from highest to lowest.
// This event is triggered before the top 3 torrents are analyzed.
type TorrentStreamAutoSelectTorrentsFetchedEvent struct {
	hook_resolver.Event
	Torrents []*hibiketorrent.AnimeTorrent
}

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

// TorrentStreamEpisodeCollectionRequestedEvent is triggered when the episode collection is being requested.
// Prevent default to skip the default behavior and return your own data.
type TorrentStreamEpisodeCollectionRequestedEvent struct {
	hook_resolver.Event
	MediaId int `json:"mediaId"`
	// Empty episode collection object, will be used if the hook prevents the default behavior
	EpisodeCollection *EpisodeCollection `json:"episodeCollection"`
}

// TorrentStreamEpisodeCollectionEvent is triggered when the episode collection is being returned.
type TorrentStreamEpisodeCollectionEvent struct {
	hook_resolver.Event
	EpisodeCollection *EpisodeCollection `json:"episodeCollection"`
}
