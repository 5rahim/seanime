package events

const (
	EventScanProgress                = "scan-progress"                        // Progress of the scan
	EventScanStatus                  = "scan-status"                          // Status text of the scan
	RefreshedAnilistCollection       = "refreshed-anilist-collection"         // The anilist collection has been refreshed
	MediaPlayerTrackingStopped       = "media-player-tracking-stopped"        // The media player tracking has stopped
	MediaPlayerTrackingStarted       = "media-player-tracking-started"        // The media player tracking has started
	MediaPlayerVideoCompleted        = "media-player-video-completed"         // The video has been completed
	MediaPlayerProgressUpdateRequest = "media-player-progress-update-request" // The media player progress update request has been sent to the client
	MediaPlayerPlaybackStatus        = "media-player-playback-status"         // The playback status of the media player
	LibraryWatcherFileAdded          = "library-watcher-file-added"           // A new file has been added to the library
	LibraryWatcherFileRemoved        = "library-watcher-file-removed"         // A file has been removed from the library
	AutoDownloaderItemAdded          = "auto-downloader-item-added"           // An item has been added to the auto downloader queue
	AutoScanStarted                  = "auto-scan-started"                    // The auto scan has started
	AutoScanCompleted                = "auto-scan-completed"                  // The auto scan has stopped
)
