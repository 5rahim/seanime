package qbittorrent_model

type GetLogOptions struct {
	Normal      bool `url:"normal"`
	Info        bool `url:"info"`
	Warning     bool `url:"warning"`
	Critical    bool `url:"critical"`
	LastKnownID int  `url:"lastKnownId"`
}
