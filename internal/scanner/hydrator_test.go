package scanner

import (
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/anizip"
	"github.com/seanime-app/seanime-server/internal/limiter"
	"testing"
)

func TestFileHydrator_HydrateMetadata(t *testing.T) {

	media := MockAllMedia()
	baseMediaCache := anilist.NewBaseMediaCache()
	anizipCache := anizip.NewCache()
	aniliztClient := MockGetAnilistClient()
	anilistRateLimiter := limiter.NewAnilistLimiter()

	localFiles, ok := MockGetSelectTestLocalFiles()
	if !ok {
		t.Fatal("expected local files, got error")
	}

	mc := NewMediaContainer(&MediaContainerOptions{
		allMedia: *media,
	})

	matcher := NewMatcher(&MatcherOptions{
		localFiles:     localFiles,
		mediaContainer: mc,
		baseMediaCache: baseMediaCache,
	})
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
	}

	fh.HydrateMetadata()

	for _, lf := range fh.localFiles {
		if lf == nil {
			t.Fatal("expected base media, got nil")
		}
		t.Logf("LocalFile: %+v", lf.Metadata)
	}

}
