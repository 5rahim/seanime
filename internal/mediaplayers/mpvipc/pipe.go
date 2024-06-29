//go:build !windows
// +build !windows

package mpvipc

import "net"

func dial(path string) (net.Conn, error) {
	return net.Dial("unix", path)
}
