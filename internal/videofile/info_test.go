package videofile

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMediaInfoExtractor_GetInfo(t *testing.T) {

	//filep := "E:/ANIME/[Judas] Blue Lock (Season 1) [1080p][HEVC x265 10bit][Dual-Audio][Multi-Subs]/[Judas] Blue Lock - S01E03v2.mkv"
	filep := "E:\\COLLECTION\\Dungeon Meshi\\[EMBER] Dungeon Meshi - 04.mkv"

	me, err := NewMediaInfoExtractor(filep, "")

	if assert.NoError(t, err) {

		info, err := me.GetInfo()
		if assert.NoError(t, err) {

			spew.Dump(info)

		}

	}

}
