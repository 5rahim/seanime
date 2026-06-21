package directstream

import (
	"context"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata_provider"
	"seanime/internal/continuity"
	discordrpc_presence "seanime/internal/discordrpc/presence"
	"seanime/internal/events"
	"seanime/internal/library/anime"
	"seanime/internal/mediacore"
	"seanime/internal/mkvparser"
	"seanime/internal/nativeplayer"
	"seanime/internal/platforms/platform"
	"seanime/internal/util"
	"seanime/internal/util/result"
	"seanime/internal/videocore"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/mo"
)

// Manager handles direct stream playback and progress tracking for the built-in video player.
// It is similar to playbackmanager.PlaybackManager.
type (
	Manager struct {
		Logger *zerolog.Logger

		// ------------ Modules ------------- //

		wsEventManager             events.WSEventManagerInterface
		continuityManager          *continuity.Manager
		metadataProviderRef        *util.Ref[metadata_provider.Provider]
		discordPresence            *discordrpc_presence.Presence
		platformRef                *util.Ref[platform.Platform]
		refreshAnimeCollectionFunc func()                                      // This function is called to refresh the AniList collection
		hmacTokenFunc              func(endpoint string, symbol string) string // Generates HMAC token query param for stream URLs

		nativePlayer         *nativeplayer.NativePlayer
		videoCore            *videocore.VideoCore
		mediacoreCoordinator *mediacore.Coordinator
		mediacoreSubscriber  *mediacore.Subscriber

		// --------- Playback Context -------- //

		playbackMu            sync.Mutex
		playbackCtx           context.Context
		playbackCtxCancelFunc context.CancelFunc

		// ---------- Playback State ---------- //

		currentStream          mo.Option[Stream] // The current stream being played
		currentPlaybackId      string
		currentPlaybackClient  string
		replacedPlaybackId     string
		replacedPlaybackClient string
		preparingClientID      string
		preparingTarget        PlaybackTarget
		preparationCanceled    bool
		preparationCancelFunc  func()
		currentPlaybackTarget  PlaybackTarget
		defaultPlaybackTarget  PlaybackTarget

		// \/ Stream playback
		// This is set by [SetStreamEpisodeCollection]
		currentStreamEpisodeCollection mo.Option[*anime.EpisodeCollection]

		settings *Settings

		isOfflineRef    *util.Ref[bool]
		animeCollection mo.Option[*anilist.AnimeCollection]
		animeCache      *result.Cache[int, *anilist.BaseAnime]

		parserCache *result.Cache[string, *mkvparser.MetadataParser]
		//playbackStatusSubscribers *result.Map[string, *PlaybackStatusSubscriber]
	}

	Settings struct {
		AutoPlayNextEpisode bool
		AutoUpdateProgress  bool
	}

	NewManagerOptions struct {
		Logger                     *zerolog.Logger
		WSEventManager             events.WSEventManagerInterface
		MetadataProviderRef        *util.Ref[metadata_provider.Provider]
		ContinuityManager          *continuity.Manager
		DiscordPresence            *discordrpc_presence.Presence
		PlatformRef                *util.Ref[platform.Platform]
		RefreshAnimeCollectionFunc func()
		IsOfflineRef               *util.Ref[bool]
		NativePlayer               *nativeplayer.NativePlayer
		VideoCore                  *videocore.VideoCore
		MediacoreCoordinator       *mediacore.Coordinator
		HMACTokenFunc              func(endpoint string, symbol string) string
	}
)

func NewManager(options NewManagerOptions) *Manager {
	ret := &Manager{
		Logger:                     options.Logger,
		wsEventManager:             options.WSEventManager,
		metadataProviderRef:        options.MetadataProviderRef,
		continuityManager:          options.ContinuityManager,
		discordPresence:            options.DiscordPresence,
		platformRef:                options.PlatformRef,
		refreshAnimeCollectionFunc: options.RefreshAnimeCollectionFunc,
		hmacTokenFunc:              options.HMACTokenFunc,
		isOfflineRef:               options.IsOfflineRef,
		currentStream:              mo.None[Stream](),
		nativePlayer:               options.NativePlayer,
		videoCore:                  options.VideoCore,
		mediacoreCoordinator:       options.MediacoreCoordinator,
		defaultPlaybackTarget:      PlaybackTargetVideoCore,
		parserCache:                result.NewCache[string, *mkvparser.MetadataParser](),
	}
	if ret.mediacoreCoordinator != nil {
		ret.mediacoreSubscriber = ret.mediacoreCoordinator.Subscribe("directstream")
	}
	ret.listenToPlayerEvents()

	return ret
}

type PlaybackTarget string

const (
	PlaybackTargetVideoCore PlaybackTarget = "videocore"
	PlaybackTargetMpvCore   PlaybackTarget = "mpvcore"
)

func (m *Manager) SetPlaybackTarget(target PlaybackTarget) {
	if target != PlaybackTargetVideoCore && target != PlaybackTargetMpvCore {
		return
	}
	m.playbackMu.Lock()
	m.defaultPlaybackTarget = target
	m.playbackMu.Unlock()
}

func (m *Manager) GetPlaybackTarget() PlaybackTarget {
	m.playbackMu.Lock()
	defer m.playbackMu.Unlock()
	return m.defaultPlaybackTarget
}

func (m *Manager) SetAnimeCollection(ac *anilist.AnimeCollection) {
	m.animeCollection = mo.Some(ac)
}

func (m *Manager) SetSettings(s *Settings) {
	m.settings = s
}

// GetHMACTokenQueryParam returns an HMAC token query param for the given endpoint, or empty string if not available.
func (m *Manager) GetHMACTokenQueryParam(endpoint string, symbol string) string {
	if m.hmacTokenFunc != nil {
		return m.hmacTokenFunc(endpoint, symbol)
	}
	return ""
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (m *Manager) getAnime(ctx context.Context, mediaId int) (*anilist.BaseAnime, error) {
	media, ok := m.animeCache.Get(mediaId)
	if ok {
		return media, nil
	}

	// Find in anime collection
	animeCollection, ok := m.animeCollection.Get()
	if ok {
		media, ok := animeCollection.FindAnime(mediaId)
		if ok {
			return media, nil
		}
	}

	// Find in platform
	media, err := m.platformRef.Get().GetAnime(ctx, mediaId)
	if err != nil {
		return nil, err
	}

	// Cache
	m.animeCache.SetT(mediaId, media, 1*time.Hour)

	return media, nil
}
