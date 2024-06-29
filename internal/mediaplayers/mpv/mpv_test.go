package mpv

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/seanime-app/seanime/internal/test_utils"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var testFilePath = "E:\\ANIME\\Sousou no Frieren\\[SubsPlease] Sousou no Frieren - 01 (1080p) [F02B9CEE].mkv"

func TestMpv_OpenAndPlay(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.MediaPlayer())

	m := New(util.NewLogger(), "", "")

	err := m.OpenAndPlay(testFilePath)
	if err != nil {
		t.Fatal(err)
	}

	sub := m.Subscribe("test")

	select {
	case v, _ := <-sub.Done():
		t.Logf("mpv exited, %+v", v)
		break
	}

	t.Log("Done")

}

func TestMpv_OpenAndPlayPath(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.MediaPlayer())

	m := New(util.NewLogger(), "", test_utils.ConfigData.Provider.MpvPath)

	err := m.OpenAndPlay(testFilePath)
	if err != nil {
		t.Fatal(err)
	}

	sub := m.Subscribe("test")

	select {
	case v, _ := <-sub.Done():
		t.Logf("mpv exited, %+v", v)
		break
	}

	t.Log("Done")

}

func TestMpv_Playback(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.MediaPlayer())

	m := New(util.NewLogger(), "", "")

	err := m.OpenAndPlay(testFilePath)
	if err != nil {
		t.Fatal(err)
	}

	sub := m.Subscribe("test")

loop:
	for {
		select {
		case v, _ := <-sub.Done():
			t.Logf("mpv exited, %+v", v)
			break loop
		default:
			spew.Dump(m.GetPlaybackStatus())
			time.Sleep(2 * time.Second)
		}
	}

	t.Log("Done")

}

func TestMpv_Multiple(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.MediaPlayer())

	m := New(util.NewLogger(), "", "")

	err := m.OpenAndPlay(testFilePath)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(2 * time.Second)

	err = m.OpenAndPlay(testFilePath)
	if !assert.NoError(t, err) {
		t.Log("error opening mpv instance twice")
	}

	sub := m.Subscribe("test")

	go func() {
		time.Sleep(2 * time.Second)
		m.CloseAll()
	}()

	select {
	case v, _ := <-sub.Done():
		t.Logf("mpv exited, %+v", v)
		break
	}

	t.Log("Done")

}
