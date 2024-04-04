package manga

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/manga/providers"
	"github.com/seanime-app/seanime/internal/util/filecache"
	_ "golang.org/x/image/webp" // Register WebP format
	"image"
	_ "image/jpeg" // Register JPEG format
	_ "image/png"  // Register PNG format
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type (
	Repository struct {
		logger           *zerolog.Logger
		fileCacher       *filecache.Cacher
		comick           *manga_providers.ComicK
		mangasee         *manga_providers.Mangasee
		downloader       *downloader
		backupDir        string
		serverUri        string
		backupMap        DownloadMap
		wsEventManager   events.IWSEventManager
		mu               sync.Mutex
		downloadContexts map[DownloadID]context.CancelFunc
	}

	NewRepositoryOptions struct {
		Logger         *zerolog.Logger
		FileCacher     *filecache.Cacher
		BackupDir      string
		ServerURI      string
		WsEventManager events.IWSEventManager
	}
)

func NewRepository(opts *NewRepositoryOptions) *Repository {
	r := &Repository{
		logger:           opts.Logger,
		fileCacher:       opts.FileCacher,
		comick:           manga_providers.NewComicK(opts.Logger),
		mangasee:         manga_providers.NewMangasee(opts.Logger),
		downloader:       newDownloader(opts.Logger, opts.WsEventManager),
		backupDir:        opts.BackupDir,
		serverUri:        opts.ServerURI,
		backupMap:        make(DownloadMap),
		downloadContexts: make(map[DownloadID]context.CancelFunc),
	}

	//go r.hydrateBackupMap()

	return r
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Backups
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type EntryBackupContainer struct {
	Provider   string          `json:"provider"`
	MediaId    int             `json:"mediaId"`
	ChapterIds map[string]bool `json:"chapterIds"` // Using map for O(1) lookup in the client
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// File Cache
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type bucketType string

const (
	bucketTypeChapter        bucketType = "chapters"
	bucketTypePage           bucketType = "pages"
	bucketTypePageDimensions bucketType = "page-dimensions"
)

// getFcProviderBucket returns a bucket for the provider and mediaId.
//
//	e.g., manga_comick_chapters_123, manga_mangasee_pages_456
func (r *Repository) getFcProviderBucket(provider manga_providers.Provider, mediaId int, bucketType bucketType) filecache.Bucket {
	return filecache.NewBucket("manga_"+string(provider)+"_"+string(bucketType)+"_"+strconv.Itoa(mediaId), time.Hour*24*7)
}

// EmptyMangaCache deletes all manga buckets associated with the given mediaId.
func (r *Repository) EmptyMangaCache(mediaId int) (err error) {
	// Empty the manga cache
	err = r.fileCacher.RemoveAllBy(func(filename string) bool {
		return strings.HasPrefix(filename, "manga_") && strings.Contains(filename, strconv.Itoa(mediaId))
	})
	return
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Backups
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// GetMangaEntryBackups returns the backup chapters for the given manga entry.
// Used by the client to display the downloaded chapters / allow user to download chapters.
// Never returns nil.
func (r *Repository) GetMangaEntryBackups(provider manga_providers.Provider, mediaId int) *EntryBackupContainer {

	// Get the backup chapters for the given manga entry
	backupContainer := &EntryBackupContainer{
		ChapterIds: make(map[string]bool),
		Provider:   string(provider),
		MediaId:    mediaId,
	}

	if r.backupMap == nil {
		return backupContainer
	}

	storedChapterIds, found := r.backupMap[DownloadID{Provider: string(provider), MediaID: mediaId}]
	if !found {
		return backupContainer
	}

	for _, chapterId := range storedChapterIds {
		backupContainer.ChapterIds[chapterId] = true
	}

	return backupContainer
}

func (r *Repository) hydrateBackupMap() {
	// Get the backup folders
	backupMap, err := r.downloader.getDownloads(r.backupDir)
	if err != nil {
		r.logger.Error().Err(err).Msg("manga: failed to hydrate backup map")
		return
	}

	// Set the backup map
	r.backupMap = backupMap
}
func (r *Repository) GetStoredChapterIdsFromBackup(c DownloadID) ([]string, bool, error) {
	if r.backupMap == nil {
		return nil, false, nil
	}

	storedChapterIds, found := r.backupMap[c]
	return storedChapterIds, found, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func getImageNaturalSize(url string) (int, int, error) {
	// Fetch the image
	resp, err := http.Get(url)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	// Decode the image
	img, _, err := image.DecodeConfig(resp.Body)
	if err != nil {
		return 0, 0, err
	}

	// Return the natural size
	return img.Width, img.Height, nil
}
