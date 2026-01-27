package torrent

import (
	"context"
	"seanime/internal/api/anilist"
	"seanime/internal/database/db"
	"seanime/internal/extension"
	"seanime/internal/platforms/anilist_platform"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"testing"
)

func TestSmartSearch(t *testing.T) {
	t.Skip("Add fake provider")
	test_utils.InitTestProvider(t)

	anilistClient := anilist.TestGetMockAnilistClient()
	logger := util.NewLogger()
	database, err := db.NewDatabase(test_utils.ConfigData.Path.DataDir, test_utils.ConfigData.Database.Name, logger)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	extensionBankRef := util.NewRef(extension.NewUnifiedBank())
	anilistPlatform := anilist_platform.NewAnilistPlatform(util.NewRef(anilistClient), extensionBankRef, logger, database)

	repo := getTestRepo(t)

	tests := []struct {
		smartSearch    bool
		query          string
		episodeNumber  int
		batch          bool
		mediaId        int
		absoluteOffset int
		resolution     string
		provider       string
	}{
		{
			smartSearch:    true,
			query:          "",
			episodeNumber:  5,
			batch:          false,
			mediaId:        162670, // Dr. Stone S3
			absoluteOffset: 48,
			resolution:     "1080",
			provider:       "tosho",
		},
		{
			smartSearch:    true,
			query:          "",
			episodeNumber:  1,
			batch:          true,
			mediaId:        77, // Mahou Shoujo Lyrical Nanoha A's
			absoluteOffset: 0,
			resolution:     "1080",
			provider:       "tosho",
		},
		{
			smartSearch:    true,
			query:          "",
			episodeNumber:  1,
			batch:          true,
			mediaId:        109731, // Hibike Season 3
			absoluteOffset: 0,
			resolution:     "1080",
			provider:       "tosho",
		},
		{
			smartSearch:    true,
			query:          "",
			episodeNumber:  1,
			batch:          true,
			mediaId:        1915, // Magical Girl Lyrical Nanoha StrikerS
			absoluteOffset: 0,
			resolution:     "",
			provider:       "tosho",
		},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {

			media, err := anilistPlatform.GetAnime(t.Context(), tt.mediaId)
			if err != nil {
				t.Fatalf("could not fetch media id %d", tt.mediaId)
			}

			data, err := repo.SearchAnime(context.Background(), AnimeSearchOptions{
				Provider:      tt.provider,
				Type:          AnimeSearchTypeSmart,
				Media:         media,
				Query:         "",
				Batch:         tt.batch,
				EpisodeNumber: tt.episodeNumber,
				BestReleases:  false,
				Resolution:    tt.resolution,
			})
			if err != nil {
				t.Errorf("NewSmartSearch() failed: %v", err)
			}

			t.Log("----------------------- Previews --------------------------")
			for _, preview := range data.Previews {
				t.Logf("> %s", preview.Torrent.Name)
				if preview.Episode != nil {
					t.Logf("\t\t %s", preview.Episode.DisplayTitle)
				} else {
					t.Logf("\t\t Batch")
				}
			}
			t.Log("----------------------- Torrents --------------------------")
			for _, torrent := range data.Torrents {
				t.Logf("> %s", torrent.Name)
			}

		})
	}
}
