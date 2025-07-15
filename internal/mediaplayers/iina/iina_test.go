package iina

import (
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testFilePath = "/Users/rahim/Documents/collection/Bocchi the Rock/[ASW] Bocchi the Rock! - 01 [1080p HEVC][EDC91675].mkv"
var testFilePath2 = "/Users/rahim/Documents/collection/One Piece/[Erai-raws] One Piece - 1072 [1080p][Multiple Subtitle][51CB925F].mkv"

func TestIina_OpenPlayPauseSeekClose(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.MediaPlayer())

	i := New(util.NewLogger(), "", "")

	// Test Open and Play
	t.Log("Open and Play...")
	err := i.OpenAndPlay(testFilePath)
	if err != nil {
		t.Skipf("Skipping test: %v", err)
	}

	// Subscribe to events
	sub := i.Subscribe("test")

	time.Sleep(3 * time.Second)

	t.Log("Get Playback Status...")
	status, err := i.GetPlaybackStatus()
	if err != nil {
		t.Logf("Warning: Could not get playback status: %v", err)
	} else {
		t.Logf("Playback Status: Duration=%.2f, Position=%.2f, Playing=%t, Filename=%s",
			status.Duration, status.Position, !status.Paused, status.Filename)
		assert.True(t, status.IsRunning, "Player should be running")
		assert.Greater(t, status.Duration, 0.0, "Duration should be greater than 0")
	}

	t.Log("Pause...")
	err = i.Pause()
	if err != nil {
		t.Logf("Warning: Could not pause: %v", err)
	} else {
		time.Sleep(2 * time.Second)
		status, err := i.GetPlaybackStatus()
		if err == nil {
			t.Logf("After pause - Paused: %t", status.Paused)
			assert.True(t, status.Paused, "Player should be paused")
		}
	}

	t.Log("Resume...")
	err = i.Resume()
	if err != nil {
		t.Logf("Warning: Could not resume: %v", err)
	} else {
		time.Sleep(2 * time.Second)
		status, err := i.GetPlaybackStatus()
		if err == nil {
			t.Logf("After resume - Paused: %t", status.Paused)
			assert.False(t, status.Paused, "Player should not be paused")
		}
	}

	t.Log("Seek...")
	seekPosition := 30.0 // Seek to 30 seconds
	err = i.Seek(seekPosition)
	if err != nil {
		t.Logf("Warning: Could not seek: %v", err)
	} else {
		time.Sleep(2 * time.Second)
		status, err := i.GetPlaybackStatus()
		if err == nil {
			t.Logf("After seek - Position: %.2f", status.Position)
			assert.InDelta(t, seekPosition, status.Position, 5.0, "Position should be close to seek position")
		}
	}

	t.Log("SeekTo...")
	seekToPosition := 60.0 // Seek to 60 seconds
	err = i.SeekTo(seekToPosition)
	if err != nil {
		t.Logf("Warning: Could not seek to position: %v", err)
	} else {
		time.Sleep(2 * time.Second)
		status, err := i.GetPlaybackStatus()
		if err == nil {
			t.Logf("After seekTo - Position: %.2f", status.Position)
			assert.InDelta(t, seekToPosition, status.Position, 5.0, "Position should be close to seekTo position")
		}
	}

	// Test loading another file
	t.Log("Open another file...")
	err = i.OpenAndPlay(testFilePath2)
	if err != nil {
		t.Logf("Warning: Could not open another file: %v", err)
	} else {
		time.Sleep(2 * time.Second) // Wait for the new file to load
		status, err := i.GetPlaybackStatus()
		if err != nil {
			t.Logf("Warning: Could not get playback status after opening another file: %v", err)
		} else {
			t.Logf("New Playback Status: Duration=%.2f, Position=%.2f, Playing=%t, Filename=%s",
				status.Duration, status.Position, !status.Paused, status.Filename)
			assert.True(t, status.IsRunning, "Player should be running after opening another file")
			assert.Greater(t, status.Duration, 0.0, "Duration should be greater than 0 after opening another file")
		}
	}

	// Test Close
	t.Log("Close...")
	go func() {
		time.Sleep(2 * time.Second)
		i.CloseAll()
	}()

	// Wait for close event
	select {
	case <-sub.Closed():
		t.Log("IINA exited successfully")
	case <-time.After(10 * time.Second):
		t.Log("Timeout waiting for IINA to close")
		i.CloseAll() // Force close
	}

	// Verify player is not running
	time.Sleep(1 * time.Second)
	status, err = i.GetPlaybackStatus()
	if err != nil {
		t.Log("Confirmed: Player is no longer running")
	} else if status != nil && !status.IsRunning {
		t.Log("Confirmed: Player status shows not running")
	}

	t.Log("Test completed successfully")
}
