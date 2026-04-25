package debrid_client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"seanime/internal/debrid/debrid"
	"seanime/internal/events"
	"seanime/internal/hook"
	"seanime/internal/notifier"
	"seanime/internal/util"
	"seanime/internal/util/result"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	debridDownloadMaxAttempts      = 8
	debridDownloadInitialBackoff   = time.Second
	debridDownloadMaxBackoff       = 30 * time.Second
	errInvalidDownloadStatus       = errors.New("invalid download response status")
	errInvalidDownloadContentRange = errors.New("invalid download content range")
)

func (r *Repository) launchDownloadLoop(ctx context.Context) {
	r.logger.Trace().Msg("debrid: Starting download loop")
	go func() {
		for {
			select {
			case <-ctx.Done():
				r.logger.Trace().Msg("debrid: Download loop destroy request received")
				// Destroy the loop
				return
			case <-time.After(time.Minute * 1):
				// Every minute, check if there are any completed downloads
				provider, found := r.provider.Get()
				if !found {
					continue
				}

				// Get the list of completed downloads
				items, err := provider.GetTorrents()
				if err != nil {
					r.logger.Err(err).Msg("debrid: Failed to get torrents")
					continue
				}

				readyItems := make([]*debrid.TorrentItem, 0)
				for _, item := range items {
					if item.IsReady {
						readyItems = append(readyItems, item)
					}
				}

				dbItems, err := r.db.GetDebridTorrentItems()
				if err != nil {
					r.logger.Err(err).Msg("debrid: Failed to get debrid torrent items")
					continue
				}

				for _, dbItem := range dbItems {
					// Check if the item is ready for download
					for _, readyItem := range readyItems {
						if dbItem.TorrentItemID == readyItem.ID {
							r.logger.Debug().Str("torrentItemId", dbItem.TorrentItemID).Msg("debrid: Torrent is ready for download")
							// Remove the item from the database
							err = r.db.DeleteDebridTorrentItemByDbId(dbItem.ID)
							if err != nil {
								r.logger.Err(err).Msg("debrid: Failed to remove debrid torrent item")
								continue
							}
							time.Sleep(1 * time.Second)
							// Download the torrent locally
							err = r.downloadTorrentItem(readyItem.ID, readyItem.Name, dbItem.Destination)
							if err != nil {
								r.logger.Err(err).Msg("debrid: Failed to download torrent")
								continue
							}
						}
					}
				}

			}
		}
	}()
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) DownloadTorrent(item debrid.TorrentItem, destination string) error {
	return r.downloadTorrentItem(item.ID, item.Name, destination)
}

type downloadStatus struct {
	TotalBytes int64
	TotalSize  int64
}

