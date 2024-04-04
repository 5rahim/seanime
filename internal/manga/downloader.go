package manga

import (
	"bytes"
	"context"
	"fmt"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/manga/providers"
	"github.com/seanime-app/seanime/internal/util"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type downloader struct {
	logger         *zerolog.Logger
	wsEventManager events.IWSEventManager
}

func newDownloader(logger *zerolog.Logger, wsEventManager events.IWSEventManager) *downloader {
	return &downloader{logger: logger, wsEventManager: wsEventManager}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// PageMap is used to store the mapping of page index to page details of a specific downloaded chapter in a main.txt file.
type PageMap map[int]PageInfo

type PageInfo struct {
	Index       int    `json:"index"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Filename    string `json:"filename"`
	OriginalURL string `json:"original_url"`
	Size        int64  `json:"size"`
}

func (p *PageInfo) ToChapterPage(key string) *manga_providers.ChapterPage {
	return &manga_providers.ChapterPage{
		Index: p.Index,
		URL:   fmt.Sprintf("/%s/%s", key, p.Filename),
	}
}

// DownloadMap is used to store the mapping of a DownloadID to a list of chapterIDs.
//
//	e.g., downloadDir/comick_1234_One-Piece$10010/
//	      downloadDir/comick_1234_One-Piece$10023/
//	-> map[DownloadID{ Provider: "comick", MediaID: 1234 }] = []string{"One-Piece$10010", "One-Piece$10023"}
type DownloadMap map[DownloadID][]string

// DownloadID represents the unique identifier for a backup folder group
type DownloadID struct {
	Provider string
	MediaID  int
}

// getDownloads scans the backup folder and creates a DownloadMap
func (d *downloader) getDownloads(downloadDir string) (DownloadMap, error) {
	ret := make(DownloadMap)

	files, err := os.ReadDir(downloadDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache directory: %v", err)
	}

	for _, file := range files {
		if file.IsDir() {
			parts := strings.SplitN(file.Name(), "_", 3)
			if len(parts) != 3 {
				continue
			}

			provider := parts[0]
			mediaID, _ := strconv.Atoi(parts[1])
			chapterID := parts[2]

			comicID := DownloadID{Provider: provider, MediaID: mediaID}

			if _, ok := ret[comicID]; !ok {
				ret[comicID] = []string{chapterID}
			} else {
				ret[comicID] = append(ret[comicID], chapterID)
			}
		}
	}

	return ret, nil
}

// deleteDownloads deletes the cache directory for a given provider, mediaID, and chapterID.
// If the cache directory does not exist, an error is returned.
func (d *downloader) deleteDownloads(provider string, mediaID int, chapterID string, downloadDir string) error {
	comicDir := fmt.Sprintf("%s_%d_%s", provider, mediaID, chapterID)
	comicPath := filepath.Join(downloadDir, comicDir)

	if _, err := os.Stat(comicPath); os.IsNotExist(err) {
		return fmt.Errorf("manga downloader: cache directory does not exist")
	}

	if err := os.RemoveAll(comicPath); err != nil {
		return fmt.Errorf("manga downloader: failed to delete cache directory: %v", err)
	}

	//fmt.Printf("Cache directory %s deleted successfully\n", comicPath)
	return nil
}

// getPageMap retrieves page details based on provider, mediaID, chapterID, and downloadDir.
func (d *downloader) getPageMap(provider string, mediaID int, chapterID string, downloadDir string) (pm *PageMap, dirName string, err error) {
	defer util.HandlePanicInModuleThen("manga/downloader/downloadImages", func() {
		err = fmt.Errorf("manga downloader: failed to get page map")
	})

	comicDir := fmt.Sprintf("%s_%d_%s", provider, mediaID, chapterID)
	comicPath := filepath.Join(downloadDir, comicDir)

	mainFilePath := filepath.Join(comicPath, "main.txt")

	file, err := os.Open(mainFilePath)
	if err != nil {
		return nil, "", fmt.Errorf("manga downloader: failed to open main.txt: %v", err)
	}
	defer file.Close()

	var pages PageMap
	if err := json.NewDecoder(file).Decode(&pages); err != nil {
		return nil, "", fmt.Errorf("manga downloader: failed to decode main.txt: %v", err)
	}

	return &pages, comicDir, nil
}

var (
	errDownloadCanceled = fmt.Errorf("manga downloader: download process canceled")
)

// downloadImages concurrently downloads images from given URLs and saves them to a directory
// with the specified provider, media ID, and chapter ID.
//
//	e.g., downloadDir/comick_1234_One-Piece$10010/...
//	e.g., downloadDir/comick_1234_One-Piece$10023/...
func (d *downloader) downloadImages(ctx context.Context, provider string, mediaID int, chapterID string, pages []*manga_providers.ChapterPage, downloadDir string) (err error) {
	// Channel to receive errors from goroutines
	errCh := make(chan error, len(pages))

	defer util.HandlePanicInModuleThen("manga/downloader/downloadImages", func() {
		err = fmt.Errorf("manga downloader: failed to download images")
	})

	defer func() {
		d.wsEventManager.SendEvent(events.MangaDownloaderDownloadingProgress, struct {
			ChapterId string `json:"chapterId"`
			Number    int    `json:"number"`
		}{
			ChapterId: chapterID,
			Number:    0, // Signal that all images have been downloaded
		})
	}()

	// Create directory for the comic
	comicDir := fmt.Sprintf("%s_%d_%s", provider, mediaID, chapterID)
	comicPath := filepath.Join(downloadDir, comicDir)
	if err := os.MkdirAll(comicPath, 0755); err != nil {
		return fmt.Errorf("manga downloader: failed to create comic directory: %v", err)
	}

	// Create main.txt file to store image filenames and dimensions
	mainFile, err := os.Create(filepath.Join(comicPath, "main.txt"))
	if err != nil {
		return fmt.Errorf("manga downloader: failed to create main.txt: %v", err)
	}
	defer mainFile.Close()

	var wg sync.WaitGroup
	var mu sync.Mutex // Mutex to protect access to mainFile

	// Create a map to store image metadata
	imageMetadata := make(map[int]PageInfo)

	d.wsEventManager.SendEvent(events.MangaDownloaderDownloadingProgress, struct {
		ChapterId string `json:"chapterId"`
		Number    int    `json:"number"`
	}{
		ChapterId: chapterID,
		Number:    len(pages),
	})

	// Start goroutines to download images
	for _, page := range pages {
		wg.Add(1)
		go func(page *manga_providers.ChapterPage, ctx context.Context, wg *sync.WaitGroup) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				d.logger.Warn().Msg("manga downloader: download process canceled")
				errCh <- errDownloadCanceled
				return
			default:
				url := page.URL

				// Download image from URL
				resp, err := http.Get(url)
				if err != nil {
					d.logger.Error().Err(err).Msgf("manga downloader: failed to download image from URL %s", url)
					errCh <- fmt.Errorf("manga downloader: failed to download image from URL %s", url)
					return
				}
				defer resp.Body.Close()

				// Determine file extension based on Content-Type header
				ext := ".webp" // Default to webp
				contentType := resp.Header.Get("Content-Type")
				if contentType == "image/jpeg" || contentType == "image/jpg" {
					ext = ".jpg"
				} else if contentType == "image/png" {
					ext = ".png"
				}

				// Create filename for the downloaded image
				filename := fmt.Sprintf("%d_%s%s", page.Index, comicDir, ext)
				filePath := filepath.Join(comicPath, filename)

				// Create and write image data to file
				file, err := os.Create(filePath)
				if err != nil {
					d.logger.Error().Err(err).Msgf("manga downloader: failed to create file for image %s", filename)
					errCh <- fmt.Errorf("manga downloader: failed to create file for image %s", filename)
					return
				}
				defer file.Close()

				var (
					buf           []byte
					contentLength int64
				)

				// if the content length is unknown
				if resp.ContentLength == -1 {
					buf, err = io.ReadAll(resp.Body)
					contentLength = int64(len(buf))
				} else {
					contentLength = resp.ContentLength
					buf = make([]byte, resp.ContentLength)
					_, err = io.ReadFull(resp.Body, buf)
				}

				if err != nil {
					d.logger.Error().Err(err).Msgf("manga downloader: failed to read image data from URL %s", url)
					errCh <- fmt.Errorf("manga downloader: failed to read image data from URL %s", url)
					return
				}

				if _, err := file.Write(buf); err != nil {
					d.logger.Error().Err(err).Msgf("manga downloader: failed to write image data to file %s", filename)
					errCh <- fmt.Errorf("manga downloader: failed to write image data to file %s", filename)
					return
				}

				// Decode image to get its dimensions
				img, _, err := image.DecodeConfig(bytes.NewReader(buf))
				if err != nil {
					d.logger.Error().Err(err).Msgf("manga downloader: failed to decode image %s", filename)
					errCh <- fmt.Errorf("manga downloader: failed to decode image %s", filename)
					return
				}
				width := img.Width
				height := img.Height

				mu.Lock()
				defer mu.Unlock()

				imageMetadata[page.Index] = PageInfo{
					Index:       page.Index,
					Width:       width,
					Height:      height,
					Filename:    filename,
					OriginalURL: url,
					Size:        contentLength,
				}

				d.logger.Debug().Str("filename", filename).Msg("image downloaded")
			}
		}(page, ctx, &wg)
	}

	// Wait for all goroutines to finish or for cancellation
	go func() {
		wg.Wait()
		close(errCh)
	}()

	// Handle errors from goroutines
	for err := range errCh {
		if err != nil {
			return err
		}
	}

	// Write imageMetadata map to main.txt file
	jsonBytes, err := json.Marshal(imageMetadata)
	if err != nil {
		return fmt.Errorf("manga downloader: failed to encode image metadata to JSON: %v", err)
	}

	if _, err := mainFile.Write(jsonBytes); err != nil {
		return fmt.Errorf("manga downloader: failed to write image metadata to main.txt: %v", err)
	}

	d.logger.Info().Msgf("manga downloader: chapter downloaded successfully")
	return nil
}
