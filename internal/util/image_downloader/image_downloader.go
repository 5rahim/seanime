package image_downloader

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"seanime/internal/util"
	"seanime/internal/util/limiter"
	"slices"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

const (
	RegistryFilename = "registry.json"
)

type (
	ImageDownloader struct {
		downloadDir   string
		registry      Registry
		cancelChannel chan struct{}
		logger        *zerolog.Logger
		actionMu      sync.Mutex
		registryMu    sync.Mutex
	}

	Registry struct {
		content      *RegistryContent
		logger       *zerolog.Logger
		downloadDir  string
		registryPath string
		mu           sync.Mutex
	}
	RegistryContent struct {
		UrlToId map[string]string `json:"url_to_id"`
		IdToUrl map[string]string `json:"id_to_url"`
		IdToExt map[string]string `json:"id_to_ext"`
	}
)

func NewImageDownloader(downloadDir string, logger *zerolog.Logger) *ImageDownloader {
	_ = os.MkdirAll(downloadDir, os.ModePerm)

	return &ImageDownloader{
		downloadDir: downloadDir,
		logger:      logger,
		registry: Registry{
			logger:       logger,
			registryPath: filepath.Join(downloadDir, RegistryFilename),
			downloadDir:  downloadDir,
			content:      &RegistryContent{},
		},
		cancelChannel: make(chan struct{}),
	}
}

// DownloadImages downloads multiple images concurrently.
func (id *ImageDownloader) DownloadImages(urls []string) (err error) {
	id.cancelChannel = make(chan struct{})

	if err = id.registry.setup(); err != nil {
		return
	}

	rateLimiter := limiter.NewLimiter(1*time.Second, 10)
	var wg sync.WaitGroup
	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			select {
			case <-id.cancelChannel:
				id.logger.Warn().Msg("image downloader: Download process canceled")
				return
			default:
				rateLimiter.Wait()
				id.downloadImage(url)
			}
		}(url)
	}
	wg.Wait()

	if err = id.registry.save(urls); err != nil {
		return
	}

	return
}

func (id *ImageDownloader) DeleteDownloads() {
	id.actionMu.Lock()
	defer id.actionMu.Unlock()

	id.registryMu.Lock()
	defer id.registryMu.Unlock()

	_ = os.RemoveAll(id.downloadDir)
	id.registry.content = &RegistryContent{}
}

// CancelDownload cancels the download process.
func (id *ImageDownloader) CancelDownload() {
	close(id.cancelChannel)
}

func (id *ImageDownloader) GetImageFilenameByUrl(url string) (filename string, ok bool) {
	id.actionMu.Lock()
	defer id.actionMu.Unlock()

	id.registryMu.Lock()
	defer id.registryMu.Unlock()

	if err := id.registry.setup(); err != nil {
		return
	}

	var imgID string
	imgID, ok = id.registry.content.UrlToId[url]
	if !ok {
		return
	}

	filename = imgID + "." + id.registry.content.IdToExt[imgID]
	return
}

// GetImageFilenamesByUrls returns a map of URLs to image filenames.
//
//	e.g., {"url1": "filename1.png", "url2": "filename2.jpg"}
func (id *ImageDownloader) GetImageFilenamesByUrls(urls []string) (ret map[string]string, err error) {
	id.actionMu.Lock()
	defer id.actionMu.Unlock()

	id.registryMu.Lock()
	defer id.registryMu.Unlock()

	ret = make(map[string]string)

	if err = id.registry.setup(); err != nil {
		return nil, err
	}

	for _, url := range urls {
		imgID, ok := id.registry.content.UrlToId[url]
		if !ok {
			continue
		}

		ret[url] = imgID + "." + id.registry.content.IdToExt[imgID]
	}
	return
}

func (id *ImageDownloader) DeleteImagesByUrls(urls []string) (err error) {
	id.actionMu.Lock()
	defer id.actionMu.Unlock()

	id.registryMu.Lock()
	defer id.registryMu.Unlock()

	if err = id.registry.setup(); err != nil {
		return
	}

	for _, url := range urls {
		imgID, ok := id.registry.content.UrlToId[url]
		if !ok {
			continue
		}

		err = os.Remove(filepath.Join(id.downloadDir, imgID+"."+id.registry.content.IdToExt[imgID]))
		if err != nil {
			continue
		}

		delete(id.registry.content.UrlToId, url)
		delete(id.registry.content.IdToUrl, imgID)
		delete(id.registry.content.IdToExt, imgID)
	}
	return
}

