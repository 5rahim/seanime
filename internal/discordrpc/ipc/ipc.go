package discordrpc_ipc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
)

// GetIpcPath chooses the correct directory to the ipc socket and returns it
func GetIpcPath() string {
	vn := []string{"XDG_RUNTIME_DIR", "TMPDIR", "TMP", "TEMP"}

	for _, name := range vn {
		path, exists := os.LookupEnv(name)

		if exists {
			return path
		}
	}

	return "/tmp"
}

// Socket extends net.Conn methods
type Socket struct {
	net.Conn
}

// Read reads one frame: 8-byte header (opcode, length) then payload
func (socket *Socket) Read() (string, error) {
	header := make([]byte, 8)
	if _, err := io.ReadFull(socket.Conn, header); err != nil {
		return "", err
	}

	payload := make([]byte, binary.LittleEndian.Uint32(header[4:]))
	if _, err := io.ReadFull(socket.Conn, payload); err != nil {
		return "", err
	}

	if len(payload) == 0 {
		return "", fmt.Errorf("empty response")
	}

	return string(payload), nil
}

// Send opcode and payload to the unix socket
func (socket *Socket) Send(opcode int, payload string) (string, error) {
	buf := new(bytes.Buffer)

	err := binary.Write(buf, binary.LittleEndian, int32(opcode))
	if err != nil {
		return "", err
	}

	err = binary.Write(buf, binary.LittleEndian, int32(len(payload)))
	if err != nil {
		return "", err
	}

	buf.Write([]byte(payload))
	_, err = socket.Write(buf.Bytes())
	if err != nil {
		return "", err
	}

	return socket.Read()
}
