package discordrpc_ipc

import (
	"bytes"
	"encoding/binary"
	"net"
	"strings"
	"testing"
)

func frame(opcode int, payload string) []byte {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, int32(opcode))
	_ = binary.Write(buf, binary.LittleEndian, int32(len(payload)))
	buf.WriteString(payload)
	return buf.Bytes()
}

func TestRead(t *testing.T) {
	ready := `{"cmd":"DISPATCH","data":{"v":1,"config":{"cdn_host":"cdn.discordapp.com","api_endpoint":"//discord.com/api","environment":"production"},"user":{"id":"123456789012345678","username":"user","discriminator":"0","global_name":"User","avatar":"a_0123456789abcdef0123456789abcdef","avatar_decoration_data":{"asset":"a_fedcba9876543210fedcba9876543210","sku_id":"1144058844004233369","expires_at":null},"bot":false,"flags":0,"premium_type":2}},"evt":"READY","nonce":null,"padding":"` + strings.Repeat("x", 300) + `"}`
	activity := `{"cmd":"SET_ACTIVITY","data":{"name":"Seanime"},"evt":null,"nonce":"abc"}`

	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	go func() {
		// split the first frame and send the second right after it
		data := append(frame(1, ready), frame(1, activity)...)
		for len(data) > 0 {
			n := min(100, len(data))
			_, _ = server.Write(data[:n])
			data = data[n:]
		}
	}()

	socket := &Socket{client}
	for _, want := range []string{ready, activity} {
		got, err := socket.Read()
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	}
}
