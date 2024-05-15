package animetosho

import (
	"fmt"
	"github.com/goccy/go-json"
	"github.com/seanime-app/seanime/internal/util/result"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	SearchUrl   = "https://animetosho.org/search"
	FeedUrl     = "https://feed.animetosho.org/rss2"
	JsonFeedUrl = "https://feed.animetosho.org/json"
)

type (
	SearchCache struct {
		*result.Cache[string, []*Torrent]
	}
)

func NewSearchCache() *SearchCache {
	return &SearchCache{result.NewCache[string, []*Torrent]()}
}

func GetLatest() (torrents []*Torrent, err error) {
	query := "?qx=1&q=&filter[0][t]=nyaa_class&order="
	return fetchTorrents(query)
}

func Search(show string) (torrents []*Torrent, err error) {
	format := "?qx=1&q=%s&filter[0][t]=nyaa_class&order="
	query := fmt.Sprintf(format, url.QueryEscape(show))
	return fetchTorrents(query)
}

func formatCommonQuery(quality string) string {
	quality = strings.TrimSuffix(quality, "p")
	if quality == "1080" {
		return `((e*|a*|r*|i*|o*|"1080") !"720" !"540" !"480")`
	} else if quality == "720" {
		return `((e*|a*|r*|i*|o*|"720") !"1080" !"540" !"480")`
	} else if quality == "540" {
		return `((e*|a*|r*|i*|o*|"540") !"1080" !"720" !"480")`
	} else if quality == "480" {
		return `((e*|a*|r*|i*|o*|"480") !"1080" !"720" !"540")`
	} else {
		return `(e*|a*|r*|i*|o*)`
	}
}

func SearchByAID(aid int, quality string) (torrents []*Torrent, err error) {
	q := url.QueryEscape(formatCommonQuery(quality))
	query := fmt.Sprintf(`?qx=1&order=size-d&aid=%d&q=%s`, aid, q)
	return fetchTorrents(query)
}

func SearchByAIDLikelyBatch(aid int, quality string) (torrents []*Torrent, err error) {
	q := url.QueryEscape(formatCommonQuery(quality))
	query := fmt.Sprintf(`?qx=1&order=size-d&aid=%d&q=%s`, aid, q)
	return fetchTorrents(query)
}

func SearchByEID(eid int, quality string) (torrents []*Torrent, err error) {
	q := url.QueryEscape(formatCommonQuery(quality))
	query := fmt.Sprintf(`?qx=1&eid=%d&q=%s`, eid, q)
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
