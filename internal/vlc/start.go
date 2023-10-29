package vlc

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func (vlc *VLC) getVLCExecutableName() string {
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

func (vlc *VLC) getVLCExecutablePath() string {
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

func (vlc *VLC) isVLCRunning(executable string) bool {
	cmd := exec.Command("tasklist")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error checking for VLC:", err)
		return false
	}

	return strings.Contains(string(output), executable)
}

func (vlc *VLC) StartVLC() error {
	name := vlc.getVLCExecutableName()
	exe := vlc.getVLCExecutablePath()
	if vlc.isVLCRunning(name) {
		return nil
	}

	cmd := exec.Command(exe)
	err := cmd.Start()
	if err != nil {
		return errors.New("failed to start VLC")
	}

	time.Sleep(1 * time.Second)

	return nil
}
