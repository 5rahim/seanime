package torrent

import "regexp"

func ExtractHashFromMagnet(magnetLink string) (string, bool) {
	re := regexp.MustCompile(`magnet:\?xt=urn:btih:([^&]+)`)
	match := re.FindStringSubmatch(magnetLink)
	if len(match) > 1 {
		return match[1], true
	} else {
		return "", false // Magnet link format not recognized or no hash found
	}
}
