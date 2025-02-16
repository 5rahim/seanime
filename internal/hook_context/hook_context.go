package hook_context

type AppContext interface {
	Store() *Store[string, any]
}

type AppContextImpl struct {
	store *Store[string, any]
}

func NewAppContext() AppContext {
	appCtx := &AppContextImpl{
		store: New[string, any](nil),
	}

	return appCtx
}

func (a *AppContextImpl) Store() *Store[string, any] {
	return a.store
}
