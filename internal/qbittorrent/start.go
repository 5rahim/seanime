package qbittorrent

import (
	"errors"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func (c *Client) getExecutableName() string {
	switch runtime.GOOS {
	case "windows":
		return "qbittorrent.exe"
	case "linux":
		return "qbittorrent"
	case "darwin":
		return "qbittorrent"
	default:
		return "qbittorrent"
	}
}

func (c *Client) getExecutablePath() string {

	if len(c.Path) > 0 {
		return c.Path
	}

	switch runtime.GOOS {
	case "windows":
		return "C:\\Program Files\\VideoLAN\\Client\\c.exe"
	case "linux":
		return "/usr/bin/c" // Default path for Client on most Linux distributions
	case "darwin":
		return "/Applications/Client.app/Contents/MacOS/Client" // Default path for Client on macOS
	default:
		return "C:\\Program Files\\VideoLAN\\Client\\c.exe"
	}
}

func (c *Client) isRunning(executable string) bool {
	cmd := exec.Command("tasklist")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.Contains(string(output), executable)
}

func (c *Client) Start() error {
	name := c.getExecutableName()
	exe := c.getExecutablePath()
	if c.isRunning(name) {
		return nil
	}

	cmd := exec.Command(exe)
	err := cmd.Start()
	if err != nil {
		return errors.New("failed to start Client")
	}

	time.Sleep(1 * time.Second)

	return nil
}
