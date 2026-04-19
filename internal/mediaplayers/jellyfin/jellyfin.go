package jellyfin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/rs/zerolog"
)

type Jellyfin struct {
	ServerUrl string
	ApiKey    string
	UserId    string // hyphens will be stripped before comparison
	Logger    *zerolog.Logger
}

type PlaybackState struct {
	FilePath             string
	CurrentTimeInSeconds float64
	DurationInSeconds    float64
	CompletionPercentage float64 // 0.0–1.0
	IsPaused             bool
}

// sessionsResponse mirrors the subset of Jellyfin's GET /Sessions response we care about.
type sessionsResponse []struct {
	UserId string `json:"UserId"`

	NowPlayingItem *struct {
		Type         string `json:"Type"`
		RunTimeTicks int64  `json:"RunTimeTicks"`
		Path         string `json:"Path"`
	} `json:"NowPlayingItem"`

	PlayState *struct {
		PositionTicks int64 `json:"PositionTicks"`
		IsPaused      bool  `json:"IsPaused"`
	} `json:"PlayState"`

	NowPlayingQueueFullItems []struct {
		Path string `json:"Path"`
	} `json:"NowPlayingQueueFullItems"`
}

// GetPlaybackState polls GET /Sessions and returns the active episode session for the
// configured user, or nil if nothing is playing or the user has no active episode.
func (j *Jellyfin) GetPlaybackState() (*PlaybackState, error) {
	sessions, err := j.fetchSessions()
	if err != nil {
		return nil, err
	}

	wantUserId := strings.ReplaceAll(j.UserId, "-", "")

	for _, s := range sessions {
		if s.NowPlayingItem == nil || s.PlayState == nil {
			continue
		}
		if s.NowPlayingItem.Type != "Episode" {
			continue
		}
		if strings.ReplaceAll(s.UserId, "-", "") != wantUserId {
			continue
		}

		runtimeTicks := s.NowPlayingItem.RunTimeTicks
		if runtimeTicks == 0 {
			continue
		}

		positionTicks := s.PlayState.PositionTicks
		completion := clamp(float64(positionTicks)/float64(runtimeTicks), 0, 1)

		filePath := s.NowPlayingItem.Path
		if len(s.NowPlayingQueueFullItems) > 0 && s.NowPlayingQueueFullItems[0].Path != "" {
			filePath = s.NowPlayingQueueFullItems[0].Path
		}

		return &PlaybackState{
			FilePath:             filePath,
			CurrentTimeInSeconds: float64(positionTicks) / 10_000_000,
			DurationInSeconds:    float64(runtimeTicks) / 10_000_000,
			CompletionPercentage: completion,
			IsPaused:             s.PlayState.IsPaused,
		}, nil
	}

	return nil, nil
}

func (j *Jellyfin) fetchSessions() (sessionsResponse, error) {
	req, err := http.NewRequest("GET", j.ServerUrl+"/Sessions", nil)
	if err != nil {
		return nil, fmt.Errorf("jellyfin: build request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf(`MediaBrowser Token="%s"`, j.ApiKey))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("jellyfin: GET /Sessions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("jellyfin: GET /Sessions returned HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("jellyfin: read response: %w", err)
	}

	var sessions sessionsResponse
	if err := json.Unmarshal(body, &sessions); err != nil {
		return nil, fmt.Errorf("jellyfin: parse response: %w", err)
	}

	return sessions, nil
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
