package manga

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/manga/providers"
	"github.com/seanime-app/seanime/internal/util/filecache"
	"image"
	_ "image/jpeg" // Register JPEG format
	_ "image/png"  // Register PNG format
	"net/http"
	"strconv"
	"time"
)

type (
	Repository struct {
		logger         *zerolog.Logger
		fileCacher     *filecache.Cacher
		comick         *manga_providers.ComicK
		mangasee       *manga_providers.Mangasee
		downloader     *downloader
		backupDir      string
		serverUri      string
		backupMap      BackupMap
		wsEventManager events.IWSEventManager
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
		logger:     opts.Logger,
		fileCacher: opts.FileCacher,
		comick:     manga_providers.NewComicK(opts.Logger),
		mangasee:   manga_providers.NewMangasee(opts.Logger),
		downloader: newDownloader(opts.Logger, opts.WsEventManager),
		backupDir:  opts.BackupDir,
		serverUri:  opts.ServerURI,
		backupMap:  make(BackupMap),
	}

	r.hydrateBackupMap()

	return r
}

// RefreshBackups is a client action.
func (r *Repository) RefreshBackups() {
	r.hydrateBackupMap()
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) hydrateBackupMap() {
	go func() {
		// Get the backup folders
		backupMap, err := r.downloader.getBackups(r.backupDir)
		if err != nil {
			//r.logger.Error().Err(err).Msg("manga: failed to hydrate backup map")
			return
		}

		// Set the backup map
		r.backupMap = backupMap
	}()
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type bucketType string

const (
	bucketTypeChapter bucketType = "chapters"
	bucketTypePage    bucketType = "pages"
)

// getFcProviderBucket returns a bucket for the provider and mediaId.
//
//	e.g., manga_comick_chapters_123, manga_mangasee_pages_456
func (r *Repository) getFcProviderBucket(provider manga_providers.Provider, mediaId int, bucketType bucketType) filecache.Bucket {
	return filecache.NewBucket("manga"+"_"+string(provider)+"_"+string(bucketType)+"_"+strconv.Itoa(mediaId), time.Hour*24*7)
}

func getImageNaturalSize(url string) (int, int, error) {
	// Fetch the image
	resp, err := http.Head(url) // FIXME Use HEAD to avoid downloading the entire image, this only works for ComicK
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	// Extract image dimensions from the Content-Length header
	contentLength := resp.Header.Get("Content-Length")
	if contentLength == "" {
		return 0, 0, fmt.Errorf("Content-Length header not found")
	}

	// Decode the image
	img, _, err := image.DecodeConfig(resp.Body)
	if err != nil {
		return 0, 0, err
	}

	// Return the natural size
	return img.Width, img.Height, nil
}
