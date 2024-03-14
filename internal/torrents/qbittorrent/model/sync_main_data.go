package qbittorrent_model

type SyncMainData struct {
	RID               int                 `json:"rid"`
	FullUpdate        bool                `json:"full_update"`
	Torrents          map[string]*Torrent `json:"torrents"`
	TorrentsRemoved   []string            `json:"torrents_removed"`
	Categories        map[string]Category `json:"categories"`
	CategoriesRemoved map[string]Category `json:"categories_removed"`
	Queueing          bool                `json:"queueing"`
	ServerState       ServerState         `json:"server_state"`
}