func (r *Repository) downloadTorrentItem(tId string, torrentName string, destination string) (err error) {
	defer util.HandlePanicInModuleWithError("debrid/client/downloadTorrentItem", &err)

	provider, err := r.GetProvider()
	if err != nil {
		return err
	}

	r.logger.Debug().Str("torrentName", torrentName).Str("destination", destination).Msg("debrid: Downloading torrent")

	// Get the download URL
	downloadUrl, err := provider.GetTorrentDownloadUrl(debrid.DownloadTorrentOptions{
		ID: tId,
	})
	if err != nil {
		return err
	}

	event := &DebridLocalDownloadRequestedEvent{
		TorrentName: torrentName,
		Destination: destination,
		DownloadUrl: downloadUrl,
	}
	err = hook.GlobalHookManager.OnDebridLocalDownloadRequested().Trigger(event)
	if err != nil {
		return err
	}

	if event.DefaultPrevented {
		r.logger.Debug().Msg("debrid: Download prevented by hook")
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	r.ctxMap.Set(tId, cancel)
	if err := r.sendDownloadStartedEvent(tId, torrentName, destination, downloadUrl); err != nil {
		cancel()
		r.ctxMap.Delete(tId)
		return err
	}

	go func(ctx context.Context) {
		defer func() {
			cancel()
			r.ctxMap.Delete(tId)
		}()

		wg := sync.WaitGroup{}
		downloadUrls := strings.Split(downloadUrl, ",")
		downloadMap := result.NewMap[string, downloadStatus]()
		var failed atomic.Bool
		var cancelOnce sync.Once

		for _, url := range downloadUrls {
			wg.Add(1)
			go func(ctx context.Context, url string) {
				defer wg.Done()

				// Download the file
				ok := r.downloadFile(ctx, tId, url, destination, downloadMap)
				if !ok {
					failed.Store(true)
					cancelOnce.Do(cancel)
					return
				}
			}(ctx, url)
		}
		wg.Wait()

		if failed.Load() {
			return
		}

		r.sendDownloadCompletedEvent(tId, torrentName, destination)
		notifier.GlobalNotifier.Notify(notifier.Debrid, fmt.Sprintf("Downloaded %q", torrentName))
	}(ctx)

	return nil
}

func (r *Repository) downloadFile(ctx context.Context, tId string, downloadUrl string, destination string, downloadMap *result.Map[string, downloadStatus]) (ok bool) {
	defer util.HandlePanicInModuleThen("debrid/client/downloadFile", func() {
		ok = false
	})

	_ = os.MkdirAll(destination, os.ModePerm)

	// Download the files to a temporary folder
	//	/path/to/destination
	//		/.tmp-123456789
	tmpDirPath, err := os.MkdirTemp(destination, ".tmp-")
	if err != nil {
		r.logger.Err(err).Str("destination", destination).Msg("debrid: Failed to create temp folder")
		r.wsEventManager.SendEvent(events.ErrorToast, fmt.Sprintf("debrid: Failed to create temp folder: %v", err))
		return false
	}
	defer os.RemoveAll(tmpDirPath) // Clean up temp folder on exit

	if runtime.GOOS == "windows" {
		r.logger.Debug().Str("tmpDirPath", tmpDirPath).Msg("debrid: Hiding temp folder")
		util.HideFile(tmpDirPath)
		time.Sleep(time.Millisecond * 500)
	}

	r.logger.Debug().Str("downloadUrl", downloadUrl).Msg("debrid: Starting download")
	downloadedFile, err := r.downloadHTTPFile(ctx, tId, downloadUrl, tmpDirPath, downloadMap)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			r.logger.Debug().Msg("debrid: Download cancelled")
		} else {
			r.logger.Err(err).Str("downloadUrl", downloadUrl).Msg("debrid: Download failed")
			r.wsEventManager.SendEvent(events.ErrorToast, fmt.Sprintf("debrid: Download failed: %v", err))
		}
		r.sendDownloadCancelledEvent(tId, downloadUrl, downloadMap)
		return false
	}

	filename := downloadedFile.Filename
	ext := downloadedFile.Ext

	r.logger.Debug().Str("filename", filename).Str("ext", ext).Msg("debrid: Downloaded file ready")

	downloadMap.Delete(downloadUrl)

	if len(downloadMap.Values()) == 0 {
		r.wsEventManager.SendEvent(events.DebridDownloadProgress, map[string]interface{}{
			"status":     "downloading",
			"itemID":     tId,
			"totalBytes": "Extracting...",
			"totalSize":  "-",
			"speed":      "",
		})
	}

	r.logger.Debug().Msg("debrid: Download completed")

	switch runtime.GOOS {
	case "windows":
		time.Sleep(time.Second * 1)
	}

	// Extract the downloaded file
	var extractedDir string
	tmpDownloadedFilePath := downloadedFile.Path
	switch ext {
	case ".zip":
		//	/path/to/destination/.tmp-123456789/downlooaded_torrent.zip -> /path/to/destination/.tmp-123456789/extracted-1234/...
		extractedDir, err = unzipFile(tmpDownloadedFilePath, tmpDirPath)
		//	/path/to/destination/.tmp-123456789/downlooaded_torrent.rar -> /path/to/destination/.tmp-123456789/extracted-1234/...
		r.logger.Debug().Str("extractedDir", extractedDir).Msg("debrid: Extracted zip file")
	case ".rar":
		extractedDir, err = unrarFile(tmpDownloadedFilePath, tmpDirPath)
		r.logger.Debug().Str("extractedDir", extractedDir).Msg("debrid: Extracted rar file")
	default:
		// No extraction needed which means we downloaded a file
		//	/path/to/destination/.tmp-123456789/Episode.mkv -> /path/to/destination/Episode.mkv
		r.logger.Debug().Str("tmpDownloadedFilePath", tmpDownloadedFilePath).Str("destination", destination).Msg("debrid: No extraction needed, moving file directly")
		err = moveContentsTo(filepath.Dir(tmpDownloadedFilePath), destination)
		if err != nil {
			r.logger.Err(err).Str("tmpDownloadedFilePath", tmpDownloadedFilePath).Str("destination", destination).Msg("debrid: Failed to move downloaded file")
			r.wsEventManager.SendEvent(events.ErrorToast, fmt.Sprintf("debrid: Failed to move downloaded file: %v", err))
			r.sendDownloadCancelledEvent(tId, downloadUrl, downloadMap)
			return false
		}
		return true
	}
	if err != nil {
		r.logger.Err(err).Str("tmpDownloadedFilePath", tmpDownloadedFilePath).Msg("debrid: Failed to extract downloaded file")
		r.wsEventManager.SendEvent(events.ErrorToast, fmt.Sprintf("debrid: Failed to extract downloaded file: %v", err))
		r.sendDownloadCancelledEvent(tId, downloadUrl, downloadMap)
		return false
	}

	r.logger.Debug().Msg("debrid: Extraction completed, deleting temporary files")

	// Delete the downloaded file (/path/to/destination/.tmp-123456789/downlooaded_torrent.zip)
	err = os.Remove(tmpDownloadedFilePath)
	if err != nil {
		r.logger.Err(err).Str("tmpDownloadedFilePath", tmpDownloadedFilePath).Msg("debrid: Failed to delete downloaded file")
		// Do not stop here, continue with the extracted files
	}

	r.logger.Debug().Str("extractedDir", extractedDir).Str("destination", destination).Msg("debrid: Moving extracted files to destination")

	// Move the extracted files to the destination
	// /path/to/destination/.tmp-123456789/extracted-1234/{files} -> /path/to/destination/{files}
	err = moveContentsTo(extractedDir, destination)
	if err != nil {
		r.logger.Err(err).Str("extractedDir", extractedDir).Str("destination", destination).Msg("debrid: Failed to move downloaded files")
		r.wsEventManager.SendEvent(events.ErrorToast, fmt.Sprintf("debrid: Failed to move downloaded files: %v", err))
		r.sendDownloadCancelledEvent(tId, downloadUrl, downloadMap)
		return false
	}

	return true
}

