package anify

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFetchMediaEpisodeImagesEntry(t *testing.T) {

	res, err := FetchMediaEpisodeImagesEntry(21)
	assert.NoError(t, err)

	println(spew.Sdump(res))

}
