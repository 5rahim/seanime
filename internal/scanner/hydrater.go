package scanner

import (
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/anizip"
)

type FileHydrater struct {
	localFiles     []*LocalFile
	media          []*anilist.BaseMedia
	baseMediaCache *anilist.BaseMediaCache
	anizipCache    *anizip.Cache
}

// HydrateMetadata will hydrate the metadata of each LocalFile with the metadata of the matched anilist.BaseMedia
func (fh *FileHydrater) HydrateMetadata() {
	//rateLimiter := limiter.NewLimiter(5*time.Second, 20)
}
