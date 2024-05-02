package transcoder

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/seanime-app/seanime/internal/util"
	"testing"
)

func TestTranscoder(t *testing.T) {

	transcoder, err := NewTranscoder(util.NewLogger())
	if err != nil {
		t.Fatal(err)
	}

	spew.Dump(transcoder.GetMaster("E:\\COLLECTION\\Dungeon Meshi\\[EMBER] Dungeon Meshi - 04.mkv", "client"))

	ret, err := transcoder.GetVideoIndex("E:\\COLLECTION\\Dungeon Meshi\\[EMBER] Dungeon Meshi - 04.mkv", "480p", "client")
	if err != nil {
		t.Fatal(err)
	}

	spew.Dump(ret)
}
