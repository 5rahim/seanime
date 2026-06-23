package directstream

import (
	"context"
	"testing"

	"seanime/internal/mkvparser"
	"seanime/internal/player"
	"seanime/internal/util"
	"seanime/internal/util/result"

	"github.com/samber/mo"
	"github.com/stretchr/testify/require"
)

func TestSubtitleOffsetForTimeUsesPlaybackProgress(t *testing.T) {
	// keeps seek-based subtitle refresh near the current playback position
	playbackInfo := &player.PlaybackInfo{ContentLength: defaultSubtitleBackoffBytes * 4}

	require.Equal(t, defaultSubtitleBackoffBytes, subtitleOffsetForTime(playbackInfo, 25, 100))
}

func TestSubtitleOffsetForTimeFallsBackToMetadataDuration(t *testing.T) {
	// falls back to mkv metadata when the player duration is not available yet
	playbackInfo := &player.PlaybackInfo{
		ContentLength: defaultSubtitleBackoffBytes * 4,
		MkvMetadata: &mkvparser.Metadata{
			Duration: 200,
		},
	}

	require.Equal(t, defaultSubtitleBackoffBytes, subtitleOffsetForTime(playbackInfo, 50, 0))
}

func TestSubtitleOffsetForTimeClampsNearEnd(t *testing.T) {
	// leaves enough room for the subtitle parser backoff near eof
	playbackInfo := &player.PlaybackInfo{ContentLength: defaultSubtitleBackoffBytes * 2}

	require.Equal(t, defaultSubtitleBackoffBytes, subtitleOffsetForTime(playbackInfo, 199, 200))
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

	stream.StartSubtitleStreamP(stream, context.Background(), reader, 8*1024*1024+256*1024, defaultSubtitleBackoffBytes)

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
