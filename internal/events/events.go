package events

type WebsocketClientEventType string

const (
	PluginEvent WebsocketClientEventType = "plugin"
)

type WebsocketClientEvent struct {
	ClientID string                   `json:"clientId"`
	Type     WebsocketClientEventType `json:"type"`
	Payload  interface{}              `json:"payload"`
}

const (
	ServerReady = "server-ready" // The anilist data has been loaded

	EventScanProgress               = "scan-progress"                      // Progress of the scan
	EventScanStatus                 = "scan-status"                        // Status text of the scan
	RefreshedAnilistAnimeCollection = "refreshed-anilist-anime-collection" // The anilist collection has been refreshed
	RefreshedAnilistMangaCollection = "refreshed-anilist-manga-collection" // The manga collection has been refreshed
	LibraryWatcherFileAdded         = "library-watcher-file-added"         // A new file has been added to the library
	LibraryWatcherFileRemoved       = "library-watcher-file-removed"       // A file has been removed from the library
	AutoDownloaderItemAdded         = "auto-downloader-item-added"         // An item has been added to the auto downloader queue

	AutoScanStarted   = "auto-scan-started"   // The auto scan has started
	AutoScanCompleted = "auto-scan-completed" // The auto scan has stopped

	PlaybackManagerProgressTrackingStarted     = "playback-manager-progress-tracking-started"      // The video progress tracking has started
	PlaybackManagerProgressTrackingStopped     = "playback-manager-progress-tracking-stopped"      // The video progress tracking has stopped
	PlaybackManagerProgressVideoCompleted      = "playback-manager-progress-video-completed"       // The video progress has been completed
	PlaybackManagerProgressPlaybackState       = "playback-manager-progress-playback-state"        // Dispatches the current playback state
	PlaybackManagerProgressUpdated             = "playback-manager-progress-updated"               // Signals that the progress has been updated
	PlaybackManagerPlaylistState               = "playback-manager-playlist-state"                 // Dispatches the current playlist state
	PlaybackManagerManualTrackingPlaybackState = "playback-manager-manual-tracking-playback-state" // Dispatches the current playback state
	PlaybackManagerManualTrackingStopped       = "playback-manager-manual-tracking-stopped"        // The manual tracking has been stopped

	ExternalPlayerOpenURL = "external-player-open-url" // Open a URL to send media to an external media player

	InfoToast    = "info-toast"
	ErrorToast   = "error-toast"
	WarningToast = "warning-toast"
	SuccessToast = "success-toast"

	CheckForUpdates = "check-for-updates"

	RefreshedMangaDownloadData  = "refreshed-manga-download-data"
	ChapterDownloadQueueUpdated = "chapter-download-queue-updated"
	OfflineSnapshotCreated      = "offline-snapshot-created"

	MediastreamShutdownStream = "mediastream-shutdown-stream"

	ExtensionsReloaded    = "extensions-reloaded"
	ExtensionUpdatesFound = "extension-updates-found"
	PluginUnloaded        = "plugin-unloaded"
	PluginLoaded          = "plugin-loaded"

	ActiveTorrentCountUpdated = "active-torrent-count-updated"

	SyncLocalQueueState = "sync-local-queue-state"
	SyncLocalFinished   = "sync-local-finished"
	SyncAnilistFinished = "sync-anilist-finished"

	DebridDownloadProgress = "debrid-download-progress"
	DebridStreamState      = "debrid-stream-state"

	InvalidateQueries = "invalidate-queries"
	ConsoleLog        = "console-log"
	ConsoleWarn       = "console-warn"
)
