package torrent

import (
	"errors"
	"github.com/seanime-app/seanime/internal/animetosho"
	"github.com/seanime-app/seanime/internal/nyaa"
	"regexp"
	"strings"
)

// GetTorrentMagnetFromUrl will return the magnet link of a torrent from its page URL.
// It supports Nyaa and AnimeTosho.
func GetTorrentMagnetFromUrl(viewUrl string) (string, error) {

	str := strings.ToLower(viewUrl)

	if strings.Contains(str, "nyaa.si") {
		return nyaa.TorrentMagnet(viewUrl)
	} else if strings.Contains(str, "animetosho.org") {
		return animetosho.TorrentMagnet(viewUrl)
	}

	return "", errors.New("could not determine the torrent provider from the URL")
}

func ExtractHashFromMagnet(magnetLink string) (string, bool) {
	re := regexp.MustCompile(`magnet:\?xt=urn:btih:([^&]+)`)
	match := re.FindStringSubmatch(magnetLink)
	if len(match) > 1 {
		return match[1], true
	} else {
		return "", false // Magnet link format not recognized or no hash found
	}
}
