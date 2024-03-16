package onlinestream_sources

// VideoSource represents a video source.
type VideoSource struct {
	URL     string          `json:"url"`
	Type    VideoSourceType `json:"type"`
	Quality string          `json:"quality"`
	Size    float64         `json:"size"`
}

type VideoSourceType int

const (
	VideoSourceMP4 VideoSourceType = iota + 1
	VideoSourceM3U8
	VideoSourceDash
)

type VideoSubtitle struct {
	URL       string `json:"url"`
	Id        string `json:"id"`
	Language  string `json:"language"`
	IsDefault bool   `json:"isDefault"`
}
