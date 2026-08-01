package dummy

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"seanime/internal/database/models"
	"seanime/internal/debrid/debrid"
	"seanime/internal/util"
	httputil "seanime/internal/util/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

const (
	id   = "dummy"
	name = "Dummy Debrid"
)

const (
	chunkSize  = 64 * 1024
	progressMs = 250
)

type SettingsProvider interface {
	GetDummyDebridSettings() (*models.DummyDebridSettings, bool)
}

type Dummy struct {
	logger           *zerolog.Logger
	settingsProvider SettingsProvider

	mu      sync.Mutex
	server  *http.Server
	baseURL string
	items   map[string]*debrid.TorrentItem
}

func New(logger *zerolog.Logger, settingsProvider SettingsProvider) debrid.Provider {
	return &Dummy{
		logger:           logger,
		settingsProvider: settingsProvider,
		items:            make(map[string]*debrid.TorrentItem),
	}
}

func (d *Dummy) GetSettings() debrid.Settings {
	return debrid.Settings{ID: id, Name: name}
}

func (d *Dummy) Authenticate(_ string) error {
	settings, err := d.settings()
	if err != nil {
		return err
	}
	if !settings.Enabled {
		return errors.New("dummy debrid profile is disabled")
	}
	return d.startServer()
}

func (d *Dummy) AddTorrent(opts debrid.AddTorrentOptions) (string, error) {
	settings, err := d.settings()
	if err != nil {
		return "", err
	}
	if !settings.Enabled {
		return "", errors.New("dummy debrid profile is disabled")
	}

	files := d.files(settings)
	hash := hashFromOptions(opts.InfoHash, opts.MagnetLink)
	torrentID := "dummy-" + hash
	item := d.newItem(torrentID, hash, settings, d.selectFiles(files, opts.SelectFileId), debrid.TorrentItemStatusDownloading, 0)

	d.mu.Lock()
	d.items[torrentID] = item
	d.mu.Unlock()

	return torrentID, nil
}

func (d *Dummy) GetTorrentStreamUrl(ctx context.Context, opts debrid.StreamTorrentOptions, itemCh chan debrid.TorrentItem) (string, error) {
	settings, err := d.settings()
	if err != nil {
		return "", err
	}
	if !settings.Enabled {
		return "", errors.New("dummy debrid profile is disabled")
	}

	files := d.files(settings)
	file, err := findFile(files, opts.FileId)
	if err != nil {
		return "", err
	}

	item := d.getOrCreateItem(opts.ID, "", settings, files)
	if err := d.awaitReady(ctx, settings, item, itemCh); err != nil {
		return "", err
	}

	return d.fileURL(file)
}

func (d *Dummy) GetTorrentDownloadUrl(opts debrid.DownloadTorrentOptions) (string, error) {
	settings, err := d.settings()
	if err != nil {
		return "", err
	}
	if !settings.Enabled {
		return "", errors.New("dummy debrid profile is disabled")
	}

	files := d.files(settings)
	file, err := findFile(files, opts.FileId)
	if err != nil {
		return "", err
	}

	return d.fileURL(file)
}

func (d *Dummy) GetInstantAvailability(hashes []string) map[string]debrid.TorrentItemInstantAvailability {
	settings, err := d.settings()
	if err != nil || !settings.Enabled || !settings.Cached {
		return map[string]debrid.TorrentItemInstantAvailability{}
	}

	files := d.files(settings)
	cachedFiles := make(map[string]*debrid.CachedFile, len(files))
	for _, file := range files {
		cachedFiles[file.ID] = &debrid.CachedFile{
			Size: file.Size,
			Name: file.Name,
		}
	}

	ret := make(map[string]debrid.TorrentItemInstantAvailability, len(hashes))
	for _, hash := range hashes {
		if hash == "" {
			continue
		}
		ret[hash] = debrid.TorrentItemInstantAvailability{CachedFiles: cachedFiles}
	}
	return ret
}

func (d *Dummy) GetTorrent(id string) (*debrid.TorrentItem, error) {
	d.mu.Lock()
	item, found := d.items[id]
	d.mu.Unlock()
	if found {
		return copyItem(item), nil
	}

	settings, err := d.settings()
	if err != nil {
		return nil, err
	}
	if !settings.Enabled {
		return nil, errors.New("dummy debrid profile is disabled")
	}

	item = d.newItem(id, "", settings, d.files(settings), debrid.TorrentItemStatusCompleted, 100)
	return item, nil
}

