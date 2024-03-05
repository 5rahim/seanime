package torrent_client

import (
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/nyaa"
	"github.com/seanime-app/seanime/internal/qbittorrent"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

var destination = "E:/Anime/Temp"

func TestSmartSelect(t *testing.T) {

	anilistClientWrapper := anilist.MockAnilistClientWrapper()

	// get repo
	repo := getRepo(t, destination)

	tests := []struct {
		name             string
		mediaId          int
		url              string
		selectedEpisodes []int
		absoluteOffset   int
	}{
		{
			name:             "Kakegurui xx",
			mediaId:          1553978,
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

			err := repo.QbittorrentClient.Start()
			assert.NoError(t, err)

			// get magnet
			magnet, err := nyaa.TorrentMagnet(tt.url)
			assert.NoError(t, err)

			// get hash
			hash, ok := nyaa.ExtractHashFromMagnet(magnet)
			assert.True(t, ok)

			t.Log(tt.name, hash)

			// get media
			media, err := anilist.GetBaseMediaById(anilistClientWrapper.Client, tt.mediaId)
			if err != nil {
				t.Fatalf("error getting media: %s", err.Error())
			}

			err = repo.AddMagnets([]string{magnet})
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

			if assert.NoError(t, err) {
				err = repo.PauseTorrents([]string{hash})
			}

		})

	}

}

// Clean up
func TestRemoveTorrents(t *testing.T) {

	const url = "https://nyaa.si/view/1553978"

	// get repo
	repo := getRepo(t, destination)
	// get magnet
	magnet, err := nyaa.TorrentMagnet(url)
	assert.NoError(t, err)
	// get hash
	hash, ok := nyaa.ExtractHashFromMagnet(magnet)
	assert.True(t, ok)

	t.Log(hash)

	err = repo.RemoveTorrents([]string{hash})
	assert.NoError(t, err)

}

//----------------------------------------------------------------------------------------------------------------------

func getRepo(t *testing.T, destination string) *TorrentClientRepository {

	logger := util.NewLogger()
	WSEventManager := events.NewMockWSEventManager(logger)

	qBittorrentClient := qbittorrent.NewClient(&qbittorrent.NewClientOptions{
		Logger:   logger,
		Username: "admin",
		Password: "adminadmin",
		Port:     8081,
		Host:     "127.0.0.1",
		Path:     "C:/Program Files/qBittorrent/qbittorrent.exe",
	})

	err := qBittorrentClient.Login()
	assert.NoError(t, err)

	// create repository
	repo := &TorrentClientRepository{
		Logger:            logger,
		QbittorrentClient: qBittorrentClient,
		WSEventManager:    WSEventManager,
		Destination:       destination,
	}

	return repo
}
