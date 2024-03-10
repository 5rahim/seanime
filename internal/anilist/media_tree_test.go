package anilist

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/seanime-app/seanime/internal/limiter"
	"github.com/seanime-app/seanime/internal/test_utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBaseMedia_FetchMediaTree(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	acw := TestGetAnilistClientWrapper()
	lim := limiter.NewAnilistLimiter()
	baseMediaCache := NewBaseMediaCache()

	tests := []struct {
		name    string
		mediaId int
		edgeIds []int
	}{
		{
			name:    "Bungo Stray Dogs",
			mediaId: 103223,
			edgeIds: []int{
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

			mediaF, err := acw.Client.BaseMediaByID(context.Background(), &tt.mediaId)

			if assert.NoError(t, err) {

				media := mediaF.GetMedia()

				tree := NewBaseMediaRelationTree()

				err = media.FetchMediaTree(
					FetchMediaTreeAll,
					acw,
					lim,
					tree,
					baseMediaCache,
				)

				if assert.NoError(t, err) {

					for _, treeId := range tt.edgeIds {
						a, found := tree.Get(treeId)
						assert.Truef(t, found, "expected tree to contain %d", treeId)
						spew.Dump(a.GetTitleSafe())
					}

				}

			}
		})

	}

}

func TestBasicMedia_FetchMediaTree(t *testing.T) {

	acw := TestGetAnilistClientWrapper()
	lim := limiter.NewAnilistLimiter()
	baseMediaCache := NewBaseMediaCache()

	tests := []struct {
		name    string
		mediaId int
		edgeIds []int
	}{
		{
			name:    "Bungo Stray Dogs",
			mediaId: 103223,
			edgeIds: []int{
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

			mediaF, err := acw.Client.BasicMediaByID(context.Background(), &tt.mediaId)

			if assert.NoError(t, err) {

				media := mediaF.GetMedia()

				tree := NewBaseMediaRelationTree()

				err = media.FetchMediaTree(
					FetchMediaTreeAll,
					acw,
					lim,
					tree,
					baseMediaCache,
				)

				if assert.NoError(t, err) {

					for _, treeId := range tt.edgeIds {
						a, found := tree.Get(treeId)
						assert.Truef(t, found, "expected tree to contain %d", treeId)
						spew.Dump(a.GetTitleSafe())
					}
				}

			}

		})

	}

}
