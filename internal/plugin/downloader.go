package plugin

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"seanime/internal/extension"
	goja_util "seanime/internal/util/goja"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type DownloadStatus string

const (
	DownloadStatusDownloading DownloadStatus = "downloading"
	DownloadStatusCompleted   DownloadStatus = "completed"
	DownloadStatusCancelled   DownloadStatus = "cancelled"
	DownloadStatusError       DownloadStatus = "error"
)

type DownloadProgress struct {
	ID             string    `json:"id"`
	URL            string    `json:"url"`
	Destination    string    `json:"destination"`
	TotalBytes     int64     `json:"totalBytes"`
	TotalSize      int64     `json:"totalSize"`
	Speed          int64     `json:"speed"`
	Percentage     float64   `json:"percentage"`
	Status         string    `json:"status"`
	Error          string    `json:"error,omitempty"`
	LastUpdateTime time.Time `json:"lastUpdate"`
	StartTime      time.Time `json:"startTime"`

	lastBytes int64 `json:"-"`
}

// IsFinished returns true if the download has completed, errored, or been cancelled
func (p *DownloadProgress) IsFinished() bool {
	return p.Status == string(DownloadStatusCompleted) || p.Status == string(DownloadStatusCancelled) || p.Status == string(DownloadStatusError)
}

type progressSubscriber struct {
	ID       string
	Channel  chan map[string]interface{}
	Cancel   context.CancelFunc
	LastSent time.Time
}

