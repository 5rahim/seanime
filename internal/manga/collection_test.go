package manga

import (
	"context"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/test_utils"
	"testing"
)

func TestNewCollection(t *testing.T) {
	test_utils.SetTwoLevelDeep()
	test_utils.InitTestProvider(t, test_utils.Anilist())

	anilistClientWrapper := anilist.TestGetMockAnilistClientWrapper()

	mangaCollection, err := anilistClientWrapper.MangaCollection(context.Background(), &test_utils.ConfigData.Provider.AnilistUsername)
	if err != nil {
		t.Fatalf("Failed to get manga collection: %v", err)
	}

	opts := &NewCollectionOptions{
		MangaCollection:      mangaCollection,
		AnilistClientWrapper: anilistClientWrapper,
	}

	collection, err := NewCollection(opts)
	if err != nil {
		t.Fatalf("Failed to create collection: %v", err)
	}

	if len(collection.Lists) == 0 {
		t.Skip("No lists found")
	}

	for _, list := range collection.Lists {
		t.Logf("List: %s", list.Type)
		for _, entry := range list.Entries {
			t.Logf("\tEntry: %s", entry.Media.GetPreferredTitle())
			t.Logf("\t\tProgress: %d", entry.EntryListData.Progress)
		}
		t.Log("---------------------------------------")
	}
}
