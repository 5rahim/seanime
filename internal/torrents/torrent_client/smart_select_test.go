package torrent_client

import (
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/platform"
	"github.com/seanime-app/seanime/internal/test_utils"
	"github.com/seanime-app/seanime/internal/torrents/torrent"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSmartSelect(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.TorrentClient())

	destination := t.TempDir()

	anilistClientWrapper := anilist.TestGetMockAnilistClientWrapper()
	anilistPlatform := platform.NewAnilistPlatform(anilistClientWrapper, util.NewLogger())

	// get repo

	tests := []struct {
		name             string
		mediaId          int
		url              string
		selectedEpisodes []int
		client           string
	}{
		{
			name:             "Kakegurui xx (Season 2)",
			mediaId:          100876,
			url:              "https://nyaa.si/view/1553978", // kakegurui season 1 + season 2
			selectedEpisodes: []int{10, 11, 12},              // should select 10, 11, 12 in season 2
			client:           QbittorrentClient,
		},
		{
			name:             "Spy x Family",
			mediaId:          140960,
			url:              "https://nyaa.si/view/1661695", // spy x family (01-25)
			selectedEpisodes: []int{10, 11, 12},              // should select 10, 11, 12
			client:           QbittorrentClient,
		},
		{
			name:             "Spy x Family Part 2",
			mediaId:          142838,
			url:              "https://nyaa.si/view/1661695", // spy x family (01-25)
			selectedEpisodes: []int{10, 11, 12, 13},          // should select 22, 23, 24, 25
			client:           QbittorrentClient,
		},
		{
			name:             "Kakegurui xx (Season 2)",
			mediaId:          100876,
			url:              "https://nyaa.si/view/1553978", // kakegurui season 1 + season 2
			selectedEpisodes: []int{10, 11, 12},              // should select 10, 11, 12 in season 2
			client:           TransmissionClient,
		},
		{
			name:             "Spy x Family",
			mediaId:          140960,
			url:              "https://nyaa.si/view/1661695", // spy x family (01-25)
			selectedEpisodes: []int{10, 11, 12},              // should select 10, 11, 12
			client:           TransmissionClient,
		},
		{
			name:             "Spy x Family Part 2",
			mediaId:          142838,
			url:              "https://nyaa.si/view/1661695", // spy x family (01-25)
			selectedEpisodes: []int{10, 11, 12, 13},          // should select 22, 23, 24, 25
			client:           TransmissionClient,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			repo := getTestRepo(t, tt.client)

			ok := repo.Start()
			if !assert.True(t, ok) {
				return
			}

			// get media
			completeAnime, err := anilistPlatform.GetAnimeWithRelations(tt.mediaId)
			if err != nil {
				t.Fatalf("error getting media: %s", err.Error())
			}

			hash, err := torrent.ScrapeHash(tt.url)

			err = repo.SmartSelect(&SmartSelectParams{
				Url:              tt.url,
				EpisodeNumbers:   tt.selectedEpisodes,
				Media:            completeAnime,
				Platform:         anilistPlatform,
				Destination:      destination,
				ShouldAddTorrent: true,
			})
			// Remove torrent
			defer repo.RemoveTorrents([]string{hash})

			if assert.NoError(t, err) {

				// Pause the torrent
				err = repo.PauseTorrents([]string{hash})

				repo.logger.Info().Msg("[TEST] SMART SELECT SUCCESSFUL, CHECK MANUALLY")

				time.Sleep(20 * time.Second) // /!\ Can't verify programmatically that the files have been deselected, so check manually

			}

		})

	}

}
