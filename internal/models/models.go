package models

import (
	"time"
)

type BaseModel struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Token struct {
	BaseModel
	Value string `json:"value"`
}

type Account struct {
	BaseModel
	Username string `gorm:"column:username" json:"username"`
	Token    string `gorm:"column:token" json:"token"`
	Viewer   []byte `gorm:"column:viewer" json:"viewer"`
}

// +---------------------+
// |     LocalFiles      |
// +---------------------+

type LocalFiles struct {
	BaseModel
	Value []byte `gorm:"column:value" json:"value"`
}

// +---------------------+
// |       Settings      |
// +---------------------+

type Settings struct {
	BaseModel
	Library        *LibrarySettings        `gorm:"embedded" json:"library"`
	MediaPlayer    *MediaPlayerSettings    `gorm:"embedded" json:"mediaPlayer"`
	Torrent        *TorrentSettings        `gorm:"embedded" json:"torrent"`
	Anilist        *AnilistSettings        `gorm:"embedded" json:"anilist"`
	ListSync       *ListSyncSettings       `gorm:"embedded" json:"listSync"`
	AutoDownloader *AutoDownloaderSettings `gorm:"embedded" json:"autoDownloader"`
}

type AnilistSettings struct {
	HideAudienceScore bool `gorm:"column:hide_audience_score" json:"hideAudienceScore"`
}

type MediaPlayerSettings struct {
	Default     string `gorm:"column:default_player" json:"defaultPlayer"` // "vlc" or "mpc-hc"
	Host        string `gorm:"column:player_host" json:"host"`
	VlcUsername string `gorm:"column:vlc_username" json:"vlcUsername"`
	VlcPassword string `gorm:"column:vlc_password" json:"vlcPassword"`
	VlcPort     int    `gorm:"column:vlc_port" json:"vlcPort"`
	VlcPath     string `gorm:"column:vlc_path" json:"vlcPath"`
	MpcPort     int    `gorm:"column:mpc_port" json:"mpcPort"`
	MpcPath     string `gorm:"column:mpc_path" json:"mpcPath"`
}

type LibrarySettings struct {
	LibraryPath        string `gorm:"column:library_path" json:"libraryPath"`
	AutoUpdateProgress bool   `gorm:"column:auto_update_progress" json:"autoUpdateProgress"`
}

type TorrentSettings struct {
	QBittorrentPath     string `gorm:"column:qbittorrent_path" json:"qbittorrentPath"`
	QBittorrentHost     string `gorm:"column:qbittorrent_host" json:"qbittorrentHost"`
	QBittorrentPort     int    `gorm:"column:qbittorrent_port" json:"qbittorrentPort"`
	QBittorrentUsername string `gorm:"column:qbittorrent_username" json:"qbittorrentUsername"`
	QBittorrentPassword string `gorm:"column:qbittorrent_password" json:"qbittorrentPassword"`
}

type ListSyncSettings struct {
	Automatic bool   `gorm:"column:automatic_sync" json:"automatic"`
	Origin    string `gorm:"column:sync_origin" json:"origin"`
}

// +---------------------+
// |         MAL         |
// +---------------------+

type Mal struct {
	BaseModel
	Username       string    `gorm:"column:username" json:"username"`
	AccessToken    string    `gorm:"column:access_token" json:"accessToken"`
	RefreshToken   string    `gorm:"column:refresh_token" json:"refreshToken"`
	TokenExpiresAt time.Time `gorm:"column:token_expires_at" json:"tokenExpiresAt"`
}

// +---------------------+
// |    Scan Summary     |
// +---------------------+

type ScanSummary struct {
	BaseModel
	Value []byte `gorm:"column:value" json:"value"`
}

// +---------------------+
// |   Auto downloader   |
// +---------------------+

type AutoDownloaderRule struct {
	BaseModel
	Value []byte `gorm:"column:value" json:"value"`
}

type AutoDownloaderItem struct {
	BaseModel
	RuleID      uint   `gorm:"column:rule_id" json:"ruleId"`
	MediaID     int    `gorm:"column:media_id" json:"mediaId"`
	Episode     int    `gorm:"column:episode" json:"episode"`
	Link        string `gorm:"column:link" json:"link"`
	Hash        string `gorm:"column:hash" json:"hash"`
	Magnet      string `gorm:"column:magnet" json:"magnet"`
	TorrentName string `gorm:"column:torrent_name" json:"torrentName"`
	Downloaded  bool   `gorm:"column:downloaded" json:"downloaded"`
}

type AutoDownloaderSettings struct {
	Provider              string `gorm:"column:auto_downloader_provider" json:"provider"`
	Interval              int    `gorm:"column:auto_downloader_interval" json:"interval"`
	Enabled               bool   `gorm:"column:auto_downloader_enabled" json:"enabled"`
	DownloadAutomatically bool   `gorm:"column:auto_downloader_download_automatically" json:"downloadAutomatically"`
}
