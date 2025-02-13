package hook

import (
	"github.com/rs/zerolog"
	"seanime/internal/hook_event"
)

// HookManager manages all hooks in the application
type HookManager interface {

	// AniList Platform
	OnGetBaseAnime() *Hook[*hook_event.GetBaseAnimeEvent]
	OnGetBaseAnimeError() *Hook[*hook_event.GetBaseAnimeErrorEvent]
}

type HookManagerImpl struct {
	logger *zerolog.Logger

	onGetBaseAnime      *Hook[*hook_event.GetBaseAnimeEvent]
	onGetBaseAnimeError *Hook[*hook_event.GetBaseAnimeErrorEvent]
}

type NewHookManagerOptions struct {
	Logger *zerolog.Logger
}

func NewHookManager(opts NewHookManagerOptions) HookManager {
	ret := &HookManagerImpl{
		logger: opts.Logger,
	}

	ret.initHooks()

	return ret
}

func (m *HookManagerImpl) initHooks() {
	m.initAniListPlatformHooks()
}
