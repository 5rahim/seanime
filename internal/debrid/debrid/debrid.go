package debrid

import "time"

type (
	Provider interface {
		GetSettings() Settings
		Authenticate(apiKey string) error
		AddTorrent(opts AddTorrentOptions) error
		StreamTorrent(id string) (streamUrl string, err error)
		DownloadTorrent(id string) (downloadUrl string, err error)
		GetTorrent(id string) (*TorrentItem, error)
		GetTorrents() ([]*TorrentItem, error)
		DeleteTorrent(id string) error
	}

	AddTorrentOptions struct {
		MagnetLink string `json:"magnetLink"`
	}

	TorrentItem struct {
		ID                   string            `json:"id"`
		Name                 string            `json:"name"`              // Name of the torrent or file
		Hash                 string            `json:"hash"`              // SHA1 hash of the torrent
		Bytes                int64             `json:"bytes"`             // Size of the selected files (size in bytes)
		CompletionPercentage int               `json:"progress"`          // Progress percentage (0 to 100)
		Status               TorrentItemStatus `json:"status"`            // Current download status
		AddedAt              time.Time         `json:"added"`             // Date when the torrent was added
		EndedAt              *time.Time        `json:"ended,omitempty"`   // Date when the torrent finished (optional, only when finished)
		Speed                int64             `json:"speed,omitempty"`   // Current download speed (optional, present in downloading state)
		Seeders              int               `json:"seeders,omitempty"` // Number of seeders (optional, present in downloading state)
	}

	TorrentItemStatus string

	////////////////////////////////////////////////////////////////////

	Settings struct {
		CanStream           bool `json:"canStream"`
		CanSelectStreamFile bool `json:"canSelectStreamFile"`
	}
)

const (
	TorrentItemStatusDownloading TorrentItemStatus = "downloading"
	TorrentItemStatusFinished    TorrentItemStatus = "finished"
	TorrentItemStatusSeeding     TorrentItemStatus = "seeding"
	TorrentItemStatusError       TorrentItemStatus = "error"
	TorrentItemStatusStalled     TorrentItemStatus = "stalled"
)
