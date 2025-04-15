package models

import (
	"database/sql/driver"
	"errors"
	"strings"
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
	Manga          *MangaSettings          `gorm:"embedded" json:"manga"`
	Anilist        *AnilistSettings        `gorm:"embedded" json:"anilist"`
	ListSync       *ListSyncSettings       `gorm:"embedded" json:"listSync"`
	AutoDownloader *AutoDownloaderSettings `gorm:"embedded" json:"autoDownloader"`
	Discord        *DiscordSettings        `gorm:"embedded" json:"discord"`
	Notifications  *NotificationSettings   `gorm:"embedded" json:"notifications"`
}

type AnilistSettings struct {
	//AnilistClientId    string `gorm:"column:anilist_client_id" json:"anilistClientId"`
	HideAudienceScore  bool `gorm:"column:hide_audience_score" json:"hideAudienceScore"`
	EnableAdultContent bool `gorm:"column:enable_adult_content" json:"enableAdultContent"`
	BlurAdultContent   bool `gorm:"column:blur_adult_content" json:"blurAdultContent"`
}

type LibrarySettings struct {
	LibraryPath                     string `gorm:"column:library_path" json:"libraryPath"`
	AutoUpdateProgress              bool   `gorm:"column:auto_update_progress" json:"autoUpdateProgress"`
	DisableUpdateCheck              bool   `gorm:"column:disable_update_check" json:"disableUpdateCheck"`
	TorrentProvider                 string `gorm:"column:torrent_provider" json:"torrentProvider"`
	AutoScan                        bool   `gorm:"column:auto_scan" json:"autoScan"`
	EnableOnlinestream              bool   `gorm:"column:enable_onlinestream" json:"enableOnlinestream"`
	IncludeOnlineStreamingInLibrary bool   `gorm:"column:include_online_streaming_in_library" json:"includeOnlineStreamingInLibrary"`
	DisableAnimeCardTrailers        bool   `gorm:"column:disable_anime_card_trailers" json:"disableAnimeCardTrailers"`
	EnableManga                     bool   `gorm:"column:enable_manga" json:"enableManga"`
	DOHProvider                     string `gorm:"column:doh_provider" json:"dohProvider"`
	OpenTorrentClientOnStart        bool   `gorm:"column:open_torrent_client_on_start" json:"openTorrentClientOnStart"`
	OpenWebURLOnStart               bool   `gorm:"column:open_web_url_on_start" json:"openWebURLOnStart"`
	RefreshLibraryOnStart           bool   `gorm:"column:refresh_library_on_start" json:"refreshLibraryOnStart"`
	// v2.1+
	AutoPlayNextEpisode bool `gorm:"column:auto_play_next_episode" json:"autoPlayNextEpisode"`
	// v2.2+
	EnableWatchContinuity    bool         `gorm:"column:enable_watch_continuity" json:"enableWatchContinuity"`
	LibraryPaths             LibraryPaths `gorm:"column:library_paths;type:text" json:"libraryPaths"`
	AutoSyncOfflineLocalData bool         `gorm:"column:auto_sync_offline_local_data" json:"autoSyncOfflineLocalData"`
	// v2.6+
	ScannerMatchingThreshold float64 `gorm:"column:scanner_matching_threshold" json:"scannerMatchingThreshold"`
	ScannerMatchingAlgorithm string  `gorm:"column:scanner_matching_algorithm" json:"scannerMatchingAlgorithm"`
}

func (o *LibrarySettings) GetLibraryPaths() (ret []string) {
	ret = make([]string, len(o.LibraryPaths)+1)
	ret[0] = o.LibraryPath
	if len(o.LibraryPaths) > 0 {
		copy(ret[1:], o.LibraryPaths)
	}
	return
}

type LibraryPaths []string

func (o *LibraryPaths) Scan(src interface{}) error {
	str, ok := src.(string)
	if !ok {
		return errors.New("src value cannot cast to string")
	}
	*o = strings.Split(str, ",")
	return nil
}
func (o LibraryPaths) Value() (driver.Value, error) {
	if len(o) == 0 {
		return nil, nil
	}
	return strings.Join(o, ","), nil
}

