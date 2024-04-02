package discordrpc_client

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ActivityEventData struct {
	Secret string `json:"secret"`
	User   *User  `json:"user"`
}

type event string

var (
	ActivityJoinEvent        event = "ACTIVITY_JOIN"
	ActivitySpectateEvent    event = "ACTIVITY_SPECTATE"
	ActivityJoinRequestEvent event = "ACTIVITY_JOIN_REQUEST"
)

func (c *Client) RegisterEvent(ch chan ActivityEventData, evt event) error {
	if c == nil {
		return nil
	}

	payload := Payload{
		Cmd:   SubscribeCommand,
		Event: evt,
		Nonce: uuid.New(),
	}

	err := c.SendPayload(payload)
	if err != nil {
		return nil
	}

	go func() {
		for {
			r, err := c.Socket.Read()
			if err != nil {
				continue
			}

			var response struct {
				Event event              `json:"event"`
				Data  *ActivityEventData `json:"data"`
			}

			if err := json.Unmarshal([]byte(r), &response); err != nil {
				continue
			}

			if response.Event == evt {
				continue
			}

			ch <- *response.Data

			time.Sleep(10 * time.Millisecond)
		}
	}()

	return nil
}
