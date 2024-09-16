package anime

import (
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/platforms/anilist_platform"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"testing"
)

// TestNewAnimeEntry tests /library/entry endpoint.
// /!\ MAKE SURE TO HAVE THE MEDIA ADDED TO YOUR LIST TEST ACCOUNT LISTS
func TestNewAnimeEntry(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	metadataProvider := metadata.GetMockProvider(t)

	tests := []struct {
		name                              string
		mediaId                           int
		localFiles                        []*LocalFile
		currentProgress                   int
		expectedNextEpisodeNumber         int
		expectedNextEpisodeProgressNumber int
	}{
		{
			name:    "Sousou no Frieren",
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
			currentProgress:                   4,
			expectedNextEpisodeNumber:         5,
			expectedNextEpisodeProgressNumber: 5,
		},
		{
			name:    "Mushoku Tensei II Isekai Ittara Honki Dasu",
			mediaId: 146065,
			localFiles: MockHydratedLocalFiles(
				MockGenerateHydratedLocalFileGroupOptions("E:/Anime", "E:/Anime/Mushoku Tensei II Isekai Ittara Honki Dasu/[SubsPlease] Mushoku Tensei S2 - 00 (1080p) [9C362DC3].mkv", 146065, []MockHydratedLocalFileWrapperOptionsMetadata{
					{MetadataEpisode: 0, MetadataAniDbEpisode: "S1", MetadataType: LocalFileTypeMain}, // Special episode
					{MetadataEpisode: 1, MetadataAniDbEpisode: "1", MetadataType: LocalFileTypeMain},
					{MetadataEpisode: 2, MetadataAniDbEpisode: "2", MetadataType: LocalFileTypeMain},
					{MetadataEpisode: 3, MetadataAniDbEpisode: "3", MetadataType: LocalFileTypeMain},
					{MetadataEpisode: 4, MetadataAniDbEpisode: "4", MetadataType: LocalFileTypeMain},
					{MetadataEpisode: 5, MetadataAniDbEpisode: "5", MetadataType: LocalFileTypeMain},
					{MetadataEpisode: 6, MetadataAniDbEpisode: "6", MetadataType: LocalFileTypeMain},
					{MetadataEpisode: 7, MetadataAniDbEpisode: "7", MetadataType: LocalFileTypeMain},
					{MetadataEpisode: 8, MetadataAniDbEpisode: "8", MetadataType: LocalFileTypeMain},
					{MetadataEpisode: 9, MetadataAniDbEpisode: "9", MetadataType: LocalFileTypeMain},
					{MetadataEpisode: 10, MetadataAniDbEpisode: "10", MetadataType: LocalFileTypeMain},
					{MetadataEpisode: 11, MetadataAniDbEpisode: "11", MetadataType: LocalFileTypeMain},
					{MetadataEpisode: 12, MetadataAniDbEpisode: "12", MetadataType: LocalFileTypeMain},
				}),
			),
			currentProgress:                   0,
			expectedNextEpisodeNumber:         0,
			expectedNextEpisodeProgressNumber: 1,
		},
	}

	anilistClient := anilist.TestGetMockAnilistClient()
	anilistPlatform := anilist_platform.NewAnilistPlatform(anilistClient, util.NewLogger())
	animeCollection, err := anilistPlatform.GetAnimeCollection(false)
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			anilist.TestModifyAnimeCollectionEntry(animeCollection, tt.mediaId, anilist.TestModifyAnimeCollectionEntryInput{
				Progress: lo.ToPtr(tt.currentProgress), // Mock progress
			})

			entry, err := NewEntry(&NewEntryOptions{
				MediaId:          tt.mediaId,
				LocalFiles:       tt.localFiles,
				AnimeCollection:  animeCollection,
				Platform:         anilistPlatform,
				MetadataProvider: metadataProvider,
			})

			if assert.NoErrorf(t, err, "Failed to get mock data") {

				if assert.NoError(t, err) {

					// Mock progress is 4
					nextEp, found := entry.FindNextEpisode()
					if assert.True(t, found, "did not find next episode") {
						assert.Equal(t, tt.expectedNextEpisodeNumber, nextEp.EpisodeNumber, "next episode number mismatch")
						assert.Equal(t, tt.expectedNextEpisodeProgressNumber, nextEp.ProgressNumber, "next episode progress number mismatch")
					}

					t.Logf("Found %v episodes", len(entry.Episodes))

				}

			}

		})

	}
}
