package qbittorrent

import (
	"github.com/seanime-app/seanime/internal/util"
	"testing"
)

func TestNewClient(t *testing.T) {

	client := NewClient(&NewClientOptions{
		Logger:   util.NewLogger(),
		Username: "admin",
		Password: "adminadmin",
		Port:     8081,
		Host:     "127.0.0.1",
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
