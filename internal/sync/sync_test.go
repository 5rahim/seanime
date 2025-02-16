package sync

import (
	"errors"
	"fmt"
	"seanime/internal/api/anilist"
	"seanime/internal/database/db"
	"seanime/internal/platforms/anilist_platform"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
)

func testSetupManager(t *testing.T) (Manager, *anilist.AnimeCollection, *anilist.MangaCollection) {

	logger := util.NewLogger()

	anilistClient := anilist.NewAnilistClient(test_utils.ConfigData.Provider.AnilistJwt)
	anilistPlatform := anilist_platform.NewAnilistPlatform(anilistClient, logger)
	anilistPlatform.SetUsername(test_utils.ConfigData.Provider.AnilistUsername)
	animeCollection, err := anilistPlatform.GetAnimeCollection(true)
	require.NoError(t, err)
	mangaCollection, err := anilistPlatform.GetMangaCollection(true)
	require.NoError(t, err)

	database, err := db.NewDatabase(test_utils.ConfigData.Path.DataDir, test_utils.ConfigData.Database.Name, logger)
	require.NoError(t, err)

	manager := GetMockManager(t, database)

	manager.SetAnimeCollection(animeCollection)
	manager.SetMangaCollection(mangaCollection)

	return manager, animeCollection, mangaCollection
}

func TestSync2(t *testing.T) {
	test_utils.SetTwoLevelDeep()
	test_utils.InitTestProvider(t, test_utils.Anilist())

	manager, animeCollection, _ := testSetupManager(t)

	err := manager.AddAnime(130003) // Bocchi the rock
	if err != nil && !errors.Is(err, ErrAlreadyTracked) {
		require.NoError(t, err)
	}
	err = manager.AddAnime(10800) // Chihayafuru
	if err != nil && !errors.Is(err, ErrAlreadyTracked) {
		require.NoError(t, err)
	}
	err = manager.AddAnime(171457) // Make Heroine ga Oosugiru!
	if err != nil && !errors.Is(err, ErrAlreadyTracked) {
		require.NoError(t, err)
	}
	err = manager.AddManga(101517) // JJK
	if err != nil && !errors.Is(err, ErrAlreadyTracked) {
		require.NoError(t, err)
	}

	err = manager.SynchronizeLocal()
	require.NoError(t, err)

	select {
	case <-manager.GetQueue().doneUpdatingLocalCollections:
		util.Spew(manager.GetLocalAnimeCollection().MustGet())
		util.Spew(manager.GetLocalMangaCollection().MustGet())
		break
	case <-time.After(10 * time.Second):
		t.Log("Timeout")
		break
	}

	anilist.TestModifyAnimeCollectionEntry(animeCollection, 130003, anilist.TestModifyAnimeCollectionEntryInput{
		Status:   lo.ToPtr(anilist.MediaListStatusCompleted),
		Progress: lo.ToPtr(12), // Mock progress
	})

	fmt.Println("================================================================================================")
	fmt.Println("================================================================================================")

	err = manager.SynchronizeLocal()
	require.NoError(t, err)

	select {
	case <-manager.GetQueue().doneUpdatingLocalCollections:
		util.Spew(manager.GetLocalAnimeCollection().MustGet())
		util.Spew(manager.GetLocalMangaCollection().MustGet())
		break
	case <-time.After(10 * time.Second):
		t.Log("Timeout")
		break
	}

}

func TestSync(t *testing.T) {
	test_utils.SetTwoLevelDeep()
	test_utils.InitTestProvider(t, test_utils.Anilist())

	manager, _, _ := testSetupManager(t)

	err := manager.AddAnime(130003) // Bocchi the rock
	if err != nil && !errors.Is(err, ErrAlreadyTracked) {
		require.NoError(t, err)
	}
	err = manager.AddAnime(10800) // Chihayafuru
	if err != nil && !errors.Is(err, ErrAlreadyTracked) {
		require.NoError(t, err)
	}
	err = manager.AddAnime(171457) // Make Heroine ga Oosugiru!
	if err != nil && !errors.Is(err, ErrAlreadyTracked) {
		require.NoError(t, err)
	}
	err = manager.AddManga(101517) // JJK
	if err != nil && !errors.Is(err, ErrAlreadyTracked) {
		require.NoError(t, err)
	}

	err = manager.SynchronizeLocal()
	require.NoError(t, err)

	select {
	case <-manager.GetQueue().doneUpdatingLocalCollections:
		util.Spew(manager.GetLocalAnimeCollection().MustGet())
		util.Spew(manager.GetLocalMangaCollection().MustGet())
		break
	case <-time.After(10 * time.Second):
		t.Log("Timeout")
		break
	}

}
