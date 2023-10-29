package nyaa

import (
	"github.com/mmcdole/gofeed"
)

// https://github.com/irevenko/go-nyaa

type SearchOptions struct {
	Provider string
	Query    string
	Category string
	SortBy   string
	Filter   string
}

func Search(opts SearchOptions) ([]Torrent, error) {

	fp := gofeed.NewParser()

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
