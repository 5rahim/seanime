package directstream

import (
	"context"
	"encoding/json"
	"seanime/internal/events"
	"seanime/internal/mkvparser"
	"seanime/internal/nativeplayer"
	"seanime/internal/player"
	"seanime/internal/util"
	"seanime/internal/util/result"
	"testing"

	"github.com/samber/mo"
	"github.com/stretchr/testify/require"
)

type subtitleTestVideoCore struct{}

func (subtitleTestVideoCore) RecordEvent(*mkvparser.SubtitleEvent) {}
func (subtitleTestVideoCore) Reset()                               {}
func (subtitleTestVideoCore) Terminate()                           {}

func TestSubtitleOffsetForTimeUsesPlaybackProgress(t *testing.T) {
	// keeps seek-based subtitle refresh near the current playback position
	playbackInfo := &player.PlaybackInfo{ContentLength: subtitleBackoffBytes * 4}

	require.Equal(t, subtitleBackoffBytes, subtitleOffsetForTime(playbackInfo, 25, 100))
}

func TestSubtitleOffsetForTimeFallsBackToMetadataDuration(t *testing.T) {
	// falls back to mkv metadata when the player duration is not available yet
	playbackInfo := &player.PlaybackInfo{
		ContentLength: subtitleBackoffBytes * 4,
		MkvMetadata: &mkvparser.Metadata{
			Duration: 200,
		},
	}

	require.Equal(t, subtitleBackoffBytes, subtitleOffsetForTime(playbackInfo, 50, 0))
}

func TestSubtitleOffsetForTimeClampsNearEnd(t *testing.T) {
	// leaves enough room for the subtitle parser backoff near eof
	playbackInfo := &player.PlaybackInfo{ContentLength: subtitleBackoffBytes * 2}

	require.Equal(t, subtitleBackoffBytes, subtitleOffsetForTime(playbackInfo, 199, 200))
}

func TestStartSubtitleStreamPSkipsNearbyActiveStream(t *testing.T) {
	// avoids starting a duplicate stream when seek and range land in the same area
	reader := &trackingReadSeekCloser{}
	stream := &BaseStream{
		logger: util.NewLogger(),
		playbackInfo: &player.PlaybackInfo{
			MkvMetadataParser: mo.Some(&mkvparser.MetadataParser{}),
		},
		activeSubtitleStreams: result.NewMap[string, *SubtitleStream](),
	}
	stream.activeSubtitleStreams.Set("existing", &SubtitleStream{offset: 8 * 1024 * 1024})

	stream.StartSubtitleStreamP(stream, context.Background(), reader, 8*1024*1024+256*1024, subtitleBackoffBytes)

	require.True(t, reader.closed)

	count := 0
	stream.activeSubtitleStreams.Range(func(_ string, _ *SubtitleStream) bool {
		count++
		return true
	})
	require.Equal(t, 1, count)
}

func TestSubtitleFlushConfigForTorrentThrottlesBatches(t *testing.T) {
	// torrent subtitle extraction can outrun the UI, so its batches stay smaller
	defaultConfig := subtitleFlushConfigFor(player.PlaybackTypeDebrid, 0)
	torrentConfig := subtitleFlushConfigFor(player.PlaybackTypeTorrent, 0)
	torrentSeekConfig := subtitleFlushConfigFor(player.PlaybackTypeTorrent, 8*1024*1024)

	require.Less(t, torrentConfig.maxBatchSize, defaultConfig.maxBatchSize)
	require.Greater(t, torrentConfig.flushInterval, defaultConfig.flushInterval)
	require.Zero(t, defaultConfig.minSendInterval)
	require.NotZero(t, torrentConfig.minSendInterval)
	require.Less(t, torrentSeekConfig.flushInterval, torrentConfig.flushInterval)
	require.Less(t, torrentSeekConfig.minSendInterval, torrentConfig.minSendInterval)
}

