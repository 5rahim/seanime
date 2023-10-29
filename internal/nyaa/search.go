package nyaa

import (
	"github.com/mmcdole/gofeed"
)

type SearchOptions struct {
	Provider string
	Query    string
	Category string
	SortBy   string
	Filter   string
}

var (
	fp = gofeed.NewParser()
)

func Search(opts SearchOptions) ([]Torrent, error) {
	url, err := buildURL(opts)
	if err != nil {
		return nil, err
	}

	feed, err := fp.ParseURL(url)
	if err != nil {
		return nil, err
	}

	res := convertRSS(feed)

	return res, nil
}
