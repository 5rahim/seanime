package autodownloader

import (
	"github.com/seanime-app/seanime/internal/nyaa"
	"github.com/seanime-app/seanime/seanime-parser"
	"strconv"
)

type (
	NormalizedTorrent struct {
		Name       string
		Link       string
		Hash       string
		Size       string
		Seeders    int
		ParsedInfo *seanime_parser.Metadata
		Provider   string
		magnet     string // Access using GetMagnet()
	}
)

func (ad *AutoDownloader) getCurrentTorrentsFromNyaa() ([]*NormalizedTorrent, error) {
	ad.Logger.Debug().Msg("autodownloader: Checking for new episodes from Nyaa")

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
		parsedInfo := seanime_parser.Parse(t.Name)
		if err != nil {
			ad.Logger.Error().Err(err).Msg("autodownloader: Failed to parse torrent title")
			continue
		}

		seedersInt := 0
		if t.Seeders != "" {
			seedersInt, _ = strconv.Atoi(t.Seeders)
		}

		normalizedTs = append(normalizedTs, &NormalizedTorrent{
			Name:       t.Name,
			Link:       t.Link,
			Hash:       t.InfoHash,
			Size:       t.Size,
			Seeders:    seedersInt,
			magnet:     "", // Nyaa doesn't provide the magnet link in the RSS feed
			ParsedInfo: parsedInfo,
			Provider:   NyaaProvider,
		})
	}

	return normalizedTs, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (t *NormalizedTorrent) GetMagnet() (string, bool) {
	if t.magnet != "" {
		return t.magnet, true
	}

	if t.Provider == NyaaProvider {
		// Fetch the view page to get the magnet link
		magnet, err := nyaa.TorrentMagnet(t.Link)
		if err != nil {
			return "", false
		}
		return magnet, true
	}

	return "", false
}
