package hook

import "seanime/internal/hook_event"

func (m *HookManagerImpl) initAniListPlatformHooks() {
	m.onGetBaseAnime = &Hook[*hook_event.GetBaseAnimeEvent]{}
	m.onGetBaseAnimeError = &Hook[*hook_event.GetBaseAnimeErrorEvent]{}
}

func (m *HookManagerImpl) OnGetBaseAnime() *Hook[*hook_event.GetBaseAnimeEvent] {
	return m.onGetBaseAnime
}

func (m *HookManagerImpl) OnGetBaseAnimeError() *Hook[*hook_event.GetBaseAnimeErrorEvent] {
	return m.onGetBaseAnimeError
}
