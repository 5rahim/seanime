package torrent_client

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"seanime/internal/test_utils"
	"seanime/internal/torrents/qbittorrent"
	"seanime/internal/torrents/torrent"
	"seanime/internal/torrents/transmission"
	"seanime/internal/util"
	"testing"
	"time"
)

func TestRepository_GetFiles(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.TorrentClient())

	destination := t.TempDir()

	tests := []struct {
		name            string
		url             string
		expectedNbFiles int
		client          string
	}{
		{
			name:            "[EMBER] Demon Slayer (2023) (Season 3)",
			url:             "https://animetosho.org/view/ember-demon-slayer-2023-season-3-bdrip-1080p.n1778316",
			expectedNbFiles: 11,
			client:          TransmissionClient,
		},
		{
			name:            "[Tenrai-Sensei] Kakegurui (Season 1-2 + OVAs)",
			url:             "https://nyaa.si/view/1553978",
			expectedNbFiles: 27,
			client:          TransmissionClient,
		},
		{
			name:            "[EMBER] Demon Slayer (2023) (Season 3)",
			url:             "https://animetosho.org/view/ember-demon-slayer-2023-season-3-bdrip-1080p.n1778316",
			expectedNbFiles: 11,
			client:          QbittorrentClient,
		},
		{
			name:            "[Tenrai-Sensei] Kakegurui (Season 1-2 + OVAs)",
			url:             "https://nyaa.si/view/1553978",
			expectedNbFiles: 27,
			client:          QbittorrentClient,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			repo := getTestRepo(t, tt.client)

			started := repo.Start()
			assert.True(t, started)

			// Get magnet
			magnet, err := torrent.ScrapeMagnet(tt.url)

			if assert.NoError(t, err, "error scraping magnet") {

				// Get hash
				hash, err := torrent.ScrapeHash(tt.url)
				t.Log(hash)

				if assert.NoError(t, err, "hash not found") {

					// Add torrent
					err = repo.AddMagnets([]string{magnet}, destination)

					if assert.NoError(t, err, "error adding magnet") {

						files, err := repo.GetFiles(hash)

						if assert.NoError(t, err, "error getting files") {

							assert.Len(t, files, tt.expectedNbFiles)

							spew.Dump(files)

						}

					}

					// Remove torrent
					err = repo.RemoveTorrents([]string{hash})

				}
			}

		})

	}

}
func TestRepository_DeselectFiles(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.TorrentClient())

	destination := t.TempDir()

	tests := []struct {
		name            string
		url             string
		deselectIndices []int
		client          string
	}{
		{
			name:            "[EMBER] Demon Slayer (2023) (Season 3)",
			url:             "https://animetosho.org/view/ember-demon-slayer-2023-season-3-bdrip-1080p.n1778316",
			deselectIndices: []int{0, 1, 2, 3, 4},
			client:          TransmissionClient,
		},
		{
			name:            "[EMBER] Demon Slayer (2023) (Season 3)",
			url:             "https://animetosho.org/view/ember-demon-slayer-2023-season-3-bdrip-1080p.n1778316",
			deselectIndices: []int{0, 1, 2, 3, 4},
			client:          QbittorrentClient,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			repo := getTestRepo(t, tt.client)

			started := repo.Start()
			assert.True(t, started)

			// Get magnet
			magnet, err := torrent.ScrapeMagnet(tt.url)

			if assert.NoError(t, err, "error scraping magnet") {

				// Get hash
				hash, err := torrent.ScrapeHash(tt.url)
				t.Log(hash)

				if assert.NoError(t, err, "hash not found") {

					// Add torrent
					err = repo.AddMagnets([]string{magnet}, destination)

					if assert.NoError(t, err, "error adding magnet") {

						_, err := repo.GetFiles(hash)

						// Pause torrent
						err = repo.PauseTorrents([]string{hash})

						repo.logger.Info().Msg("[TEST] TORRENT PAUSED, CHECK MANUALLY")

						if assert.NoError(t, err, "error getting files") {

							err = repo.DeselectFiles(hash, tt.deselectIndices)

							if assert.NoError(t, err, "error deselecting files") {

								time.Sleep(20 * time.Second) // /!\ Can't verify programmatically that the files have been deselected, so check manually

								// Remove torrent
								err = repo.RemoveTorrents([]string{hash})

							}

						}

					}

					// Remove torrent
					err = repo.RemoveTorrents([]string{hash})

				}
			}

		})

	}

}

// Add and remove
func TestAddAndRemove(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.TorrentClient())

	destination := t.TempDir()

	tests := []struct {
		name   string
		url    string
		client string
	}{
		{
			name:   "Sousou no Frieren",
			url:    "https://animetosho.org/view/subsplease-sousou-no-frieren-24-480p-c467b289-mkv.1847941",
			client: TransmissionClient,
		},
		{
			name:   "Sousou no Frieren",
			url:    "https://animetosho.org/view/subsplease-sousou-no-frieren-24-480p-c467b289-mkv.1847941",
			client: QbittorrentClient,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// get repo
			repo := getTestRepo(t, tt.client)

			ok := repo.Start()
			if !assert.True(t, ok) {
				return
			}

			// get magnet
			magnet, err := torrent.ScrapeMagnet(tt.url)
			assert.NoError(t, err)
			// get hash
			hash, err := torrent.ScrapeHash(tt.url)
			assert.NoError(t, err)

			err = repo.AddMagnets([]string{magnet}, destination)
			if err != nil {
				t.Fatalf("error adding magnet: %s", err.Error())
			}

			t.Log(hash)

			time.Sleep(5 * time.Second)

			err = repo.RemoveTorrents([]string{hash})
			assert.NoError(t, err)

		})

	}

}

//----------------------------------------------------------------------------------------------------------------------

func getTestRepo(t *testing.T, provider string) *Repository {
	test_utils.InitTestProvider(t, test_utils.TorrentClient())

	logger := util.NewLogger()

	qBittorrentClient := qbittorrent.NewClient(&qbittorrent.NewClientOptions{
		Logger:   logger,
		Username: test_utils.ConfigData.Provider.QbittorrentUsername,
		Password: test_utils.ConfigData.Provider.QbittorrentPassword,
		Port:     test_utils.ConfigData.Provider.QbittorrentPort,
		Host:     test_utils.ConfigData.Provider.QbittorrentHost,
		Path:     test_utils.ConfigData.Provider.QbittorrentPath,
	})

	trans, err := transmission.New(&transmission.NewTransmissionOptions{
		Logger:   logger,
		Host:     test_utils.ConfigData.Provider.TransmissionHost,
		Path:     test_utils.ConfigData.Provider.TransmissionPath,
		Port:     test_utils.ConfigData.Provider.TransmissionPort,
		Username: test_utils.ConfigData.Provider.TransmissionUsername,
		Password: test_utils.ConfigData.Provider.TransmissionPassword,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = qBittorrentClient.Login()
	assert.NoError(t, err)

	// create repository
	repo := &Repository{
		logger:            logger,
		qBittorrentClient: qBittorrentClient,
		transmission:      trans,
		provider:          provider,
	}

	return repo
}
