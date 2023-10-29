package qbittorrent_model

type SearchStatus struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
	Total  int    `json:"total"`
}
