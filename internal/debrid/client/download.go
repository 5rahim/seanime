package debrid_client

import (
	"context"
	"errors"
	"fmt"
	"github.com/dustin/go-humanize"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"seanime/internal/debrid/debrid"
	"seanime/internal/events"
	"seanime/internal/notifier"
	"time"
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

func (r *Repository) downloadTorrentItem(tId string, torrentName string, destination string) error {
	provider, err := r.GetProvider()
	if err != nil {
		return err
	}

	// Get the download URL
	downloadUrl, err := provider.GetTorrentDownloadUrl(debrid.DownloadTorrentOptions{
		ID: tId,
	})
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	r.ctxMap.Set(tId, cancel)

	go func(ctx context.Context) {
		defer func() {
			cancel()
			r.ctxMap.Delete(tId)
		}()

		// Create a cancellable HTTP request
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadUrl, nil)
		if err != nil {
			r.logger.Err(err).Msg("debrid: Failed to create request")
			return
		}

		// Download the files to a temporary folder
		tmpDirPath, err := os.MkdirTemp("", "torrent-")
		if err != nil {
			r.logger.Err(err).Msg("debrid: Failed to create temp folder")
			r.wsEventManager.SendEvent(events.ErrorToast, fmt.Sprintf("debrid: Failed to create temp folder: %v", err))
			return
		}
		defer os.RemoveAll(tmpDirPath) // Clean up temp folder on exit

		// Execute the request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			r.logger.Err(err).Msg("debrid: Failed to execute request")
			r.wsEventManager.SendEvent(events.ErrorToast, fmt.Sprintf("debrid: Failed to execute download request: %v", err))
			return
		}
		defer resp.Body.Close()

		// e.g. "my-torrent.zip", "downloaded_torrent"
		filename := "downloaded_torrent"
		ext := ""

		// Try to get the file name from the Content-Disposition header
		hFilename, err := getFilenameFromHeaders(downloadUrl)
		if err == nil {
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
			}
		}

		if filename == "downloaded_torrent" && ext != "" {
			filename = fmt.Sprintf("%s%s", filename, ext)
		}

		r.logger.Debug().Str("filename", filename).Str("ext", ext).Msg("debrid: Downloading torrent")

		// Create a file in the temporary folder to store the download
		// e.g. "/tmp/torrent-123456789/my-torrent.zip"
		tmpDownloadedFilePath := filepath.Join(tmpDirPath, filename)
		file, err := os.Create(tmpDownloadedFilePath)
		if err != nil {
			r.logger.Err(err).Msg("debrid: Failed to create temp file")
			r.wsEventManager.SendEvent(events.ErrorToast, fmt.Sprintf("debrid: Failed to create temp file: %v", err))
			return
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
					r.logger.Err(writeErr).Msg("debrid: Failed to write to temp file")
					r.wsEventManager.SendEvent(events.ErrorToast, fmt.Sprintf("debrid: Download failed / Failed to write to temp file: %v", writeErr))
					r.sendDownloadCancelledEvent(tId)
					return
				}
				totalBytes += int64(n)
				if totalSize > 0 {
					speed = int((totalBytes - lastBytes) / 1024) // KB/s
					lastBytes = totalBytes
				}

				if time.Since(lastSent) > time.Second*2 {
					// Notify progress
					r.wsEventManager.SendEvent(events.DebridDownloadProgress, map[string]interface{}{
						"status":     "downloading",
						"itemID":     tId,
						"totalBytes": humanize.Bytes(uint64(totalBytes)),
						"totalSize":  humanize.Bytes(uint64(totalSize)),
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
					r.sendDownloadCancelledEvent(tId)
					return
				}
				_ = file.Close()
				r.logger.Err(err).Msg("debrid: Failed to read from response body")
				r.wsEventManager.SendEvent(events.ErrorToast, fmt.Sprintf("debrid: Download failed / Failed to read from response body: %v", err))
				r.sendDownloadCancelledEvent(tId)
				return
			}
		}

		_ = file.Close()

		r.wsEventManager.SendEvent(events.DebridDownloadProgress, map[string]interface{}{
			"status":     "downloading",
			"itemID":     tId,
			"totalBytes": "Extracting...",
			"totalSize":  humanize.Bytes(uint64(totalSize)),
			"speed":      "",
		})

		switch runtime.GOOS {
		case "windows":
			time.Sleep(time.Second * 1)
		}

		// Extract the downloaded file
		var extractedDir string
		switch ext {
		case ".zip":
			extractedDir, err = unzipFile(tmpDownloadedFilePath, tmpDirPath)
		case ".rar":
			extractedDir, err = unrarFile(tmpDownloadedFilePath, tmpDirPath)
		default:
			// Move the file directly to the destination
			err = moveFolderOrFileTo(tmpDownloadedFilePath, destination)
			if err != nil {
				r.logger.Err(err).Msg("debrid: Failed to move downloaded file")
				r.wsEventManager.SendEvent(events.ErrorToast, fmt.Sprintf("debrid: Failed to move downloaded file: %v", err))
				r.sendDownloadCancelledEvent(tId)
				return
			}
			return // Exit early
		}
		if err != nil {
			r.logger.Err(err).Msg("debrid: Failed to extract downloaded file")
			r.wsEventManager.SendEvent(events.ErrorToast, fmt.Sprintf("debrid: Failed to extract downloaded file: %v", err))
			r.sendDownloadCancelledEvent(tId)
			return
		}

		// Delete the downloaded file
		err = os.Remove(tmpDownloadedFilePath)
		if err != nil {
			r.logger.Err(err).Msg("debrid: Failed to delete downloaded file")
			// Do not stop here, continue with the extracted files
		}

		// Move the extracted files to the destination
		err = moveContentsTo(extractedDir, destination)
		if err != nil {
			r.logger.Err(err).Msg("debrid: Failed to move downloaded files")
			r.wsEventManager.SendEvent(events.ErrorToast, fmt.Sprintf("debrid: Failed to move downloaded files: %v", err))
			r.sendDownloadCancelledEvent(tId)
			return
		}

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

func (r *Repository) sendDownloadCancelledEvent(tId string) {
	r.wsEventManager.SendEvent(events.DebridDownloadProgress, map[string]interface{}{
		"status": "cancelled",
		"itemID": tId,
	})
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
