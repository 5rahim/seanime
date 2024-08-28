package qbittorrent

import (
	"github.com/stretchr/testify/require"
	"seanime/internal/test_utils"
	"seanime/internal/torrent_clients/qbittorrent/model"
	"seanime/internal/util"
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
	require.NoError(t, err)

	for _, torrent := range res {
		t.Logf("%+v", torrent)
	}

}

func TestGetMainDataList(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.TorrentClient())

	client := NewClient(&NewClientOptions{
		Logger:   util.NewLogger(),
		Username: test_utils.ConfigData.Provider.QbittorrentUsername,
		Password: test_utils.ConfigData.Provider.QbittorrentPassword,
		Port:     test_utils.ConfigData.Provider.QbittorrentPort,
		Host:     test_utils.ConfigData.Provider.QbittorrentHost,
		Path:     test_utils.ConfigData.Provider.QbittorrentPath,
	})

	res, err := client.Sync.GetMainData(0)
	require.NoError(t, err)

	for _, torrent := range res.Torrents {
		t.Logf("%+v", torrent)
	}

	res2, err := client.Sync.GetMainData(res.RID)
	require.NoError(t, err)

	require.Equal(t, 0, len(res2.Torrents))

	for _, torrent := range res2.Torrents {
		t.Logf("%+v", torrent)
	}

}

func TestGetActiveTorrents(t *testing.T) {
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
		Filter: "active",
	})
	require.NoError(t, err)

	for _, torrent := range res {
		t.Logf("%+v", torrent.Name)
	}

}
