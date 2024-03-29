package events

const (
	EventScanProgress                      = "scan-progress"                              // Progress of the scan
	EventScanStatus                        = "scan-status"                                // Status text of the scan
	RefreshedAnilistCollection             = "refreshed-anilist-collection"               // The anilist collection has been refreshed
	MediaPlayerTrackingStopped             = "media-player-tracking-stopped"              // DEPRECATED: The media player tracking has stopped
	MediaPlayerTrackingStarted             = "media-player-tracking-started"              // DEPRECATED: The media player tracking has started
	MediaPlayerVideoCompleted              = "media-player-video-completed"               // DEPRECATED: The video has been completed
	MediaPlayerProgressUpdateRequest       = "media-player-progress-update-request"       // The media player progress update request has been sent to the client
	MediaPlayerPlaybackStatus              = "media-player-playback-status"               // The playback status of the media player
	LibraryWatcherFileAdded                = "library-watcher-file-added"                 // A new file has been added to the library
	LibraryWatcherFileRemoved              = "library-watcher-file-removed"               // A file has been removed from the library
	AutoDownloaderItemAdded                = "auto-downloader-item-added"                 // An item has been added to the auto downloader queue
	AutoScanStarted                        = "auto-scan-started"                          // The auto scan has started
	AutoScanCompleted                      = "auto-scan-completed"                        // The auto scan has stopped
	PlaybackManagerProgressTrackingStarted = "playback-manager-progress-tracking-started" // The video progress tracking has started
	PlaybackManagerProgressTrackingStopped = "playback-manager-progress-tracking-stopped" // The video progress tracking has stopped
	PlaybackManagerProgressTrackingError   = "playback-manager-progress-tracking-error"   // The video progress tracking has an error
	PlaybackManagerProgressUpdateError     = "playback-manager-progress-update-error"     // The video progress update has an error
	PlaybackManagerProgressMetadataError   = "playback-manager-progress-metadata-error"   // Error occurred while fetching metadata
	PlaybackManagerProgressVideoCompleted  = "playback-manager-progress-video-completed"  // The video progress has been completed
	PlaybackManagerProgressPlaybackState   = "playback-manager-progress-playback-state"   // Dispatches the current playback state
	PlaybackManagerProgressUpdated         = "playback-manager-progress-updated"          // Signals that the progress has been updated
	PlaybackManagerNotifyInfo              = "playback-manager-notify-info"
	PlaybackManagerNotifyError             = "playback-manager-notify-error"
	PlaybackManagerPlaylistState           = "playback-manager-playlist-state" // Dispatches the current playlist state
)
