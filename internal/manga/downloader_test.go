package manga

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/manga/providers"
	"github.com/seanime-app/seanime/internal/test_utils"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDownloader(t *testing.T) {
	test_utils.SetTwoLevelDeep()
	test_utils.InitTestProvider(t)

	cacheDir := t.TempDir()

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

	logger := util.NewLogger()
	d := newDownloader(logger, events.NewMockWSEventManager(logger))

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
					if assert.NoError(t, err, "comick.FindChapterPages() error") {
						assert.NotEmpty(t, pages, "pages is empty")

						//
						// TEST
						//

						ctx, cancel := context.WithCancel(context.Background())
						defer cancel()
						err := d.downloadImages(ctx, string(tt.providerName), tt.mediaId, chapterInfo.ID, pages, cacheDir)
						assert.NoError(t, err, "downloadImages() error")

						backupMap, err := d.getDownloads(cacheDir)
						assert.NoError(t, err, "getDownloads() error")

						assert.NotEmpty(t, backupMap, "backupMap is empty")
						spew.Dump(backupMap)

						pageMap, _, err := d.getPageMap(string(tt.providerName), tt.mediaId, chapterInfo.ID, cacheDir)
						assert.NoError(t, err, "getPageMap() error")

						assert.NotEmpty(t, pageMap, "pageMap is empty")

						for _, image := range *pageMap {
							t.Logf("Index: %d", image.Index)
							t.Logf("\tURL: %s", image.Filename)
							t.Log("--------------------------------------------------")
						}

						err = d.deleteDownloads(string(tt.providerName), tt.mediaId, chapterInfo.ID, cacheDir)
						assert.NoError(t, err, "deleteDownloads() error")

					}
				}
			}

		})

	}

}
