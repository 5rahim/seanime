package qbittorrent_model

type Preferences struct {
	// Currently selected language (e.g. en_GB for English)
	Locale string `json:"locale"`
	// True if a subfolder should be created when adding a torrent
	CreateSubfolderEnabled bool `json:"create_subfolder_enabled"`
	// True if torrents should be added in a Paused state
	StartPausedEnabled bool `json:"start_paused_enabled"`
	// No documentation provided
	AutoDeleteMode int `json:"auto_delete_mode"`
	// True if disk space should be pre-allocated for all files
	PreallocateAll bool `json:"preallocate_all"`
	// True if ".!qB" should be appended to incomplete files
	IncompleteFilesExt bool `json:"incomplete_files_ext"`
	// True if Automatic Torrent Management is enabled by default
	AutoTmmEnabled bool `json:"auto_tmm_enabled"`
	// True if torrent should be relocated when its Category changes
	TorrentChangedTmmEnabled bool `json:"torrent_changed_tmm_enabled"`
	// True if torrent should be relocated when the default save path changes
	SavePathChangedTmmEnabled bool `json:"save_path_changed_tmm_enabled"`
	// True if torrent should be relocated when its Category's save path changes
	CategoryChangedTmmEnabled bool `json:"category_changed_tmm_enabled"`
	// Default save path for torrents, separated by slashes
	SavePath string `json:"save_path"`
	// True if folder for incomplete torrents is enabled
	TempPathEnabled bool `json:"temp_path_enabled"`
	// Path for incomplete torrents, separated by slashes
	TempPath string `json:"temp_path"`
	// Property: directory to watch for torrent files, value: where torrents loaded from this directory should be downloaded to (see list of possible values below). Slashes are used as path separators; multiple key/value pairs can be specified
	ScanDirs map[string]interface{} `json:"scan_dirs"`
	// Path to directory to copy .torrent files to. Slashes are used as path separators
	ExportDir string `json:"export_dir"`
	// Path to directory to copy .torrent files of completed downloads to. Slashes are used as path separators
	ExportDirFin string `json:"export_dir_fin"`
	// True if e-mail notification should be enabled
	MailNotificationEnabled bool `json:"mail_notification_enabled"`
	// e-mail where notifications should originate from
	MailNotificationSender string `json:"mail_notification_sender"`
	// e-mail to send notifications to
	MailNotificationEmail string `json:"mail_notification_email"`
	// smtp server for e-mail notifications
	MailNotificationSmtp string `json:"mail_notification_smtp"`
	// True if smtp server requires SSL connection
	MailNotificationSslEnabled bool `json:"mail_notification_ssl_enabled"`
	// True if smtp server requires authentication
	MailNotificationAuthEnabled bool `json:"mail_notification_auth_enabled"`
	// Username for smtp authentication
	MailNotificationUsername string `json:"mail_notification_username"`
	// Password for smtp authentication
	MailNotificationPassword string `json:"mail_notification_password"`
	// True if external program should be run after torrent has finished downloading
	AutorunEnabled bool `json:"autorun_enabled"`
	// Program path/name/arguments to run if autorun_enabled is enabled; path is separated by slashes; you can use %f and %n arguments, which will be expanded by qBittorent as path_to_torrent_file and torrent_name (from the GUI; not the .torrent file name) respectively
	AutorunProgram string `json:"autorun_program"`
	// True if torrent queuing is enabled
	QueueingEnabled bool `json:"queueing_enabled"`
	// Maximum number of active simultaneous downloads
	MaxActiveDownloads int `json:"max_active_downloads"`
	// Maximum number of active simultaneous downloads and uploads
	MaxActiveTorrents int `json:"max_active_torrents"`
	// Maximum number of active simultaneous uploads
	MaxActiveUploads int `json:"max_active_uploads"`
	// If true torrents w/o any activity (stalled ones) will not be counted towards max_active_* limits; see dont_count_slow_torrents for more information
	DontCountSlowTorrents bool `json:"dont_count_slow_torrents"`
	// Download rate in KiB/s for a torrent to be considered "slow"
	SlowTorrentDlRateThreshold int `json:"slow_torrent_dl_rate_threshold"`
	// Upload rate in KiB/s for a torrent to be considered "slow"
	SlowTorrentUlRateThreshold int `json:"slow_torrent_ul_rate_threshold"`
	// Seconds a torrent should be inactive before considered "slow"
	SlowTorrentInactiveTimer int `json:"slow_torrent_inactive_timer"`
	// True if share ratio limit is enabled
	MaxRatioEnabled bool `json:"max_ratio_enabled"`
	// Get the global share ratio limit
	MaxRatio float64 `json:"max_ratio"`
	// Action performed when a torrent reaches the maximum share ratio. See list of possible values here below.
	MaxRatioAct MaxRatioAction `json:"max_ratio_act"`
	// Port for incoming connections
	ListenPort int `json:"listen_port"`
	// True if UPnP/NAT-PMP is enabled
	Upnp bool `json:"upnp"`
	// True if the port is randomly selected
	RandomPort bool `json:"random_port"`
	// Global download speed limit in KiB/s; -1 means no limit is applied
	DlLimit int `json:"dl_limit"`
	// Global upload speed limit in KiB/s; -1 means no limit is applied
	UpLimit int `json:"up_limit"`
	// Maximum global number of simultaneous connections
	MaxConnec int `json:"max_connec"`
	// Maximum number of simultaneous connections per torrent
	MaxConnecPerTorrent int `json:"max_connec_per_torrent"`
	// Maximum number of upload slots
	MaxUploads int `json:"max_uploads"`
	// Maximum number of upload slots per torrent
	MaxUploadsPerTorrent int `json:"max_uploads_per_torrent"`
	// True if uTP protocol should be enabled; this option is only available in qBittorent built against libtorrent version 0.16.X and higher
	EnableUtp bool `json:"enable_utp"`
	// True if [du]l_limit should be applied to uTP connections; this option is only available in qBittorent built against libtorrent version 0.16.X and higher
	LimitUtpRate bool `json:"limit_utp_rate"`
	// True if [du]l_limit should be applied to estimated TCP overhead (service data: e.g. packet headers)
	LimitTcpOverhead bool `json:"limit_tcp_overhead"`
	// True if [du]l_limit should be applied to peers on the LAN
	LimitLanPeers bool `json:"limit_lan_peers"`
	// Alternative global download speed limit in KiB/s
	AltDlLimit int `json:"alt_dl_limit"`
	// Alternative global upload speed limit in KiB/s
	AltUpLimit int `json:"alt_up_limit"`
	// True if alternative limits should be applied according to schedule
	SchedulerEnabled bool `json:"scheduler_enabled"`
	// Scheduler starting hour
	ScheduleFromHour int `json:"schedule_from_hour"`
	// Scheduler starting minute
	ScheduleFromMin int `json:"schedule_from_min"`
	// Scheduler ending hour
	ScheduleToHour int `json:"schedule_to_hour"`
	// Scheduler ending minute
	ScheduleToMin int `json:"schedule_to_min"`
	// Scheduler days. See possible values here below
	SchedulerDays int `json:"scheduler_days"`
	// True if DHT is enabled
	Dht bool `json:"dht"`
	// True if DHT port should match TCP port
	DhtSameAsBT bool `json:"dhtSameAsBT"`
	// DHT port if dhtSameAsBT is false
	DhtPort int `json:"dht_port"`
	// True if PeX is enabled
	Pex bool `json:"pex"`
	// True if LSD is enabled
	Lsd bool `json:"lsd"`
	// See list of possible values here below
	Encryption int `json:"encryption"`
	// If true anonymous mode will be enabled; read more here; this option is only available in qBittorent built against libtorrent version 0.16.X and higher
	AnonymousMode bool `json:"anonymous_mode"`
	// See list of possible values here below
	ProxyType int `json:"proxy_type"`
	// Proxy IP address or domain name
	ProxyIp string `json:"proxy_ip"`
	// Proxy port
	ProxyPort int `json:"proxy_port"`
	// True if peer and web seed connections should be proxified; this option will have any effect only in qBittorent built against libtorrent version 0.16.X and higher
	ProxyPeerConnections bool `json:"proxy_peer_connections"`
	// True if the connections not supported by the proxy are disabled
	ForceProxy bool `json:"force_proxy"`
	// True proxy requires authentication; doesn't apply to SOCKS4 proxies
	ProxyAuthEnabled bool `json:"proxy_auth_enabled"`
	// Username for proxy authentication
	ProxyUsername string `json:"proxy_username"`
	// Password for proxy authentication
	ProxyPassword string `json:"proxy_password"`
	// True if external IP filter should be enabled
	IpFilterEnabled bool `json:"ip_filter_enabled"`
	// Path to IP filter file (.dat, .p2p, .p2b files are supported); path is separated by slashes
	IpFilterPath string `json:"ip_filter_path"`
	// True if IP filters are applied to trackers
	IpFilterTrackers bool `json:"ip_filter_trackers"`
	// Comma-separated list of domains to accept when performing Host header validation
	WebUiDomainList string `json:"web_ui_domain_list"`
	// IP address to use for the WebUI
	WebUiAddress string `json:"web_ui_address"`
	// WebUI port
	WebUiPort int `json:"web_ui_port"`
	// True if UPnP is used for the WebUI port
	WebUiUpnp bool `json:"web_ui_upnp"`
	// WebUI username
	WebUiUsername string `json:"web_ui_username"`
	// For API ≥ v2.3.0: Plaintext WebUI password, not readable, write-only. For API < v2.3.0: MD5 hash of WebUI password, hash is generated from the following string: username:Web UI Access:plain_text_web_ui_password
	WebUiPassword string `json:"web_ui_password"`
	// True if WebUI CSRF protection is enabled
	WebUiCsrfProtectionEnabled bool `json:"web_ui_csrf_protection_enabled"`
	// True if WebUI clickjacking protection is enabled
	WebUiClickjackingProtectionEnabled bool `json:"web_ui_clickjacking_protection_enabled"`
	// True if authentication challenge for loopback address (127.0.0.1) should be disabled
	BypassLocalAuth bool `json:"bypass_local_auth"`
	// True if webui authentication should be bypassed for clients whose ip resides within (at least) one of the subnets on the whitelist
	BypassAuthSubnetWhitelistEnabled bool `json:"bypass_auth_subnet_whitelist_enabled"`
	// (White)list of ipv4/ipv6 subnets for which webui authentication should be bypassed; list entries are separated by commas
	BypassAuthSubnetWhitelist string `json:"bypass_auth_subnet_whitelist"`
	// True if an alternative WebUI should be used
	AlternativeWebuiEnabled bool `json:"alternative_webui_enabled"`
	// File path to the alternative WebUI
	AlternativeWebuiPath string `json:"alternative_webui_path"`
	// True if WebUI HTTPS access is enabled
	UseHttps bool `json:"use_https"`
	// SSL keyfile contents (this is a not a path)
	SslKey string `json:"ssl_key"`
	// SSL certificate contents (this is a not a path)
	SslCert string `json:"ssl_cert"`
	// True if server DNS should be updated dynamically
	DyndnsEnabled bool `json:"dyndns_enabled"`
	// See list of possible values here below
	DyndnsService int `json:"dyndns_service"`
	// Username for DDNS service
	DyndnsUsername string `json:"dyndns_username"`
	// Password for DDNS service
	DyndnsPassword string `json:"dyndns_password"`
	// Your DDNS domain name
	DyndnsDomain string `json:"dyndns_domain"`
	// RSS refresh interval
	RssRefreshInterval int `json:"rss_refresh_interval"`
	// Max stored articles per RSS feed
	RssMaxArticlesPerFeed int `json:"rss_max_articles_per_feed"`
	// Enable processing of RSS feeds
	RssProcessingEnabled bool `json:"rss_processing_enabled"`
	// Enable auto-downloading of torrents from the RSS feeds
	RssAutoDownloadingEnabled bool `json:"rss_auto_downloading_enabled"`
}

type MaxRatioAction int

const (
	ActionPause  MaxRatioAction = 0
	ActionRemove                = 1
)
