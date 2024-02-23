package animetosho

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/seanime-app/seanime/internal/anilist"
	"testing"
)

func TestSearchQuery(t *testing.T) {

	_, anilistClientWrapper, _ := anilist.MockAnilistClientWrappers()

	tests := []struct {
		name           string
		mId            int
		batch          bool
		episodeNumber  int
		absoluteOffset int
		resolution     string
	}{
		{
			name:           "Bungou Stray Dogs 5th Season Episode 11",
			mId:            163263,
			batch:          false,
			episodeNumber:  11,
			absoluteOffset: 45,
			resolution:     "1080p",
		},
		{
			name:           "SPY×FAMILY Season 1 Part 2",
			mId:            142838,
			batch:          false,
			episodeNumber:  12,
			absoluteOffset: 12,
			resolution:     "1080p",
		},
		{
			name:           "Jujutsu Kaisen Season 2",
			mId:            145064,
			batch:          false,
			episodeNumber:  2,
			absoluteOffset: 24,
			resolution:     "",
		},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {
			mediaRes, err := anilistClientWrapper.Client.BaseMediaByID(context.Background(), &test.mId)
			if err != nil {
				t.Fatal(err)
			}

			torrents, err := SearchQuery(&BuildSearchQueryOptions{
				Media:          mediaRes.GetMedia(),
				Batch:          &test.batch,
				EpisodeNumber:  &test.episodeNumber,
				AbsoluteOffset: &test.absoluteOffset,
				Resolution:     &test.resolution,
				Cache:          NewSearchCache(),
			})
			if err != nil {
				t.Fatal(err)
			}

			if len(torrents) == 0 {
				t.Fatal("no torrents found")
			}

			spew.Dump(torrents)
		})

	}
}

func TestSearch2(t *testing.T) {
	torrents, err := Search("Kusuriya no Hitorigoto 05")
	if err != nil {
		t.Fatal(err)
	}

	spew.Dump(torrents)
}
