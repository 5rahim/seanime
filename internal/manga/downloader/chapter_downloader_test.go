package chapter_downloader

import (
	"github.com/seanime-app/seanime/internal/database/db"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/manga/providers"
	"github.com/seanime-app/seanime/internal/test_utils"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestQueue(t *testing.T) {
	test_utils.InitTestProvider(t)

	tempDir := t.TempDir()

	logger := util.NewLogger()
	database, err := db.NewDatabase(tempDir, test_utils.ConfigData.Database.Name, logger)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	t.Log(tempDir)

	downloadDir := "./test"

	tests := []struct {
		name         string
		providerName manga_providers.Provider
		provider     manga_providers.MangaProvider
		mangaId      string
		mediaId      int
		chapterIndex uint
	}{
		{
			providerName: manga_providers.ComickProvider,
			provider:     manga_providers.NewComicK(util.NewLogger()),
			name:         "Jujutsu Kaisen",
			mangaId:      "TA22I5O7",
			chapterIndex: 258,
			mediaId:      101517,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			// SETUP
			chapters, err := tt.provider.FindChapters(tt.mangaId)
			if assert.NoError(t, err, "comick.FindChapters() error") {

				assert.NotEmpty(t, chapters, "chapters is empty")

				var chapterInfo *manga_providers.ChapterDetails
				for _, chapter := range chapters {
					if chapter.Index == tt.chapterIndex {
						chapterInfo = chapter
						break
					}
				}

				if assert.NotNil(t, chapterInfo, "chapter not found") {
					pages, err := tt.provider.FindChapterPages(chapterInfo.ID)
					if assert.NoError(t, err, "provider.FindChapterPages() error") {
						assert.NotEmpty(t, pages, "pages is empty")

						//
						// TEST
						//

						downloader := NewDownloader(&NewDownloaderOptions{
							Logger:         logger,
							WSEventManager: events.NewMockWSEventManager(logger),
							Database:       database,
							DownloadDir:    downloadDir,
						})

						downloader.Start()

						err := downloader.DownloadChapter(string(tt.providerName), tt.mediaId, chapterInfo.ID, pages)
						if err != nil {
							t.Fatalf("Failed to download chapter: %v", err)
						}

						time.Sleep(10 * time.Second)

					}
				}
			}

		})

	}

}
