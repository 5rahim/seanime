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
	OnAutoDownloaderRunStarted() *Hook[hook_resolver.Resolver]
	OnAutoDownloaderMatchVerified() *Hook[hook_resolver.Resolver]
	OnAutoDownloaderSettingsUpdated() *Hook[hook_resolver.Resolver]
	OnAutoDownloaderTorrentsFetched() *Hook[hook_resolver.Resolver]
	OnAutoDownloaderBeforeDownloadTorrent() *Hook[hook_resolver.Resolver]
	OnAutoDownloaderAfterDownloadTorrent() *Hook[hook_resolver.Resolver]

	// Scanner events
	OnScanStarted() *Hook[hook_resolver.Resolver]
	OnScanFilePathsRetrieved() *Hook[hook_resolver.Resolver]
	OnScanLocalFilesParsed() *Hook[hook_resolver.Resolver]
	OnScanCompleted() *Hook[hook_resolver.Resolver]
	OnScanMediaFetcherStarted() *Hook[hook_resolver.Resolver]
	OnScanMediaFetcherCompleted() *Hook[hook_resolver.Resolver]
	OnScanMatchingStarted() *Hook[hook_resolver.Resolver]
	OnScanLocalFileMatched() *Hook[hook_resolver.Resolver]
	OnScanMatchingCompleted() *Hook[hook_resolver.Resolver]
	OnScanHydrationStarted() *Hook[hook_resolver.Resolver]
	OnScanLocalFileHydrationStarted() *Hook[hook_resolver.Resolver]
	OnScanLocalFileHydrated() *Hook[hook_resolver.Resolver]

	// Anime metadata events
	OnAnimeMetadataRequested() *Hook[hook_resolver.Resolver]
	OnAnimeMetadataEvent() *Hook[hook_resolver.Resolver]
	OnAnimeEpisodeMetadataRequested() *Hook[hook_resolver.Resolver]
	OnAnimeEpisodeMetadataEvent() *Hook[hook_resolver.Resolver]

	// Manga events
	OnMangaEntryRequested() *Hook[hook_resolver.Resolver]
	OnMangaEntry() *Hook[hook_resolver.Resolver]
	OnMangaLibraryCollectionRequested() *Hook[hook_resolver.Resolver]
	OnMangaLibraryCollection() *Hook[hook_resolver.Resolver]
	OnMangaDownloadedChapterContainersRequested() *Hook[hook_resolver.Resolver]
	OnMangaDownloadedChapterContainers() *Hook[hook_resolver.Resolver]
	OnMangaLatestChapterNumbersMap() *Hook[hook_resolver.Resolver]
	OnMangaDownloadMap() *Hook[hook_resolver.Resolver]
	OnMangaChapterContainerRequested() *Hook[hook_resolver.Resolver]
	OnMangaChapterContainer() *Hook[hook_resolver.Resolver]

	// Playback events
	OnLocalFilePlaybackRequested() *Hook[hook_resolver.Resolver]
	OnPlaybackBeforeTracking() *Hook[hook_resolver.Resolver]
	OnStreamPlaybackRequested() *Hook[hook_resolver.Resolver]
	OnPlaybackLocalFileDetailsRequested() *Hook[hook_resolver.Resolver]
	OnPlaybackStreamDetailsRequested() *Hook[hook_resolver.Resolver]

	// Media player events
	OnMediaPlayerLocalFileTrackingRequested() *Hook[hook_resolver.Resolver]
	OnMediaPlayerStreamTrackingRequested() *Hook[hook_resolver.Resolver]

	// Debrid events
	OnDebridAutoSelectTorrentsFetched() *Hook[hook_resolver.Resolver]
	OnDebridSendStreamToMediaPlayer() *Hook[hook_resolver.Resolver]
	OnDebridLocalDownloadRequested() *Hook[hook_resolver.Resolver]
	OnDebridSkipStreamCheck() *Hook[hook_resolver.Resolver]

	// Torrent stream events
	OnTorrentStreamAutoSelectTorrentsFetched() *Hook[hook_resolver.Resolver]
	OnTorrentStreamSendStreamToMediaPlayer() *Hook[hook_resolver.Resolver]

	// Continuity events
	OnWatchHistoryItemRequested() *Hook[hook_resolver.Resolver]
	OnWatchHistoryLocalFileEpisodeItemRequested() *Hook[hook_resolver.Resolver]
	OnWatchHistoryStreamEpisodeItemRequested() *Hook[hook_resolver.Resolver]
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
	onAutoDownloaderMatchVerified         *Hook[hook_resolver.Resolver]
	onAutoDownloaderRunStarted            *Hook[hook_resolver.Resolver]
	onAutoDownloaderRunCompleted          *Hook[hook_resolver.Resolver]
	onAutoDownloaderSettingsUpdated       *Hook[hook_resolver.Resolver]
	onAutoDownloaderTorrentsFetched       *Hook[hook_resolver.Resolver]
	onAutoDownloaderBeforeDownloadTorrent *Hook[hook_resolver.Resolver]
	onAutoDownloaderAfterDownloadTorrent  *Hook[hook_resolver.Resolver]
	// Scanner events
	onScanStarted                   *Hook[hook_resolver.Resolver]
	onScanFilePathsRetrieved        *Hook[hook_resolver.Resolver]
	onScanLocalFilesParsed          *Hook[hook_resolver.Resolver]
	onScanCompleted                 *Hook[hook_resolver.Resolver]
	onScanMediaFetcherStarted       *Hook[hook_resolver.Resolver]
	onScanMediaFetcherCompleted     *Hook[hook_resolver.Resolver]
	onScanMatchingStarted           *Hook[hook_resolver.Resolver]
	onScanLocalFileMatched          *Hook[hook_resolver.Resolver]
	onScanMatchingCompleted         *Hook[hook_resolver.Resolver]
	onScanHydrationStarted          *Hook[hook_resolver.Resolver]
	onScanLocalFileHydrationStarted *Hook[hook_resolver.Resolver]
	onScanLocalFileHydrated         *Hook[hook_resolver.Resolver]
	// Anime metadata events
	onAnimeMetadataRequested        *Hook[hook_resolver.Resolver]
	onAnimeMetadataEvent            *Hook[hook_resolver.Resolver]
	onAnimeEpisodeMetadataRequested *Hook[hook_resolver.Resolver]
	onAnimeEpisodeMetadataEvent     *Hook[hook_resolver.Resolver]
	// Manga events
	onMangaEntryRequested                       *Hook[hook_resolver.Resolver]
	onMangaEntry                                *Hook[hook_resolver.Resolver]
	onMangaLibraryCollectionRequested           *Hook[hook_resolver.Resolver]
	onMangaLibraryCollection                    *Hook[hook_resolver.Resolver]
	onMangaDownloadedChapterContainersRequested *Hook[hook_resolver.Resolver]
	onMangaDownloadedChapterContainers          *Hook[hook_resolver.Resolver]
	onMangaLatestChapterNumbersMap              *Hook[hook_resolver.Resolver]
	onMangaDownloadMap                          *Hook[hook_resolver.Resolver]
	onMangaChapterContainerRequested            *Hook[hook_resolver.Resolver]
	onMangaChapterContainer                     *Hook[hook_resolver.Resolver]
	// Playback events
	onLocalFilePlaybackRequested        *Hook[hook_resolver.Resolver]
	onPlaybackBeforeTracking            *Hook[hook_resolver.Resolver]
	onStreamPlaybackRequested           *Hook[hook_resolver.Resolver]
	onPlaybackLocalFileDetailsRequested *Hook[hook_resolver.Resolver]
	onPlaybackStreamDetailsRequested    *Hook[hook_resolver.Resolver]
	// Media player events
	onMediaPlayerLocalFileTrackingRequested *Hook[hook_resolver.Resolver]
	onMediaPlayerStreamTrackingRequested    *Hook[hook_resolver.Resolver]
	// Debrid events
	onDebridAutoSelectTorrentsFetched *Hook[hook_resolver.Resolver]
	onDebridSendStreamToMediaPlayer   *Hook[hook_resolver.Resolver]
	onDebridLocalDownloadRequested    *Hook[hook_resolver.Resolver]
	onDebridSkipStreamCheck           *Hook[hook_resolver.Resolver]
	// Torrent stream events
	onTorrentStreamAutoSelectTorrentsFetched *Hook[hook_resolver.Resolver]
	onTorrentStreamSendStreamToMediaPlayer   *Hook[hook_resolver.Resolver]
	// Continuity events
	onWatchHistoryItemRequested                 *Hook[hook_resolver.Resolver]
	onWatchHistoryLocalFileEpisodeItemRequested *Hook[hook_resolver.Resolver]
	onWatchHistoryStreamEpisodeItemRequested    *Hook[hook_resolver.Resolver]
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
	m.onAutoDownloaderMatchVerified = &Hook[hook_resolver.Resolver]{}
	m.onAutoDownloaderRunStarted = &Hook[hook_resolver.Resolver]{}
	m.onAutoDownloaderRunCompleted = &Hook[hook_resolver.Resolver]{}
	m.onAutoDownloaderSettingsUpdated = &Hook[hook_resolver.Resolver]{}
	m.onAutoDownloaderTorrentsFetched = &Hook[hook_resolver.Resolver]{}
	m.onAutoDownloaderBeforeDownloadTorrent = &Hook[hook_resolver.Resolver]{}
	m.onAutoDownloaderAfterDownloadTorrent = &Hook[hook_resolver.Resolver]{}
	// Scanner events
	m.onScanStarted = &Hook[hook_resolver.Resolver]{}
	m.onScanFilePathsRetrieved = &Hook[hook_resolver.Resolver]{}
	m.onScanLocalFilesParsed = &Hook[hook_resolver.Resolver]{}
	m.onScanCompleted = &Hook[hook_resolver.Resolver]{}
	m.onScanMediaFetcherStarted = &Hook[hook_resolver.Resolver]{}
	m.onScanMediaFetcherCompleted = &Hook[hook_resolver.Resolver]{}
	m.onScanMatchingStarted = &Hook[hook_resolver.Resolver]{}
	m.onScanLocalFileMatched = &Hook[hook_resolver.Resolver]{}
	m.onScanMatchingCompleted = &Hook[hook_resolver.Resolver]{}
	m.onScanHydrationStarted = &Hook[hook_resolver.Resolver]{}
	m.onScanLocalFileHydrationStarted = &Hook[hook_resolver.Resolver]{}
	m.onScanLocalFileHydrated = &Hook[hook_resolver.Resolver]{}
	// Anime metadata events
	m.onAnimeMetadataRequested = &Hook[hook_resolver.Resolver]{}
	m.onAnimeMetadataEvent = &Hook[hook_resolver.Resolver]{}
	m.onAnimeEpisodeMetadataRequested = &Hook[hook_resolver.Resolver]{}
	m.onAnimeEpisodeMetadataEvent = &Hook[hook_resolver.Resolver]{}
	// Manga events
	m.onMangaEntryRequested = &Hook[hook_resolver.Resolver]{}
	m.onMangaEntry = &Hook[hook_resolver.Resolver]{}
	m.onMangaLibraryCollectionRequested = &Hook[hook_resolver.Resolver]{}
	m.onMangaLibraryCollection = &Hook[hook_resolver.Resolver]{}
	m.onMangaDownloadedChapterContainersRequested = &Hook[hook_resolver.Resolver]{}
	m.onMangaDownloadedChapterContainers = &Hook[hook_resolver.Resolver]{}
	m.onMangaLatestChapterNumbersMap = &Hook[hook_resolver.Resolver]{}
	m.onMangaDownloadMap = &Hook[hook_resolver.Resolver]{}
	m.onMangaChapterContainerRequested = &Hook[hook_resolver.Resolver]{}
	m.onMangaChapterContainer = &Hook[hook_resolver.Resolver]{}
	// Playback events
	m.onLocalFilePlaybackRequested = &Hook[hook_resolver.Resolver]{}
	m.onPlaybackBeforeTracking = &Hook[hook_resolver.Resolver]{}
	m.onStreamPlaybackRequested = &Hook[hook_resolver.Resolver]{}
	m.onPlaybackLocalFileDetailsRequested = &Hook[hook_resolver.Resolver]{}
	m.onPlaybackStreamDetailsRequested = &Hook[hook_resolver.Resolver]{}
	// Media player events
	m.onMediaPlayerLocalFileTrackingRequested = &Hook[hook_resolver.Resolver]{}
	m.onMediaPlayerStreamTrackingRequested = &Hook[hook_resolver.Resolver]{}
	// Debrid events
	m.onDebridAutoSelectTorrentsFetched = &Hook[hook_resolver.Resolver]{}
	m.onDebridSendStreamToMediaPlayer = &Hook[hook_resolver.Resolver]{}
	m.onDebridLocalDownloadRequested = &Hook[hook_resolver.Resolver]{}
	m.onDebridSkipStreamCheck = &Hook[hook_resolver.Resolver]{}
	// Torrent stream events
	m.onTorrentStreamAutoSelectTorrentsFetched = &Hook[hook_resolver.Resolver]{}
	m.onTorrentStreamSendStreamToMediaPlayer = &Hook[hook_resolver.Resolver]{}
	// Continuity events
	m.onWatchHistoryItemRequested = &Hook[hook_resolver.Resolver]{}
	m.onWatchHistoryLocalFileEpisodeItemRequested = &Hook[hook_resolver.Resolver]{}
	m.onWatchHistoryStreamEpisodeItemRequested = &Hook[hook_resolver.Resolver]{}
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

