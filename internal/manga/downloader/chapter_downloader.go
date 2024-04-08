package chapter_downloader

import (
	"bytes"
	"fmt"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/database/db"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/manga/providers"
	"github.com/seanime-app/seanime/internal/util"
	_ "golang.org/x/image/webp"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

var (
	ErrChapterAlreadyDownloaded = fmt.Errorf("chapter already downloaded")
)

// 📁 cache/manga
// └── 📁 {provider}_{mediaId}_{chapterId}      <- Downloader generates
//     ├── 📄 registry.json						<- Contains ChapterDownloadRegistry
//     ├── 📄 1.jpg
//     ├── 📄 2.jpg
//     └── 📄 ...
//

type (
	// Downloader is used to download chapters from various manga providers.
	Downloader struct {
		logger         *zerolog.Logger
		wsEventManager events.IWSEventManager
		database       *db.Database
		downloadDir    string
		mu             sync.Mutex
		downloadMu     sync.Mutex
		// cancelChannel is used to cancel some or all downloads.
		cancelChannels map[DownloadID]chan struct{}
		// downloadQueue is used to keep track of the progress of each download.
		queue    *Queue
		runCh    chan *QueueInfo // QueueInfo from queue
		cancelCh chan struct{}   // Called by client

		mediaMap MediaMap // Refreshed on start and after each download
	}

	//+-------------------------------------------------------------------------------------------------------------------+

	// MediaMap is used to store all downloaded chapters for each media.
	MediaMap map[int]MediaMapInfo

	// MediaMapInfo stores all downloaded chapters for a specific media.
	MediaMapInfo struct {
		Provider   string
		ChapterIds []string
	}

	DownloadID struct {
		Provider  string `json:"provider"`
		MediaId   int    `json:"mediaId"`
		ChapterId string `json:"chapterId"`
	}

	//+-------------------------------------------------------------------------------------------------------------------+

	// Registry stored in 📄 registry.json for each chapter download.
	Registry map[int]PageInfo

	PageInfo struct {
		Index       int    `json:"index"`
		Filename    string `json:"filename"`
		OriginalURL string `json:"original_url"`
		Size        int64  `json:"size"`
		Width       int    `json:"width"`
		Height      int    `json:"height"`
	}
)

type (
	NewDownloaderOptions struct {
		Logger         *zerolog.Logger
		WSEventManager events.IWSEventManager
		DownloadDir    string
		Database       *db.Database
	}
)

func NewDownloader(opts *NewDownloaderOptions) *Downloader {
	runCh := make(chan *QueueInfo, 1)

	d := &Downloader{
		logger:         opts.Logger,
		wsEventManager: opts.WSEventManager,
		downloadDir:    opts.DownloadDir,
		cancelChannels: make(map[DownloadID]chan struct{}),
		runCh:          runCh,
		queue:          NewQueue(opts.Database, opts.Logger, runCh),
	}

	return d
}

// Start spins up a goroutine that will listen to queue events.
func (cd *Downloader) Start() {
	go func() {
		for {
			select {
			case queueInfo := <-cd.runCh:
				cd.logger.Debug().Msgf("chapter downloader: Received queue item to download: %s", queueInfo.ChapterId)
				cd.run(queueInfo)
			default:
			}
		}
	}()
}

// DownloadChapter downloads a chapter from a manga provider.
// It is the higher-order function called by the client.
// If the chapter is already downloaded, it will delete the previous data and re-download it.
//
// It will spin up a goroutine to add the chapter to the download queue.
func (cd *Downloader) DownloadChapter(provider string, mediaId int, chapterId string, pages []*manga_providers.ChapterPage) error {
	cd.mu.Lock()
	defer cd.mu.Unlock()

	downloadId := DownloadID{Provider: provider, MediaId: mediaId, ChapterId: chapterId}

	// Check if chapter is already downloaded
	registryPath := cd.getChapterRegistryPath(downloadId)
	if _, err := os.Stat(registryPath); err == nil {
		cd.logger.Warn().Msg("chapter downloader: directory already exists, deleting")
		// Delete folder
		_ = os.RemoveAll(cd.getChapterDownloadDir(downloadId))
	}

	// Start download
	cd.logger.Debug().Msgf("chapter downloader: Adding chapter to download queue: %s", chapterId)
	// Add to queue
	return cd.queue.Add(downloadId, pages)
}

// Run starts the downloader.
// It is a higher-order function called by the client.
func (cd *Downloader) Run() {
	cd.mu.Lock()
	defer cd.mu.Unlock()

	cd.queue.Start()
}

// run downloads the chapter based on the queue info provided.
// It's run in a goroutine.
// Ideally this will not be run concurrently.
func (cd *Downloader) run(queueInfo *QueueInfo) {

	// Catch panic in runNext, so it doesn't bubble up and stop goroutines.
	defer util.HandlePanicInModuleThen("internal/manga/downloader/runNext", func() {
		cd.logger.Error().Msg("chapter downloader: Panic in 'run'")
	})

	// Download chapter images
	if err := cd.downloadChapterImages(queueInfo); err != nil {
		return
	}

}

func (cd *Downloader) downloadChapterImages(queueInfo *QueueInfo) (err error) {
	cd.cancelCh = make(chan struct{})

	// Create download directory
	// 📁 {provider}_{mediaId}_{chapterId}
	destination := cd.getChapterDownloadDir(queueInfo.DownloadID)
	if err = os.MkdirAll(destination, os.ModePerm); err != nil {
		cd.logger.Error().Err(err).Msgf("chapter downloader: Failed to create download directory for chapter %s", queueInfo.ChapterId)
		return err
	}

	cd.logger.Debug().Msgf("chapter downloader: Downloading chapter %s images to %s", queueInfo.ChapterId, destination)

	registry := make(Registry)

	// calculateBatchSize calculates the batch size based on the number of URLs.
	calculateBatchSize := func(numURLs int) int {
		maxBatchSize := 1
		batchSize := numURLs / 10
		if batchSize < 1 {
			return 1
		} else if batchSize > maxBatchSize {
			return maxBatchSize
		}
		return batchSize
	}

	// Download images
	batchSize := calculateBatchSize(len(queueInfo.Pages))

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, batchSize) // Semaphore to control concurrency
	for _, page := range queueInfo.Pages {
		semaphore <- struct{}{} // Acquire semaphore
		wg.Add(1)
		go func(page *manga_providers.ChapterPage, registry *Registry) {
			defer func() {
				<-semaphore // Release semaphore
				wg.Done()
			}()
			select {
			case <-cd.cancelCh:
				cd.logger.Warn().Msg("chapter downloader: Download process canceled")
				return
			default:
				cd.downloadPage(page, destination, registry)
			}
		}(page, &registry)
	}
	wg.Wait()

	// Write the registry
	_ = registry.save(queueInfo, destination, cd.logger)

	cd.queue.HasCompleted(queueInfo.DownloadID)

	if queueInfo.Status != QueueStatusErrored {
		cd.logger.Info().Msgf("chapter downloader: Finished downloading chapter %s", queueInfo.ChapterId)
	}

	return
}

