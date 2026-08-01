package iina

import (
	"seanime/internal/util"
	"testing"
)

func TestFileReadyUsesObservedDuration(t *testing.T) {
	player := New(util.NewLogger(), "", "")
	player.playbackMu.Lock()
	player.Playback.Duration = 1440
	player.freshDuration = true
	player.playbackMu.Unlock()

	if !player.isFileReady() {
		t.Fatal("expected observed duration to mark the file ready")
	}

	player.playbackMu.Lock()
	player.freshDuration = false
	player.playbackMu.Unlock()
	if player.isFileReady() {
		t.Fatal("expected stale duration to be ignored")
	}
}
