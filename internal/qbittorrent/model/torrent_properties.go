package qbittorrent_model

import (
	"encoding/json"
	"time"
)

type TorrentProperties struct {
	// Torrent save path
	SavePath string `json:"save_path"`
	// Torrent creation date (Unix timestamp)
	CreationDate time.Time `json:"creation_date"`
	// Torrent piece size (bytes)
	PieceSize int `json:"piece_size"`
	// Torrent comment
	Comment string `json:"comment"`
	// Total data wasted for torrent (bytes)
	TotalWasted int `json:"total_wasted"`
	// Total data uploaded for torrent (bytes)
	TotalUploaded int `json:"total_uploaded"`
	// Total data uploaded this session (bytes)
	TotalUploadedSession int `json:"total_uploaded_session"`
	// Total data downloaded for torrent (bytes)
	TotalDownloaded int `json:"total_downloaded"`
	// Total data downloaded this session (bytes)
	TotalDownloadedSession int `json:"total_downloaded_session"`
	// Torrent upload limit (bytes/s)
	UpLimit int `json:"up_limit"`
	// Torrent download limit (bytes/s)
	DlLimit int `json:"dl_limit"`
	// Torrent elapsed time (seconds)
	TimeElapsed int `json:"time_elapsed"`
	// Torrent elapsed time while complete (seconds)
	SeedingTime time.Duration `json:"seeding_time"`
	// Torrent connection count
	NbConnections int `json:"nb_connections"`
	// Torrent connection count limit
	NbConnectionsLimit int `json:"nb_connections_limit"`
	// Torrent share ratio
	ShareRatio float64 `json:"share_ratio"`
	// When this torrent was added (unix timestamp)
	AdditionDate time.Time `json:"addition_date"`
	// Torrent completion date (unix timestamp)
	CompletionDate time.Time `json:"completion_date"`
	// Torrent creator
	CreatedBy string `json:"created_by"`
	// Torrent average download speed (bytes/second)
	DlSpeedAvg int `json:"dl_speed_avg"`
	// Torrent download speed (bytes/second)
	DlSpeed int `json:"dl_speed"`
	// Torrent ETA (seconds)
	Eta time.Duration `json:"eta"`
	// Last seen complete date (unix timestamp)
	LastSeen time.Time `json:"last_seen"`
	// Number of peers connected to
	Peers int `json:"peers"`
	// Number of peers in the swarm
	PeersTotal int `json:"peers_total"`
	// Number of pieces owned
	PiecesHave int `json:"pieces_have"`
	// Number of pieces of the torrent
	PiecesNum int `json:"pieces_num"`
	// Number of seconds until the next announce
	Reannounce time.Duration `json:"reannounce"`
	// Number of seeds connected to
	Seeds int `json:"seeds"`
	// Number of seeds in the swarm
	SeedsTotal int `json:"seeds_total"`
	// Torrent total size (bytes)
	TotalSize int `json:"total_size"`
	// Torrent average upload speed (bytes/second)
	UpSpeedAvg int `json:"up_speed_avg"`
	// Torrent upload speed (bytes/second)
	UpSpeed int `json:"up_speed"`
}

