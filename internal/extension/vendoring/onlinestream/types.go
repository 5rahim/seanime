package vender_hibike_onlinestream

type (
	Provider interface {
		Search(query string, dub bool) ([]*SearchResult, error)
		// FindEpisode returns the episode details for the given anime ID.
		// The ID is the anime slug.
		FindEpisode(id string) ([]*EpisodeDetails, error)
		// FindEpisodeServer returns the episode server for the given episode.
		// "server" can be "default"
		FindEpisodeServer(episode *EpisodeDetails, server string) (*EpisodeServer, error)
	}

	SearchResult struct {
		// ID is the anime slug.
		// It is used to fetch the episode details.
		ID string `json:"id"`
		// Title is the anime title.
		Title string `json:"title"`
		// URL is the anime page URL.
		URL      string   `json:"url"`
		SubOrDub SubOrDub `json:"subOrDub"`
	}

	// EpisodeDetails contains the episode information from a provider.
	// It is obtained by scraping the list of episodes.
	EpisodeDetails struct {
		// Provider is the ID of the provider.
		// This should be the same as the extension ID and follow the same format.
		Provider string `json:"provider"`
		// ID is the episode slug.
		// e.g. "the-apothecary-diaries-18578".
		ID string `json:"id"`
		// Episode number.
		// From 0 to n.
		Number int `json:"number"`
		// Episode page URL.
		URL string `json:"url"`
		// Episode title.
		// Leave it empty if the episode title is not available.
		Title string `json:"title,omitempty"`
	}

	// EpisodeServer contains the server, headers and video sources for an episode.
	EpisodeServer struct {
		// Provider is the ID of the provider.
		// This should be the same as the extension ID and follow the same format.
		Provider string `json:"provider"`
		// Server is the video server name.
		// e.g. "vidcloud".
		Server string `json:"server"`
		// Headers are the HTTP headers for the video request.
		Headers map[string]string `json:"headers"`
		// Video sources for the episode.
		VideoSources []*VideoSource `json:"videoSources"`
	}

	SubOrDub string

	VideoSourceType string

	VideoSource struct {
		// URL of the video source.
		URL string `json:"url"`
		// Type of the video source.
		Type VideoSourceType `json:"type"`
		// Quality of the video source.
		// e.g. "default", "auto", "1080p".
		Quality string `json:"quality"`
		// Subtitles for the video source.
		Subtitles []*VideoSubtitle `json:"subtitles"`
	}

	VideoSubtitle struct {
		ID        string `json:"id"`
		URL       string `json:"url"`
		Language  string `json:"language"`
		IsDefault bool   `json:"isDefault"`
	}

	VideoExtractor interface {
		Extract(uri string) ([]*VideoSource, error)
	}
)

const (
	Sub       SubOrDub = "sub"
	Dub       SubOrDub = "dub"
	SubAndDub SubOrDub = "both"
)

const (
	VideoSourceMP4  VideoSourceType = "mp4"
	VideoSourceM3U8 VideoSourceType = "m3u8"
	VideoSourceDash VideoSourceType = "dash"
)