func (d *Dummy) GetTorrentInfo(opts debrid.GetTorrentInfoOptions) (*debrid.TorrentInfo, error) {
	settings, err := d.settings()
	if err != nil {
		return nil, err
	}
	if !settings.Enabled {
		return nil, errors.New("dummy debrid profile is disabled")
	}

	files := d.files(settings)
	hash := hashFromOptions(opts.InfoHash, opts.MagnetLink)
	size := totalSize(files)
	retFiles := make([]*debrid.TorrentItemFile, 0, len(files))
	for i, file := range files {
		retFiles = append(retFiles, &debrid.TorrentItemFile{
			ID:    file.ID,
			Index: i,
			Name:  file.Name,
			Path:  file.Path,
			Size:  file.Size,
		})
	}

	return &debrid.TorrentInfo{
		Name:  profileName(settings),
		Hash:  hash,
		Size:  size,
		Files: retFiles,
	}, nil
}

func (d *Dummy) GetTorrents() ([]*debrid.TorrentItem, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	ret := make([]*debrid.TorrentItem, 0, len(d.items))
	for _, item := range d.items {
		ret = append(ret, copyItem(item))
	}
	return ret, nil
}

func (d *Dummy) DeleteTorrent(id string) error {
	d.mu.Lock()
	delete(d.items, id)
	d.mu.Unlock()
	return nil
}

func (d *Dummy) Close() error {
	d.mu.Lock()
	server := d.server
	d.server = nil
	d.baseURL = ""
	d.mu.Unlock()

	if server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return server.Shutdown(ctx)
}

func (d *Dummy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fileID, ok := fileIDFromPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	settings, err := d.settings()
	if err != nil || !settings.Enabled {
		http.Error(w, "dummy debrid profile is unavailable", http.StatusServiceUnavailable)
		return
	}

	file, err := findFile(d.files(settings), fileID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	localPath := file.LocalFilePath
	if localPath == "" {
		localPath = settings.FallbackFilePath
	}

	handle, err := os.Open(localPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defer handle.Close()

	stat, err := handle.Stat()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	size := stat.Size()
	ranges, err := httputil.ParseRange(r.Header.Get("Range"), size)
	if err != nil {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", size))
		http.Error(w, err.Error(), http.StatusRequestedRangeNotSatisfiable)
		return
	}

	start := int64(0)
	length := size
	status := http.StatusOK
	if len(ranges) > 0 {
		start = ranges[0].Start
		length = ranges[0].Length
		status = http.StatusPartialContent
		w.Header().Set("Content-Range", ranges[0].ContentRange(size))
	}

	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Type", contentType(file.Name))
	w.Header().Set("Content-Length", strconv.FormatInt(length, 10))
	w.WriteHeader(status)

	if r.Method == http.MethodHead || length == 0 {
		return
	}

	if _, err = handle.Seek(start, io.SeekStart); err != nil {
		return
	}
	_ = d.copyThrottled(r.Context(), w, handle, length, settings)
}

func (d *Dummy) settings() (*models.DummyDebridSettings, error) {
	if d.settingsProvider == nil {
		return nil, errors.New("dummy debrid settings provider is not set")
	}

	settings, found := d.settingsProvider.GetDummyDebridSettings()
	if !found || settings == nil {
		return nil, errors.New("dummy debrid settings not found")
	}

	return normalizeSettings(settings), nil
}

func (d *Dummy) startServer() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.server != nil {
		return nil
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}

	server := &http.Server{Handler: d}
	d.server = server
	d.baseURL = "http://" + listener.Addr().String()

	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) && d.logger != nil {
			d.logger.Warn().Err(err).Msg("dummy debrid: HTTP server stopped")
		}
	}()

	return nil
}

func (d *Dummy) fileURL(file models.DummyDebridFile) (string, error) {
	if err := d.startServer(); err != nil {
		return "", err
	}

	d.mu.Lock()
	baseURL := d.baseURL
	d.mu.Unlock()

	return fmt.Sprintf("%s/dummy-debrid/files/%s/%s", baseURL, url.PathEscape(file.ID), url.PathEscape(file.Name)), nil
}

func (d *Dummy) awaitReady(ctx context.Context, settings *models.DummyDebridSettings, item *debrid.TorrentItem, itemCh chan debrid.TorrentItem) error {
	readyDelay := time.Duration(settings.ReadyDelayMs) * time.Millisecond
	if readyDelay <= 0 {
		d.updateItem(item, debrid.TorrentItemStatusCompleted, 100, itemCh)
		return nil
	}

	pi := time.Duration(settings.ProgressIntervalMs) * time.Millisecond
	if pi <= 0 {
		pi = progressMs * time.Millisecond
	}

	start := time.Now()
	timer := time.NewTimer(readyDelay)
	ticker := time.NewTicker(pi)
	defer timer.Stop()
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			progress := int(float64(time.Since(start)) / float64(readyDelay) * 100)
			if progress > 99 {
				progress = 99
			}
			d.updateItem(item, debrid.TorrentItemStatusDownloading, progress, itemCh)
		case <-timer.C:
			d.updateItem(item, debrid.TorrentItemStatusCompleted, 100, itemCh)
			return nil
		}
	}
}

