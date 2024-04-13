package manga

import (
	"bytes"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/manga/providers"
	"github.com/seanime-app/seanime/internal/util/filecache"
	_ "golang.org/x/image/bmp"  // Register BMP format
	_ "golang.org/x/image/tiff" // Register Tiff format
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
		logger         *zerolog.Logger
		fileCacher     *filecache.Cacher
		comick         *manga_providers.ComicK
		mangasee       *manga_providers.Mangasee
		mangadex       *manga_providers.Mangadex
		mangapill      *manga_providers.Mangapill
		serverUri      string
		wsEventManager events.IWSEventManager
		mu             sync.Mutex
		downloadDir    string
	}

	NewRepositoryOptions struct {
		Logger         *zerolog.Logger
		FileCacher     *filecache.Cacher
		BackupDir      string
		ServerURI      string
		WsEventManager events.IWSEventManager
		DownloadDir    string
	}
)

func NewRepository(opts *NewRepositoryOptions) *Repository {
	r := &Repository{
		logger:         opts.Logger,
		fileCacher:     opts.FileCacher,
		comick:         manga_providers.NewComicK(opts.Logger),
		mangasee:       manga_providers.NewMangasee(opts.Logger),
		mangadex:       manga_providers.NewMangadex(opts.Logger),
		mangapill:      manga_providers.NewMangapill(opts.Logger),
		serverUri:      opts.ServerURI,
		wsEventManager: opts.WsEventManager,
		downloadDir:    opts.DownloadDir,
	}
	return r
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

func getImageNaturalSizeB(data []byte) (int, int, error) {
	// Decode the image
	img, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return 0, 0, err
	}

	// Return the natural size
	return img.Width, img.Height, nil
}
