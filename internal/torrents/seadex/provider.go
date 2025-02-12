package seadex

import (
	"context"
	"github.com/5rahim/habari"
	"github.com/rs/zerolog"
	"net/http"
	"seanime/internal/torrents/nyaa"
	"sync"
	"time"

	hibiketorrent "seanime/internal/extension/hibike/torrent"
)

const (
	ProviderName = "seadex"
)

type Provider struct {
	logger *zerolog.Logger
	seadex *SeaDex
}

func NewProvider(logger *zerolog.Logger) hibiketorrent.AnimeProvider {
	return &Provider{
		logger: logger,
		seadex: New(logger),
	}
}

func (n *Provider) GetSettings() hibiketorrent.AnimeProviderSettings {
	return hibiketorrent.AnimeProviderSettings{
		Type:           hibiketorrent.AnimeProviderTypeSpecial,
		CanSmartSearch: true, // Setting to true to allow previews
		SupportsAdult:  false,
	}
}

func (n *Provider) GetType() hibiketorrent.AnimeProviderType {
	return hibiketorrent.AnimeProviderTypeSpecial
}

func (n *Provider) GetLatest() (ret []*hibiketorrent.AnimeTorrent, err error) {
	return
}

func (n *Provider) Search(opts hibiketorrent.AnimeSearchOptions) (ret []*hibiketorrent.AnimeTorrent, err error) {
	return n.findTorrents(&opts.Media)
}

func (n *Provider) SmartSearch(opts hibiketorrent.AnimeSmartSearchOptions) (ret []*hibiketorrent.AnimeTorrent, err error) {
	return n.findTorrents(&opts.Media)
}

func (n *Provider) findTorrents(media *hibiketorrent.Media) (ret []*hibiketorrent.AnimeTorrent, err error) {
	seadexTorrents, err := n.seadex.FetchTorrents(media.ID, media.RomajiTitle)
	if err != nil {
		return nil, err
	}

	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	wg.Add(len(seadexTorrents))

	for _, t := range seadexTorrents {
		go func(t *Torrent) {
			defer wg.Done()
			mu.Lock()
			ret = append(ret, t.toAnimeTorrent(ProviderName))
			mu.Unlock()
		}(t)
	}

	wg.Wait()

	return
}

//--------------------------------------------------------------------------------------------------------------------------------------------------//

func (n *Provider) GetTorrentInfoHash(torrent *hibiketorrent.AnimeTorrent) (string, error) {
	return torrent.MagnetLink, nil
}

func (n *Provider) GetTorrentMagnetLink(torrent *hibiketorrent.AnimeTorrent) (string, error) {
	return nyaa.TorrentMagnet(torrent.Link)
}

func (t *Torrent) toAnimeTorrent(providerName string) *hibiketorrent.AnimeTorrent {
	metadata := habari.Parse(t.Name)

	ret := &hibiketorrent.AnimeTorrent{
		Name:          t.Name,
		Date:          t.Date,
		Size:          0,  // Should be scraped
		FormattedSize: "", // Should be scraped
		Seeders:       0,  // Should be scraped
		Leechers:      0,  // Should be scraped
		DownloadCount: 0,  // Should be scraped
		Link:          t.Link,
		DownloadUrl:   "", // Should be scraped
		InfoHash:      t.InfoHash,
		MagnetLink:    "",   // Should be scraped
		Resolution:    "",   // Should be parsed
		IsBatch:       true, // Should be parsed
		EpisodeNumber: -1,   // Should be parsed
		ReleaseGroup:  "",   // Should be parsed
		Provider:      providerName,
		IsBestRelease: true,
		Confirmed:     true,
	}

	var seeders, leechers, downloads int
	var title, downloadUrl, formattedSize string

	// Try scraping from Nyaa
	// Since nyaa tends to be blocked, try for a few seconds only
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if t.Link != "" {
		downloadUrl = t.Link

		client := http.DefaultClient
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, ret.Link, nil)
		if err == nil {
			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()

				title, seeders, leechers, downloads, formattedSize, _, _, err = nyaa.TorrentInfo(ret.Link)
				if err == nil && title != "" {
					ret.Name = title // Override title
					ret.Seeders = seeders
					ret.Leechers = leechers
					ret.DownloadCount = downloads
					ret.DownloadUrl = downloadUrl
					ret.Size = 1
					ret.FormattedSize = formattedSize
				}
			}
		}
	}

	ret.Resolution = metadata.VideoResolution
	ret.ReleaseGroup = metadata.ReleaseGroup

	return ret
}
