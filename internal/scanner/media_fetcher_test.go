package scanner

import (
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/anizip"
	"github.com/seanime-app/seanime-server/internal/util"
	"testing"
)

func TestFetchMediaFromLocalFiles(t *testing.T) {

	anilistClient := MockGetAnilistClient()
	anizipCache := anizip.NewCache()
	baseMediaCache := anilist.NewBaseMediaCache()

	localFiles, ok := MockGetTestLocalFiles()
	if !ok {
		t.Fatal("expected local files, got error")
	}

	_, ok = FetchMediaFromLocalFiles(anilistClient, localFiles, baseMediaCache, anizipCache)

	if !ok {
		t.Fatal("expected result, got error")
	}

	baseMediaCache.Range(func(key int, value *anilist.BaseMedia) bool {
		t.Log(value.GetTitleSafe())
		return true
	})

}

func TestNewMediaFetcher(t *testing.T) {

	anilistClient := MockGetAnilistClient()
	anizipCache := anizip.NewCache()
	baseMediaCache := anilist.NewBaseMediaCache()
	logger := util.NewLogger()

	localFiles, ok := MockGetTestLocalFiles()
	if !ok {
		t.Fatal("expected local files, got error")
	}

	mc, err := NewMediaFetcher(&MediaFetcherOptions{
		Enhanced:       false,
		Username:       "5unwired",
		AnilistClient:  anilistClient,
		LocalFiles:     localFiles,
		BaseMediaCache: baseMediaCache,
		AnizipCache:    anizipCache,
		Logger:         logger,
	})

	if err != nil {
		t.Fatal("expected result, got error: ", err)
	}

	for _, media := range mc.AllMedia {
		t.Log(media.GetTitleSafe())
	}
}

func TestEnhancedNewMediaFetcher(t *testing.T) {

	anilistClient := MockGetAnilistClient()
	anizipCache := anizip.NewCache()
	baseMediaCache := anilist.NewBaseMediaCache()
	logger := util.NewLogger()

	localFiles, ok := MockGetTestLocalFiles()
	if !ok {
		t.Fatal("expected local files, got error")
	}

	mc, err := NewMediaFetcher(&MediaFetcherOptions{
		Enhanced:       true,
		Username:       "5unwired",
		AnilistClient:  anilistClient,
		LocalFiles:     localFiles,
		BaseMediaCache: baseMediaCache,
		AnizipCache:    anizipCache,
		Logger:         logger,
	})

	if err != nil {
		t.Fatal("expected result, got error: ", err)
	}

	for _, media := range mc.AllMedia {
		t.Log(media.GetTitleSafe())
	}
}
