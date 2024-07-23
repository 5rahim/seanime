package qbittorrent_model

type TorrentTracker struct {
	URL           string        `json:"url"`
	Status        TrackerStatus `json:"status"`
	Tier          int           `json:"tier"`
	NumPeers      int           `json:"num_peers"`
	NumSeeds      int           `json:"num_seeds"`
	NumLeeches    int           `json:"num_leeches"`
	NumDownloaded int           `json:"num_downloaded"`
	Message       string        `json:"msg"`
}

type TrackerStatus int

const (
	TrackerStatusDisabled TrackerStatus = iota
	TrackerStatusNotContacted
	TrackerStatusWorking
	TrackerStatusUpdating
	TrackerStatusNotWorking
)
