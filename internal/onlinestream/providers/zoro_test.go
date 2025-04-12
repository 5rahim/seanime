package onlinestream_providers

import (
	"errors"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	hibikeonlinestream "seanime/internal/extension/hibike/onlinestream"
	"seanime/internal/util"
	"testing"
)

func TestZoro_Search(t *testing.T) {

	logger := util.NewLogger()
	zoro := NewZoro(logger)

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
			name:   "Dungeon Meshi",
			query:  "Dungeon Meshi",
			dubbed: false,
		},
		{
			name:   "Omoi, Omoware, Furi, Furare",
			query:  "Omoi, Omoware, Furi, Furare",
			dubbed: false,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			results, err := zoro.Search(hibikeonlinestream.SearchOptions{
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

			spew.Dump(results)

		})

	}

}

func TestZoro_FetchEpisodes(t *testing.T) {
	logger := util.NewLogger()

	tests := []struct {
		name string
		id   string
	}{
		{
			name: "One Piece",
			id:   "one-piece-100",
		},
		{
			name: "The Apothecary Diaries",
			id:   "the-apothecary-diaries-18578",
		},
	}

	zoro := NewZoro(logger)

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			episodes, err := zoro.FindEpisodes(tt.id)
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
	logger := util.NewLogger()

	tests := []struct {
		name    string
		episode *hibikeonlinestream.EpisodeDetails
		server  string
	}{
		{
			name: "One Piece",
			episode: &hibikeonlinestream.EpisodeDetails{
				ID:     "one-piece-100$episode$120118$both",
				Number: 1095,
				URL:    "https://hianime.to/watch/one-piece-100?ep=120118",
			},
			server: VidcloudServer,
		},
		{
			name: "One Piece",
			episode: &hibikeonlinestream.EpisodeDetails{
				ID:     "one-piece-100$episode$120118$both",
				Number: 1095,
				URL:    "https://hianime.to/watch/one-piece-100?ep=120118",
			},
			server: VidstreamingServer,
		},
		{
			name: "One Piece",
			episode: &hibikeonlinestream.EpisodeDetails{
				ID:     "one-piece-100$episode$120118$both",
				Number: 1095,
				URL:    "https://hianime.to/watch/one-piece-100?ep=120118",
			},
			server: StreamtapeServer,
		},
		{
			name: "One Piece",
			episode: &hibikeonlinestream.EpisodeDetails{
				ID:     "one-piece-100$episode$120118$both",
				Number: 1095,
				URL:    "https://hianime.to/watch/one-piece-100?ep=120118",
			},
			server: StreamSBServer,
		},
		{
			name: "Apothecary Diaries",
			episode: &hibikeonlinestream.EpisodeDetails{
				ID:     "the-apothecary-diaries-18578$episode$122954$sub",
				Number: 24,
				URL:    "https://hianime.to/watch/the-apothecary-diaries-18578?ep=122954",
			},
			server: StreamSBServer,
		},
	}
	zoro := NewZoro(logger)

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			serverSources, err := zoro.FindEpisodeServer(tt.episode, tt.server)
			if err != nil {
				if !errors.Is(err, ErrSourceNotFound) && !errors.Is(err, ErrServerNotFound) {
					t.Fatal(err)
				}
			}

			if err != nil {
				t.Skip(err.Error())
			}

			assert.NotEmpty(t, serverSources)

			for _, s := range serverSources.VideoSources {
				assert.NotEmpty(t, s, "Source is empty")
			}

			spew.Dump(serverSources)

		})

	}

}
