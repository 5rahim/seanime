package qbittorrent_model

type GetTorrentListOptions struct {
	Filter   TorrentListFilter `url:"filter,omitempty"`
	Category *string           `url:"category,omitempty"`
	Sort     string            `url:"sort,omitempty"`
	Reverse  bool              `url:"reverse,omitempty"`
	Limit    int               `url:"limit,omitempty"`
	Offset   int               `url:"offset,omitempty"`
	Hashes   string            `url:"hashes,omitempty"`
}

type TorrentListFilter string

const (
	FilterAll         TorrentListFilter = "all"
	FilterDownloading                   = "downloading"
	FilterCompleted                     = "completed"
	FilterPaused                        = "paused"
	FilterActive                        = "active"
	FilterInactive                      = "inactive"
	FilterResumed                       = "resumed"
)
