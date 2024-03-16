package onlinestream_sources

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGogoCDN_Extract(t *testing.T) {
	gogo := NewGogoCDN()

	ret, err := gogo.Extract("https://embtaku.pro/streaming.php?id=MjExNjU5&title=One+Piece+Episode+1075")
	assert.NoError(t, err)

	spew.Dump(ret)
}
