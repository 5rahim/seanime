package torrentstream

import (
	"context"
	"errors"
	"fmt"
	alog "github.com/anacrolix/log"
	"github.com/anacrolix/sync"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/storage"
	"github.com/dustin/go-humanize"
	"github.com/samber/mo"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"seanime/internal/mediaplayers/mediaplayer"
	"strings"
	"time"
)

type (
	Client struct {
		repository *Repository

		torrentClient        mo.Option[*torrent.Client]
		currentTorrent       mo.Option[*torrent.Torrent]
		currentFile          mo.Option[*torrent.File]
		currentTorrentStatus TorrentStatus
		cancelFunc           context.CancelFunc

		mu                          sync.Mutex
		stopCh                      chan struct{}                    // Closed when the media player stops
		mediaPlayerPlaybackStatusCh chan *mediaplayer.PlaybackStatus // Continuously receives playback status
		timeSinceLoggedSeeding      time.Time
	}

	TorrentStatus struct {
		UploadProgress     int64   `json:"uploadProgress"`
		DownloadProgress   int64   `json:"downloadProgress"`
		ProgressPercentage float64 `json:"progressPercentage"`
		DownloadSpeed      string  `json:"downloadSpeed"`
		UploadSpeed        string  `json:"uploadSpeed"`
		Size               string  `json:"size"`
		Seeders            int     `json:"seeders"`
	}

	NewClientOptions struct {
		Repository *Repository
	}
)

func NewClient(repository *Repository) *Client {
	ret := &Client{
		repository:                  repository,
		torrentClient:               mo.None[*torrent.Client](),
		currentFile:                 mo.None[*torrent.File](),
		currentTorrent:              mo.None[*torrent.Torrent](),
		stopCh:                      make(chan struct{}),
		mediaPlayerPlaybackStatusCh: make(chan *mediaplayer.PlaybackStatus, 1),
	}

	return ret
}

