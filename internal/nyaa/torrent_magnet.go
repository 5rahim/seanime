package nyaa

import (
	"errors"
	"github.com/gocolly/colly"
	"regexp"
)

func TorrentMagnet(viewURL string) (string, error) {
	var magnetLink string

	c := colly.NewCollector()

	c.OnHTML("a.card-footer-item", func(e *colly.HTMLElement) {
		magnetLink = e.Attr("href")
	})

	var e error
	c.OnError(func(r *colly.Response, err error) {
		e = err
	})
	if e != nil {
		return "", e
	}

	c.Visit(viewURL)

	if magnetLink == "" {
		return "", errors.New("magnet link not found")
	}

	return magnetLink, nil
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