func (d *Dummy) updateItem(item *debrid.TorrentItem, status debrid.TorrentItemStatus, progress int, itemCh chan debrid.TorrentItem) {
	item.Status = status
	item.CompletionPercentage = progress
	item.IsReady = status == debrid.TorrentItemStatusCompleted

	d.mu.Lock()
	d.items[item.ID] = copyItem(item)
	d.mu.Unlock()

	if itemCh != nil {
		select {
		case itemCh <- *copyItem(item):
		default:
		}
	}
}

func (d *Dummy) getOrCreateItem(itemID string, hash string, settings *models.DummyDebridSettings, files []models.DummyDebridFile) *debrid.TorrentItem {
	d.mu.Lock()
	item, found := d.items[itemID]
	d.mu.Unlock()
	if found {
		return item
	}

	if itemID == "" {
		itemID = "dummy-" + hashFromOptions(hash, profileName(settings))
	}

	item = d.newItem(itemID, hash, settings, files, debrid.TorrentItemStatusDownloading, 0)
	d.mu.Lock()
	d.items[itemID] = item
	d.mu.Unlock()
	return item
}

func (d *Dummy) newItem(itemID string, hash string, settings *models.DummyDebridSettings, files []models.DummyDebridFile, status debrid.TorrentItemStatus, progress int) *debrid.TorrentItem {
	retFiles := make([]*debrid.TorrentItemFile, 0, len(files))
	for i, file := range files {
		retFiles = append(retFiles, &debrid.TorrentItemFile{
			ID:    file.ID,
			Index: i,
			Name:  file.Name,
			Path:  file.Path,
			Size:  file.Size,
		})
	}

	size := totalSize(files)
	return &debrid.TorrentItem{
		ID:                   itemID,
		Name:                 profileName(settings),
		Hash:                 hash,
		Size:                 size,
		FormattedSize:        util.Bytes(uint64(size)),
		CompletionPercentage: progress,
		Status:               status,
		AddedAt:              time.Now().Format(time.RFC3339),
		IsReady:              status == debrid.TorrentItemStatusCompleted,
		Files:                retFiles,
	}
}

func (d *Dummy) files(settings *models.DummyDebridSettings) []models.DummyDebridFile {
	files := []models.DummyDebridFile(settings.Files)
	if len(files) == 0 && settings.FallbackFilePath != "" {
		files = []models.DummyDebridFile{{
			ID:            "1",
			Path:          filepath.Base(settings.FallbackFilePath),
			Name:          filepath.Base(settings.FallbackFilePath),
			EpisodeNumber: 1,
			LocalFilePath: settings.FallbackFilePath,
		}}
	}

	ret := make([]models.DummyDebridFile, 0, len(files))
	for i, file := range files {
		if file.ID == "" {
			file.ID = strconv.Itoa(i + 1)
		}
		if file.LocalFilePath == "" {
			file.LocalFilePath = settings.FallbackFilePath
		}
		if file.Name == "" {
			file.Name = filepath.Base(file.Path)
			if file.Name == "." || file.Name == string(filepath.Separator) || file.Name == "" {
				file.Name = filepath.Base(file.LocalFilePath)
			}
		}
		if file.Path == "" {
			file.Path = file.Name
		}
		if file.Size <= 0 && file.LocalFilePath != "" {
			if stat, err := os.Stat(file.LocalFilePath); err == nil {
				file.Size = stat.Size()
			}
		}
		ret = append(ret, file)
	}

	return ret
}

func (d *Dummy) selectFiles(files []models.DummyDebridFile, fileID string) []models.DummyDebridFile {
	if fileID == "" || fileID == "all" {
		return files
	}

	file, err := findFile(files, fileID)
	if err != nil {
		return files
	}
	return []models.DummyDebridFile{file}
}

