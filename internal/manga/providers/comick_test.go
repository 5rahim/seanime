package manga_providers

import (
	"github.com/seanime-app/seanime/internal/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestComicK_Search(t *testing.T) {

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

	comick := NewComicK(util.NewLogger())

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			searchRes, err := comick.Search(SearchOptions{
				Query: tt.query,
			})
			if assert.NoError(t, err, "comick.Search() error") {
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

func TestComicK_FindChapters(t *testing.T) {

	tests := []struct {
		name    string
		id      string
		atLeast int
	}{
		{
			name:    "Jujutsu Kaisen",
			id:      "TA22I5O7",
			atLeast: 250,
		},
	}

	comick := NewComicK(util.NewLogger())

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			chapters, err := comick.FindChapters(tt.id)
			if assert.NoError(t, err, "comick.FindChapters() error") {

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

func TestComicK_FindChapterPages(t *testing.T) {

	tests := []struct {
		name  string
		id    string
		index uint
	}{
		{
			name:  "Jujutsu Kaisen",
			id:    "TA22I5O7",
			index: 258,
		},
	}

	comick := NewComicK(util.NewLogger())

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			chapters, err := comick.FindChapters(tt.id)
			if assert.NoError(t, err, "comick.FindChapters() error") {

				assert.NotEmpty(t, chapters, "chapters is empty")

				var chapterInfo *ChapterDetails
				for _, chapter := range chapters {
					if chapter.Index == tt.index {
						chapterInfo = chapter
						break
					}
				}

				if assert.NotNil(t, chapterInfo, "chapter not found") {
					pages, err := comick.FindChapterPages(chapterInfo)
					if assert.NoError(t, err, "comick.FindChapterPages() error") {
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
