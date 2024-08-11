package manga_providers

import (
	hibikemanga "github.com/5rahim/hibike/pkg/extension/manga"
	"github.com/stretchr/testify/assert"
	"seanime/internal/util"
	"testing"
)

func TestMangasee_Search(t *testing.T) {

	tests := []struct {
		name  string
		query string
	}{
		{
			name:  "One Piece",
			query: "One Piece",
		},
		{
			name:  "Jujutsu Kaisen",
			query: "Jujutsu Kaisen",
		},
	}

	mangasee := NewMangasee(util.NewLogger())

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			searchRes, err := mangasee.Search(hibikemanga.SearchOptions{
				Query: tt.query,
			})
			if assert.NoError(t, err, "mangasee.Search() error") {
				assert.NotEmpty(t, searchRes, "search result is empty")

				for _, res := range searchRes {
					t.Logf("Title: %s", res.Title)
					t.Logf("\tID: %s", res.ID)
					t.Logf("\tYear: %d", res.Year)
					t.Logf("\tImage: %s", res.Image)
					t.Logf("\tProvider: %s", res.Provider)
					t.Logf("\tSearchRating: %f", res.SearchRating)
					t.Logf("\tSynonyms: %v", res.Synonyms)
					t.Log("--------------------------------------------------")
				}
			}

		})

	}

}

func TestMangasee_FindChapters(t *testing.T) {

	tests := []struct {
		name    string
		id      string
		atLeast int
	}{
		{
			name:    "One Piece",
			id:      "One-Piece",
			atLeast: 1100,
		},
		{
			name:    "Jujutsu Kaisen",
			id:      "Jujutsu-Kaisen",
			atLeast: 250,
		},
	}

	mangasee := NewMangasee(util.NewLogger())

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			chapters, err := mangasee.FindChapters(tt.id)
			if assert.NoError(t, err, "mangasee.FindChapters() error") {

				assert.NotEmpty(t, chapters, "chapters is empty")

				assert.GreaterOrEqual(t, len(chapters), tt.atLeast, "chapters length is less than expected")

				for _, chapter := range chapters {
					t.Logf("Title: %s", chapter.Title)
					t.Logf("\tSlug: %s", chapter.ID)
					t.Logf("\tURL: %s", chapter.URL)
					t.Logf("\tIndex: %d", chapter.Index)
					t.Logf("\tChapter: %s", chapter.Chapter)
					t.Logf("\tUpdatedAt: %s", chapter.UpdatedAt)
					t.Log("--------------------------------------------------")
				}
			}

		})

	}

}

func TestMangasee_FindChapterPages(t *testing.T) {

	tests := []struct {
		name  string
		id    string
		index uint
	}{
		{
			name:  "One Piece",
			id:    "One-Piece",
			index: 1110,
		},
	}

	mangasee := NewMangasee(util.NewLogger())

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			chapters, err := mangasee.FindChapters(tt.id)
			if assert.NoError(t, err, "mangasee.FindChapters() error") {

				assert.NotEmpty(t, chapters, "chapters is empty")

				var chapterInfo *hibikemanga.ChapterDetails
				for _, chapter := range chapters {
					if chapter.Index == tt.index {
						chapterInfo = chapter
						break
					}
				}

				if assert.NotNil(t, chapterInfo, "chapter not found") {
					pages, err := mangasee.FindChapterPages(chapterInfo.ID)
					if assert.NoError(t, err, "mangasee.FindChapterPages() error") {
						assert.NotEmpty(t, pages, "pages is empty")

						for _, page := range pages {
							t.Logf("Index: %d", page.Index)
							t.Logf("\tURL: %s", page.URL)
							t.Log("--------------------------------------------------")
						}
					}
				}
			}

		})

	}

}