func (m *ManagerImpl) OnAutoDownloaderMatchVerified() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onAutoDownloaderMatchVerified
}

func (m *ManagerImpl) OnAutoDownloaderRunStarted() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onAutoDownloaderRunStarted
}

func (m *ManagerImpl) OnAutoDownloaderSettingsUpdated() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onAutoDownloaderSettingsUpdated
}

func (m *ManagerImpl) OnAutoDownloaderTorrentsFetched() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onAutoDownloaderTorrentsFetched
}

func (m *ManagerImpl) OnAutoDownloaderBeforeDownloadTorrent() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onAutoDownloaderBeforeDownloadTorrent
}

func (m *ManagerImpl) OnAutoDownloaderAfterDownloadTorrent() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onAutoDownloaderAfterDownloadTorrent
}

// Scanner events
func (m *ManagerImpl) OnScanStarted() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onScanStarted
}

func (m *ManagerImpl) OnScanFilePathsRetrieved() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onScanFilePathsRetrieved
}

func (m *ManagerImpl) OnScanLocalFilesParsed() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onScanLocalFilesParsed
}

func (m *ManagerImpl) OnScanCompleted() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onScanCompleted
}

func (m *ManagerImpl) OnScanMediaFetcherStarted() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onScanMediaFetcherStarted
}