type resumableDownloadResult struct {
	Path     string
	Filename string
	Ext      string
}

type downloadContentRange struct {
	Start int64
	End   int64
	Size  int64
}

func (r *Repository) downloadHTTPFile(ctx context.Context, tId string, downloadUrl string, tmpDirPath string, downloadMap *result.Map[string, downloadStatus]) (*resumableDownloadResult, error) {
	client := &http.Client{}
	backoff := debridDownloadInitialBackoff
	written := int64(0)
	expectedSize := int64(-1)
	lastBytes := int64(0)
	lastSent := time.Now()

	result := &resumableDownloadResult{
		Filename: "downloaded_torrent",
	}

	var lastErr error
	for attempt := 1; attempt <= debridDownloadMaxAttempts; attempt++ {
		rangeStart := written
		resp, err := r.openDownloadResponse(ctx, client, downloadUrl, rangeStart)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil, err
			}
			lastErr = err
			r.logger.Warn().Err(err).Int("attempt", attempt).Str("downloadUrl", downloadUrl).Msg("debrid: Download request failed, retrying")
			if err := waitBeforeDebridDownloadRetry(ctx, attempt, backoff); err != nil {
				return nil, err
			}
			backoff = nextDebridDownloadBackoff(backoff)
			continue
		}

		appendFile, responseExpectedSize, err := validateDebridDownloadResponse(resp, rangeStart)
		if err != nil {
			_ = resp.Body.Close()
			if errors.Is(err, errInvalidDownloadContentRange) {
				return nil, err
			}
			if errors.Is(err, errInvalidDownloadStatus) && !isRetryableDebridDownloadStatus(resp.StatusCode) {
				return nil, err
			}
			lastErr = err
			r.logger.Warn().Err(err).Int("attempt", attempt).Int("status", resp.StatusCode).Str("downloadUrl", downloadUrl).Msg("debrid: Download response status failed, retrying")
			if err := waitBeforeDebridDownloadRetry(ctx, attempt, backoff); err != nil {
				return nil, err
			}
			backoff = nextDebridDownloadBackoff(backoff)
			continue
		}

		if responseExpectedSize >= 0 {
			expectedSize = responseExpectedSize
		}

		if result.Path == "" {
			result.Filename, result.Ext = getDownloadFilename(downloadUrl, resp.Header)
			result.Path = filepath.Join(tmpDirPath, result.Filename)
		}

		if !appendFile {
			written = 0
			lastBytes = 0
			rangeStart = 0
		}

		fileFlag := os.O_WRONLY | os.O_CREATE
		if appendFile {
			fileFlag |= os.O_APPEND
		} else {
			fileFlag |= os.O_TRUNC
		}

		file, err := os.OpenFile(result.Path, fileFlag, 0644)
		if err != nil {
			_ = resp.Body.Close()
			return nil, fmt.Errorf("failed to create temp file: %w", err)
		}

		err = r.copyDownloadResponseBody(ctx, resp.Body, file, tId, downloadUrl, downloadMap, &written, &lastBytes, expectedSize, &lastSent)
		bodyCloseErr := resp.Body.Close()
		fileCloseErr := file.Close()
		if bodyCloseErr != nil && err == nil {
			err = bodyCloseErr
		}
		if fileCloseErr != nil {
			return nil, fmt.Errorf("failed to close temp file: %w", fileCloseErr)
		}

		if err == nil && expectedSize >= 0 && written != expectedSize {
			err = fmt.Errorf("%w: downloaded %d of %d bytes", io.ErrUnexpectedEOF, written, expectedSize)
		}

		if err == nil {
			return result, nil
		}

		if errors.Is(err, context.Canceled) {
			return nil, err
		}

		lastErr = err
		r.logger.Warn().Err(err).Int("attempt", attempt).Int64("written", written).Int64("expectedSize", expectedSize).Str("downloadUrl", downloadUrl).Msg("debrid: Download attempt failed, retrying")
		if err := waitBeforeDebridDownloadRetry(ctx, attempt, backoff); err != nil {
			return nil, err
		}
		backoff = nextDebridDownloadBackoff(backoff)
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("download failed")
	}
	return nil, fmt.Errorf("download failed after %d attempts: %w", debridDownloadMaxAttempts, lastErr)
}

