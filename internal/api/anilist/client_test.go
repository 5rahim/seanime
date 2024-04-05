package anilist

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/test_utils"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetBaseMediaById(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	anilistClientWrapper := TestGetMockAnilistClientWrapper()

	tests := []struct {
		name    string
		mediaId int
	}{
		{
			name:    "Cowboy Bebop",
			mediaId: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := anilistClientWrapper.BaseMediaByID(context.Background(), &tt.mediaId)
			assert.NoError(t, err)
			assert.NotNil(t, res)
		})
	}
}

func TestListMedia(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	tests := []struct {
		name                string
		Page                *int
		Search              *string
		PerPage             *int
		Sort                []*MediaSort
		Status              []*MediaStatus
		Genres              []*string
		AverageScoreGreater *int
		Season              *MediaSeason
		SeasonYear          *int
		Format              *MediaFormat
		IsAdult             *bool
	}{
		{
			name:                "Popular",
			Page:                lo.ToPtr(1),
			Search:              nil,
			PerPage:             lo.ToPtr(20),
			Sort:                []*MediaSort{lo.ToPtr(MediaSortTrendingDesc)},
			Status:              nil,
			Genres:              nil,
			AverageScoreGreater: nil,
			Season:              nil,
			SeasonYear:          nil,
			Format:              nil,
			IsAdult:             nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			cacheKey := ListMediaCacheKey(
				tt.Page,
				tt.Search,
				tt.PerPage,
				tt.Sort,
				tt.Status,
				tt.Genres,
				tt.AverageScoreGreater,
				tt.Season,
				tt.SeasonYear,
				tt.Format,
				tt.IsAdult,
			)

			t.Log(cacheKey)

			res, err := ListMediaM(
				tt.Page,
				tt.Search,
				tt.PerPage,
				tt.Sort,
				tt.Status,
				tt.Genres,
				tt.AverageScoreGreater,
				tt.Season,
				tt.SeasonYear,
				tt.Format,
				tt.IsAdult,
				util.NewLogger(),
			)
			assert.NoError(t, err)

			assert.Equal(t, *tt.PerPage, len(res.GetPage().GetMedia()))

			spew.Dump(res)
		})
	}
}