func (a *AppContextImpl) BindDownloaderToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *goja_util.Scheduler) {
	downloadObj := vm.NewObject()

	progressMap := sync.Map{}
	downloadCancels := sync.Map{}
	progressSubscribers := sync.Map{}

	_ = downloadObj.Set("watch", func(downloadID string, callback goja.Callable) goja.Value {
		// Create cancellable context for the subscriber
		ctx, cancel := context.WithCancel(context.Background())

		// Create a new subscriber
		subscriber := &progressSubscriber{
			ID:       downloadID,
			Channel:  make(chan map[string]interface{}, 1),
			Cancel:   cancel,
			LastSent: time.Now(),
		}

		// Store the subscriber
		if existing, ok := progressSubscribers.Load(downloadID); ok {
			// Cancel existing subscriber if any
			existing.(*progressSubscriber).Cancel()
		}
		progressSubscribers.Store(downloadID, subscriber)

		// Start watching for progress updates
		go func() {
			defer func() {
				close(subscriber.Channel)
				progressSubscribers.Delete(downloadID)
			}()

			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					// If download is complete/cancelled/errored, send one last update and stop
					if progress, ok := progressMap.Load(downloadID); ok {
						p := progress.(*DownloadProgress)
						scheduler.ScheduleAsync(func() error {
							p.Speed = 0
							callback(goja.Undefined(), vm.ToValue(p))
							return nil
						})
					}
					return
				case <-ticker.C:
					if progress, ok := progressMap.Load(downloadID); ok {
						p := progress.(*DownloadProgress)
						scheduler.ScheduleAsync(func() error {
							callback(goja.Undefined(), vm.ToValue(p))
							return nil
						})
						// If download is complete/cancelled/errored, send one last update and stop
						if p.IsFinished() {
							return
						}
					} else {
						// Download not found or already completed
						return
					}
				}
			}
		}()

		// Return a function to cancel the watch
		return vm.ToValue(func() {
			if subscriber, ok := progressSubscribers.Load(downloadID); ok {
				subscriber.(*progressSubscriber).Cancel()
			}
		})
	})

	_ = downloadObj.Set("download", func(url string, destination string, options map[string]interface{}) (string, error) {
		if !a.isAllowedPath(ext, destination, AllowPathWrite) {
			return "", ErrPathNotAuthorized
		}

		// Generate unique download ID
		downloadID := uuid.New().String()

		// Create context with optional timeout
		var ctx context.Context
		var cancel context.CancelFunc
		if timeout, ok := options["timeout"].(float64); ok {
			ctx, cancel = context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
		} else {
			ctx, cancel = context.WithCancel(context.Background())
		}
		downloadCancels.Store(downloadID, cancel)

		logger.Trace().Str("url", url).Str("destination", destination).Msg("plugin: Starting download")

		// Initialize progress tracking
		now := time.Now()
		progress := &DownloadProgress{
			ID:             downloadID,
			URL:            url,
			Destination:    destination,
			Status:         string(DownloadStatusDownloading),
			LastUpdateTime: now,
			StartTime:      now,
		}
		progressMap.Store(downloadID, progress)

		// Start download in a goroutine
		go func() {
			defer downloadCancels.Delete(downloadID)
			defer func() {
				// Clean up subscriber if it exists
				if subscriber, ok := progressSubscribers.Load(downloadID); ok {
					subscriber.(*progressSubscriber).Cancel()
				}
			}()

			// Create request
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				progress.Status = string(DownloadStatusError)
				progress.Error = err.Error()
				return
			}

			// Add headers if provided
			if headers, ok := options["headers"].(map[string]interface{}); ok {
				for k, v := range headers {
					if strVal, ok := v.(string); ok {
						req.Header.Set(k, strVal)
					}
				}
			}

			// Execute request
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				progress.Status = string(DownloadStatusError)
				progress.Error = err.Error()
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode < 200 || resp.StatusCode > 299 {
				progress.Status = string(DownloadStatusError)
				progress.Error = fmt.Sprintf("server returned status code %d", resp.StatusCode)
				return
			}

			// Update progress with content length
			progress.TotalSize = resp.ContentLength

			// Create destination directory if it doesn't exist
			if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
				progress.Status = string(DownloadStatusError)
				progress.Error = err.Error()
				return
			}

			// Create destination file
			file, err := os.Create(destination)
			if err != nil {
				progress.Status = string(DownloadStatusError)
				progress.Error = err.Error()
				return
			}
			defer file.Close()

			// Create buffer for copying
			buffer := make([]byte, 32*1024)
			lastUpdateTime := now

			logger.Trace().Str("url", url).Str("destination", destination).Msg("plugin: Download started")

			for {
				select {
				case <-ctx.Done():
					progress.Status = string(DownloadStatusCancelled)
					logger.Trace().Str("url", url).Str("destination", destination).Msg("plugin: Download cancelled")
					return
				default:
					n, err := resp.Body.Read(buffer)
					if n > 0 {
						_, writeErr := file.Write(buffer[:n])
						if writeErr != nil {
							progress.Status = string(DownloadStatusError)
							progress.Error = writeErr.Error()
							return
						}

						progress.TotalBytes += int64(n)
						if progress.TotalSize > 0 {
							progress.Percentage = float64(progress.TotalBytes) / float64(progress.TotalSize) * 100
						}

						// Update speed every 500ms
						if time.Since(lastUpdateTime) > 500*time.Millisecond {
							elapsed := time.Since(lastUpdateTime).Seconds()
							bytesInPeriod := progress.TotalBytes - progress.lastBytes
							progress.Speed = int64(float64(bytesInPeriod) / elapsed)
							progress.lastBytes = progress.TotalBytes
							progress.LastUpdateTime = time.Now()
							lastUpdateTime = time.Now()
						}
					}

					if err != nil {
						if err == io.EOF {
							progress.Status = string(DownloadStatusCompleted)
							logger.Trace().Str("url", url).Str("destination", destination).Msg("plugin: Download completed")
							return
						}
						if errors.Is(err, context.Canceled) {
							progress.Status = string(DownloadStatusCancelled)
							logger.Trace().Str("url", url).Str("destination", destination).Msg("plugin: Download cancelled")
							return
						}
						progress.Status = string(DownloadStatusError)
						progress.Error = err.Error()
						logger.Error().Err(err).Str("url", url).Str("destination", destination).Msg("plugin: Download error")
						return
					}
				}
			}
		}()

		return downloadID, nil
	})

	_ = downloadObj.Set("getProgress", func(downloadID string) *DownloadProgress {
		if progress, ok := progressMap.Load(downloadID); ok {
			return progress.(*DownloadProgress)
		}
		return nil
	})

	_ = downloadObj.Set("listDownloads", func() []*DownloadProgress {
		downloads := make([]*DownloadProgress, 0)
		progressMap.Range(func(key, value interface{}) bool {
			downloads = append(downloads, value.(*DownloadProgress))
			return true
		})
		return downloads
	})

	_ = downloadObj.Set("cancel", func(downloadID string) {
		if cancel, ok := downloadCancels.Load(downloadID); ok {
			if cancel == nil {
				return
			}
			logger.Trace().Str("downloadID", downloadID).Msg("plugin: Cancelling download")
			cancel.(context.CancelFunc)()
		}
	})

	_ = downloadObj.Set("cancelAll", func() {
		logger.Trace().Msg("plugin: Cancelling all downloads")
		downloadCancels.Range(func(key, value interface{}) bool {
			if value == nil {
				return true
			}
			logger.Trace().Str("downloadID", key.(string)).Msg("plugin: Cancelling download")
			value.(context.CancelFunc)()
			return true
		})
	})

	_ = obj.Set("downloader", downloadObj)
}
