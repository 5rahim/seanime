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
		FindEpisodesInfo(id string) ([]*ProviderEpisodeInfo, error)
		FindEpisodeServerSources(episodeInfo *ProviderEpisodeInfo, server Server) (*ProviderServerSources, error)
	}

	SearchResult struct {
		ID       string   `json:"id"`       // Anime slug
		Title    string   `json:"title"`    // Anime title
		URL      string   `json:"url"`      // Anime page URL
		SubOrDub SubOrDub `json:"subOrDub"` // Sub or Dub
	}

	ProviderEpisodeInfo struct {
		ID     string `json:"id"`     // Episode slug
		Number int    `json:"number"` // Episode number
		URL    string `json:"url"`    // Watch URL
		Title  string `json:"title"`  // Episode title
	}

	ProviderServerSources struct {
		Server       Server                              `json:"server"`
		Headers      map[string]string                   `json:"headers"`
		VideoSources []*onlinestream_sources.VideoSource `json:"videoSources"`
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
