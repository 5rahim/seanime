package entities

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

// TestNewMediaEntry tests /library/entry endpoint.
// /!\ MAKE SURE TO HAVE THE MEDIA ADDED TO YOUR LIST TEST ACCOUNT LISTS
func TestNewMediaEntry(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	metadataProvider := metadata.TestGetMockProvider(t)

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
					{metadataEpisode: 1, metadataAniDbEpisode: "1", metadataType: LocalFileTypeMain},
					{metadataEpisode: 2, metadataAniDbEpisode: "2", metadataType: LocalFileTypeMain},
					{metadataEpisode: 3, metadataAniDbEpisode: "3", metadataType: LocalFileTypeMain},
					{metadataEpisode: 4, metadataAniDbEpisode: "4", metadataType: LocalFileTypeMain},
					{metadataEpisode: 5, metadataAniDbEpisode: "5", metadataType: LocalFileTypeMain},
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
					{metadataEpisode: 0, metadataAniDbEpisode: "S1", metadataType: LocalFileTypeMain}, // Special episode
					{metadataEpisode: 1, metadataAniDbEpisode: "1", metadataType: LocalFileTypeMain},
					{metadataEpisode: 2, metadataAniDbEpisode: "2", metadataType: LocalFileTypeMain},
					{metadataEpisode: 3, metadataAniDbEpisode: "3", metadataType: LocalFileTypeMain},
					{metadataEpisode: 4, metadataAniDbEpisode: "4", metadataType: LocalFileTypeMain},
					{metadataEpisode: 5, metadataAniDbEpisode: "5", metadataType: LocalFileTypeMain},
					{metadataEpisode: 6, metadataAniDbEpisode: "6", metadataType: LocalFileTypeMain},
					{metadataEpisode: 7, metadataAniDbEpisode: "7", metadataType: LocalFileTypeMain},
					{metadataEpisode: 8, metadataAniDbEpisode: "8", metadataType: LocalFileTypeMain},
					{metadataEpisode: 9, metadataAniDbEpisode: "9", metadataType: LocalFileTypeMain},
					{metadataEpisode: 10, metadataAniDbEpisode: "10", metadataType: LocalFileTypeMain},
					{metadataEpisode: 11, metadataAniDbEpisode: "11", metadataType: LocalFileTypeMain},
					{metadataEpisode: 12, metadataAniDbEpisode: "12", metadataType: LocalFileTypeMain},
				}),
			),
			currentProgress:                   0,
			expectedNextEpisodeNumber:         0,
			expectedNextEpisodeProgressNumber: 1,
		},
	}

	anilistClientWrapper := anilist.TestGetMockAnilistClientWrapper()
	anilistCollection, err := anilistClientWrapper.AnimeCollection(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	aniZipCache := anizip.NewCache()

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			anilist.TestModifyAnimeCollectionEntry(anilistCollection, tt.mediaId, anilist.TestModifyAnimeCollectionEntryInput{
				Progress: lo.ToPtr(tt.currentProgress), // Mock progress
			})

			entry, err := NewMediaEntry(&NewMediaEntryOptions{
				MediaId:              tt.mediaId,
				LocalFiles:           tt.localFiles,
				AnizipCache:          aniZipCache,
				AnilistCollection:    anilistCollection,
				AnilistClientWrapper: anilistClientWrapper,
				MetadataProvider:     metadataProvider,
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
