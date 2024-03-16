package onlinestream_sources

import (
	"errors"
)

var (
	ErrNoVideoSourceFound    = errors.New("no episode source found")
	ErrVideoSourceExtraction = errors.New("error while extracting video sources")
)

type VideoExtractor interface {
	Extract(uri string) ([]*VideoSource, error)
}
