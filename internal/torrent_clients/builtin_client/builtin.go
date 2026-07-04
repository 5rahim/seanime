package builtin_client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"seanime/internal/database/db"
	"seanime/internal/database/models"
	"seanime/internal/util"
	"sort"
	"strings"
	"sync"
	"time"

	g "github.com/anacrolix/generics"
	alog "github.com/anacrolix/log"
	anacrolix "github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/storage"
	"github.com/rs/zerolog"
	"golang.org/x/time/rate"
)

type noClosePieceCompletion struct {
	storage.PieceCompletion
}

func (noClosePieceCompletion) Close() error {
	return nil
}

const sequentialWindowPieces = 24

type NewClientOptions struct {
	Logger             *zerolog.Logger
	Database           *db.Database
	Dir                string
	Port               int
	MaxConnections     int
	DownloadLimitKB    int
	UploadLimitKB      int
	MaxActiveDownloads int
	DisableNetwork     bool
}

type Client struct {
	mu                 sync.RWMutex
	logger             *zerolog.Logger
	database           *db.Database
	client             *anacrolix.Client
	torrents           map[string]*torrentEntry
	downloadLimiter    *rate.Limiter
	uploadLimiter      *rate.Limiter
	maxActiveDownloads int
	closeCh            chan struct{}
	closed             bool
	pieceCompletion    storage.PieceCompletion
}

type torrentEntry struct {
	client                *Client
	model                 *models.LocalTorrent
	torrent               *anacrolix.Torrent
	lastSample            time.Time
	lastDownload          int64
	lastUpload            int64
	downSpeed             int64
	upSpeed               int64
	sequentialInitialized bool
	sequentialStart       int
	filePriorities        map[int]int
	storageCloser         io.Closer
	writeError            error
	writeErrorMu          sync.RWMutex
}

func (e *torrentEntry) getWriteError() error {
	e.writeErrorMu.RLock()
	defer e.writeErrorMu.RUnlock()
	return e.writeError
}

func (e *torrentEntry) setWriteError(err error) {
	e.writeErrorMu.Lock()
	defer e.writeErrorMu.Unlock()
	e.writeError = err
}

type TorrentSnapshot struct {
	Name        string    `json:"name"`
	Hash        string    `json:"hash"`
	Destination string    `json:"destination"`
	Paused      bool      `json:"paused"`
	Queued      bool      `json:"queued"`
	ForceStart  bool      `json:"forceStart"`
	Sequential  bool      `json:"sequential"`
	QueueIndex  int       `json:"queueIndex"`
	Length      int64     `json:"length"`
	Completed   int64     `json:"completed"`
	DownSpeed   int64     `json:"downSpeed"`
	UpSpeed     int64     `json:"upSpeed"`
	Seeds       int       `json:"seeds"`
	Peers       int       `json:"peers"`
	Downloaded  int64     `json:"downloaded"`
	Uploaded    int64     `json:"uploaded"`
	AddedAt     time.Time `json:"addedAt"`
	Error       string    `json:"error"`
}

type TorrentDetails struct {
	Torrent  TorrentSnapshot `json:"torrent"`
	Files    []FileDetails   `json:"files"`
	Trackers []string        `json:"trackers"`
	Peers    []PeerDetails   `json:"peers"`
}

type FileDetails struct {
	Index     int     `json:"index"`
	Path      string  `json:"path"`
	Length    int64   `json:"length"`
	Completed int64   `json:"completed"`
	Progress  float64 `json:"progress"`
	Priority  int     `json:"priority"`
}

type PeerDetails struct {
	Address string `json:"address"`
	Client  string `json:"client"`
}

