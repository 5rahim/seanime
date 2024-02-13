package mpv_player

import (
	"github.com/davecgh/go-spew/spew"
	"testing"
	"time"
)

// var testFilePath = "E:/ANIME\\Sousou no Frieren\\[SubsPlease] Sousou no Frieren - 01 (1080p) [F02B9CEE].mkv"
var testFilePath = "E:/ANIME/One Piece/[SubsPlease] One Piece - 1092 (1080p) [507B5014].mkv"

func TestMpvPlayer_OpenAndPlay(t *testing.T) {

	m := New()

	m.Start()

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
		case <-m.ExitCh:
			break loop
		case <-time.After(3 * time.Second):
			spew.Dump(m.Playback)
		}
	}

	t.Log("TestMpvPlayer_OpenAndPlay: Done")

}
