package manga

import (
	"errors"
	_ "image/jpeg" // Register JPEG format
	_ "image/png"  // Register PNG format
	"io"
	"net/http"
	"seanime/internal/database/db"
	"seanime/internal/database/models"
	"seanime/internal/events"
	"seanime/internal/extension"
	"seanime/internal/util"
	"seanime/internal/util/filecache"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	_ "golang.org/x/image/bmp"  // Register BMP format
	_ "golang.org/x/image/tiff" // Register Tiff format
	_ "golang.org/x/image/webp" // Register WebP format
)

var (
	ErrNoResults            = errors.New("no results found for this media")
	ErrNoChapters           = errors.New("no manga chapters found")
	ErrChapterNotFound      = errors.New("chapter not found")
	ErrChapterNotDownloaded = errors.New("chapter not downloaded")
	ErrNoTitlesProvided     = errors.New("no titles provided")
)

type (
	Repository struct {
		logger           *zerolog.Logger
		fileCacher       *filecache.Cacher
		cacheDir         string
		extensionBankRef *util.Ref[*extension.UnifiedBank]
		serverUri        string
		wsEventManager   events.WSEventManagerInterface
		mu               sync.Mutex
		preferencesMu    sync.Mutex
		sourceRefreshMu  sync.Mutex
		sourceRefresh    *mangaSourceRefreshState
		sourceRefreshLog map[string]mangaSourceRefreshCompleted
		downloadDir      string
		db               *db.Database

		settings *models.Settings
	}

	NewRepositoryOptions struct {
		Logger           *zerolog.Logger
		CacheDir         string
		FileCacher       *filecache.Cacher
		ServerURI        string
		WsEventManager   events.WSEventManagerInterface
		DownloadDir      string
		Database         *db.Database
		ExtensionBankRef *util.Ref[*extension.UnifiedBank]
	}
)

func NewRepository(opts *NewRepositoryOptions) *Repository {
	r := &Repository{
		logger:           opts.Logger,
		fileCacher:       opts.FileCacher,
		cacheDir:         opts.CacheDir,
		serverUri:        opts.ServerURI,
		wsEventManager:   opts.WsEventManager,
		downloadDir:      opts.DownloadDir,
		extensionBankRef: opts.ExtensionBankRef,
		db:               opts.Database,
		sourceRefreshLog: make(map[string]mangaSourceRefreshCompleted),
	}
	return r
}

func (r *Repository) SetSettings(settings *models.Settings) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.settings = settings
}

func (r *Repository) RemoveProvider(id string) {
	r.extensionBankRef.Get().Delete(id)
}

func (r *Repository) GetProviderExtensionBank() *extension.UnifiedBank {
	return r.extensionBankRef.Get()
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// File Cache
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type bucketType string

const (
	bucketTypeChapterKey                = "1"
	bucketTypeChapter        bucketType = "chapters"
	bucketTypePage           bucketType = "pages"
	bucketTypePageDimensions bucketType = "page-dimensions"
)

// getFcProviderBucket returns a bucket for the provider and mediaId.
//
//	e.g., manga_comick_chapters_123, manga_mangasee_pages_456
//
// Note: Each bucket contains only 1 key-value pair.
func (r *Repository) getFcProviderBucket(provider string, mediaId int, bucketType bucketType) filecache.Bucket {
	return filecache.NewBucket("manga_"+provider+"_"+string(bucketType)+"_"+strconv.Itoa(mediaId), time.Hour*24*7)
}

// EmptyMangaCache deletes all manga buckets associated with the given mediaId.
func (r *Repository) EmptyMangaCache(mediaId int) (err error) {
	// Empty the manga cache
	err = r.fileCacher.RemoveAllBy(func(filename string) bool {
		return strings.HasPrefix(filename, "manga_") && strings.Contains(filename, strconv.Itoa(mediaId))
	})

	_ = r.fileCacher.Clear()
	return
}

func ParseChapterContainerFileName(filename string) (provider string, bucketType bucketType, mediaId int, ok bool) {
	filename = strings.TrimSuffix(filename, ".json")
	filename = strings.TrimSuffix(filename, ".cache")
	filename = strings.TrimSuffix(filename, ".txt")
	parts := strings.Split(filename, "_")
	if len(parts) != 4 {
		return "", "", 0, false
	}

	provider = parts[1]
	var err error
	mediaId, err = strconv.Atoi(parts[3])
	if err != nil {
		return "", "", 0, false
	}

	switch parts[2] {
	case "chapters":
		bucketType = bucketTypeChapter
	case "pages":
		bucketType = bucketTypePage
	case "page-dimensions":
		bucketType = bucketTypePageDimensions
	default:
		return "", "", 0, false
	}

	ok = true
	return
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func getImageNaturalSize(url string) (int, int, error) {
	// Fetch the image
	resp, err := http.Get(url)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, err
	}

	return getImageNaturalSizeB(buf)
}

func getImageNaturalSizeB(data []byte) (int, int, error) {
	width, height, _, err := util.DetectImageFormatAndDimensions(data, "")
	if err != nil {
		return 0, 0, err
	}

	// Return the natural size
	return width, height, nil
}