func New(opts *NewClientOptions) (*Client, error) {
	if opts == nil || opts.Database == nil || opts.Logger == nil {
		return nil, errors.New("builtin torrent: missing dependencies")
	}
	if opts.Dir == "" {
		return nil, errors.New("builtin torrent: data directory is empty")
	}
	if opts.Port == 0 {
		opts.Port = 50007
	} else if opts.Port < 0 {
		opts.Port = 0
	}
	if opts.MaxConnections <= 0 {
		opts.MaxConnections = 50
	}
	if opts.MaxActiveDownloads <= 0 {
		opts.MaxActiveDownloads = 3
	}
	if err := os.MkdirAll(opts.Dir, 0700); err != nil {
		return nil, fmt.Errorf("create torrent data directory: %w", err)
	}

	pc, err := storage.NewDefaultPieceCompletionForDir(opts.Dir)
	if err != nil {
		opts.Logger.Warn().Err(err).Msg("builtin torrent: piece completion db is corrupt, deleting it")
		dbFile := filepath.Join(opts.Dir, ".torrent.db")
		_ = os.Remove(dbFile)
		_ = os.Remove(dbFile + "-journal")
		_ = os.Remove(dbFile + "-wal")
		_ = os.Remove(dbFile + "-shm")
		pc, err = storage.NewDefaultPieceCompletionForDir(opts.Dir)
		if err != nil {
			opts.Logger.Warn().Err(err).Msg("builtin torrent: failed to initialize persistent piece completion DB, falling back to in-memory map")
			pc = storage.NewMapPieceCompletion()
		}
	}

	downloadLimiter := newRateLimiter(opts.DownloadLimitKB)
	uploadLimiter := newRateLimiter(opts.UploadLimitKB)
	cfg := anacrolix.NewDefaultClientConfig()
	cfg.DataDir = opts.Dir
	cfg.ListenPort = opts.Port
	cfg.Seed = true
	cfg.DownloadRateLimiter = downloadLimiter
	cfg.UploadRateLimiter = uploadLimiter
	cfg.Logger = alog.Logger{}.FilterLevel(alog.Never)
	cfg.DefaultStorage = newTorrentStorage(opts.Dir, pc)
	if opts.DisableNetwork {
		cfg.NoDHT = true
		cfg.DisablePEX = true
		cfg.NoDefaultPortForwarding = true
		cfg.DialForPeerConns = false
	}
	if runtime.GOOS == "ios" || runtime.GOOS == "android" {
		cfg.DisableIPv6 = true
		cfg.NoDefaultPortForwarding = true
	}
	if opts.MaxConnections > 0 {
		cfg.EstablishedConnsPerTorrent = opts.MaxConnections
	}

	inner, err := anacrolix.NewClient(cfg)
	if err != nil {
		if cfg.ListenPort != 0 {
			opts.Logger.Warn().Err(err).Msgf("builtin torrent: failed to start client on port %d, retrying with random port", cfg.ListenPort)
			cfg.ListenPort = 0
			inner, err = anacrolix.NewClient(cfg)
		}
		if err != nil {
			pc.Close()
			return nil, fmt.Errorf("create torrent client: %w", err)
		}
	}

	c := &Client{
		logger:             opts.Logger,
		database:           opts.Database,
		client:             inner,
		torrents:           make(map[string]*torrentEntry),
		downloadLimiter:    downloadLimiter,
		uploadLimiter:      uploadLimiter,
		maxActiveDownloads: opts.MaxActiveDownloads,
		closeCh:            make(chan struct{}),
		pieceCompletion:    pc,
	}
	if err := c.restore(); err != nil {
		inner.Close()
		pc.Close()
		return nil, err
	}
	go c.runScheduler()
	return c, nil
}

func newRateLimiter(limitKB int) *rate.Limiter {
	if limitKB <= 0 {
		return rate.NewLimiter(rate.Inf, 1<<20)
	}
	bytesPerSecond := limitKB * 1024
	burst := max(bytesPerSecond, 1<<20)
	return rate.NewLimiter(rate.Limit(bytesPerSecond), burst)
}

func newTorrentStorage(baseDir string, pc storage.PieceCompletion) storage.ClientImplCloser {
	if util.IsMobile() {
		return newClassicFileStorage(baseDir, pc)
	}
	return storage.NewFileOpts(storage.NewFileClientOpts{
		ClientBaseDir:   baseDir,
		PieceCompletion: noClosePieceCompletion{pc},
		UsePartFiles:    g.Some(false),
	})
}

func (c *Client) Start() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.client != nil && !c.closed
}

func (c *Client) Close() {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	c.closed = true
	close(c.closeCh)
	inner := c.client
	var closers []io.Closer
	for _, entry := range c.torrents {
		if entry.storageCloser != nil {
			closers = append(closers, entry.storageCloser)
		}
	}
	c.mu.Unlock()

	if inner != nil {
		for _, err := range inner.Close() {
			c.logger.Warn().Err(err).Msg("builtin torrent: close error")
		}
	}

	for _, closer := range closers {
		_ = closer.Close()
	}

	if c.pieceCompletion != nil {
		_ = c.pieceCompletion.Close()
	}
}

func (c *Client) restore() error {
	defer util.HandlePanicInModuleThen("builtin_client/restore", func() {})
	items, err := c.database.GetLocalTorrents()
	if err != nil {
		return fmt.Errorf("load persisted torrents: %w", err)
	}
	for _, item := range items {
		if _, err := c.addPersisted(item); err != nil {
			c.logger.Warn().Err(err).Str("hash", item.Hash).Msg("builtin torrent: failed to restore torrent")
		}
	}
	c.reconcileQueue()
	return nil
}

func (c *Client) AddMagnet(magnet, destination string) (*anacrolix.Torrent, error) {
	defer util.HandlePanicInModuleThen("builtin_client/AddMagnet", func() {})
	if destination == "" {
		return nil, errors.New("destination is required")
	}
	destination = filepath.Clean(destination)
	if err := os.MkdirAll(destination, 0755); err != nil {
		return nil, fmt.Errorf("create destination: %w", err)
	}
	spec, err := anacrolix.TorrentSpecFromMagnetUri(magnet)
	if err != nil {
		return nil, fmt.Errorf("parse magnet: %w", err)
	}
	hash := strings.ToLower(spec.InfoHash.HexString())
	if hash == strings.Repeat("0", 40) {
		return nil, errors.New("magnet does not contain a v1 info hash")
	}

	c.mu.RLock()
	existing := c.torrents[hash]
	c.mu.RUnlock()
	if existing != nil {
		return existing.torrent, nil
	}

	items, err := c.database.GetLocalTorrents()
	if err != nil {
		return nil, err
	}
	queueIndex := len(items)
	item := &models.LocalTorrent{
		Hash:        hash,
		Magnet:      magnet,
		Name:        spec.DisplayName,
		Destination: destination,
		QueueIndex:  queueIndex,
	}
	t, err := c.addPersisted(item)
	if err != nil {
		return nil, err
	}
	if err := c.database.UpsertLocalTorrent(item); err != nil {
		c.removeRuntime(hash)
		return nil, fmt.Errorf("persist torrent: %w", err)
	}
	c.reconcileQueue()
	return t, nil
}

