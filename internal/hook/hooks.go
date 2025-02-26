package hook

import (
	"seanime/internal/hook_resolver"
	"seanime/internal/util"

	"github.com/rs/zerolog"
)

// Manager manages all hooks in the application
type Manager interface {
	// AniList events
	OnGetAnime() *Hook[hook_resolver.Resolver]
	OnGetAnimeDetails() *Hook[hook_resolver.Resolver]
	OnGetManga() *Hook[hook_resolver.Resolver]
	OnGetMangaDetails() *Hook[hook_resolver.Resolver]
	OnGetAnimeCollection() *Hook[hook_resolver.Resolver]
	OnGetMangaCollection() *Hook[hook_resolver.Resolver]
	OnGetRawAnimeCollection() *Hook[hook_resolver.Resolver]
	OnGetRawMangaCollection() *Hook[hook_resolver.Resolver]
	OnGetStudioDetails() *Hook[hook_resolver.Resolver]
	OnPreUpdateEntry() *Hook[hook_resolver.Resolver]
	OnPostUpdateEntry() *Hook[hook_resolver.Resolver]
	OnPreUpdateEntryProgress() *Hook[hook_resolver.Resolver]
	OnPostUpdateEntryProgress() *Hook[hook_resolver.Resolver]
	OnPreUpdateEntryRepeat() *Hook[hook_resolver.Resolver]
	OnPostUpdateEntryRepeat() *Hook[hook_resolver.Resolver]

	// Anime library events
	OnAnimeEntryRequested() *Hook[hook_resolver.Resolver]
	OnAnimeEntry() *Hook[hook_resolver.Resolver]

	OnAnimeEntryFillerHydration() *Hook[hook_resolver.Resolver]

	OnAnimeEntryLibraryDataRequested() *Hook[hook_resolver.Resolver]
	OnAnimeEntryLibraryData() *Hook[hook_resolver.Resolver]

	OnAnimeEntryManualMatchBeforeSave() *Hook[hook_resolver.Resolver]

	OnMissingEpisodesRequested() *Hook[hook_resolver.Resolver]
	OnMissingEpisodes() *Hook[hook_resolver.Resolver]

	// Anime library collection events
	OnAnimeLibraryCollectionRequested() *Hook[hook_resolver.Resolver]
	OnAnimeLibraryCollection() *Hook[hook_resolver.Resolver]

	OnAnimeLibraryStreamCollectionRequested() *Hook[hook_resolver.Resolver]
	OnAnimeLibraryStreamCollection() *Hook[hook_resolver.Resolver]

	// Auto Downloader events
	OnAutoDownloaderQueueOrDownloadTorrent() *Hook[hook_resolver.Resolver]
	OnAutoDownloaderTorrentMatched() *Hook[hook_resolver.Resolver]
	OnAutoDownloaderRuleVerifyMatch() *Hook[hook_resolver.Resolver]
	OnAutoDownloaderRunStarted() *Hook[hook_resolver.Resolver]
	OnAutoDownloaderRunCompleted() *Hook[hook_resolver.Resolver]
	OnAutoDownloaderSettingsUpdated() *Hook[hook_resolver.Resolver]
}

