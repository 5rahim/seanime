package torrent

import (
	"github.com/seanime-app/seanime/internal/torrents/animetosho"
	"github.com/seanime-app/seanime/internal/torrents/nyaa"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/seanime-app/seanime/internal/util/comparison"
	"github.com/seanime-app/seanime/seanime-parser"
	"strconv"
	"time"
)

type (
	AnimeTorrent struct {
		Name          string `json:"name"`
		Date          string `json:"date"`
		Size          int64  `json:"size"`
		FormattedSize string `json:"formattedSize"`
		Seeders       int    `json:"seeders"`
		Leechers      int    `json:"leechers"`
		DownloadCount int    `json:"downloadCount"`
		Link          string `json:"link"`
		DownloadUrl   string `json:"downloadUrl"`
		InfoHash      string `json:"infoHash"`
		Resolution    string `json:"resolution,omitempty"`
		IsBatch       bool   `json:"isBatch"`
		EpisodeNumber int    `json:"episodeNumber,omitempty"`
		ReleaseGroup  string `json:"releaseGroup,omitempty"`
		Provider      string `json:"provider,omitempty"`
	}
)

func NewAnimeTorrentFromNyaa(torrent *nyaa.DetailedTorrent) *AnimeTorrent {
	metadata := seanime_parser.Parse(torrent.Name)

	seeders, _ := strconv.Atoi(torrent.Seeders)
	leechers, _ := strconv.Atoi(torrent.Leechers)
	downloads, _ := strconv.Atoi(torrent.Downloads)

	formattedDate := ""
	parsedDate, err := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", torrent.Date)
	if err == nil {
		formattedDate = parsedDate.Format(time.RFC3339)
	}

	t := &AnimeTorrent{
		Name:          torrent.Name,
		Date:          formattedDate,
		Size:          torrent.GetSizeInBytes(),
		FormattedSize: torrent.Size,
		Seeders:       seeders,
		Leechers:      leechers,
		DownloadCount: downloads,
		Link:          torrent.GUID,
		DownloadUrl:   torrent.Link,
		InfoHash:      torrent.InfoHash,
		Provider:      "nyaa",
	}

	hydrateMetadata(t, metadata)

	return t
}

func NewAnimeTorrentFromAnimeTosho(torrent *animetosho.Torrent) *AnimeTorrent {
	metadata := seanime_parser.Parse(torrent.Title)

	formattedDate := ""
	parsedDate := time.Unix(int64(torrent.Timestamp), 0)
	formattedDate = parsedDate.Format(time.RFC3339)

	t := &AnimeTorrent{
		Name:          torrent.Title,
		Date:          formattedDate,
		Size:          torrent.TotalSize,
		FormattedSize: util.ToHumanReadableSize(int(torrent.TotalSize)),
		Seeders:       torrent.Seeders,
		Leechers:      torrent.Leechers,
		DownloadCount: torrent.TorrentDownloadCount,
		Link:          torrent.Link,
		DownloadUrl:   torrent.TorrentUrl,
		InfoHash:      torrent.InfoHash,
		Provider:      "animetosho",
	}

	hydrateMetadata(t, metadata)

	return t
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func hydrateMetadata(t *AnimeTorrent, metadata *seanime_parser.Metadata) {
	if metadata == nil {
		return
	}

	isBatch := false
	episode := -1

	if len(metadata.EpisodeNumber) > 1 || comparison.ValueContainsBatchKeywords(t.Name) {
		isBatch = true
	}
	if len(metadata.EpisodeNumber) == 1 {
		episode, _ = util.StringToInt(metadata.EpisodeNumber[0])
	}

	t.Resolution = metadata.VideoResolution
	t.ReleaseGroup = metadata.ReleaseGroup
	t.IsBatch = isBatch
	t.EpisodeNumber = episode

	return
}
