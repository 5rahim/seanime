package hibiketorrent

// Resolutions represent resolution filters available to the user.
var Resolutions = []string{"1080", "720", "540", "480"}

const (
	// AnimeProviderTypeMain providers can be used as default providers.
	AnimeProviderTypeMain AnimeProviderType = "main"
	// AnimeProviderTypeSpecial providers cannot be set as default provider.
	// Providers that return only specific content (e.g. adult content).
	// These providers should not return anything from "GetLatest".
	AnimeProviderTypeSpecial AnimeProviderType = "special"
)

const (
	AnimeProviderSmartSearchFilterBatch         AnimeProviderSmartSearchFilter = "batch"
	AnimeProviderSmartSearchFilterEpisodeNumber AnimeProviderSmartSearchFilter = "episodeNumber"
	AnimeProviderSmartSearchFilterResolution    AnimeProviderSmartSearchFilter = "resolution"
	AnimeProviderSmartSearchFilterQuery         AnimeProviderSmartSearchFilter = "query"
	AnimeProviderSmartSearchFilterBestReleases  AnimeProviderSmartSearchFilter = "bestReleases"
)

type (
	AnimeProviderType string

	AnimeProviderSmartSearchFilter string

	AnimeProviderSettings struct {
		CanSmartSearch     bool                             `json:"canSmartSearch"`
		SmartSearchFilters []AnimeProviderSmartSearchFilter `json:"smartSearchFilters"`
		SupportsAdult      bool                             `json:"supportsAdult"`
		Type               AnimeProviderType                `json:"type"`
	}

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
		// GetSettings returns the provider settings.
		GetSettings() AnimeProviderSettings
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
		// Absolute offset of the media's season.
		// This will be 0 if the media is not seasonal or the offset is unknown.
		AbsoluteSeasonOffset int `json:"absoluteSeasonOffset,omitempty"`
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

	// AnimeSearchOptions represents the options to search for torrents without filters.
	AnimeSearchOptions struct {
		// The media object provided by Seanime.
		Media Media `json:"media"`
		// The user search query.
		Query string `json:"query"`
	}

	AnimeSmartSearchOptions struct {
		// The media object provided by Seanime.
		Media Media `json:"media"`
		// The user search query.
		// This will be empty if your extension does not support custom queries.
		Query string `json:"query"`
		// Indicates whether the user wants to search for batch torrents.
		// This will be false if your extension does not support batch torrents.
		Batch bool `json:"batch"`
		// The episode number the user wants to search for.
		// This will be 0 if your extension does not support episode number filtering.
		EpisodeNumber int `json:"episodeNumber"`
		// The resolution the user wants to search for.
		// This will be empty if your extension does not support resolution filtering.
		// e.g. "1080", "720"
		Resolution string `json:"resolution"`
		// AniDB Anime ID of the media.
		AnidbAID int `json:"anidbAID"`
		// AniDB Episode ID of the media.
		AnidbEID int `json:"anidbEID"`
		// Indicates whether the user wants to search for the best releases.
		// This will be false if your extension does not support filtering by best releases.
		BestReleases bool `json:"bestReleases"`
	}

	AnimeTorrent struct {
		// "ID" of the extension.
		Provider string `json:"provider,omitempty"`
		// Title of the torrent.
		Name string `json:"name"`
		// Date of the torrent.
		// The date should have RFC3339 format. e.g. "2006-01-02T15:04:05Z07:00"
		Date string `json:"date"`
		// Size of the torrent in bytes.
		Size int64 `json:"size"`
		// Formatted size of the torrent. e.g. "1.2 GB"
		// Leave this empty if you want Seanime to format the size.
		FormattedSize string `json:"formattedSize"`
		// Number of seeders.
		Seeders int `json:"seeders"`
		// Number of leechers.
		Leechers int `json:"leechers"`
		// Number of downloads.
		DownloadCount int `json:"downloadCount"`
		// Link to the torrent page.
		Link string `json:"link"`
		// Download URL of the torrent.
		// Leave this empty if you cannot provide a direct download URL.
		DownloadUrl string `json:"downloadUrl"`
		// Magnet link of the torrent.
		// Leave this empty if you cannot provide a magnet link without scraping.
		MagnetLink string `json:"magnetLink,omitempty"`
		// InfoHash of the torrent.
		// Leave empty if it should be scraped later.
		InfoHash string `json:"infoHash,omitempty"`
		// Resolution of the video.
		// e.g. "1080p", "720p"
		Resolution string `json:"resolution,omitempty"`
		// Set this to true if you can confirm that the torrent is a batch.
		// Else, Seanime will parse the torrent name to determine if it's a batch.
		IsBatch bool `json:"isBatch,omitempty"`
		// Episode number of the torrent.
		// Return -1 if unknown / unable to determine and Seanime will parse the torrent name.
		EpisodeNumber int `json:"episodeNumber,omitempty"`
		// Release group of the torrent.
		// Leave this empty if you want Seanime to parse the release group from the name.
		ReleaseGroup string `json:"releaseGroup,omitempty"`
		// Set this to true if you can confirm that the torrent is the best release.
		IsBestRelease bool `json:"isBestRelease"`
		// Set this to true if you can confirm that the torrent matches the anime the user is searching for.
		// e.g. If the torrent was found using the AniDB anime or episode ID
		Confirmed bool `json:"confirmed"`
	}
)
