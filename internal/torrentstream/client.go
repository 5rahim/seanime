package torrentstream

import (
	"errors"
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/storage"
	"github.com/samber/mo"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
)

type (
	Client struct {
		repository *Repository

		torrentClient mo.Option[*torrent.Client]
		srv           *http.Server
		torrents      []*torrent.Torrent
	}

	NewClientOptions struct {
		Repository *Repository
	}
)

func NewClient(opts *NewClientOptions) *Client {
	return &Client{
		repository: opts.Repository,
	}
}

func (c *Client) createTorrentClient() error {

	if err := c.repository.FailIfNoSettings(); err != nil {
		return err
	}

	settings := c.repository.settings.MustGet()

	cfg := torrent.NewDefaultClientConfig()
	cfg.NoUpload = false
	cfg.DisableIPv6 = settings.DisableIPV6
	cfg.ListenPort = settings.TorrentClientPort
	//cfg.SetListenAddr(settings.ListenAddr)
	cfg.DefaultStorage = storage.NewFileByInfoHash(settings.DownloadDir)

	client, err := torrent.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("error creating a new torrent client: %v", err)
	}

	c.repository.logger.Info().Msg("torrentstream: Starting torrent client")

	c.torrentClient = mo.Some(client)
	return nil
}

func (c *Client) DownloadTorrent(torrent string) error {
	t, err := c.AddTorrent(torrent)
	if err != nil {
		return err
	}
	t.DownloadAll()
	return nil
}

func (c *Client) ShowTorrents() ([]*torrent.Torrent, bool) {
	if c.torrentClient.IsAbsent() {
		return nil, false
	}

	return c.torrentClient.MustGet().Torrents(), true
}

func (c *Client) AddTorrent(id string) (*torrent.Torrent, error) {
	if c.torrentClient.IsAbsent() {
		return nil, errors.New("torrent client is not initialized")
	}

	if strings.HasPrefix(id, "magnet") {
		return c.AddMagnet(id)
	} else if strings.HasPrefix(id, "http") {
		return c.AddTorrentURL(id)
	} else {
		return c.AddTorrentFile(id)
	}
}

func (c *Client) AddMagnet(magnet string) (*torrent.Torrent, error) {
	if c.torrentClient.IsAbsent() {
		return nil, errors.New("torrent client is not initialized")
	}

	t, err := c.torrentClient.MustGet().AddMagnet(magnet)
	if err != nil {
		return nil, err
	}
	<-t.GotInfo()
	return t, nil
}

func (c *Client) AddTorrentFile(file string) (*torrent.Torrent, error) {
	if c.torrentClient.IsAbsent() {
		return nil, errors.New("torrent client is not initialized")
	}

	t, err := c.torrentClient.MustGet().AddTorrentFromFile(file)
	if err != nil {
		return nil, err
	}
	<-t.GotInfo()
	return t, nil
}

func (c *Client) AddTorrentURL(url string) (*torrent.Torrent, error) {
	if c.torrentClient.IsAbsent() {
		return nil, errors.New("torrent client is not initialized")
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	fname := path.Base(url)
	tmp := os.TempDir()
	path.Join(tmp, fname)

	file, err := os.Create(fname)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return nil, err
	}

	t, err := c.torrentClient.MustGet().AddTorrentFromFile(file.Name())
	if err != nil {
		return nil, err
	}
	<-t.GotInfo()
	return t, nil
}

// Close the client and all torrents
func (c *Client) Close() (errs []error) {
	if c.torrentClient.IsAbsent() {
		return
	}
	return c.torrentClient.MustGet().Close()
}

func (c *Client) FindByInfoHash(infoHash string) (*torrent.Torrent, error) {
	if c.torrentClient.IsAbsent() {
		return nil, errors.New("torrent client is not initialized")
	}

	torrents := c.torrentClient.MustGet().Torrents()
	for _, t := range torrents {
		if t.InfoHash().AsString() == infoHash {
			return t, nil
		}
	}
	return nil, fmt.Errorf("no torrents match info hash: %v", infoHash)
}
