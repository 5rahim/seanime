package qbittorrent_model

type AddTorrentsOptions struct {
	// Download folder
	Savepath string `json:"savepath,omitempty"`
	// Cookie sent to download the .torrent file
	Cookie string `json:"cookie,omitempty"`
	// Category for the torrent
	Category string `json:"category,omitempty"`
	// Skip hash checking.
	SkipChecking bool `json:"skip_checking,omitempty"`
	// Add torrents in the paused state.
	Paused string `json:"paused,omitempty"`
	// Create the root folder. Possible values are true, false, unset (default)
	RootFolder string `json:"root_folder,omitempty"`
	// Rename torrent
	Rename string `json:"rename,omitempty"`
	// Set torrent upload speed limit. Unit in bytes/second
	UpLimit int `json:"upLimit,omitempty"`
	// Set torrent download speed limit. Unit in bytes/second
	DlLimit int `json:"dlLimit,omitempty"`
	// Whether Automatic Torrent Management should be used
	UseAutoTMM bool `json:"useAutoTMM,omitempty"`
	// Enable sequential download. Possible values are true, false (default)
	SequentialDownload bool `json:"sequentialDownload,omitempty"`
	// Prioritize download first last piece. Possible values are true, false (default)
	FirstLastPiecePrio bool `json:"firstLastPiecePrio,omitempty"`
	// Tags for the torrent, split by ','
	Tags string `json:"tags,omitempty"`
}
