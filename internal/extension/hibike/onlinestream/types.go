package hibikeonlinestream

type (
	Provider interface {
		Search(opts SearchOptions) ([]*SearchResult, error)
		// FindEpisodes returns the episodes for the given anime ID.
		FindEpisodes(id string) ([]*EpisodeDetails, error)
		// FindEpisodeServer returns the episode server for the given episode.
		// The "server" argument can be "default"
		FindEpisodeServer(episode *EpisodeDetails, server string) (*EpisodeServer, error)
		// GetSettings returns the provider settings.
		GetSettings() Settings
	}

	SearchOptions struct {
		// The media object provided by Seanime.
		Media Media `json:"media"`
		// The search query.
		Query string `json:"query"`
		// Whether to search for subbed or dubbed anime.
		Dub bool `json:"dub"`
		// The year the anime was released.
		// Will be 0 if the year is not available.
		Year int `json:"year"`
	}

	Media struct {
		// AniList ID of the media.
		ID int `json:"id"`
		// MyAnimeList ID of the media.
		IDMal *int `json:"idMal,omitempty"`
		// e.g. "FINISHED", "RELEASING", "NOT_YET_RELEASED", "CANCELLED", "HIATUS"
		// This will be set to "NOT_YET_RELEASED" if the status is unknown.
		Status string `json:"status,omitempty"`
		// e.g. "TV", "TV_SHORT", "MOVIE", "SPECIAL", "OVA", "ONA", "MUSIC"
		// This will be set to "TV" if the format is unknown.
		Format string `json:"format,omitempty"`
		// e.g. "Attack on Titan"
		// This will be undefined if the english title is unknown.
		EnglishTitle *string `json:"englishTitle,omitempty"`
		// e.g. "Shingeki no Kyojin"
		RomajiTitle string `json:"romajiTitle,omitempty"`
		// TotalEpisodes is total number of episodes of the media.
		// This will be -1 if the total number of episodes is unknown / not applicable.
		EpisodeCount int `json:"episodeCount,omitempty"`
		// All alternative titles of the media.
		Synonyms []string `json:"synonyms"`
		// Whether the media is NSFW.
		IsAdult bool `json:"isAdult"`
		// Start date of the media.
		// This will be undefined if it has no start date.
		StartDate *FuzzyDate `json:"startDate,omitempty"`
	}

	FuzzyDate struct {
		Year  int  `json:"year"`
		Month *int `json:"month"`
		Day   *int `json:"day"`
	}

	Settings struct {
		EpisodeServers []string `json:"episodeServers"`
		SupportsDub    bool     `json:"supportsDub"`
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
		// "ID" of the extension.
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
		// "ID" of the extension.
		Provider string `json:"provider"`
		// Episode server name.
		// e.g. "vidcloud".
		Server string `json:"server"`
		// HTTP headers for the video request.
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
		ID  string `json:"id"`
		URL string `json:"url"`
		// e.g. "en", "fr"
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
)