func TestShouldSendSubtitleEventSkipsCachedEvents(t *testing.T) {
	// seeks can rediscover old subtitles, but the browser only needs each event once
	stream := &BaseStream{subtitleEventCache: result.NewMap[string, *mkvparser.SubtitleEvent]()}
	event := &mkvparser.SubtitleEvent{
		TrackNumber: 1,
		Text:        "hello",
		StartTime:   10,
		Duration:    2,
		CodecID:     "S_TEXT/ASS",
		ExtraData:   map[string]string{"style": "Default"},
	}

	require.True(t, stream.shouldSendSubtitleEvent(event))
	require.False(t, stream.shouldSendSubtitleEvent(event))
	require.True(t, stream.shouldSendSubtitleEvent(&mkvparser.SubtitleEvent{
		TrackNumber: 1,
		Text:        "hello",
		StartTime:   12,
		Duration:    2,
		CodecID:     "S_TEXT/ASS",
		ExtraData:   map[string]string{"style": "Default"},
	}))
}

func TestBeginSubtitleSeekCancelsPreviousGeneration(t *testing.T) {
	stream := &BaseStream{
		logger:                util.NewLogger(),
		playbackInfo:          &player.PlaybackInfo{ID: "playback-1"},
		activeSubtitleStreams: result.NewMap[string, *SubtitleStream](),
	}
	stopped := false
	previous := &SubtitleStream{
		logger: util.NewLogger(),
		onStop: func() {
			stopped = true
		},
	}
	stream.activeSubtitleStreams.Set("previous", previous)

	request := stream.beginSubtitleSeek(125.5)

	require.True(t, stopped)
	require.Equal(t, "playback-1", request.playbackID)
	require.Equal(t, int64(1), request.generation)
	require.Equal(t, 125.5, request.seekTime)
	require.Equal(t, request.generation, stream.subtitleGeneration.Load())
}

func TestStartSubtitleStreamPRejectsStaleGeneration(t *testing.T) {
	reader := &trackingReadSeekCloser{}
	stream := &BaseStream{
		logger: util.NewLogger(),
		playbackInfo: &player.PlaybackInfo{
			MkvMetadataParser: mo.Some(&mkvparser.MetadataParser{}),
		},
		activeSubtitleStreams: result.NewMap[string, *SubtitleStream](),
	}
	stream.subtitleGeneration.Store(2)

	stream.startSubtitleStreamP(stream, context.Background(), reader, 0, subtitleBackoffBytes, subtitleRequest{generation: 1})

	require.True(t, reader.closed)
}

func TestSendSubtitleEventsRejectsStaleGeneration(t *testing.T) {
	stream := &BaseStream{}
	stream.subtitleGeneration.Store(2)

	sent := stream.sendSubtitleEvents(context.Background(), stream, []*mkvparser.SubtitleEvent{{Text: "stale"}}, subtitleFlushConfig{}, subtitleRequest{generation: 1})

	require.False(t, sent)
}

func TestSendSubtitleEventsIncludesPlaybackGeneration(t *testing.T) {
	logger := util.NewLogger()
	ws := events.NewMockWSEventManager(logger)
	nativePlayer := nativeplayer.New(nativeplayer.NewNativePlayerOptions{
		WsEventManager: ws,
		Logger:         logger,
		VideoCore:      subtitleTestVideoCore{},
	})
	manager := &Manager{
		nativePlayer:          nativePlayer,
		currentPlaybackTarget: PlaybackTargetVideoCore,
	}
	stream := &LocalFileStream{BaseStream: BaseStream{
		manager:  manager,
		clientId: "client",
	}}
	stream.subtitleGeneration.Store(3)
	event := &mkvparser.SubtitleEvent{Text: "current"}

	sent := stream.sendSubtitleEvents(context.Background(), stream, []*mkvparser.SubtitleEvent{event}, subtitleFlushConfig{}, subtitleRequest{
		playbackID: "playback-1",
		generation: 3,
		seekTime:   125.5,
	})

	require.True(t, sent)
	require.Len(t, ws.Events(), 1)
	payload, err := json.Marshal(ws.Events()[0].Payload)
	require.NoError(t, err)
	var message struct {
		Type    string                             `json:"type"`
		Payload nativeplayer.SubtitleEventsPayload `json:"payload"`
	}
	require.NoError(t, json.Unmarshal(payload, &message))
	require.Equal(t, string(nativeplayer.ServerEventSubtitleEvent), message.Type)
	require.Equal(t, "playback-1", message.Payload.PlaybackID)
	require.Equal(t, int64(3), message.Payload.GenerationID)
	require.Equal(t, 125.5, message.Payload.SeekTime)
	require.Equal(t, "current", message.Payload.Events[0].Text)
}