// initializeClient will create and torrent client.
// The client is designed to support only one torrent at a time, and seed it.
// Upon initialization, the client will drop all torrents.
func (c *Client) initializeClient() error {
	// Fail if no settings
	if err := c.repository.FailIfNoSettings(); err != nil {
		return err
	}

	// Cancel the previous context, terminating the goroutine if it's running
	if c.cancelFunc != nil {
		c.cancelFunc()
	}

	// Context for the client's goroutine
	var ctx context.Context
	ctx, c.cancelFunc = context.WithCancel(context.Background())

	// Get the settings
	settings := c.repository.settings.MustGet()

	// Define torrent client settings
	cfg := torrent.NewDefaultClientConfig()
	cfg.Seed = true
	cfg.DisableIPv6 = true
	cfg.Logger = alog.Logger{}
	//cfg.DisableAggressiveUpload = true
	//cfg.Debug = true

	if settings.TorrentClientHost != "" {
		cfg.ListenHost = func(network string) string { return settings.TorrentClientHost }
	}

	if settings.TorrentClientPort == 0 {
		settings.TorrentClientPort = 43213
	}
	cfg.ListenPort = settings.TorrentClientPort
	// Set the download directory
	// e.g. /path/to/temp/seanime/torrentstream/{infohash}
	cfg.DefaultStorage = storage.NewFileByInfoHash(settings.DownloadDir)

	c.mu.Lock()
	// Create the torrent client
	client, err := torrent.NewClient(cfg)
	if err != nil {
		c.mu.Unlock()
		return fmt.Errorf("error creating a new torrent client: %v", err)
	}
	c.repository.logger.Info().Msgf("torrentstream: Initialized torrent client on port %d", settings.TorrentClientPort)
	c.torrentClient = mo.Some(client)
	c.dropTorrents()
	c.mu.Unlock()

	go func(ctx context.Context) {

		for {
			select {
			case <-ctx.Done():
				c.repository.logger.Debug().Msg("torrentstream: Context cancelled, stopping torrent client")
				return
			//case <-c.stopCh:
			//	c.mu.Lock()
			//	c.stopCh = make(chan struct{})
			//	c.repository.logger.Debug().Msg("torrentstream: Handling media player stopped event")
			//	// This is to prevent the client from downloading the whole torrent when the user stops watching
			//	// Also, the torrent might be a batch - so we don't want to download the whole thing
			//	if c.currentTorrent.IsPresent() {
			//		if c.currentTorrentStatus.ProgressPercentage < 70 {
			//			c.repository.logger.Debug().Msg("torrentstream: Dropping torrent, completion is less than 70%")
			//			c.dropTorrents()
			//		}
			//		c.repository.logger.Debug().Msg("torrentstream: Resetting current torrent and status")
			//	}
			//	c.currentTorrent = mo.None[*torrent.Torrent]()                  // Reset the current torrent
			//	c.currentFile = mo.None[*torrent.File]()                        // Reset the current file
			//	c.currentTorrentStatus = TorrentStatus{}                        // Reset the torrent status
			//	c.repository.serverManager.stopServer()                         // Stop streaming server
			//	c.repository.wsEventManager.SendEvent(eventTorrentStopped, nil) // Send torrent stopped event
			//	c.repository.mediaPlayerRepository.Stop()                       // Stop the media player gracefully if it's running
			//	c.mu.Unlock()

			case status := <-c.mediaPlayerPlaybackStatusCh:
				// DEVNOTE: When this is received, "default" case is executed right after
				if status != nil && c.currentFile.IsPresent() && c.repository.playback.currentVideoDuration == 0 {
					// If the stored video duration is 0 but the media player status shows a duration that is not 0
					// we know that the video has been loaded and is playing
					if c.repository.playback.currentVideoDuration == 0 && status.Duration > 0 {
						// The media player has started playing the video
						c.repository.logger.Debug().Msg("torrentstream: Media player started playing the video, sending event")
						c.repository.wsEventManager.SendEvent(eventTorrentStartedPlaying, nil)
						// Update the stored video duration
						c.repository.playback.currentVideoDuration = status.Duration
					}
				}
			default:
				c.mu.Lock()
				if c.torrentClient.IsPresent() && c.currentTorrent.IsPresent() && c.currentFile.IsPresent() {
					t := c.currentTorrent.MustGet()
					f := c.currentFile.MustGet()

					// downloadProgress is the number of bytes downloaded
					downloadProgress := t.BytesCompleted()
					// Difference between the current download progress and the last download progress
					progressDiff := downloadProgress - c.currentTorrentStatus.DownloadProgress
					// Get the download speed based on the difference
					downloadSpeed := ""
					if progressDiff > 0 {
						downloadSpeed = fmt.Sprintf("%s/s", humanize.Bytes(uint64(progressDiff)))
					}
					size := humanize.Bytes(uint64(f.Length()))

					bytesWrittenData := t.Stats().BytesWrittenData
					// uploadProgress is the number of bytes uploaded
					uploadProgress := (&bytesWrittenData).Int64() - c.currentTorrentStatus.UploadProgress
					uploadSpeed := ""
					if uploadProgress > 0 {
						uploadSpeed = fmt.Sprintf("%s/s", humanize.Bytes(uint64(uploadProgress)))
					}
					if t.PeerConns() != nil {
						c.currentTorrentStatus.Seeders = len(t.PeerConns())
					}

					//// If the torrent status went from 0% to > 0%, send an event, as the torrent has been loaded
					//if c.currentTorrentStatus.ProgressPercentage == 0. && c.getTorrentPercentage(c.currentTorrent) > 0. {
					//	c.repository.logger.Debug().Msg("torrentstream: Torrent loaded, sending event")
					//	c.repository.wsEventManager.SendEvent(eventTorrentLoaded, nil)
					//}

					c.currentTorrentStatus = TorrentStatus{
						Size:               size,
						UploadProgress:     uploadProgress,
						DownloadSpeed:      downloadSpeed,
						UploadSpeed:        uploadSpeed,
						DownloadProgress:   downloadProgress,
						ProgressPercentage: c.getTorrentPercentage(c.currentTorrent, c.currentFile),
						Seeders:            t.Stats().ConnectedSeeders,
					}
					c.repository.wsEventManager.SendEvent(eventTorrentStatus, c.currentTorrentStatus)
					// Always log the progress so the user knows what's happening
					c.repository.logger.Trace().Msgf("torrentstream: Progress: %.2f%%, Download speed: %s, Upload speed: %s, Size: %s",
						c.currentTorrentStatus.ProgressPercentage,
						c.currentTorrentStatus.DownloadSpeed,
						c.currentTorrentStatus.UploadSpeed,
						c.currentTorrentStatus.Size)
					c.timeSinceLoggedSeeding = time.Now()
				}
				c.mu.Unlock()
				if c.torrentClient.IsPresent() {
					if time.Since(c.timeSinceLoggedSeeding) > 20*time.Second {
						c.timeSinceLoggedSeeding = time.Now()
						for _, t := range c.torrentClient.MustGet().Torrents() {
							if t.Seeding() {
								c.repository.logger.Trace().Msgf("torrentstream: Seeding last torrent, %d peers", t.Stats().ActivePeers)
							}
						}
					}
				}
				time.Sleep(3 * time.Second)
			}
		}
	}(ctx)

	return nil
}