func (r *Repository) openDownloadResponse(ctx context.Context, client *http.Client, downloadUrl string, rangeStart int64) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	if rangeStart > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", rangeStart))
	}
	return client.Do(req)
}

func isRetryableDebridDownloadStatus(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func validateDebridDownloadResponse(resp *http.Response, rangeStart int64) (appendFile bool, expectedSize int64, err error) {
	expectedSize = -1

	switch resp.StatusCode {
	case http.StatusOK:
		if resp.ContentLength >= 0 {
			expectedSize = resp.ContentLength
		}
		return false, expectedSize, nil
	case http.StatusPartialContent:
		contentRange, err := parseDownloadContentRange(resp.Header.Get("Content-Range"))
		if err != nil {
			return false, -1, fmt.Errorf("%w: %v", errInvalidDownloadContentRange, err)
		}
		if contentRange.Start != rangeStart {
			return false, -1, fmt.Errorf("%w: expected start %d, got %d", errInvalidDownloadContentRange, rangeStart, contentRange.Start)
		}
		if contentRange.Size >= 0 {
			expectedSize = contentRange.Size
		}
		return rangeStart > 0, expectedSize, nil
	default:
		return false, -1, fmt.Errorf("%w: %s", errInvalidDownloadStatus, resp.Status)
	}
}

func parseDownloadContentRange(value string) (downloadContentRange, error) {
	matches := regexp.MustCompile(`^bytes (\d+)-(\d+)/(\d+|\*)$`).FindStringSubmatch(value)
	if len(matches) != 4 {
		return downloadContentRange{}, fmt.Errorf("malformed Content-Range %q", value)
	}

	start, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return downloadContentRange{}, err
	}
	end, err := strconv.ParseInt(matches[2], 10, 64)
	if err != nil {
		return downloadContentRange{}, err
	}
	if end < start {
		return downloadContentRange{}, fmt.Errorf("end before start")
	}

	size := int64(-1)
	if matches[3] != "*" {
		size, err = strconv.ParseInt(matches[3], 10, 64)
		if err != nil {
			return downloadContentRange{}, err
		}
		if size <= end {
			return downloadContentRange{}, fmt.Errorf("size before end")
		}
	}

	return downloadContentRange{Start: start, End: end, Size: size}, nil
}

