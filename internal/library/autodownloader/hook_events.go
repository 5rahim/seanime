package autodownloader

import (
	"seanime/internal/api/anilist"
	"seanime/internal/database/models"
	"seanime/internal/hook_resolver"
	"seanime/internal/library/anime"
)

// AutoDownloaderRunStartedEvent is triggered when the autodownloader starts checking for new episodes
type AutoDownloaderRunStartedEvent struct {
	hook_resolver.Event
	Rules []*anime.AutoDownloaderRule `json:"rules"`
}

// AutoDownloaderTorrentsFetchedEvent is triggered when the autodownloader fetches torrents from the provider
type AutoDownloaderTorrentsFetchedEvent struct {
	hook_resolver.Event
	Torrents []*NormalizedTorrent `json:"torrents"`
}

// AutoDownloaderMatchVerifiedEvent is triggered when a torrent is verified to follow a rule
type AutoDownloaderMatchVerifiedEvent struct {
	hook_resolver.Event
	// Fetched torrent
	Torrent    *NormalizedTorrent           `json:"torrent"`
	Rule       *anime.AutoDownloaderRule    `json:"rule"`
	ListEntry  *anilist.AnimeListEntry      `json:"listEntry"`
	LocalEntry *anime.LocalFileWrapperEntry `json:"localEntry"`
	// The episode number found for the match
	// If the match failed, this will be 0
	Episode int `json:"episode"`
	// Whether the torrent matches the rule
	// Changing this value to true will trigger a download even if the match failed;
	Ok bool `json:"ok"`
}

// AutoDownloaderSettingsUpdatedEvent is triggered when the autodownloader settings are updated
type AutoDownloaderSettingsUpdatedEvent struct {
	hook_resolver.Event
	Settings *models.AutoDownloaderSettings `json:"settings"`
}