type MangaSettings struct {
	DefaultProvider    string `gorm:"column:default_manga_provider" json:"defaultMangaProvider"`
	AutoUpdateProgress bool   `gorm:"column:manga_auto_update_progress" json:"mangaAutoUpdateProgress"`
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

type TorrentSettings struct {
	Default              string `gorm:"column:default_torrent_client" json:"defaultTorrentClient"`
	QBittorrentPath      string `gorm:"column:qbittorrent_path" json:"qbittorrentPath"`
	QBittorrentHost      string `gorm:"column:qbittorrent_host" json:"qbittorrentHost"`
	QBittorrentPort      int    `gorm:"column:qbittorrent_port" json:"qbittorrentPort"`
	QBittorrentUsername  string `gorm:"column:qbittorrent_username" json:"qbittorrentUsername"`
	QBittorrentPassword  string `gorm:"column:qbittorrent_password" json:"qbittorrentPassword"`
	QBittorrentTags      string `gorm:"column:qbittorrent_tags" json:"qbittorrentTags"`
	TransmissionPath     string `gorm:"column:transmission_path" json:"transmissionPath"`
	TransmissionHost     string `gorm:"column:transmission_host" json:"transmissionHost"`
	TransmissionPort     int    `gorm:"column:transmission_port" json:"transmissionPort"`
	TransmissionUsername string `gorm:"column:transmission_username" json:"transmissionUsername"`
	TransmissionPassword string `gorm:"column:transmission_password" json:"transmissionPassword"`
	// v2.1+
	ShowActiveTorrentCount bool `gorm:"column:show_active_torrent_count" json:"showActiveTorrentCount"`
	// v2.2+
	HideTorrentList bool `gorm:"column:hide_torrent_list" json:"hideTorrentList"`
}

type ListSyncSettings struct {
	Automatic bool   `gorm:"column:automatic_sync" json:"automatic"`
	Origin    string `gorm:"column:sync_origin" json:"origin"`
}

type DiscordSettings struct {
	EnableRichPresence                      bool `gorm:"column:enable_rich_presence" json:"enableRichPresence"`
	EnableAnimeRichPresence                 bool `gorm:"column:enable_anime_rich_presence" json:"enableAnimeRichPresence"`
	EnableMangaRichPresence                 bool `gorm:"column:enable_manga_rich_presence" json:"enableMangaRichPresence"`
	RichPresenceHideSeanimeRepositoryButton bool `gorm:"column:rich_presence_hide_seanime_repository_button" json:"richPresenceHideSeanimeRepositoryButton"`
	RichPresenceShowAniListMediaButton      bool `gorm:"column:rich_presence_show_anilist_media_button" json:"richPresenceShowAniListMediaButton"`
	RichPresenceShowAniListProfileButton    bool `gorm:"column:rich_presence_show_anilist_profile_button" json:"richPresenceShowAniListProfileButton"`
}

type NotificationSettings struct {
	DisableNotifications               bool `gorm:"column:disable_notifications" json:"disableNotifications"`
	DisableAutoDownloaderNotifications bool `gorm:"column:disable_auto_downloader_notifications" json:"disableAutoDownloaderNotifications"`
	DisableAutoScannerNotifications    bool `gorm:"column:disable_auto_scanner_notifications" json:"disableAutoScannerNotifications"`
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
	EnableEnhancedQueries bool   `gorm:"column:auto_downloader_enable_enhanced_queries" json:"enableEnhancedQueries"`
	EnableSeasonCheck     bool   `gorm:"column:auto_downloader_enable_season_check" json:"enableSeasonCheck"`
	UseDebrid             bool   `gorm:"column:auto_downloader_use_debrid" json:"useDebrid"`
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
	// Main
	EnableColorSettings              bool   `gorm:"column:enable_color_settings" json:"enableColorSettings"`
	BackgroundColor                  string `gorm:"column:background_color" json:"backgroundColor"`
	AccentColor                      string `gorm:"column:accent_color" json:"accentColor"`
	SidebarBackgroundColor           string `gorm:"column:sidebar_background_color" json:"sidebarBackgroundColor"`  // DEPRECATED
	AnimeEntryScreenLayout           string `gorm:"column:anime_entry_screen_layout" json:"animeEntryScreenLayout"` // DEPRECATED
	ExpandSidebarOnHover             bool   `gorm:"column:expand_sidebar_on_hover" json:"expandSidebarOnHover"`
	HideTopNavbar                    bool   `gorm:"column:hide_top_navbar" json:"hideTopNavbar"`
	EnableMediaCardBlurredBackground bool   `gorm:"column:enable_media_card_blurred_background" json:"enableMediaCardBlurredBackground"`
	// Note: These are named "libraryScreen" but are used on all pages
	LibraryScreenCustomBackgroundImage   string `gorm:"column:library_screen_custom_background_image" json:"libraryScreenCustomBackgroundImage"`
	LibraryScreenCustomBackgroundOpacity int    `gorm:"column:library_screen_custom_background_opacity" json:"libraryScreenCustomBackgroundOpacity"`
	// Anime
	SmallerEpisodeCarouselSize bool `gorm:"column:smaller_episode_carousel_size" json:"smallerEpisodeCarouselSize"`
	// Library Screen (Anime & Manga)
	// LibraryScreenBannerType: "dynamic", "custom"
	LibraryScreenBannerType           string `gorm:"column:library_screen_banner_type" json:"libraryScreenBannerType"`
	LibraryScreenCustomBannerImage    string `gorm:"column:library_screen_custom_banner_image" json:"libraryScreenCustomBannerImage"`
	LibraryScreenCustomBannerPosition string `gorm:"column:library_screen_custom_banner_position" json:"libraryScreenCustomBannerPosition"`
	LibraryScreenCustomBannerOpacity  int    `gorm:"column:library_screen_custom_banner_opacity" json:"libraryScreenCustomBannerOpacity"`
	DisableLibraryScreenGenreSelector bool   `gorm:"column:disable_library_screen_genre_selector" json:"disableLibraryScreenGenreSelector"`

	LibraryScreenCustomBackgroundBlur string `gorm:"column:library_screen_custom_background_blur" json:"libraryScreenCustomBackgroundBlur"`
	EnableMediaPageBlurredBackground  bool   `gorm:"column:enable_media_page_blurred_background" json:"enableMediaPageBlurredBackground"`
	DisableSidebarTransparency        bool   `gorm:"column:disable_sidebar_transparency" json:"disableSidebarTransparency"`
	UseLegacyEpisodeCard              bool   `gorm:"column:use_legacy_episode_card" json:"useLegacyEpisodeCard"` // DEPRECATED
	DisableCarouselAutoScroll         bool   `gorm:"column:disable_carousel_auto_scroll" json:"disableCarouselAutoScroll"`

	// v2.6+
	MediaPageBannerType        string `gorm:"column:media_page_banner_type" json:"mediaPageBannerType"`
	MediaPageBannerSize        string `gorm:"column:media_page_banner_size" json:"mediaPageBannerSize"`
	MediaPageBannerInfoBoxSize string `gorm:"column:media_page_banner_info_box_size" json:"mediaPageBannerInfoBoxSize"`

	// v2.7+
	ShowEpisodeCardAnimeInfo             bool   `gorm:"column:show_episode_card_anime_info" json:"showEpisodeCardAnimeInfo"`
	ContinueWatchingDefaultSorting       string `gorm:"column:continue_watching_default_sorting" json:"continueWatchingDefaultSorting"`
	AnimeLibraryCollectionDefaultSorting string `gorm:"column:anime_library_collection_default_sorting" json:"animeLibraryCollectionDefaultSorting"`
	MangaLibraryCollectionDefaultSorting string `gorm:"column:manga_library_collection_default_sorting" json:"mangaLibraryCollectionDefaultSorting"`
	ShowAnimeUnwatchedCount              bool   `gorm:"column:show_anime_unwatched_count" json:"showAnimeUnwatchedCount"`
	ShowMangaUnreadCount                 bool   `gorm:"column:show_manga_unread_count" json:"showMangaUnreadCount"`

	// v2.8+
	HideEpisodeCardDescription        bool `gorm:"column:hide_episode_card_description" json:"hideEpisodeCardDescription"`
	HideDownloadedEpisodeCardFilename bool `gorm:"column:hide_downloaded_episode_card_filename" json:"hideDownloadedEpisodeCardFilename"`

	// v2.9+
	CustomCSS       string `gorm:"column:custom_css" json:"customCSS"`
	MobileCustomCSS string `gorm:"column:mobile_custom_css" json:"mobileCustomCSS"`
}

// +---------------------+
// |      Playlist       |
// +---------------------+

type PlaylistEntry struct {
	BaseModel
	Name  string `gorm:"column:name" json:"name"`
	Value []byte `gorm:"column:value" json:"value"`
}

// +------------------------+
// | Chapter Download Queue |
// +------------------------+

type ChapterDownloadQueueItem struct {
	BaseModel
	Provider      string `gorm:"column:provider" json:"provider"`
	MediaID       int    `gorm:"column:media_id" json:"mediaId"`
	ChapterID     string `gorm:"column:chapter_id" json:"chapterId"`
	ChapterNumber string `gorm:"column:chapter_number" json:"chapterNumber"`
	PageData      []byte `gorm:"column:page_data" json:"pageData"` // Contains map of page index to page details
	Status        string `gorm:"column:status" json:"status"`
}

// +---------------------+
// |     MediaStream     |
// +---------------------+

type MediastreamSettings struct {
	BaseModel
	// DEVNOTE: Should really be "Enabled"
	TranscodeEnabled              bool   `gorm:"column:transcode_enabled" json:"transcodeEnabled"`
	TranscodeHwAccel              string `gorm:"column:transcode_hw_accel" json:"transcodeHwAccel"`
	TranscodeThreads              int    `gorm:"column:transcode_threads" json:"transcodeThreads"`
	TranscodePreset               string `gorm:"column:transcode_preset" json:"transcodePreset"`
	DisableAutoSwitchToDirectPlay bool   `gorm:"column:disable_auto_switch_to_direct_play" json:"disableAutoSwitchToDirectPlay"`
	DirectPlayOnly                bool   `gorm:"column:direct_play_only" json:"directPlayOnly"`
	PreTranscodeEnabled           bool   `gorm:"column:pre_transcode_enabled" json:"preTranscodeEnabled"`
	PreTranscodeLibraryDir        string `gorm:"column:pre_transcode_library_dir" json:"preTranscodeLibraryDir"`
	FfmpegPath                    string `gorm:"column:ffmpeg_path" json:"ffmpegPath"`
	FfprobePath                   string `gorm:"column:ffprobe_path" json:"ffprobePath"`
	// v2.2+
	TranscodeHwAccelCustomSettings string `gorm:"column:transcode_hw_accel_custom_settings" json:"transcodeHwAccelCustomSettings"`

	//TranscodeTempDir              string `gorm:"column:transcode_temp_dir" json:"transcodeTempDir"` // DEPRECATED
}

// +---------------------+
// |    TorrentStream    |
// +---------------------+

type TorrentstreamSettings struct {
	BaseModel
	Enabled             bool   `gorm:"column:enabled" json:"enabled"`
	AutoSelect          bool   `gorm:"column:auto_select" json:"autoSelect"`
	PreferredResolution string `gorm:"column:preferred_resolution" json:"preferredResolution"`
	DisableIPV6         bool   `gorm:"column:disable_ipv6" json:"disableIPV6"`
	DownloadDir         string `gorm:"column:download_dir" json:"downloadDir"`
	AddToLibrary        bool   `gorm:"column:add_to_library" json:"addToLibrary"`
	TorrentClientHost   string `gorm:"column:torrent_client_host" json:"torrentClientHost"`
	TorrentClientPort   int    `gorm:"column:torrent_client_port" json:"torrentClientPort"`
	StreamingServerHost string `gorm:"column:streaming_server_host" json:"streamingServerHost"`
	StreamingServerPort int    `gorm:"column:streaming_server_port" json:"streamingServerPort"`
	//FallbackToTorrentStreamingView bool   `gorm:"column:fallback_to_torrent_streaming_view" json:"fallbackToTorrentStreamingView"` // DEPRECATED
	IncludeInLibrary bool `gorm:"column:include_in_library" json:"includeInLibrary"`
	// v2.6+
	StreamUrlAddress string `gorm:"column:stream_url_address" json:"streamUrlAddress"`
	// v2.7+
	SlowSeeding bool `gorm:"column:slow_seeding" json:"slowSeeding"`
}

type TorrentstreamHistory struct {
	BaseModel
	MediaId int    `gorm:"column:media_id" json:"mediaId"`
	Torrent []byte `gorm:"column:torrent" json:"torrent"`
}

// +---------------------+
// |        Filler       |
// +---------------------+

type MediaFiller struct {
	BaseModel
	Provider      string    `gorm:"column:provider" json:"provider"`
	Slug          string    `gorm:"column:slug" json:"slug"`
	MediaID       int       `gorm:"column:media_id" json:"mediaId"`
	LastFetchedAt time.Time `gorm:"column:last_fetched_at" json:"lastFetchedAt"`
	Data          []byte    `gorm:"column:data" json:"data"`
}

// +---------------------+
// |        Manga        |
// +---------------------+

type MangaMapping struct {
	BaseModel
	Provider string `gorm:"column:provider" json:"provider"`
	MediaID  int    `gorm:"column:media_id" json:"mediaId"`
	MangaID  string `gorm:"column:manga_id" json:"mangaId"` // ID from search result, used to fetch chapters
}

type MangaChapterContainer struct {
	BaseModel
	Provider  string `gorm:"column:provider" json:"provider"`
	MediaID   int    `gorm:"column:media_id" json:"mediaId"`
	ChapterID string `gorm:"column:chapter_id" json:"chapterId"`
	Data      []byte `gorm:"column:data" json:"data"`
}

// +---------------------+
// |  Online streaming   |
// +---------------------+

type OnlinestreamMapping struct {
	BaseModel
	Provider string `gorm:"column:provider" json:"provider"`
	MediaID  int    `gorm:"column:media_id" json:"mediaId"`
	AnimeID  string `gorm:"column:anime_id" json:"anime_id"` // ID from search result, used to fetch episodes
}

// +---------------------+
// |       Debrid        |
// +---------------------+

type DebridSettings struct {
	BaseModel
	Enabled  bool   `gorm:"column:enabled" json:"enabled"`
	Provider string `gorm:"column:provider" json:"provider"`
	ApiKey   string `gorm:"column:api_key" json:"apiKey"`
	//FallbackToDebridStreamingView bool   `gorm:"column:fallback_to_debrid_streaming_view" json:"fallbackToDebridStreamingView"` // DEPRECATED
	IncludeDebridStreamInLibrary bool   `gorm:"column:include_debrid_stream_in_library" json:"includeDebridStreamInLibrary"`
	StreamAutoSelect             bool   `gorm:"column:stream_auto_select" json:"streamAutoSelect"`
	StreamPreferredResolution    string `gorm:"column:stream_preferred_resolution" json:"streamPreferredResolution"`
}

type DebridTorrentItem struct {
	BaseModel
	TorrentItemID string `gorm:"column:torrent_item_id" json:"torrentItemId"`
	Destination   string `gorm:"column:destination" json:"destination"`
	Provider      string `gorm:"column:provider" json:"provider"`
	MediaId       int    `gorm:"column:media_id" json:"mediaId"`
}

// +---------------------+
// |       Plugin        |
// +---------------------+

type PluginData struct {
	BaseModel
	PluginID string `gorm:"column:plugin_id;index" json:"pluginId"`
	Data     []byte `gorm:"column:data" json:"data"`
}
