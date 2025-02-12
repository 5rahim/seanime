package hook

func (m *HookManager) initAniListPlatformHooks() {
	m.onGetBaseAnime = &Hook[Resolver]{}
	m.onGetBaseAnimeError = &Hook[Resolver]{}
}

func (m *HookManager) OnGetBaseAnime() *Hook[Resolver] {
	return m.onGetBaseAnime
}

func (m *HookManager) OnGetBaseAnimeError() *Hook[Resolver] {
	return m.onGetBaseAnimeError
}

func (m *HookManager) BindAniListPlatformHooks() {
	m.onGetBaseAnime.BindFunc(func(e Resolver) error {
		// plugin code execution goes here
		return e.Next()
	})

	m.onGetBaseAnimeError.BindFunc(func(e Resolver) error {
		// plugin code execution goes here
		return e.Next()
	})
}
