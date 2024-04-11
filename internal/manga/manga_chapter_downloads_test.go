package manga

import (
	"context"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/test_utils"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/seanime-app/seanime/internal/util/filecache"
	"path/filepath"
	"testing"
)

func TestGetDownloadedChapterContainers(t *testing.T) {
	test_utils.SetTwoLevelDeep()
	test_utils.InitTestProvider(t, test_utils.Anilist())

	anilistClientWrapper := anilist.TestGetMockAnilistClientWrapper()

	mangaCollection, err := anilistClientWrapper.MangaCollection(context.Background(), &test_utils.ConfigData.Provider.AnilistUsername)
	if err != nil {
		t.Fatal(err)
	}

	logger := util.NewLogger()
	fileCacher, err := filecache.NewCacher(filepath.Join(test_utils.ConfigData.Path.DataDir, "cache"))
	if err != nil {
		t.Fatal(err)
	}

	repository := NewRepository(&NewRepositoryOptions{
		Logger:         logger,
		FileCacher:     fileCacher,
		BackupDir:      "",
		ServerURI:      "",
		WsEventManager: events.NewMockWSEventManager(logger),
		DownloadDir:    filepath.Join(test_utils.ConfigData.Path.DataDir, "manga"),
	})

	// Test
	containers, err := repository.GetDownloadedChapterContainers(mangaCollection)
	if err != nil {
		t.Fatal(err)
	}

	for _, container := range containers {
		t.Logf("MediaId: %d", container.MediaId)
		t.Logf("Provider: %s", container.Provider)
		t.Logf("Chapters: ")
		for _, chapter := range container.Chapters {
			t.Logf("  %s", chapter.Title)
		}
		t.Log("-----------------------------------")
		t.Log("")
	}

}
