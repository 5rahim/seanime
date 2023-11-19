package anilist

import (
	"context"
	"github.com/seanime-app/seanime-server/internal/limiter"
	"sync"
	"testing"
	"time"
)

func TestBaseMedia_FetchMediaTree(t *testing.T) {

	anilistClient := NewAuthedClient("")
	cache := NewBaseMediaCache()

	id := 103223  // BSD3
	id2 := 145064 // JJK2
	bsdMediaF, err := anilistClient.BaseMediaByID(context.Background(), &id)
	jjkMediaF, err := anilistClient.BaseMediaByID(context.Background(), &id2)

	if err != nil {
		t.Fatalf("error while fetching media")
	}

	bsdMedia := bsdMediaF.GetMedia()
	jjkMedia := jjkMediaF.GetMedia()

	rateLimit := limiter.NewLimiter(time.Minute, 90)

	tree := NewBaseMediaRelationTree()

	wg := sync.WaitGroup{}

	for _, m := range []*BaseMedia{bsdMedia, jjkMedia} {
		wg.Add(1)
		go func(_m *BaseMedia) {
			defer wg.Done()
			err := _m.FetchMediaTree(FetchMediaTreeAll, anilistClient, rateLimit, tree, cache)
			if err != nil {
				t.Error("error while fetching tree,", err)
				return
			}
		}(m)
	}

	wg.Wait()

	tree.Range(func(key int, value *BaseMedia) bool {
		t.Log(value.GetTitleSafe())
		return true
	})

}
