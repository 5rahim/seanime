package transmission

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/hekmon/transmissionrpc/v3"
	"github.com/seanime-app/seanime/internal/test_utils"
	"github.com/seanime-app/seanime/internal/torrent"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGetFiles(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.TorrentClient())

	tempDir := t.TempDir()

	tests := []struct {
		name            string
		url             string
		mediaId         int
		expectedNbFiles int
	}{
		{
			name:            "[EMBER] Demon Slayer (2023) (Season 3)",
			url:             "https://animetosho.org/view/ember-demon-slayer-2023-season-3-bdrip-1080p.n1778316",
			mediaId:         145139,
			expectedNbFiles: 11,
		},
		{
			name:            "[Tenrai-Sensei] Kakegurui (Season 1-2 + OVAs)",
			url:             "https://nyaa.si/view/1553978",
			mediaId:         98314,
			expectedNbFiles: 27,
		},
	}

	trans, err := New(&NewTransmissionOptions{
		Host:     test_utils.ConfigData.Provider.TransmissionHost,
		Path:     test_utils.ConfigData.Provider.TransmissionPath,
		Port:     test_utils.ConfigData.Provider.TransmissionPort,
		Username: test_utils.ConfigData.Provider.TransmissionUsername,
		Password: test_utils.ConfigData.Provider.TransmissionPassword,
		Logger:   util.NewLogger(),
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			// Get magnet
			magnet, found := torrent.ScrapeMagnet(tt.url)

			if assert.NoError(t, found) {

				to, err := trans.Client.TorrentAdd(context.Background(), transmissionrpc.TorrentAddPayload{
					Filename:    &magnet,
					DownloadDir: &tempDir,
				})

				if assert.NoError(t, err) {

					time.Sleep(20 * time.Second)

					// Get files
					torrents, err := trans.Client.TorrentGetAllFor(context.Background(), []int64{*to.ID})
					to = torrents[0]

					spew.Dump(to.Files)

					// Remove torrent
					err = trans.Client.TorrentRemove(context.Background(), transmissionrpc.TorrentRemovePayload{
						IDs:             []int64{*to.ID},
						DeleteLocalData: true,
					})

					assert.NoError(t, err)

				}

			}

		})

	}

}
