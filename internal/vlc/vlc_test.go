package vlc

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestVLC_Play(t *testing.T) {

	vlc := NewVLC("127.0.0.1", 8080, "seanime", nil)

	err := vlc.StartVLC()
	assert.Nil(t, err)

	err = vlc.AddStart("E:\\ANIME\\Violet.Evergarden.The.Movie.1080p.Dual.Audio.BDRip.10.bits.DD.x265-EMBER.mkv")
	time.Sleep(400 * time.Millisecond)
	vlc.ForcePause()

	status, err := vlc.GetStatus()
	assert.Nil(t, err)

	assert.Equal(t, "paused", status.State)

	if err != nil {
		t.Fatal(err)
	}

}
