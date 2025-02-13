package hook

import "github.com/rs/zerolog"

// HookManager manages all hooks in the application
type HookManager interface {

	// AniList Platform
	OnGetBaseAnime() *Hook[Resolver]
	OnGetBaseAnimeError() *Hook[Resolver]
}

type HookManagerImpl struct {
	logger *zerolog.Logger

	onGetBaseAnime      *Hook[Resolver]
	onGetBaseAnimeError *Hook[Resolver]
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
