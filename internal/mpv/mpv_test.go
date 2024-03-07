package mpv

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var testFilePath = "E:\\ANIME\\Sousou no Frieren\\[SubsPlease] Sousou no Frieren - 01 (1080p) [F02B9CEE].mkv"

func TestMpv_OpenAndPlay(t *testing.T) {

	m := New(util.NewLogger(), "", "")

	err := m.OpenAndPlay(testFilePath, StartExecCommand)
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

	m := New(util.NewLogger(), "", "C:\\Program Files\\mpv.net\\mpvnet.exe")

	err := m.OpenAndPlay(testFilePath, StartExec)
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

	m := New(util.NewLogger(), "", "")

	err := m.OpenAndPlay(testFilePath, StartExecCommand)
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

	m := New(util.NewLogger(), "", "")

	//sub := m.Subscribe("test")

	err := m.OpenAndPlay(testFilePath, StartExecCommand)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(4 * time.Second)

	err = m.OpenAndPlay(testFilePath, StartExecCommand)
	if assert.Error(t, err, "mpv instance should not be initialized twice") {
		t.Log("mpv instance already initialized")
	}

	t.Log("Tried to open mpv instance twice")

	time.Sleep(4 * time.Second)

	t.Log("Done")

}
