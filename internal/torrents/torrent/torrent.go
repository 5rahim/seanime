package torrent

import (
	"context"
	"github.com/seanime-app/seanime/internal/torrents/animetosho"
	"github.com/seanime-app/seanime/internal/torrents/nyaa"
	"github.com/seanime-app/seanime/internal/torrents/seadex"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/seanime-app/seanime/internal/util/comparison"
	"github.com/seanime-app/seanime/seanime-parser"
	"net/http"
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
		EpisodeNumber int    `json:"episodeNumber"`
		ReleaseGroup  string `json:"releaseGroup,omitempty"`
		Provider      string `json:"provider,omitempty"`
		IsBestRelease bool   `json:"isBestRelease"`
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

func NewAnimeTorrentFromSeaDex(torrent *seadex.Torrent) *AnimeTorrent {

	var seeders, leechers, downloads int
	var downloadUrl, title string

	t := &AnimeTorrent{
		Name:          torrent.Name,
		Date:          torrent.Date,
		Size:          torrent.Size,
		FormattedSize: util.ToHumanReadableSize(int(torrent.Size)),
		Seeders:       seeders,
		Leechers:      leechers,
		DownloadCount: downloads,
		Link:          torrent.Link,
		DownloadUrl:   downloadUrl,
		InfoHash:      torrent.InfoHash,
		Provider:      "seadex",
		Resolution:    "",
		IsBatch:       true,
		EpisodeNumber: 0,
		ReleaseGroup:  torrent.ReleaseGroup,
		IsBestRelease: true,
	}

	// Try scraping from Nyaa
	// Since nyaa tends to be blocked, try for a few seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if torrent.Link != "" {
		downloadUrl = torrent.Link

		client := http.DefaultClient
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, torrent.Link, nil)
		if err == nil {
			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				title, seeders, leechers, downloads, _, _, err = nyaa.TorrentInfo(torrent.Link + "fail")
				if err == nil && title != "" {
					t.Name = title
					t.Seeders = seeders
					t.Leechers = leechers
					t.DownloadCount = downloads
					t.DownloadUrl = downloadUrl

					hydrateMetadata(t, seanime_parser.Parse(title))
				}
			}
		}
	}

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
