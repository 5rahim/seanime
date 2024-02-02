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
	Library     *LibrarySettings     `gorm:"embedded" json:"library"`
	MediaPlayer *MediaPlayerSettings `gorm:"embedded" json:"mediaPlayer"`
	Torrent     *TorrentSettings     `gorm:"embedded" json:"torrent"`
	Anilist     *AnilistSettings     `gorm:"embedded" json:"anilist"`
	ListSync    *ListSyncSettings    `gorm:"embedded" json:"listSync"`
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
	LibraryPath string `gorm:"column:library_path" json:"libraryPath"`
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
