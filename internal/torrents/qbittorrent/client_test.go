package qbittorrent

import (
	"github.com/seanime-app/seanime/internal/test_utils"
	"github.com/seanime-app/seanime/internal/torrents/qbittorrent/model"
	"github.com/seanime-app/seanime/internal/util"
	"testing"
)

func TestGetList(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.TorrentClient())

	client := NewClient(&NewClientOptions{
		Logger:   util.NewLogger(),
		Username: test_utils.ConfigData.Provider.QbittorrentUsername,
		Password: test_utils.ConfigData.Provider.QbittorrentPassword,
		Port:     test_utils.ConfigData.Provider.QbittorrentPort,
		Host:     test_utils.ConfigData.Provider.QbittorrentHost,
		Path:     test_utils.ConfigData.Provider.QbittorrentPath,
	})

	res, err := client.Torrent.GetList(&qbittorrent_model.GetTorrentListOptions{
		Filter:   "",
		Category: nil,
		Sort:     "",
		Reverse:  false,
		Limit:    0,
		Offset:   0,
		Hashes:   "",
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	for _, torrent := range res {
		t.Logf("%+v", torrent)
	}

}
