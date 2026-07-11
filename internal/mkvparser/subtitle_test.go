package mkvparser

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSendSubtitleEventSkipsEventsBeforeSeekTarget(t *testing.T) {
	subtitleCh := make(chan *SubtitleEvent, 1)
	event := &SubtitleEvent{StartTime: 1000, Duration: 2000}

	require.True(t, sendSubtitleEvent(context.Background(), subtitleCh, event, 4000))
	require.Empty(t, subtitleCh)
}

func TestSendSubtitleEventKeepsEventsSpanningSeekTarget(t *testing.T) {
	subtitleCh := make(chan *SubtitleEvent, 1)
	event := &SubtitleEvent{StartTime: 1000, Duration: 4000}

	require.True(t, sendSubtitleEvent(context.Background(), subtitleCh, event, 4000))
	require.Same(t, event, <-subtitleCh)
}