func (r *Repository) copyDownloadResponseBody(
	ctx context.Context,
	body io.Reader,
	file *os.File,
	tId string,
	downloadUrl string,
	downloadMap *result.Map[string, downloadStatus],
	written *int64,
	lastBytes *int64,
	expectedSize int64,
	lastSent *time.Time,
) error {
	buffer := make([]byte, 32*1024)
	for {
		n, err := body.Read(buffer)
		if n > 0 {
			if _, writeErr := file.Write(buffer[:n]); writeErr != nil {
				return fmt.Errorf("failed to write to temp file: %w", writeErr)
			}
			*written += int64(n)
			r.updateDownloadProgress(tId, downloadUrl, downloadMap, *written, expectedSize, lastBytes, lastSent)
		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			if errors.Is(ctx.Err(), context.Canceled) {
				return context.Canceled
			}
			return err
		}
		if errors.Is(ctx.Err(), context.Canceled) {
			return context.Canceled
		}
	}
}

func (r *Repository) updateDownloadProgress(tId string, downloadUrl string, downloadMap *result.Map[string, downloadStatus], totalBytes int64, totalSize int64, lastBytes *int64, lastSent *time.Time) {
	speed := 0
	if totalSize > 0 {
		speed = int((totalBytes - *lastBytes) / 1024) // KB/s
		*lastBytes = totalBytes
	}

	progressSize := totalSize
	if progressSize < 0 {
		progressSize = 0
	}

	downloadMap.Set(downloadUrl, downloadStatus{
		TotalBytes: totalBytes,
		TotalSize:  progressSize,
	})

	if time.Since(*lastSent) <= time.Second*2 {
		return
	}

	_totalBytes := uint64(0)
	_totalSize := uint64(0)
	downloadMap.Range(func(key string, value downloadStatus) bool {
		_totalBytes += uint64(value.TotalBytes)
		_totalSize += uint64(value.TotalSize)
		return true
	})
	// Notify progress
	r.wsEventManager.SendEvent(events.DebridDownloadProgress, map[string]interface{}{
		"status":     "downloading",
		"itemID":     tId,
		"totalBytes": util.Bytes(_totalBytes),
		"totalSize":  util.Bytes(_totalSize),
		"speed":      speed,
	})
	*lastSent = time.Now()
}

