package manga_providers

import (
	"github.com/seanime-app/seanime/internal/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMangadex_Search(t *testing.T) {

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

	mangadex := NewMangadex(util.NewLogger())

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			searchRes, err := mangadex.Search(SearchOptions{
				Query: tt.query,
			})
			if assert.NoError(t, err, "mangadex.Search() error") {
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

func TestMangadex_FindChapters(t *testing.T) {

	tests := []struct {
		name    string
		id      string
		atLeast int
	}{
		//{
		//	name:    "One Piece",
		//	id:      "One-Piece",
		//	atLeast: 1100,
		//},
		{
			name:    "Jujutsu Kaisen",
			id:      "c52b2ce3-7f95-469c-96b0-479524fb7a1a",
			atLeast: 250,
		},
	}

	mangadex := NewMangadex(util.NewLogger())

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			chapters, err := mangadex.FindChapters(tt.id)
			if assert.NoError(t, err, "mangadex.FindChapters() error") {

				assert.NotEmpty(t, chapters, "chapters is empty")

				assert.GreaterOrEqual(t, len(chapters), tt.atLeast, "chapters length is less than expected")

				for _, chapter := range chapters {
					t.Logf("Title: %s", chapter.Title)
					t.Logf("\tSlug: %s", chapter.ID)
					t.Logf("\tURL: %s", chapter.URL)
					t.Logf("\tIndex: %d", chapter.Index)
					t.Logf("\tUpdatedAt: %s", chapter.UpdatedAt)
					t.Log("--------------------------------------------------")
				}
			}

		})

	}

}

//
//func TestMangadex_FindChapterPages(t *testing.T) {
//
//	tests := []struct {
//		name          string
//		id            string
//		chapterNumber int
//	}{
//		{
//			name:          "One Piece",
//			id:            "One-Piece",
//			chapterNumber: 1111,
//		},
//	}
//
//	mangadex := NewMangadex(util.NewLogger())
//
//	for _, tt := range tests {
//
//		t.Run(tt.name, func(t *testing.T) {
//
//			chapters, err := mangadex.FindChapters(tt.id)
//			if assert.NoError(t, err, "mangadex.FindChapters() error") {
//
//				assert.NotEmpty(t, chapters, "chapters is empty")
//
//				var chapterInfo *ChapterDetails
//				for _, chapter := range chapters {
//					if chapter.Number == tt.chapterNumber {
//						chapterInfo = chapter
//						break
//					}
//				}
//
//				if assert.NotNil(t, chapterInfo, "chapter not found") {
//					pages, err := mangadex.FindChapterPages(chapterInfo)
//					if assert.NoError(t, err, "mangadex.FindChapterPages() error") {
//						assert.NotEmpty(t, pages, "pages is empty")
//
//						spew.Dump(pages)
//					}
//				}
//			}
//
//		})
//
//	}
//
//}
