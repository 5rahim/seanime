package onlinestream

import (
	"errors"
	"github.com/davecgh/go-spew/spew"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGogoanime_Search(t *testing.T) {

	gogo := NewGogoanime(util.NewLogger())

	tests := []struct {
		name   string
		query  string
		dubbed bool
	}{
		{
			name:   "One Piece",
			query:  "One Piece",
			dubbed: false,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			results, err := gogo.Search(tt.query, tt.dubbed)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			assert.NotEmpty(t, results)

			for _, r := range results {
				assert.NotEmpty(t, r.ID, "ID is empty")
				assert.NotEmpty(t, r.Title, "Title is empty")
				assert.NotEmpty(t, r.URL, "URL is empty")
			}

			spew.Dump(results)

		})

	}

}

func TestGogoanime_FetchEpisodes(t *testing.T) {

	tests := []struct {
		name string
		id   string
	}{
		{
			name: "One Piece",
			id:   "one-piece",
		},
		{
			name: "One Piece (Dub)",
			id:   "one-piece-dub",
		},
	}

	gogo := NewGogoanime(util.NewLogger())

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			episodes, err := gogo.FetchAnimeEpisodes(tt.id)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			assert.NotEmpty(t, episodes)

			for _, e := range episodes {
				assert.NotEmpty(t, e.ID, "ID is empty")
				assert.NotEmpty(t, e.Number, "Number is empty")
				assert.NotEmpty(t, e.URL, "URL is empty")
			}

			spew.Dump(episodes)

		})

	}

}

func TestGogoanime_FetchSources(t *testing.T) {

	tests := []struct {
		name    string
		episode *AnimeEpisode
		server  Server
	}{
		{
			name: "One Piece",
			episode: &AnimeEpisode{
				ID:     "one-piece-episode-1075",
				Number: 1075,
				URL:    "https://anitaku.to/one-piece-episode-1075",
			},
			server: VidstreamingServer,
		},
		{
			name: "One Piece",
			episode: &AnimeEpisode{
				ID:     "one-piece-episode-1075",
				Number: 1075,
				URL:    "https://anitaku.to/one-piece-episode-1075",
			},
			server: StreamSBServer,
		},
		{
			name: "One Piece",
			episode: &AnimeEpisode{
				ID:     "one-piece-episode-1075",
				Number: 1075,
				URL:    "https://anitaku.to/one-piece-episode-1075",
			},
			server: GogocdnServer,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			gogo := NewGogoanime(util.NewLogger())

			sources, err := gogo.FetchEpisodeSources(tt.episode, tt.server)
			if err != nil {
				if !errors.Is(err, ErrSourceNotFound) {
					t.Fatal(err)
				}
			}

			if err != nil {
				t.Skip("Source not found")
			}

			assert.NotEmpty(t, sources)

			for _, s := range sources.Sources {
				assert.NotEmpty(t, s, "Source is empty")
			}

			spew.Dump(sources)

		})

	}

}
