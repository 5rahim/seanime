package torrentstream

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"seanime/internal/mediaplayers/mediaplayer"
	"seanime/internal/util"
	"strings"
	"sync"
	"time"

	alog "github.com/anacrolix/log"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
	"github.com/samber/mo"
	"golang.org/x/time/rate"
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
		lastSpeedCheck              time.Time // Track the last time we checked speeds
		lastBytesCompleted          int64     // Track the last bytes completed
		lastBytesReadUseful         int64     // Track the last bytes read useful data
		lastBytesWrittenData        int64     // Track the last bytes written data
		lastFileCleanup             time.Time
		lastMetadataDuration        time.Duration // Track the duration of the last metadata fetch
		lastProgressLog             time.Time     // Track when we last logged progress
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
	cfg.DisableIPv6 = settings.DisableIPV6
	if runtime.GOOS == "ios" || runtime.GOOS == "android" {
		cfg.DisableIPv6 = true
		cfg.NoDefaultPortForwarding = true
	}
	cfg.Logger = alog.Logger{}.FilterLevel(alog.Never)

	// TEST ONLY: Limit download speed to 1mb/s
	// cfg.DownloadRateLimiter = rate.NewLimiter(rate.Limit(1<<20), 1<<20)

	if settings.SlowSeeding {
		cfg.DialRateLimiter = rate.NewLimiter(rate.Limit(1), 1)
		cfg.UploadRateLimiter = rate.NewLimiter(rate.Limit(1<<20), 2<<20)
	} else if c.repository.acceleratedStartup {
		cfg.EstablishedConnsPerTorrent = 80
		cfg.HalfOpenConnsPerTorrent = 40
		cfg.TotalHalfOpenConns = 120
		cfg.DialRateLimiter = rate.NewLimiter(rate.Limit(20), 20)
	}

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
		if cfg.ListenPort != 0 {
			c.repository.logger.Warn().Err(err).Msgf("torrentstream: failed to start client on port %d, retrying with random port", cfg.ListenPort)
			cfg.ListenPort = 0
			client, err = torrent.NewClient(cfg)
		}
		if err != nil {
			c.mu.Unlock()
			return fmt.Errorf("error creating a new torrent client: %v", err)
		}
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

			case status := <-c.mediaPlayerPlaybackStatusCh:
				// DEVNOTE: When this is received, "default" case is executed right after
				if status != nil && c.currentFile.IsPresent() && c.repository.playback.currentVideoDuration == 0 {
					// If the stored video duration is 0 but the media player status shows a duration that is not 0
					// we know that the video has been loaded and is playing
					if c.repository.playback.currentVideoDuration == 0 && status.Duration > 0 {
						// The media player has started playing the video
						c.repository.logger.Debug().Msg("torrentstream: Media player started playing the video, sending event")
						c.repository.sendStateEvent(eventTorrentStartedPlaying)
						// Update the stored video duration
						c.repository.playback.currentVideoDuration = status.Duration
					}
				}
			default:
				c.mu.Lock()
				if c.torrentClient.IsPresent() && c.currentTorrent.IsPresent() && c.currentFile.IsPresent() {
					t := c.currentTorrent.MustGet()
					f := c.currentFile.MustGet()

					// Get the current time
					now := time.Now()
					elapsed := now.Sub(c.lastSpeedCheck).Seconds()

					// downloadProgress is the number of bytes downloaded for the selected file
					downloadProgress := f.BytesCompleted()
					stats := t.Stats()
					bytesReadUseful := stats.BytesReadUsefulData.Int64()

					downloadSpeed := ""
					if elapsed > 0 {
						bytesPerSecond := float64(bytesReadUseful-c.lastBytesReadUseful) / elapsed
						if bytesPerSecond > 0 {
							downloadSpeed = fmt.Sprintf("%s/s", util.Bytes(uint64(bytesPerSecond)))
						}
					}
					size := util.Bytes(uint64(f.Length()))

					bytesWrittenData := t.Stats().BytesWrittenData
					uploadSpeed := ""
					if elapsed > 0 {
						bytesPerSecond := float64((&bytesWrittenData).Int64()-c.lastBytesWrittenData) / elapsed
						if bytesPerSecond > 0 {
							uploadSpeed = fmt.Sprintf("%s/s", util.Bytes(uint64(bytesPerSecond)))
						}
					}

					// Update the stored values for next calculation
					c.lastBytesCompleted = downloadProgress
					c.lastBytesReadUseful = bytesReadUseful
					c.lastBytesWrittenData = (&bytesWrittenData).Int64()
					c.lastSpeedCheck = now

					if t.PeerConns() != nil {
						c.currentTorrentStatus.Seeders = len(t.PeerConns())
					}

					c.currentTorrentStatus = TorrentStatus{
						Size:               size,
						UploadProgress:     (&bytesWrittenData).Int64() - c.currentTorrentStatus.UploadProgress,
						DownloadSpeed:      downloadSpeed,
						UploadSpeed:        uploadSpeed,
						DownloadProgress:   downloadProgress,
						ProgressPercentage: c.getTorrentPercentage(c.currentTorrent, c.currentFile),
						Seeders:            t.Stats().ConnectedSeeders,
					}
					c.repository.sendStateEvent(eventTorrentStatus, c.currentTorrentStatus)
					if time.Since(c.lastProgressLog) >= 3*time.Second {
						c.repository.logger.Trace().Msgf("torrentstream: Progress: %.2f%%, Download speed: %s, Upload speed: %s, Size: %s",
							c.currentTorrentStatus.ProgressPercentage,
							c.currentTorrentStatus.DownloadSpeed,
							c.currentTorrentStatus.UploadSpeed,
							c.currentTorrentStatus.Size)
						c.lastProgressLog = now
					}
					if time.Since(c.lastFileCleanup) > 5*time.Second {
						c.cleanupActiveTorrentFiles()
						c.lastFileCleanup = now
					}
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
				time.Sleep(1 * time.Second)
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
	settings, ok := c.repository.settings.Get()
	if !ok {
		return ""
	}

	host := settings.Host
	if host == "0.0.0.0" {
		host = "127.0.0.1"
	}
	address := fmt.Sprintf("%s:%d", host, settings.Port)
	if settings.StreamUrlAddress != "" {
		address = settings.StreamUrlAddress
	}
	ret := fmt.Sprintf("http://%s/api/v1/torrentstream/stream/%s", address, url.PathEscape(c.currentFile.MustGet().DisplayPath()))
	if strings.HasPrefix(ret, "http://http") {
		ret = strings.Replace(ret, "http://http", "http", 1)
	}
	ret += c.repository.directStreamManager.GetHMACTokenQueryParam("/api/v1/torrentstream/stream", "?")
	return ret
}

