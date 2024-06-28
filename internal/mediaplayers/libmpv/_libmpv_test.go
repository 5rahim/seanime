package libmpv

import (
	"fmt"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var testFilePath = "E:\\ANIME\\Dungeon Meshi\\[Lazy] Dungeon Meshi - 01 (WEB 1080p EAC3 5.1) [Dual Audio] [2669B61B].mkv"

func TestMpv_Multiple(t *testing.T) {
	//test_utils.InitTestProvider(t, test_utils.MediaPlayer())

	fmt.Println("Using cgo:", isUsingCgo())

	m := New(util.NewLogger())
	defer m.CloseAll()

	err := m.OpenAndPlay(testFilePath)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(2 * time.Second)

	err = m.OpenAndPlay(testFilePath)
	if !assert.NoError(t, err) {
		t.Fatal("error opening mpv instance twice")
	}

	sub := m.Subscribe("test")

	go func() {
		time.Sleep(5 * time.Second)
		m.CloseAll()
	}()

loop:
	for {
		select {
		case v, _ := <-sub.Done():
			t.Logf("mpv exited, %+v", v)
			break loop
		default:
			time.Sleep(3 * time.Second)
			status, err := m.GetPlaybackStatus()
			if err != nil {
				t.Errorf("error getting playback status: %v", err)
				break loop
			}
			t.Logf("Playback status: %+v", status)
		}
	}

	t.Log("Done")
}
