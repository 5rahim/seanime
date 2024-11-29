package qbittorrent

import (
	"errors"
	"runtime"
	"seanime/internal/util"
	"time"
)

func (c *Client) getExecutableName() string {
	switch runtime.GOOS {
	case "windows":
		return "qbittorrent.exe"
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
		return "C:/Program Files/qBittorrent/qbittorrent.exe"
	case "linux":
		return "/usr/bin/qbittorrent" // Default path for Client on most Linux distributions
	case "darwin":
		return "/Applications/qbittorrent.app/Contents/MacOS/qbittorrent" // Default path for Client on macOS
	default:
		return "C:/Program Files/qBittorrent/qbittorrent.exe"
	}
}

func (c *Client) Start() error {

	// If the path is empty, do not check if qBittorrent is running
	if c.Path == "" {
		return nil
	}

	name := c.getExecutableName()
	if util.ProgramIsRunning(name) {
		return nil
	}

	exe := c.getExecutablePath()
	cmd := util.NewCmd(exe)
	err := cmd.Start()
	if err != nil {
		return errors.New("failed to start qBittorrent")
	}

	time.Sleep(1 * time.Second)

	return nil
}

func (c *Client) CheckStart() bool {
	if c == nil {
		return false
	}

	// If the path is empty, assume it's running
	if c.Path == "" {
		return true
	}

	_, err := c.Application.GetAppVersion()
	if err == nil {
		return true
	}

	err = c.Start()
	timeout := time.After(30 * time.Second)
	ticker := time.Tick(1 * time.Second)
	for {
		select {
		case <-ticker:
			_, err = c.Application.GetAppVersion()
			if err == nil {
				return true
			}
		case <-timeout:
			return false
		}
	}
}