func (c *Client) addPersisted(item *models.LocalTorrent) (*anacrolix.Torrent, error) {
	defer util.HandlePanicInModuleThen("builtin_client/addPersisted", func() {})
	if item.Paused {
		entry := &torrentEntry{
			client: c, model: item, torrent: nil, lastSample: time.Now(), sequentialStart: -1,
			filePriorities: make(map[int]int),
			storageCloser:  nil,
		}
		if item.FilePriorities != "" {
			_ = json.Unmarshal([]byte(item.FilePriorities), &entry.filePriorities)
		}
		c.mu.Lock()
		c.torrents[item.Hash] = entry
		c.mu.Unlock()
		return nil, nil
	}

	spec, err := anacrolix.TorrentSpecFromMagnetUri(item.Magnet)
	if err != nil {
		return nil, fmt.Errorf("parse persisted magnet: %w", err)
	}
	fc := newTorrentStorage(item.Destination, c.pieceCompletion)
	spec.Storage = fc
	spec.DisallowDataDownload = true
	spec.DisallowDataUpload = true
	t, _, err := c.client.AddTorrentSpec(spec)
	if err != nil {
		fc.Close()
		return nil, fmt.Errorf("add torrent: %w", err)
	}
	if item.Name != "" {
		t.SetDisplayName(item.Name)
	}
	filePriorities := make(map[int]int)
	if item.FilePriorities != "" {
		if err := json.Unmarshal([]byte(item.FilePriorities), &filePriorities); err != nil {
			c.logger.Warn().Err(err).Str("hash", item.Hash).Msg("builtin torrent: invalid persisted file priorities")
			filePriorities = make(map[int]int)
		}
	}
	entry := &torrentEntry{
		client: c, model: item, torrent: t, lastSample: time.Now(), sequentialStart: -1,
		filePriorities: filePriorities,
		storageCloser:  fc,
	}
	t.SetOnWriteChunkError(func(err error) {
		if err != nil {
			prevErr := entry.getWriteError()
			if prevErr == nil || prevErr.Error() != err.Error() {
				entry.setWriteError(err)
				c.logger.Error().Err(err).Str("hash", item.Hash).Msg("builtin torrent: write chunk error")
				t.DisallowDataDownload()
				go c.reconcileQueue()
			}
		}
	})
	c.mu.Lock()
	c.torrents[item.Hash] = entry
	c.mu.Unlock()
	return t, nil
}

func (c *Client) TorrentExists(hash string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.torrents[strings.ToLower(hash)]
	return ok
}

func (c *Client) RemoveTorrent(hash string, deleteFiles bool) error {
	defer util.HandlePanicInModuleThen("builtin_client/RemoveTorrent", func() {})
	hash = strings.ToLower(hash)
	c.mu.RLock()
	entry := c.torrents[hash]
	c.mu.RUnlock()
	if entry == nil {
		return errors.New("torrent not found")
	}
	if entry.torrent != nil {
		entry.torrent.DisallowDataDownload()
		entry.torrent.DisallowDataUpload()
	}
	var paths []string
	var root string
	var err error
	if deleteFiles {
		if entry.torrent != nil {
			paths, root, err = torrentFilePaths(entry.model.Destination, entry.torrent)
			if err != nil {
				c.logger.Warn().Err(err).Str("hash", hash).Msg("builtin torrent: could not determine file paths for deletion")
				root, err = torrentRootFromModel(entry.model.Destination, entry.model.Name)
				if err != nil {
					c.logger.Warn().Err(err).Str("hash", hash).Msg("builtin torrent: could not determine fallback root for deletion")
					root = ""
				}
				err = nil
			}
		} else {
			root, err = torrentRootFromModel(entry.model.Destination, entry.model.Name)
			if err != nil {
				c.logger.Warn().Err(err).Str("hash", hash).Msg("builtin torrent: could not determine fallback root for deletion")
				root = ""
				err = nil
			}
		}
	}
	c.removeRuntime(hash)
	if len(paths) > 0 || root != "" {
		if deleteErr := removeTorrentFiles(paths, root); deleteErr != nil {
			c.logger.Warn().Err(deleteErr).Str("hash", hash).Msg("builtin torrent: some files or directories could not be removed from disk")
		}
	}
	if err = c.database.DeleteLocalTorrent(hash); err != nil {
		return err
	}
	c.compactQueue()
	c.reconcileQueue()
	return nil
}

func (c *Client) removeRuntime(hash string) {
	c.mu.Lock()
	entry := c.torrents[hash]
	delete(c.torrents, hash)
	c.mu.Unlock()
	if entry != nil {
		if entry.torrent != nil {
			entry.torrent.Drop()
		}
		if entry.storageCloser != nil {
			_ = entry.storageCloser.Close()
		}
	}
}

func torrentFilePaths(destination string, t *anacrolix.Torrent) ([]string, string, error) {
	if t.Info() == nil {
		return nil, "", errors.New("torrent metadata is not available; files were not removed")
	}
	base, err := filepath.Abs(destination)
	if err != nil {
		return nil, "", err
	}
	paths := make([]string, 0, len(t.Files()))
	for _, file := range t.Files() {
		path := filepath.Join(base, filepath.FromSlash(file.Path()))
		rel, err := filepath.Rel(base, path)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return nil, "", errors.New("torrent file path escapes destination")
		}
		paths = append(paths, path)
	}
	root := filepath.Join(base, filepath.FromSlash(t.Info().BestName()))
	return paths, root, nil
}

