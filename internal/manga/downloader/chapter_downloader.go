package chapter_downloader

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"seanime/internal/database/db"
	"seanime/internal/events"
	hibikemanga "seanime/internal/extension/hibike/manga"
	manga_providers "seanime/internal/manga/providers"
	"seanime/internal/util"
	"strconv"
	"strings"
	"sync"

	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
	_ "golang.org/x/image/bmp"  // Register BMP format
	_ "golang.org/x/image/tiff" // Register Tiff format
)

// 📁 cache/manga
// └── 📁 {provider}_{mediaId}_{chapterId}_{chapterNumber}      <- Downloader generates
//     ├── 📄 registry.json						                <- Contains Registry
//     ├── 📄 1.jpg
//     ├── 📄 2.jpg
//     └── 📄 ...
//

type (
	// Downloader is used to download chapters from various manga providers.
	Downloader struct {
		logger         *zerolog.Logger
		wsEventManager events.WSEventManagerInterface
		database       *db.Database
		downloadDir    string
		mu             sync.Mutex
		downloadMu     sync.Mutex
		// cancelChannel is used to cancel some or all downloads.
		cancelChannels      map[DownloadID]chan struct{}
		queue               *Queue
		cancelCh            chan struct{}   // Close to cancel the download process
		runCh               chan *QueueInfo // Receives a signal to download the next item
		chapterDownloadedCh chan DownloadID // Sends a signal when a chapter has been downloaded
	}

	//+-------------------------------------------------------------------------------------------------------------------+

	DownloadID struct {
		Provider      string `json:"provider"`
		MediaId       int    `json:"mediaId"`
		ChapterId     string `json:"chapterId"`
		ChapterNumber string `json:"chapterNumber"`
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
		WSEventManager events.WSEventManagerInterface
		DownloadDir    string
		Database       *db.Database
	}

	DownloadOptions struct {
		DownloadID
		Pages    []*hibikemanga.ChapterPage
		StartNow bool
	}
)

func NewDownloader(opts *NewDownloaderOptions) *Downloader {
	runCh := make(chan *QueueInfo, 1)

	d := &Downloader{
		logger:              opts.Logger,
		wsEventManager:      opts.WSEventManager,
		downloadDir:         opts.DownloadDir,
		cancelChannels:      make(map[DownloadID]chan struct{}),
		runCh:               runCh,
		queue:               NewQueue(opts.Database, opts.Logger, opts.WSEventManager, runCh),
		chapterDownloadedCh: make(chan DownloadID, 100),
	}

	return d
}

// Start spins up a goroutine that will listen to queue events.
func (cd *Downloader) Start() {
	go func() {
		for {
			select {
			// Listen for new queue items
			case queueInfo := <-cd.runCh:
				cd.logger.Debug().Msgf("chapter downloader: Received queue item to download: %s", queueInfo.ChapterId)
				cd.run(queueInfo)
			}
		}
	}()
}

func (cd *Downloader) ChapterDownloaded() <-chan DownloadID {
	return cd.chapterDownloadedCh
}

// AddToQueue adds a chapter to the download queue.
// If the chapter is already downloaded (i.e. a folder already exists), it will delete the previous data and re-download it.
func (cd *Downloader) AddToQueue(opts DownloadOptions) error {
	cd.mu.Lock()
	defer cd.mu.Unlock()

	downloadId := opts.DownloadID

	// Check if chapter is already downloaded
	registryPath := cd.getChapterRegistryPath(downloadId)
	if _, err := os.Stat(registryPath); err == nil {
		cd.logger.Warn().Msg("chapter downloader: directory already exists, deleting")
		// Delete folder
		_ = os.RemoveAll(cd.getChapterDownloadDir(downloadId))
	}

	// Start download
	cd.logger.Debug().Msgf("chapter downloader: Adding chapter to download queue: %s", opts.ChapterId)
	// Add to queue
	return cd.queue.Add(downloadId, opts.Pages, opts.StartNow)
}

// DeleteChapter deletes a chapter directory from the download directory.
func (cd *Downloader) DeleteChapter(id DownloadID) error {
	cd.mu.Lock()
	defer cd.mu.Unlock()

	cd.logger.Debug().Msgf("chapter downloader: Deleting chapter %s", id.ChapterId)

	_ = os.RemoveAll(cd.getChapterDownloadDir(id))
	cd.logger.Debug().Msgf("chapter downloader: Removed chapter %s", id.ChapterId)
	return nil
}

// Run starts the downloader if it's not already running.
func (cd *Downloader) Run() {
	cd.mu.Lock()
	defer cd.mu.Unlock()

	cd.logger.Debug().Msg("chapter downloader: Starting queue")

	cd.cancelCh = make(chan struct{})

	cd.queue.Run()
}

// Stop cancels the download process and stops the queue from running.
func (cd *Downloader) Stop() {
	cd.mu.Lock()
	defer cd.mu.Unlock()

	defer func() {
		if r := recover(); r != nil {
			cd.logger.Error().Msgf("chapter downloader: cancelCh is already closed")
		}
	}()

	cd.cancelCh = make(chan struct{})

	close(cd.cancelCh) // Cancel download process

	cd.queue.Stop()
}

// run downloads the chapter based on the QueueInfo provided.
// This is called successively for each current item being processed.
// It invokes downloadChapterImages to download the chapter pages.
func (cd *Downloader) run(queueInfo *QueueInfo) {

	defer util.HandlePanicInModuleThen("internal/manga/downloader/runNext", func() {
		cd.logger.Error().Msg("chapter downloader: Panic in 'run'")
	})

	// Download chapter images
	if err := cd.downloadChapterImages(queueInfo); err != nil {
		return
	}

	cd.chapterDownloadedCh <- queueInfo.DownloadID
}

