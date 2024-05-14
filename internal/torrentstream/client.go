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
		torrents      []*torrent.Torrent
	}

	NewClientOptions struct {
		Repository *Repository
	}
)

func NewClient(repository *Repository) *Client {
	ret := &Client{
		repository:    repository,
		torrentClient: mo.None[*torrent.Client](),
	}

	return ret
}

// InitializeClient will create and torrent client and server
func (c *Client) InitializeClient() error {
	// Fail if no settings
	if err := c.repository.FailIfNoSettings(); err != nil {
		return err
	}

	// Get the settings
	settings := c.repository.settings.MustGet()

	// Define torrent client settings
	cfg := torrent.NewDefaultClientConfig()
	//cfg.SetListenAddr(settings.ListenAddr)
	cfg.NoUpload = false
	cfg.DisableIPv6 = settings.DisableIPV6
	if settings.TorrentClientPort == 0 {
		settings.TorrentClientPort = 43213
	}
	cfg.ListenPort = settings.TorrentClientPort
	// Set the download directory
	// e.g. /path/to/temp/seanime/torrentstream/{infohash}
	cfg.DefaultStorage = storage.NewFileByInfoHash(settings.DownloadDir)

	// Create the torrent client
	client, err := torrent.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("error creating a new torrent client: %v", err)
	}
	c.repository.logger.Info().Msg("torrentstream: Initialized torrent client")
	c.torrentClient = mo.Some(client)

	return nil
}

func (c *Client) GetStreamingUrl() string {
	if c.torrentClient.IsAbsent() {
		return ""
	}

	settings := c.repository.settings.MustGet()
	if settings.StreamingServerHost == "0.0.0.0" {
		return fmt.Sprintf("http://127.0.0.1:%d/stream", settings.StreamingServerPort)
	}
	return fmt.Sprintf("http://%s:%d/stream", settings.StreamingServerHost, settings.StreamingServerPort)
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

func (c *Client) AddTorrentFile(fp string) (*torrent.Torrent, error) {
	if c.torrentClient.IsAbsent() {
		return nil, errors.New("torrent client is not initialized")
	}

	t, err := c.torrentClient.MustGet().AddTorrentFromFile(fp)
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

	filename := path.Base(url)
	tmp := os.TempDir()
	path.Join(tmp, filename)

	file, err := os.Create(filename)
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

func (c *Client) Close() (errs []error) {
	if c.torrentClient.IsAbsent() {
		return
	}
	return c.torrentClient.MustGet().Close()
}

func (c *Client) FindTorrent(infoHash string) (*torrent.Torrent, error) {
	if c.torrentClient.IsAbsent() {
		return nil, errors.New("torrent client is not initialized")
	}

	torrents := c.torrentClient.MustGet().Torrents()
	for _, t := range torrents {
		if t.InfoHash().AsString() == infoHash {
			return t, nil
		}
	}
	return nil, fmt.Errorf("no torrent found")
}

func (c *Client) RemoveTorrent(infoHash string) error {
	if c.torrentClient.IsAbsent() {
		return errors.New("torrent client is not initialized")
	}

	torrents := c.torrentClient.MustGet().Torrents()
	for _, t := range torrents {
		if t.InfoHash().AsString() == infoHash {
			t.Drop()
			return nil
		}
	}
	return fmt.Errorf("no torrent found")
}
