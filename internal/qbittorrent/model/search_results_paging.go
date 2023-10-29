package qbittorrent_model

type SearchResultsPaging struct {
	Results []SearchResult `json:"results"`
	Status  string         `json:"status"`
	Total   int            `json:"total"`
}
