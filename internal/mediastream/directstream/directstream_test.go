package directstream

import (
	"github.com/seanime-app/seanime/internal/util"
	"testing"
)

func TestDirectStream_CopyToHLS(t *testing.T) {

	ds := NewDirectStream(util.NewLogger())

	opts := &CopyToHLSOptions{
		Filepath:         "E:\\ANIME\\Dungeon Meshi\\[EMBER] Dungeon Meshi - 15.mkv",
		OutDir:           "test",
		AudioStreamIndex: 1,
	}

	ds.CopyToHLS(opts)

}
