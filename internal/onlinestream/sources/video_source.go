package onlinestream_sources

// VideoSource represents a video source.
type VideoSource struct {
	URL     string          `json:"url"`
	Type    VideoSourceType `json:"type"`
	Quality string          `json:"quality"`
}

type VideoSourceType int

const (
	VideoSourceMP4 VideoSourceType = iota + 1
	VideoSourceM3U8
	VideoSourceDash
)
