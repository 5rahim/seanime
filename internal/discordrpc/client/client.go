package discordrpc_client

import (
	"fmt"
	"github.com/goccy/go-json"
	"seanime/internal/discordrpc/ipc"
)

// Client wrapper for the Discord RPC client
type Client struct {
	ClientID string
	Socket   *discordrpc_ipc.Socket
}

func (c *Client) Close() {
	if c == nil {
		return
	}
	c.Socket.Close()
}

// New sends a handshake in the socket and returns an error or nil and an instance of Client
func New(clientId string) (*Client, error) {
	if clientId == "" {
		return nil, fmt.Errorf("no clientId set")
	}

	payload, err := json.Marshal(handshake{"1", clientId})
	if err != nil {
		return nil, err
	}

	sock, err := discordrpc_ipc.NewConnection()
	if err != nil {
		return nil, err
	}

	c := &Client{Socket: sock, ClientID: clientId}

	r, err := c.Socket.Send(0, string(payload))
	if err != nil {
		return nil, err
	}

	var responseBody Data
	if err := json.Unmarshal([]byte(r), &responseBody); err != nil {
		return nil, err
	}

	if responseBody.Code > 1000 {
		return nil, fmt.Errorf(responseBody.Message)
	}

	return c, nil
}
