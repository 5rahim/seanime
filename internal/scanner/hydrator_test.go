package scanner

import (
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/anizip"
	"github.com/seanime-app/seanime-server/internal/entities"
	"github.com/seanime-app/seanime-server/internal/limiter"
	"github.com/seanime-app/seanime-server/internal/util"
	"testing"
)

func TestFileHydrator_HydrateMetadata(t *testing.T) {

	media := anilist.MockGetAllMedia()
	baseMediaCache := anilist.NewBaseMediaCache()
	anizipCache := anizip.NewCache()
	aniliztClient := anilist.MockGetAnilistClient()
	anilistRateLimiter := limiter.NewAnilistLimiter()
	logger := util.NewLogger()

	localFiles, ok := entities.MockGetSelectedLocalFiles()
	if !ok {
		t.Fatal("expected local files, got error")
	}

	mc := NewMediaContainer(&MediaContainerOptions{
		allMedia: *media,
	})

	matcher := &Matcher{
		localFiles:     localFiles,
		mediaContainer: mc,
		baseMediaCache: baseMediaCache,
		logger:         logger,
	}
	if err := matcher.MatchLocalFilesWithMedia(); err != nil {
		t.Fatal("expected result, got error:", err.Error())
	}

	fh := FileHydrator{
		localFiles:         localFiles,
		media:              *media,
		baseMediaCache:     baseMediaCache,
		anizipCache:        anizipCache,
		anilistClient:      aniliztClient,
		anilistRateLimiter: anilistRateLimiter,
		logger:             logger,
	}

	fh.HydrateMetadata()

	for _, lf := range fh.localFiles {
		if lf == nil {
			t.Fatal("expected base media, got nil")
		}
		t.Logf("LocalFile: %+v\nMetadata: %+v\n\n", lf, lf.Metadata)
	}

}