func (m *ManagerImpl) OnScanMediaFetcherCompleted() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onScanMediaFetcherCompleted
}

func (m *ManagerImpl) OnScanMatchingStarted() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onScanMatchingStarted
}

func (m *ManagerImpl) OnScanLocalFileMatched() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onScanLocalFileMatched
}

func (m *ManagerImpl) OnScanMatchingCompleted() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onScanMatchingCompleted
}

func (m *ManagerImpl) OnScanHydrationStarted() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onScanHydrationStarted
}

func (m *ManagerImpl) OnScanLocalFileHydrationStarted() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onScanLocalFileHydrationStarted
}

func (m *ManagerImpl) OnScanLocalFileHydrated() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onScanLocalFileHydrated
}

// Anime metadata events

func (m *ManagerImpl) OnAnimeMetadataRequested() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onAnimeMetadataRequested
}

func (m *ManagerImpl) OnAnimeMetadataEvent() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onAnimeMetadataEvent
}

func (m *ManagerImpl) OnAnimeEpisodeMetadataRequested() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onAnimeEpisodeMetadataRequested
}

func (m *ManagerImpl) OnAnimeEpisodeMetadataEvent() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onAnimeEpisodeMetadataEvent
}

// Manga events

func (m *ManagerImpl) OnMangaEntryRequested() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onMangaEntryRequested
}

