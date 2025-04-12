package animetosho

import (
	"seanime/internal/util"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMagnet(t *testing.T) {

	url := util.Decode("aHR0cHM6Ly9hbmltZXRvc2hvLm9yZy92aWV3L2thaXpva3UtanVqdXRzdS1rYWlzZW4tMjYtYTFjOWJhYjEtc2Vhc29uLTIubjE3MTAxMTY=")

	magnet, err := TorrentMagnet(url)

	if assert.NoError(t, err) {
		if assert.NotEmptyf(t, magnet, "magnet link not found") {
			t.Log(magnet)
		}
	}
}

func TestTorrentFile(t *testing.T) {

	url := util.Decode("aHR0cHM6Ly9hbmltZXRvc2hvLm9yZy92aWV3L2thaXpva3UtanVqdXRzdS1rYWlzZW4tMjYtYTFjOWJhYjEtc2Vhc29uLTIubjE3MTAxMTY=")

	link, err := TorrentFile(url)

	if assert.NoError(t, err) {
		if assert.NotEmptyf(t, link, "download link not found") {
			t.Log(link)
		}
	}
}

func TestTorrentHash(t *testing.T) {

	url := util.Decode("aHR0cHM6Ly9hbmltZXRvc2hvLm9yZy92aWV3L2thaXpva3UtanVqdXRzdS1rYWlzZW4tMjYtYTFjOWJhYjEtc2Vhc29uLTIubjE3MTAxMTY=")

	hash, err := TorrentHash(url)

	if assert.NoError(t, err) {
		if assert.NotEmptyf(t, hash, "hash not found") {
			t.Log(hash)
		}
	}
}
