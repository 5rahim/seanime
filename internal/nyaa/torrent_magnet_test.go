package nyaa

import (
	"github.com/seanime-app/seanime/internal/torrent"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTorrentMagnet(t *testing.T) {

	magnet, err := TorrentMagnet("https://nyaa.si/view/1741691")
	assert.NoError(t, err)

	t.Log(magnet)
	assert.NotEmpty(t, magnet)

}

func TestExtractHashFromMagnet(t *testing.T) {

	magnet, err := TorrentMagnet("https://nyaa.si/view/1741691")
	assert.NoError(t, err)

	t.Log(magnet)
	assert.NotEmpty(t, magnet)

	hash, ok := torrent.ExtractHashFromMagnet(magnet)
	assert.True(t, ok)

	t.Log(hash)
	assert.NotEmpty(t, hash)

}