func (m *ManagerImpl) OnMangaEntry() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onMangaEntry
}

func (m *ManagerImpl) OnMangaLibraryCollectionRequested() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onMangaLibraryCollectionRequested
}

func (m *ManagerImpl) OnMangaLibraryCollection() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onMangaLibraryCollection
}

func (m *ManagerImpl) OnMangaDownloadedChapterContainersRequested() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onMangaDownloadedChapterContainersRequested
}

func (m *ManagerImpl) OnMangaDownloadedChapterContainers() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onMangaDownloadedChapterContainers
}

func (m *ManagerImpl) OnMangaLatestChapterNumbersMap() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onMangaLatestChapterNumbersMap
}

func (m *ManagerImpl) OnMangaDownloadMap() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onMangaDownloadMap
}

func (m *ManagerImpl) OnMangaChapterContainerRequested() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onMangaChapterContainerRequested
}

func (m *ManagerImpl) OnMangaChapterContainer() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onMangaChapterContainer
}

// Playback events

func (m *ManagerImpl) OnLocalFilePlaybackRequested() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onLocalFilePlaybackRequested
}

func (m *ManagerImpl) OnPlaybackBeforeTracking() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onPlaybackBeforeTracking
}

func (m *ManagerImpl) OnStreamPlaybackRequested() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onStreamPlaybackRequested
}