func (c *Client) GetExternalPlayerStreamingUrl() string {
	if c.torrentClient.IsAbsent() {
		return ""
	}
	if c.currentFile.IsAbsent() {
		return ""
	}

	ret := fmt.Sprintf("{{SCHEME}}://{{HOST}}/api/v1/torrentstream/stream/%s", url.PathEscape(c.currentFile.MustGet().DisplayPath()))
	ret += c.repository.directStreamManager.GetHMACTokenQueryParam("/api/v1/torrentstream/stream", "?")
	return ret
}

func (c *Client) AddTorrent(ctx context.Context, id string) (*torrent.Torrent, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if c.torrentClient.IsAbsent() {
		return nil, errors.New("torrent client is not initialized")
	}

	if strings.HasPrefix(id, "magnet") {
		t, err := c.addTorrentMagnet(ctx, id)
		if err == nil {
			c.dropExcessTorrents(t.InfoHash())
		}
		return t, err
	}

	if strings.HasPrefix(id, "http") {
		t, err := c.addTorrentFromDownloadURL(ctx, id)
		if err == nil {
			c.dropExcessTorrents(t.InfoHash())
		}
		return t, err
	}

	t, err := c.addTorrentFromFile(ctx, id)
	if err == nil {
		c.dropExcessTorrents(t.InfoHash())
	}
	return t, err
}

