package videocore

import (
	"seanime/internal/events"
	"seanime/internal/mkvparser"
)

func (vc *VideoCore) setupEffects() {
	vc.setupSharedEffects()
	vc.setupOnlinestreamEffects()
}

func (vc *VideoCore) setupSharedEffects() {
	// noop
}

func (vc *VideoCore) setupOnlinestreamEffects() {
	subscriber := vc.Subscribe("videocore:onlinestream")

	go func(subscriber *Subscriber) {
		for e := range subscriber.Events() {
			if !e.IsOnlinestream() && !e.IsWebPlayer() {
				continue
			}
			switch event := e.(type) {
			case *SubtitleFileUploadedEvent:
				vc.logger.Trace().Msgf("videocore: Subtitle file uploaded: %s", event.Filename)
				mkvTrack, err := vc.GenerateMkvSubtitleTrack(GenerateSubtitleFileOptions{
					Filename:  event.Filename,
					Content:   event.Content,
					Number:    0,
					ConvertTo: mkvparser.SubtitleTypeASS,
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
					Type:              new("ass"),
					Default:           new(false),
					UseLibassRenderer: nil,
				}
				vc.AddExternalSubtitleTrack(track)
				vc.logger.Debug().Msgf("videocore: Sent converted subtitle tracks")
			}
		}
	}(subscriber)
}
