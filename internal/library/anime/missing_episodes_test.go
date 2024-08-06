package anime

import (
	"context"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"seanime/internal/api/anilist"
	"seanime/internal/api/anizip"
	"seanime/internal/api/metadata"
	"seanime/internal/test_utils"
	"testing"
)

// Test to retrieve accurate missing episodes
// DEPRECATED
func TestNewMissingEpisodes(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	metadataProvider := metadata.TestGetMockProvider(t)

	anilistClient := anilist.TestGetMockAnilistClient()
	animeCollection, err := anilistClient.AnimeCollection(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name                    string
		mediaId                 int
		localFiles              []*LocalFile
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
			localFiles: MockHydratedLocalFiles(
				MockGenerateHydratedLocalFileGroupOptions("E:/Anime", "E:\\Anime\\Sousou no Frieren\\[SubsPlease] Sousou no Frieren - %ep (1080p) [F02B9CEE].mkv", 154587, []MockHydratedLocalFileWrapperOptionsMetadata{
					{MetadataEpisode: 1, MetadataAniDbEpisode: "1", MetadataType: LocalFileTypeMain},
					{MetadataEpisode: 2, MetadataAniDbEpisode: "2", MetadataType: LocalFileTypeMain},
					{MetadataEpisode: 3, MetadataAniDbEpisode: "3", MetadataType: LocalFileTypeMain},
					{MetadataEpisode: 4, MetadataAniDbEpisode: "4", MetadataType: LocalFileTypeMain},
					{MetadataEpisode: 5, MetadataAniDbEpisode: "5", MetadataType: LocalFileTypeMain},
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
				Progress:      lo.ToPtr(tt.currentProgress), // Mock progress
				AiredEpisodes: lo.ToPtr(tt.mediaAiredEpisodes),
				NextAiringEpisode: &anilist.BaseAnime_NextAiringEpisode{
					Episode: tt.mediaAiredEpisodes + 1,
				},
			})

		})

		if assert.NoError(t, err) {
			missingData := NewMissingEpisodes(&NewMissingEpisodesOptions{
				AnimeCollection:  animeCollection,
				LocalFiles:       tt.localFiles,
				AnizipCache:      anizip.NewCache(),
				MetadataProvider: metadataProvider,
			})

			assert.Equal(t, tt.expectedMissingEpisodes, len(missingData.Episodes))
		}

	}

}