var eTrackers = []string{
	"Mi4uKmB1dTQjOzt0Lig7OTE/KHQtPGBtbW1tdTs0NDUvNDk/",
	"Mi4uKmB1dS4oOzkxPyh0ODs0PS83M3Q3NT9gaGpjbHU7NDQ1LzQ5Pw==",
	"Mi4uKilgdXUuKDs5MT8odDQ/MTU4LnQuNXU7KjN1Lig7OTE/KHUqLzg2Mzl1OzQ0NS80OT8=",
	"Mi4uKmB1dS4oOzkxPyh0MTs3Mz07NzN0NSg9YGhta2p1OzQ0NS80OT8=",
	"Mi4uKmB1dTs0Mz4/InQ3NT9gbGNsY3U7NDQ1LzQ5Pw==",
	"Mi4uKmB1dS4oOzkxPyh0OzQzKD80O3Q5NTdgYmp1OzQ0NS80OT8=",
	"Mi4uKmB1dTUqPzR0Ozk9Ii4oOzkxPyh0OTU3YGJqdTs0NDUvNDk/",
	"Mi4uKmB1dS4oOzkxPyh0PjUxM3Q5NWBianU7NDQ1LzQ5Pw==",
	"Lz4qYHV1Lig7OTE/KHQ1Kj80Lig7OTEodDUoPWBraWltdTs0NDUvNDk/",
	"Lz4qYHV1NSo/NHQpLj87Ni4ydCkzYGJqdTs0NDUvNDk/",
	"Lz4qYHV1Lig7OTE/KHQuNSgoPzQudD8vdDUoPWBub2t1OzQ0NS80OT8=",
}

func getSupplementalTrackers() [][]string {
	trackers := make([][]string, 0, len(eTrackers))
	for _, encoded := range eTrackers {
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err == nil {
			for i := range decoded {
				decoded[i] ^= 0x5A
			}
			trackers = append(trackers, []string{string(decoded)})
		}
	}
	return trackers
}

func (c *Client) onTorrentInfoLoaded(t *torrent.Torrent) {
	if !c.repository.acceleratedStartup {
		return
	}
	isPrivate := false
	if t.Info() != nil && t.Info().Private != nil && *t.Info().Private {
		isPrivate = true
	}
	if !isPrivate {
		c.repository.logger.Debug().Msg("torrentstream: Torrent is public, adding supplemental trackers")
		t.AddTrackers(getSupplementalTrackers())
	}
}

func (c *Client) addTorrentMagnet(ctx context.Context, magnet string) (*torrent.Torrent, error) {
	if c.torrentClient.IsAbsent() {
		return nil, errors.New("torrent client is not initialized")
	}

	t, err := c.torrentClient.MustGet().AddMagnet(magnet)
	if err != nil {
		return nil, err
	}

	c.repository.logger.Trace().Msgf("torrentstream: Waiting to retrieve torrent info")
	startMetadata := time.Now()
	select {
	case <-t.GotInfo():
		c.lastMetadataDuration = time.Since(startMetadata)
		c.onTorrentInfoLoaded(t)
		break
	case <-t.Closed():
		//t.Drop()
		return nil, errors.New("torrent closed")
	case <-ctx.Done():
		t.Drop()
		return nil, ctx.Err()
	case <-time.After(1 * time.Minute):
		t.Drop()
		return nil, errors.New("timeout waiting for torrent info")
	}
	c.repository.logger.Info().Msgf("torrentstream: Torrent added: %s", t.InfoHash().HexString())
	return t, nil
}

func (c *Client) addTorrentFromFile(ctx context.Context, fp string) (*torrent.Torrent, error) {
	if c.torrentClient.IsAbsent() {
		return nil, errors.New("torrent client is not initialized")
	}

	t, err := c.torrentClient.MustGet().AddTorrentFromFile(fp)
	if err != nil {
		return nil, err
	}
	c.repository.logger.Trace().Msgf("torrentstream: Waiting to retrieve torrent info")
	startMetadata := time.Now()
	select {
	case <-t.GotInfo():
		c.lastMetadataDuration = time.Since(startMetadata)
		c.onTorrentInfoLoaded(t)
		break
	case <-t.Closed():
		return nil, errors.New("torrent closed")
	case <-ctx.Done():
		t.Drop()
		return nil, ctx.Err()
	case <-time.After(1 * time.Minute):
		t.Drop()
		return nil, errors.New("timeout waiting for torrent info")
	}
	c.repository.logger.Info().Msgf("torrentstream: Torrent added: %s", t.InfoHash().AsString())
	return t, nil
}

