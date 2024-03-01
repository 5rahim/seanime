package streamer

import (
	"github.com/davecgh/go-spew/spew"
	"testing"
)

func TestTranscoder(t *testing.T) {

	transcoder, err := NewTranscoder()
	if err != nil {
		t.Fatal(err)
	}

	spew.Dump(transcoder.GetMaster("E:\\COLLECTION\\Dungeon Meshi\\[EMBER] Dungeon Meshi - 04.mkv", "client", "route"))

	ret, err := transcoder.GetVideoIndex("E:\\COLLECTION\\Dungeon Meshi\\[EMBER] Dungeon Meshi - 04.mkv", "480p", "client", "route")
	if err != nil {
		t.Fatal(err)
	}

	spew.Dump(ret)
}
