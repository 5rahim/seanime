package qbittorrent_model

type BuildInfo struct {
	QT         string `json:"qt"`
	LibTorrent string `json:"libtorrent"`
	Boost      string `json:"boost"`
	OpenSSL    string `json:"openssl"`
	Bitness    string `json:"bitness"`
}
