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
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
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

	go func(ctx context.Context) {
		defer func() {
			cancel()
			r.ctxMap.Delete(tId)
		}()

		wg := sync.WaitGroup{}
		downloadUrls := strings.Split(downloadUrl, ",")
		downloadMap := result.NewResultMap[string, downloadStatus]()

		for _, url := range downloadUrls {
			wg.Add(1)
			go func(ctx context.Context, url string) {
				defer wg.Done()

				// Download the file
				ok := r.downloadFile(ctx, tId, url, destination, downloadMap)
				if !ok {
					return
				}
			}(ctx, url)
		}
		wg.Wait()

		r.sendDownloadCompletedEvent(tId)
		notifier.GlobalNotifier.Notify(notifier.Debrid, fmt.Sprintf("Downloaded %q", torrentName))
	}(ctx)

	// Send a starting event
	r.wsEventManager.SendEvent(events.DebridDownloadProgress, map[string]interface{}{
		"status":     "downloading",
		"itemID":     tId,
		"totalBytes": "0 B",
		"totalSize":  "-",
		"speed":      "",
	})

	return nil
}

func (r *Repository) downloadFile(ctx context.Context, tId string, downloadUrl string, destination string, downloadMap *result.Map[string, downloadStatus]) (ok bool) {
	defer util.HandlePanicInModuleThen("debrid/client/downloadFile", func() {
		ok = false
	})

	// Create a cancellable HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadUrl, nil)
	if err != nil {
		r.logger.Err(err).Str("downloadUrl", downloadUrl).Msg("debrid: Failed to create request")
		return false
	}

	_ = os.MkdirAll(destination, os.ModePerm)

	// Download the files to a temporary folder
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

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		r.logger.Err(err).Str("downloadUrl", downloadUrl).Msg("debrid: Failed to execute request")
		r.wsEventManager.SendEvent(events.ErrorToast, fmt.Sprintf("debrid: Failed to execute download request: %v", err))
		return false
	}
	defer resp.Body.Close()

	// e.g. "my-torrent.zip", "downloaded_torrent"
	filename := "downloaded_torrent"
	ext := ""

	// Try to get the file name from the Content-Disposition header
	hFilename, err := getFilenameFromHeaders(downloadUrl)
	if err == nil {
		r.logger.Warn().Str("newFilename", hFilename).Str("defaultFilename", filename).Msg("debrid: Filename found in headers, overriding default")
		filename = hFilename
	}

	if ct := resp.Header.Get("Content-Type"); ct != "" {
		mediaType, _, err := mime.ParseMediaType(ct)
		if err == nil {
			switch mediaType {
			case "application/zip":
				ext = ".zip"
			case "application/x-rar-compressed":
				ext = ".rar"
			default:
			}
			r.logger.Debug().Str("mediaType", mediaType).Str("ext", ext).Msg("debrid: Detected media type and extension")
		}
	}

	if filename == "downloaded_torrent" && ext != "" {
		filename = fmt.Sprintf("%s%s", filename, ext)
	}

	// Check if the download URL has the extension
	urlExt := filepath.Ext(downloadUrl)
	if filename == "downloaded_torrent" && urlExt != "" {
		filename = filepath.Base(downloadUrl)
		filename, _ = url.PathUnescape(filename)
		ext = urlExt
		r.logger.Warn().Str("urlExt", urlExt).Str("filename", filename).Str("downloadUrl", downloadUrl).Msg("debrid: Extension found in URL, using it as file extension and file name")
	}

	r.logger.Debug().Str("filename", filename).Str("ext", ext).Msg("debrid: Starting download")

	// Create a file in the temporary folder to store the download
	// e.g. "/tmp/torrent-123456789/my-torrent.zip"
	tmpDownloadedFilePath := filepath.Join(tmpDirPath, filename)
	file, err := os.Create(tmpDownloadedFilePath)
	if err != nil {
		r.logger.Err(err).Str("tmpDownloadedFilePath", tmpDownloadedFilePath).Msg("debrid: Failed to create temp file")
		r.wsEventManager.SendEvent(events.ErrorToast, fmt.Sprintf("debrid: Failed to create temp file: %v", err))
		return false
	}

	totalSize := resp.ContentLength
	speed := 0

	lastSent := time.Now()

	// Copy response body to the temporary file
	buffer := make([]byte, 32*1024)
	var totalBytes int64
	var lastBytes int64
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			_, writeErr := file.Write(buffer[:n])
			if writeErr != nil {
				_ = file.Close()
				r.logger.Err(writeErr).Str("tmpDownloadedFilePath", tmpDownloadedFilePath).Msg("debrid: Failed to write to temp file")
				r.wsEventManager.SendEvent(events.ErrorToast, fmt.Sprintf("debrid: Download failed / Failed to write to temp file: %v", writeErr))
				r.sendDownloadCancelledEvent(tId, downloadUrl, downloadMap)
				return false
			}
			totalBytes += int64(n)
			if totalSize > 0 {
				speed = int((totalBytes - lastBytes) / 1024) // KB/s
				lastBytes = totalBytes
			}

			downloadMap.Set(downloadUrl, downloadStatus{
				TotalBytes: totalBytes,
				TotalSize:  totalSize,
			})

			if time.Since(lastSent) > time.Second*2 {
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
					"totalBytes": humanize.Bytes(_totalBytes),
					"totalSize":  humanize.Bytes(_totalSize),
					"speed":      speed,
				})
				lastSent = time.Now()
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			if errors.Is(err, context.Canceled) {
				_ = file.Close()
				r.logger.Debug().Msg("debrid: Download cancelled")
				r.sendDownloadCancelledEvent(tId, downloadUrl, downloadMap)
				return false
			}
			_ = file.Close()
			r.logger.Err(err).Str("downloadUrl", downloadUrl).Msg("debrid: Failed to read from response body")
			r.wsEventManager.SendEvent(events.ErrorToast, fmt.Sprintf("debrid: Download failed / Failed to read from response body: %v", err))
			r.sendDownloadCancelledEvent(tId, downloadUrl, downloadMap)
			return false
		}
	}

	_ = file.Close()

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
	switch ext {
	case ".zip":
		extractedDir, err = unzipFile(tmpDownloadedFilePath, tmpDirPath)
		r.logger.Debug().Str("extractedDir", extractedDir).Msg("debrid: Extracted zip file")
	case ".rar":
		extractedDir, err = unrarFile(tmpDownloadedFilePath, tmpDirPath)
		r.logger.Debug().Str("extractedDir", extractedDir).Msg("debrid: Extracted rar file")
	default:
		r.logger.Debug().Str("tmpDownloadedFilePath", tmpDownloadedFilePath).Str("destination", destination).Msg("debrid: No extraction needed, moving file directly")
		// Move the file directly to the destination
		err = moveFolderOrFileTo(tmpDownloadedFilePath, destination)
		if err != nil {
			r.logger.Err(err).Str("tmpDownloadedFilePath", tmpDownloadedFilePath).Str("destination", destination).Msg("debrid: Failed to move downloaded file")
			r.wsEventManager.SendEvent(events.ErrorToast, fmt.Sprintf("debrid: Failed to move downloaded file: %v", err))
			r.sendDownloadCancelledEvent(tId, downloadUrl, downloadMap)
			return false
		}
		return false
	}
	if err != nil {
		r.logger.Err(err).Str("tmpDownloadedFilePath", tmpDownloadedFilePath).Msg("debrid: Failed to extract downloaded file")
		r.wsEventManager.SendEvent(events.ErrorToast, fmt.Sprintf("debrid: Failed to extract downloaded file: %v", err))
		r.sendDownloadCancelledEvent(tId, downloadUrl, downloadMap)
		return false
	}

	r.logger.Debug().Msg("debrid: Extraction completed, deleting temporary files")

	// Delete the downloaded file
	err = os.Remove(tmpDownloadedFilePath)
	if err != nil {
		r.logger.Err(err).Str("tmpDownloadedFilePath", tmpDownloadedFilePath).Msg("debrid: Failed to delete downloaded file")
		// Do not stop here, continue with the extracted files
	}

	r.logger.Debug().Str("extractedDir", extractedDir).Str("destination", destination).Msg("debrid: Moving extracted files to destination")

	// Move the extracted files to the destination
	err = moveContentsTo(extractedDir, destination)
	if err != nil {
		r.logger.Err(err).Str("extractedDir", extractedDir).Str("destination", destination).Msg("debrid: Failed to move downloaded files")
		r.wsEventManager.SendEvent(events.ErrorToast, fmt.Sprintf("debrid: Failed to move downloaded files: %v", err))
		r.sendDownloadCancelledEvent(tId, downloadUrl, downloadMap)
		return false
	}

	return true
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

func (r *Repository) sendDownloadCompletedEvent(tId string) {
	r.wsEventManager.SendEvent(events.DebridDownloadProgress, map[string]interface{}{
		"status": "completed",
		"itemID": tId,
	})
}

func getFilenameFromHeaders(url string) (string, error) {
	resp, err := http.Head(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Get the Content-Disposition header
	contentDisposition := resp.Header.Get("Content-Disposition")
	if contentDisposition == "" {
		return "", fmt.Errorf("no Content-Disposition header found")
	}

	// Use a regex to extract the filename from Content-Disposition
	re := regexp.MustCompile(`filename="(.+)"`)
	matches := re.FindStringSubmatch(contentDisposition)
	if len(matches) > 1 {
		return matches[1], nil
	}
	return "", fmt.Errorf("filename not found in Content-Disposition header")
}
