package tvdb

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"seanime/internal/api/anilist"
	"seanime/internal/api/anizip"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"testing"
)

func TestTVDB_FetchSeriesEpisodes(t *testing.T) {
	test_utils.InitTestProvider(t)

	anilistClient := anilist.TestGetMockAnilistClient()

	tests := []struct {
		name          string
		anilistId     int
		episodeNumber int
	}{
		{
			name:          "Dungeon Meshi",
			anilistId:     153518,
			episodeNumber: 1,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			mediaF, err := anilistClient.BaseAnimeByID(context.Background(), &tt.anilistId)
			if err != nil {
				t.Fatalf("could not media")
			}

			media := mediaF.GetMedia()

			anizipMedia, err := anizip.FetchAniZipMedia("anilist", tt.anilistId)
			if err != nil {
				t.Fatalf("could not fetch anime metadata for %s", tt.name)
			}

			tvdbId := anizipMedia.Mappings.ThetvdbID
			if tvdbId == 0 {
				t.Fatalf("could not find tvdb id for %s", tt.name)
			}

			// Create TVDB instance
			tvdb := NewTVDB(&NewTVDBOptions{
				ApiKey: "",
				Logger: util.NewLogger(),
			})

			episodes, err := tvdb.FetchSeriesEpisodes(tvdbId, FilterEpisodeMediaInfo{
				Year:           media.GetStartDate().GetYear(),
				Month:          media.GetStartDate().GetMonth(),
				TotalEp:        anizipMedia.GetMainEpisodeCount(),
				AbsoluteOffset: anizipMedia.GetOffset(),
			})
			if err != nil {
				t.Fatalf("could not fetch episodes for %s: %s", tt.name, err)
			}

			for _, episode := range episodes {

				t.Log("Episode ID:", episode.ID)
				t.Log("\t Number:", episode.Number)
				t.Log("\t Episode Number:", episode.Number)
				t.Log("\t Image:", episode.Image)
				t.Log("\t AiredAt:", episode.AiredAt)
				t.Log("")

			}

		})

	}

}

func TestTVDB_FetchSeasons(t *testing.T) {
	test_utils.InitTestProvider(t)

	tests := []struct {
		name      string
		anilistId int
	}{
		{
			name:      "Dungeon Meshi",
			anilistId: 153518,
		},
		{
			name:      "Boku no Kokoro no Yabai Yatsu 2nd Season",
			anilistId: 166216,
		},
		{
			name:      "Horiyima Piece",
			anilistId: 163132,
		},
		{
			name:      "Kusuriya no Hitorigoto",
			anilistId: 161645,
		},
		{
			name:      "Spy x Family Part 2",
			anilistId: 142838,
		},
		{
			name:      "Kimi No Todoke Season 2",
			anilistId: 9656,
		},
		{
			name:      "One Piece",
			anilistId: 21,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			anizipMedia, err := anizip.FetchAniZipMedia("anilist", tt.anilistId)
			if err != nil {
				t.Fatalf("could not fetch anime metadata for %s", tt.name)
			}

			tvdbId := anizipMedia.Mappings.ThetvdbID
			if tvdbId == 0 {
				t.Fatalf("could not find tvdb id for %s", tt.name)
			}

			// Create TVDB instance
			tvdb := NewTVDB(&NewTVDBOptions{
				ApiKey: "",
				Logger: util.NewLogger(),
			})

			// Get token
			_, err = tvdb.getTokenWithTries()
			if err != nil {
				t.Fatalf("could not get token: %s", err)
			}

			// Fetch seasons
			seasons, err := tvdb.fetchSeasons(tvdbId)
			if err != nil {
				t.Fatalf("could not fetch metadata for %s: %s", tt.name, err)
			}

			for _, season := range seasons {

				t.Log("Season ID:", season.ID)
				t.Log("\t Name:", season.Type.Name)
				t.Log("\t Number:", season.Number)
				t.Log("\t Type:", season.Type.Type)
				t.Log("\t LastUpdated:", season.LastUpdated)
				t.Log("\t Year:", season.Year)
				t.Log("")

			}

		})

	}

}

