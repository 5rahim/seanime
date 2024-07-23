package qbittorrent

import (
	"github.com/stretchr/testify/assert"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"testing"
)

func TestClient_Start(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.TorrentClient())

	client := NewClient(&NewClientOptions{
		Logger:   util.NewLogger(),
		Username: test_utils.ConfigData.Provider.QbittorrentUsername,
		Password: test_utils.ConfigData.Provider.QbittorrentPassword,
		Port:     test_utils.ConfigData.Provider.QbittorrentPort,
		Host:     test_utils.ConfigData.Provider.QbittorrentHost,
		Path:     test_utils.ConfigData.Provider.QbittorrentPath,
	})

	err := client.Start()
	assert.Nil(t, err)

}

func TestClient_CheckStart(t *testing.T) {

	client := NewClient(&NewClientOptions{
		Logger:   util.NewLogger(),
		Username: test_utils.ConfigData.Provider.QbittorrentUsername,
		Password: test_utils.ConfigData.Provider.QbittorrentPassword,
		Port:     test_utils.ConfigData.Provider.QbittorrentPort,
		Host:     test_utils.ConfigData.Provider.QbittorrentHost,
		Path:     test_utils.ConfigData.Provider.QbittorrentPath,
	})

	started := client.CheckStart()
	assert.True(t, started)

}
