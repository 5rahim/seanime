package transmission

import (
	"context"
	"errors"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func (c *Transmission) getExecutableName() string {
	switch runtime.GOOS {
	case "windows":
		return "transmission-qt.exe"
	default:
		return "transmission-qt"
	}
}

func (c *Transmission) getExecutablePath() string {

	if len(c.Path) > 0 {
		return c.Path
	}

	switch runtime.GOOS {
	case "windows":
		return "C:/Program Files/Transmission/transmission-qt.exe"
	case "linux":
		return "/usr/bin/transmission-qt" // Default path for Transmission on most Linux distributions
	case "darwin":
		return "/Applications/Transmission.app/Contents/MacOS/transmission-qt"
		// Default path for Transmission on macOS
	default:
		return "C:/Program Files/Transmission/transmission-qt.exe"
	}
}

func (c *Transmission) isRunning(executable string) bool {
	cmd := exec.Command("tasklist")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.Contains(string(output), executable)
}

func (c *Transmission) Start() error {
	name := c.getExecutableName()
	exe := c.getExecutablePath()
	if c.isRunning(name) {
		return nil
	}

	cmd := exec.Command(exe)
	err := cmd.Start()
	if err != nil {
		return errors.New("failed to start Transmission")
	}

	time.Sleep(1 * time.Second)

	return nil
}

func (c *Transmission) CheckStart() bool {
	if c == nil {
		return false
	}

	_, _, _, err := c.Client.RPCVersion(context.Background())
	if err == nil {
		return true
	}

	err = c.Start()
	timeout := time.After(30 * time.Second)
	ticker := time.Tick(1 * time.Second)
	for {
		select {
		case <-ticker:
			_, _, _, err := c.Client.RPCVersion(context.Background())
			if err == nil {
				return true
			}
		case <-timeout:
			return false
		}
	}
}