func torrentRootFromModel(destination, name string) (string, error) {
	if name == "" {
		return "", errors.New("torrent name is empty")
	}
	base, err := filepath.Abs(destination)
	if err != nil {
		return "", err
	}
	root := filepath.Join(base, filepath.FromSlash(name))
	root = filepath.Clean(root)
	rel, err := filepath.Rel(base, root)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", errors.New("torrent root path escapes destination")
	}
	return root, nil
}

func removeTorrentFiles(paths []string, root string) error {
	var retErr error
	for _, path := range paths {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			retErr = errors.Join(retErr, err)
		}
		for dir := filepath.Dir(path); root != "" && strings.HasPrefix(dir, root+string(filepath.Separator)); dir = filepath.Dir(dir) {
			_ = os.Remove(dir)
		}
	}
	if root != "" {
		if err := os.RemoveAll(root); err != nil && !os.IsNotExist(err) {
			retErr = errors.Join(retErr, err)
		}
	}
	return retErr
}

func (c *Client) PauseTorrent(hash string) error {
	defer util.HandlePanicInModuleThen("builtin_client/PauseTorrent", func() {})
	return c.setPaused(hash, true)
}

func (c *Client) ResumeTorrent(hash string) error {
	defer util.HandlePanicInModuleThen("builtin_client/ResumeTorrent", func() {})
	return c.setPaused(hash, false)
}

func (c *Client) setPaused(hash string, paused bool) error {
	defer util.HandlePanicInModuleThen("builtin_client/setPaused", func() {})
	hash = strings.ToLower(hash)
	c.mu.Lock()
	entry := c.torrents[hash]
	if entry == nil {
		c.mu.Unlock()
		return errors.New("torrent not found")
	}

	if paused {
		entry.model.Paused = true
		entry.model.ForceStart = false
		t := entry.torrent
		closer := entry.storageCloser
		if t != nil {
			entry.model.Length = t.Length()
			entry.model.Completed = t.BytesCompleted()
		}
		entry.torrent = nil
		entry.storageCloser = nil
		c.mu.Unlock()
		if t != nil {
			t.Drop()
		}
		if closer != nil {
			_ = closer.Close()
		}
	} else {
		// Resume: Check directory existence first!
		if _, err := os.Stat(entry.model.Destination); err != nil {
			c.mu.Unlock()
			if os.IsNotExist(err) || errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("cannot resume: save directory not found (%s)", entry.model.Destination)
			}
			return fmt.Errorf("cannot resume: %w", err)
		}

		entry.model.Paused = false
		entry.setWriteError(nil)
		c.mu.Unlock()

		_, err := c.addPersisted(entry.model)
		if err != nil {
			c.mu.Lock()
			entry.model.Paused = true
			c.mu.Unlock()
			return err
		}
	}
	if err := c.database.UpdateLocalTorrent(hash, map[string]interface{}{
		"paused":      paused,
		"force_start": entry.model.ForceStart,
		"length":      entry.model.Length,
		"completed":   entry.model.Completed,
	}); err != nil {
		return err
	}
	c.reconcileQueue()
	return nil
}

