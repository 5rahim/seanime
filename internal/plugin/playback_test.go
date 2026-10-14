package plugin

import (
	"encoding/json"
	"seanime/internal/library/playbackmanager"
	"seanime/internal/mediaplayers/mediaplayer"
	gojautil "seanime/internal/util/goja"
	"testing"
	"time"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/require"
)

func TestPlaybackEvents(t *testing.T) {
	vm := goja.New()
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
	scheduler := gojautil.NewScheduler()
	t.Cleanup(scheduler.Stop)
	p := &Playback{vm: vm, scheduler: scheduler}
	subscriber := &playbackmanager.PlaybackStatusSubscriber{EventCh: make(chan playbackmanager.PlaybackEvent, 10)}
	received := make(chan string, 10)
	require.NoError(t, vm.Set("report", func(event string) { received <- event }))
	listener, err := vm.RunString(`(event) => {
		if (!event.state || !event.status) return;
		report(JSON.stringify(event));
	}`)
	require.NoError(t, err)
	var callback func(*PlaybackEvent)
	require.NoError(t, vm.ExportTo(listener, &callback))
	go p.listen(subscriber, callback)
	defer close(subscriber.EventCh)

	state := playbackmanager.PlaybackState{MediaId: 977, EpisodeNumber: 1}
	status := mediaplayer.PlaybackStatus{Filename: "episode.mkv", CompletionPercentage: 0.95, PlaybackType: mediaplayer.PlaybackTypeFile}
	streamState := playbackmanager.PlaybackState{MediaId: 978, EpisodeNumber: 2}
	streamStatus := mediaplayer.PlaybackStatus{Filename: "stream", CompletionPercentage: 0.99, PlaybackType: mediaplayer.PlaybackTypeStream}
	events := []playbackmanager.PlaybackEvent{
		playbackmanager.PlaybackStatusChangedEvent{State: state, Status: status},
		playbackmanager.VideoStartedEvent{Filename: status.Filename},
		playbackmanager.VideoCompletedEvent{Filename: status.Filename},
		playbackmanager.VideoStoppedEvent{Reason: "closed"},
		playbackmanager.PlaybackStatusChangedEvent{State: streamState, Status: streamStatus},
		playbackmanager.StreamStartedEvent{Filename: streamStatus.Filename},
		playbackmanager.StreamCompletedEvent{Filename: streamStatus.Filename},
		playbackmanager.StreamStoppedEvent{Reason: "closed"},
		playbackmanager.PlaybackStatusChangedEvent{State: state, Status: status},
	}
	for _, event := range events {
		subscriber.EventCh <- event
	}
	for i := range events {
		select {
		case data := <-received:
			var event PlaybackEvent
			require.NoError(t, json.Unmarshal([]byte(data), &event))
			require.NotNil(t, event.State)
			require.NotNil(t, event.Status)
			wantState, wantStatus := state, status
			if i >= 4 && i <= 7 {
				wantState, wantStatus = streamState, streamStatus
			}
			require.Equal(t, wantState, *event.State)
			require.Equal(t, wantStatus, *event.Status)
			switch i {
			case 1, 5:
				require.Equal(t, i == 1, event.IsVideoStarted)
				require.Equal(t, i == 5, event.IsStreamStarted)
				require.NotNil(t, event.StartedEvent)
				require.Equal(t, wantStatus.Filename, event.StartedEvent.Filename)
			case 2, 6:
				require.Equal(t, i == 2, event.IsVideoCompleted)
				require.Equal(t, i == 6, event.IsStreamCompleted)
				require.NotNil(t, event.CompletedEvent)
				require.Equal(t, wantStatus.Filename, event.CompletedEvent.Filename)
			case 3, 7:
				require.Equal(t, i == 3, event.IsVideoStopped)
				require.Equal(t, i == 7, event.IsStreamStopped)
				require.NotNil(t, event.StoppedEvent)
				require.Equal(t, "closed", event.StoppedEvent.Reason)
			default:
				require.False(t, event.IsVideoStarted || event.IsVideoCompleted || event.IsVideoStopped ||
					event.IsStreamStarted || event.IsStreamCompleted || event.IsStreamStopped)
				require.Nil(t, event.StartedEvent)
				require.Nil(t, event.CompletedEvent)
				require.Nil(t, event.StoppedEvent)
			}
		case <-time.After(time.Second):
			t.Fatalf("missing callback for %s", events[i].Type())
		}
	}
}
