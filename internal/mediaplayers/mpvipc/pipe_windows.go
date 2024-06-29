//go:build windows
// +build windows

package mpvipc

import (
	"net"
	"time"

	winio "github.com/Microsoft/go-winio"
)

func dial(path string) (net.Conn, error) {
	timeout := time.Second * 10
	return winio.DialPipe(path, &timeout)
}
