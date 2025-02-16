package autodownloader

import (
	"seanime/internal/api/anilist"
	"seanime/internal/database/models"
	"seanime/internal/hook_resolver"
	"seanime/internal/library/anime"
)

// AutoDownloaderQueueOrDownloadTorrentEvent is triggered when a torrent is added to the queue or downloaded
type AutoDownloaderQueueOrDownloadTorrentEvent struct {
	hook_resolver.Event
	TorrentName string `json:"torrentName"`
	MediaId     int    `json:"mediaId"`
	Episode     int    `json:"episode"`
	Link        string `json:"link"`
	Hash        string `json:"hash"`
	Magnet      string `json:"magnet"`
	Downloaded  bool   `json:"downloaded"`
}

// AutoDownloaderTorrentMatchedEvent is triggered when a torrent matches a rule
type AutoDownloaderTorrentMatchedEvent struct {
	hook_resolver.Event
	TorrentName string                    `json:"torrentName"`
	Rule        *anime.AutoDownloaderRule `json:"rule"`
	Episode     int                       `json:"episode"`
}

// AutoDownloaderRuleVerifyMatchEvent is triggered when checking if a torrent matches a rule
type AutoDownloaderRuleVerifyMatchEvent struct {
	hook_resolver.Event
	TorrentName string                       `json:"torrentName"`
	Rule        *anime.AutoDownloaderRule    `json:"rule"`
	ListEntry   *anilist.AnimeListEntry      `json:"listEntry"`
	LocalEntry  *anime.LocalFileWrapperEntry `json:"localEntry"`
}

// AutoDownloaderRunStartedEvent is triggered when the autodownloader starts checking for new episodes
type AutoDownloaderRunStartedEvent struct {
	hook_resolver.Event
	Rules []*anime.AutoDownloaderRule `json:"rules"`
}

// AutoDownloaderRunCompletedEvent is triggered when the autodownloader finishes checking for new episodes
type AutoDownloaderRunCompletedEvent struct {
	hook_resolver.Event
	TorrentsAdded int `json:"torrentsAdded"`
}

// AutoDownloaderSettingsUpdatedEvent is triggered when the autodownloader settings are updated
type AutoDownloaderSettingsUpdatedEvent struct {
	hook_resolver.Event
	Settings *models.AutoDownloaderSettings `json:"settings"`
}
