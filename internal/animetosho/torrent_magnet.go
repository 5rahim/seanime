package animetosho

import (
	"errors"
	"github.com/gocolly/colly"
	"strings"
)

func TorrentMagnet(viewURL string) (string, error) {
	var magnetLink string

	c := colly.NewCollector()

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		if strings.HasPrefix(e.Attr("href"), "magnet:?xt=") {
			magnetLink = e.Attr("href")
		}
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
