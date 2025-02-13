package hook

func (m *HookManagerImpl) initAniListPlatformHooks() {
	m.onGetBaseAnime = &Hook[Resolver]{}
	m.onGetBaseAnimeError = &Hook[Resolver]{}
}

func (m *HookManagerImpl) OnGetBaseAnime() *Hook[Resolver] {
	return m.onGetBaseAnime
}

func (m *HookManagerImpl) OnGetBaseAnimeError() *Hook[Resolver] {
	return m.onGetBaseAnimeError
}
