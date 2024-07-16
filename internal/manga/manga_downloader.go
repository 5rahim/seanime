package manga

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/api/anilist"
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
		wsEventManager    events.WSEventManagerInterface
		database          *db.Database
		downloadDir       string
		chapterDownloader *chapter_downloader.Downloader
		repository        *Repository

		mediaMap   *MediaMap // Refreshed on start and after each download
		mediaMapMu sync.RWMutex

		chapterDownloadedCh chan chapter_downloader.DownloadID
		readingDownloadDir  bool
	}

	// MediaMap is created after reading the download directory.
	// It is used to store all downloaded chapters for each media.
	// The key is the media ID and the value is a map of provider to a list of chapters.
	//
	//	e.g., downloadDir/comick_1234_abc_13/
	//	      downloadDir/comick_1234_def_13.5/
	// -> { 1234: { "comick": [ { "chapterId": "abc", "chapterNumber": "13" }, { "chapterId": "def", "chapterNumber": "13.5" } ] } }
	MediaMap map[int]ProviderDownloadMap

	// ProviderDownloadMap is used to store all downloaded chapters for a specific media and provider.
	ProviderDownloadMap map[string][]ProviderDownloadMapChapterInfo

	ProviderDownloadMapChapterInfo struct {
		ChapterID     string `json:"chapterId"`
		ChapterNumber string `json:"chapterNumber"`
	}

	MediaDownloadData struct {
		Downloaded ProviderDownloadMap `json:"downloaded"`
		Queued     ProviderDownloadMap `json:"queued"`
	}
)

