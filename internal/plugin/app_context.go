package plugin

type AppContext interface {
	Store() *Store[string, any]
}

type AppContextImpl struct {
	store *Store[string, any]
}

func NewAppContext() AppContext {
	appCtx := &AppContextImpl{
		store: NewStore[string, any](nil),
	}

	return appCtx
}

func (a *AppContextImpl) Store() *Store[string, any] {
	return a.store
}
