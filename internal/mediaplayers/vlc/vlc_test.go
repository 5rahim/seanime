package vlc

import (
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVLC_Play(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.MediaPlayer())

	vlc := &VLC{
		Host:     test_utils.ConfigData.Provider.VlcHost,
		Port:     test_utils.ConfigData.Provider.VlcPort,
		Password: test_utils.ConfigData.Provider.VlcPassword,
		Path:     test_utils.ConfigData.Provider.VlcPath,
		Logger:   util.NewLogger(),
	}

	err := vlc.Start()
	require.NoError(t, err)

	err = vlc.AddAndPlay("E:\\Anime\\[Judas] Golden Kamuy (Seasons 1-2) [BD 1080p][HEVC x265 10bit][Eng-Subs]\\[Judas] Golden Kamuy - S2\\[Judas] Golden Kamuy S2 - 16.mkv")

	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(400 * time.Millisecond)

	vlc.ForcePause()

	time.Sleep(400 * time.Millisecond)

	status, err := vlc.GetStatus()
	require.NoError(t, err)

	assert.Equal(t, "paused", status.State)

	if err != nil {
		t.Fatal(err)
	}

}

func TestVLC_Seek(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.MediaPlayer())

	vlc := &VLC{
		Host:     test_utils.ConfigData.Provider.VlcHost,
		Port:     test_utils.ConfigData.Provider.VlcPort,
		Password: test_utils.ConfigData.Provider.VlcPassword,
		Path:     test_utils.ConfigData.Provider.VlcPath,
		Logger:   util.NewLogger(),
	}

	err := vlc.Start()
	require.NoError(t, err)

	err = vlc.AddAndPlay("E:\\ANIME\\[SubsPlease] Bocchi the Rock! (01-12) (1080p) [Batch]\\[SubsPlease] Bocchi the Rock! - 01v2 (1080p) [ABDDAE16].mkv")

	time.Sleep(400 * time.Millisecond)

	vlc.ForcePause()

	time.Sleep(400 * time.Millisecond)

	vlc.SeekTo("100")

	time.Sleep(400 * time.Millisecond)

	status, err := vlc.GetStatus()
	require.NoError(t, err)

	assert.Equal(t, "paused", status.State)

	spew.Dump(status)

	if err != nil {
		t.Fatal(err)
	}

}
