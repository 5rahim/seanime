package vlc

import (
	"fmt"
	"runtime"
	"seanime/internal/util"
	"time"
)

func (vlc *VLC) getExecutableName() string {
	switch runtime.GOOS {
	case "windows":
		return "vlc.exe"
	case "linux":
		return "vlc"
	case "darwin":
		return "vlc"
	default:
		return "vlc"
	}
}

func (vlc *VLC) GetExecutablePath() string {

	if len(vlc.Path) > 0 {
		return vlc.Path
	}

	switch runtime.GOOS {
	case "windows":
		return "C:\\Program Files\\VideoLAN\\VLC\\vlc.exe"
	case "linux":
		return "/usr/bin/vlc" // Default path for VLC on most Linux distributions
	case "darwin":
		return "/Applications/VLC.app/Contents/MacOS/VLC" // Default path for VLC on macOS
	default:
		return "C:\\Program Files\\VideoLAN\\VLC\\vlc.exe"
	}
}

func (vlc *VLC) Start() error {

	// If the path is empty, do not check if VLC is running
	if vlc.Path == "" {
		return nil
	}

	// Check if VLC is already running
	name := vlc.getExecutableName()
	if util.ProgramIsRunning(name) {
		return nil
	}

	// Start VLC
	exe := vlc.GetExecutablePath()
	cmd := util.NewCmd(exe)
	err := cmd.Start()
	if err != nil {
		vlc.Logger.Error().Err(err).Msg("vlc: Error starting VLC")
		return fmt.Errorf("error starting VLC: %w", err)
	}

	time.Sleep(1 * time.Second)

	return nil
}
