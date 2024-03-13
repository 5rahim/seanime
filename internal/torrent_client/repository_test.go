package torrent_client

import (
	"github.com/seanime-app/seanime/internal/qbittorrent"
	"github.com/seanime-app/seanime/internal/test_utils"
	"github.com/seanime-app/seanime/internal/torrent"
	"github.com/seanime-app/seanime/internal/transmission"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var testDefaultClient = TransmissionProvider

// Add and remove
func TestAddAndRemove(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.TorrentClient())

	destination := t.TempDir()

	const url = "https://animetosho.org/view/subsplease-sousou-no-frieren-24-480p-c467b289-mkv.1847941"

	// get repo
	repo := getTestRepo(t)

	ok := repo.Start()
	if !assert.True(t, ok) {
		return
	}

	// get magnet
	magnet, err := torrent.ScrapeMagnet(url)
	assert.NoError(t, err)
	// get hash
	hash, ok := torrent.ExtractHashFromMagnet(magnet)
	assert.True(t, ok)

	err = repo.AddMagnets([]string{magnet}, destination)
	if err != nil {
		t.Fatalf("error adding magnet: %s", err.Error())
	}

	t.Log(hash)

	time.Sleep(5 * time.Second)

	err = repo.RemoveTorrents([]string{hash})
	assert.NoError(t, err)

}

// Clean up
func TestRemoveTorrents(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.TorrentClient())

	const url = "https://animetosho.org/view/subsplease-sousou-no-frieren-24-480p-c467b289-mkv.1847941"

	// get repo
	repo := getTestRepo(t)
	// get magnet
	magnet, err := torrent.ScrapeMagnet(url)
	assert.NoError(t, err)
	// get hash
	hash, ok := torrent.ExtractHashFromMagnet(magnet)
	assert.True(t, ok)

	t.Log(hash)

	err = repo.RemoveTorrents([]string{hash})
	assert.NoError(t, err)

}

//----------------------------------------------------------------------------------------------------------------------

func getTestRepo(t *testing.T) *Repository {
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
		Logger:            logger,
		QbittorrentClient: qBittorrentClient,
		Transmission:      trans,
		Provider:          testDefaultClient,
	}

	return repo
}
