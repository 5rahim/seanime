package manga

import (
	"errors"
	"fmt"
	"os"
	"seanime/internal/api/anilist"
	"seanime/internal/database/db"
	"seanime/internal/database/models"
	"seanime/internal/events"
	"seanime/internal/hook"
	chapter_downloader "seanime/internal/manga/downloader"
	manga_providers "seanime/internal/manga/providers"
	"seanime/internal/util"
	"seanime/internal/util/filecache"
	"sync"

	"github.com/rs/zerolog"
)

type (
	Downloader struct {
		logger            *zerolog.Logger
		wsEventManager    events.WSEventManagerInterface
		database          *db.Database
		downloadDir       string
		chapterDownloader *chapter_downloader.Downloader
		repository        *Repository
		filecacher        *filecache.Cacher

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
	// The key is the provider and the value is a list of chapters.
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
		Provider  string
		MediaId   int
		ChapterId string
		StartNow  bool
	}
)

func NewDownloader(opts *NewDownloaderOptions) *Downloader {
	_ = os.MkdirAll(opts.DownloadDir, os.ModePerm)
	filecacher, _ := filecache.NewCacher(opts.DownloadDir)

	d := &Downloader{
		logger:         opts.Logger,
		wsEventManager: opts.WSEventManager,
		database:       opts.Database,
		downloadDir:    opts.DownloadDir,
		repository:     opts.Repository,
		mediaMap:       new(MediaMap),
		filecacher:     filecacher,
	}

	d.chapterDownloader = chapter_downloader.NewDownloader(&chapter_downloader.NewDownloaderOptions{
		Logger:         opts.Logger,
		WSEventManager: opts.WSEventManager,
		Database:       opts.Database,
		DownloadDir:    opts.DownloadDir,
	})

	go d.hydrateMediaMap()

	return d
}

// Start is called once to start the Chapter downloader 's main goroutine.
func (d *Downloader) Start() {
	d.chapterDownloader.Start()
	go func() {
		for {
			select {
			// Listen for downloaded chapters
			case downloadId := <-d.chapterDownloader.ChapterDownloaded():
				// When a chapter is downloaded, fetch the chapter container from the file cache
				// and store it in the permanent bucket.
				// DEVNOTE: This will be useful to avoid re-fetching the chapter container when the cache expires.
				// This is deleted when a chapter is deleted.
				go func() {
					chapterContainerKey := getMangaChapterContainerCacheKey(downloadId.Provider, downloadId.MediaId)
					chapterContainer, found := d.repository.getChapterContainerFromFilecache(downloadId.Provider, downloadId.MediaId)
					if found {
						// Store the chapter container in the permanent bucket
						permBucket := getPermanentChapterContainerCacheBucket(downloadId.Provider, downloadId.MediaId)
						_ = d.filecacher.SetPerm(permBucket, chapterContainerKey, chapterContainer)
					}
				}()

				// Refresh the media map when a chapter is downloaded
				d.hydrateMediaMap()
			}
		}
	}()
}

// The bucket for storing downloaded chapter containers.
// e.g. manga_downloaded_comick_chapters_1234
// The key is the chapter ID.
func getPermanentChapterContainerCacheBucket(provider string, mId int) filecache.PermanentBucket {
	return filecache.NewPermanentBucket(fmt.Sprintf("manga_downloaded_%s_chapters_%d", provider, mId))
}

// getChapterContainerFromFilecache returns the chapter container from the temporary file cache.
func (r *Repository) getChapterContainerFromFilecache(provider string, mId int) (*ChapterContainer, bool) {
	// Find chapter container in the file cache
	chapterBucket := r.getFcProviderBucket(provider, mId, bucketTypeChapter)

	chapterContainerKey := getMangaChapterContainerCacheKey(provider, mId)

	var chapterContainer *ChapterContainer
	// Get the key-value pair in the bucket
	if found, _ := r.fileCacher.Get(chapterBucket, chapterContainerKey, &chapterContainer); !found {
		// If the chapter container is not found, return an error
		// since it means that it wasn't fetched (for some reason) -- This shouldn't happen
		return nil, false
	}

	return chapterContainer, true
}

