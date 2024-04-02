//go:build windows
// +build windows

package discordrpc_ipc

import (
	"time"

	"github.com/Microsoft/go-winio"
)

// NewConnection opens the discord-ipc-0 named pipe
func NewConnection() (*Socket, error) {
	// Connect to the Windows named pipe, this is a well known name
	// We use DialTimeout since it will block forever (or very, very long) on Windows
	// if the pipe is not available (Discord not running)
	t := 2 * time.Second
	sock, err := winio.DialPipe(`\\.\pipe\discord-ipc-0`, &t)
	if err != nil {
		return nil, err
	}

	return &Socket{sock}, nil
}
