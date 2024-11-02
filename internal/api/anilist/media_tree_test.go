package anilist

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"seanime/internal/test_utils"
	"seanime/internal/util/limiter"
	"testing"
)

func TestBaseAnime_FetchMediaTree_BaseAnime(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	anilistClient := TestGetMockAnilistClient()
	lim := limiter.NewAnilistLimiter()
	completeAnimeCache := NewCompleteAnimeCache()

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
		{
			name:    "Re:Zero",
			mediaId: 21355,
			edgeIds: []int{
				21355,  // Re:Zero 1
				108632, // Re:Zero 2
				119661, // Re:Zero 2 Part 2
				163134, // Re:Zero 3
			},
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			mediaF, err := anilistClient.CompleteAnimeByID(context.Background(), &tt.mediaId)

			if assert.NoError(t, err) {

				media := mediaF.GetMedia()

				tree := NewCompleteAnimeRelationTree()

				err = media.FetchMediaTree(
					FetchMediaTreeAll,
					anilistClient,
					lim,
					tree,
					completeAnimeCache,
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
