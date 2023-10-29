package qbittorrent_model

type TorrentPieceState int

const (
	PieceStateNotDownloaded TorrentPieceState = iota
	PieceStateDownloading
	PieceStateDownloaded
)