func (d *Dummy) copyThrottled(ctx context.Context, w http.ResponseWriter, r io.Reader, length int64, settings *models.DummyDebridSettings) error {
	if settings.FirstByteDelayMs > 0 {
		if err := sleepContext(ctx, time.Duration(settings.FirstByteDelayMs)*time.Millisecond); err != nil {
			return err
		}
	}

	chunkSize := settings.ChunkSize
	if chunkSize <= 0 {
		chunkSize = chunkSize
	}
	buf := make([]byte, chunkSize)
	left := length
	random := rand.New(rand.NewSource(time.Now().UnixNano()))

	for left > 0 {
		if err := ctx.Err(); err != nil {
			return err
		}

		chunkLen := len(buf)
		if left < int64(chunkLen) {
			chunkLen = int(left)
		}

		read, readErr := io.ReadFull(r, buf[:chunkLen])
		if readErr != nil && readErr != io.EOF && !errors.Is(readErr, io.ErrUnexpectedEOF) {
			return readErr
		}
		if read == 0 {
			break
		}

		if settings.JitterMs > 0 {
			if err := sleepContext(ctx, time.Duration(random.Intn(settings.JitterMs+1))*time.Millisecond); err != nil {
				return err
			}
		}

		if settings.BandwidthBytesPerSecond > 0 {
			wait := time.Duration(float64(read) / float64(settings.BandwidthBytesPerSecond) * float64(time.Second))
			if err := sleepContext(ctx, wait); err != nil {
				return err
			}
		}

		written, err := w.Write(buf[:read])
		if err != nil {
			return err
		}
		if written != read {
			return io.ErrShortWrite
		}
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}

		left -= int64(written)
		if readErr == io.EOF || errors.Is(readErr, io.ErrUnexpectedEOF) {
			break
		}
	}

	return nil
}

func normalizeSettings(settings *models.DummyDebridSettings) *models.DummyDebridSettings {
	ret := *settings
	if ret.ProfileName == "" {
		ret.ProfileName = name
	}
	if ret.ReadyDelayMs < 0 {
		ret.ReadyDelayMs = 0
	}
	if ret.ProgressIntervalMs <= 0 {
		ret.ProgressIntervalMs = progressMs
	}
	if ret.ChunkSize <= 0 {
		ret.ChunkSize = chunkSize
	}
	if ret.BandwidthBytesPerSecond < 0 {
		ret.BandwidthBytesPerSecond = 0
	}
	if ret.JitterMs < 0 {
		ret.JitterMs = 0
	}
	if ret.FirstByteDelayMs < 0 {
		ret.FirstByteDelayMs = 0
	}
	return &ret
}

func findFile(files []models.DummyDebridFile, fileID string) (models.DummyDebridFile, error) {
	if fileID == "" && len(files) > 0 {
		return files[0], nil
	}

	for i, file := range files {
		if file.ID == fileID || strconv.Itoa(i) == fileID || strconv.Itoa(i+1) == fileID {
			return file, nil
		}
	}

	return models.DummyDebridFile{}, fmt.Errorf("dummy debrid file %q not found", fileID)
}

func totalSize(files []models.DummyDebridFile) int64 {
	var size int64
	for _, file := range files {
		if file.Size > 0 {
			size += file.Size
		}
	}
	return size
}

func profileName(settings *models.DummyDebridSettings) string {
	if settings.ProfileName != "" {
		return settings.ProfileName
	}
	return name
}

func hashFromOptions(infoHash string, magnet string) string {
	if infoHash != "" {
		return infoHash
	}

	if magnet != "" {
		if parsed, err := url.Parse(magnet); err == nil {
			xt := parsed.Query().Get("xt")
			if strings.HasPrefix(xt, "urn:btih:") {
				return strings.TrimPrefix(xt, "urn:btih:")
			}
		}
	}

	sum := sha1.Sum([]byte(magnet))
	return hex.EncodeToString(sum[:])
}

func fileIDFromPath(path string) (string, bool) {
	const prefix = "/dummy-debrid/files/"
	if !strings.HasPrefix(path, prefix) {
		return "", false
	}

	rest := strings.TrimPrefix(path, prefix)
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) == 0 || parts[0] == "" {
		return "", false
	}

	fileID, err := url.PathUnescape(parts[0])
	if err != nil {
		return "", false
	}

	return fileID, true
}

func contentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == ".mkv" {
		return "video/x-matroska"
	}
	if ctype := mime.TypeByExtension(ext); ctype != "" {
		return ctype
	}
	return "application/octet-stream"
}

func sleepContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}

	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func copyItem(item *debrid.TorrentItem) *debrid.TorrentItem {
	if item == nil {
		return nil
	}
	ret := *item
	if item.Files != nil {
		ret.Files = make([]*debrid.TorrentItemFile, len(item.Files))
		for i, file := range item.Files {
			if file == nil {
				continue
			}
			ret.Files[i] = new(*file)
		}
	}
	return &ret
}
