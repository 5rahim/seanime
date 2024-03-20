package onlinestream_providers

import (
	"errors"
	"github.com/seanime-app/seanime/internal/onlinestream/sources"
)

var (
	ErrSourceNotFound = errors.New("video source not found")
	ErrServerNotFound = errors.New("server not found")
)

const (
	GogoanimeProvider Provider = "gogoanime"
	ZoroProvider      Provider = "zoro"
)

type (
	AnimeProvider interface {
		Search(query string, dub bool) ([]*SearchResult, error)
		FindEpisodeDetails(id string) ([]*EpisodeDetails, error)
		FindEpisodeServer(episodeInfo *EpisodeDetails, server Server) (*EpisodeServer, error)
	}

	SearchResult struct {
		ID       string   `json:"id"`       // Anime slug
		Title    string   `json:"title"`    // Anime title
		URL      string   `json:"url"`      // Anime page URL
		SubOrDub SubOrDub `json:"subOrDub"` // Sub or Dub
	}

	// EpisodeDetails contains the episode information from a provider.
	// It is obtained by scraping the list of episodes.
	EpisodeDetails struct {
		Provider Provider `json:"provider"`
		ID       string   `json:"id"`              // Episode slug
		Number   int      `json:"number"`          // Episode number
		URL      string   `json:"url"`             // Watch URL
		Title    string   `json:"title,omitempty"` // Episode title
	}

	// EpisodeServer contains the server, headers and video sources for an episode.
	EpisodeServer struct {
		Provider     Provider                            `json:"provider"`
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

type Provider string
