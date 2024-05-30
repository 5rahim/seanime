package vlc

import (
	"errors"
	"os/exec"
	"runtime"
	"strings"
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

func (vlc *VLC) getExecutablePath() string {

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

func (vlc *VLC) isRunning(executable string) bool {
	cmd := exec.Command("tasklist")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// TODO
	//var cmd *exec.Cmd
	//switch runtime.GOOS {
	//case "windows":
	//	cmd = exec.Command("tasklist")
	//case "linux":
	//	cmd = exec.Command("pgrep", executable)
	//case "darwin":
	//	cmd = exec.Command("pgrep", executable)
	//default:
	//	return false
	//}

	return strings.Contains(string(output), executable)
}

func (vlc *VLC) Start() error {
	name := vlc.getExecutableName()
	exe := vlc.getExecutablePath()
	if vlc.isRunning(name) {
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
