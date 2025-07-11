package nakama

import (
	"seanime/internal/events"
	"time"

	"github.com/goccy/go-json"
)

const (
	OnlineStreamStartedEvent        = "online-stream-started" // reported by host when onCanPlay is called
	OnlineStreamPlaybackStatusEvent = "online-stream-playback-status"
)

type OnlineStreamStartedEventPayload struct {
	MediaId       int    `json:"mediaId"`
	EpisodeNumber int    `json:"episodeNumber"`
	Provider      string `json:"provider"`
	Server        string `json:"server"`
	Dubbed        bool   `json:"dubbed"`
	Quality       string `json:"quality"`
}

func (wpm *WatchPartyManager) listenToOnlineStreaming() {
	go func() {
		listener := wpm.manager.wsEventManager.SubscribeToClientNakamaEvents("watch_party")

		for {
			select {
			case <-wpm.sessionCtx.Done():
				wpm.logger.Debug().Msg("nakama: Stopping online stream listener")
				return
			case clientEvent := <-listener.Channel:

				marshaled, _ := json.Marshal(clientEvent.Payload)

				var event NakamaEvent
				err := json.Unmarshal(marshaled, &event)
				if err != nil {
					return
				}

				marshaledPayload, _ := json.Marshal(event.Payload)

				session, ok := wpm.currentSession.Get()
				if !ok {
					continue
				}

				switch event.Type {
				case OnlineStreamStartedEvent:
					wpm.logger.Debug().Msg("nakama: Received online stream started event")

					var payload OnlineStreamStartedEventPayload
					if err := json.Unmarshal(marshaledPayload, &payload); err != nil {
						wpm.logger.Error().Err(err).Msg("nakama: Failed to unmarshal online stream started event")
						return
					}
					wpm.logger.Debug().Interface("payload", payload).Msg("nakama: Received online stream started event")

					newCurrentMediaInfo := &WatchPartySessionMediaInfo{
						MediaId:       payload.MediaId,
						EpisodeNumber: payload.EpisodeNumber,
						AniDBEpisode:  "",
						StreamType:    "online",
						StreamPath:    "",
						OnlineStreamParams: &OnlineStreamParams{
							MediaId:       payload.MediaId,
							Provider:      payload.Provider,
							EpisodeNumber: payload.EpisodeNumber,
							Server:        payload.Server,
							Dubbed:        payload.Dubbed,
							Quality:       payload.Quality,
						},
					}

					session.CurrentMediaInfo = newCurrentMediaInfo

					// Pause immediately and wait for peers to be ready
					//_ = wpm.manager.playbackManager.Pause()
					wpm.sendCommandToOnlineStream(OnlineStreamCommandPause)

					// Reset buffering state for new playback
					wpm.bufferMu.Lock()
					wpm.isWaitingForBuffers = true
					wpm.bufferWaitStart = time.Now()

					// Cancel existing waitForPeersReady goroutine
					if wpm.waitForPeersCancel != nil {
						wpm.waitForPeersCancel()
						wpm.waitForPeersCancel = nil
					}
					wpm.bufferMu.Unlock()

					wpm.broadcastSessionStateToPeers()

					// Start checking peer readiness
					go wpm.waitForPeersReady(func() {
						wpm.sendCommandToOnlineStream(OnlineStreamCommandPlay)
					})
				}
			}
		}
	}()
}

type OnlineStreamCommand string

type OnlineStreamCommandPayload struct {
	Type    OnlineStreamCommand `json:"type"`              // The command type
	Payload interface{}         `json:"payload,omitempty"` // Optional payload for the command
}

const (
	OnlineStreamCommandStart  OnlineStreamCommand = "start" // Start the online stream
	OnlineStreamCommandPlay   OnlineStreamCommand = "play"
	OnlineStreamCommandPause  OnlineStreamCommand = "pause"
	OnlineStreamCommandSeek   OnlineStreamCommand = "seek"
	OnlineStreamCommandSeekTo OnlineStreamCommand = "seekTo" // Seek to a specific time in seconds
)

func (wpm *WatchPartyManager) sendCommandToOnlineStream(cmd OnlineStreamCommand, payload ...interface{}) {
	session, ok := wpm.currentSession.Get()
	if !ok {
		return
	}

	if session.CurrentMediaInfo == nil || session.CurrentMediaInfo.OnlineStreamParams == nil {
		wpm.logger.Warn().Msg("nakama: No online stream params available for sending command")
		return
	}

	commandPayload := OnlineStreamCommandPayload{
		Type:    cmd,
		Payload: nil,
	}

	if len(payload) > 0 {
		commandPayload.Payload = payload[0]
	}

	event := NakamaEvent{
		Type:    OnlineStreamPlaybackStatusEvent,
		Payload: commandPayload,
	}

	wpm.manager.wsEventManager.SendEvent(events.NakamaOnlineStreamEvent, event)
}
