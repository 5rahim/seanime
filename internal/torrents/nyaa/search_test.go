package nyaa

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/platform"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/seanime-app/seanime/internal/util/limiter"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSearch(t *testing.T) {

	res, err := Search(SearchOptions{
		Provider: "nyaa",
		Query:    "one piece",
		Category: "anime-eng",
		SortBy:   "seeders",
		Filter:   "",
	})

	if err != nil {
		t.Fatal(err)
	}

	for _, torrent := range res {
		t.Log(torrent)
	}
}

func TestNsfwSearch(t *testing.T) {

	res, err := Search(SearchOptions{
		Provider: "sukebei",
		Query:    "Araiya-san!: Ore to Aitsu ga Onnayu de!?",
		Category: "art-anime",
		SortBy:   "seeders",
		Filter:   "",
	})

	if err != nil {
		t.Fatal(err)
	}

	for _, torrent := range res {
		t.Logf("Title: %s", torrent.Name)
		t.Logf("Seeders: %s", torrent.Seeders)
	}
}

func TestBuildSearchQuery(t *testing.T) {

	anilistLimiter := limiter.NewAnilistLimiter()
	anilistClient := anilist.TestGetMockAnilistClient()
	anilistPlatform := platform.NewAnilistPlatform(anilistClient, util.NewLogger())

	tests := []struct {
		name           string
		mediaId        int
		batch          bool
		episodeNumber  int
		absoluteOffset int
		resolution     string
		title          *string
	}{
		{
			name:           "ReZero kara Hajimeru Isekai Seikatsu 2nd-Season",
			batch:          false,
			mediaId:        108632,
			episodeNumber:  1,
			absoluteOffset: 24,
			resolution:     "",
			title:          nil,
		},
	}

	for _, tt := range tests {

		anilistLimiter.Wait()

		t.Run(tt.name, func(t *testing.T) {

			media, err := anilistPlatform.GetAnime(tt.mediaId)

			if assert.NoError(t, err) &&
				assert.NotNil(t, media) {

				queries, ok := BuildSearchQuery(&BuildSearchQueryOptions{
					Media:          media,
					Batch:          lo.ToPtr(tt.batch),
					EpisodeNumber:  lo.ToPtr(tt.episodeNumber),
					AbsoluteOffset: lo.ToPtr(tt.absoluteOffset),
					Resolution:     lo.ToPtr(tt.resolution),
					Title:          tt.title,
				})

				if assert.True(t, ok) {

					res, err := SearchMultiple(SearchMultipleOptions{
						Provider: "nyaa",
						Query:    queries,
						Category: "anime-eng",
						SortBy:   "seeders",
						Filter:   "",
					})
					if assert.NoError(t, err, "error searching nyaa") {
						for _, torrent := range res {
							t.Log(spew.Sdump(torrent.Name))
						}
					}
				}

			}

		})

	}

}
