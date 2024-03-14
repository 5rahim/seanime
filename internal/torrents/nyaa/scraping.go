package nyaa

import (
	"errors"
	"github.com/gocolly/colly"
	"regexp"
)

func TorrentFiles(viewURL string) ([]string, error) {
	var folders []string
	var files []string

	c := colly.NewCollector()

	c.OnHTML(".folder", func(e *colly.HTMLElement) {
		folders = append(folders, e.Text)
	})

	c.OnHTML(".torrent-file-list", func(e *colly.HTMLElement) {
		files = append(files, e.ChildText("li"))
	})

	var e error
	c.OnError(func(r *colly.Response, err error) {
		e = err
	})
	if e != nil {
		return nil, e
	}

	c.Visit(viewURL)

	if len(folders) == 0 {
		return files, nil
	}

	return folders, nil
}

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

func TorrentHash(viewURL string) (string, error) {
	magnet, err := TorrentMagnet(viewURL)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`magnet:\?xt=urn:btih:([^&]+)`)
	match := re.FindStringSubmatch(magnet)
	if len(match) > 1 {
		return match[1], nil
	}
	return "", errors.New("could not extract hash")
}