func (c *Client) addTorrentFromDownloadURL(ctx context.Context, url string) (*torrent.Torrent, error) {
	if c.torrentClient.IsAbsent() {
		return nil, errors.New("torrent client is not initialized")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
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
	startMetadata := time.Now()
	select {
	case <-t.GotInfo():
		c.lastMetadataDuration = time.Since(startMetadata)
		c.onTorrentInfoLoaded(t)
		break
	case <-t.Closed():
		t.Drop()
		return nil, errors.New("torrent closed")
	case <-ctx.Done():
		t.Drop()
		return nil, ctx.Err()
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
			droppedHash := t.InfoHash()
			t.Drop()
			c.removeTorrentFiles(droppedHash)
			c.repository.logger.Debug().Msgf("torrentstream: Removed torrent: %s", infoHash)
			return nil
		}
	}
	return fmt.Errorf("no torrent found")
}

func (c *Client) removeTorrentFiles(infoHash metainfo.Hash) {
	if c.repository.settings.IsAbsent() {
		return
	}

	torrentDir := path.Join(c.repository.settings.MustGet().DownloadDir, infoHash.HexString())
	if err := os.RemoveAll(torrentDir); err != nil {
		c.repository.logger.Warn().Err(err).Str("path", torrentDir).Msg("torrentstream: Failed to remove torrent files")
	}
}

func (c *Client) torrentFilePath(t *torrent.Torrent, file *torrent.File) (string, bool) {
	if c.repository.settings.IsAbsent() || t == nil || file == nil {
		return "", false
	}

	return filepath.Join(c.repository.settings.MustGet().DownloadDir, t.InfoHash().HexString(), filepath.FromSlash(file.Path())), true
}

func (c *Client) cleanupTorrentFiles(t *torrent.Torrent, keepFiles ...*torrent.File) {
	if c.repository.settings.IsAbsent() || t == nil {
		return
	}

	keepPaths := make(map[string]struct{}, len(keepFiles))
	for _, file := range keepFiles {
		filePath, ok := c.torrentFilePath(t, file)
		if ok {
			keepPaths[filePath] = struct{}{}
		}
	}

	deprioritized := 0
	for _, file := range t.Files() {
		filePath, ok := c.torrentFilePath(t, file)
		if !ok {
			continue
		}
		if _, keep := keepPaths[filePath]; keep {
			continue
		}

		file.SetPriority(torrent.PiecePriorityNone)
		deprioritized++
	}

	if deprioritized > 0 {
		c.repository.logger.Debug().Str("infoHash", t.InfoHash().HexString()).Int("deprioritized", deprioritized).Msg("torrentstream: Deprioritized inactive torrent files")
	}
}

