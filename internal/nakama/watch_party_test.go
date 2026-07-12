package nakama

import (
	"context"
	"seanime/internal/api/anilist"
	"seanime/internal/customsource"
	"seanime/internal/extension"
	"seanime/internal/platforms/platform"
	"seanime/internal/testmocks"
	"seanime/internal/util"
	"testing"

	"github.com/rs/zerolog"
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

func TestWatchPartyCustomSourceTranslation(t *testing.T) {
	// Create mock custom source extensions
	// On Host: id "demo-ext", extensionIdentifier 456
	// On Peer (Relay Origin): id "demo-ext", extensionIdentifier 123
	ext := &extension.Extension{
		ID:   "demo-ext",
		Type: extension.TypeCustomSource,
	}
	hostExt := extension.NewCustomSourceExtension(ext, nil)
	hostExt.SetExtensionIdentifier(456)

	hostBank := extension.NewUnifiedBank()
	hostBank.Set("demo-ext", hostExt)
	hostCSM := customsource.NewManager(util.NewRef(hostBank), nil, new(zerolog.Nop()))

	// Create host WatchPartyManager
	hostPlatform := &mockPlatformWithCustomSource{
		Platform: testmocks.NewFakePlatformBuilder().Build(),
		csm:      hostCSM,
	}
	hostWPM := newWatchPartyMediaTestManager(hostPlatform)

	// Peer media ID is offset + (123 << 40) + localId
	peerMediaId := customsource.GenerateMediaId(123, 789)
	hostMediaId := customsource.GenerateMediaId(456, 789)

	// Peer media URL contains "ext_custom_source_demo-ext"
	siteURL := "ext_custom_source_demo-ext|END|https://example.com/anime/789"
	media := &anilist.BaseAnime{
		ID:      peerMediaId,
		SiteURL: &siteURL,
	}

	// 1. Test translation helper getLocalMediaIdOfCustomSource
	gotId := hostWPM.getLocalMediaIdOfCustomSource(peerMediaId, media)
	require.Equal(t, hostMediaId, gotId)

	// 2. Test translateMediaInfo
	info := &WatchPartySessionMediaInfo{
		MediaId: peerMediaId,
		Media:   media,
	}
	hostWPM.translateMediaInfo(info)
	require.Equal(t, hostMediaId, info.MediaId)
	require.Equal(t, hostMediaId, info.Media.ID)

	// 3. Test translateOriginStreamStartedPayload
	payload := &WatchPartyRelayModeOriginStreamStartedPayload{
		Media: media,
		State: &WatchPartyPlaybackState{
			MediaId: peerMediaId,
		},
	}
	hostWPM.translateOriginStreamStartedPayload(payload)
	require.Equal(t, hostMediaId, payload.Media.ID)
	require.Equal(t, hostMediaId, payload.State.MediaId)
}

func newWatchPartyMediaTestManager(p platform.Platform) *WatchPartyManager {
	return NewWatchPartyManager(&Manager{
		logger:      new(zerolog.Nop()),
		platformRef: util.NewRef[platform.Platform](p),
	})
}

type mockPlatformWithCustomSource struct {
	platform.Platform
	csm *customsource.Manager
}

func (m *mockPlatformWithCustomSource) GetCustomSourceManager() *customsource.Manager {
	return m.csm
}
