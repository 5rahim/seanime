package anime

import (
	"context"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/api/anizip"
	"github.com/seanime-app/seanime/internal/api/metadata"
	"github.com/seanime-app/seanime/internal/test_utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewMediaEntryDownloadInfo(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	metadataProvider := metadata.TestGetMockProvider(t)

	anilistClientWrapper := anilist.TestGetMockAnilistClientWrapper()
	anilistCollection, err := anilistClientWrapper.AnimeCollection(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name                             string
		localFiles                       []*LocalFile
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

	anizipCache := anizip.NewCache()

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			anizipData, err := anizip.FetchAniZipMediaC("anilist", tt.mediaId, anizipCache)
			if err != nil {
				t.Fatal(err)
			}

			anilistEntry, _ := anilistCollection.GetListEntryFromMediaId(tt.mediaId)

			info, err := NewMediaEntryDownloadInfo(&NewMediaEntryDownloadInfoOptions{
				LocalFiles:       tt.localFiles,
				AnizipMedia:      anizipData,
				Progress:         &tt.currentProgress,
				Status:           &tt.status,
				Media:            anilistEntry.Media,
				MetadataProvider: metadataProvider,
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
