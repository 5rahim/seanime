package manga

import (
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/manga/providers"
	"github.com/seanime-app/seanime/internal/util/filecache"
	"time"
)

type (
	Repository struct {
		logger                           *zerolog.Logger
		fileCacher                       *filecache.Cacher
		comick                           *manga_providers.ComicK
		mangasee                         *manga_providers.Mangasee
		fcComicKChapterContainerBucket   filecache.Bucket
		fcComicKPageContainerBucket      filecache.Bucket
		fcMangaseeChapterContainerBucket filecache.Bucket
		fcMangaseePageContainerBucket    filecache.Bucket
	}

	NewRepositoryOptions struct {
		Logger     *zerolog.Logger
		FileCacher *filecache.Cacher
	}
)

func NewRepository(opts *NewRepositoryOptions) *Repository {
	return &Repository{
		logger:                           opts.Logger,
		fileCacher:                       opts.FileCacher,
		comick:                           manga_providers.NewComicK(opts.Logger),
		mangasee:                         manga_providers.NewMangasee(opts.Logger),
		fcComicKChapterContainerBucket:   filecache.NewBucket("comick_manga_chapters", time.Hour*24*7),
		fcComicKPageContainerBucket:      filecache.NewBucket("comick_manga_pages", time.Hour*24*7),
		fcMangaseeChapterContainerBucket: filecache.NewBucket("mangasee_manga_chapters", time.Hour*24*7),
		fcMangaseePageContainerBucket:    filecache.NewBucket("mangasee_manga_pages", time.Hour*24*7),
	}
}
