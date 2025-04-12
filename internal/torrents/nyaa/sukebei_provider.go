package nyaa

import (
	"github.com/mmcdole/gofeed"
	"github.com/rs/zerolog"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"sync"
)

const (
	SukebeiProviderName = "nyaa-sukebei"
)

type SukebeiProvider struct {
	logger *zerolog.Logger
}

func NewSukebeiProvider(logger *zerolog.Logger) hibiketorrent.AnimeProvider {
	return &SukebeiProvider{
		logger: logger,
	}
}

func (n *SukebeiProvider) GetSettings() hibiketorrent.AnimeProviderSettings {
	return hibiketorrent.AnimeProviderSettings{
		Type:           hibiketorrent.AnimeProviderTypeSpecial,
		CanSmartSearch: false,
		SupportsAdult:  true,
	}
}

func (n *SukebeiProvider) GetLatest() (ret []*hibiketorrent.AnimeTorrent, err error) {
	fp := gofeed.NewParser()

	url, err := buildURL(BuildURLOptions{
		Provider: "sukebei",
		Query:    "",
		Category: "art-anime",
		SortBy:   "seeders",
		Filter:   "",
	})
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

	wg := sync.WaitGroup{}
	mu := sync.Mutex{}

	for _, torrent := range res {
		wg.Add(1)
		go func(torrent Torrent) {
			defer wg.Done()
			mu.Lock()
			ret = append(ret, torrent.toAnimeTorrent(SukebeiProviderName))
			mu.Unlock()
		}(torrent)
	}

	wg.Wait()

	return
}

func (n *SukebeiProvider) Search(opts hibiketorrent.AnimeSearchOptions) (ret []*hibiketorrent.AnimeTorrent, err error) {
	fp := gofeed.NewParser()

	n.logger.Trace().Str("query", opts.Query).Msg("nyaa: Search query")

	url, err := buildURL(BuildURLOptions{
		Provider: "sukebei",
		Query:    opts.Query,
		Category: "art-anime",
		SortBy:   "seeders",
		Filter:   "",
	})
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

	wg := sync.WaitGroup{}
	mu := sync.Mutex{}

	for _, torrent := range res {
		wg.Add(1)
		go func(torrent Torrent) {
			defer wg.Done()
			mu.Lock()
			ret = append(ret, torrent.toAnimeTorrent(SukebeiProviderName))
			mu.Unlock()
		}(torrent)
	}

	wg.Wait()

	return
}

func (n *SukebeiProvider) SmartSearch(opts hibiketorrent.AnimeSmartSearchOptions) (ret []*hibiketorrent.AnimeTorrent, err error) {
	return
}

func (n *SukebeiProvider) GetTorrentInfoHash(torrent *hibiketorrent.AnimeTorrent) (string, error) {
	return TorrentHash(torrent.Link)
}

func (n *SukebeiProvider) GetTorrentMagnetLink(torrent *hibiketorrent.AnimeTorrent) (string, error) {
	return TorrentMagnet(torrent.Link)
}
