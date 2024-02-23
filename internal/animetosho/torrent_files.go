package animetosho

import (
	"errors"
	"github.com/gocolly/colly"
	"strings"
)

func TorrentFile(viewURL string) (string, error) {
	var torrentLink string

	c := colly.NewCollector()

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		if strings.HasSuffix(e.Attr("href"), ".torrent") {
			torrentLink = e.Attr("href")
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

	if torrentLink == "" {
		return "", errors.New("download link not found")
	}

	return torrentLink, nil
}