// downloadChapterImages creates a directory for the chapter and downloads each image to that directory.
// It also creates a Registry file that contains information about each image.
//
//	e.g.,
//	📁 {provider}_{mediaId}_{chapterId}_{chapterNumber}
//	   ├── 📄 registry.json
//	   ├── 📄 1.jpg
//	   ├── 📄 2.jpg
//	   └── 📄 ...
func (cd *Downloader) downloadChapterImages(queueInfo *QueueInfo) (err error) {

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
		maxBatchSize := 5
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
		go func(page *hibikemanga.ChapterPage, registry *Registry) {
			defer func() {
				<-semaphore // Release semaphore
				wg.Done()
			}()
			select {
			case <-cd.cancelCh:
				//cd.logger.Warn().Msg("chapter downloader: Download goroutine canceled")
				return
			default:
				cd.downloadPage(page, destination, registry)
			}
		}(page, &registry)
	}
	wg.Wait()

	// Write the registry
	_ = registry.save(queueInfo, destination, cd.logger)

	cd.queue.HasCompleted(queueInfo)

	if queueInfo.Status != QueueStatusErrored {
		cd.logger.Info().Msgf("chapter downloader: Finished downloading chapter %s", queueInfo.ChapterId)
	}

	if queueInfo.Status == QueueStatusErrored {
		return fmt.Errorf("chapter downloader: Failed to download chapter %s", queueInfo.ChapterId)
	}

	return
}

// downloadPage downloads a single page from the URL and saves it to the destination directory.
// It also updates the Registry with the page information.
func (cd *Downloader) downloadPage(page *hibikemanga.ChapterPage, destination string, registry *Registry) {

	defer util.HandlePanicInModuleThen("manga/downloader/downloadImage", func() {
	})

	// Download image from URL

	imgID := fmt.Sprintf("%02d", page.Index+1)

	buf, err := manga_providers.GetImageByProxy(page.URL, page.Headers)
	if err != nil {
		cd.logger.Error().Err(err).Msgf("chapter downloader: Failed to get image from URL %s", page.URL)
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

// save saves the Registry content to a file in the chapter directory.
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
		logger.Error().Msg("chapter downloader: Not all images have been downloaded, aborting")
		queueInfo.Status = QueueStatusErrored
		// Delete directory
		go os.RemoveAll(destination)
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
	return filepath.Join(cd.downloadDir, FormatChapterDirName(downloadId.Provider, downloadId.MediaId, downloadId.ChapterId, downloadId.ChapterNumber))
}

func FormatChapterDirName(provider string, mediaId int, chapterId string, chapterNumber string) string {
	return fmt.Sprintf("%s_%d_%s_%s", provider, mediaId, EscapeChapterID(chapterId), chapterNumber)
}

// ParseChapterDirName parses a chapter directory name and returns the DownloadID.
// e.g. comick_1234_chapter$UNDERSCORE$id_13.5 -> {Provider: "comick", MediaId: 1234, ChapterId: "chapter_id", ChapterNumber: "13.5"}
func ParseChapterDirName(dirName string) (id DownloadID, ok bool) {
	parts := strings.Split(dirName, "_")
	if len(parts) != 4 {
		return id, false
	}

	id.Provider = parts[0]
	var err error
	id.MediaId, err = strconv.Atoi(parts[1])
	if err != nil {
		return id, false
	}
	id.ChapterId = UnescapeChapterID(parts[2])
	id.ChapterNumber = parts[3]

	ok = true
	return
}

func EscapeChapterID(id string) string {
	id = strings.ReplaceAll(id, "/", "$SLASH$")
	id = strings.ReplaceAll(id, "\\", "$BSLASH$")
	id = strings.ReplaceAll(id, ":", "$COLON$")
	id = strings.ReplaceAll(id, "*", "$ASTERISK$")
	id = strings.ReplaceAll(id, "?", "$QUESTION$")
	id = strings.ReplaceAll(id, "\"", "$QUOTE$")
	id = strings.ReplaceAll(id, "<", "$LT$")
	id = strings.ReplaceAll(id, ">", "$GT$")
	id = strings.ReplaceAll(id, "|", "$PIPE$")
	id = strings.ReplaceAll(id, ".", "$DOT$")
	id = strings.ReplaceAll(id, " ", "$SPACE$")
	id = strings.ReplaceAll(id, "_", "$UNDERSCORE$")
	return id
}

func UnescapeChapterID(id string) string {
	id = strings.ReplaceAll(id, "$SLASH$", "/")
	id = strings.ReplaceAll(id, "$BSLASH$", "\\")
	id = strings.ReplaceAll(id, "$COLON$", ":")
	id = strings.ReplaceAll(id, "$ASTERISK$", "*")
	id = strings.ReplaceAll(id, "$QUESTION$", "?")
	id = strings.ReplaceAll(id, "$QUOTE$", "\"")
	id = strings.ReplaceAll(id, "$LT$", "<")
	id = strings.ReplaceAll(id, "$GT$", ">")
	id = strings.ReplaceAll(id, "$PIPE$", "|")
	id = strings.ReplaceAll(id, "$DOT$", ".")
	id = strings.ReplaceAll(id, "$SPACE$", " ")
	id = strings.ReplaceAll(id, "$UNDERSCORE$", "_")
	return id
}

func (cd *Downloader) getChapterRegistryPath(downloadId DownloadID) string {
	return filepath.Join(cd.getChapterDownloadDir(downloadId), "registry.json")
}
