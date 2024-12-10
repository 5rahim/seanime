package onlinestream_providers

import "errors"

// Built-in
const (
	GogoanimeProvider string = "gogoanime"
	ZoroProvider      string = "zoro"
)

// Built-in
const (
	DefaultServer      = "default"
	GogocdnServer      = "gogocdn"
	VidstreamingServer = "vidstreaming"
	StreamSBServer     = "streamsb"
	VidcloudServer     = "vidcloud"
	StreamtapeServer   = "streamtape"
	KwikServer         = "kwik"
)

var (
	ErrSourceNotFound = errors.New("video source not found")
	ErrServerNotFound = errors.New("server not found")
)
