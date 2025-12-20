package manga

import (
	"context"
	"seanime/internal/api/anilist"
	"seanime/internal/database/db"
	"seanime/internal/extension"
	"seanime/internal/platforms/anilist_platform"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"testing"
)

func TestNewCollection(t *testing.T) {
	test_utils.SetTwoLevelDeep()
	test_utils.InitTestProvider(t, test_utils.Anilist())

	db, err := db.NewDatabase(test_utils.ConfigData.Path.DataDir, test_utils.ConfigData.Database.Name, util.NewLogger())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	anilistClient := anilist.TestGetMockAnilistClient()
	logger := util.NewLogger()
	extensionBankRef := util.NewRef(extension.NewUnifiedBank())
	anilistPlatform := anilist_platform.NewAnilistPlatform(util.NewRef(anilistClient), extensionBankRef, logger, db)

	mangaCollection, err := anilistClient.MangaCollection(context.Background(), &test_utils.ConfigData.Provider.AnilistUsername)
	if err != nil {
		t.Fatalf("Failed to get manga collection: %v", err)
	}

	opts := &NewCollectionOptions{
		MangaCollection: mangaCollection,
		PlatformRef:     util.NewRef(anilistPlatform),
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
