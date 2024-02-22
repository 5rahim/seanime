package torrent

type (
	AnimeTorrent struct {
		Name          string `json:"name"`
		Date          string `json:"date"`
		Size          string `json:"size"`
		Seeders       string `json:"seeders"`
		Leechers      string `json:"leechers"`
		DownloadCount string `json:"downloads"`
		Link          string `json:"link"`
		GUID          string `json:"guid"`
		InfoHash      string `json:"infoHash"`
	}
)