func (cd *Downloader) downloadPage(page *manga_providers.ChapterPage, destination string, registry *Registry) {

	defer util.HandlePanicInModuleThen("manga/downloader/downloadImage", func() {
	})

	// Download image from URL
	cd.logger.Debug().Msgf("chapter downloader: Downloading page %d", page.Index)

	imgID := uuid.NewString()

	// Download the image
	resp, err := http.Get(page.URL)
	if err != nil {
		cd.logger.Error().Err(err).Msgf("chapter downloader: Failed to download image from URL %s", page.URL)
		return
	}
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		cd.logger.Error().Err(err).Msgf("chapter downloader: Failed to read image data from URL %s", page.URL)
		return
	}

	// Get the image format
	config, format, err := image.DecodeConfig(bytes.NewReader(buf))
	if err != nil {
		cd.logger.Error().Err(err).Msgf("chapter downloader: Failed to decode image format from URL %s", page.URL)
		return
	}

	filename := imgID + "." + format

	// Create the file
	filePath := filepath.Join(destination, filename)
	file, err := os.Create(filePath)
	if err != nil {
		cd.logger.Error().Err(err).Msgf("chapter downloader: Failed to create file for image %s", imgID)
		return
	}
	defer file.Close()

	// Copy the image data to the file
	_, err = io.Copy(file, bytes.NewReader(buf))
	if err != nil {
		cd.logger.Error().Err(err).Msgf("image downloader: Failed to write image data to file for image from %s", page.URL)
		return
	}

	// Update registry
	cd.downloadMu.Lock()
	(*registry)[page.Index] = PageInfo{
		Index:       page.Index,
		Width:       config.Width,
		Height:      config.Height,
		Filename:    filename,
		OriginalURL: page.URL,
		Size:        int64(len(buf)),
	}
	cd.downloadMu.Unlock()

	return
}

////////////////////////

func (r *Registry) save(queueInfo *QueueInfo, destination string, logger *zerolog.Logger) (err error) {

	defer util.HandlePanicInModuleThen("manga/downloader/save", func() {
		err = fmt.Errorf("chapter downloader: Failed to save registry content")
	})

	// Verify all images have been downloaded
	allDownloaded := true
	for _, page := range queueInfo.Pages {
		if _, ok := (*r)[page.Index]; !ok {
			allDownloaded = false
			break
		}
	}

	if !allDownloaded {
		// Clean up downloaded images
		go func() {
			logger.Error().Msg("chapter downloader: Not all images have been downloaded, aborting")
			// Delete directory
			_ = os.RemoveAll(destination)

			queueInfo.Status = QueueStatusErrored
		}()
		return fmt.Errorf("chapter downloader: Not all images have been downloaded, operation aborted")
	}

	// Create registry file
	var data []byte
	data, err = json.Marshal(*r)
	if err != nil {
		return err
	}

	registryFilePath := filepath.Join(destination, "registry.json")
	err = os.WriteFile(registryFilePath, data, 0644)
	if err != nil {
		return err
	}

	return
}

func (cd *Downloader) getChapterDownloadDir(downloadId DownloadID) string {
	return filepath.Join(cd.downloadDir, fmt.Sprintf("%s_%d_%s", downloadId.Provider, downloadId.MediaId, downloadId.ChapterId))
}

func (cd *Downloader) getChapterRegistryPath(downloadId DownloadID) string {
	return filepath.Join(cd.getChapterDownloadDir(downloadId), "registry.json")
}
