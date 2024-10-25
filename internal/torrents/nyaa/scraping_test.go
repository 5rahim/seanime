package nyaa

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTorrentFiles(t *testing.T) {

	files, err := TorrentFiles("https://nyaa.si/view/1542057") // durarara complete series
	assert.NoError(t, err)

	t.Log(spew.Sdump(files))
	assert.NotEmpty(t, files)

}

func TestTorrentMagnet(t *testing.T) {

	magnet, err := TorrentMagnet("https://nyaa.si/view/1886886")
	assert.NoError(t, err)

	t.Log(magnet)
	assert.NotEmpty(t, magnet)

}

func TestTorrentInfo(t *testing.T) {

	title, a, b, c, fs, d, e, err := TorrentInfo("https://nyaa.si/view/1727922")
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

	hash, err := TorrentHash("https://nyaa.si/view/1741691")
	assert.NoError(t, err)

	t.Log(hash)
	assert.NotEmpty(t, hash)

}
