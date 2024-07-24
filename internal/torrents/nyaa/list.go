package nyaa

import (
	"github.com/mmcdole/gofeed"
)

func GetTorrentList(opts BuildURLOptions) ([]Torrent, error) {

	fp := gofeed.NewParser()

	// create search url
	url, err := buildURL(opts)
	if err != nil {
		return nil, err
	}

	// get content
	feed, err := fp.ParseURL(url)
	if err != nil {
		return nil, err
	}

	// parse content
	res := convertRSS(feed)

	return res, nil
}
