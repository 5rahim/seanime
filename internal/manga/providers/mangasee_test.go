package manga_providers

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/stretchr/testify/assert"
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

			searchRes, err := mangasee.Search(SearchOptions{
				Query: tt.query,
			})
			if assert.NoError(t, err, "mangasee.Search() error") {
				assert.NotEmpty(t, searchRes, "search result is empty")

				spew.Dump(searchRes)
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
			}

		})

	}

}

func TestMangasee_FindChapterPages(t *testing.T) {

	tests := []struct {
		name          string
		id            string
		chapterNumber int
	}{
		{
			name:          "One Piece",
			id:            "One-Piece",
			chapterNumber: 1111,
		},
	}

	mangasee := NewMangasee(util.NewLogger())

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			chapters, err := mangasee.FindChapters(tt.id)
			if assert.NoError(t, err, "mangasee.FindChapters() error") {

				assert.NotEmpty(t, chapters, "chapters is empty")

				var chapterInfo *ChapterDetails
				for _, chapter := range chapters {
					if chapter.Number == tt.chapterNumber {
						chapterInfo = chapter
						break
					}
				}

				if assert.NotNil(t, chapterInfo, "chapter not found") {
					pages, err := mangasee.FindChapterPages(chapterInfo)
					if assert.NoError(t, err, "mangasee.FindChapterPages() error") {
						assert.NotEmpty(t, pages, "pages is empty")

						spew.Dump(pages)
					}
				}
			}

		})

	}

}
