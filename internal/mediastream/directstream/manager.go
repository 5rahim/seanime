package directstream

import (
	"context"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/continuity"
	discordrpc_presence "seanime/internal/discordrpc/presence"
	"seanime/internal/events"
	"seanime/internal/library/anime"
	"seanime/internal/mediastream/nativeplayer"
	"seanime/internal/platforms/platform"
	"seanime/internal/util/result"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/mo"
)

// Manager handles direct stream playback and progress tracking for the built-in video player.
// It is similar to [playbackmanager.PlaybackManager].
type (
	Manager struct {
		Logger *zerolog.Logger

		// ------------ Modules ------------- //

		wsEventManager             events.WSEventManagerInterface
		continuityManager          *continuity.Manager
		metadataProvider           metadata.Provider
		discordPresence            *discordrpc_presence.Presence
		platform                   platform.Platform
		refreshAnimeCollectionFunc func() // This function is called to refresh the AniList collection

		nativePlayer           *nativeplayer.NativePlayer
		nativePlayerSubscriber *nativeplayer.Subscriber

		// --------- Playback Context -------- //

		playbackMu         sync.Mutex
		playbackCancelFunc context.CancelFunc

		// ---------- Playback State ---------- //

		currentStream mo.Option[Stream] // The current stream being played

		// \/ Stream playback
		// This is set by [SetStreamEpisodeCollection]
		currentStreamEpisodeCollection mo.Option[*anime.EpisodeCollection]

		settings *Settings

		isOffline       bool
		animeCollection mo.Option[*anilist.AnimeCollection]
		animeCache      *result.Cache[int, *anilist.BaseAnime]

		//playbackStatusSubscribers *result.Map[string, *PlaybackStatusSubscriber]
	}

	Settings struct {
		AutoPlayNextEpisode bool
	}

	NewManagerOptions struct {
		Logger                     *zerolog.Logger
		WSEventManager             events.WSEventManagerInterface
		MetadataProvider           metadata.Provider
		ContinuityManager          *continuity.Manager
		DiscordPresence            *discordrpc_presence.Presence
		Platform                   platform.Platform
		RefreshAnimeCollectionFunc func()
		IsOffline                  bool
		NativePlayer               *nativeplayer.NativePlayer
	}
)

func NewManager(options NewManagerOptions) *Manager {
	ret := &Manager{
		Logger:                     options.Logger,
		wsEventManager:             options.WSEventManager,
		metadataProvider:           options.MetadataProvider,
		continuityManager:          options.ContinuityManager,
		discordPresence:            options.DiscordPresence,
		platform:                   options.Platform,
		refreshAnimeCollectionFunc: options.RefreshAnimeCollectionFunc,
		isOffline:                  options.IsOffline,
		currentStream:              mo.None[Stream](),
		nativePlayer:               options.NativePlayer,
	}

	ret.nativePlayerSubscriber = ret.nativePlayer.Subscribe("directstream")

	return ret
}

func (m *Manager) SetAnimeCollection(ac *anilist.AnimeCollection) {
	m.animeCollection = mo.Some(ac)
}

func (m *Manager) SetSettings(s *Settings) {
	m.settings = s
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (m *Manager) getAnime(mediaId int) (*anilist.BaseAnime, error) {
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
	media, err := m.platform.GetAnime(mediaId)
	if err != nil {
		return nil, err
	}

	// Cache
	m.animeCache.SetT(mediaId, media, 1*time.Hour)

	return media, nil
}
