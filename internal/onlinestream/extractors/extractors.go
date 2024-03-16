package onlinestream_extractors

import "errors"

// VideoSource represents a video source.
type VideoSource struct {
	URL     string
	IsM3U8  bool
	Quality string
}

var (
	// ErrNoEpisodeSourceFound is an error indicating no episode source is found.
	ErrNoEpisodeSourceFound  = errors.New("no episode source found")
	ErrVideoSourceExtraction = errors.New("error while extracting video sources")
)

type VideoExtractor interface {
	Extract(uri string) ([]*VideoSource, error)
}
