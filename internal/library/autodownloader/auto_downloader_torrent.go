package autodownloader

import (
	"github.com/seanime-app/seanime/internal/torrents/animetosho"
	"github.com/seanime-app/seanime/internal/torrents/nyaa"
	"github.com/seanime-app/seanime/internal/torrents/torrent"
	"github.com/seanime-app/seanime/seanime-parser"
	"strconv"
)

type (
	// NormalizedTorrent is a struct built from torrent from a provider.
	// It is used to normalize the data from different providers so that it can be used by the AutoDownloader.
	NormalizedTorrent struct {
		Name       string
		Link       string
		Hash       string
		Size       int64
		Seeders    int
		ParsedData *seanime_parser.Metadata
		Provider   string
		magnet     string // Access using GetMagnet()
	}
)

func (ad *AutoDownloader) getCurrentTorrentsFromNyaa() ([]*NormalizedTorrent, error) {
	ad.logger.Debug().Msg("autodownloader: Checking for new episodes from Nyaa")

	// Fetch the RSS feed
	torrents, err := nyaa.GetTorrentList(nyaa.SearchOptions{
		Provider: "nyaa",
		Query:    "",
		Category: "anime-eng",
		SortBy:   "seeders",
		Filter:   "",
	})
	if err != nil {
		return nil, err
	}

	normalizedTs := make([]*NormalizedTorrent, 0)
	for _, t := range torrents {
		parsedData := seanime_parser.Parse(t.Name)
		if err != nil {
			ad.logger.Error().Err(err).Msg("autodownloader: Failed to parse torrent title")
			continue
		}

		seedersInt := 0
		if t.Seeders != "" {
			seedersInt, _ = strconv.Atoi(t.Seeders)
		}

		normalizedTs = append(normalizedTs, &NormalizedTorrent{
			Name:       t.Name,
			Link:       t.GUID,
			Hash:       t.InfoHash,
			Size:       t.GetSizeInBytes(),
			Seeders:    seedersInt,
			magnet:     "", // Nyaa doesn't provide the magnet link in the RSS feed
			ParsedData: parsedData,
			Provider:   torrent.ProviderNyaa,
		})
	}

	return normalizedTs, nil
}

func (ad *AutoDownloader) getCurrentTorrentsFromAnimeTosho() ([]*NormalizedTorrent, error) {
	ad.logger.Debug().Msg("autodownloader: Checking for new episodes from AnimeTosho")
	normalizedTs := make([]*NormalizedTorrent, 0)

	// Fetch the latest torrents
	torrents, err := animetosho.GetLatest()
	if err != nil {
		return nil, err
	}

	for _, t := range torrents {
		parsedData := seanime_parser.Parse(t.Title)
		if err != nil {
			ad.logger.Error().Err(err).Msg("autodownloader: Failed to parse torrent title")
			continue
		}

		normalizedTs = append(normalizedTs, &NormalizedTorrent{
			Name:       t.Title,
			Link:       t.Link,
			Hash:       t.InfoHash,
			Size:       t.TotalSize,
			Seeders:    t.Seeders,
			magnet:     t.MagnetUrl, // AnimeTosho doesn't seem to provide the magnet link for newer torrents
			ParsedData: parsedData,
			Provider:   torrent.ProviderAnimeTosho,
		})
	}

	return normalizedTs, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (t *NormalizedTorrent) GetMagnet() (string, bool) {
	if t.magnet != "" {
		return t.magnet, true
	}

	// Scrape the page to get the magnet link
	if t.Provider == torrent.ProviderNyaa {
		magnet, err := nyaa.TorrentMagnet(t.Link)
		if err != nil {
			return "", false
		}
		return magnet, true
	} else if t.Provider == torrent.ProviderAnimeTosho {
		magnet, err := animetosho.TorrentMagnet(t.Link)
		if err != nil {
			return "", false
		}
		return magnet, true
	}

	return "", false
}