type ManagerImpl struct {
	logger *zerolog.Logger
	// AniList events
	onGetAnime                *Hook[hook_resolver.Resolver]
	onGetAnimeDetails         *Hook[hook_resolver.Resolver]
	onGetManga                *Hook[hook_resolver.Resolver]
	onGetMangaDetails         *Hook[hook_resolver.Resolver]
	onGetAnimeCollection      *Hook[hook_resolver.Resolver]
	onGetMangaCollection      *Hook[hook_resolver.Resolver]
	onGetRawAnimeCollection   *Hook[hook_resolver.Resolver]
	onGetRawMangaCollection   *Hook[hook_resolver.Resolver]
	onGetStudioDetails        *Hook[hook_resolver.Resolver]
	onPreUpdateEntry          *Hook[hook_resolver.Resolver]
	onPostUpdateEntry         *Hook[hook_resolver.Resolver]
	onPreUpdateEntryProgress  *Hook[hook_resolver.Resolver]
	onPostUpdateEntryProgress *Hook[hook_resolver.Resolver]
	onPreUpdateEntryRepeat    *Hook[hook_resolver.Resolver]
	onPostUpdateEntryRepeat   *Hook[hook_resolver.Resolver]
	// Anime library events
	onAnimeEntryRequested             *Hook[hook_resolver.Resolver]
	onAnimeEntry                      *Hook[hook_resolver.Resolver]
	onAnimeEntryFillerHydration       *Hook[hook_resolver.Resolver]
	onAnimeEntryLibraryDataRequested  *Hook[hook_resolver.Resolver]
	onAnimeEntryLibraryData           *Hook[hook_resolver.Resolver]
	onAnimeEntryManualMatchBeforeSave *Hook[hook_resolver.Resolver]
	onMissingEpisodesRequested        *Hook[hook_resolver.Resolver]
	onMissingEpisodes                 *Hook[hook_resolver.Resolver]
	// Anime library collection events
	onAnimeLibraryCollectionRequested       *Hook[hook_resolver.Resolver]
	onAnimeLibraryCollection                *Hook[hook_resolver.Resolver]
	onAnimeLibraryStreamCollectionRequested *Hook[hook_resolver.Resolver]
	onAnimeLibraryStreamCollection          *Hook[hook_resolver.Resolver]
	// Auto Downloader events
	onAutoDownloaderQueueOrDownloadTorrent *Hook[hook_resolver.Resolver]
	onAutoDownloaderTorrentMatched         *Hook[hook_resolver.Resolver]
	onAutoDownloaderRuleVerifyMatch        *Hook[hook_resolver.Resolver]
	onAutoDownloaderRunStarted             *Hook[hook_resolver.Resolver]
	onAutoDownloaderRunCompleted           *Hook[hook_resolver.Resolver]
	onAutoDownloaderSettingsUpdated        *Hook[hook_resolver.Resolver]
}

type NewHookManagerOptions struct {
	Logger *zerolog.Logger
}

var GlobalHookManager = NewHookManager(NewHookManagerOptions{
	Logger: util.NewLogger(),
})

func SetGlobalHookManager(manager Manager) {
	GlobalHookManager = manager
}

func NewHookManager(opts NewHookManagerOptions) Manager {
	ret := &ManagerImpl{
		logger: opts.Logger,
	}

	ret.initHooks()

	return ret
}

func (m *ManagerImpl) initHooks() {
	// AniList events
	m.onGetAnime = &Hook[hook_resolver.Resolver]{}
	m.onGetAnimeDetails = &Hook[hook_resolver.Resolver]{}
	m.onGetManga = &Hook[hook_resolver.Resolver]{}
	m.onGetMangaDetails = &Hook[hook_resolver.Resolver]{}
	m.onGetAnimeCollection = &Hook[hook_resolver.Resolver]{}
	m.onGetMangaCollection = &Hook[hook_resolver.Resolver]{}
	m.onGetRawAnimeCollection = &Hook[hook_resolver.Resolver]{}
	m.onGetRawMangaCollection = &Hook[hook_resolver.Resolver]{}
	m.onGetStudioDetails = &Hook[hook_resolver.Resolver]{}
	m.onPreUpdateEntry = &Hook[hook_resolver.Resolver]{}
	m.onPostUpdateEntry = &Hook[hook_resolver.Resolver]{}
	m.onPreUpdateEntryProgress = &Hook[hook_resolver.Resolver]{}
	m.onPostUpdateEntryProgress = &Hook[hook_resolver.Resolver]{}
	m.onPreUpdateEntryRepeat = &Hook[hook_resolver.Resolver]{}
	m.onPostUpdateEntryRepeat = &Hook[hook_resolver.Resolver]{}
	// Anime library events
	m.onAnimeEntryRequested = &Hook[hook_resolver.Resolver]{}
	m.onAnimeEntry = &Hook[hook_resolver.Resolver]{}
	m.onAnimeEntryFillerHydration = &Hook[hook_resolver.Resolver]{}
	m.onAnimeEntryLibraryDataRequested = &Hook[hook_resolver.Resolver]{}
	m.onAnimeEntryLibraryData = &Hook[hook_resolver.Resolver]{}
	m.onAnimeEntryManualMatchBeforeSave = &Hook[hook_resolver.Resolver]{}
	m.onMissingEpisodesRequested = &Hook[hook_resolver.Resolver]{}
	m.onMissingEpisodes = &Hook[hook_resolver.Resolver]{}
	// Anime library collection events
	m.onAnimeLibraryCollectionRequested = &Hook[hook_resolver.Resolver]{}
	m.onAnimeLibraryCollection = &Hook[hook_resolver.Resolver]{}
	m.onAnimeLibraryStreamCollectionRequested = &Hook[hook_resolver.Resolver]{}
	m.onAnimeLibraryStreamCollection = &Hook[hook_resolver.Resolver]{}
	// Auto Downloader events
	m.onAutoDownloaderQueueOrDownloadTorrent = &Hook[hook_resolver.Resolver]{}
	m.onAutoDownloaderTorrentMatched = &Hook[hook_resolver.Resolver]{}
	m.onAutoDownloaderRuleVerifyMatch = &Hook[hook_resolver.Resolver]{}
	m.onAutoDownloaderRunStarted = &Hook[hook_resolver.Resolver]{}
	m.onAutoDownloaderRunCompleted = &Hook[hook_resolver.Resolver]{}
	m.onAutoDownloaderSettingsUpdated = &Hook[hook_resolver.Resolver]{}
}

