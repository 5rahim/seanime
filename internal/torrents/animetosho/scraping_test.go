package animetosho

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMagnet(t *testing.T) {

	url := "https://animetosho.org/view/kaizoku-jujutsu-kaisen-26-a1c9bab1-season-2.n1710116"

	magnet, err := TorrentMagnet(url)

	if assert.NoError(t, err) {
		if assert.NotEmptyf(t, magnet, "magnet link not found") {
			t.Log(magnet)
		}
	}
}

func TestTorrentFile(t *testing.T) {

	url := "https://animetosho.org/view/kaizoku-jujutsu-kaisen-26-a1c9bab1-season-2.n1710116"

	link, err := TorrentFile(url)

	if assert.NoError(t, err) {
		if assert.NotEmptyf(t, link, "download link not found") {
			t.Log(link)
		}
	}
}

func TestTorrentHash(t *testing.T) {

	url := "https://animetosho.org/view/kaizoku-jujutsu-kaisen-26-a1c9bab1-season-2.n1710116"

	hash, err := TorrentHash(url)

	if assert.NoError(t, err) {
		if assert.NotEmptyf(t, hash, "hash not found") {
			t.Log(hash)
		}
	}
}
