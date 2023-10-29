package qbittorrent_model

import (
	"encoding/json"
	"time"
)

type PeerLogEntry struct {
	ID        int       `json:"id"`
	IP        string    `json:"ip"`
	Timestamp time.Time `json:"timestamp"`
	Blocked   bool      `json:"blocked"`
	Reason    string    `json:"reason"`
}

func (l *PeerLogEntry) UnmarshalJSON(data []byte) error {
	var raw rawPeerLogEntry
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	t := time.Unix(0, int64(raw.Timestamp)*int64(time.Millisecond))
	*l = PeerLogEntry{
		ID:        raw.ID,
		IP:        raw.IP,
		Timestamp: t,
		Blocked:   raw.Blocked,
		Reason:    raw.Reason,
	}
	return nil
}

type rawPeerLogEntry struct {
	ID        int    `json:"id"`
	IP        string `json:"ip"`
	Timestamp int    `json:"timestamp"`
	Blocked   bool   `json:"blocked"`
	Reason    string `json:"reason"`
}
