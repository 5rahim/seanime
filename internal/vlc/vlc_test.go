package vlc

import (
	"github.com/seanime-app/seanime-server/internal/util"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestVLC_Play(t *testing.T) {

	vlc := NewVLC(&NewVLCOptions{
		Host:     "localhost",
		Port:     8080,
		Password: "seanime",
		Logger:   util.NewLogger(),
	})

	err := vlc.StartVLC()
	assert.Nil(t, err)

	err = vlc.AddStart("E:\\ANIME\\Violet.Evergarden.The.Movie.1080p.Dual.Audio.BDRip.10.bits.DD.x265-EMBER.mkv")
	time.Sleep(400 * time.Millisecond)
	vlc.ForcePause()
	time.Sleep(400 * time.Millisecond)

	status, err := vlc.GetStatus()
	assert.Nil(t, err)

	assert.Equal(t, "paused", status.State)

	if err != nil {
		t.Fatal(err)
	}

}
