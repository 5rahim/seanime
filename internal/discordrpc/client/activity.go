package discordrpc_client

import (
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// Activity holds the data for discord rich presence
//
// See https://discord.com/developers/docs/game-sdk/activities#data-models-activity-struct
type Activity struct {
	Name       string `json:"name,omitempty"`
	Details    string `json:"details,omitempty"`
	DetailsURL string `json:"details_url,omitempty"` // URL to details
	State      string `json:"state,omitempty"`
	StateURL   string `json:"state_url,omitempty"` // URL to state

	Timestamps *Timestamps `json:"timestamps,omitempty"`
	Assets     *Assets     `json:"assets,omitempty"`
	Party      *Party      `json:"party,omitempty"`
	Secrets    *Secrets    `json:"secrets,omitempty"`
	Buttons    []*Button   `json:"buttons,omitempty"`

	Instance          bool `json:"instance"`
	Type              int  `json:"type"`
	StatusDisplayType int  `json:"status_display_type,omitempty"` // 1 = name, 2 = details, 3 = state
}

// Timestamps holds unix timestamps for start and/or end of the game
//
// See https://discord.com/developers/docs/game-sdk/activities#data-models-activitytimestamps-struct
type Timestamps struct {
	Start *Epoch `json:"start,omitempty"`
	End   *Epoch `json:"end,omitempty"`
}

// Epoch wrapper around time.Time to ensure times are sent as a unix epoch int
type Epoch struct{ time.Time }

// MarshalJSON converts time.Time to unix time int
func (t Epoch) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(t.Unix(), 10)), nil
}

// Assets passes image references for inclusion in rich presence
//
// See https://discord.com/developers/docs/game-sdk/activities#data-models-activityassets-struct
type Assets struct {
	LargeImage string `json:"large_image,omitempty"`
	LargeText  string `json:"large_text,omitempty"`
	LargeURL   string `json:"large_url,omitempty"` // URL to large image, if any
	SmallImage string `json:"small_image,omitempty"`
	SmallText  string `json:"small_text,omitempty"`
	SmallURL   string `json:"small_url,omitempty"` // URL to small image, if any
}

// Party holds information for the current party of the player
type Party struct {
	ID   string `json:"id"`
	Size []int  `json:"size"` // seems to be element [0] is count and [1] is max
}

// Secrets holds secrets for Rich Presence joining and spectating
type Secrets struct {
	Join     string `json:"join,omitempty"`
	Spectate string `json:"spectate,omitempty"`
	Match    string `json:"match,omitempty"`
}

type Button struct {
	Label string `json:"label,omitempty"`
	Url   string `json:"url,omitempty"`
}

// SetActivity sets the Rich Presence activity for the running application
func (c *Client) SetActivity(activity Activity) error {
	payload := Payload{
		Cmd: SetActivityCommand,
		Args: Args{
			Pid:      os.Getpid(),
			Activity: &activity,
		},
		Nonce: uuid.New(),
	}
	return c.SendPayload(payload)
}

func (c *Client) CancelActivity() error {
	payload := Payload{
		Cmd: SetActivityCommand,
		Args: Args{
			Pid:      os.Getpid(),
			Activity: nil,
		},
		Nonce: uuid.New(),
	}
	return c.SendPayload(payload)
}
