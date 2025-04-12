package manga_providers

import (
	"github.com/stretchr/testify/assert"
	hibikemanga "seanime/internal/extension/hibike/manga"
	"seanime/internal/util"
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
		{
			name:  "Komi-san wa, Komyushou desu",
			query: "Komi-san wa, Komyushou desu",
		},
		{
			name:  "Boku no Kokoro no Yabai Yatsu",
			query: "Boku no Kokoro no Yabai Yatsu",
		},
	}

	comick := NewComicK(util.NewLogger())

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			searchRes, err := comick.Search(hibikemanga.SearchOptions{
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
		{
			name:    "Komi-san wa, Komyushou desu",
			id:      "K_Dn8VW7",
			atLeast: 250,
		},
		{
			name:    "Boku no Kokoro no Yabai Yatsu",
			id:      "pYN47sZm",
			atLeast: 141,
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

func TestComicKMulti_FindChapters(t *testing.T) {

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
		{
			name:    "Komi-san wa, Komyushou desu",
			id:      "K_Dn8VW7",
			atLeast: 250,
		},
		{
			name:    "Boku no Kokoro no Yabai Yatsu",
			id:      "pYN47sZm",
			atLeast: 141,
		},
	}

	comick := NewComicKMulti(util.NewLogger())

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			chapters, err := comick.FindChapters(tt.id)
			if assert.NoError(t, err, "comick.FindChapters() error") {

				assert.NotEmpty(t, chapters, "chapters is empty")

				assert.GreaterOrEqual(t, len(chapters), tt.atLeast, "chapters length is less than expected")

				for _, chapter := range chapters {
					t.Logf("Title: %s", chapter.Title)
					t.Logf("\tLanguage: %s", chapter.Language)
					t.Logf("\tScanlator: %s", chapter.Scanlator)
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

				var chapterInfo *hibikemanga.ChapterDetails
				for _, chapter := range chapters {
					if chapter.Index == tt.index {
						chapterInfo = chapter
						break
					}
				}

				if assert.NotNil(t, chapterInfo, "chapter not found") {
					pages, err := comick.FindChapterPages(chapterInfo.ID)
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
