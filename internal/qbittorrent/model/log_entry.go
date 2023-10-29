package qbittorrent_model

import (
	"encoding/json"
	"time"
)

type LogEntry struct {
	ID        int       `json:"id"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Type      LogType   `json:"type"`
}

func (l *LogEntry) UnmarshalJSON(data []byte) error {
	var raw rawLogEntry
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	t := time.Unix(0, int64(raw.Timestamp)*int64(time.Millisecond))
	*l = LogEntry{
		ID:        raw.ID,
		Message:   raw.Message,
		Timestamp: t,
		Type:      raw.Type,
	}
	return nil
}

type LogType int

const (
	TypeNormal LogType = iota << 1
	TypeInfo
	TypeWarning
	TypeCritical
)

type rawLogEntry struct {
	ID        int     `json:"id"`
	Message   string  `json:"message"`
	Timestamp int     `json:"timestamp"`
	Type      LogType `json:"type"`
}
