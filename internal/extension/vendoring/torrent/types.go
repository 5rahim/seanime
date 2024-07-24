package vendor_hibike_torrent

type (
	AnimeProviderType string

	AnimeProvider interface {
		// Search for torrents.
		Search(opts AnimeSearchOptions) ([]*AnimeTorrent, error)
		// SmartSearch for torrents.
		SmartSearch(opts AnimeSmartSearchOptions) ([]*AnimeTorrent, error)
		// GetTorrentInfoHash returns the info hash of the torrent.
		// This should just return the info hash without scraping the torrent page if already available.
		GetTorrentInfoHash(torrent *AnimeTorrent) (string, error)
		// GetTorrentMagnetLink returns the magnet link of the torrent.
		// This should just return the magnet link without scraping the torrent page if already available.
		GetTorrentMagnetLink(torrent *AnimeTorrent) (string, error)
		// GetLatest returns the latest torrents.
		GetLatest() ([]*AnimeTorrent, error)
		// CanSmartSearch returns true if the provider supports smart search.
		// i.e. Searching related torrents without direct user query, based on the media.
		CanSmartSearch() bool
		// CanFindBestRelease returns true if the provider supports finding the best release.
		CanFindBestRelease() bool
		// SupportsAdult returns true if the provider supports searching for adult content.
		SupportsAdult() bool
		// GetType returns the provider type.
		GetType() AnimeProviderType
	}

	Media struct {
		// AniList ID of the media.
		ID int `json:"id"`
		// MyAnimeList ID of the media.
		IDMal *int `json:"idMal,omitempty"`
		// e.g. "FINISHED", "RELEASING", "NOT_YET_RELEASED", "CANCELLED", "HIATUS"
		// This should be set to "NOT_YET_RELEASED" if the status is unknown.
		Status string `json:"status,omitempty"`
		// e.g. "TV", "TV_SHORT", "MOVIE", "SPECIAL", "OVA", "ONA", "MUSIC"
		// This should be set to "TV" if the format is unknown.
		Format string `json:"format,omitempty"`
		// e.g. "Attack on Titan"
		EnglishTitle *string `json:"englishTitle,omitempty"`
		// e.g. "Shingeki no Kyojin"
		RomajiTitle string `json:"romajiTitle,omitempty"`
		// TotalEpisodes returns the total number of episodes of the media.
		// This should be set to -1 if the total number of episodes is unknown.
		EpisodeCount int `json:"episodeCount,omitempty"`
		// Absolute offset of the media's season.
		// This should be set to 0 if the media is not seasonal or the offset is unknown.
		AbsoluteSeasonOffset int `json:"absoluteSeasonOffset,omitempty"`
		// All alternative titles of the media.
		Synonyms []string `json:"synonyms"`
		// Whether the media is NSFW.
		IsAdult bool `json:"isAdult"`
		// StartDate of the media.
		// This should be nil if it has no start data.
		StartDate *FuzzyDate `json:"startDate,omitempty"`
	}

	FuzzyDate struct {
		Year  int  `json:"year"`
		Month *int `json:"month"`
		Day   *int `json:"day"`
	}

	// AnimeSearchOptions represents the options to search for torrents without filters.
	AnimeSearchOptions struct {
		Media Media
		// User query
		Query string `json:"query"`
		// Indicates if the search is for a batch torrent.
		Batch bool `json:"batch"`
	}

	AnimeSmartSearchOptions struct {
		Media Media `json:"media"`
		// Optional user query
		Query string `json:"query"`
		// Indicates if the search is for a batch torrent.
		Batch bool `json:"batch"`
		// Episode number of the torrent.
		EpisodeNumber int `json:"episodeNumber"`
		// Resolution of the torrent.
		// e.g. "1080", "720"
		Resolution string `json:"resolution"`
		// AniDB Anime ID of the media.
		AniDbAID int `json:"aniDbAID"`
		// AniDB Episode ID of the media.
		AniDbEID int `json:"aniDbEID"`
		// Look for the best release.
		BestReleases bool `json:"bestReleases"`
	}

	AnimeTorrent struct {
		Name string `json:"name"`
		// Date of the torrent.
		// The date should have RFC3339 format. e.g. "2006-01-02T15:04:05Z07:00"
		Date string `json:"date"`
		// Size of the torrent in bytes.
		Size int64 `json:"size"`
		// Human-readable size. e.g. "1.2 GB"
		FormattedSize string `json:"formattedSize"`
		// Number of seeders.
		Seeders int `json:"seeders"`
		// Number of leechers.
		Leechers int `json:"leechers"`
		// Number of downloads.
		DownloadCount int `json:"downloadCount"`
		// Link to the torrent page.
		Link string `json:"link"`
		// Direct download link to the torrent.
		DownloadUrl string `json:"downloadUrl"`
		// Magnet link of the torrent.
		// Leave empty if it should be scraped later.
		MagnetLink string `json:"magnetLink,omitempty"`
		// InfoHash of the torrent.
		// Leave empty if it should be scraped later.
		InfoHash string `json:"infoHash,omitempty"`
		// Resolution of the video.
		// e.g. "1080p", "720p"
		Resolution string `json:"resolution,omitempty"`
		// Indicates if the torrent is a batch.
		// Leave it as false if not a batch or unknown.
		IsBatch bool `json:"isBatch,omitempty"`
		// Episode number of the torrent.
		// This can be inferred from the query.
		// Return -1 if unknown.
		EpisodeNumber int `json:"episodeNumber,omitempty"`
		// Release group of the torrent.
		// Leave empty if unknown.
		ReleaseGroup string `json:"releaseGroup,omitempty"`
		// Provider of the torrent.
		// e.g. "Nyaa", "AnimeTosho"
		Provider string `json:"provider,omitempty"`
		// Indicates if the torrent is the best release for the specific media.
		// Should be a batch torrent unless the media is a movie.
		IsBestRelease bool `json:"isBestRelease"`
		// Indicates if the torrent is certainly related to the media.
		// i.e. the torrent is not a false positive.
		// e.g. If the torrent was found using the AniDB anime or episode ID
		Confirmed bool `json:"confirmed"`
	}
)
