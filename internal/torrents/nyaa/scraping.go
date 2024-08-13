package nyaa

import (
	"errors"
	"github.com/gocolly/colly"
	"regexp"
	"strconv"
	"strings"
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

func TorrentInfo(viewURL string) (title string, seeders int, leechers int, completed int, formattedSize string, infoHash string, magnetLink string, err error) {

	c := colly.NewCollector()

	c.OnHTML("a.card-footer-item", func(e *colly.HTMLElement) {
		magnetLink = e.Attr("href")
	})

	c.OnHTML(".panel-title", func(e *colly.HTMLElement) {
		if title == "" {
			title = strings.TrimSpace(e.Text)
		}
	})

	// Find and extract information from the specified div elements
	c.OnHTML(".panel-body", func(e *colly.HTMLElement) {

		if seeders == 0 {
			// Extract seeders
			e.ForEach("div:contains('Seeders:') span", func(_ int, el *colly.HTMLElement) {
				if el.Attr("style") == "color: green;" {
					seeders, _ = strconv.Atoi(el.Text)
				}
			})
		}

		if leechers == 0 {
			// Extract leechers
			e.ForEach("div:contains('Leechers:') span", func(_ int, el *colly.HTMLElement) {
				if el.Attr("style") == "color: red;" {
					leechers, _ = strconv.Atoi(el.Text)
				}
			})
		}

		if completed == 0 {
			// Extract completed
			e.ForEach("div:contains('Completed:')", func(_ int, el *colly.HTMLElement) {
				completed, _ = strconv.Atoi(el.DOM.Parent().Find("div").Next().Next().Next().Text())
			})
		}

		if formattedSize == "" {
			// Extract completed
			e.ForEach("div:contains('File size:')", func(_ int, el *colly.HTMLElement) {
				text := el.DOM.Parent().ChildrenFiltered("div:nth-child(2)").Text()
				if !strings.Contains(text, "\t") {
					formattedSize = text
				}
			})
		}

		if infoHash == "" {
			// Extract info hash
			e.ForEach("div:contains('Info hash:') kbd", func(_ int, el *colly.HTMLElement) {
				infoHash = el.Text
			})
		}
	})

	var e error
	c.OnError(func(r *colly.Response, err error) {
		e = err
	})
	if e != nil {
		err = e
		return
	}

	_ = c.Visit(viewURL)

	if magnetLink == "" {
		err = errors.New("magnet link not found")
		return
	}

	return
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