// downloadImage downloads an image from a URL.
func (id *ImageDownloader) downloadImage(url string) {

	defer util.HandlePanicInModuleThen("util/image_downloader/downloadImage", func() {
	})

	if url == "" {
		id.logger.Warn().Msg("image downloader: Empty URL provided, skipping download")
		return
	}

	// Check if the image has already been downloaded
	id.registryMu.Lock()
	if _, ok := id.registry.content.UrlToId[url]; ok {
		id.registryMu.Unlock()
		id.logger.Debug().Msgf("image downloader: Image from URL %s has already been downloaded", url)
		return
	}
	id.registryMu.Unlock()

	// Download image from URL
	id.logger.Info().Msgf("image downloader: Downloading image from URL: %s", url)

	imgID := uuid.NewString()

	// Download the image
	resp, err := http.Get(url)
	if err != nil {
		id.logger.Error().Err(err).Msgf("image downloader: Failed to download image from URL %s", url)
		return
	}
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		id.logger.Error().Err(err).Msgf("image downloader: Failed to read image data from URL %s", url)
		return
	}

	// Get the image format
	_, format, err := image.DecodeConfig(bytes.NewReader(buf))
	if err != nil {
		id.logger.Error().Err(err).Msgf("image downloader: Failed to decode image format from URL %s", url)
		return
	}

	// Create the file
	filePath := filepath.Join(id.downloadDir, imgID+"."+format)
	file, err := os.Create(filePath)
	if err != nil {
		id.logger.Error().Err(err).Msgf("image downloader: Failed to create file for image %s", imgID)
		return
	}
	defer file.Close()

	// Copy the image data to the file
	_, err = io.Copy(file, bytes.NewReader(buf))
	if err != nil {
		id.logger.Error().Err(err).Msgf("image downloader: Failed to write image data to file for image from %s", url)
		return
	}

	// Update registry
	id.registryMu.Lock()
	id.registry.addUrl(imgID, url, format)
	id.registryMu.Unlock()

	return
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Registry) setup() (err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	defer util.HandlePanicInModuleThen("util/image_downloader/setup", func() {
		err = fmt.Errorf("image downloader: Failed to setup registry")
	})

	if r.content.IdToUrl != nil && r.content.UrlToId != nil {
		return nil
	}

	r.content.UrlToId = make(map[string]string)
	r.content.IdToUrl = make(map[string]string)
	r.content.IdToExt = make(map[string]string)

	// Check if the registry exists
	_ = os.MkdirAll(filepath.Dir(r.registryPath), os.ModePerm)
	_, err = os.Stat(r.registryPath)
	if os.IsNotExist(err) {
		// Create the registry file
		err = os.WriteFile(r.registryPath, []byte("{}"), os.ModePerm)
		if err != nil {
			return err
		}
	}

	// Read the registry file
	file, err := os.Open(r.registryPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Decode the registry file if there is content
	if file != nil {
		r.logger.Debug().Msg("image downloader: Reading registry content")
		err = json.NewDecoder(file).Decode(&r.content)
		if err != nil {
			return err
		}
	}

	if r.content == nil {
		r.content = &RegistryContent{
			UrlToId: make(map[string]string),
			IdToUrl: make(map[string]string),
			IdToExt: make(map[string]string),
		}
	}

	return nil
}

// save verifies and saves the registry content.
func (r *Registry) save(urls []string) (err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	defer util.HandlePanicInModuleThen("util/image_downloader/save", func() {
		err = fmt.Errorf("image downloader: Failed to save registry content")
	})

	// Verify all images have been downloaded
	allDownloaded := true
	for _, url := range urls {
		if url == "" {
			continue
		}
		if _, ok := r.content.UrlToId[url]; !ok {
			allDownloaded = false
			break
		}
	}

	if !allDownloaded {
		// Clean up downloaded images
		go func() {
			r.logger.Error().Msg("image downloader: Not all images have been downloaded, aborting")
			// Read the directory
			files, err := os.ReadDir(r.downloadDir)
			if err != nil {
				r.logger.Error().Err(err).Msg("image downloader: Failed to abort")
				return
			}
			// Delete all files that have been downloaded (are in the registry)
			for _, file := range files {
				fileNameWithoutExt := file.Name()[:len(file.Name())-len(filepath.Ext(file.Name()))]
				if url, ok := r.content.IdToUrl[fileNameWithoutExt]; ok && slices.Contains(urls, url) {
					err = os.Remove(filepath.Join(r.downloadDir, file.Name()))
					if err != nil {
						r.logger.Error().Err(err).Msgf("image downloader: Failed to delete file %s", file.Name())
					}
				}
			}
		}()
		return fmt.Errorf("image downloader: Not all images have been downloaded, operation aborted")
	}

	data, err := json.Marshal(r.content)
	if err != nil {
		r.logger.Error().Err(err).Msg("image downloader: Failed to marshal registry content")
	}
	// Overwrite the registry file
	err = os.WriteFile(r.registryPath, data, 0644)
	if err != nil {
		r.logger.Error().Err(err).Msg("image downloader: Failed to write registry content")
		return err
	}

	return nil
}

func (r *Registry) addUrl(imgID, url, format string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.content.UrlToId[url] = imgID
	r.content.IdToUrl[imgID] = url
	r.content.IdToExt[imgID] = format
}
