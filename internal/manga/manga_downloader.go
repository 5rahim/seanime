package manga

import (
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/database/db"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/manga/downloader"
	"github.com/seanime-app/seanime/internal/manga/providers"
	"os"
	"strconv"
	"strings"
	"sync"
)

type (
	Downloader struct {
		logger            *zerolog.Logger
		wsEventManager    events.IWSEventManager
		database          *db.Database
		downloadDir       string
		chapterDownloader *chapter_downloader.Downloader

		mediaMap   MediaMap // Refreshed on start and after each download
		mediaMapMu sync.RWMutex
	}

	// MediaMap is used to store all downloaded chapters for each media.
	//
	//	e.g., downloadDir/comick_1234_abc/
	//	      downloadDir/comick_1234_def/
	// -> map[1234]["comick"] = ["abc", "def"]
	MediaMap map[int]map[string][]string
)

type (
	NewDownloaderOptions struct {
		Database       *db.Database
		Logger         *zerolog.Logger
		WSEventManager events.IWSEventManager
		DownloadDir    string
	}
)

func NewDownloader(opts *NewDownloaderOptions) *Downloader {
	d := &Downloader{
		logger:         opts.Logger,
		wsEventManager: opts.WSEventManager,
		database:       opts.Database,
		downloadDir:    opts.DownloadDir,
	}

	d.chapterDownloader = chapter_downloader.NewDownloader(&chapter_downloader.NewDownloaderOptions{
		Logger:         opts.Logger,
		WSEventManager: opts.WSEventManager,
		Database:       opts.Database,
		DownloadDir:    opts.DownloadDir,
	})

	return d
}

func (d *Downloader) Start() {
	d.chapterDownloader.Start()
}

func (d *Downloader) DownloadChapter(provider string, mediaId int, chapterId string, pages []*manga_providers.ChapterPage) error {
	return d.chapterDownloader.DownloadChapter(provider, mediaId, chapterId, pages)
}

func (d *Downloader) DeleteChapter(provider string, mediaId int, chapterId string) error {
	return d.chapterDownloader.DeleteChapter(provider, mediaId, chapterId)
}

func (d *Downloader) GetMediaMap() MediaMap {
	d.mediaMapMu.RLock()
	defer d.mediaMapMu.RUnlock()
	return d.mediaMap
}

func (d *Downloader) refreshMediaMap() {
	d.mediaMapMu.Lock()
	defer d.mediaMapMu.Unlock()

	ret := make(MediaMap)

	files, err := os.ReadDir(d.downloadDir)
	if err != nil {
		d.logger.Error().Err(err).Msg("manga downloader: Failed to read download directory")
	}

	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	for _, file := range files {
		wg.Add(1)
		go func(file os.DirEntry) {
			defer wg.Done()

			if file.IsDir() {
				parts := strings.SplitN(file.Name(), "_", 3)
				if len(parts) != 3 {
					return
				}

				provider := parts[0]
				mediaID, err := strconv.Atoi(parts[1])
				chapterID := parts[2]

				if err != nil {
					return
				}

				mu.Lock()
				if _, ok := ret[mediaID]; !ok {
					ret[mediaID] = make(map[string][]string)
					ret[mediaID][provider] = []string{chapterID}
				} else {
					if _, ok := ret[mediaID][provider]; !ok {
						ret[mediaID][provider] = []string{chapterID}
					} else {
						ret[mediaID][provider] = append(ret[mediaID][provider], chapterID)
					}
				}
				mu.Unlock()
			}
		}(file)
	}
	wg.Wait()

	d.mediaMap = ret
}
