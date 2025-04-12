package manga_providers

import (
	"github.com/stretchr/testify/assert"
	hibikemanga "seanime/internal/extension/hibike/manga"
	"seanime/internal/util"
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
		{
			name:  "Boku no Kokoro no Yabai Yatsu",
			query: "Boku no Kokoro no Yabai Yatsu",
		},
	}

	mangadex := NewMangadex(util.NewLogger())

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			searchRes, err := mangadex.Search(hibikemanga.SearchOptions{
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
		{
			name:    "The Dangers in My Heart",
			id:      "3df1a9a3-a1be-47a3-9e90-9b3e55b1d0ac",
			atLeast: 141,
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

func TestMangadex_FindChapterPages(t *testing.T) {

	tests := []struct {
		name      string
		id        string
		chapterId string
	}{
		{
			name:      "The Dangers in My Heart",
			id:        "3df1a9a3-a1be-47a3-9e90-9b3e55b1d0ac",
			chapterId: "5145ea39-be4b-4bf9-81e7-4f90961db857", // Chapter 1
		},
		{
			name:      "Kagurabachi",
			id:        "",
			chapterId: "9c9652fc-10d2-40b3-9382-16fb072d3068", // Chapter 1
		},
	}

	mangadex := NewMangadex(util.NewLogger())

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			pages, err := mangadex.FindChapterPages(tt.chapterId)
			if assert.NoError(t, err, "mangadex.FindChapterPages() error") {
				assert.NotEmpty(t, pages, "pages is empty")

				for _, page := range pages {
					t.Logf("Index: %d", page.Index)
					t.Logf("\tURL: %s", page.URL)
					t.Log("--------------------------------------------------")
				}
			}

		})

	}

}
