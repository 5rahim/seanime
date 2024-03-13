package torrent_client

import (
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/nyaa"
	"github.com/seanime-app/seanime/internal/test_utils"
	"github.com/seanime-app/seanime/internal/torrent"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSmartSelect(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.TorrentClient())

	destination := t.TempDir()

	anilistClientWrapper := anilist.TestGetMockAnilistClientWrapper()

	// get repo
	repo := getTestRepo(t)

	tests := []struct {
		name             string
		mediaId          int
		url              string
		selectedEpisodes []int
		absoluteOffset   int
	}{
		{
			name:             "Kakegurui xx",
			mediaId:          100876,
			url:              "https://nyaa.si/view/1553978", // kakegurui season 1 + season 2
			selectedEpisodes: []int{10, 11, 12},
			absoluteOffset:   12,
		},
		{
			name:             "Spy x Family",
			mediaId:          1661695,
			url:              "https://nyaa.si/view/1661695", // spy x family (01-25)
			selectedEpisodes: []int{10, 11, 12},
			absoluteOffset:   0,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			ok := repo.Start()
			if !assert.True(t, ok) {
				return
			}

			// get magnet
			magnet, err := nyaa.TorrentMagnet(tt.url)
			assert.NoError(t, err)

			// get hash
			hash, ok := torrent.ExtractHashFromMagnet(magnet)
			assert.True(t, ok)

			t.Log(tt.name, hash)

			// get media
			media, err := anilist.GetBaseMediaById(anilistClientWrapper, tt.mediaId)
			if err != nil {
				t.Fatalf("error getting media: %s", err.Error())
			}

			err = repo.AddMagnets([]string{magnet}, destination)
			if err != nil {
				t.Fatalf("error adding magnet: %s", err.Error())
			}

			err = repo.SmartSelect(&SmartSelect{
				Magnets:               []string{magnet},
				Enabled:               true,
				MissingEpisodeNumbers: tt.selectedEpisodes,
				AbsoluteOffset:        tt.absoluteOffset,
				Media:                 media,
			})

			if testDefaultClient == TransmissionProvider {
				assert.Error(t, err)
			} else if testDefaultClient == QbittorrentProvider {
				assert.NoError(t, err)
			}

			err = repo.PauseTorrents([]string{hash})
			assert.NoError(t, err)

		})

	}

}
