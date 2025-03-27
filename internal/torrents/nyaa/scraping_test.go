package nyaa

import (
	"seanime/internal/util"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
)

func TestTorrentFiles(t *testing.T) {

	files, err := TorrentFiles(util.Decode("aHR0cHM6Ly9ueWFhLnNpL3ZpZXcvMTU0MjA1Nw==")) // durarara complete series
	assert.NoError(t, err)

	t.Log(spew.Sdump(files))
	assert.NotEmpty(t, files)

}

func TestTorrentMagnet(t *testing.T) {

	magnet, err := TorrentMagnet(util.Decode("aHR0cHM6Ly9ueWFhLnNpL3ZpZXcvMTg4Njg4Ng=="))
	assert.NoError(t, err)

	t.Log(magnet)
	assert.NotEmpty(t, magnet)

}

func TestTorrentInfo(t *testing.T) {

	title, a, b, c, fs, d, e, err := TorrentInfo(util.Decode("aHR0cHM6Ly9ueWFhLnNpL3ZpZXcvMTcyNzkyMg=="))
	assert.NoError(t, err)

	t.Logf("Title: %s\n", title)
	t.Logf("Seeders: %d\n", a)
	t.Logf("Leechers: %d\n", b)
	t.Logf("Downloads: %d\n", c)
	t.Logf("Formatted Size: %s\n", fs)
	t.Logf("Info Hash: %s\n", d)
	t.Logf("Download link: %s\n", e)

}

func TestTorrentHash(t *testing.T) {

	hash, err := TorrentHash(util.Decode("aHR0cHM6Ly9ueWFhLnNpL3ZpZXcvMTc0MTY5MQ=="))
	assert.NoError(t, err)

	t.Log(hash)
	assert.NotEmpty(t, hash)

}
