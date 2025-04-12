package onlinestream_sources

import (
	"errors"
	hibikeonlinestream "seanime/internal/extension/hibike/onlinestream"
)

var (
	ErrNoVideoSourceFound    = errors.New("no episode source found")
	ErrVideoSourceExtraction = errors.New("error while extracting video sources")
)

type VideoExtractor interface {
	Extract(uri string) ([]*hibikeonlinestream.VideoSource, error)
}

const (
	QualityDefault = "default"
	QualityAuto    = "auto"
	Quality360     = "360"
	Quality480     = "480"
	Quality720     = "720"
	Quality1080    = "1080"
)
