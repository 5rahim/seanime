package qbittorrent

import (
	"github.com/seanime-app/seanime-server/internal/qbittorrent/model"
	"github.com/seanime-app/seanime-server/internal/util"
	"testing"
)

func TestNewClient(t *testing.T) {

	client := NewClient("http://localhost:8081", util.NewLogger())

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
