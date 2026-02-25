package anime_test

import (
	"context"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata_provider"
	"seanime/internal/database/db"
	"seanime/internal/library/anime"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test to retrieve accurate missing episodes
// DEPRECATED
func TestNewMissingEpisodes(t *testing.T) {
	t.Skip("Outdated test")
	test_utils.InitTestProvider(t, test_utils.Anilist())
	logger := util.NewLogger()
	database, _ := db.NewDatabase(t.TempDir(), "test", logger)

	metadataProvider := metadata_provider.GetFakeProvider(t, database)

	anilistClient := anilist.TestGetMockAnilistClient()
	animeCollection, err := anilistClient.AnimeCollection(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name                    string
		mediaId                 int
		localFiles              []*anime.LocalFile
		mediaAiredEpisodes      int
		currentProgress         int
		expectedMissingEpisodes int
	}{
		{
			// Sousou no Frieren - 10 currently aired episodes
			// User has 5 local files from ep 1 to 5, but only watched 4 episodes
			// So we should expect to see 5 missing episodes
			name:    "Sousou no Frieren, missing 5 episodes",
			mediaId: 154587,
			localFiles: anime.MockHydratedLocalFiles(
				anime.MockGenerateHydratedLocalFileGroupOptions("E:/Anime", "E:\\Anime\\Sousou no Frieren\\[SubsPlease] Sousou no Frieren - %ep (1080p) [F02B9CEE].mkv", 154587, []anime.MockHydratedLocalFileWrapperOptionsMetadata{
					{MetadataEpisode: 1, MetadataAniDbEpisode: "1", MetadataType: anime.LocalFileTypeMain},
					{MetadataEpisode: 2, MetadataAniDbEpisode: "2", MetadataType: anime.LocalFileTypeMain},
					{MetadataEpisode: 3, MetadataAniDbEpisode: "3", MetadataType: anime.LocalFileTypeMain},
					{MetadataEpisode: 4, MetadataAniDbEpisode: "4", MetadataType: anime.LocalFileTypeMain},
					{MetadataEpisode: 5, MetadataAniDbEpisode: "5", MetadataType: anime.LocalFileTypeMain},
				}),
			),
			mediaAiredEpisodes: 10,
			currentProgress:    4,
			//expectedMissingEpisodes: 5,
			expectedMissingEpisodes: 1, // DEVNOTE: Now the value is 1 at most because everything else is merged
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			// Mock Anilist collection
			anilist.TestModifyAnimeCollectionEntry(animeCollection, tt.mediaId, anilist.TestModifyAnimeCollectionEntryInput{
				Progress:      new(tt.currentProgress), // Mock progress
				AiredEpisodes: new(tt.mediaAiredEpisodes),
				NextAiringEpisode: &anilist.BaseAnime_NextAiringEpisode{
					Episode: tt.mediaAiredEpisodes + 1,
				},
			})

		})

		if assert.NoError(t, err) {
			missingData := anime.NewMissingEpisodes(&anime.NewMissingEpisodesOptions{
				AnimeCollection:     animeCollection,
				LocalFiles:          tt.localFiles,
				MetadataProviderRef: util.NewRef(metadataProvider),
			})

			assert.Equal(t, tt.expectedMissingEpisodes, len(missingData.Episodes))
		}

	}

}
