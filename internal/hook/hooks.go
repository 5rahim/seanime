package hook

import (
	"seanime/internal/hook_event"

	"github.com/rs/zerolog"
)

// Manager manages all hooks in the application
type Manager interface {
	OnGetAnime() *Hook[*hook_event.GetAnimeEvent]
	OnGetAnimeDetails() *Hook[*hook_event.GetAnimeDetailsEvent]
	OnGetManga() *Hook[*hook_event.GetMangaEvent]
	OnGetMangaDetails() *Hook[*hook_event.GetMangaDetailsEvent]
	OnGetAnimeCollection() *Hook[*hook_event.GetAnimeCollectionEvent]
	OnGetMangaCollection() *Hook[*hook_event.GetMangaCollectionEvent]
	OnGetRawAnimeCollection() *Hook[*hook_event.GetRawAnimeCollectionEvent]
	OnGetRawMangaCollection() *Hook[*hook_event.GetRawMangaCollectionEvent]
	OnGetStudioDetails() *Hook[*hook_event.GetStudioDetailsEvent]
	OnPreUpdateEntry() *Hook[*hook_event.PreUpdateEntryEvent]
	OnPostUpdateEntry() *Hook[*hook_event.PostUpdateEntryEvent]
	OnPreUpdateEntryProgress() *Hook[*hook_event.PreUpdateEntryProgressEvent]
	OnPostUpdateEntryProgress() *Hook[*hook_event.PostUpdateEntryProgressEvent]
	OnPreUpdateEntryRepeat() *Hook[*hook_event.PreUpdateEntryRepeatEvent]
	OnPostUpdateEntryRepeat() *Hook[*hook_event.PostUpdateEntryRepeatEvent]
}

type ManagerImpl struct {
	logger *zerolog.Logger

	onGetAnime                *Hook[*hook_event.GetAnimeEvent]
	onGetAnimeDetails         *Hook[*hook_event.GetAnimeDetailsEvent]
	onGetManga                *Hook[*hook_event.GetMangaEvent]
	onGetMangaDetails         *Hook[*hook_event.GetMangaDetailsEvent]
	onGetAnimeCollection      *Hook[*hook_event.GetAnimeCollectionEvent]
	onGetMangaCollection      *Hook[*hook_event.GetMangaCollectionEvent]
	onGetRawAnimeCollection   *Hook[*hook_event.GetRawAnimeCollectionEvent]
	onGetRawMangaCollection   *Hook[*hook_event.GetRawMangaCollectionEvent]
	onGetStudioDetails        *Hook[*hook_event.GetStudioDetailsEvent]
	onPreUpdateEntry          *Hook[*hook_event.PreUpdateEntryEvent]
	onPostUpdateEntry         *Hook[*hook_event.PostUpdateEntryEvent]
	onPreUpdateEntryProgress  *Hook[*hook_event.PreUpdateEntryProgressEvent]
	onPostUpdateEntryProgress *Hook[*hook_event.PostUpdateEntryProgressEvent]
	onPreUpdateEntryRepeat    *Hook[*hook_event.PreUpdateEntryRepeatEvent]
	onPostUpdateEntryRepeat   *Hook[*hook_event.PostUpdateEntryRepeatEvent]
}

type NewHookManagerOptions struct {
	Logger *zerolog.Logger
}

func NewHookManager(opts NewHookManagerOptions) Manager {
	ret := &ManagerImpl{
		logger: opts.Logger,
	}

	ret.initHooks()

	return ret
}

func (m *ManagerImpl) initHooks() {
	m.onGetAnime = &Hook[*hook_event.GetAnimeEvent]{}
	m.onGetAnimeDetails = &Hook[*hook_event.GetAnimeDetailsEvent]{}
	m.onGetManga = &Hook[*hook_event.GetMangaEvent]{}
	m.onGetMangaDetails = &Hook[*hook_event.GetMangaDetailsEvent]{}
	m.onGetAnimeCollection = &Hook[*hook_event.GetAnimeCollectionEvent]{}
	m.onGetMangaCollection = &Hook[*hook_event.GetMangaCollectionEvent]{}
	m.onGetRawAnimeCollection = &Hook[*hook_event.GetRawAnimeCollectionEvent]{}
	m.onGetRawMangaCollection = &Hook[*hook_event.GetRawMangaCollectionEvent]{}
	m.onGetStudioDetails = &Hook[*hook_event.GetStudioDetailsEvent]{}
	m.onPreUpdateEntry = &Hook[*hook_event.PreUpdateEntryEvent]{}
	m.onPostUpdateEntry = &Hook[*hook_event.PostUpdateEntryEvent]{}
	m.onPreUpdateEntryProgress = &Hook[*hook_event.PreUpdateEntryProgressEvent]{}
	m.onPostUpdateEntryProgress = &Hook[*hook_event.PostUpdateEntryProgressEvent]{}
	m.onPreUpdateEntryRepeat = &Hook[*hook_event.PreUpdateEntryRepeatEvent]{}
	m.onPostUpdateEntryRepeat = &Hook[*hook_event.PostUpdateEntryRepeatEvent]{}
}

func (m *ManagerImpl) OnGetAnime() *Hook[*hook_event.GetAnimeEvent] {
	return m.onGetAnime
}

func (m *ManagerImpl) OnGetAnimeDetails() *Hook[*hook_event.GetAnimeDetailsEvent] {
	return m.onGetAnimeDetails
}

func (m *ManagerImpl) OnGetManga() *Hook[*hook_event.GetMangaEvent] {
	return m.onGetManga
}

func (m *ManagerImpl) OnGetMangaDetails() *Hook[*hook_event.GetMangaDetailsEvent] {
	return m.onGetMangaDetails
}

func (m *ManagerImpl) OnGetAnimeCollection() *Hook[*hook_event.GetAnimeCollectionEvent] {
	return m.onGetAnimeCollection
}

func (m *ManagerImpl) OnGetMangaCollection() *Hook[*hook_event.GetMangaCollectionEvent] {
	return m.onGetMangaCollection
}

func (m *ManagerImpl) OnGetRawAnimeCollection() *Hook[*hook_event.GetRawAnimeCollectionEvent] {
	return m.onGetRawAnimeCollection
}

func (m *ManagerImpl) OnGetRawMangaCollection() *Hook[*hook_event.GetRawMangaCollectionEvent] {
	return m.onGetRawMangaCollection
}

func (m *ManagerImpl) OnGetStudioDetails() *Hook[*hook_event.GetStudioDetailsEvent] {
	return m.onGetStudioDetails
}

func (m *ManagerImpl) OnPreUpdateEntry() *Hook[*hook_event.PreUpdateEntryEvent] {
	return m.onPreUpdateEntry
}

func (m *ManagerImpl) OnPostUpdateEntry() *Hook[*hook_event.PostUpdateEntryEvent] {
	return m.onPostUpdateEntry
}

func (m *ManagerImpl) OnPreUpdateEntryProgress() *Hook[*hook_event.PreUpdateEntryProgressEvent] {
	return m.onPreUpdateEntryProgress
}

func (m *ManagerImpl) OnPostUpdateEntryProgress() *Hook[*hook_event.PostUpdateEntryProgressEvent] {
	return m.onPostUpdateEntryProgress
}

func (m *ManagerImpl) OnPreUpdateEntryRepeat() *Hook[*hook_event.PreUpdateEntryRepeatEvent] {
	return m.onPreUpdateEntryRepeat
}

func (m *ManagerImpl) OnPostUpdateEntryRepeat() *Hook[*hook_event.PostUpdateEntryRepeatEvent] {
	return m.onPostUpdateEntryRepeat
}
