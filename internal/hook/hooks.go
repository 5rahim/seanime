package hook

import "github.com/rs/zerolog"

// HookManager manages all hooks in the application
type HookManager struct {
	logger *zerolog.Logger

	// Anime Library
	onRequestAnimeLibraryCollection *Hook[*AnimeLibraryCollectionRequestEvent]
}

type NewHookManagerOptions struct {
	Logger *zerolog.Logger
}

func NewHookManager(opts NewHookManagerOptions) *HookManager {
	return &HookManager{
		logger: opts.Logger,
	}
}

func (m *HookManager) initHooks() {
	m.onRequestAnimeLibraryCollection = &Hook[*AnimeLibraryCollectionRequestEvent]{}
}
