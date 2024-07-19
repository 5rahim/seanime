package torrent

import (
	"errors"
	"seanime/internal/torrents/animetosho"
	"seanime/internal/torrents/nyaa"
	"strings"
)

// ScrapeMagnet will return the magnet link of a torrent from its page URL.
// It supports Nyaa and AnimeTosho.
func ScrapeMagnet(viewUrl string) (string, error) {

	str := strings.ToLower(viewUrl)

	if strings.Contains(str, "nyaa.si") {
		return nyaa.TorrentMagnet(viewUrl)
	} else if strings.Contains(str, "animetosho.org") {
		return animetosho.TorrentMagnet(viewUrl)
	}

	return "", errors.New("could not determine the torrent provider from the URL")
}

func ScrapeHash(viewUrl string) (string, error) {
	if strings.Contains(viewUrl, "nyaa.si") {
		return nyaa.TorrentHash(viewUrl)
	} else if strings.Contains(viewUrl, "animetosho.org") {
		return animetosho.TorrentHash(viewUrl)
	}
	return "", errors.New("could not determine the torrent provider from the URL")
}
