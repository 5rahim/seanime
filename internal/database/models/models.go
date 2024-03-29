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
	MpvSocket   string `gorm:"column:mpv_socket" json:"mpvSocket"`
	MpvPath     string `gorm:"column:mpv_path" json:"mpvPath"`
}

type LibrarySettings struct {
	LibraryPath              string `gorm:"column:library_path" json:"libraryPath"`
	AutoUpdateProgress       bool   `gorm:"column:auto_update_progress" json:"autoUpdateProgress"`
	DisableUpdateCheck       bool   `gorm:"column:disable_update_check" json:"disableUpdateCheck"`
	TorrentProvider          string `gorm:"column:torrent_provider" json:"torrentProvider"`
	AutoScan                 bool   `gorm:"column:auto_scan" json:"autoScan"`
	EnableOnlinestream       bool   `gorm:"column:enable_onlinestream" json:"enableOnlinestream"`
	DisableAnimeCardTrailers bool   `gorm:"column:disable_anime_card_trailers" json:"disableAnimeCardTrailers"`
}

type TorrentSettings struct {
	Default              string `gorm:"column:default_torrent_client" json:"defaultTorrentClient"`
	QBittorrentPath      string `gorm:"column:qbittorrent_path" json:"qbittorrentPath"`
	QBittorrentHost      string `gorm:"column:qbittorrent_host" json:"qbittorrentHost"`
	QBittorrentPort      int    `gorm:"column:qbittorrent_port" json:"qbittorrentPort"`
	QBittorrentUsername  string `gorm:"column:qbittorrent_username" json:"qbittorrentUsername"`
	QBittorrentPassword  string `gorm:"column:qbittorrent_password" json:"qbittorrentPassword"`
	TransmissionPath     string `gorm:"column:transmission_path" json:"transmissionPath"`
	TransmissionHost     string `gorm:"column:transmission_host" json:"transmissionHost"`
	TransmissionPort     int    `gorm:"column:transmission_port" json:"transmissionPort"`
	TransmissionUsername string `gorm:"column:transmission_username" json:"transmissionUsername"`
	TransmissionPassword string `gorm:"column:transmission_password" json:"transmissionPassword"`
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

// +---------------------+
// |     Media Entry     |
// +---------------------+

type SilencedMediaEntry struct {
	BaseModel
}

// +---------------------+
// |        Theme        |
// +---------------------+

type Theme struct {
	BaseModel
	AnimeEntryScreenLayout     string `gorm:"column:anime_entry_screen_layout" json:"animeEntryScreenLayout"`
	SmallerEpisodeCarouselSize bool   `gorm:"column:smaller_episode_carousel_size" json:"smallerEpisodeCarouselSize"`
	ExpandSidebarOnHover       bool   `gorm:"column:expand_sidebar_on_hover" json:"expandSidebarOnHover"`
	BackgroundColor            string `gorm:"column:background_color" json:"backgroundColor"`
	SidebarBackgroundColor     string `gorm:"column:sidebar_background_color" json:"sidebarBackgroundColor"`
	// Library Screen Banner
	LibraryScreenBannerType              string `gorm:"column:library_screen_banner_type" json:"libraryScreenBannerType"`
	LibraryScreenCustomBannerImage       string `gorm:"column:library_screen_custom_banner_image" json:"libraryScreenCustomBannerImage"`
	LibraryScreenCustomBannerPosition    string `gorm:"column:library_screen_custom_banner_position" json:"libraryScreenCustomBannerPosition"`
	LibraryScreenCustomBannerOpacity     int    `gorm:"column:library_screen_custom_banner_opacity" json:"libraryScreenCustomBannerOpacity"`
	LibraryScreenCustomBackgroundImage   string `gorm:"column:library_screen_custom_background_image" json:"libraryScreenCustomBackgroundImage"`
	LibraryScreenCustomBackgroundOpacity int    `gorm:"column:library_screen_custom_background_opacity" json:"libraryScreenCustomBackgroundOpacity"`
}

// +---------------------+
// |      Playlist       |
// +---------------------+

type PlaylistEntry struct {
	BaseModel
	Name  string `gorm:"column:name" json:"name"`
	Value []byte `gorm:"column:value" json:"value"`
}
