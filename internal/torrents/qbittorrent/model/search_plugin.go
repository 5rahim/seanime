package qbittorrent_model

type SearchPlugin struct {
	Enabled             bool     `json:"enabled"`
	FullName            string   `json:"fullName"`
	Name                string   `json:"name"`
	SupportedCategories []string `json:"supportedCategories"`
	URL                 string   `json:"url"`
	Version             string   `json:"version"`
}
