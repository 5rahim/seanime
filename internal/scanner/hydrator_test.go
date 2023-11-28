package scanner

import (
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/anizip"
	"github.com/seanime-app/seanime/internal/entities"
	"github.com/seanime-app/seanime/internal/limiter"
	"github.com/seanime-app/seanime/internal/util"
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
		LocalFiles:         localFiles,
		Media:              *media,
		BaseMediaCache:     baseMediaCache,
		AnizipCache:        anizipCache,
		AnilistClient:      aniliztClient,
		AnilistRateLimiter: anilistRateLimiter,
		Logger:             logger,
	}

	fh.HydrateMetadata()

	for _, lf := range fh.LocalFiles {
		if lf == nil {
			t.Fatal("expected base media, got nil")
		}
		t.Logf("LocalFile: %+v\nMetadata: %+v\n\n", lf, lf.Metadata)
	}

}
