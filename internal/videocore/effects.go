package videocore

import (
	"seanime/internal/continuity"
	"seanime/internal/discordrpc/presence"
	"seanime/internal/events"
	"seanime/internal/mkvparser"

	"github.com/samber/lo"
)

func (vc *VideoCore) setupEffects() {
	vc.setupSharedEffects()
	vc.setupOnlinestreamEffects()
}

func (vc *VideoCore) setupSharedEffects() {
	subscriber := vc.Subscribe("video-core:shared")

	go func(subscriber *Subscriber) {
		for e := range subscriber.Events() {
			switch event := e.(type) {
			case *VideoPausedEvent:
				if vc.discordPresence != nil && !vc.isOfflineRef.Get() {
					go vc.discordPresence.UpdateAnimeActivity(int(event.CurrentTime), int(event.Duration), true)
				}
			case *VideoResumedEvent:
				if vc.discordPresence != nil && !vc.isOfflineRef.Get() {
					go vc.discordPresence.UpdateAnimeActivity(int(event.CurrentTime), int(event.Duration), false)
				}
			case *VideoEndedEvent:
				if vc.discordPresence != nil && !vc.isOfflineRef.Get() {
					go vc.discordPresence.Close()
				}
			case *VideoLoadedMetadataEvent:
				state, ok := vc.GetPlaybackState()
				if !ok {
					continue
				}
				if vc.discordPresence != nil && !vc.isOfflineRef.Get() {
					go vc.discordPresence.SetAnimeActivity(&discordrpc_presence.AnimeActivity{
						ID:            state.PlaybackInfo.Media.GetID(),
						Title:         state.PlaybackInfo.Media.GetPreferredTitle(),
						Image:         state.PlaybackInfo.Media.GetCoverImageSafe(),
						IsMovie:       state.PlaybackInfo.Media.IsMovie(),
						EpisodeNumber: state.PlaybackInfo.Episode.EpisodeNumber,
						Progress:      int(event.CurrentTime),
						Duration:      int(event.Duration),
					})
				}
			case *VideoErrorEvent:
				if vc.discordPresence != nil && !vc.isOfflineRef.Get() {
					go vc.discordPresence.Close()
				}
			case *VideoTerminatedEvent:
				if vc.discordPresence != nil && !vc.isOfflineRef.Get() {
					go vc.discordPresence.Close()
				}
			case *VideoStatusEvent:
				state, ok := vc.GetPlaybackState()
				if !ok {
					continue
				}
				if event.Duration != 0 {
					_ = vc.continuityManager.UpdateWatchHistoryItem(&continuity.UpdateWatchHistoryItemOptions{
						CurrentTime:   event.CurrentTime,
						Duration:      event.Duration,
						MediaId:       state.PlaybackInfo.Media.GetID(),
						EpisodeNumber: state.PlaybackInfo.Episode.GetEpisodeNumber(),
						Kind:          continuity.MediastreamKind,
					})
				}

				if vc.discordPresence != nil && !vc.isOfflineRef.Get() {
					go vc.discordPresence.UpdateAnimeActivity(int(event.CurrentTime), int(event.Duration), event.Paused)
				}
			}
		}
	}(subscriber)
}

func (vc *VideoCore) setupOnlinestreamEffects() {
	subscriber := vc.Subscribe("video-core:onlinestream")

	go func(subscriber *Subscriber) {
		for e := range subscriber.Events() {
			if !e.IsOnlinestream() && !e.IsWebPlayer() {
				continue
			}
			switch event := e.(type) {
			case *SubtitleFileUploadedEvent:
				vc.logger.Trace().Msgf("videocore: Subtitle file uploaded: %s", event.Filename)
				// Send WebVTT track for VideoCore MediaCaptionsManager
				mkvTrack, err := vc.GenerateMkvSubtitleTrack(GenerateSubtitleFileOptions{
					Filename:  event.Filename,
					Content:   event.Content,
					Number:    0,
					ConvertTo: mkvparser.SubtitleTypeWEBVTT,
				})
				if err != nil {
					vc.wsEventManager.SendEventTo(vc.GetCurrentClientId(), events.ErrorToast, "Failed to upload subtitle file: "+err.Error())
					continue
				}
				track := &VideoSubtitleTrack{
					Index:             0,
					Src:               nil,
					Content:           &mkvTrack.CodecPrivate,
					Label:             mkvTrack.Name,
					Language:          mkvTrack.Language,
					Type:              lo.ToPtr("vtt"),
					Default:           lo.ToPtr(false),
					UseLibassRenderer: nil,
				}
				vc.AddExternalSubtitleTrack(track)
				// Send ASS track for VideoCore SubtitleManager
				mkvTrack, err = vc.GenerateMkvSubtitleTrack(GenerateSubtitleFileOptions{
					Filename:  event.Filename,
					Content:   event.Content,
					Number:    0,
					ConvertTo: mkvparser.SubtitleTypeASS,
				})
				if err != nil {
					vc.wsEventManager.SendEventTo(vc.GetCurrentClientId(), events.ErrorToast, "Failed to upload subtitle file: "+err.Error())
					continue
				}
				track = &VideoSubtitleTrack{
					Index:             0,
					Src:               nil,
					Content:           &mkvTrack.CodecPrivate,
					Label:             mkvTrack.Name,
					Language:          mkvTrack.Language,
					Type:              lo.ToPtr("ass"),
					Default:           lo.ToPtr(false),
					UseLibassRenderer: nil,
				}
				vc.AddExternalSubtitleTrack(track)
				vc.logger.Debug().Msgf("videocore: Sent converted subtitle tracks")
			}
		}
	}(subscriber)
}
