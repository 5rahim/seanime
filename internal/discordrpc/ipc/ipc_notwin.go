//go:build !windows
// +build !windows

package discordrpc_ipc

import (
	"net"
	"time"
)

// NewConnection opens the discord-ipc-0 unix socket
func NewConnection() (*Socket, error) {
	sock, err := net.DialTimeout("unix", GetIpcPath()+"/discord-ipc-0", time.Second*2)
	if err != nil {
		return nil, err
	}

	return &Socket{sock}, nil
}
