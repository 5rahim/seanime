package handlers

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/anilist"
	"testing"
)

// DOES NOT WORK
func TestHandleSearchAnilistMediaList(t *testing.T) {

	anilistClient := anilist.MockGetAnilistClient()

	type params struct {
		Page                *int                   `json:"page,omitempty"`
		Search              *string                `json:"search,omitempty"`
		PerPage             *int                   `json:"perPage,omitempty"`
		Sort                []*anilist.MediaSort   `json:"sort,omitempty"`
		Status              []*anilist.MediaStatus `json:"status,omitempty"`
		Genres              []*string              `json:"genres,omitempty"`
		AverageScoreGreater *int                   `json:"averageScoreGreater,omitempty"`
		Season              *anilist.MediaSeason   `json:"season,omitempty"`
		SeasonYear          *int                   `json:"seasonYear,omitempty"`
		Format              *anilist.MediaFormat   `json:"format,omitempty"`
	}

	b := new(params)

	b.Page = lo.ToPtr(1)
	b.PerPage = lo.ToPtr(10)
	b.Sort = []*anilist.MediaSort{lo.ToPtr(anilist.MediaSortPopularityDesc)}

	// Function to set default values for nil fields in params
	if b.Page == nil {
		b.Page = lo.ToPtr(1)
	}
	if b.PerPage == nil {
		b.PerPage = lo.ToPtr(10)
	}
	if b.Sort == nil {
		b.Sort = lo.ToSlicePtr(anilist.AllMediaSort)
	}
	if b.Status == nil {
		b.Status = lo.ToSlicePtr(anilist.AllMediaStatus)
	}
	if b.AverageScoreGreater == nil {
		var defaultValue int
		b.AverageScoreGreater = &defaultValue
	}

	res, err := anilistClient.ListMedia(
		context.Background(),
		b.Page,
		b.Search,
		b.PerPage,
		b.Sort,
		b.Status,
		nil,
		b.AverageScoreGreater,
		b.Season,
		nil,
		b.Format,
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(spew.Sprint(res))

}
