package vlc

import (
	"github.com/seanime-app/seanime-server/internal/util"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestVLC_Play(t *testing.T) {

	vlc := &VLC{
		Host:     "localhost",
		Port:     8080,
		Password: "seanime",
		Logger:   util.NewLogger(),
	}

	err := vlc.Start()
	assert.Nil(t, err)

	err = vlc.AddAndPlay("E:\\Anime\\[Judas] Golden Kamuy (Seasons 1-2) [BD 1080p][HEVC x265 10bit][Eng-Subs]\\[Judas] Golden Kamuy - S2\\[Judas] Golden Kamuy S2 - 16.mkv")

	if err != nil {
		t.Fatal(err)
	}

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

func TestVLC_Seek(t *testing.T) {

	vlc := &VLC{
		Host:     "localhost",
		Port:     8080,
		Password: "seanime",
		Logger:   util.NewLogger(),
	}

	err := vlc.Start()
	assert.Nil(t, err)

	err = vlc.AddAndPlay("E:\\ANIME\\Violet.Evergarden.The.Movie.1080p.Dual.Audio.BDRip.10.bits.DD.x265-EMBER.mkv")

	time.Sleep(400 * time.Millisecond)

	vlc.ForcePause()

	time.Sleep(400 * time.Millisecond)

	vlc.Seek("100")

	time.Sleep(400 * time.Millisecond)

	status, err := vlc.GetStatus()
	assert.Nil(t, err)

	assert.Equal(t, "paused", status.State)

	if err != nil {
		t.Fatal(err)
	}

}
