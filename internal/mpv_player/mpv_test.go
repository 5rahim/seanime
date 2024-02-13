package mpv_player

import (
	"testing"
	"time"
)

var testFilePath = "E:\\ANIME\\Sousou no Frieren\\[SubsPlease] Sousou no Frieren - 01 (1080p) [F02B9CEE].mkv"

func TestMpvPlayer_OpenAndPlay(t *testing.T) {

	m := NewMpvPlayer()

	quit := make(chan struct{})

	go func() {
		err := m.OpenAndPlay(testFilePath)
		if err != nil {
			t.Error(err)
			return
		}
		select {
		case <-quit:
			return
		}
	}()

	go func() {
		select {
		case <-m.paused:
			t.Log("paused")
		case <-quit:
			return
		}
	}()

	time.Sleep(5 * time.Second)
	close(quit)

}
