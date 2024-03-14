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

func TorrentHash(viewURL string) (string, error) {

	file, err := TorrentFile(viewURL)
	if err != nil {
		return "", err
	}

	file = strings.Replace(file, "https://", "", 1)
	//template := "%s/storage/torrent/%s/%s"
	parts := strings.Split(file, "/")

	if len(parts) < 4 {
		return "", errors.New("hash not found")
	}

	return parts[3], nil
}
