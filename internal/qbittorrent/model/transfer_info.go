package qbittorrent_model

type TransferInfo struct {
	ConnectionStatus  ConnectionStatus `json:"connection_status"`
	DhtNodes          int              `json:"dht_nodes"`
	DlInfoData        int              `json:"dl_info_data"`
	DlInfoSpeed       int              `json:"dl_info_speed"`
	DlRateLimit       int              `json:"dl_rate_limit"`
	UpInfoData        int              `json:"up_info_data"`
	UpInfoSpeed       int              `json:"up_info_speed"`
	UpRateLimit       int              `json:"up_rate_limit"`
	UseAltSpeedLimits bool             `json:"use_alt_speed_limits"`
	Queueing          bool             `json:"queueing"`
	RefreshInterval   int              `json:"refresh_interval"`
}

type ConnectionStatus string

const (
	StatusConnected    ConnectionStatus = "connected"
	StatusFirewalled                    = "firewalled"
	StatusDisconnected                  = "disconnected"
)
