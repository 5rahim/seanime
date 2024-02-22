package animetosho

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/goccy/go-json"
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
	SearchResult struct {
		Title      string
		URL        string
		MagnetURL  string
		TorrentURL string
	}

	Torrent struct {
		Id                   int         `json:"id"`
		Title                string      `json:"title"`
		Link                 string      `json:"link"`
		Timestamp            int         `json:"timestamp"`
		Status               string      `json:"status"`
		ToshoId              int         `json:"tosho_id,omitempty"`
		NyaaId               int         `json:"nyaa_id,omitempty"`
		NyaaSubdom           interface{} `json:"nyaa_subdom,omitempty"`
		AniDexId             int         `json:"anidex_id,omitempty"`
		TorrentUrl           string      `json:"torrent_url"`
		InfoHash             string      `json:"info_hash"`
		InfoHashV2           string      `json:"info_hash_v2,omitempty"`
		MagnetUrl            string      `json:"magnet_url"`
		Seeders              int         `json:"seeders"`
		Leechers             int         `json:"leechers"`
		TorrentDownloadCount int         `json:"torrent_download_count"`
		TrackerUpdated       interface{} `json:"tracker_updated,omitempty"`
		NzbUrl               string      `json:"nzb_url,omitempty"`
		TotalSize            int64       `json:"total_size"`
		NumFiles             int         `json:"num_files"`
		AniDbAid             int         `json:"anidb_aid"`
		AniDbEid             int         `json:"anidb_eid"`
		AniDbFid             int         `json:"anidb_fid"`
		ArticleUrl           string      `json:"article_url"`
		ArticleTitle         string      `json:"article_title"`
		WebsiteUrl           string      `json:"website_url"`
	}
)

func GetLatest() (torrents []*Torrent, err error) {
	query := "?only_tor=1&q=&filter[0][t]=nyaa_class&order="
	return fetchTorrents(query)
}

func Search(show string) (torrents []*Torrent, err error) {

	//format := "%s?only_tor=1&q=%s&filter[0][t]=nyaa_class&filter[0][v]=trusted"
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

func SearchRSS(terms string) ([]SearchResult, error) {
	var (
		err error

		searchURL *url.URL
		resp      *http.Response
		doc       *goquery.Document
	)

	searchURL, err = url.Parse(SearchUrl)
	if err != nil {
		return nil, err
	}

	qs := searchURL.Query()
	qs.Set("q", terms)

	searchURL.RawQuery = qs.Encode()

	resp, err = http.Get(searchURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err = goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var results []SearchResult
	doc.Find(".home_list_entry").Each(func(i int, entry *goquery.Selection) {
		var (
			err error

			infoURL    *url.URL
			magnetURL  *url.URL
			torrentURL *url.URL
		)

		titleSel := entry.Find(".link a").First()
		if titleSel.Length() != 1 {
			return
		}

		infoLink := titleSel.AttrOr("href", "")
		if infoLink == "" {
			return
		}

		infoURL, err = searchURL.Parse(infoLink)
		if err != nil {
			return
		}

		title := strings.TrimSpace(titleSel.Text())
		if title == "" {
			return
		}

		dlLink := entry.Find(".dllink").First().AttrOr("href", "")
		if dlLink == "" {
			return
		}

		torrentURL, err = searchURL.Parse(dlLink)
		if err != nil {
			return
		}

		magnetLink := entry.Find(`a[href^="magnet:"]`).First().AttrOr("href", "")
		if magnetLink == "" {
			return
		}

		magnetURL, err = searchURL.Parse(magnetLink)
		if err != nil {
			return
		}

		result := SearchResult{
			Title:      title,
			URL:        infoURL.String(),
			MagnetURL:  magnetURL.String(),
			TorrentURL: torrentURL.String(),
		}

		results = append(results, result)
	})

	return results, nil
}
