package qbittorrent_model

type SearchResult struct {
	// URL of the torrent's description page
	DescriptionLink string `json:"descrLink"`
	// Name of the file
	FileName string `json:"fileName"`
	// Size of the file in Bytes
	FileSize int `json:"fileSize"`
	// Torrent download link (usually either .torrent file or magnet link)
	FileUrl string `json:"fileUrl"`
	// Number of leechers
	NumLeechers int `json:"nbLeechers"`
	// int of seeders
	NumSeeders int `json:"nbSeeders"`
	// URL of the torrent site
	SiteUrl string `json:"siteUrl"`
}
