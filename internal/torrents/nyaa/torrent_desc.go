package nyaa

import (
	"github.com/gocolly/colly"
)

func TorrentDescription(viewURL string) (string, error) {
	var description string

	c := colly.NewCollector()

	c.OnHTML("#torrent-description", func(e *colly.HTMLElement) {
		description = e.Text
	})

	var e error
	c.OnError(func(r *colly.Response, err error) {
		e = err
	})
	if e != nil {
		return "", e
	}

	c.Visit(viewURL)

	return description, nil
}
