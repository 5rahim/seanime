package anilist

import (
	"context"
	"os"
	"seanime/internal/test_utils"
	"testing"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
)

// USE CASE: Generate a boilerplate Anilist AnimeCollection for testing purposes and save it to 'test/data/BoilerplateAnimeCollection'.
// The generated AnimeCollection will have all entries in the 'Planning' status.
// The generated AnimeCollection will be used to test various Anilist API methods.
// You can use TestModifyAnimeCollectionEntry to modify the generated AnimeCollection before using it in a test.
//   - DO NOT RUN IF YOU DON'T PLAN TO GENERATE A NEW 'test/data/BoilerplateAnimeCollection'
func TestGenerateBoilerplateAnimeCollection(t *testing.T) {
	t.Skip("This test is not meant to be run")
	test_utils.InitTestProvider(t, test_utils.Anilist())

	anilistClient := TestGetMockAnilistClient()

	ac, err := anilistClient.AnimeCollection(context.Background(), &test_utils.ConfigData.Provider.AnilistUsername)

	if assert.NoError(t, err) {

		lists := ac.GetMediaListCollection().GetLists()

		entriesToAddToPlanning := make([]*AnimeListEntry, 0)

		if assert.NoError(t, err) {

			for _, list := range lists {
				if list.Status != nil {
					if list.GetStatus().String() != string(MediaListStatusPlanning) {
						entries := list.GetEntries()
						for _, entry := range entries {
							entry.Progress = new(0)
							entry.Score = new(0.0)
							entry.Status = new(MediaListStatusPlanning)
							entriesToAddToPlanning = append(entriesToAddToPlanning, entry)
						}
						list.Entries = make([]*AnimeListEntry, 0)
					}
				}
			}

			newLists := make([]*AnimeCollection_MediaListCollection_Lists, 0)
			for _, list := range lists {
				if list.Status == nil {
					continue
				}
				if *list.GetStatus() == MediaListStatusPlanning {
					list.Entries = append(list.Entries, entriesToAddToPlanning...)
					newLists = append(newLists, list)
				} else {
					newLists = append(newLists, list)
				}
			}

			ac.MediaListCollection.Lists = newLists

			data, err := json.Marshal(ac)
			if assert.NoError(t, err) {
				err = os.WriteFile(test_utils.GetDataPath("BoilerplateAnimeCollection"), data, 0644)
				assert.NoError(t, err)
			}
		}

	}

}