func (c *Client) PauseAll() error {
	defer util.HandlePanicInModuleThen("builtin_client/PauseAll", func() {})
	for _, snapshot := range c.Snapshots() {
		if err := c.PauseTorrent(snapshot.Hash); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) ResumeAll() error {
	defer util.HandlePanicInModuleThen("builtin_client/ResumeAll", func() {})
	for _, snapshot := range c.Snapshots() {
		if err := c.ResumeTorrent(snapshot.Hash); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) SetForceStart(hash string, enabled bool) error {
	defer util.HandlePanicInModuleThen("builtin_client/SetForceStart", func() {})
	hash = strings.ToLower(hash)
	c.mu.Lock()
	entry := c.torrents[hash]
	if entry == nil {
		c.mu.Unlock()
		return errors.New("torrent not found")
	}
	entry.model.ForceStart = enabled
	if enabled {
		entry.model.Paused = false
	}
	c.mu.Unlock()
	if err := c.database.UpdateLocalTorrent(hash, map[string]interface{}{"force_start": enabled, "paused": entry.model.Paused}); err != nil {
		return err
	}
	c.reconcileQueue()
	return nil
}

func (c *Client) MoveQueue(hash string, direction int) error {
	defer util.HandlePanicInModuleThen("builtin_client/MoveQueue", func() {})
	hash = strings.ToLower(hash)
	c.mu.Lock()
	entries := c.sortedEntriesLocked()
	index := -1
	for i, entry := range entries {
		if entry.model.Hash == hash {
			index = i
			break
		}
	}
	target := index + direction
	if index < 0 || target < 0 || target >= len(entries) {
		c.mu.Unlock()
		return nil
	}
	entries[index], entries[target] = entries[target], entries[index]
	for i, entry := range entries {
		entry.model.QueueIndex = i
	}
	c.mu.Unlock()
	for _, entry := range entries {
		if err := c.database.UpdateLocalTorrent(entry.model.Hash, map[string]interface{}{"queue_index": entry.model.QueueIndex}); err != nil {
			return err
		}
	}
	c.reconcileQueue()
	return nil
}

func (c *Client) SetLimits(downloadKB, uploadKB int) {
	setRateLimit(c.downloadLimiter, downloadKB)
	setRateLimit(c.uploadLimiter, uploadKB)
}

func setRateLimit(limiter *rate.Limiter, limitKB int) {
	if limitKB <= 0 {
		limiter.SetLimit(rate.Inf)
		limiter.SetBurst(1 << 20)
		return
	}
	bytesPerSecond := limitKB * 1024
	limiter.SetLimit(rate.Limit(bytesPerSecond))
	limiter.SetBurst(max(bytesPerSecond, 1<<20))
}

func (c *Client) SetSequential(hash string, enabled bool) error {
	defer util.HandlePanicInModuleThen("builtin_client/SetSequential", func() {})
	entry, err := c.getEntry(hash)
	if err != nil {
		return err
	}
	c.mu.Lock()
	entry.model.Sequential = enabled
	entry.sequentialInitialized = false
	entry.sequentialStart = -1
	c.mu.Unlock()
	if err := c.database.UpdateLocalTorrent(entry.model.Hash, map[string]interface{}{"sequential": enabled}); err != nil {
		return err
	}
	if entry.torrent != nil {
		if !enabled && entry.torrent.Info() != nil {
			entry.torrent.DownloadAll()
		}
		c.reconcileQueue()
	}
	return nil
}

func (c *Client) SetFilePriority(hash string, index, priority int) error {
	defer util.HandlePanicInModuleThen("builtin_client/SetFilePriority", func() {})
	entry, err := c.getEntry(hash)
	if err != nil {
		return err
	}
	if entry.torrent == nil {
		c.mu.Lock()
		entry.filePriorities[index] = priority
		encoded, err := json.Marshal(entry.filePriorities)
		if err == nil {
			entry.model.FilePriorities = string(encoded)
		}
		c.mu.Unlock()
		if err != nil {
			return err
		}
		return c.database.UpdateLocalTorrent(entry.model.Hash, map[string]interface{}{"file_priorities": entry.model.FilePriorities})
	}
	if entry.torrent.Info() == nil {
		return errors.New("torrent metadata is not available")
	}
	files := entry.torrent.Files()
	if index < 0 || index >= len(files) {
		return errors.New("file index is out of range")
	}
	if priority < 0 || priority > 2 {
		return errors.New("file priority must be 0, 1 or 2")
	}
	c.mu.Lock()
	entry.filePriorities[index] = priority
	entry.sequentialInitialized = false
	entry.sequentialStart = -1
	encoded, err := json.Marshal(entry.filePriorities)
	if err == nil {
		entry.model.FilePriorities = string(encoded)
	}
	c.mu.Unlock()
	if err != nil {
		return err
	}
	applyFilePriorities(entry, entry.torrent)
	return c.database.UpdateLocalTorrent(entry.model.Hash, map[string]interface{}{"file_priorities": entry.model.FilePriorities})
}

func (c *Client) AddTracker(hash, tracker string) error {
	defer util.HandlePanicInModuleThen("builtin_client/AddTracker", func() {})
	entry, err := c.getEntry(hash)
	if err != nil {
		return err
	}
	if entry.torrent == nil {
		return errors.New("torrent is paused")
	}
	tracker = strings.TrimSpace(tracker)
	if tracker == "" {
		return errors.New("tracker URL is required")
	}
	entry.torrent.AddTrackers([][]string{{tracker}})
	return nil
}

func (c *Client) RemoveTracker(hash, tracker string) (err error) {
	defer util.HandlePanicInModuleWithError("builtin_client/RemoveTracker", &err)
	entry, err := c.getEntry(hash)
	if err != nil {
		return err
	}
	if entry.torrent == nil {
		return errors.New("torrent is paused")
	}
	trackers := trackerList(entry.torrent)
	filtered := make([][]string, 0, len(trackers))
	for _, t := range trackers {
		if t != tracker {
			filtered = append(filtered, []string{t})
		}
	}
	entry.torrent.ModifyTrackers(filtered)
	return nil
}

func (c *Client) ReannounceTorrent(hash string) (err error) {
	defer util.HandlePanicInModuleWithError("builtin_client/ReannounceTorrent", &err)
	entry, err := c.getEntry(hash)
	if err != nil {
		return err
	}
	if entry.torrent == nil {
		return errors.New("torrent is paused")
	}
	trackers := trackerList(entry.torrent)
	tiers := make([][]string, 0, len(trackers))
	for _, tracker := range trackers {
		tiers = append(tiers, []string{tracker})
	}
	entry.torrent.ModifyTrackers(nil)
	entry.torrent.AddTrackers(tiers)
	return nil
}

func (c *Client) RecheckTorrent(hash string) (err error) {
	defer util.HandlePanicInModuleWithError("builtin_client/RecheckTorrent", &err)
	entry, err := c.getEntry(hash)
	if err != nil {
		return err
	}
	if entry.torrent == nil {
		return errors.New("torrent is paused")
	}
	if entry.torrent.Info() == nil {
		return errors.New("torrent metadata is not available")
	}
	entry.setWriteError(nil)
	return entry.torrent.VerifyData()
}

func (c *Client) RenameTorrent(hash, name string) (err error) {
	defer util.HandlePanicInModuleWithError("builtin_client/RenameTorrent", &err)
	entry, err := c.getEntry(hash)
	if err != nil {
		return err
	}
	c.mu.Lock()
	entry.model.Name = name
	if entry.torrent != nil {
		entry.torrent.SetDisplayName(name)
	}
	c.mu.Unlock()
	return c.database.UpdateLocalTorrent(entry.model.Hash, map[string]interface{}{"name": name})
}

func (c *Client) MoveStorage(hash, newDestination string) (err error) {
	defer util.HandlePanicInModuleWithError("builtin_client/MoveStorage", &err)
	entry, err := c.getEntry(hash)
	if err != nil {
		return err
	}
	newDestination = filepath.Clean(newDestination)
	if newDestination == "" || newDestination == "." {
		return errors.New("new destination is required")
	}
	if err := os.MkdirAll(newDestination, 0755); err != nil {
		return err
	}
	if entry.torrent == nil {
		c.mu.Lock()
		entry.model.Destination = newDestination
		c.mu.Unlock()
		if err := c.database.UpdateLocalTorrent(entry.model.Hash, map[string]interface{}{"destination": newDestination}); err != nil {
			return err
		}
		return nil
	}
	if entry.torrent.Info() == nil {
		return errors.New("torrent metadata is not available")
	}
	wasPaused := entry.model.Paused
	entry.torrent.DisallowDataDownload()
	entry.torrent.DisallowDataUpload()
	info := entry.torrent.Info()
	oldRoot := filepath.Join(entry.model.Destination, filepath.FromSlash(info.BestName()))
	newRoot := filepath.Join(newDestination, filepath.FromSlash(info.BestName()))
	if err := os.MkdirAll(filepath.Dir(newRoot), 0755); err != nil {
		return err
	}
	if err := os.Rename(oldRoot, newRoot); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("move torrent data: %w", err)
	}
	if entry.storageCloser != nil {
		_ = entry.storageCloser.Close()
	}
	entry.torrent.Drop()
	spec := anacrolix.TorrentSpecFromMetaInfo(new(entry.torrent.Metainfo()))
	fc := newTorrentStorage(newDestination, c.pieceCompletion)
	spec.Storage = fc
	spec.DisallowDataDownload = true
	spec.DisallowDataUpload = wasPaused
	readded, _, err := c.client.AddTorrentSpec(spec)
	if err != nil {
		fc.Close()
		return err
	}
	c.mu.Lock()
	entry.torrent = readded
	entry.model.Destination = newDestination
	entry.lastSample = time.Now()
	entry.lastDownload = 0
	entry.lastUpload = 0
	entry.sequentialInitialized = false
	entry.sequentialStart = -1
	entry.storageCloser = fc
	c.mu.Unlock()
	if err := c.database.UpdateLocalTorrent(entry.model.Hash, map[string]interface{}{"destination": newDestination}); err != nil {
		return err
	}
	c.reconcileQueue()
	return nil
}

func (c *Client) GetTorrentDetails(hash string) (*TorrentDetails, error) {
	defer util.HandlePanicInModuleThen("builtin_client/GetTorrentDetails", func() {})
	entry, err := c.getEntry(hash)
	if err != nil {
		return nil, err
	}
	details := &TorrentDetails{Torrent: c.snapshotEntry(entry)}
	if entry.torrent == nil {
		return details, nil
	}
	filePriorities := c.filePrioritiesSnapshot(entry)
	if entry.torrent.Info() != nil {
		for index, file := range entry.torrent.Files() {
			length := file.Length()
			completed := file.BytesCompleted()
			progress := 0.0
			if length > 0 {
				progress = float64(completed) / float64(length)
			}
			priority := 1
			if persistedPriority, ok := filePriorities[index]; ok {
				priority = persistedPriority
			}
			details.Files = append(details.Files, FileDetails{
				Index: index, Path: file.DisplayPath(), Length: length, Completed: completed,
				Progress: progress, Priority: priority,
			})
		}
	}
	details.Trackers = trackerList(entry.torrent)
	for _, peer := range entry.torrent.PeerConns() {
		clientName, _ := peer.PeerClientName.Load().(string)
		address := ""
		if peer.RemoteAddr != nil {
			address = peer.RemoteAddr.String()
		}
		details.Peers = append(details.Peers, PeerDetails{Address: address, Client: clientName})
	}
	return details, nil
}

func trackerList(t *anacrolix.Torrent) []string {
	seen := make(map[string]struct{})
	var trackers []string
	meta := t.Metainfo()
	for _, tier := range meta.UpvertedAnnounceList() {
		for _, tracker := range tier {
			if _, ok := seen[tracker]; ok || tracker == "" {
				continue
			}
			seen[tracker] = struct{}{}
			trackers = append(trackers, tracker)
		}
	}
	return trackers
}

func (c *Client) Snapshots() []TorrentSnapshot {
	defer util.HandlePanicInModuleThen("builtin_client/Snapshots", func() {})
	c.mu.RLock()
	entries := c.sortedEntriesLocked()
	allowed := c.allowedEntriesLocked()
	c.mu.RUnlock()
	ret := make([]TorrentSnapshot, 0, len(entries))
	for _, entry := range entries {
		ret = append(ret, c.snapshotEntryA(entry, allowed))
	}
	return ret
}

func (c *Client) snapshotEntry(entry *torrentEntry) TorrentSnapshot {
	return c.snapshotEntryA(entry, c.SnapshotsOrder())
}

func (c *Client) snapshotEntryA(entry *torrentEntry, allowed map[string]bool) TorrentSnapshot {
	var stats anacrolix.TorrentStats
	length := entry.model.Length
	completed := entry.model.Completed
	if entry.torrent != nil {
		stats = entry.torrent.Stats()
		if entry.torrent.Info() != nil {
			length = entry.torrent.Length()
			completed = entry.torrent.BytesCompleted()

			entry.model.Length = length
			entry.model.Completed = completed
		}
	}
	errStr := ""
	if wErr := entry.getWriteError(); wErr != nil {
		errStr = wErr.Error()
	}
	return TorrentSnapshot{
		Name:        displayName(entry),
		Hash:        entry.model.Hash,
		Destination: entry.model.Destination,
		Paused:      entry.model.Paused,
		Queued:      !entry.model.Paused && !entry.model.ForceStart && !dataDownloadAllowed(entry, c.maxActiveDownloads, allowed),
		ForceStart:  entry.model.ForceStart,
		Sequential:  entry.model.Sequential,
		QueueIndex:  entry.model.QueueIndex,
		Length:      length,
		Completed:   completed,
		DownSpeed:   entry.downSpeed,
		UpSpeed:     entry.upSpeed,
		Seeds:       stats.ConnectedSeeders,
		Peers:       stats.ActivePeers,
		Downloaded:  stats.BytesReadUsefulData.Int64(),
		Uploaded:    stats.BytesWrittenData.Int64(),
		AddedAt:     entry.model.CreatedAt,
		Error:       errStr,
	}
}

func (c *Client) SnapshotsOrder() map[string]bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.allowedEntriesLocked()
}