func (m *ManagerImpl) OnGetAnime() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onGetAnime
}

func (m *ManagerImpl) OnGetAnimeDetails() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onGetAnimeDetails
}

func (m *ManagerImpl) OnGetManga() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onGetManga
}

func (m *ManagerImpl) OnGetMangaDetails() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onGetMangaDetails
}

func (m *ManagerImpl) OnGetAnimeCollection() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onGetAnimeCollection
}

func (m *ManagerImpl) OnGetMangaCollection() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onGetMangaCollection
}

func (m *ManagerImpl) OnGetRawAnimeCollection() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onGetRawAnimeCollection
}

func (m *ManagerImpl) OnGetRawMangaCollection() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onGetRawMangaCollection
}

func (m *ManagerImpl) OnGetStudioDetails() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onGetStudioDetails
}

func (m *ManagerImpl) OnPreUpdateEntry() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onPreUpdateEntry
}

func (m *ManagerImpl) OnPostUpdateEntry() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onPostUpdateEntry
}

func (m *ManagerImpl) OnPreUpdateEntryProgress() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onPreUpdateEntryProgress
}

func (m *ManagerImpl) OnPostUpdateEntryProgress() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onPostUpdateEntryProgress
}

func (m *ManagerImpl) OnPreUpdateEntryRepeat() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onPreUpdateEntryRepeat
}

func (m *ManagerImpl) OnPostUpdateEntryRepeat() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onPostUpdateEntryRepeat
}

// Anime entry events

func (m *ManagerImpl) OnAnimeEntryRequested() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onAnimeEntryRequested
}

func (m *ManagerImpl) OnAnimeEntry() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onAnimeEntry
}

func (m *ManagerImpl) OnAnimeEntryFillerHydration() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onAnimeEntryFillerHydration
}

func (m *ManagerImpl) OnAnimeEntryLibraryDataRequested() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onAnimeEntryLibraryDataRequested
}

func (m *ManagerImpl) OnAnimeEntryLibraryData() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onAnimeEntryLibraryData
}

func (m *ManagerImpl) OnAnimeEntryManualMatchBeforeSave() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onAnimeEntryManualMatchBeforeSave
}

func (m *ManagerImpl) OnMissingEpisodesRequested() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onMissingEpisodesRequested
}

func (m *ManagerImpl) OnMissingEpisodes() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onMissingEpisodes
}

// Anime library collection events

func (m *ManagerImpl) OnAnimeLibraryCollectionRequested() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onAnimeLibraryCollectionRequested
}

func (m *ManagerImpl) OnAnimeLibraryStreamCollectionRequested() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onAnimeLibraryStreamCollectionRequested
}

func (m *ManagerImpl) OnAnimeLibraryCollection() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onAnimeLibraryCollection
}

func (m *ManagerImpl) OnAnimeLibraryStreamCollection() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onAnimeLibraryStreamCollection
}

// Auto Downloader events

func (m *ManagerImpl) OnAutoDownloaderQueueOrDownloadTorrent() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onAutoDownloaderQueueOrDownloadTorrent
}

func (m *ManagerImpl) OnAutoDownloaderTorrentMatched() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onAutoDownloaderTorrentMatched
}

func (m *ManagerImpl) OnAutoDownloaderRuleVerifyMatch() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onAutoDownloaderRuleVerifyMatch
}

func (m *ManagerImpl) OnAutoDownloaderRunStarted() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onAutoDownloaderRunStarted
}

func (m *ManagerImpl) OnAutoDownloaderRunCompleted() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onAutoDownloaderRunCompleted
}

func (m *ManagerImpl) OnAutoDownloaderSettingsUpdated() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onAutoDownloaderSettingsUpdated
}
