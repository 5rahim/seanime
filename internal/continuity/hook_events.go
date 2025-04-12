package continuity

import (
	"seanime/internal/hook_resolver"
	"seanime/internal/library/anime"
)

// WatchHistoryItemRequestedEvent is triggered when a watch history item is requested.
// Prevent default to skip getting the watch history item from the file cache, in this case the event should have a valid WatchHistoryItem object or set it to nil to indicate that the watch history item was not found.
type WatchHistoryItemRequestedEvent struct {
	hook_resolver.Event
	MediaId int `json:"mediaId"`
	// Empty WatchHistoryItem object, will be used if the hook prevents the default behavior
	WatchHistoryItem *WatchHistoryItem `json:"watchHistoryItem"`
}

type WatchHistoryLocalFileEpisodeItemRequestedEvent struct {
	hook_resolver.Event
	Path string
	// All scanned local files
	LocalFiles []*anime.LocalFile
	// Empty WatchHistoryItem object, will be used if the hook prevents the default behavior
	WatchHistoryItem *WatchHistoryItem `json:"watchHistoryItem"`
}

type WatchHistoryStreamEpisodeItemRequestedEvent struct {
	hook_resolver.Event
	Episode int
	MediaId int
	// Empty WatchHistoryItem object, will be used if the hook prevents the default behavior
	WatchHistoryItem *WatchHistoryItem `json:"watchHistoryItem"`
}
