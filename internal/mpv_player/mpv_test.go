package mpv_player

import (
	"github.com/davecgh/go-spew/spew"
	"testing"
	"time"
)

var testFilePath = "E:\\ANIME\\Sousou no Frieren"

func TestMpvPlayer_OpenAndPlay(t *testing.T) {

	m := NewMpvPlayer()

	go func() {
		err := m.OpenAndPlay(testFilePath)
		if err != nil {
			t.Error(err)
			return
		}
	}()

loop:
	for {
		select {
		case <-m.exit:
			t.Log("Exited")
			break loop
		case <-time.After(2 * time.Second):
			spew.Dump(m.Playback)
		}
	}

	t.Log("TestMpvPlayer_OpenAndPlay: Done")

}
