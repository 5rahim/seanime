package anilist

import (
	"context"
	"github.com/seanime-app/seanime/internal/limiter"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBaseMedia_FetchMediaTree(t *testing.T) {

	anilistClient := NewAuthedClient("")

	id := 103223 // BSD3
	bsdMediaF, err := anilistClient.BaseMediaByID(context.Background(), &id)

	if err != nil {
		t.Fatalf("error while fetching media")
	}

	bsdMedia := bsdMediaF.GetMedia()

	tree := NewBaseMediaRelationTree()

	err = bsdMedia.FetchMediaTree(
		FetchMediaTreeAll,
		anilistClient,
		limiter.NewAnilistLimiter(),
		tree,
		NewBaseMediaCache())

	if assert.NoError(t, err) {
		tree.Range(func(key int, value *BaseMedia) bool {
			t.Log(value.GetTitleSafe())
			return true
		})
	}

}
