package qbittorrent_model

type SyncPeersData struct {
	FullUpdate bool            `json:"full_update"`
	Peers      map[string]Peer `json:"peers"`
	RID        int             `json:"rid"`
	ShowFlags  bool            `json:"show_flags"`
}
