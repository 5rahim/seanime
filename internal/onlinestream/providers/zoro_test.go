package onlinestream_providers

import (
	"errors"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestZoro_Search(t *testing.T) {

	zoro := NewZoro()

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

			results, err := zoro.Search(tt.query, tt.dubbed)
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

func TestZoro_FetchEpisodes(t *testing.T) {

	tests := []struct {
		name string
		id   string
	}{
		{
			name: "One Piece",
			id:   "one-piece-100",
		},
	}

	zoro := NewZoro()

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			episodes, err := zoro.FetchEpisodes(tt.id)
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

func TestZoro_FetchSources(t *testing.T) {

	tests := []struct {
		name    string
		episode *ProviderEpisode
		server  Server
	}{
		{
			name: "One Piece",
			episode: &ProviderEpisode{
				ID:     "one-piece-100$episode$120118$both",
				Number: 1095,
				URL:    "https://hianime.to/watch/one-piece-100?ep=120118",
			},
			server: VidcloudServer,
		},
		{
			name: "One Piece",
			episode: &ProviderEpisode{
				ID:     "one-piece-100$episode$120118$both",
				Number: 1095,
				URL:    "https://hianime.to/watch/one-piece-100?ep=120118",
			},
			server: VidstreamingServer,
		},
		{
			name: "One Piece",
			episode: &ProviderEpisode{
				ID:     "one-piece-100$episode$120118$both",
				Number: 1095,
				URL:    "https://hianime.to/watch/one-piece-100?ep=120118",
			},
			server: StreamtapeServer,
		},
		{
			name: "One Piece",
			episode: &ProviderEpisode{
				ID:     "one-piece-100$episode$120118$both",
				Number: 1095,
				URL:    "https://hianime.to/watch/one-piece-100?ep=120118",
			},
			server: StreamSBServer,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			zoro := NewZoro()

			sources, err := zoro.FetchEpisodeSources(tt.episode, tt.server)
			if err != nil {
				if !errors.Is(err, ErrSourceNotFound) && !errors.Is(err, ErrServerNotFound) {
					t.Fatal(err)
				}
			}

			if err != nil {
				t.Skip(err.Error())
			}

			assert.NotEmpty(t, sources)

			for _, s := range sources.Sources {
				assert.NotEmpty(t, s, "Source is empty")
			}

			spew.Dump(sources)

		})

	}

}