func (p *TorrentProperties) UnmarshalJSON(data []byte) error {
	var raw rawTorrentProperties
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	creationDate := time.Unix(int64(raw.CreationDate), 0)
	seedingTime := time.Duration(raw.SeedingTime) * time.Second
	additionDate := time.Unix(int64(raw.AdditionDate), 0)
	completionDate := time.Unix(int64(raw.CompletionDate), 0)
	eta := time.Duration(raw.Eta) * time.Second
	lastSeen := time.Unix(int64(raw.LastSeen), 0)
	reannounce := time.Duration(raw.Reannounce) * time.Second
	*p = TorrentProperties{
		SavePath:               raw.SavePath,
		CreationDate:           creationDate,
		PieceSize:              raw.PieceSize,
		Comment:                raw.Comment,
		TotalWasted:            raw.TotalWasted,
		TotalUploaded:          raw.TotalUploaded,
		TotalUploadedSession:   raw.TotalUploadedSession,
		TotalDownloaded:        raw.TotalDownloaded,
		TotalDownloadedSession: raw.TotalDownloadedSession,
		UpLimit:                raw.UpLimit,
		DlLimit:                raw.DlLimit,
		TimeElapsed:            raw.TimeElapsed,
		SeedingTime:            seedingTime,
		NbConnections:          raw.NbConnections,
		NbConnectionsLimit:     raw.NbConnectionsLimit,
		ShareRatio:             raw.ShareRatio,
		AdditionDate:           additionDate,
		CompletionDate:         completionDate,
		CreatedBy:              raw.CreatedBy,
		DlSpeedAvg:             raw.DlSpeedAvg,
		DlSpeed:                raw.DlSpeed,
		Eta:                    eta,
		LastSeen:               lastSeen,
		Peers:                  raw.Peers,
		PeersTotal:             raw.PeersTotal,
		PiecesHave:             raw.PiecesHave,
		PiecesNum:              raw.PiecesNum,
		Reannounce:             reannounce,
		Seeds:                  raw.Seeds,
		SeedsTotal:             raw.SeedsTotal,
		TotalSize:              raw.TotalSize,
		UpSpeedAvg:             raw.UpSpeedAvg,
		UpSpeed:                raw.UpSpeed,
	}
	return nil
}

type rawTorrentProperties struct {
	// Torrent save path
	SavePath string `json:"save_path"`
	// Torrent creation date (Unix timestamp)
	CreationDate int `json:"creation_date"`
	// Torrent piece size (bytes)
	PieceSize int `json:"piece_size"`
	// Torrent comment
	Comment string `json:"comment"`
	// Total data wasted for torrent (bytes)
	TotalWasted int `json:"total_wasted"`
	// Total data uploaded for torrent (bytes)
	TotalUploaded int `json:"total_uploaded"`
	// Total data uploaded this session (bytes)
	TotalUploadedSession int `json:"total_uploaded_session"`
	// Total data downloaded for torrent (bytes)
	TotalDownloaded int `json:"total_downloaded"`
	// Total data downloaded this session (bytes)
	TotalDownloadedSession int `json:"total_downloaded_session"`
	// Torrent upload limit (bytes/s)
	UpLimit int `json:"up_limit"`
	// Torrent download limit (bytes/s)
	DlLimit int `json:"dl_limit"`
	// Torrent elapsed time (seconds)
	TimeElapsed int `json:"time_elapsed"`
	// Torrent elapsed time while complete (seconds)
	SeedingTime int `json:"seeding_time"`
	// Torrent connection count
	NbConnections int `json:"nb_connections"`
	// Torrent connection count limit
	NbConnectionsLimit int `json:"nb_connections_limit"`
	// Torrent share ratio
	ShareRatio float64 `json:"share_ratio"`
	// When this torrent was added (unix timestamp)
	AdditionDate int `json:"addition_date"`
	// Torrent completion date (unix timestamp)
	CompletionDate int `json:"completion_date"`
	// Torrent creator
	CreatedBy string `json:"created_by"`
	// Torrent average download speed (bytes/second)
	DlSpeedAvg int `json:"dl_speed_avg"`
	// Torrent download speed (bytes/second)
	DlSpeed int `json:"dl_speed"`
	// Torrent ETA (seconds)
	Eta int `json:"eta"`
	// Last seen complete date (unix timestamp)
	LastSeen int `json:"last_seen"`
	// Number of peers connected to
	Peers int `json:"peers"`
	// Number of peers in the swarm
	PeersTotal int `json:"peers_total"`
	// Number of pieces owned
	PiecesHave int `json:"pieces_have"`
	// Number of pieces of the torrent
	PiecesNum int `json:"pieces_num"`
	// Number of seconds until the next announce
	Reannounce int `json:"reannounce"`
	// Number of seeds connected to
	Seeds int `json:"seeds"`
	// Number of seeds in the swarm
	SeedsTotal int `json:"seeds_total"`
	// Torrent total size (bytes)
	TotalSize int `json:"total_size"`
	// Torrent average upload speed (bytes/second)
	UpSpeedAvg int `json:"up_speed_avg"`
	// Torrent upload speed (bytes/second)
	UpSpeed int `json:"up_speed"`
}
