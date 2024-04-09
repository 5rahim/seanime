package manga

import (
	"errors"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/database/db"
	"github.com/seanime-app/seanime/internal/database/models"
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
		repository        *Repository

		mediaMap   *MediaMap // Refreshed on start and after each download
		mediaMapMu sync.RWMutex
	}

	// MediaMap is used to store all downloaded chapters for each media.
	//
	//	e.g., downloadDir/comick_1234_abc/
	//	      downloadDir/comick_1234_def/
	// -> map[1234]["comick"] = ["abc", "def"]
	MediaMap map[int]ProviderDownloadMap

	// ProviderDownloadMap is used to store all downloaded chapters for a specific media and provider.
	ProviderDownloadMap map[string][]string

	MediaDownloadData struct {
		Downloaded ProviderDownloadMap `json:"downloaded"`
		Queued     ProviderDownloadMap `json:"queued"`
	}
)

type (
	NewDownloaderOptions struct {
		Database       *db.Database
		Logger         *zerolog.Logger
		WSEventManager events.IWSEventManager
		DownloadDir    string
		Repository     *Repository
	}

	DownloadChapterOptions struct {
		Provider  manga_providers.Provider
		MediaId   int
		ChapterId string
		StartNow  bool
	}
)

func NewDownloader(opts *NewDownloaderOptions) *Downloader {
	d := &Downloader{
		logger:         opts.Logger,
		wsEventManager: opts.WSEventManager,
		database:       opts.Database,
		downloadDir:    opts.DownloadDir,
		repository:     opts.Repository,
		mediaMap:       new(MediaMap),
	}

	d.chapterDownloader = chapter_downloader.NewDownloader(&chapter_downloader.NewDownloaderOptions{
		Logger:         opts.Logger,
		WSEventManager: opts.WSEventManager,
		Database:       opts.Database,
		DownloadDir:    opts.DownloadDir,
	})

	go d.refreshMediaMap()

	return d
}

// Start is called once to start the Chapter downloader 's main goroutine.
func (d *Downloader) Start() {
	d.chapterDownloader.Start()
}

// DownloadChapter is called by the client to download a chapter.
// It fetches the chapter pages by using Repository.GetMangaPageContainer
// and invokes the chapter_downloader.Downloader 'Download' method to add the chapter to the download queue.
func (d *Downloader) DownloadChapter(opts DownloadChapterOptions) error {

	// Fetch the chapter pages
	pageContainer, err := d.repository.GetMangaPageContainer(opts.Provider, opts.MediaId, opts.ChapterId, false)
	if err != nil {
		return err
	}

	// Add the chapter to the download queue
	return d.chapterDownloader.Download(chapter_downloader.DownloadOptions{
		DownloadID: chapter_downloader.DownloadID{
			Provider:  string(opts.Provider),
			MediaId:   opts.MediaId,
			ChapterId: opts.ChapterId,
		},
		Pages: pageContainer.Pages,
	})
}

func (d *Downloader) DeleteChapter(provider string, mediaId int, chapterId string) error {
	return d.chapterDownloader.DeleteChapter(chapter_downloader.DownloadID{
		Provider:  provider,
		MediaId:   mediaId,
		ChapterId: chapterId,
	})
}

func (d *Downloader) GetMediaDownloads(mediaId int) (MediaDownloadData, error) {
	return d.mediaMap.getMediaDownload(mediaId, d.database)
}

func (d *Downloader) RefreshMediaMap() *MediaMap {
	d.refreshMediaMap()
	return d.mediaMap
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Media map
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (mm *MediaMap) getMediaDownload(mediaId int, db *db.Database) (MediaDownloadData, error) {
	if mm == nil {
		return MediaDownloadData{}, errors.New("could not check downloaded chapters")
	}

	// Get all downloaded chapters for the media
	downloads, ok := (*mm)[mediaId]
	if !ok {
		downloads = make(map[string][]string)
	}

	// Get all queued chapters for the media
	queued, err := db.GetMediaQueuedChapters(mediaId)
	if err != nil {
		queued = make([]*models.ChapterDownloadQueueItem, 0)
	}

	qm := make(ProviderDownloadMap)
	for _, item := range queued {
		if _, ok := qm[item.Provider]; !ok {
			qm[item.Provider] = []string{item.ChapterID}
		} else {
			qm[item.Provider] = append(qm[item.Provider], item.ChapterID)
		}
	}

	data := MediaDownloadData{
		Downloaded: downloads,
		Queued:     qm,
	}

	return data, nil

}

func (d *Downloader) refreshMediaMap() {
	d.mediaMapMu.Lock()
	defer d.mediaMapMu.Unlock()

	ret := make(MediaMap)

	_ = os.MkdirAll(d.downloadDir, os.ModePerm)
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

	d.mediaMap = &ret
}
