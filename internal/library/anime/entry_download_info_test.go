package anime_test

import (
	"context"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/library/anime"
	"seanime/internal/test_utils"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEntryDownloadInfo(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	metadataProvider := metadata.GetMockProvider(t)

	anilistClient := anilist.TestGetMockAnilistClient()
	animeCollection, err := anilistClient.AnimeCollection(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name                             string
		localFiles                       []*anime.LocalFile
		mediaId                          int
		currentProgress                  int
		status                           anilist.MediaListStatus
		expectedEpisodeNumbersToDownload []struct {
			episodeNumber int
			aniDbEpisode  string
		}
	}{
		{
			// AniList includes episode 0 as a main episode but AniDB lists it as a special S1
			// So we should expect to see episode 0 (S1) in the list of episodes to download
			name:            "Mushoku Tensei: Jobless Reincarnation Season 2",
			localFiles:      nil,
			mediaId:         146065,
			currentProgress: 0,
			status:          anilist.MediaListStatusCurrent,
			expectedEpisodeNumbersToDownload: []struct {
				episodeNumber int
				aniDbEpisode  string
			}{
				{episodeNumber: 0, aniDbEpisode: "S1"},
				{episodeNumber: 1, aniDbEpisode: "1"},
				{episodeNumber: 2, aniDbEpisode: "2"},
				{episodeNumber: 3, aniDbEpisode: "3"},
				{episodeNumber: 4, aniDbEpisode: "4"},
				{episodeNumber: 5, aniDbEpisode: "5"},
				{episodeNumber: 6, aniDbEpisode: "6"},
				{episodeNumber: 7, aniDbEpisode: "7"},
				{episodeNumber: 8, aniDbEpisode: "8"},
				{episodeNumber: 9, aniDbEpisode: "9"},
				{episodeNumber: 10, aniDbEpisode: "10"},
				{episodeNumber: 11, aniDbEpisode: "11"},
				{episodeNumber: 12, aniDbEpisode: "12"},
			},
		},
		{
			// Same as above but progress of 1 should just eliminate episode 0 from the list and not episode 1
			name:            "Mushoku Tensei: Jobless Reincarnation Season 2 - 2",
			localFiles:      nil,
			mediaId:         146065,
			currentProgress: 1,
			status:          anilist.MediaListStatusCurrent,
			expectedEpisodeNumbersToDownload: []struct {
				episodeNumber int
				aniDbEpisode  string
			}{
				{episodeNumber: 1, aniDbEpisode: "1"},
				{episodeNumber: 2, aniDbEpisode: "2"},
				{episodeNumber: 3, aniDbEpisode: "3"},
				{episodeNumber: 4, aniDbEpisode: "4"},
				{episodeNumber: 5, aniDbEpisode: "5"},
				{episodeNumber: 6, aniDbEpisode: "6"},
				{episodeNumber: 7, aniDbEpisode: "7"},
				{episodeNumber: 8, aniDbEpisode: "8"},
				{episodeNumber: 9, aniDbEpisode: "9"},
				{episodeNumber: 10, aniDbEpisode: "10"},
				{episodeNumber: 11, aniDbEpisode: "11"},
				{episodeNumber: 12, aniDbEpisode: "12"},
			},
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			anilistEntry, _ := animeCollection.GetListEntryFromAnimeId(tt.mediaId)

			animeMetadata, err := metadataProvider.GetAnimeMetadata(metadata.AnilistPlatform, tt.mediaId)
			require.NoError(t, err)

			info, err := anime.NewEntryDownloadInfo(&anime.NewEntryDownloadInfoOptions{
				LocalFiles:       tt.localFiles,
				Progress:         &tt.currentProgress,
				Status:           &tt.status,
				Media:            anilistEntry.Media,
				MetadataProvider: metadataProvider,
				AnimeMetadata:    animeMetadata,
			})

			if assert.NoError(t, err) && assert.NotNil(t, info) {

				foundEpToDownload := make([]struct {
					episodeNumber int
					aniDbEpisode  string
				}, 0)
				for _, ep := range info.EpisodesToDownload {
					foundEpToDownload = append(foundEpToDownload, struct {
						episodeNumber int
						aniDbEpisode  string
					}{
						episodeNumber: ep.EpisodeNumber,
						aniDbEpisode:  ep.AniDBEpisode,
					})
				}

				assert.ElementsMatch(t, tt.expectedEpisodeNumbersToDownload, foundEpToDownload)

			}

		})

	}

}

func TestNewEntryDownloadInfo2(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	mediaId := 21

	metadataProvider := metadata.GetMockProvider(t)

	anilistClient := anilist.TestGetMockAnilistClient()
	animeCollection, err := anilistClient.AnimeCollection(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	anilistEntry, _ := animeCollection.GetListEntryFromAnimeId(mediaId)

	animeMetadata, err := metadataProvider.GetAnimeMetadata(metadata.AnilistPlatform, mediaId)
	require.NoError(t, err)

	info, err := anime.NewEntryDownloadInfo(&anime.NewEntryDownloadInfoOptions{
		LocalFiles:       nil,
		Progress:         lo.ToPtr(0),
		Status:           lo.ToPtr(anilist.MediaListStatusCurrent),
		Media:            anilistEntry.Media,
		MetadataProvider: metadataProvider,
		AnimeMetadata:    animeMetadata,
	})
	require.NoError(t, err)

	require.NotNil(t, info)

	t.Log(len(info.EpisodesToDownload))
	assert.GreaterOrEqual(t, len(info.EpisodesToDownload), 1096)
}