func TestTVDB_fetchEpisodes(t *testing.T) {
	test_utils.InitTestProvider(t)

	anilistClient := anilist.TestGetMockAnilistClient()

	tests := []struct {
		name      string
		anilistId int
	}{
		{
			name:      "Dungeon Meshi",
			anilistId: 153518,
		},
		{
			name:      "Boku no Kokoro no Yabai Yatsu 2nd Season",
			anilistId: 166216,
		},
		{
			name:      "Horiyima Piece",
			anilistId: 163132,
		},
		{
			name:      "Spy x Family Part 2",
			anilistId: 142838,
		},
		{
			name:      "Kusuriya no Hitorigoto",
			anilistId: 161645,
		},
		{
			name:      "Kimi No Todoke",
			anilistId: 6045,
		},
		{
			name:      "Kimi No Todoke Season 2",
			anilistId: 9656,
		},
		{
			name:      "One Piece",
			anilistId: 21,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			mediaF, err := anilistClient.BaseAnimeByID(context.Background(), &tt.anilistId)
			if err != nil {
				t.Fatalf("could not media")
			}

			media := mediaF.GetMedia()

			anizipMedia, err := anizip.FetchAniZipMedia("anilist", tt.anilistId)
			if err != nil {
				t.Fatalf("could not fetch anime metadata for %s", tt.name)
			}

			tvdbId := anizipMedia.Mappings.ThetvdbID
			if tvdbId == 0 {
				t.Fatalf("could not find tvdb id for %s", tt.name)
			}

			// Create TVDB instance
			tvdb := NewTVDB(&NewTVDBOptions{
				ApiKey: "",
				Logger: util.NewLogger(),
			})

			// Get token
			_, err = tvdb.getTokenWithTries()
			if err != nil {
				t.Fatalf("could not get token: %s", err)
			}

			// Fetch seasons
			seasons, err := tvdb.fetchSeasons(tvdbId)
			if err != nil {
				t.Fatalf("could not fetch metadata for %s: %s", tt.name, err)
			}

			// Fetch episodes
			res, err := tvdb.fetchEpisodes(seasons, false)
			if err != nil {
				t.Fatalf("could not fetch episode metadata for %s: %s", tt.name, err)
			}

			spew.Sdump(media)
			res, err = tvdb.filterEpisodes(res, FilterEpisodeMediaInfo{
				Year:           media.GetStartDate().GetYear(),
				Month:          media.GetStartDate().GetMonth(),
				TotalEp:        anizipMedia.GetMainEpisodeCount(),
				AbsoluteOffset: anizipMedia.GetOffset(),
			}, false)
			if err != nil {
				t.Fatalf("could not filter episodes for %s: %s", tt.name, err)
			}

			for _, episode := range res {

				t.Log("Episode ID:", episode.ID)
				t.Log("\t Number:", episode.Number)
				t.Log("\t Episode Number:", episode.Number)
				t.Log("\t Image:", episode.Image)
				t.Log("\t Season Number:", episode.SeasonNumber)
				t.Log("\t Year:", episode.Year)
				t.Log("\t Aired:", episode.Aired)

				t.Log("")

			}

		})

	}

}

func TestTVDB_fetchEpisodesAbsolute(t *testing.T) {
	test_utils.InitTestProvider(t)

	anilistClient := anilist.TestGetMockAnilistClient()

	tests := []struct {
		name          string
		anilistId     int
		episodeNumber int
	}{
		{
			name:          "Dungeon Meshi",
			anilistId:     153518,
			episodeNumber: 1,
		},
		{
			name:      "Boku no Kokoro no Yabai Yatsu 2nd Season",
			anilistId: 166216,
		},
		{
			name:      "Horiyima Piece",
			anilistId: 163132,
		},
		{
			name:      "Spy x Family Part 2",
			anilistId: 142838,
		},
		{
			name:      "Kusuriya no Hitorigoto",
			anilistId: 161645,
		},
		{
			name:      "Kimi No Todoke",
			anilistId: 6045,
		},
		{
			name:      "Kimi No Todoke Season 2",
			anilistId: 9656,
		},
		{
			name:      "One Piece",
			anilistId: 21,
		},
		{
			name:      "Hibike! Euphonium 3",
			anilistId: 109731,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			mediaF, err := anilistClient.BaseAnimeByID(context.Background(), &tt.anilistId)
			if err != nil {
				t.Fatalf("could not media")
			}

			media := mediaF.GetMedia()

			anizipMedia, err := anizip.FetchAniZipMedia("anilist", tt.anilistId)
			if err != nil {
				t.Fatalf("could not fetch anime metadata for %s", tt.name)
			}

			tvdbId := anizipMedia.Mappings.ThetvdbID
			if tvdbId == 0 {
				t.Fatalf("could not find tvdb id for %s", tt.name)
			}

			// Create TVDB instance
			tvdb := NewTVDB(&NewTVDBOptions{
				ApiKey: "",
				Logger: util.NewLogger(),
			})

			// Get token
			_, err = tvdb.getTokenWithTries()
			if err != nil {
				t.Fatalf("could not get token: %s", err)
			}

			// Fetch seasons
			seasons, err := tvdb.fetchSeasons(tvdbId)
			if err != nil {
				t.Fatalf("could not fetch metadata for %s: %s", tt.name, err)
			}

			// Fetch episodes
			res, err := tvdb.fetchEpisodes(seasons, true)
			if err != nil {
				t.Fatalf("could not fetch episode metadata for %s: %s", tt.name, err)
			}

			res, err = tvdb.filterEpisodes(res, FilterEpisodeMediaInfo{
				Year:           media.GetStartDate().GetYear(),
				Month:          media.GetStartDate().GetMonth(),
				TotalEp:        anizipMedia.GetMainEpisodeCount(),
				AbsoluteOffset: anizipMedia.GetOffset(),
			}, true)
			if err != nil {
				t.Fatalf("could not filter episodes for %s: %s", tt.name, err)
			}

			for _, episode := range res {

				t.Log("Episode ID:", episode.ID)
				t.Log("\t Number:", episode.Number)
				t.Log("\t Episode Number:", episode.Number)
				t.Log("\t Image:", episode.Image)
				t.Log("\t Season Number:", episode.SeasonNumber)
				t.Log("\t Year:", episode.Year)
				t.Log("\t Aired:", episode.Aired)

				t.Log("")

			}

		})

	}

}
