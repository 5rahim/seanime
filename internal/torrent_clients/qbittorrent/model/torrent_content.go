package qbittorrent_model

type TorrentContent struct {
	// File name (including relative path)
	Name string `json:"	name"`
	// File size (bytes)
	Size int `json:"	size"`
	// File progress (percentage/100)
	Progress float64 `json:"	progress"`
	// File priority. See possible values here below
	Priority TorrentPriority `json:"	priority"`
	// True if file is seeding/complete
	IsSeed bool `json:"	is_seed"`
	// The first number is the starting piece index and the second number is the ending piece index (inclusive)
	PieceRange []int `json:"	piece_range"`
	// Percentage of file pieces currently available
	Availability float64 `json:"	availability"`
}

type TorrentPriority int

const (
	PriorityDoNotDownload TorrentPriority = 0
	PriorityNormal                        = 1
	PriorityHigh                          = 6
	PriorityMaximum                       = 7
)
