package seadex

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/test_utils"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSeaDex(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	anilistClientWrapper := anilist.TestGetMockAnilistClientWrapper()

	tests := []struct {
		name    string
		mediaId int
	}{
		{
			name:    "86 - Eighty Six Part 2",
			mediaId: 131586,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mediaF, err := anilistClientWrapper.BaseAnimeByID(context.Background(), &tt.mediaId)
			if assert.NoErrorf(t, err, "error getting media: %v", tt.mediaId) {

				media := mediaF.GetMedia()

				torrents, err := New(util.NewLogger()).FetchTorrents(tt.mediaId, media.GetRomajiTitleSafe())
				if assert.NoErrorf(t, err, "error fetching records: %v", tt.mediaId) {

					spew.Dump(torrents)

				}

			}

		})
	}

}
