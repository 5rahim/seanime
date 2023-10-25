package scanner

import (
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/anizip"
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

func TestNewMediaContainer(t *testing.T) {

	anilistClient := MockGetAnilistClient()
	anizipCache := anizip.NewCache()
	baseMediaCache := anilist.NewBaseMediaCache()

	localFiles, ok := MockGetTestLocalFiles()
	if !ok {
		t.Fatal("expected local files, got error")
	}

	mc, err := NewMediaContainer(&MediaContainerOptions{
		Enhanced:       false,
		Username:       "5unwired",
		AnilistClient:  anilistClient,
		LocalFiles:     localFiles,
		BaseMediaCache: baseMediaCache,
		AnizipCache:    anizipCache,
	})

	if err != nil {
		t.Fatal("expected result, got error: ", err)
	}

	for _, media := range mc.AllMedia {
		t.Log(media.GetTitleSafe())
	}
}

func TestEnhancedNewMediaContainer(t *testing.T) {

	anilistClient := MockGetAnilistClient()
	anizipCache := anizip.NewCache()
	baseMediaCache := anilist.NewBaseMediaCache()

	localFiles, ok := MockGetTestLocalFiles()
	if !ok {
		t.Fatal("expected local files, got error")
	}

	mc, err := NewMediaContainer(&MediaContainerOptions{
		Enhanced:       true,
		Username:       "5unwired",
		AnilistClient:  anilistClient,
		LocalFiles:     localFiles,
		BaseMediaCache: baseMediaCache,
		AnizipCache:    anizipCache,
	})

	if err != nil {
		t.Fatal("expected result, got error: ", err)
	}

	for _, media := range mc.AllMedia {
		t.Log(media.GetTitleSafe())
	}
}
