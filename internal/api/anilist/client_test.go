package anilist

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"testing"
)

//func TestHiddenFromStatus(t *testing.T) {
//	test_utils.InitTestProvider(t, test_utils.Anilist())
//
//	token := test_utils.ConfigData.Provider.AnilistJwt
//	logger := util.NewLogger()
//	//anilistClient := NewAnilistClient(test_utils.ConfigData.Provider.AnilistJwt)
//
//	variables := map[string]interface{}{}
//
//	variables["userName"] = test_utils.ConfigData.Provider.AnilistUsername
//	variables["type"] = "ANIME"
//
//	requestBody, err := json.Marshal(map[string]interface{}{
//		"query":     testQuery,
//		"variables": variables,
//	})
//	require.NoError(t, err)
//
//	data, err := customQuery(requestBody, logger, token)
//	require.NoError(t, err)
//
//	var mediaLists []*MediaList
//
//	type retData struct {
//		Page     Page
//		PageInfo PageInfo
//	}
//
//	var ret retData
//	m, err := json.Marshal(data)
//	require.NoError(t, err)
//	if err := json.Unmarshal(m, &ret); err != nil {
//		t.Fatalf("Failed to unmarshal data: %v", err)
//	}
//
//	mediaLists = append(mediaLists, ret.Page.MediaList...)
//
//	util.Spew(ret.Page.PageInfo)
//
//	var currentPage = 1
//	var hasNextPage = false
//	if ret.Page.PageInfo != nil && ret.Page.PageInfo.HasNextPage != nil {
//		hasNextPage = *ret.Page.PageInfo.HasNextPage
//	}
//	for hasNextPage {
//		currentPage++
//		variables["page"] = currentPage
//		requestBody, err = json.Marshal(map[string]interface{}{
//			"query":     testQuery,
//			"variables": variables,
//		})
//		require.NoError(t, err)
//		data, err = customQuery(requestBody, logger, token)
//		require.NoError(t, err)
//		m, err = json.Marshal(data)
//		require.NoError(t, err)
//		if err := json.Unmarshal(m, &ret); err != nil {
//			t.Fatalf("Failed to unmarshal data: %v", err)
//		}
//		util.Spew(ret.Page.PageInfo)
//		if ret.Page.PageInfo != nil && ret.Page.PageInfo.HasNextPage != nil {
//			hasNextPage = *ret.Page.PageInfo.HasNextPage
//		}
//		mediaLists = append(mediaLists, ret.Page.MediaList...)
//	}
//
//	//res, err := anilistClient.AnimeCollection(context.Background(), &test_utils.ConfigData.Provider.AnilistUsername)
//	//assert.NoError(t, err)
//
//	for _, mediaList := range mediaLists {
//		util.Spew(mediaList.Media.ID)
//		if mediaList.Media.ID == 151514 {
//			util.Spew(mediaList)
//		}
//	}
//
//}
//
//const testQuery = `query ($page: Int, $userName: String, $type: MediaType) {
//      Page (page: $page, perPage: 100) {
//        pageInfo {
//          hasNextPage
//		  total
//		  perPage
//		  currentPage
//		  lastPage
//        }
//        mediaList (type: $type, userName: $userName) {
//          status
//          startedAt {
//            year
//            month
//            day
//          }
//          completedAt {
//            year
//            month
//            day
//          }
//          repeat
//          score(format: POINT_100)
//          progress
//          progressVolumes
//          notes
//          media {
//            siteUrl
//            id
//            idMal
//            episodes
//            chapters
//            volumes
//            status
//            averageScore
//            coverImage{
//              large
//              extraLarge
//            }
//            bannerImage
//            title {
//              userPreferred
//            }
//          }
//        }
//      }
//    }`

func TestGetAnimeById(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	anilistClient := TestGetMockAnilistClient()

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
			res, err := anilistClient.BaseAnimeByID(context.Background(), &tt.mediaId)
			assert.NoError(t, err)
			assert.NotNil(t, res)
		})
	}
}

func TestListAnime(t *testing.T) {
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

			cacheKey := ListAnimeCacheKey(
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

			res, err := ListAnimeM(
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
				"",
			)
			assert.NoError(t, err)

			assert.Equal(t, *tt.PerPage, len(res.GetPage().GetMedia()))

			spew.Dump(res)
		})
	}
}
