package anizip

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFetchAniZipMedia(t *testing.T) {

	media, err := FetchAniZipMedia("anilist", 1)

	if assert.NoError(t, err) {
		t.Log(spew.Sdump(media))
	}

}
