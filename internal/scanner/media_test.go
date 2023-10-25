package scanner

import (
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/anizip"
	"testing"
)

func TestFetchMediaTrees(t *testing.T) {

	anilistClient := MockGetAnilistClient()
	anizipCache := anizip.NewCache()
	baseMediaCache := anilist.NewBaseMediaCache()

	localFiles, ok := MockGetTestLocalFiles()
	if !ok {
		t.Fatal("expected local files, got error")
	}

	ret, ok := FetchMediaTrees(anilistClient, localFiles, baseMediaCache, anizipCache)

	if !ok {
		t.Fatal("expected result, got error")
	}

	for _, media := range ret {
		t.Log(*media.GetTitleSafe())
	}

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
		t.Log(*media.GetTitleSafe())
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
		t.Log(*media.GetTitleSafe())
	}
}
