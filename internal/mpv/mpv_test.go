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

	select {
	case v, _ := <-m.ExitCh:
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

	select {
	case v, _ := <-m.ExitCh:
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

loop:
	for {
		select {
		case v, _ := <-m.ExitCh:
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

	err := m.OpenAndPlay(testFilePath, StartExecCommand)
	if err != nil {
		t.Fatal(err)
	}

	err = m.OpenAndPlay(testFilePath, StartExecCommand)
	if assert.Error(t, err, "mpv instance should not be initialized twice") {
		t.Log("mpv instance already initialized")
	}

	t.Log("Tried to open mpv instance twice")

	select {
	case v, _ := <-m.ExitCh:
		t.Logf("mpv exited, %+v", v)
		break
	}

	t.Log("Done")

}

func TestMpv_CloseReopen(t *testing.T) {

	m := New(util.NewLogger(), "", "")

	err := m.OpenAndPlay(testFilePath, StartExecCommand)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		<-time.After(2 * time.Second)
		m.Close()
	}()

	select {
	case v, _ := <-m.ExitCh:
		t.Logf("mpv exited, %+v", v)
		break
	}

	time.Sleep(1 * time.Second) // Wait a second before reopening

	err = m.OpenAndPlay(testFilePath, StartExecCommand)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		<-time.After(2 * time.Second)
		m.Close()
	}()

	select {
	case v, _ := <-m.ExitCh:
		t.Logf("mpv exited again, %+v", v)
		break
	}

	t.Log("Done")

}
