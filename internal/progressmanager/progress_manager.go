package progressmanager

import (
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/events"
)

type (
	// ProgressManager is used as an interface between the video playback and progress tracking.
	// It can receive progress updates and dispatch appropriate events for:
	//  - syncing progress with AniList, MAL, etc.
	//  - sending notifications to the client
	//  - DEVNOTE: in the future, it could also be used to implement w2g, handle built-in player or allow multiple watchers
	ProgressManager struct {
		wsEventManager       events.IWSEventManager
		Logger               *zerolog.Logger
		anilistClientWrapper *anilist.ClientWrapper
		anilistCollection    *anilist.AnimeCollection
	}
	NewProgressManagerOptions struct {
		WSEventManager       events.IWSEventManager
		Logger               *zerolog.Logger
		AnilistClientWrapper *anilist.ClientWrapper
		AnilistCollection    *anilist.AnimeCollection
	}
)

func New(opts *NewProgressManagerOptions) *ProgressManager {
	return &ProgressManager{
		Logger:               opts.Logger,
		wsEventManager:       opts.WSEventManager,
		anilistClientWrapper: opts.AnilistClientWrapper,
		anilistCollection:    opts.AnilistCollection,
	}
}

func (p *ProgressManager) SetAnilistClientWrapper(anilistClientWrapper *anilist.ClientWrapper) {
	p.anilistClientWrapper = anilistClientWrapper
}

func (p *ProgressManager) SetAnilistCollection(anilistCollection *anilist.AnimeCollection) {
	p.anilistCollection = anilistCollection
}
