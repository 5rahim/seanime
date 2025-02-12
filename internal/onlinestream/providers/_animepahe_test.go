package onlinestream_providers

import (
	"errors"
	"github.com/stretchr/testify/assert"
	hibikeonlinestream "seanime/internal/extension/hibike/onlinestream"
	"seanime/internal/util"
	"testing"
)

func TestAnimepahe_Search(t *testing.T) {

	ap := NewAnimepahe(util.NewLogger())

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
		{
			name:   "Blue Lock Season 2",
			query:  "Blue Lock Season 2",
			dubbed: false,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			results, err := ap.Search(hibikeonlinestream.SearchOptions{
				Query: tt.query,
				Dub:   tt.dubbed,
			})
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			assert.NotEmpty(t, results)

			for _, r := range results {
				assert.NotEmpty(t, r.ID, "ID is empty")
				assert.NotEmpty(t, r.Title, "Title is empty")
				assert.NotEmpty(t, r.URL, "URL is empty")
			}

			util.Spew(results)

		})

	}

}

func TestAnimepahe_FetchEpisodes(t *testing.T) {

	tests := []struct {
		name string
		id   string
	}{
		{
			name: "One Piece",
			id:   "4",
		},
		{
			name: "Blue Lock Season 2",
			id:   "5648",
		},
	}

	ap := NewAnimepahe(util.NewLogger())

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			episodes, err := ap.FindEpisodes(tt.id)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			assert.NotEmpty(t, episodes)

			for _, e := range episodes {
				assert.NotEmpty(t, e.ID, "ID is empty")
				assert.NotEmpty(t, e.Number, "Number is empty")
				assert.NotEmpty(t, e.URL, "URL is empty")
			}

			util.Spew(episodes)

		})

	}

}

func TestAnimepahe_FetchSources(t *testing.T) {

	tests := []struct {
		name    string
		episode *hibikeonlinestream.EpisodeDetails
		server  string
	}{
		{
			name: "One Piece",
			episode: &hibikeonlinestream.EpisodeDetails{
				ID:     "63391$4",
				Number: 1115,
				URL:    "",
			},
			server: KwikServer,
		},
		{
			name: "Blue Lock Season 2 - Episode 1",
			episode: &hibikeonlinestream.EpisodeDetails{
				ID:     "64056$5648",
				Number: 1,
				URL:    "",
			},
			server: KwikServer,
		},
	}
	ap := NewAnimepahe(util.NewLogger())

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			sources, err := ap.FindEpisodeServer(tt.episode, tt.server)
			if err != nil {
				if !errors.Is(err, ErrSourceNotFound) {
					t.Fatal(err)
				}
			}

			if err != nil {
				t.Skip("Source not found")
			}

			assert.NotEmpty(t, sources)

			for _, s := range sources.VideoSources {
				assert.NotEmpty(t, s, "Source is empty")
			}

			util.Spew(sources)

		})

	}

}
