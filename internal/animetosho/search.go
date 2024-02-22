package animetosho

import (
	"fmt"
	"github.com/goccy/go-json"
	"io"
	"net/http"
	"net/url"
)

const (
	SearchUrl   = "https://animetosho.org/search"
	FeedUrl     = "https://feed.animetosho.org/rss2"
	JsonFeedUrl = "https://feed.animetosho.org/json"
)

func GetLatest() (torrents []*Torrent, err error) {
	query := "?only_tor=1&q=&filter[0][t]=nyaa_class&order="
	return fetchTorrents(query)
}

func Search(show string) (torrents []*Torrent, err error) {
	format := "?only_tor=1&q=%s&filter[0][t]=nyaa_class&order="
	query := fmt.Sprintf(format, url.QueryEscape(show))
	return fetchTorrents(query)
}

func fetchTorrents(query string) (torrents []*Torrent, err error) {

	//format := "%s?only_tor=1&q=%s&filter[0][t]=nyaa_class&filter[0][v]=trusted"
	//format := "%s?only_tor=1&q=%s&filter[0][t]=nyaa_class&order="
	furl := JsonFeedUrl + query
	resp, err := http.Get(furl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check if the request was successful (status code 200)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch torrents, %s", resp.Status)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse the feed
	var ret []*Torrent
	if err := json.Unmarshal(b, &ret); err != nil {
		return nil, err
	}

	for _, t := range ret {
		if t.Seeders > 30000 {
			t.Seeders = 0
		}
		if t.Leechers > 30000 {
			t.Leechers = 0
		}
	}

	return ret, nil
}
