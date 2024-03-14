package nyaa

import (
	"errors"
	"github.com/gocolly/colly"
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
