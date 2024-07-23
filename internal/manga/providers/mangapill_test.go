package manga_providers

import (
	hibikemanga "github.com/5rahim/hibike/pkg/extension/manga"
	"github.com/stretchr/testify/assert"
	"seanime/internal/util"
	"testing"
)

func TestMangapill_Search(t *testing.T) {

	tests := []struct {
		name  string
		query string
	}{
		{
			name:  "Boku no Kokoro no Yabai Yatsu",
			query: "Boku no Kokoro no Yabai Yatsu",
		},
	}

	provider := NewMangapill(util.NewLogger())

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			searchRes, err := provider.Search(hibikemanga.SearchOptions{
				Query: tt.query,
			})
			if assert.NoError(t, err, "provider.Search() error") {
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

func TestMangapill_FindChapters(t *testing.T) {

	tests := []struct {
		name    string
		id      string
		atLeast int
	}{
		{
			name:    "The Dangers in My Heart",
			id:      "5232$boku-no-kokoro-no-yabai-yatsu",
			atLeast: 141,
		},
	}

	provider := NewMangapill(util.NewLogger())

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			chapters, err := provider.FindChapters(tt.id)
			if assert.NoError(t, err, "provider.FindChapters() error") {

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

func TestMangapill_FindChapterPages(t *testing.T) {

	tests := []struct {
		name      string
		chapterId string
	}{
		{
			name:      "The Dangers in My Heart",
			chapterId: "5232-10001000$boku-no-kokoro-no-yabai-yatsu-chapter-1", // Chapter 1
		},
	}

	provider := NewMangapill(util.NewLogger())

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			pages, err := provider.FindChapterPages(tt.chapterId)
			if assert.NoError(t, err, "provider.FindChapterPages() error") {
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
