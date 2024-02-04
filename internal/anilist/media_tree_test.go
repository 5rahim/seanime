package anilist

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/seanime-app/seanime/internal/limiter"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBaseMedia_FetchMediaTree(t *testing.T) {

	anilistClientWrapper := MockAnilistClientWrapper()
	lim := limiter.NewAnilistLimiter()
	baseMediaCache := NewBaseMediaCache()

	tests := []struct {
		name    string
		mediaId int
		treeIds []int
	}{
		{
			name:    "Bungo Stray Dogs",
			mediaId: 103223,
			treeIds: []int{
				21311,  // BSD1
				21679,  // BSD2
				103223, // BSD3
				141249, // BSD4
				163263, // BSD5
			},
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			mediaF, err := anilistClientWrapper.Client.BaseMediaByID(context.Background(), &tt.mediaId)
			if err != nil {
				t.Fatalf("error while fetching media, %v", err)
			}
			media := mediaF.GetMedia()

			tree := NewBaseMediaRelationTree()

			err = media.FetchMediaTree(
				FetchMediaTreeAll,
				anilistClientWrapper,
				lim,
				tree,
				baseMediaCache,
			)

			if assert.NoError(t, err) {

				for _, treeId := range tt.treeIds {
					_, found := tree.Get(treeId)
					assert.Truef(t, found, "expected tree to contain %d", treeId)
				}
			}

		})

	}

}

func TestBasicMedia_FetchMediaTree(t *testing.T) {

	anilistClientWrapper := MockAnilistClientWrapper()
	lim := limiter.NewAnilistLimiter()
	baseMediaCache := NewBaseMediaCache()

	tests := []struct {
		name    string
		mediaId int
		treeIds []int
	}{
		{
			name:    "Bungo Stray Dogs",
			mediaId: 103223,
			treeIds: []int{
				21311,  // BSD1
				21679,  // BSD2
				103223, // BSD3
				141249, // BSD4
				163263, // BSD5
			},
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			mediaF, err := anilistClientWrapper.Client.BasicMediaByID(context.Background(), &tt.mediaId)
			if err != nil {
				t.Fatalf("error while fetching media, %v", err)
			}
			media := mediaF.GetMedia()

			tree := NewBaseMediaRelationTree()

			err = media.FetchMediaTree(
				FetchMediaTreeAll,
				anilistClientWrapper,
				lim,
				tree,
				baseMediaCache,
			)

			if assert.NoError(t, err) {

				for _, treeId := range tt.treeIds {
					a, found := tree.Get(treeId)
					assert.Truef(t, found, "expected tree to contain %d", treeId)
					spew.Dump(a.GetTitleSafe())
				}
			}

		})

	}

}