// getChapterContainerFromPermanentFilecache returns the chapter container from the permanent file cache.
func (r *Repository) getChapterContainerFromPermanentFilecache(provider string, mId int) (*ChapterContainer, bool) {
	permBucket := getPermanentChapterContainerCacheBucket(provider, mId)

	chapterContainerKey := getMangaChapterContainerCacheKey(provider, mId)

	var chapterContainer *ChapterContainer
	// Get the key-value pair in the bucket
	if found, _ := r.fileCacher.GetPerm(permBucket, chapterContainerKey, &chapterContainer); !found {
		// If the chapter container is not found, return an error
		// since it means that it wasn't fetched (for some reason) -- This shouldn't happen
		return nil, false
	}

	return chapterContainer, true
}

// DownloadChapter is called by the client to download a chapter.
// It fetches the chapter pages by using Repository.GetMangaPageContainer
// and invokes the chapter_downloader.Downloader 'Download' method to add the chapter to the download queue.
func (d *Downloader) DownloadChapter(opts DownloadChapterOptions) error {

	chapterContainer, found := d.repository.getChapterContainerFromFilecache(opts.Provider, opts.MediaId)
	if !found {
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
			Provider:      opts.Provider,
			MediaId:       opts.MediaId,
			ChapterId:     opts.ChapterId,
			ChapterNumber: manga_providers.GetNormalizedChapter(chapter.Chapter),
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

	permBucket := getPermanentChapterContainerCacheBucket(provider, mediaId)
	_ = d.filecacher.DeletePerm(permBucket, chapterId)

	d.hydrateMediaMap()

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

		permBucket := getPermanentChapterContainerCacheBucket(id.Provider, id.MediaId)
		_ = d.filecacher.DeletePerm(permBucket, id.ChapterId)
	}
	if err != nil {
		return err
	}

	d.hydrateMediaMap()

	return nil
}

func (d *Downloader) GetMediaDownloads(mediaId int, cached bool) (ret MediaDownloadData, err error) {
	defer util.HandlePanicInModuleWithError("manga/GetMediaDownloads", &err)

	if !cached {
		d.hydrateMediaMap()
	}

	return d.mediaMap.getMediaDownload(mediaId, d.database)
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
		MediaId int `json:"mediaId"`
		// Media will be nil if the manga is no longer in the user's collection.
		// The client should handle this case by displaying the download data without the media data.
		Media        *anilist.BaseManga  `json:"media"`
		DownloadData ProviderDownloadMap `json:"downloadData"`
	}
)

// NewDownloadList returns a list of DownloadListItem for the client to display.
func (d *Downloader) NewDownloadList(opts *NewDownloadListOptions) (ret []*DownloadListItem, err error) {
	defer util.HandlePanicInModuleWithError("manga/NewDownloadList", &err)

	mm := d.mediaMap

	ret = make([]*DownloadListItem, 0)

	for mId, data := range *mm {
		listEntry, ok := opts.MangaCollection.GetListEntryFromMangaId(mId)
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

// hydrateMediaMap hydrates the MediaMap by reading the download directory.
func (d *Downloader) hydrateMediaMap() {

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

	// Hydrate MediaMap by going through all chapter directories
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	for _, file := range files {
		wg.Add(1)
		go func(file os.DirEntry) {
			defer wg.Done()

			if file.IsDir() {
				// e.g. comick_1234_abc_13.5
				id, ok := chapter_downloader.ParseChapterDirName(file.Name())
				if !ok {
					return
				}

				mu.Lock()
				newMapInfo := ProviderDownloadMapChapterInfo{
					ChapterID:     id.ChapterId,
					ChapterNumber: id.ChapterNumber,
				}

				if _, ok := ret[id.MediaId]; !ok {
					ret[id.MediaId] = make(map[string][]ProviderDownloadMapChapterInfo)
					ret[id.MediaId][id.Provider] = []ProviderDownloadMapChapterInfo{newMapInfo}
				} else {
					if _, ok := ret[id.MediaId][id.Provider]; !ok {
						ret[id.MediaId][id.Provider] = []ProviderDownloadMapChapterInfo{newMapInfo}
					} else {
						ret[id.MediaId][id.Provider] = append(ret[id.MediaId][id.Provider], newMapInfo)
					}
				}
				mu.Unlock()
			}
		}(file)
	}
	wg.Wait()

	// Trigger hook event
	ev := &MangaDownloadMapEvent{
		MediaMap: &ret,
	}
	_ = hook.GlobalHookManager.OnMangaDownloadMap().Trigger(ev) // ignore the error
	// make sure the media map is not nil
	if ev.MediaMap != nil {
		ret = *ev.MediaMap
	}

	d.mediaMap = &ret

	// When done refreshing, send a message to the client to refetch the download data
	d.wsEventManager.SendEvent(events.RefreshedMangaDownloadData, nil)
}