func (c *Client) GetStreamingUrl() string {
	if c.torrentClient.IsAbsent() {
		return ""
	}

	if c.currentFile.IsAbsent() {
		return ""
	}

	settings := c.repository.settings.MustGet()
	if settings.StreamingServerHost == "0.0.0.0" {
		return fmt.Sprintf("http://127.0.0.1:%d/stream/%s", settings.StreamingServerPort, url.PathEscape(c.currentFile.MustGet().DisplayPath()))
	}
	return fmt.Sprintf("http://%s:%d/stream/%s", settings.StreamingServerHost, settings.StreamingServerPort, url.PathEscape(c.currentFile.MustGet().DisplayPath()))
}

func (c *Client) AddTorrent(id string) (*torrent.Torrent, error) {
	if c.torrentClient.IsAbsent() {
		return nil, errors.New("torrent client is not initialized")
	}

	// Drop all torrents
	for _, t := range c.torrentClient.MustGet().Torrents() {
		t.Drop()
	}

	if strings.HasPrefix(id, "magnet") {
		return c.addTorrentMagnet(id)
	}

	if strings.HasPrefix(id, "http") {
		return c.addTorrentFromDownloadURL(id)
	}

	return c.addTorrentFromFile(id)
}

func (c *Client) addTorrentMagnet(magnet string) (*torrent.Torrent, error) {
	if c.torrentClient.IsAbsent() {
		return nil, errors.New("torrent client is not initialized")
	}

	t, err := c.torrentClient.MustGet().AddMagnet(magnet)
	if err != nil {
		return nil, err
	}

	c.repository.logger.Trace().Msgf("torrentstream: Waiting to retrieve torrent info")
	select {
	case <-t.GotInfo():
		break
	case <-t.Closed():
		t.Drop()
		return nil, errors.New("torrent closed")
	case <-time.After(1 * time.Minute):
		t.Drop()
		return nil, errors.New("timeout waiting for torrent info")
	}
	c.repository.logger.Info().Msgf("torrentstream: Torrent added: %s", t.InfoHash().AsString())
	return t, nil
}

func (c *Client) addTorrentFromFile(fp string) (*torrent.Torrent, error) {
	if c.torrentClient.IsAbsent() {
		return nil, errors.New("torrent client is not initialized")
	}

	t, err := c.torrentClient.MustGet().AddTorrentFromFile(fp)
	if err != nil {
		return nil, err
	}
	c.repository.logger.Trace().Msgf("torrentstream: Waiting to retrieve torrent info")
	<-t.GotInfo()
	c.repository.logger.Info().Msgf("torrentstream: Torrent added: %s", t.InfoHash().AsString())
	return t, nil
}

