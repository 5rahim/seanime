package mpv

import (
	"errors"
	"seanime/internal/mediaplayers/mpvipc"
	"seanime/internal/util"
	"testing"
)

func TestNewUsesGeneratedSocketForLegacyDefault(t *testing.T) {
	legacySocket := getLegacySocketName()
	player := New(util.NewLogger(), legacySocket, "")

	if !player.autoSocket {
		t.Fatal("expected legacy default socket to opt into generated sockets")
	}
	if player.SocketName == "" {
		t.Fatal("expected generated socket name")
	}
	if player.SocketName == legacySocket {
		t.Fatalf("expected generated socket instead of legacy socket %q", legacySocket)
	}
}

func TestGetPlaybackStatusFileSwitch(t *testing.T) {
	player := New(util.NewLogger(), "", "")

	player.playbackMu.Lock()
	player.Playback = &Playback{
		Filename:  "episode-01.mkv",
		Filepath:  "/anime/episode-01.mkv",
		Position:  1410,
		Duration:  1440,
		IsRunning: true,
	}
	player.playbackMu.Unlock()

	player.applyPlaybackEvent(&mpvipc.Event{Name: "start-file"})
	player.applyPlaybackEvent(&mpvipc.Event{ID: 45, Data: "episode-02.mkv"})
	player.applyPlaybackEvent(&mpvipc.Event{ID: 46, Data: "/anime/episode-02.mkv"})
	player.applyPlaybackEvent(&mpvipc.Event{ID: 42, Data: 1320.0})

	status, err := player.GetPlaybackStatus()
	if err != nil {
		t.Fatalf("expected playback status during transition: %v", err)
	}
	if status.Filename != "episode-02.mkv" {
		t.Fatalf("expected switched filename, got %q", status.Filename)
	}
	if status.Position != 0 {
		t.Fatalf("expected position to stay reset while the new duration is pending, got %v", status.Position)
	}

	player.applyPlaybackEvent(&mpvipc.Event{ID: 44, Data: 1440.0})

	status, err = player.GetPlaybackStatus()
	if err != nil {
		t.Fatalf("expected playback status after transition settled: %v", err)
	}
	if status.Position != 1320 {
		t.Fatalf("expected fresh position after duration update, got %v", status.Position)
	}
}

func TestFileReadyUsesObservedDuration(t *testing.T) {
	player := New(util.NewLogger(), "", "")
	player.applyPlaybackEvent(&mpvipc.Event{ID: 44, Data: 1440.0})

	if !player.isFileReady() {
		t.Fatal("expected observed duration to mark the file ready")
	}

	player.applyPlaybackEvent(&mpvipc.Event{Name: "start-file"})
	if player.isFileReady() {
		t.Fatal("expected start-file to ignore the previous duration")
	}
}

func TestExtractLaunchErrorMessage(t *testing.T) {
	content := "[   0.036][e][ipc] Could not bind IPC socket\n[   0.040][e][file] Cannot open file '/tmp/missing.mkv': No such file or directory\n[   0.041][e][stream] Failed to open /tmp/missing.mkv.\n"

	message := extractLaunchErrorMessage(content)
	if message != "Cannot open file '/tmp/missing.mkv': No such file or directory" {
		t.Fatalf("unexpected launch error message %q", message)
	}
}

func TestFormatLaunchExitErrorFallsBackToWaitError(t *testing.T) {
	waitErr := errors.New("exit status 2")
	err := formatLaunchExitError(waitErr, "/path/that/does/not/exist")
	if !errors.Is(err, waitErr) {
		t.Fatalf("expected wrapped wait error, got %v", err)
	}
	if err.Error() != waitErr.Error() {
		t.Fatalf("expected fallback wait error, got %q", err.Error())
	}
}