func dataDownloadAllowed(entry *torrentEntry, _ int, allowed map[string]bool) bool {
	return allowed[entry.model.Hash]
}

func displayName(entry *torrentEntry) string {
	if entry.model.Name != "" {
		return entry.model.Name
	}
	if entry.torrent != nil {
		if name := entry.torrent.Name(); name != "" {
			return name
		}
	}
	return entry.model.Hash
}

func (c *Client) getEntry(hash string) (*torrentEntry, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry := c.torrents[strings.ToLower(hash)]
	if entry == nil {
		return nil, errors.New("torrent not found")
	}
	return entry, nil
}

func (c *Client) sortedEntriesLocked() []*torrentEntry {
	entries := make([]*torrentEntry, 0, len(c.torrents))
	for _, entry := range c.torrents {
		entries = append(entries, entry)
	}
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].model.QueueIndex < entries[j].model.QueueIndex
	})
	return entries
}

func (c *Client) allowedEntriesLocked() map[string]bool {
	allowed := make(map[string]bool, len(c.torrents))
	active := 0
	for _, entry := range c.sortedEntriesLocked() {
		if entry.model.Paused || entry.torrent == nil {
			continue
		}
		if entry.getWriteError() != nil {
			continue
		}
		if length := entry.torrent.Length(); length > 0 && entry.torrent.BytesCompleted() >= length {
			allowed[entry.model.Hash] = true
			continue
		}
		if entry.model.ForceStart || c.maxActiveDownloads <= 0 || active < c.maxActiveDownloads {
			allowed[entry.model.Hash] = true
			if !entry.model.ForceStart {
				active++
			}
		}
	}
	return allowed
}

