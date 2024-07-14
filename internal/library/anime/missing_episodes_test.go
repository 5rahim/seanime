package anime

import (
	"context"
	"github.com/samber/lo"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/api/anizip"
	"github.com/seanime-app/seanime/internal/api/metadata"
	"github.com/seanime-app/seanime/internal/test_utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Test to retrieve accurate missing episodes
func TestNewMissingEpisodes(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	metadataProvider := metadata.TestGetMockProvider(t)

	anilistClientWrapper := anilist.TestGetMockAnilistClientWrapper()
	animeCollection, err := anilistClientWrapper.AnimeCollection(context.Background(), nil)
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
					{metadataEpisode: 1, metadataAniDbEpisode: "1", metadataType: LocalFileTypeMain},
					{metadataEpisode: 2, metadataAniDbEpisode: "2", metadataType: LocalFileTypeMain},
					{metadataEpisode: 3, metadataAniDbEpisode: "3", metadataType: LocalFileTypeMain},
					{metadataEpisode: 4, metadataAniDbEpisode: "4", metadataType: LocalFileTypeMain},
					{metadataEpisode: 5, metadataAniDbEpisode: "5", metadataType: LocalFileTypeMain},
				}),
			),
			mediaAiredEpisodes: 10,
			currentProgress:    4,
			//expectedMissingEpisodes: 5,
			expectedMissingEpisodes: 1, // DEVNOTE: Now the value is 1 at most because everything else is aggregated
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
