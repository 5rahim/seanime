package core

import (
	"seanime/internal/hook"
	"seanime/internal/library/anime"
)

type AnimeLibraryCollectionRequestEvent struct {
	hook.Event

	LibraryCollection *anime.LibraryCollection
}

/**
	event := &hook.AnimeLibraryCollectionRequestEvent{
		LibraryCollection: libraryCollection,
	}

	return h.App.HookManager.OnRequestAnimeLibraryCollection().Trigger(event, func(e *hook.AnimeLibraryCollectionRequestEvent) error {
		return h.RespondWithData(c, e.LibraryCollection)
	})
**/

// func (a *App) OnRequestAnimeLibraryCollection() *hook.Hook[*AnimeLibraryCollectionRequestEvent] {
// 	return a.HookManager.onRequestAnimeLibraryCollection
// }
