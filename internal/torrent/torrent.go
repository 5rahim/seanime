package torrent

import (
	"github.com/seanime-app/seanime/internal/animetosho"
	"github.com/seanime-app/seanime/internal/comparison"
	"github.com/seanime-app/seanime/internal/nyaa"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/seanime-app/seanime/seanime-parser"
	"strconv"
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
		IsBatch       bool   `json:"isBatch,omitempty"`
		EpisodeNumber int    `json:"episodeNumber,omitempty"`
		ReleaseGroup  string `json:"releaseGroup,omitempty"`
	}
)

func NewAnimeTorrentFromNyaa(torrent *nyaa.Torrent) *AnimeTorrent {
	metadata := seanime_parser.Parse(torrent.Name)

	seeders, _ := strconv.Atoi(torrent.Seeders)
	leechers, _ := strconv.Atoi(torrent.Leechers)
	downloads, _ := strconv.Atoi(torrent.Downloads)

	t := &AnimeTorrent{
		Name:          torrent.Name,
		Date:          torrent.Date,
		Size:          torrent.GetSizeInBytes(),
		FormattedSize: torrent.Size,
		Seeders:       seeders,
		Leechers:      leechers,
		DownloadCount: downloads,
		Link:          torrent.GUID,
		DownloadUrl:   torrent.Link,
		InfoHash:      torrent.InfoHash,
	}

	hydrateMetadata(t, metadata)

	return t
}

func NewAnimeTorrentFromAnimeTosho(torrent *animetosho.Torrent) *AnimeTorrent {
	metadata := seanime_parser.Parse(torrent.Title)

	t := &AnimeTorrent{
		Name:          torrent.Title,
		Date:          util.TimestampToDateStr(int64(torrent.Timestamp)),
		Size:          torrent.TotalSize,
		FormattedSize: util.ToHumanReadableSize(int(torrent.TotalSize)),
		Seeders:       torrent.Seeders,
		Leechers:      torrent.Leechers,
		DownloadCount: torrent.TorrentDownloadCount,
		Link:          torrent.Link,
		DownloadUrl:   torrent.TorrentUrl,
		InfoHash:      torrent.InfoHash,
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
