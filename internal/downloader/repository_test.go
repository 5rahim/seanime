package downloader

import (
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/events"
	"github.com/seanime-app/seanime-server/internal/nyaa"
	"github.com/seanime-app/seanime-server/internal/qbittorrent"
	"github.com/seanime-app/seanime-server/internal/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func getRepo(t *testing.T) *QbittorrentRepository {

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
	repo := &QbittorrentRepository{
		Logger:         logger,
		Client:         qBittorrentClient,
		WSEventManager: WSEventManager,
		Destination:    "E:/Anime/Temp",
	}

	return repo
}

// const url = "https://nyaa.si/view/1661695" // spy x family (01-25)
// const mediaId = 142838                     // spy x family part 2
const url = "https://nyaa.si/view/1553978" // kakegurui season 1 + season 2
const mediaId = 100876                     // kakegurui xx

func TestSmartSelect(t *testing.T) {
	// get repo
	repo := getRepo(t)

	err := repo.Client.Start()
	assert.NoError(t, err)

	// get magnet
	magnet, err := nyaa.TorrentMagnet(url)
	assert.NoError(t, err)
	// get hash
	hash, ok := nyaa.ExtractHashFromMagnet(magnet)
	assert.True(t, ok)

	t.Log(hash)

	// get media
	anilistEntry, ok := anilist.MockGetCollectionEntry(mediaId)
	assert.True(t, ok)

	err = repo.AddMagnets([]string{magnet})
	assert.NoError(t, err)

	err = repo.SmartSelect(&SmartSelect{
		Magnets:               []string{magnet},
		Enabled:               true,
		MissingEpisodeNumbers: []int{10, 11, 12},
		AbsoluteOffset:        0,
		Media:                 anilistEntry.Media,
	})
	assert.NoError(t, err)

	err = repo.PauseTorrents([]string{hash})

}

func TestRemoveTorrents(t *testing.T) {
	// get repo
	repo := getRepo(t)
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
