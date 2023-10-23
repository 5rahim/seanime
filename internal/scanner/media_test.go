package scanner

import (
	"bytes"
	"context"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/anizip"
	"strconv"
	"testing"
)

func TestFetchMediaTrees(t *testing.T) {

	anilistClient := MockGetAnilistClient()
	localFiles, ok := MockGetTestLocalFiles()
	anizipCache := anizip.NewCache()
	baseMediaCache := anilist.NewBaseMediaCache()

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

type ChunkQuery struct {
	T0 struct {
		*anilist.BaseMedia
	}
	T1 struct {
		*anilist.BaseMedia
	}
	//T2 struct {
	//	*anilist.BaseMedia
	//}
}

func TestChunkedQuery(t *testing.T) {

	ac := MockGetAnilistClient()

	var resp ChunkQuery
	var vars map[string]interface{}

	ids := []int{1, 21}

	var query bytes.Buffer
	query.WriteString("query AnimeByMalId { ")
	lo.ForEach(ids, func(item int, index int) {
		_id := strconv.Itoa(item)
		_idx := strconv.Itoa(index)
		query.WriteString("t" + _idx + `: Media(id: ` + _id + `, type: ANIME) {
                id
				idMal
				siteUrl
				status(version: 2)
				season
				type
				format
				bannerImage
				episodes
				synonyms
				isAdult
				countryOfOrigin
				title {
					userPreferred
					romaji
					english
					native
				}
				coverImage {
					extraLarge
					large
					medium
					color
				}
				startDate {
					year
					month
					day
				}
				nextAiringEpisode {
					airingAt
					timeUntilAiring
					episode
				}
				relations {
					edges {
						relationType(version: 2)
						node {
							id
							idMal
							siteUrl
							status(version: 2)
							season
							type
							format
							bannerImage
							episodes
							synonyms
							isAdult
							countryOfOrigin
							title {
								userPreferred
								romaji
								english
								native
							}
							coverImage {
								extraLarge
								large
								medium
								color
							}
							startDate {
								year
								month
								day
							}
							nextAiringEpisode {
								airingAt
								timeUntilAiring
								episode
							}
						}
					}
				}
            }`)
	})
	query.WriteString("}")

	err := ac.Client.Post(context.Background(), "AnimeByMalId", query.String(), &resp, vars)

	if err != nil {
		t.Fatal("expected result, got error: ", err)
	}

	t.Log(resp)

}