func (m *ManagerImpl) OnPlaybackLocalFileDetailsRequested() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onPlaybackLocalFileDetailsRequested
}

func (m *ManagerImpl) OnPlaybackStreamDetailsRequested() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onPlaybackStreamDetailsRequested
}

// Media player events

func (m *ManagerImpl) OnMediaPlayerLocalFileTrackingRequested() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onMediaPlayerLocalFileTrackingRequested
}

func (m *ManagerImpl) OnMediaPlayerStreamTrackingRequested() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onMediaPlayerStreamTrackingRequested
}

// Debrid events

func (m *ManagerImpl) OnDebridAutoSelectTorrentsFetched() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onDebridAutoSelectTorrentsFetched
}

func (m *ManagerImpl) OnDebridSendStreamToMediaPlayer() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onDebridSendStreamToMediaPlayer
}

func (m *ManagerImpl) OnDebridLocalDownloadRequested() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onDebridLocalDownloadRequested
}

func (m *ManagerImpl) OnDebridSkipStreamCheck() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onDebridSkipStreamCheck
}

// Torrent stream events

func (m *ManagerImpl) OnTorrentStreamAutoSelectTorrentsFetched() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onTorrentStreamAutoSelectTorrentsFetched
}

func (m *ManagerImpl) OnTorrentStreamSendStreamToMediaPlayer() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onTorrentStreamSendStreamToMediaPlayer
}

// Continuity events

func (m *ManagerImpl) OnWatchHistoryItemRequested() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onWatchHistoryItemRequested
}

func (m *ManagerImpl) OnWatchHistoryLocalFileEpisodeItemRequested() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onWatchHistoryLocalFileEpisodeItemRequested
}

func (m *ManagerImpl) OnWatchHistoryStreamEpisodeItemRequested() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onWatchHistoryStreamEpisodeItemRequested
}
