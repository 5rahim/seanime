package nakama

import (
	"context"
	"seanime/internal/platforms/platform"
	"seanime/internal/testmocks"
	"seanime/internal/util"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWatchPartyGetSessionMediaUsesPayloadMedia(t *testing.T) {
	media := testmocks.NewBaseAnime(990748023463256, "Custom Source Anime")
	fakePlatform := testmocks.NewFakePlatformBuilder().Build()
	wpm := newWatchPartyMediaTestManager(fakePlatform)

	got, err := wpm.getSessionMedia(context.Background(), &WatchPartySessionMediaInfo{
		MediaId: media.ID,
		Media:   media,
	})

	require.NoError(t, err)
	require.Same(t, media, got)
	require.Zero(t, fakePlatform.AnimeCalls(media.ID))
}

func TestWatchPartyGetSessionMediaFallsBackToPlatform(t *testing.T) {
	media := testmocks.NewBaseAnime(154587, "Frieren")
	fakePlatform := testmocks.NewFakePlatformBuilder().WithAnime(media).Build()
	wpm := newWatchPartyMediaTestManager(fakePlatform)

	got, err := wpm.getSessionMedia(context.Background(), &WatchPartySessionMediaInfo{
		MediaId: media.ID,
	})

	require.NoError(t, err)
	require.Same(t, media, got)
	require.Equal(t, 1, fakePlatform.AnimeCalls(media.ID))
}

func newWatchPartyMediaTestManager(p platform.Platform) *WatchPartyManager {
	return NewWatchPartyManager(&Manager{
		platformRef: util.NewRef[platform.Platform](p),
	})
}
