package hook

import (
	"seanime/internal/library/anime"
)

type AnimeLibraryCollectionRequestEvent struct {
	Event

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

func (m *HookManager) OnRequestAnimeLibraryCollection() *Hook[*AnimeLibraryCollectionRequestEvent] {
	return m.onRequestAnimeLibraryCollection
}
