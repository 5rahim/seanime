package scanner

import (
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/anizip"
	"github.com/seanime-app/seanime/internal/entities"
	"github.com/seanime-app/seanime/internal/limiter"
	"github.com/seanime-app/seanime/internal/util"
	"testing"
)

func TestFetchMediaFromLocalFiles(t *testing.T) {

	anilistClient := anilist.MockGetAnilistClient()
	anizipCache := anizip.NewCache()
	baseMediaCache := anilist.NewBaseMediaCache()
	anilistRateLimiter := limiter.NewAnilistLimiter()

	localFiles, ok := entities.MockGetLocalFiles()
	if !ok {
		t.Fatal("expected local files, got error")
	}

	_, ok = FetchMediaFromLocalFiles(anilistClient, localFiles, baseMediaCache, anizipCache, anilistRateLimiter)

	if !ok {
		t.Fatal("expected result, got error")
	}

	baseMediaCache.Range(func(key int, value *anilist.BaseMedia) bool {
		t.Log(value.GetTitleSafe())
		return true
	})

}

func TestNewMediaFetcher(t *testing.T) {

	anilistClient := anilist.MockGetAnilistClient()
	anizipCache := anizip.NewCache()
	baseMediaCache := anilist.NewBaseMediaCache()
	logger := util.NewLogger()
	anilistRateLimiter := limiter.NewAnilistLimiter()

	localFiles, ok := entities.MockGetLocalFiles()
	if !ok {
		t.Fatal("expected local files, got error")
	}

	mc, err := NewMediaFetcher(&MediaFetcherOptions{
		Enhanced:           false,
		Username:           "5unwired",
		AnilistClient:      anilistClient,
		LocalFiles:         localFiles,
		BaseMediaCache:     baseMediaCache,
		AnizipCache:        anizipCache,
		Logger:             logger,
		AnilistRateLimiter: anilistRateLimiter,
	})

	if err != nil {
		t.Fatal("expected result, got error: ", err)
	}

	for _, media := range mc.AllMedia {
		t.Log(media.GetTitleSafe())
	}
}

func TestEnhancedNewMediaFetcher(t *testing.T) {

	anilistClient := anilist.MockGetAnilistClient()
	anizipCache := anizip.NewCache()
	baseMediaCache := anilist.NewBaseMediaCache()
	logger := util.NewLogger()
	anilistRateLimiter := limiter.NewAnilistLimiter()

	localFiles, ok := entities.MockGetLocalFiles()
	if !ok {
		t.Fatal("expected local files, got error")
	}

	mc, err := NewMediaFetcher(&MediaFetcherOptions{
		Enhanced:           true,
		Username:           "5unwired",
		AnilistClient:      anilistClient,
		LocalFiles:         localFiles,
		BaseMediaCache:     baseMediaCache,
		AnizipCache:        anizipCache,
		Logger:             logger,
		AnilistRateLimiter: anilistRateLimiter,
	})

	if err != nil {
		t.Fatal("expected result, got error: ", err)
	}

	for _, media := range mc.AllMedia {
		t.Log(media.GetTitleSafe())
	}
}
