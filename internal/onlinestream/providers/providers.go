package onlinestream_providers

import (
	"errors"
	"github.com/seanime-app/seanime/internal/onlinestream/sources"
)

var (
	ErrSourceNotFound = errors.New("video source not found")
	ErrServerNotFound = errors.New("server not found")
)

type (
	Provider interface {
		Search(query string, dub bool) ([]*SearchResult, error)
		FindEpisodes(id string) ([]*ProviderEpisode, error)
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
		Title  string `json:"title"`  // Episode title
	}

	ProviderEpisodeSource struct {
		Headers map[string]string                   `json:"headers"`
		Sources []*onlinestream_sources.VideoSource `json:"sources"`
	}

	Server string

	SubOrDub string
)

const (
	Sub       SubOrDub = "sub"
	Dub       SubOrDub = "dub"
	SubAndDub SubOrDub = "both"
)

const (
	DefaultServer      Server = "default"
	VidstreamingServer Server = "vidstreaming"
	StreamSBServer     Server = "streamsb"
	GogocdnServer      Server = "gogocdn"
	StreamtapeServer   Server = "streamtape"
	VidcloudServer     Server = "vidcloud"
)
