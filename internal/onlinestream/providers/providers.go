package onlinestream

import (
	"errors"
	onlinestream "github.com/seanime-app/seanime/internal/onlinestream/extractors"
)

var (
	ErrSourceNotFound = errors.New("source not found")
)

type AnimeProvider interface {
	Search(query string, dubbed bool) ([]*AnimeResult, error)
	FetchAnimeEpisodes(id string) ([]*AnimeEpisode, error)
	FetchEpisodeSources(episode *AnimeEpisode, server Server) (*AnimeSource, error)
}

type AnimeResult struct {
	ID    string
	Title string
	URL   string
	IsDub bool
}

type AnimeEpisode struct {
	ID     string
	Number int
	URL    string
}

type AnimeSource struct {
	Headers  map[string]string
	Sources  []*onlinestream.VideoSource
	Download string
}

type Server int

const (
	VidstreamingServer Server = iota + 1
	StreamSBServer
	GogocdnServer
)
