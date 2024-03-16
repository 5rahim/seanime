package onlinestream

import (
	"errors"
	"github.com/seanime-app/seanime/internal/onlinestream/sources"
)

var (
	ErrSourceNotFound = errors.New("video source not found")
)

type AnimeProvider interface {
	Search(query string, dub bool) ([]*AnimeResult, error)
	FindAnimeEpisodes(id string) ([]*AnimeEpisode, error)
	FindVideoSources(episode *AnimeEpisode, server Server) (*onlinestream_sources.VideoSource, error)
}

type AnimeResult struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
	IsDub bool   `json:"isDub"`
}

type AnimeEpisode struct {
	ID     string `json:"id"`
	Number int    `json:"number"`
	URL    string `json:"url"`
}

type AnimeSource struct {
	Headers  map[string]string                   `json:"headers"`
	Sources  []*onlinestream_sources.VideoSource `json:"sources"`
	Download string                              `json:"download"`
}

type Server int

const (
	VidstreamingServer Server = iota + 1
	StreamSBServer
	GogocdnServer
)