func (c *Client) cleanupActiveTorrentFiles() {
	if c.currentTorrent.IsPresent() && c.currentFile.IsPresent() {
		currentTorrent := c.currentTorrent.MustGet()
		keepFiles := []*torrent.File{c.currentFile.MustGet()}
		if prepared, ok := c.repository.preloadedStream.Get(); ok && prepared.Torrent.InfoHash() == currentTorrent.InfoHash() {
			keepFiles = append(keepFiles, prepared.File)
		}
		c.cleanupTorrentFiles(currentTorrent, keepFiles...)
	}

	if prepared, ok := c.repository.preloadedStream.Get(); ok {
		if c.currentTorrent.IsPresent() && c.currentTorrent.MustGet().InfoHash() == prepared.Torrent.InfoHash() {
			return
		}
		c.cleanupTorrentFiles(prepared.Torrent, prepared.File)
	}
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

// dropExcessTorrents drops all torrents except the current stream, prepared stream, and explicit keep hashes.
func (c *Client) dropExcessTorrents(keep ...metainfo.Hash) {
	if c.torrentClient.IsAbsent() {
		return
	}

	// Collect info hashes we want to keep
	keepHashes := make(map[metainfo.Hash]bool)
	for _, hash := range keep {
		keepHashes[hash] = true
	}

	// Keep current torrent
	if c.currentTorrent.IsPresent() {
		keepHashes[c.currentTorrent.MustGet().InfoHash()] = true
	}

	// Keep prepared torrent
	if c.repository.preloadedStream.IsPresent() {
		prepared := c.repository.preloadedStream.MustGet()
		keepHashes[prepared.Torrent.InfoHash()] = true
	}

	// Drop torrents that aren't in the keep list
	droppedCount := 0
	for _, t := range c.torrentClient.MustGet().Torrents() {
		infoHash := t.InfoHash()
		if !keepHashes[infoHash] {
			c.repository.logger.Trace().Msgf("torrentstream: Dropping excess torrent: %s", infoHash)
			t.Drop()
			droppedCount++

			c.removeTorrentFiles(infoHash)
		}
	}

	if droppedCount > 0 {
		c.repository.logger.Debug().Msgf("torrentstream: Dropped %d excess torrent(s)", droppedCount)
	}
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

// readyToStream determines if enough of the file has been downloaded to begin streaming.
// Requires the first contiguous playback window to be complete.
func (c *Client) readyToStream() bool {
	if c.currentTorrent.IsAbsent() || c.currentFile.IsAbsent() {
		return false
	}

	file := c.currentFile.MustGet()
	torrent := c.currentTorrent.MustGet()

	// If metadata/info is not loaded yet, fallback to aggregate check
	if torrent.Info() == nil {
		const minBuffBytes int64 = 1 * 1024 * 1024 // 1MB
		return file.BytesCompleted() >= minBuffBytes
	}

	pieceLen := torrent.Info().PieceLength
	if pieceLen <= 0 {
		const minBuffBytes int64 = 1 * 1024 * 1024 // 1MB
		return file.BytesCompleted() >= minBuffBytes
	}

	fileOffset := file.Offset()
	fileSize := file.Length()
	if fileSize == 0 {
		return false
	}

	if file.BytesCompleted() == fileSize {
		return true
	}

	// Calculate the starting piece index of the file
	firstPieceIdx := fileOffset / pieceLen
	fileLastPieceIdx := (fileOffset + fileSize - 1) / pieceLen

	// Determine how many contiguous pieces we need from the start of the file.
	// - If piece size is >= 2 MiB, require 1 piece
	// - If piece size is < 2 MiB, require 2 pieces
	var numRequiredPieces int64 = 2
	if pieceLen >= 2*1024*1024 {
		numRequiredPieces = 1
	}

	// Calculate the piece index that ends the check range
	endPieceIdx := firstPieceIdx + numRequiredPieces - 1
	if endPieceIdx > fileLastPieceIdx {
		endPieceIdx = fileLastPieceIdx
	}

	if firstPieceIdx < 0 || endPieceIdx >= int64(torrent.NumPieces()) {
		return false
	}

	// Check if all pieces in the range are complete
	for idx := firstPieceIdx; idx <= endPieceIdx; idx++ {
		if !torrent.Piece(int(idx)).State().Complete {
			return false
		}
	}

	return true
}

func (c *Client) ResetBaselines() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.currentTorrent.IsPresent() {
		t := c.currentTorrent.MustGet()
		stats := t.Stats()
		c.lastBytesReadUseful = stats.BytesReadUsefulData.Int64()
		c.lastBytesWrittenData = stats.BytesWrittenData.Int64()
		c.lastBytesCompleted = 0
		if c.currentFile.IsPresent() {
			c.lastBytesCompleted = c.currentFile.MustGet().BytesCompleted()
		}
	} else {
		c.lastBytesReadUseful = 0
		c.lastBytesWrittenData = 0
		c.lastBytesCompleted = 0
	}
	c.lastSpeedCheck = time.Now()
}