func (c *Client) reconcileQueue() {
	defer util.HandlePanicInModuleThen("builtin_client/reconcileQueue", func() {})
	c.mu.RLock()
	entries := c.sortedEntriesLocked()
	allowed := c.allowedEntriesLocked()
	c.mu.RUnlock()
	for _, entry := range entries {
		c.mu.RLock()
		t := entry.torrent
		isPaused := entry.model.Paused
		isAllowed := allowed[entry.model.Hash]
		isSequential := entry.model.Sequential
		c.mu.RUnlock()

		if t == nil {
			continue
		}
		if isPaused || !isAllowed {
			t.DisallowDataDownload()
			t.DisallowDataUpload()
			continue
		}
		t.AllowDataDownload()
		t.AllowDataUpload()
		if t.Info() != nil {
			if isSequential {
				applySequentialPriorities(entry, t)
			} else {
				t.DownloadAll()
				prioritizeBoundaries(entry, t)
			}
			applyFilePriorities(entry, t)
		}
	}
}

func applySequentialPriorities(entry *torrentEntry, t *anacrolix.Torrent) {
	if t.NumPieces() == 0 {
		return
	}
	firstMissing := -1
	wanted := wantedPieceSet(entry, t)
	for i := 0; i < t.NumPieces(); i++ {
		if !wanted[i] {
			t.Piece(i).SetPriority(anacrolix.PiecePriorityNone)
			continue
		}
		state := t.Piece(i).State()
		if !state.Complete && firstMissing == -1 {
			firstMissing = i
		}
		if !entry.sequentialInitialized {
			t.Piece(i).SetPriority(anacrolix.PiecePriorityNone)
		}
	}
	if firstMissing < 0 {
		return
	}
	if entry.sequentialInitialized && entry.sequentialStart == firstMissing {
		return
	}
	end := min(firstMissing+sequentialWindowPieces, t.NumPieces())
	for i := firstMissing; i < end; i++ {
		priority := anacrolix.PiecePriorityNormal
		if i == firstMissing {
			priority = anacrolix.PiecePriorityNow
		} else if i < firstMissing+4 {
			priority = anacrolix.PiecePriorityReadahead
		}
		t.Piece(i).SetPriority(priority)
	}
	prioritizeBoundaries(entry, t)
	entry.sequentialInitialized = true
	entry.sequentialStart = firstMissing
}