type (
	NewDownloaderOptions struct {
		Database       *db.Database
		Logger         *zerolog.Logger
		WSEventManager events.WSEventManagerInterface
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

	_ = os.MkdirAll(d.downloadDir, os.ModePerm)

	go d.refreshMediaMap()

	return d
}

// Start is called once to start the Chapter downloader 's main goroutine.
func (d *Downloader) Start() {
	d.chapterDownloader.Start()
	go func() {
		for {
			select {
			// Refresh the media map when a chapter is downloaded
			case _ = <-d.chapterDownloader.ChapterDownloaded():
				d.refreshMediaMap()
			}
		}
	}()
}

// DownloadChapter is called by the client to download a chapter.
// It fetches the chapter pages by using Repository.GetMangaPageContainer
// and invokes the chapter_downloader.Downloader 'Download' method to add the chapter to the download queue.
func (d *Downloader) DownloadChapter(opts DownloadChapterOptions) error {

	// Find chapter container in the file cache
	// e.g. comick$1234 from bucket 'manga_comick_chapters_1234'
	// Note: Each bucket contains only 1 key-value pair.
	chapterKey := fmt.Sprintf("%s$%d", opts.Provider, opts.MediaId)
	chapterBucket := d.repository.getFcProviderBucket(opts.Provider, opts.MediaId, bucketTypeChapter)
	var chapterContainer *ChapterContainer
	// Get the only key-value pair in the bucket
	if found, _ := d.repository.fileCacher.Get(chapterBucket, chapterKey, &chapterContainer); !found {
		// If the chapter container is not found, return an error
		// since it means that it wasn't fetched (for some reason) -- This shouldn't happen
		return errors.New("chapters not found")
	}

	// Find the chapter in the chapter container
	// e.g. Wind-Breaker$0062
	chapter, ok := chapterContainer.GetChapter(opts.ChapterId)
	if !ok {
		return errors.New("chapter not found")
	}

	// Fetch the chapter pages
	pageContainer, err := d.repository.GetMangaPageContainer(opts.Provider, opts.MediaId, opts.ChapterId, false, false)
	if err != nil {
		return err
	}

	// Add the chapter to the download queue
	return d.chapterDownloader.AddToQueue(chapter_downloader.DownloadOptions{
		DownloadID: chapter_downloader.DownloadID{
			Provider:      string(opts.Provider),
			MediaId:       opts.MediaId,
			ChapterId:     opts.ChapterId,
			ChapterNumber: chapter.GetNormalizedChapter(),
		},
		Pages: pageContainer.Pages,
	})
}

// DeleteChapter is called by the client to delete a downloaded chapter.
func (d *Downloader) DeleteChapter(provider string, mediaId int, chapterId string, chapterNumber string) (err error) {
	err = d.chapterDownloader.DeleteChapter(chapter_downloader.DownloadID{
		Provider:      provider,
		MediaId:       mediaId,
		ChapterId:     chapterId,
		ChapterNumber: chapterNumber,
	})
	if err != nil {
		return err
	}

	d.refreshMediaMap()

	return nil
}

// DeleteChapters is called by the client to delete downloaded chapters.
func (d *Downloader) DeleteChapters(ids []chapter_downloader.DownloadID) (err error) {
	for _, id := range ids {
		err = d.chapterDownloader.DeleteChapter(chapter_downloader.DownloadID{
			Provider:      id.Provider,
			MediaId:       id.MediaId,
			ChapterId:     id.ChapterId,
			ChapterNumber: id.ChapterNumber,
		})
	}
	if err != nil {
		return err
	}

	d.refreshMediaMap()

	return nil
}

func (d *Downloader) GetMediaDownloads(mediaId int, cached bool) (MediaDownloadData, error) {
	if !cached {
		d.refreshMediaMap()
	}

	return d.mediaMap.getMediaDownload(mediaId, d.database)
}

func (d *Downloader) RefreshMediaMap() *MediaMap {
	d.refreshMediaMap()
	return d.mediaMap
}

func (d *Downloader) GetAllMediaDownloads() *MediaMap {
	return d.mediaMap
}

func (d *Downloader) RunChapterDownloadQueue() {
	d.chapterDownloader.Run()
}

func (d *Downloader) StopChapterDownloadQueue() {
	_ = d.database.ResetDownloadingChapterDownloadQueueItems()
	d.chapterDownloader.Stop()
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type (
	NewDownloadListOptions struct {
		MangaCollection *anilist.MangaCollection
	}

	DownloadListItem struct {
		MediaId      int                 `json:"mediaId"`
		Media        *anilist.BaseManga  `json:"media"`
		DownloadData ProviderDownloadMap `json:"downloadData"`
	}
)

func (d *Downloader) NewDownloadList(opts *NewDownloadListOptions) (ret []*DownloadListItem, err error) {

	mm := d.mediaMap

	ret = make([]*DownloadListItem, 0)

	for mId, data := range *mm {
		listEntry, ok := opts.MangaCollection.GetListEntryFromMediaId(mId)
		if !ok {
			ret = append(ret, &DownloadListItem{
				MediaId:      mId,
				Media:        nil,
				DownloadData: data,
			})
			continue
		}

		media := listEntry.GetMedia()
		if media == nil {
			ret = append(ret, &DownloadListItem{
				MediaId:      mId,
				Media:        nil,
				DownloadData: data,
			})
			continue
		}

		item := &DownloadListItem{
			MediaId:      mId,
			Media:        media,
			DownloadData: data,
		}

		ret = append(ret, item)
	}

	return
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
		downloads = make(map[string][]ProviderDownloadMapChapterInfo)
	}

	// Get all queued chapters for the media
	queued, err := db.GetMediaQueuedChapters(mediaId)
	if err != nil {
		queued = make([]*models.ChapterDownloadQueueItem, 0)
	}

	qm := make(ProviderDownloadMap)
	for _, item := range queued {
		if _, ok := qm[item.Provider]; !ok {
			qm[item.Provider] = []ProviderDownloadMapChapterInfo{
				{
					ChapterID:     item.ChapterID,
					ChapterNumber: item.ChapterNumber,
				},
			}
		} else {
			qm[item.Provider] = append(qm[item.Provider], ProviderDownloadMapChapterInfo{
				ChapterID:     item.ChapterID,
				ChapterNumber: item.ChapterNumber,
			})
		}
	}

	data := MediaDownloadData{
		Downloaded: downloads,
		Queued:     qm,
	}

	return data, nil

}

func (d *Downloader) refreshMediaMap() {

	if d.readingDownloadDir {
		return
	}

	d.mediaMapMu.Lock()
	defer d.mediaMapMu.Unlock()

	d.readingDownloadDir = true
	defer func() {
		d.readingDownloadDir = false
	}()

	d.logger.Debug().Msg("manga downloader: Reading download directory")

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
				parts := strings.SplitN(file.Name(), "_", 4)
				if len(parts) != 4 {
					return
				}

				provider := parts[0]
				mediaID, err := strconv.Atoi(parts[1])
				chapterID := parts[2]
				chapterNumber := parts[3]

				if err != nil {
					return
				}

				mu.Lock()
				if _, ok := ret[mediaID]; !ok {
					ret[mediaID] = make(map[string][]ProviderDownloadMapChapterInfo)
					ret[mediaID][provider] = []ProviderDownloadMapChapterInfo{
						{
							ChapterID:     chapterID,
							ChapterNumber: chapterNumber,
						},
					}
				} else {
					if _, ok := ret[mediaID][provider]; !ok {
						ret[mediaID][provider] = []ProviderDownloadMapChapterInfo{
							{
								ChapterID:     chapterID,
								ChapterNumber: chapterNumber,
							},
						}
					} else {
						ret[mediaID][provider] = append(ret[mediaID][provider], ProviderDownloadMapChapterInfo{
							ChapterID:     chapterID,
							ChapterNumber: chapterNumber,
						})
					}
				}
				mu.Unlock()
			}
		}(file)
	}
	wg.Wait()

	d.mediaMap = &ret

	// When done refreshing, send a message to the client to refetch the download data
	d.wsEventManager.SendEvent(events.RefreshedMangaDownloadData, nil)
}
