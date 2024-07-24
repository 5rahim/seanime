package animetosho

import (
	"fmt"
	hibiketorrent "github.com/5rahim/hibike/pkg/extension/torrent"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
	"io"
	"net/http"
	"net/url"
	"seanime/internal/util"
	"seanime/internal/util/comparison"
	"seanime/seanime-parser"
	"strings"
	"sync"
	"time"
)

const (
	JsonFeedUrl        = "https://feed.animetosho.org/json"
	ProviderAnimeTosho = "animetosho"
)

type Provider struct {
	logger *zerolog.Logger
}

func NewProvider(logger *zerolog.Logger) hibiketorrent.Provider {
	return &Provider{
		logger: logger,
	}
}

// todo: hibike - add GetLatest method

func (at *Provider) Search(opts hibiketorrent.SearchOptions) (ret []*hibiketorrent.AnimeTorrent, err error) {
	query := fmt.Sprintf("?qx=1&q=%s&filter[0][t]=nyaa_class&order=", url.QueryEscape(sanitizeTitle(opts.Query)))
	torrents, err := fetchTorrents(query)
	if err != nil {
		return nil, err
	}

	wg := sync.WaitGroup{}
	mu := sync.Mutex{}

	for _, t := range torrents {
		wg.Add(1)
		go func(t *Torrent) {
			defer wg.Done()
			mu.Lock()
			ret = append(ret, t.toAnimeTorrent())
			mu.Unlock()
		}(t)
	}

	wg.Wait()

	return ret, nil
}

func (at *Provider) SmartSearch(opts hibiketorrent.SmartSearchOptions) ([]*hibiketorrent.AnimeTorrent, error) {
	//TODO implement me
	panic("implement me")
}

func (at *Provider) GetTorrentInfoHash(torrent *hibiketorrent.AnimeTorrent) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (at *Provider) GetTorrentMagnetLink(torrent *hibiketorrent.AnimeTorrent) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (at *Provider) CanSmartSearch() bool {
	//TODO implement me
	panic("implement me")
}

func (at *Provider) CanFindBestRelease() bool {
	//TODO implement me
	panic("implement me")
}

func (at *Provider) SupportsAdult() bool {
	//TODO implement me
	panic("implement me")
}

// GetLatest returns all the latest torrents currently visible on the site
func GetLatest() (ret []*hibiketorrent.AnimeTorrent, err error) {
	query := "?qx=1&q=&filter[0][t]=nyaa_class&order="
	torrents, err := fetchTorrents(query)
	if err != nil {
		return nil, err
	}

	wg := sync.WaitGroup{}
	mu := sync.Mutex{}

	for _, t := range torrents {
		wg.Add(1)
		go func(t *Torrent) {
			defer wg.Done()
			mu.Lock()
			ret = append(ret, t.toAnimeTorrent())
			mu.Unlock()
		}(t)
	}

	wg.Wait()

	return ret, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// formatCommonQuery adds special query filters
func formatCommonQuery(quality string) string {
	quality = strings.TrimSuffix(quality, "p")
	if quality == "1080" {
		return `("1080" !"720" !"540" !"480")`
	} else if quality == "720" {
		return `("720" !"1080" !"540" !"480")`
	} else if quality == "540" {
		return `("540" !"1080" !"720" !"480")`
	} else if quality == "480" {
		return `("480" !"1080" !"720" !"540")`
	} else {
		return ``
	}
}

// searches for torrents by Anime ID
func searchByAID(aid int, quality string) (torrents []*Torrent, err error) {
	q := url.QueryEscape(formatCommonQuery(quality))
	query := fmt.Sprintf(`?qx=1&order=size-d&aid=%d&q=%s`, aid, q)
	return fetchTorrents(query)
}

// searches for torrents by Episode ID
func searchByEID(eid int, quality string) (torrents []*Torrent, err error) {
	q := url.QueryEscape(formatCommonQuery(quality))
	query := fmt.Sprintf(`?qx=1&eid=%d&q=%s`, eid, q)
	return fetchTorrents(query)
}

// sanitizeTitle removes characters that impact the search query
func sanitizeTitle(t string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(t, "!", ""), ":", ""), "[", ""), "]", "")
}

func fetchTorrents(query string) (torrents []*Torrent, err error) {

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

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (t *Torrent) toAnimeTorrent() *hibiketorrent.AnimeTorrent {
	metadata := seanime_parser.Parse(t.Title)

	formattedDate := ""
	parsedDate := time.Unix(int64(t.Timestamp), 0)
	formattedDate = parsedDate.Format(time.RFC3339)

	ret := &hibiketorrent.AnimeTorrent{
		Name:          t.Title,
		Date:          formattedDate,
		Size:          t.TotalSize,
		FormattedSize: util.ToHumanReadableSize(t.TotalSize),
		Seeders:       t.Seeders,
		Leechers:      t.Leechers,
		DownloadCount: t.TorrentDownloadCount,
		Link:          t.Link,
		DownloadUrl:   t.TorrentUrl,
		InfoHash:      t.InfoHash,
		Provider:      ProviderAnimeTosho,
		IsBatch:       t.NumFiles > 1,
	}

	isBatchByGuess := false
	episode := -1

	if len(metadata.EpisodeNumber) > 1 || comparison.ValueContainsBatchKeywords(t.Title) {
		isBatchByGuess = true
	}
	if len(metadata.EpisodeNumber) == 1 {
		episode = util.StringToIntMust(metadata.EpisodeNumber[0])
	}

	ret.Resolution = metadata.VideoResolution
	ret.ReleaseGroup = metadata.ReleaseGroup

	// Only change batch status if it wasn't already 'true'
	if ret.IsBatch == false && isBatchByGuess {
		ret.IsBatch = true
	}

	ret.EpisodeNumber = episode

	return ret
}