func (c *Client) addTorrentFromDownloadURL(url string) (*torrent.Torrent, error) {
	if c.torrentClient.IsAbsent() {
		return nil, errors.New("torrent client is not initialized")
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	filename := path.Base(url)
	file, err := os.Create(path.Join(os.TempDir(), filename))
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
	c.repository.logger.Trace().Msgf("torrentstream: Waiting to retrieve torrent info")
	select {
	case <-t.GotInfo():
		break
	case <-t.Closed():
		t.Drop()
		return nil, errors.New("torrent closed")
	case <-time.After(1 * time.Minute):
		t.Drop()
		return nil, errors.New("timeout waiting for torrent info")
	}
	c.repository.logger.Info().Msgf("torrentstream: Added torrent: %s", t.InfoHash().AsString())
	return t, nil
}

// Shutdown closes the torrent client and drops all torrents.
// This SHOULD NOT be called if you don't intend to reinitialize the client.
func (c *Client) Shutdown() (errs []error) {
	if c.torrentClient.IsAbsent() {
		return
	}
	c.dropTorrents()
	c.currentTorrent = mo.None[*torrent.Torrent]()
	c.currentTorrentStatus = TorrentStatus{}
	c.repository.logger.Debug().Msg("torrentstream: Closing torrent client")
	return c.torrentClient.MustGet().Close()
}

func (c *Client) FindTorrent(infoHash string) (*torrent.Torrent, error) {
	if c.torrentClient.IsAbsent() {
		return nil, errors.New("torrent client is not initialized")
	}

	torrents := c.torrentClient.MustGet().Torrents()
	for _, t := range torrents {
		if t.InfoHash().AsString() == infoHash {
			c.repository.logger.Debug().Msgf("torrentstream: Found torrent: %s", infoHash)
			return t, nil
		}
	}
	return nil, fmt.Errorf("no torrent found")
}

func (c *Client) RemoveTorrent(infoHash string) error {
	if c.torrentClient.IsAbsent() {
		return errors.New("torrent client is not initialized")
	}

	c.repository.logger.Trace().Msgf("torrentstream: Removing torrent: %s", infoHash)

	torrents := c.torrentClient.MustGet().Torrents()
	for _, t := range torrents {
		if t.InfoHash().AsString() == infoHash {
			t.Drop()
			c.repository.logger.Debug().Msgf("torrentstream: Removed torrent: %s", infoHash)
			return nil
		}
	}
	return fmt.Errorf("no torrent found")
}

func (c *Client) dropTorrents() {
	if c.torrentClient.IsAbsent() {
		return
	}
	c.repository.logger.Trace().Msg("torrentstream: Dropping all torrents")

	for _, t := range c.torrentClient.MustGet().Torrents() {
		t.Drop()
	}

	if c.repository.settings.IsPresent() {
		// Delete all torrents
		fe, err := os.ReadDir(c.repository.settings.MustGet().DownloadDir)
		if err == nil {
			for _, f := range fe {
				if f.IsDir() {
					_ = os.RemoveAll(path.Join(c.repository.settings.MustGet().DownloadDir, f.Name()))
				}
			}
		}
	}

	c.repository.logger.Debug().Msg("torrentstream: Dropped all torrents")
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// getTorrentPercentage returns the percentage of the current torrent file
// If no torrent is selected, it returns -1
func (c *Client) getTorrentPercentage(t mo.Option[*torrent.Torrent], f mo.Option[*torrent.File]) float64 {
	if t.IsAbsent() || f.IsAbsent() {
		return -1
	}

	if f.MustGet().Length() == 0 {
		return 0
	}

	return float64(f.MustGet().BytesCompleted()) / float64(f.MustGet().Length()) * 100
}

func (c *Client) readyToStream() bool {
	return c.getTorrentPercentage(c.currentTorrent, c.currentFile) > 5.
}