func waitBeforeDebridDownloadRetry(ctx context.Context, attempt int, backoff time.Duration) error {
	if attempt >= debridDownloadMaxAttempts {
		return nil
	}
	timer := time.NewTimer(backoff)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return context.Canceled
	case <-timer.C:
		return nil
	}
}

func nextDebridDownloadBackoff(backoff time.Duration) time.Duration {
	backoff *= 2
	if backoff > debridDownloadMaxBackoff {
		return debridDownloadMaxBackoff
	}
	return backoff
}

func getDownloadFilename(downloadUrl string, headers http.Header) (filename string, ext string) {
	filename = "downloaded_torrent"

	if hFilename, ok := getFilenameFromContentDisposition(headers); ok {
		filename = hFilename
		ext = filepath.Ext(filename)
	}

	// The case for TorBox(?)
	// RD will return application/force-download so ext will still be empty.
	if ct := headers.Get("Content-Type"); ct != "" && ext == "" {
		mediaType, _, err := mime.ParseMediaType(ct)
		if err == nil {
			switch mediaType {
			case "application/zip":
				ext = ".zip"
			case "application/x-rar-compressed":
				ext = ".rar"
			default:
			}
		}
	}

	// Check if the download URL has the extension.
	// This works for RD, by that point we should have "Torrent Name.zip" or "Episode.mkv".
	if filename == "downloaded_torrent" && ext == "" {
		urlExt := filepath.Ext(downloadUrl)
		if urlExt != "" {
			filename = filepath.Base(downloadUrl)
			filename, _ = url.PathUnescape(filename)
			ext = urlExt
		}
	}

	// Add the file extension to downloaded_torrent if we couldn't guess the name from headers or URL.
	if filename == "downloaded_torrent" && ext != "" {
		filename = fmt.Sprintf("%s%s", filename, ext)
	}

	return filepath.Base(filename), ext
}

func getFilenameFromContentDisposition(headers http.Header) (string, bool) {
	contentDisposition := headers.Get("Content-Disposition")
	if contentDisposition == "" {
		return "", false
	}

	_, params, err := mime.ParseMediaType(contentDisposition)
	if err != nil {
		return "", false
	}

	filename := params["filename"]
	if filename == "" {
		return "", false
	}
	filename, _ = url.PathUnescape(filename)
	return filepath.Base(filename), true
}

func (r *Repository) sendDownloadCancelledEvent(tId string, url string, downloadMap *result.Map[string, downloadStatus]) {
	downloadMap.Delete(url)

	if len(downloadMap.Values()) == 0 {
		r.wsEventManager.SendEvent(events.DebridDownloadProgress, map[string]interface{}{
			"status": "cancelled",
			"itemID": tId,
		})
	}
}

func (r *Repository) sendDownloadStartedEvent(tId string, torrentName string, destination string, downloadURL string) error {
	event := &DebridLocalDownloadStartedEvent{
		TorrentItemID: tId,
		TorrentName:   torrentName,
		Destination:   destination,
		DownloadUrl:   downloadURL,
	}

	if err := hook.GlobalHookManager.OnDebridLocalDownloadStarted().Trigger(event); err != nil {
		return err
	}

	r.wsEventManager.SendEvent(events.DebridDownloadProgress, map[string]interface{}{
		"status":     "downloading",
		"itemID":     tId,
		"totalBytes": "0 B",
		"totalSize":  "-",
		"speed":      "",
	})

	return nil
}

func (r *Repository) sendDownloadCompletedEvent(tId string, torrentName string, destination string) {
	r.wsEventManager.SendEvent(events.DebridDownloadProgress, map[string]interface{}{
		"status": "completed",
		"itemID": tId,
	})

	event := &DebridLocalDownloadCompletedEvent{
		TorrentItemID: tId,
		TorrentName:   torrentName,
		Destination:   destination,
	}
	if err := hook.GlobalHookManager.OnDebridLocalDownloadCompleted().Trigger(event); err != nil {
		r.logger.Err(err).Str("torrentItemId", tId).Msg("debrid: Failed to trigger local download completed hook")
	}
}
