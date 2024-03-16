package onlinestream_providers

import (
	"errors"
	"github.com/seanime-app/seanime/internal/onlinestream/sources"
)

var (
	ErrSourceNotFound = errors.New("video source not found")
)

type (
	AnimeProvider interface {
		Search(query string, dub bool) ([]*SearchResult, error)
		FindAnimeEpisodes(id string) ([]*ProviderEpisode, error)
		FindEpisodeSources(episode *ProviderEpisode, server Server) (*ProviderEpisodeSource, error)
	}

	SearchResult struct {
		ID       string   `json:"id"`       // Anime slug
		Title    string   `json:"title"`    // Anime title
		URL      string   `json:"url"`      // Anime page URL
		SubOrDub SubOrDub `json:"subOrDub"` // Sub or Dub
	}

	ProviderEpisode struct {
		ID     string `json:"id"`     // Episode slug
		Number int    `json:"number"` // Episode number
		URL    string `json:"url"`    // Watch URL
	}

	ProviderEpisodeSource struct {
		Headers   map[string]string                     `json:"headers"`
		Sources   []*onlinestream_sources.VideoSource   `json:"sources"`
		Subtitles []*onlinestream_sources.VideoSubtitle `json:"subtitles"`
	}

	Server string

	SubOrDub string
)

const (
	Sub       SubOrDub = "sub"
	Dub       SubOrDub = "dub"
	SubAndDub SubOrDub = "subAndDub"
)

const (
	VidstreamingServer Server = "vidstreaming"
	StreamSBServer     Server = "streamsb"
	GogocdnServer      Server = "gogocdn"
)