func applyFilePriorities(entry *torrentEntry, t *anacrolix.Torrent) {
	if t.Info() == nil {
		return
	}
	files := t.Files()
	priorities := entry.client.filePrioritiesSnapshot(entry)
	for index, priority := range priorities {
		if index < 0 || index >= len(files) {
			continue
		}
		file := files[index]
		switch priority {
		case 0:
			file.SetPriority(anacrolix.PiecePriorityNone)
			if !entry.model.Sequential {
				t.CancelPieces(file.BeginPieceIndex(), file.EndPieceIndex())
			}
		case 2:
			file.SetPriority(anacrolix.PiecePriorityHigh)
		default:
			if !entry.model.Sequential {
				file.SetPriority(anacrolix.PiecePriorityNone)
			}
		}
	}
	if !entry.model.Sequential {
		wanted := wantedPieceSet(entry, t)
		for piece := 0; piece < t.NumPieces(); piece++ {
			if wanted[piece] && t.Piece(piece).State().Priority == anacrolix.PiecePriorityNone {
				t.Piece(piece).SetPriority(anacrolix.PiecePriorityNormal)
			}
		}
		prioritizeBoundaries(entry, t)
	}
}

func wantedPieceSet(entry *torrentEntry, t *anacrolix.Torrent) []bool {
	wanted := make([]bool, t.NumPieces())
	priorities := entry.client.filePrioritiesSnapshot(entry)
	for index, file := range t.Files() {
		if priority, explicitlySet := priorities[index]; explicitlySet && priority == 0 {
			continue
		}
		for piece := file.BeginPieceIndex(); piece < file.EndPieceIndex(); piece++ {
			wanted[piece] = true
		}
	}
	return wanted
}

func pieceWanted(entry *torrentEntry, t *anacrolix.Torrent, piece int) bool {
	priorities := entry.client.filePrioritiesSnapshot(entry)
	for index, file := range t.Files() {
		if piece < file.BeginPieceIndex() || piece >= file.EndPieceIndex() {
			continue
		}
		if priority, explicitlySet := priorities[index]; !explicitlySet || priority != 0 {
			return true
		}
	}
	return false
}

func (c *Client) filePrioritiesSnapshot(entry *torrentEntry) map[int]int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	priorities := make(map[int]int, len(entry.filePriorities))
	for index, priority := range entry.filePriorities {
		priorities[index] = priority
	}
	return priorities
}

func prioritizeBoundaries(entry *torrentEntry, t *anacrolix.Torrent) {
	if t.NumPieces() == 0 {
		return
	}
	if pieceWanted(entry, t, 0) {
		t.Piece(0).SetPriority(anacrolix.PiecePriorityHigh)
	}
	if pieceWanted(entry, t, t.NumPieces()-1) {
		t.Piece(t.NumPieces() - 1).SetPriority(anacrolix.PiecePriorityHigh)
	}
}

func (c *Client) runScheduler() {
	defer util.HandlePanicInModuleThen("builtin_client/runScheduler", func() {})
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-c.closeCh:
			return
		case now := <-ticker.C:
			c.sampleRates(now)
			c.reconcileQueue()
		}
	}
}

func (c *Client) sampleRates(now time.Time) {
	defer util.HandlePanicInModuleThen("builtin_client/sampleRates", func() {})
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, entry := range c.torrents {
		if entry.torrent == nil {
			continue
		}
		stats := entry.torrent.Stats()
		downloaded := stats.BytesReadUsefulData.Int64()
		uploaded := stats.BytesWrittenData.Int64()
		elapsed := now.Sub(entry.lastSample).Seconds()
		if elapsed > 0 {
			entry.downSpeed = max(0, int64(float64(downloaded-entry.lastDownload)/elapsed))
			entry.upSpeed = max(0, int64(float64(uploaded-entry.lastUpload)/elapsed))
		}
		entry.lastDownload = downloaded
		entry.lastUpload = uploaded
		entry.lastSample = now

		if entry.torrent.Info() != nil && entry.model.Name == "" {
			if name := entry.torrent.Name(); name != "" {
				entry.model.Name = name
				go func(hash, name string) {
					_ = c.database.UpdateLocalTorrent(hash, map[string]interface{}{"name": name})
				}(entry.model.Hash, name)
			}
		}

		if _, err := os.Stat(entry.model.Destination); err != nil {
			if os.IsNotExist(err) || errors.Is(err, os.ErrNotExist) {
				entry.setWriteError(fmt.Errorf("save directory not found: %s", entry.model.Destination))
			} else {
				entry.setWriteError(fmt.Errorf("save directory error: %w", err))
			}
		} else {
			wErr := entry.getWriteError()
			if wErr != nil && strings.HasPrefix(wErr.Error(), "save directory") {
				entry.setWriteError(nil)
			}
		}
	}
}

func (c *Client) compactQueue() {
	defer util.HandlePanicInModuleThen("builtin_client/compactQueue", func() {})
	c.mu.Lock()
	entries := c.sortedEntriesLocked()
	for i, entry := range entries {
		entry.model.QueueIndex = i
	}
	c.mu.Unlock()
	for _, entry := range entries {
		_ = c.database.UpdateLocalTorrent(entry.model.Hash, map[string]interface{}{"queue_index": entry.model.QueueIndex})
	}
}
