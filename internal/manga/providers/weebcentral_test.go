package manga_providers

import (
	"seanime/internal/util"
	"testing"

	"github.com/stretchr/testify/assert"
	hibikemanga "seanime/internal/extension/hibike/manga"
)

func TestWeebCentral_Search(t *testing.T) {

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

	weebcentral := NewWeebCentral(util.NewLogger())

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			searchRes, err := weebcentral.Search(hibikemanga.SearchOptions{
				Query: tt.query,
			})
			if assert.NoError(t, err, "weebcentral.Search() error") {
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

func TestWeebCentral_FindChapters(t *testing.T) {

	tests := []struct {
		name    string
		id      string
		atLeast int
	}{
		{
			name:    "One Piece",
			id:      "01J76XY7E9FNDZ1DBBM6PBJPFK",
			atLeast: 1100,
		},
		{
			name:    "Jujutsu Kaisen",
			id:      "01J76XYCERXE60T7FKXVCCAQ0H",
			atLeast: 250,
		},
	}

	weebcentral := NewWeebCentral(util.NewLogger())

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			chapters, err := weebcentral.FindChapters(tt.id)
			if assert.NoError(t, err, "weebcentral.FindChapters() error") {

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

func TestWeebCentral_FindChapterPages(t *testing.T) {

	tests := []struct {
		name  string
		id    string
		index uint
	}{
		{
			name:  "One Piece",
			id:    "01J76XY7E9FNDZ1DBBM6PBJPFK",
			index: 1110,
		},
		{
			name:  "Jujutsu Kaisen",
			id:    "01J76XYCERXE60T7FKXVCCAQ0H",
			index: 0,
		},
	}

	weebcentral := NewWeebCentral(util.NewLogger())

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			chapters, err := weebcentral.FindChapters(tt.id)
			if assert.NoError(t, err, "weebcentral.FindChapters() error") {

				assert.NotEmpty(t, chapters, "chapters is empty")

				var chapterInfo *hibikemanga.ChapterDetails
				for _, chapter := range chapters {
					if chapter.Index == tt.index {
						chapterInfo = chapter
						break
					}
				}

				if assert.NotNil(t, chapterInfo, "chapter not found") {
					pages, err := weebcentral.FindChapterPages(chapterInfo.ID)
					if assert.NoError(t, err, "weebcentral.FindChapterPages() error") {
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
