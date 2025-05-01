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
	OnGetCachedAnimeCollection() *Hook[hook_resolver.Resolver]
	OnGetCachedMangaCollection() *Hook[hook_resolver.Resolver]
	OnGetRawAnimeCollection() *Hook[hook_resolver.Resolver]
	OnGetRawMangaCollection() *Hook[hook_resolver.Resolver]
	OnGetCachedRawAnimeCollection() *Hook[hook_resolver.Resolver]
	OnGetCachedRawMangaCollection() *Hook[hook_resolver.Resolver]
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
	OnAnimeMetadata() *Hook[hook_resolver.Resolver]
	OnAnimeEpisodeMetadataRequested() *Hook[hook_resolver.Resolver]
	OnAnimeEpisodeMetadata() *Hook[hook_resolver.Resolver]

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
	OnWatchHistoryItemUpdated() *Hook[hook_resolver.Resolver]
	OnWatchHistoryLocalFileEpisodeItemRequested() *Hook[hook_resolver.Resolver]
	OnWatchHistoryStreamEpisodeItemRequested() *Hook[hook_resolver.Resolver]

	// Discord RPC events
	OnDiscordPresenceAnimeActivityRequested() *Hook[hook_resolver.Resolver]
	OnDiscordPresenceMangaActivityRequested() *Hook[hook_resolver.Resolver]
	OnDiscordPresenceClientClosed() *Hook[hook_resolver.Resolver]

	// Handlers
	//OnHandleGetAnimeCollectionRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetAnimeCollection() *Hook[hook_resolver.Resolver]
	//OnHandleGetRawAnimeCollectionRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetRawAnimeCollection() *Hook[hook_resolver.Resolver]
	//OnHandleEditAnilistListEntryRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetAnilistAnimeDetailsRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetAnilistAnimeDetails() *Hook[hook_resolver.Resolver]
	//OnHandleGetAnilistStudioDetailsRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetAnilistStudioDetails() *Hook[hook_resolver.Resolver]
	//OnHandleDeleteAnilistListEntryRequested() *Hook[hook_resolver.Resolver]
	//OnHandleAnilistListAnimeRequested() *Hook[hook_resolver.Resolver]
	//OnHandleAnilistListAnime() *Hook[hook_resolver.Resolver]
	//OnHandleAnilistListRecentAiringAnimeRequested() *Hook[hook_resolver.Resolver]
	//OnHandleAnilistListRecentAiringAnime() *Hook[hook_resolver.Resolver]
	//OnHandleAnilistListMissedSequelsRequested() *Hook[hook_resolver.Resolver]
	//OnHandleAnilistListMissedSequels() *Hook[hook_resolver.Resolver]
	//OnHandleGetAniListStatsRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetAniListStats() *Hook[hook_resolver.Resolver]
	//OnHandleGetLibraryCollectionRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetLibraryCollection() *Hook[hook_resolver.Resolver]
	//OnHandleAddUnknownMediaRequested() *Hook[hook_resolver.Resolver]
	//OnHandleAddUnknownMedia() *Hook[hook_resolver.Resolver]
	//OnHandleGetAnimeEntryRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetAnimeEntry() *Hook[hook_resolver.Resolver]
	//OnHandleAnimeEntryBulkActionRequested() *Hook[hook_resolver.Resolver]
	//OnHandleAnimeEntryBulkAction() *Hook[hook_resolver.Resolver]
	//OnHandleOpenAnimeEntryInExplorerRequested() *Hook[hook_resolver.Resolver]
	//OnHandleFetchAnimeEntrySuggestionsRequested() *Hook[hook_resolver.Resolver]
	//OnHandleFetchAnimeEntrySuggestions() *Hook[hook_resolver.Resolver]
	//OnHandleAnimeEntryManualMatchRequested() *Hook[hook_resolver.Resolver]
	//OnHandleAnimeEntryManualMatch() *Hook[hook_resolver.Resolver]
	//OnHandleGetMissingEpisodesRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetMissingEpisodes() *Hook[hook_resolver.Resolver]
	//OnHandleGetAnimeEntrySilenceStatusRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetAnimeEntrySilenceStatus() *Hook[hook_resolver.Resolver]
	//OnHandleToggleAnimeEntrySilenceStatusRequested() *Hook[hook_resolver.Resolver]
	//OnHandleUpdateAnimeEntryProgressRequested() *Hook[hook_resolver.Resolver]
	//OnHandleUpdateAnimeEntryRepeatRequested() *Hook[hook_resolver.Resolver]
	//OnHandleLoginRequested() *Hook[hook_resolver.Resolver]
	//OnHandleLogin() *Hook[hook_resolver.Resolver]
	//OnHandleLogoutRequested() *Hook[hook_resolver.Resolver]
	//OnHandleLogout() *Hook[hook_resolver.Resolver]
	//OnHandleRunAutoDownloaderRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetAutoDownloaderRuleRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetAutoDownloaderRule() *Hook[hook_resolver.Resolver]
	//OnHandleGetAutoDownloaderRulesByAnimeRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetAutoDownloaderRulesByAnime() *Hook[hook_resolver.Resolver]
	//OnHandleGetAutoDownloaderRulesRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetAutoDownloaderRules() *Hook[hook_resolver.Resolver]
	//OnHandleCreateAutoDownloaderRuleRequested() *Hook[hook_resolver.Resolver]
	//OnHandleCreateAutoDownloaderRule() *Hook[hook_resolver.Resolver]
	//OnHandleUpdateAutoDownloaderRuleRequested() *Hook[hook_resolver.Resolver]
	//OnHandleUpdateAutoDownloaderRule() *Hook[hook_resolver.Resolver]
	//OnHandleDeleteAutoDownloaderRuleRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetAutoDownloaderItemsRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetAutoDownloaderItems() *Hook[hook_resolver.Resolver]
	//OnHandleDeleteAutoDownloaderItemRequested() *Hook[hook_resolver.Resolver]
	//OnHandleUpdateContinuityWatchHistoryItemRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetContinuityWatchHistoryItemRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetContinuityWatchHistoryItem() *Hook[hook_resolver.Resolver]
	//OnHandleGetContinuityWatchHistoryRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetContinuityWatchHistory() *Hook[hook_resolver.Resolver]
	//OnHandleGetDebridSettingsRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetDebridSettings() *Hook[hook_resolver.Resolver]
	//OnHandleSaveDebridSettingsRequested() *Hook[hook_resolver.Resolver]
	//OnHandleSaveDebridSettings() *Hook[hook_resolver.Resolver]
	//OnHandleDebridAddTorrentsRequested() *Hook[hook_resolver.Resolver]
	//OnHandleDebridDownloadTorrentRequested() *Hook[hook_resolver.Resolver]
	//OnHandleDebridCancelDownloadRequested() *Hook[hook_resolver.Resolver]
	//OnHandleDebridDeleteTorrentRequested() *Hook[hook_resolver.Resolver]
	//OnHandleDebridGetTorrentsRequested() *Hook[hook_resolver.Resolver]
	//OnHandleDebridGetTorrents() *Hook[hook_resolver.Resolver]
	//OnHandleDebridGetTorrentInfoRequested() *Hook[hook_resolver.Resolver]
	//OnHandleDebridGetTorrentInfo() *Hook[hook_resolver.Resolver]
	//OnHandleDebridGetTorrentFilePreviewsRequested() *Hook[hook_resolver.Resolver]
	//OnHandleDebridGetTorrentFilePreviews() *Hook[hook_resolver.Resolver]
	//OnHandleDebridStartStreamRequested() *Hook[hook_resolver.Resolver]
	//OnHandleDebridCancelStreamRequested() *Hook[hook_resolver.Resolver]
	//OnHandleDirectorySelectorRequested() *Hook[hook_resolver.Resolver]
	//OnHandleDirectorySelector() *Hook[hook_resolver.Resolver]
	//OnHandleSetDiscordMangaActivityRequested() *Hook[hook_resolver.Resolver]
	//OnHandleSetDiscordLegacyAnimeActivityRequested() *Hook[hook_resolver.Resolver]
	//OnHandleSetDiscordAnimeActivityWithProgressRequested() *Hook[hook_resolver.Resolver]
	//OnHandleUpdateDiscordAnimeActivityWithProgressRequested() *Hook[hook_resolver.Resolver]
	//OnHandleCancelDiscordActivityRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetDocsRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetDocs() *Hook[hook_resolver.Resolver]
	//OnHandleDownloadTorrentFileRequested() *Hook[hook_resolver.Resolver]
	//OnHandleDownloadReleaseRequested() *Hook[hook_resolver.Resolver]
	//OnHandleDownloadRelease() *Hook[hook_resolver.Resolver]
	//OnHandleOpenInExplorerRequested() *Hook[hook_resolver.Resolver]
	//OnHandleFetchExternalExtensionDataRequested() *Hook[hook_resolver.Resolver]
	//OnHandleFetchExternalExtensionData() *Hook[hook_resolver.Resolver]
	//OnHandleInstallExternalExtensionRequested() *Hook[hook_resolver.Resolver]
	//OnHandleInstallExternalExtension() *Hook[hook_resolver.Resolver]
	//OnHandleUninstallExternalExtensionRequested() *Hook[hook_resolver.Resolver]
	//OnHandleUpdateExtensionCodeRequested() *Hook[hook_resolver.Resolver]
	//OnHandleReloadExternalExtensionsRequested() *Hook[hook_resolver.Resolver]
	//OnHandleReloadExternalExtensionRequested() *Hook[hook_resolver.Resolver]
	//OnHandleListExtensionDataRequested() *Hook[hook_resolver.Resolver]
	//OnHandleListExtensionData() *Hook[hook_resolver.Resolver]
	//OnHandleGetExtensionPayloadRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetExtensionPayload() *Hook[hook_resolver.Resolver]
	//OnHandleListDevelopmentModeExtensionsRequested() *Hook[hook_resolver.Resolver]
	//OnHandleListDevelopmentModeExtensions() *Hook[hook_resolver.Resolver]
	//OnHandleGetAllExtensionsRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetAllExtensions() *Hook[hook_resolver.Resolver]
	//OnHandleGetExtensionUpdateDataRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetExtensionUpdateData() *Hook[hook_resolver.Resolver]
	//OnHandleListMangaProviderExtensionsRequested() *Hook[hook_resolver.Resolver]
	//OnHandleListMangaProviderExtensions() *Hook[hook_resolver.Resolver]
	//OnHandleListOnlinestreamProviderExtensionsRequested() *Hook[hook_resolver.Resolver]
	//OnHandleListOnlinestreamProviderExtensions() *Hook[hook_resolver.Resolver]
	//OnHandleListAnimeTorrentProviderExtensionsRequested() *Hook[hook_resolver.Resolver]
	//OnHandleListAnimeTorrentProviderExtensions() *Hook[hook_resolver.Resolver]
	//OnHandleGetPluginSettingsRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetPluginSettings() *Hook[hook_resolver.Resolver]
	//OnHandleSetPluginSettingsPinnedTraysRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGrantPluginPermissionsRequested() *Hook[hook_resolver.Resolver]
	//OnHandleRunExtensionPlaygroundCodeRequested() *Hook[hook_resolver.Resolver]
	//OnHandleRunExtensionPlaygroundCode() *Hook[hook_resolver.Resolver]
	//OnHandleGetExtensionUserConfigRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetExtensionUserConfig() *Hook[hook_resolver.Resolver]
	//OnHandleSaveExtensionUserConfigRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetMarketplaceExtensionsRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetMarketplaceExtensions() *Hook[hook_resolver.Resolver]
	//OnHandleGetFileCacheTotalSizeRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetFileCacheTotalSize() *Hook[hook_resolver.Resolver]
	//OnHandleRemoveFileCacheBucketRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetFileCacheMediastreamVideoFilesTotalSizeRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetFileCacheMediastreamVideoFilesTotalSize() *Hook[hook_resolver.Resolver]
	//OnHandleClearFileCacheMediastreamVideoFilesRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetLocalFilesRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetLocalFiles() *Hook[hook_resolver.Resolver]
	//OnHandleImportLocalFilesRequested() *Hook[hook_resolver.Resolver]
	//OnHandleLocalFileBulkActionRequested() *Hook[hook_resolver.Resolver]
	//OnHandleLocalFileBulkAction() *Hook[hook_resolver.Resolver]
	//OnHandleUpdateLocalFileDataRequested() *Hook[hook_resolver.Resolver]
	//OnHandleUpdateLocalFileData() *Hook[hook_resolver.Resolver]
	//OnHandleUpdateLocalFilesRequested() *Hook[hook_resolver.Resolver]
	//OnHandleDeleteLocalFilesRequested() *Hook[hook_resolver.Resolver]
	//OnHandleRemoveEmptyDirectoriesRequested() *Hook[hook_resolver.Resolver]
	//OnHandleMALAuthRequested() *Hook[hook_resolver.Resolver]
	//OnHandleMALAuth() *Hook[hook_resolver.Resolver]
	//OnHandleEditMALListEntryProgressRequested() *Hook[hook_resolver.Resolver]
	//OnHandleMALLogoutRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetAnilistMangaCollectionRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetAnilistMangaCollection() *Hook[hook_resolver.Resolver]
	//OnHandleGetRawAnilistMangaCollectionRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetRawAnilistMangaCollection() *Hook[hook_resolver.Resolver]
	//OnHandleGetMangaCollectionRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetMangaCollection() *Hook[hook_resolver.Resolver]
	//OnHandleGetMangaEntryRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetMangaEntry() *Hook[hook_resolver.Resolver]
	//OnHandleGetMangaEntryDetailsRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetMangaEntryDetails() *Hook[hook_resolver.Resolver]
	//OnHandleGetMangaLatestChapterNumbersMapRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetMangaLatestChapterNumbersMap() *Hook[hook_resolver.Resolver]
	//OnHandleRefetchMangaChapterContainersRequested() *Hook[hook_resolver.Resolver]
	//OnHandleEmptyMangaEntryCacheRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetMangaEntryChaptersRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetMangaEntryChapters() *Hook[hook_resolver.Resolver]
	//OnHandleGetMangaEntryPagesRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetMangaEntryPages() *Hook[hook_resolver.Resolver]
	//OnHandleGetMangaEntryDownloadedChaptersRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetMangaEntryDownloadedChapters() *Hook[hook_resolver.Resolver]
	//OnHandleAnilistListMangaRequested() *Hook[hook_resolver.Resolver]
	//OnHandleAnilistListManga() *Hook[hook_resolver.Resolver]
	//OnHandleUpdateMangaProgressRequested() *Hook[hook_resolver.Resolver]
	//OnHandleMangaManualSearchRequested() *Hook[hook_resolver.Resolver]
	//OnHandleMangaManualSearch() *Hook[hook_resolver.Resolver]
	//OnHandleMangaManualMappingRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetMangaMappingRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetMangaMapping() *Hook[hook_resolver.Resolver]
	//OnHandleRemoveMangaMappingRequested() *Hook[hook_resolver.Resolver]
	//OnHandleDownloadMangaChaptersRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetMangaDownloadDataRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetMangaDownloadData() *Hook[hook_resolver.Resolver]
	//OnHandleGetMangaDownloadQueueRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetMangaDownloadQueue() *Hook[hook_resolver.Resolver]
	//OnHandleStartMangaDownloadQueueRequested() *Hook[hook_resolver.Resolver]
	//OnHandleStopMangaDownloadQueueRequested() *Hook[hook_resolver.Resolver]
	//OnHandleClearAllChapterDownloadQueueRequested() *Hook[hook_resolver.Resolver]
	//OnHandleResetErroredChapterDownloadQueueRequested() *Hook[hook_resolver.Resolver]
	//OnHandleDeleteMangaDownloadedChaptersRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetMangaDownloadsListRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetMangaDownloadsList() *Hook[hook_resolver.Resolver]
	//OnHandleTestDumpRequested() *Hook[hook_resolver.Resolver]
	//OnHandleStartDefaultMediaPlayerRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetMediastreamSettingsRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetMediastreamSettings() *Hook[hook_resolver.Resolver]
	//OnHandleSaveMediastreamSettingsRequested() *Hook[hook_resolver.Resolver]
	//OnHandleSaveMediastreamSettings() *Hook[hook_resolver.Resolver]
	//OnHandleRequestMediastreamMediaContainerRequested() *Hook[hook_resolver.Resolver]
	//OnHandleRequestMediastreamMediaContainer() *Hook[hook_resolver.Resolver]
	//OnHandlePreloadMediastreamMediaContainerRequested() *Hook[hook_resolver.Resolver]
	//OnHandleMediastreamShutdownTranscodeStreamRequested() *Hook[hook_resolver.Resolver]
	//OnHandlePopulateTVDBEpisodesRequested() *Hook[hook_resolver.Resolver]
	//OnHandlePopulateTVDBEpisodes() *Hook[hook_resolver.Resolver]
	//OnHandleEmptyTVDBEpisodesRequested() *Hook[hook_resolver.Resolver]
	//OnHandlePopulateFillerDataRequested() *Hook[hook_resolver.Resolver]
	//OnHandleRemoveFillerDataRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetOnlineStreamEpisodeListRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetOnlineStreamEpisodeList() *Hook[hook_resolver.Resolver]
	//OnHandleGetOnlineStreamEpisodeSourceRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetOnlineStreamEpisodeSource() *Hook[hook_resolver.Resolver]
	//OnHandleOnlineStreamEmptyCacheRequested() *Hook[hook_resolver.Resolver]
	//OnHandleOnlinestreamManualSearchRequested() *Hook[hook_resolver.Resolver]
	//OnHandleOnlinestreamManualSearch() *Hook[hook_resolver.Resolver]
	//OnHandleOnlinestreamManualMappingRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetOnlinestreamMappingRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetOnlinestreamMapping() *Hook[hook_resolver.Resolver]
	//OnHandleRemoveOnlinestreamMappingRequested() *Hook[hook_resolver.Resolver]
	//OnHandlePlaybackPlayVideoRequested() *Hook[hook_resolver.Resolver]
	//OnHandlePlaybackPlayRandomVideoRequested() *Hook[hook_resolver.Resolver]
	//OnHandlePlaybackSyncCurrentProgressRequested() *Hook[hook_resolver.Resolver]
	//OnHandlePlaybackSyncCurrentProgress() *Hook[hook_resolver.Resolver]
	//OnHandlePlaybackPlayNextEpisodeRequested() *Hook[hook_resolver.Resolver]
	//OnHandlePlaybackGetNextEpisodeRequested() *Hook[hook_resolver.Resolver]
	//OnHandlePlaybackGetNextEpisode() *Hook[hook_resolver.Resolver]
	//OnHandlePlaybackAutoPlayNextEpisodeRequested() *Hook[hook_resolver.Resolver]
	//OnHandlePlaybackStartPlaylistRequested() *Hook[hook_resolver.Resolver]
	//OnHandlePlaybackCancelCurrentPlaylistRequested() *Hook[hook_resolver.Resolver]
	//OnHandlePlaybackPlaylistNextRequested() *Hook[hook_resolver.Resolver]
	//OnHandlePlaybackStartManualTrackingRequested() *Hook[hook_resolver.Resolver]
	//OnHandlePlaybackCancelManualTrackingRequested() *Hook[hook_resolver.Resolver]
	//OnHandleCreatePlaylistRequested() *Hook[hook_resolver.Resolver]
	//OnHandleCreatePlaylist() *Hook[hook_resolver.Resolver]
	//OnHandleGetPlaylistsRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetPlaylists() *Hook[hook_resolver.Resolver]
	//OnHandleUpdatePlaylistRequested() *Hook[hook_resolver.Resolver]
	//OnHandleUpdatePlaylist() *Hook[hook_resolver.Resolver]
	//OnHandleDeletePlaylistRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetPlaylistEpisodesRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetPlaylistEpisodes() *Hook[hook_resolver.Resolver]
	//OnHandleInstallLatestUpdateRequested() *Hook[hook_resolver.Resolver]
	//OnHandleInstallLatestUpdate() *Hook[hook_resolver.Resolver]
	//OnHandleGetLatestUpdateRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetLatestUpdate() *Hook[hook_resolver.Resolver]
	//OnHandleGetChangelogRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetChangelog() *Hook[hook_resolver.Resolver]
	//OnHandleSaveIssueReportRequested() *Hook[hook_resolver.Resolver]
	//OnHandleDownloadIssueReportRequested() *Hook[hook_resolver.Resolver]
	//OnHandleDownloadIssueReport() *Hook[hook_resolver.Resolver]
	//OnHandleScanLocalFilesRequested() *Hook[hook_resolver.Resolver]
	//OnHandleScanLocalFiles() *Hook[hook_resolver.Resolver]
	//OnHandleGetScanSummariesRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetScanSummaries() *Hook[hook_resolver.Resolver]
	//OnHandleGetSettingsRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetSettings() *Hook[hook_resolver.Resolver]
	//OnHandleGettingStartedRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGettingStarted() *Hook[hook_resolver.Resolver]
	//OnHandleSaveSettingsRequested() *Hook[hook_resolver.Resolver]
	//OnHandleSaveSettings() *Hook[hook_resolver.Resolver]
	//OnHandleSaveAutoDownloaderSettingsRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetStatusRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetStatus() *Hook[hook_resolver.Resolver]
	//OnHandleGetLogFilenamesRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetLogFilenames() *Hook[hook_resolver.Resolver]
	//OnHandleDeleteLogsRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetLatestLogContentRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetLatestLogContent() *Hook[hook_resolver.Resolver]
	//OnHandleSyncGetTrackedMediaItemsRequested() *Hook[hook_resolver.Resolver]
	//OnHandleSyncGetTrackedMediaItems() *Hook[hook_resolver.Resolver]
	//OnHandleSyncAddMediaRequested() *Hook[hook_resolver.Resolver]
	//OnHandleSyncRemoveMediaRequested() *Hook[hook_resolver.Resolver]
	//OnHandleSyncGetIsMediaTrackedRequested() *Hook[hook_resolver.Resolver]
	//OnHandleSyncLocalDataRequested() *Hook[hook_resolver.Resolver]
	//OnHandleSyncGetQueueStateRequested() *Hook[hook_resolver.Resolver]
	//OnHandleSyncGetQueueState() *Hook[hook_resolver.Resolver]
	//OnHandleSyncAnilistDataRequested() *Hook[hook_resolver.Resolver]
	//OnHandleSyncSetHasLocalChangesRequested() *Hook[hook_resolver.Resolver]
	//OnHandleSyncGetHasLocalChangesRequested() *Hook[hook_resolver.Resolver]
	//OnHandleSyncGetLocalStorageSizeRequested() *Hook[hook_resolver.Resolver]
	//OnHandleSyncGetLocalStorageSize() *Hook[hook_resolver.Resolver]
	//OnHandleGetThemeRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetTheme() *Hook[hook_resolver.Resolver]
	//OnHandleUpdateThemeRequested() *Hook[hook_resolver.Resolver]
	//OnHandleUpdateTheme() *Hook[hook_resolver.Resolver]
	//OnHandleGetActiveTorrentListRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetActiveTorrentList() *Hook[hook_resolver.Resolver]
	//OnHandleTorrentClientActionRequested() *Hook[hook_resolver.Resolver]
	//OnHandleTorrentClientDownloadRequested() *Hook[hook_resolver.Resolver]
	//OnHandleTorrentClientAddMagnetFromRuleRequested() *Hook[hook_resolver.Resolver]
	//OnHandleSearchTorrentRequested() *Hook[hook_resolver.Resolver]
	//OnHandleSearchTorrent() *Hook[hook_resolver.Resolver]
	//OnHandleGetTorrentstreamEpisodeCollectionRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetTorrentstreamEpisodeCollection() *Hook[hook_resolver.Resolver]
	//OnHandleGetTorrentstreamSettingsRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetTorrentstreamSettings() *Hook[hook_resolver.Resolver]
	//OnHandleSaveTorrentstreamSettingsRequested() *Hook[hook_resolver.Resolver]
	//OnHandleSaveTorrentstreamSettings() *Hook[hook_resolver.Resolver]
	//OnHandleGetTorrentstreamTorrentFilePreviewsRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetTorrentstreamTorrentFilePreviews() *Hook[hook_resolver.Resolver]
	//OnHandleTorrentstreamStartStreamRequested() *Hook[hook_resolver.Resolver]
	//OnHandleTorrentstreamStopStreamRequested() *Hook[hook_resolver.Resolver]
	//OnHandleTorrentstreamDropTorrentRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetTorrentstreamBatchHistoryRequested() *Hook[hook_resolver.Resolver]
	//OnHandleGetTorrentstreamBatchHistory() *Hook[hook_resolver.Resolver]
	//OnHandleTorrentstreamServeStreamRequested() *Hook[hook_resolver.Resolver]
}

type ManagerImpl struct {
	logger *zerolog.Logger
	// AniList events
	onGetAnime                    *Hook[hook_resolver.Resolver]
	onGetAnimeDetails             *Hook[hook_resolver.Resolver]
	onGetManga                    *Hook[hook_resolver.Resolver]
	onGetMangaDetails             *Hook[hook_resolver.Resolver]
	onGetAnimeCollection          *Hook[hook_resolver.Resolver]
	onGetMangaCollection          *Hook[hook_resolver.Resolver]
	onGetCachedAnimeCollection    *Hook[hook_resolver.Resolver]
	onGetCachedMangaCollection    *Hook[hook_resolver.Resolver]
	onGetRawAnimeCollection       *Hook[hook_resolver.Resolver]
	onGetRawMangaCollection       *Hook[hook_resolver.Resolver]
	onGetCachedRawAnimeCollection *Hook[hook_resolver.Resolver]
	onGetCachedRawMangaCollection *Hook[hook_resolver.Resolver]
	onGetStudioDetails            *Hook[hook_resolver.Resolver]
	onPreUpdateEntry              *Hook[hook_resolver.Resolver]
	onPostUpdateEntry             *Hook[hook_resolver.Resolver]
	onPreUpdateEntryProgress      *Hook[hook_resolver.Resolver]
	onPostUpdateEntryProgress     *Hook[hook_resolver.Resolver]
	onPreUpdateEntryRepeat        *Hook[hook_resolver.Resolver]
	onPostUpdateEntryRepeat       *Hook[hook_resolver.Resolver]
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
	onAnimeMetadata                 *Hook[hook_resolver.Resolver]
	onAnimeEpisodeMetadataRequested *Hook[hook_resolver.Resolver]
	onAnimeEpisodeMetadata          *Hook[hook_resolver.Resolver]
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
	onWatchHistoryItemUpdated                   *Hook[hook_resolver.Resolver]
	onWatchHistoryLocalFileEpisodeItemRequested *Hook[hook_resolver.Resolver]
	onWatchHistoryStreamEpisodeItemRequested    *Hook[hook_resolver.Resolver]
	// Discord RPC events
	onDiscordPresenceAnimeActivityRequested *Hook[hook_resolver.Resolver]
	onDiscordPresenceMangaActivityRequested *Hook[hook_resolver.Resolver]
	onDiscordPresenceClientClosed           *Hook[hook_resolver.Resolver]
	// Handlers
	//onHandleGetAnimeCollectionRequested                         *Hook[hook_resolver.Resolver]
	//onHandleGetAnimeCollection                                  *Hook[hook_resolver.Resolver]
	//onHandleGetRawAnimeCollectionRequested                      *Hook[hook_resolver.Resolver]
	//onHandleGetRawAnimeCollection                               *Hook[hook_resolver.Resolver]
	//onHandleEditAnilistListEntryRequested                       *Hook[hook_resolver.Resolver]
	//onHandleGetAnilistAnimeDetailsRequested                     *Hook[hook_resolver.Resolver]
	//onHandleGetAnilistAnimeDetails                              *Hook[hook_resolver.Resolver]
	//onHandleGetAnilistStudioDetailsRequested                    *Hook[hook_resolver.Resolver]
	//onHandleGetAnilistStudioDetails                             *Hook[hook_resolver.Resolver]
	//onHandleDeleteAnilistListEntryRequested                     *Hook[hook_resolver.Resolver]
	//onHandleAnilistListAnimeRequested                           *Hook[hook_resolver.Resolver]
	//onHandleAnilistListAnime                                    *Hook[hook_resolver.Resolver]
	//onHandleAnilistListRecentAiringAnimeRequested               *Hook[hook_resolver.Resolver]
	//onHandleAnilistListRecentAiringAnime                        *Hook[hook_resolver.Resolver]
	//onHandleAnilistListMissedSequelsRequested                   *Hook[hook_resolver.Resolver]
	//onHandleAnilistListMissedSequels                            *Hook[hook_resolver.Resolver]
	//onHandleGetAniListStatsRequested                            *Hook[hook_resolver.Resolver]
	//onHandleGetAniListStats                                     *Hook[hook_resolver.Resolver]
	//onHandleGetLibraryCollectionRequested                       *Hook[hook_resolver.Resolver]
	//onHandleGetLibraryCollection                                *Hook[hook_resolver.Resolver]
	//onHandleAddUnknownMediaRequested                            *Hook[hook_resolver.Resolver]
	//onHandleAddUnknownMedia                                     *Hook[hook_resolver.Resolver]
	//onHandleGetAnimeEntryRequested                              *Hook[hook_resolver.Resolver]
	//onHandleGetAnimeEntry                                       *Hook[hook_resolver.Resolver]
	//onHandleAnimeEntryBulkActionRequested                       *Hook[hook_resolver.Resolver]
	//onHandleAnimeEntryBulkAction                                *Hook[hook_resolver.Resolver]
	//onHandleOpenAnimeEntryInExplorerRequested                   *Hook[hook_resolver.Resolver]
	//onHandleFetchAnimeEntrySuggestionsRequested                 *Hook[hook_resolver.Resolver]
	//onHandleFetchAnimeEntrySuggestions                          *Hook[hook_resolver.Resolver]
	//onHandleAnimeEntryManualMatchRequested                      *Hook[hook_resolver.Resolver]
	//onHandleAnimeEntryManualMatch                               *Hook[hook_resolver.Resolver]
	//onHandleGetMissingEpisodesRequested                         *Hook[hook_resolver.Resolver]
	//onHandleGetMissingEpisodes                                  *Hook[hook_resolver.Resolver]
	//onHandleGetAnimeEntrySilenceStatusRequested                 *Hook[hook_resolver.Resolver]
	//onHandleGetAnimeEntrySilenceStatus                          *Hook[hook_resolver.Resolver]
	//onHandleToggleAnimeEntrySilenceStatusRequested              *Hook[hook_resolver.Resolver]
	//onHandleUpdateAnimeEntryProgressRequested                   *Hook[hook_resolver.Resolver]
	//onHandleUpdateAnimeEntryRepeatRequested                     *Hook[hook_resolver.Resolver]
	//onHandleLoginRequested                                      *Hook[hook_resolver.Resolver]
	//onHandleLogin                                               *Hook[hook_resolver.Resolver]
	//onHandleLogoutRequested                                     *Hook[hook_resolver.Resolver]
	//onHandleLogout                                              *Hook[hook_resolver.Resolver]
	//onHandleRunAutoDownloaderRequested                          *Hook[hook_resolver.Resolver]
	//onHandleGetAutoDownloaderRuleRequested                      *Hook[hook_resolver.Resolver]
	//onHandleGetAutoDownloaderRule                               *Hook[hook_resolver.Resolver]
	//onHandleGetAutoDownloaderRulesByAnimeRequested              *Hook[hook_resolver.Resolver]
	//onHandleGetAutoDownloaderRulesByAnime                       *Hook[hook_resolver.Resolver]
	//onHandleGetAutoDownloaderRulesRequested                     *Hook[hook_resolver.Resolver]
	//onHandleGetAutoDownloaderRules                              *Hook[hook_resolver.Resolver]
	//onHandleCreateAutoDownloaderRuleRequested                   *Hook[hook_resolver.Resolver]
	//onHandleCreateAutoDownloaderRule                            *Hook[hook_resolver.Resolver]
	//onHandleUpdateAutoDownloaderRuleRequested                   *Hook[hook_resolver.Resolver]
	//onHandleUpdateAutoDownloaderRule                            *Hook[hook_resolver.Resolver]
	//onHandleDeleteAutoDownloaderRuleRequested                   *Hook[hook_resolver.Resolver]
	//onHandleGetAutoDownloaderItemsRequested                     *Hook[hook_resolver.Resolver]
	//onHandleGetAutoDownloaderItems                              *Hook[hook_resolver.Resolver]
	//onHandleDeleteAutoDownloaderItemRequested                   *Hook[hook_resolver.Resolver]
	//onHandleUpdateContinuityWatchHistoryItemRequested           *Hook[hook_resolver.Resolver]
	//onHandleGetContinuityWatchHistoryItemRequested              *Hook[hook_resolver.Resolver]
	//onHandleGetContinuityWatchHistoryItem                       *Hook[hook_resolver.Resolver]
	//onHandleGetContinuityWatchHistoryRequested                  *Hook[hook_resolver.Resolver]
	//onHandleGetContinuityWatchHistory                           *Hook[hook_resolver.Resolver]
	//onHandleGetDebridSettingsRequested                          *Hook[hook_resolver.Resolver]
	//onHandleGetDebridSettings                                   *Hook[hook_resolver.Resolver]
	//onHandleSaveDebridSettingsRequested                         *Hook[hook_resolver.Resolver]
	//onHandleSaveDebridSettings                                  *Hook[hook_resolver.Resolver]
	//onHandleDebridAddTorrentsRequested                          *Hook[hook_resolver.Resolver]
	//onHandleDebridDownloadTorrentRequested                      *Hook[hook_resolver.Resolver]
	//onHandleDebridCancelDownloadRequested                       *Hook[hook_resolver.Resolver]
	//onHandleDebridDeleteTorrentRequested                        *Hook[hook_resolver.Resolver]
	//onHandleDebridGetTorrentsRequested                          *Hook[hook_resolver.Resolver]
	//onHandleDebridGetTorrents                                   *Hook[hook_resolver.Resolver]
	//onHandleDebridGetTorrentInfoRequested                       *Hook[hook_resolver.Resolver]
	//onHandleDebridGetTorrentInfo                                *Hook[hook_resolver.Resolver]
	//onHandleDebridGetTorrentFilePreviewsRequested               *Hook[hook_resolver.Resolver]
	//onHandleDebridGetTorrentFilePreviews                        *Hook[hook_resolver.Resolver]
	//onHandleDebridStartStreamRequested                          *Hook[hook_resolver.Resolver]
	//onHandleDebridCancelStreamRequested                         *Hook[hook_resolver.Resolver]
	//onHandleDirectorySelectorRequested                          *Hook[hook_resolver.Resolver]
	//onHandleDirectorySelector                                   *Hook[hook_resolver.Resolver]
	//onHandleSetDiscordMangaActivityRequested                    *Hook[hook_resolver.Resolver]
	//onHandleSetDiscordLegacyAnimeActivityRequested              *Hook[hook_resolver.Resolver]
	//onHandleSetDiscordAnimeActivityWithProgressRequested        *Hook[hook_resolver.Resolver]
	//onHandleUpdateDiscordAnimeActivityWithProgressRequested     *Hook[hook_resolver.Resolver]
	//onHandleCancelDiscordActivityRequested                      *Hook[hook_resolver.Resolver]
	//onHandleGetDocsRequested                                    *Hook[hook_resolver.Resolver]
	//onHandleGetDocs                                             *Hook[hook_resolver.Resolver]
	//onHandleDownloadTorrentFileRequested                        *Hook[hook_resolver.Resolver]
	//onHandleDownloadReleaseRequested                            *Hook[hook_resolver.Resolver]
	//onHandleDownloadRelease                                     *Hook[hook_resolver.Resolver]
	//onHandleOpenInExplorerRequested                             *Hook[hook_resolver.Resolver]
	//onHandleFetchExternalExtensionDataRequested                 *Hook[hook_resolver.Resolver]
	//onHandleFetchExternalExtensionData                          *Hook[hook_resolver.Resolver]
	//onHandleInstallExternalExtensionRequested                   *Hook[hook_resolver.Resolver]
	//onHandleInstallExternalExtension                            *Hook[hook_resolver.Resolver]
	//onHandleUninstallExternalExtensionRequested                 *Hook[hook_resolver.Resolver]
	//onHandleUpdateExtensionCodeRequested                        *Hook[hook_resolver.Resolver]
	//onHandleReloadExternalExtensionsRequested                   *Hook[hook_resolver.Resolver]
	//onHandleReloadExternalExtensionRequested                    *Hook[hook_resolver.Resolver]
	//onHandleListExtensionDataRequested                          *Hook[hook_resolver.Resolver]
	//onHandleListExtensionData                                   *Hook[hook_resolver.Resolver]
	//onHandleGetExtensionPayloadRequested                        *Hook[hook_resolver.Resolver]
	//onHandleGetExtensionPayload                                 *Hook[hook_resolver.Resolver]
	//onHandleListDevelopmentModeExtensionsRequested              *Hook[hook_resolver.Resolver]
	//onHandleListDevelopmentModeExtensions                       *Hook[hook_resolver.Resolver]
	//onHandleGetAllExtensionsRequested                           *Hook[hook_resolver.Resolver]
	//onHandleGetAllExtensions                                    *Hook[hook_resolver.Resolver]
	//onHandleGetExtensionUpdateDataRequested                     *Hook[hook_resolver.Resolver]
	//onHandleGetExtensionUpdateData                              *Hook[hook_resolver.Resolver]
	//onHandleListMangaProviderExtensionsRequested                *Hook[hook_resolver.Resolver]
	//onHandleListMangaProviderExtensions                         *Hook[hook_resolver.Resolver]
	//onHandleListOnlinestreamProviderExtensionsRequested         *Hook[hook_resolver.Resolver]
	//onHandleListOnlinestreamProviderExtensions                  *Hook[hook_resolver.Resolver]
	//onHandleListAnimeTorrentProviderExtensionsRequested         *Hook[hook_resolver.Resolver]
	//onHandleListAnimeTorrentProviderExtensions                  *Hook[hook_resolver.Resolver]
	//onHandleGetPluginSettingsRequested                          *Hook[hook_resolver.Resolver]
	//onHandleGetPluginSettings                                   *Hook[hook_resolver.Resolver]
	//onHandleSetPluginSettingsPinnedTraysRequested               *Hook[hook_resolver.Resolver]
	//onHandleGrantPluginPermissionsRequested                     *Hook[hook_resolver.Resolver]
	//onHandleRunExtensionPlaygroundCodeRequested                 *Hook[hook_resolver.Resolver]
	//onHandleRunExtensionPlaygroundCode                          *Hook[hook_resolver.Resolver]
	//onHandleGetExtensionUserConfigRequested                     *Hook[hook_resolver.Resolver]
	//onHandleGetExtensionUserConfig                              *Hook[hook_resolver.Resolver]
	//onHandleSaveExtensionUserConfigRequested                    *Hook[hook_resolver.Resolver]
	//onHandleGetMarketplaceExtensionsRequested                   *Hook[hook_resolver.Resolver]
	//onHandleGetMarketplaceExtensions                            *Hook[hook_resolver.Resolver]
	//onHandleGetFileCacheTotalSizeRequested                      *Hook[hook_resolver.Resolver]
	//onHandleGetFileCacheTotalSize                               *Hook[hook_resolver.Resolver]
	//onHandleRemoveFileCacheBucketRequested                      *Hook[hook_resolver.Resolver]
	//onHandleGetFileCacheMediastreamVideoFilesTotalSizeRequested *Hook[hook_resolver.Resolver]
	//onHandleGetFileCacheMediastreamVideoFilesTotalSize          *Hook[hook_resolver.Resolver]
	//onHandleClearFileCacheMediastreamVideoFilesRequested        *Hook[hook_resolver.Resolver]
	//onHandleGetLocalFilesRequested                              *Hook[hook_resolver.Resolver]
	//onHandleGetLocalFiles                                       *Hook[hook_resolver.Resolver]
	//onHandleImportLocalFilesRequested                           *Hook[hook_resolver.Resolver]
	//onHandleLocalFileBulkActionRequested                        *Hook[hook_resolver.Resolver]
	//onHandleLocalFileBulkAction                                 *Hook[hook_resolver.Resolver]
	//onHandleUpdateLocalFileDataRequested                        *Hook[hook_resolver.Resolver]
	//onHandleUpdateLocalFileData                                 *Hook[hook_resolver.Resolver]
	//onHandleUpdateLocalFilesRequested                           *Hook[hook_resolver.Resolver]
	//onHandleDeleteLocalFilesRequested                           *Hook[hook_resolver.Resolver]
	//onHandleRemoveEmptyDirectoriesRequested                     *Hook[hook_resolver.Resolver]
	//onHandleMALAuthRequested                                    *Hook[hook_resolver.Resolver]
	//onHandleMALAuth                                             *Hook[hook_resolver.Resolver]
	//onHandleEditMALListEntryProgressRequested                   *Hook[hook_resolver.Resolver]
	//onHandleMALLogoutRequested                                  *Hook[hook_resolver.Resolver]
	//onHandleGetAnilistMangaCollectionRequested                  *Hook[hook_resolver.Resolver]
	//onHandleGetAnilistMangaCollection                           *Hook[hook_resolver.Resolver]
	//onHandleGetRawAnilistMangaCollectionRequested               *Hook[hook_resolver.Resolver]
	//onHandleGetRawAnilistMangaCollection                        *Hook[hook_resolver.Resolver]
	//onHandleGetMangaCollectionRequested                         *Hook[hook_resolver.Resolver]
	//onHandleGetMangaCollection                                  *Hook[hook_resolver.Resolver]
	//onHandleGetMangaEntryRequested                              *Hook[hook_resolver.Resolver]
	//onHandleGetMangaEntry                                       *Hook[hook_resolver.Resolver]
	//onHandleGetMangaEntryDetailsRequested                       *Hook[hook_resolver.Resolver]
	//onHandleGetMangaEntryDetails                                *Hook[hook_resolver.Resolver]
	//onHandleGetMangaLatestChapterNumbersMapRequested            *Hook[hook_resolver.Resolver]
	//onHandleGetMangaLatestChapterNumbersMap                     *Hook[hook_resolver.Resolver]
	//onHandleRefetchMangaChapterContainersRequested              *Hook[hook_resolver.Resolver]
	//onHandleEmptyMangaEntryCacheRequested                       *Hook[hook_resolver.Resolver]
	//onHandleGetMangaEntryChaptersRequested                      *Hook[hook_resolver.Resolver]
	//onHandleGetMangaEntryChapters                               *Hook[hook_resolver.Resolver]
	//onHandleGetMangaEntryPagesRequested                         *Hook[hook_resolver.Resolver]
	//onHandleGetMangaEntryPages                                  *Hook[hook_resolver.Resolver]
	//onHandleGetMangaEntryDownloadedChaptersRequested            *Hook[hook_resolver.Resolver]
	//onHandleGetMangaEntryDownloadedChapters                     *Hook[hook_resolver.Resolver]
	//onHandleAnilistListMangaRequested                           *Hook[hook_resolver.Resolver]
	//onHandleAnilistListManga                                    *Hook[hook_resolver.Resolver]
	//onHandleUpdateMangaProgressRequested                        *Hook[hook_resolver.Resolver]
	//onHandleMangaManualSearchRequested                          *Hook[hook_resolver.Resolver]
	//onHandleMangaManualSearch                                   *Hook[hook_resolver.Resolver]
	//onHandleMangaManualMappingRequested                         *Hook[hook_resolver.Resolver]
	//onHandleGetMangaMappingRequested                            *Hook[hook_resolver.Resolver]
	//onHandleGetMangaMapping                                     *Hook[hook_resolver.Resolver]
	//onHandleRemoveMangaMappingRequested                         *Hook[hook_resolver.Resolver]
	//onHandleDownloadMangaChaptersRequested                      *Hook[hook_resolver.Resolver]
	//onHandleGetMangaDownloadDataRequested                       *Hook[hook_resolver.Resolver]
	//onHandleGetMangaDownloadData                                *Hook[hook_resolver.Resolver]
	//onHandleGetMangaDownloadQueueRequested                      *Hook[hook_resolver.Resolver]
	//onHandleGetMangaDownloadQueue                               *Hook[hook_resolver.Resolver]
	//onHandleStartMangaDownloadQueueRequested                    *Hook[hook_resolver.Resolver]
	//onHandleStopMangaDownloadQueueRequested                     *Hook[hook_resolver.Resolver]
	//onHandleClearAllChapterDownloadQueueRequested               *Hook[hook_resolver.Resolver]
	//onHandleResetErroredChapterDownloadQueueRequested           *Hook[hook_resolver.Resolver]
	//onHandleDeleteMangaDownloadedChaptersRequested              *Hook[hook_resolver.Resolver]
	//onHandleGetMangaDownloadsListRequested                      *Hook[hook_resolver.Resolver]
	//onHandleGetMangaDownloadsList                               *Hook[hook_resolver.Resolver]
	//onHandleTestDumpRequested                                   *Hook[hook_resolver.Resolver]
	//onHandleStartDefaultMediaPlayerRequested                    *Hook[hook_resolver.Resolver]
	//onHandleGetMediastreamSettingsRequested                     *Hook[hook_resolver.Resolver]
	//onHandleGetMediastreamSettings                              *Hook[hook_resolver.Resolver]
	//onHandleSaveMediastreamSettingsRequested                    *Hook[hook_resolver.Resolver]
	//onHandleSaveMediastreamSettings                             *Hook[hook_resolver.Resolver]
	//onHandleRequestMediastreamMediaContainerRequested           *Hook[hook_resolver.Resolver]
	//onHandleRequestMediastreamMediaContainer                    *Hook[hook_resolver.Resolver]
	//onHandlePreloadMediastreamMediaContainerRequested           *Hook[hook_resolver.Resolver]
	//onHandleMediastreamShutdownTranscodeStreamRequested         *Hook[hook_resolver.Resolver]
	//onHandlePopulateTVDBEpisodesRequested                       *Hook[hook_resolver.Resolver]
	//onHandlePopulateTVDBEpisodes                                *Hook[hook_resolver.Resolver]
	//onHandleEmptyTVDBEpisodesRequested                          *Hook[hook_resolver.Resolver]
	//onHandlePopulateFillerDataRequested                         *Hook[hook_resolver.Resolver]
	//onHandleRemoveFillerDataRequested                           *Hook[hook_resolver.Resolver]
	//onHandleGetOnlineStreamEpisodeListRequested                 *Hook[hook_resolver.Resolver]
	//onHandleGetOnlineStreamEpisodeList                          *Hook[hook_resolver.Resolver]
	//onHandleGetOnlineStreamEpisodeSourceRequested               *Hook[hook_resolver.Resolver]
	//onHandleGetOnlineStreamEpisodeSource                        *Hook[hook_resolver.Resolver]
	//onHandleOnlineStreamEmptyCacheRequested                     *Hook[hook_resolver.Resolver]
	//onHandleOnlinestreamManualSearchRequested                   *Hook[hook_resolver.Resolver]
	//onHandleOnlinestreamManualSearch                            *Hook[hook_resolver.Resolver]
	//onHandleOnlinestreamManualMappingRequested                  *Hook[hook_resolver.Resolver]
	//onHandleGetOnlinestreamMappingRequested                     *Hook[hook_resolver.Resolver]
	//onHandleGetOnlinestreamMapping                              *Hook[hook_resolver.Resolver]
	//onHandleRemoveOnlinestreamMappingRequested                  *Hook[hook_resolver.Resolver]
	//onHandlePlaybackPlayVideoRequested                          *Hook[hook_resolver.Resolver]
	//onHandlePlaybackPlayRandomVideoRequested                    *Hook[hook_resolver.Resolver]
	//onHandlePlaybackSyncCurrentProgressRequested                *Hook[hook_resolver.Resolver]
	//onHandlePlaybackSyncCurrentProgress                         *Hook[hook_resolver.Resolver]
	//onHandlePlaybackPlayNextEpisodeRequested                    *Hook[hook_resolver.Resolver]
	//onHandlePlaybackGetNextEpisodeRequested                     *Hook[hook_resolver.Resolver]
	//onHandlePlaybackGetNextEpisode                              *Hook[hook_resolver.Resolver]
	//onHandlePlaybackAutoPlayNextEpisodeRequested                *Hook[hook_resolver.Resolver]
	//onHandlePlaybackStartPlaylistRequested                      *Hook[hook_resolver.Resolver]
	//onHandlePlaybackCancelCurrentPlaylistRequested              *Hook[hook_resolver.Resolver]
	//onHandlePlaybackPlaylistNextRequested                       *Hook[hook_resolver.Resolver]
	//onHandlePlaybackStartManualTrackingRequested                *Hook[hook_resolver.Resolver]
	//onHandlePlaybackCancelManualTrackingRequested               *Hook[hook_resolver.Resolver]
	//onHandleCreatePlaylistRequested                             *Hook[hook_resolver.Resolver]
	//onHandleCreatePlaylist                                      *Hook[hook_resolver.Resolver]
	//onHandleGetPlaylistsRequested                               *Hook[hook_resolver.Resolver]
	//onHandleGetPlaylists                                        *Hook[hook_resolver.Resolver]
	//onHandleUpdatePlaylistRequested                             *Hook[hook_resolver.Resolver]
	//onHandleUpdatePlaylist                                      *Hook[hook_resolver.Resolver]
	//onHandleDeletePlaylistRequested                             *Hook[hook_resolver.Resolver]
	//onHandleGetPlaylistEpisodesRequested                        *Hook[hook_resolver.Resolver]
	//onHandleGetPlaylistEpisodes                                 *Hook[hook_resolver.Resolver]
	//onHandleInstallLatestUpdateRequested                        *Hook[hook_resolver.Resolver]
	//onHandleInstallLatestUpdate                                 *Hook[hook_resolver.Resolver]
	//onHandleGetLatestUpdateRequested                            *Hook[hook_resolver.Resolver]
	//onHandleGetLatestUpdate                                     *Hook[hook_resolver.Resolver]
	//onHandleGetChangelogRequested                               *Hook[hook_resolver.Resolver]
	//onHandleGetChangelog                                        *Hook[hook_resolver.Resolver]
	//onHandleSaveIssueReportRequested                            *Hook[hook_resolver.Resolver]
	//onHandleDownloadIssueReportRequested                        *Hook[hook_resolver.Resolver]
	//onHandleDownloadIssueReport                                 *Hook[hook_resolver.Resolver]
	//onHandleScanLocalFilesRequested                             *Hook[hook_resolver.Resolver]
	//onHandleScanLocalFiles                                      *Hook[hook_resolver.Resolver]
	//onHandleGetScanSummariesRequested                           *Hook[hook_resolver.Resolver]
	//onHandleGetScanSummaries                                    *Hook[hook_resolver.Resolver]
	//onHandleGetSettingsRequested                                *Hook[hook_resolver.Resolver]
	//onHandleGetSettings                                         *Hook[hook_resolver.Resolver]
	//onHandleGettingStartedRequested                             *Hook[hook_resolver.Resolver]
	//onHandleGettingStarted                                      *Hook[hook_resolver.Resolver]
	//onHandleSaveSettingsRequested                               *Hook[hook_resolver.Resolver]
	//onHandleSaveSettings                                        *Hook[hook_resolver.Resolver]
	//onHandleSaveAutoDownloaderSettingsRequested                 *Hook[hook_resolver.Resolver]
	//onHandleGetStatusRequested                                  *Hook[hook_resolver.Resolver]
	//onHandleGetStatus                                           *Hook[hook_resolver.Resolver]
	//onHandleGetLogFilenamesRequested                            *Hook[hook_resolver.Resolver]
	//onHandleGetLogFilenames                                     *Hook[hook_resolver.Resolver]
	//onHandleDeleteLogsRequested                                 *Hook[hook_resolver.Resolver]
	//onHandleGetLatestLogContentRequested                        *Hook[hook_resolver.Resolver]
	//onHandleGetLatestLogContent                                 *Hook[hook_resolver.Resolver]
	//onHandleSyncGetTrackedMediaItemsRequested                   *Hook[hook_resolver.Resolver]
	//onHandleSyncGetTrackedMediaItems                            *Hook[hook_resolver.Resolver]
	//onHandleSyncAddMediaRequested                               *Hook[hook_resolver.Resolver]
	//onHandleSyncRemoveMediaRequested                            *Hook[hook_resolver.Resolver]
	//onHandleSyncGetIsMediaTrackedRequested                      *Hook[hook_resolver.Resolver]
	//onHandleSyncLocalDataRequested                              *Hook[hook_resolver.Resolver]
	//onHandleSyncGetQueueStateRequested                          *Hook[hook_resolver.Resolver]
	//onHandleSyncGetQueueState                                   *Hook[hook_resolver.Resolver]
	//onHandleSyncAnilistDataRequested                            *Hook[hook_resolver.Resolver]
	//onHandleSyncSetHasLocalChangesRequested                     *Hook[hook_resolver.Resolver]
	//onHandleSyncGetHasLocalChangesRequested                     *Hook[hook_resolver.Resolver]
	//onHandleSyncGetLocalStorageSizeRequested                    *Hook[hook_resolver.Resolver]
	//onHandleSyncGetLocalStorageSize                             *Hook[hook_resolver.Resolver]
	//onHandleGetThemeRequested                                   *Hook[hook_resolver.Resolver]
	//onHandleGetTheme                                            *Hook[hook_resolver.Resolver]
	//onHandleUpdateThemeRequested                                *Hook[hook_resolver.Resolver]
	//onHandleUpdateTheme                                         *Hook[hook_resolver.Resolver]
	//onHandleGetActiveTorrentListRequested                       *Hook[hook_resolver.Resolver]
	//onHandleGetActiveTorrentList                                *Hook[hook_resolver.Resolver]
	//onHandleTorrentClientActionRequested                        *Hook[hook_resolver.Resolver]
	//onHandleTorrentClientDownloadRequested                      *Hook[hook_resolver.Resolver]
	//onHandleTorrentClientAddMagnetFromRuleRequested             *Hook[hook_resolver.Resolver]
	//onHandleSearchTorrentRequested                              *Hook[hook_resolver.Resolver]
	//onHandleSearchTorrent                                       *Hook[hook_resolver.Resolver]
	//onHandleGetTorrentstreamEpisodeCollectionRequested          *Hook[hook_resolver.Resolver]
	//onHandleGetTorrentstreamEpisodeCollection                   *Hook[hook_resolver.Resolver]
	//onHandleGetTorrentstreamSettingsRequested                   *Hook[hook_resolver.Resolver]
	//onHandleGetTorrentstreamSettings                            *Hook[hook_resolver.Resolver]
	//onHandleSaveTorrentstreamSettingsRequested                  *Hook[hook_resolver.Resolver]
	//onHandleSaveTorrentstreamSettings                           *Hook[hook_resolver.Resolver]
	//onHandleGetTorrentstreamTorrentFilePreviewsRequested        *Hook[hook_resolver.Resolver]
	//onHandleGetTorrentstreamTorrentFilePreviews                 *Hook[hook_resolver.Resolver]
	//onHandleTorrentstreamStartStreamRequested                   *Hook[hook_resolver.Resolver]
	//onHandleTorrentstreamStopStreamRequested                    *Hook[hook_resolver.Resolver]
	//onHandleTorrentstreamDropTorrentRequested                   *Hook[hook_resolver.Resolver]
	//onHandleGetTorrentstreamBatchHistoryRequested               *Hook[hook_resolver.Resolver]
	//onHandleGetTorrentstreamBatchHistory                        *Hook[hook_resolver.Resolver]
	//onHandleTorrentstreamServeStreamRequested                   *Hook[hook_resolver.Resolver]
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
	m.onGetCachedAnimeCollection = &Hook[hook_resolver.Resolver]{}
	m.onGetCachedMangaCollection = &Hook[hook_resolver.Resolver]{}
	m.onGetRawAnimeCollection = &Hook[hook_resolver.Resolver]{}
	m.onGetRawMangaCollection = &Hook[hook_resolver.Resolver]{}
	m.onGetCachedRawAnimeCollection = &Hook[hook_resolver.Resolver]{}
	m.onGetCachedRawMangaCollection = &Hook[hook_resolver.Resolver]{}
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
	m.onAnimeMetadata = &Hook[hook_resolver.Resolver]{}
	m.onAnimeEpisodeMetadataRequested = &Hook[hook_resolver.Resolver]{}
	m.onAnimeEpisodeMetadata = &Hook[hook_resolver.Resolver]{}
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
	m.onWatchHistoryItemUpdated = &Hook[hook_resolver.Resolver]{}
	m.onWatchHistoryLocalFileEpisodeItemRequested = &Hook[hook_resolver.Resolver]{}
	m.onWatchHistoryStreamEpisodeItemRequested = &Hook[hook_resolver.Resolver]{}
	// Discord RPC events
	m.onDiscordPresenceAnimeActivityRequested = &Hook[hook_resolver.Resolver]{}
	m.onDiscordPresenceMangaActivityRequested = &Hook[hook_resolver.Resolver]{}
	m.onDiscordPresenceClientClosed = &Hook[hook_resolver.Resolver]{}
	// Handlers
	//m.onHandleGetAnimeCollectionRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetAnimeCollection = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetRawAnimeCollectionRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetRawAnimeCollection = &Hook[hook_resolver.Resolver]{}
	//m.onHandleEditAnilistListEntryRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetAnilistAnimeDetailsRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetAnilistAnimeDetails = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetAnilistStudioDetailsRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetAnilistStudioDetails = &Hook[hook_resolver.Resolver]{}
	//m.onHandleDeleteAnilistListEntryRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleAnilistListAnimeRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleAnilistListAnime = &Hook[hook_resolver.Resolver]{}
	//m.onHandleAnilistListRecentAiringAnimeRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleAnilistListRecentAiringAnime = &Hook[hook_resolver.Resolver]{}
	//m.onHandleAnilistListMissedSequelsRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleAnilistListMissedSequels = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetAniListStatsRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetAniListStats = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetLibraryCollectionRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetLibraryCollection = &Hook[hook_resolver.Resolver]{}
	//m.onHandleAddUnknownMediaRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleAddUnknownMedia = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetAnimeEntryRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetAnimeEntry = &Hook[hook_resolver.Resolver]{}
	//m.onHandleAnimeEntryBulkActionRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleAnimeEntryBulkAction = &Hook[hook_resolver.Resolver]{}
	//m.onHandleOpenAnimeEntryInExplorerRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleFetchAnimeEntrySuggestionsRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleFetchAnimeEntrySuggestions = &Hook[hook_resolver.Resolver]{}
	//m.onHandleAnimeEntryManualMatchRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleAnimeEntryManualMatch = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetMissingEpisodesRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetMissingEpisodes = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetAnimeEntrySilenceStatusRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetAnimeEntrySilenceStatus = &Hook[hook_resolver.Resolver]{}
	//m.onHandleToggleAnimeEntrySilenceStatusRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleUpdateAnimeEntryProgressRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleUpdateAnimeEntryRepeatRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleLoginRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleLogin = &Hook[hook_resolver.Resolver]{}
	//m.onHandleLogoutRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleLogout = &Hook[hook_resolver.Resolver]{}
	//m.onHandleRunAutoDownloaderRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetAutoDownloaderRuleRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetAutoDownloaderRule = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetAutoDownloaderRulesByAnimeRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetAutoDownloaderRulesByAnime = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetAutoDownloaderRulesRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetAutoDownloaderRules = &Hook[hook_resolver.Resolver]{}
	//m.onHandleCreateAutoDownloaderRuleRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleCreateAutoDownloaderRule = &Hook[hook_resolver.Resolver]{}
	//m.onHandleUpdateAutoDownloaderRuleRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleUpdateAutoDownloaderRule = &Hook[hook_resolver.Resolver]{}
	//m.onHandleDeleteAutoDownloaderRuleRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetAutoDownloaderItemsRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetAutoDownloaderItems = &Hook[hook_resolver.Resolver]{}
	//m.onHandleDeleteAutoDownloaderItemRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleUpdateContinuityWatchHistoryItemRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetContinuityWatchHistoryItemRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetContinuityWatchHistoryItem = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetContinuityWatchHistoryRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetContinuityWatchHistory = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetDebridSettingsRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetDebridSettings = &Hook[hook_resolver.Resolver]{}
	//m.onHandleSaveDebridSettingsRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleSaveDebridSettings = &Hook[hook_resolver.Resolver]{}
	//m.onHandleDebridAddTorrentsRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleDebridDownloadTorrentRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleDebridCancelDownloadRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleDebridDeleteTorrentRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleDebridGetTorrentsRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleDebridGetTorrents = &Hook[hook_resolver.Resolver]{}
	//m.onHandleDebridGetTorrentInfoRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleDebridGetTorrentInfo = &Hook[hook_resolver.Resolver]{}
	//m.onHandleDebridGetTorrentFilePreviewsRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleDebridGetTorrentFilePreviews = &Hook[hook_resolver.Resolver]{}
	//m.onHandleDebridStartStreamRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleDebridCancelStreamRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleDirectorySelectorRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleDirectorySelector = &Hook[hook_resolver.Resolver]{}
	//m.onHandleSetDiscordMangaActivityRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleSetDiscordLegacyAnimeActivityRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleSetDiscordAnimeActivityWithProgressRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleUpdateDiscordAnimeActivityWithProgressRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleCancelDiscordActivityRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetDocsRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetDocs = &Hook[hook_resolver.Resolver]{}
	//m.onHandleDownloadTorrentFileRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleDownloadReleaseRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleDownloadRelease = &Hook[hook_resolver.Resolver]{}
	//m.onHandleOpenInExplorerRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleFetchExternalExtensionDataRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleFetchExternalExtensionData = &Hook[hook_resolver.Resolver]{}
	//m.onHandleInstallExternalExtensionRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleInstallExternalExtension = &Hook[hook_resolver.Resolver]{}
	//m.onHandleUninstallExternalExtensionRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleUpdateExtensionCodeRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleReloadExternalExtensionsRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleReloadExternalExtensionRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleListExtensionDataRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleListExtensionData = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetExtensionPayloadRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetExtensionPayload = &Hook[hook_resolver.Resolver]{}
	//m.onHandleListDevelopmentModeExtensionsRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleListDevelopmentModeExtensions = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetAllExtensionsRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetAllExtensions = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetExtensionUpdateDataRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetExtensionUpdateData = &Hook[hook_resolver.Resolver]{}
	//m.onHandleListMangaProviderExtensionsRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleListMangaProviderExtensions = &Hook[hook_resolver.Resolver]{}
	//m.onHandleListOnlinestreamProviderExtensionsRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleListOnlinestreamProviderExtensions = &Hook[hook_resolver.Resolver]{}
	//m.onHandleListAnimeTorrentProviderExtensionsRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleListAnimeTorrentProviderExtensions = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetPluginSettingsRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetPluginSettings = &Hook[hook_resolver.Resolver]{}
	//m.onHandleSetPluginSettingsPinnedTraysRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGrantPluginPermissionsRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleRunExtensionPlaygroundCodeRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleRunExtensionPlaygroundCode = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetExtensionUserConfigRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetExtensionUserConfig = &Hook[hook_resolver.Resolver]{}
	//m.onHandleSaveExtensionUserConfigRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetMarketplaceExtensionsRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetMarketplaceExtensions = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetFileCacheTotalSizeRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetFileCacheTotalSize = &Hook[hook_resolver.Resolver]{}
	//m.onHandleRemoveFileCacheBucketRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetFileCacheMediastreamVideoFilesTotalSizeRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetFileCacheMediastreamVideoFilesTotalSize = &Hook[hook_resolver.Resolver]{}
	//m.onHandleClearFileCacheMediastreamVideoFilesRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetLocalFilesRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetLocalFiles = &Hook[hook_resolver.Resolver]{}
	//m.onHandleImportLocalFilesRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleLocalFileBulkActionRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleLocalFileBulkAction = &Hook[hook_resolver.Resolver]{}
	//m.onHandleUpdateLocalFileDataRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleUpdateLocalFileData = &Hook[hook_resolver.Resolver]{}
	//m.onHandleUpdateLocalFilesRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleDeleteLocalFilesRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleRemoveEmptyDirectoriesRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleMALAuthRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleMALAuth = &Hook[hook_resolver.Resolver]{}
	//m.onHandleEditMALListEntryProgressRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleMALLogoutRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetAnilistMangaCollectionRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetAnilistMangaCollection = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetRawAnilistMangaCollectionRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetRawAnilistMangaCollection = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetMangaCollectionRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetMangaCollection = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetMangaEntryRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetMangaEntry = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetMangaEntryDetailsRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetMangaEntryDetails = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetMangaLatestChapterNumbersMapRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetMangaLatestChapterNumbersMap = &Hook[hook_resolver.Resolver]{}
	//m.onHandleRefetchMangaChapterContainersRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleEmptyMangaEntryCacheRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetMangaEntryChaptersRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetMangaEntryChapters = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetMangaEntryPagesRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetMangaEntryPages = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetMangaEntryDownloadedChaptersRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetMangaEntryDownloadedChapters = &Hook[hook_resolver.Resolver]{}
	//m.onHandleAnilistListMangaRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleAnilistListManga = &Hook[hook_resolver.Resolver]{}
	//m.onHandleUpdateMangaProgressRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleMangaManualSearchRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleMangaManualSearch = &Hook[hook_resolver.Resolver]{}
	//m.onHandleMangaManualMappingRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetMangaMappingRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetMangaMapping = &Hook[hook_resolver.Resolver]{}
	//m.onHandleRemoveMangaMappingRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleDownloadMangaChaptersRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetMangaDownloadDataRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetMangaDownloadData = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetMangaDownloadQueueRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetMangaDownloadQueue = &Hook[hook_resolver.Resolver]{}
	//m.onHandleStartMangaDownloadQueueRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleStopMangaDownloadQueueRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleClearAllChapterDownloadQueueRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleResetErroredChapterDownloadQueueRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleDeleteMangaDownloadedChaptersRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetMangaDownloadsListRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetMangaDownloadsList = &Hook[hook_resolver.Resolver]{}
	//m.onHandleTestDumpRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleStartDefaultMediaPlayerRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetMediastreamSettingsRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetMediastreamSettings = &Hook[hook_resolver.Resolver]{}
	//m.onHandleSaveMediastreamSettingsRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleSaveMediastreamSettings = &Hook[hook_resolver.Resolver]{}
	//m.onHandleRequestMediastreamMediaContainerRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleRequestMediastreamMediaContainer = &Hook[hook_resolver.Resolver]{}
	//m.onHandlePreloadMediastreamMediaContainerRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleMediastreamShutdownTranscodeStreamRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandlePopulateTVDBEpisodesRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandlePopulateTVDBEpisodes = &Hook[hook_resolver.Resolver]{}
	//m.onHandleEmptyTVDBEpisodesRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandlePopulateFillerDataRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleRemoveFillerDataRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetOnlineStreamEpisodeListRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetOnlineStreamEpisodeList = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetOnlineStreamEpisodeSourceRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetOnlineStreamEpisodeSource = &Hook[hook_resolver.Resolver]{}
	//m.onHandleOnlineStreamEmptyCacheRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleOnlinestreamManualSearchRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleOnlinestreamManualSearch = &Hook[hook_resolver.Resolver]{}
	//m.onHandleOnlinestreamManualMappingRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetOnlinestreamMappingRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetOnlinestreamMapping = &Hook[hook_resolver.Resolver]{}
	//m.onHandleRemoveOnlinestreamMappingRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandlePlaybackPlayVideoRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandlePlaybackPlayRandomVideoRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandlePlaybackSyncCurrentProgressRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandlePlaybackSyncCurrentProgress = &Hook[hook_resolver.Resolver]{}
	//m.onHandlePlaybackPlayNextEpisodeRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandlePlaybackGetNextEpisodeRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandlePlaybackGetNextEpisode = &Hook[hook_resolver.Resolver]{}
	//m.onHandlePlaybackAutoPlayNextEpisodeRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandlePlaybackStartPlaylistRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandlePlaybackCancelCurrentPlaylistRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandlePlaybackPlaylistNextRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandlePlaybackStartManualTrackingRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandlePlaybackCancelManualTrackingRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleCreatePlaylistRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleCreatePlaylist = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetPlaylistsRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetPlaylists = &Hook[hook_resolver.Resolver]{}
	//m.onHandleUpdatePlaylistRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleUpdatePlaylist = &Hook[hook_resolver.Resolver]{}
	//m.onHandleDeletePlaylistRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetPlaylistEpisodesRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetPlaylistEpisodes = &Hook[hook_resolver.Resolver]{}
	//m.onHandleInstallLatestUpdateRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleInstallLatestUpdate = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetLatestUpdateRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetLatestUpdate = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetChangelogRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetChangelog = &Hook[hook_resolver.Resolver]{}
	//m.onHandleSaveIssueReportRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleDownloadIssueReportRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleDownloadIssueReport = &Hook[hook_resolver.Resolver]{}
	//m.onHandleScanLocalFilesRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleScanLocalFiles = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetScanSummariesRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetScanSummaries = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetSettingsRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetSettings = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGettingStartedRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGettingStarted = &Hook[hook_resolver.Resolver]{}
	//m.onHandleSaveSettingsRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleSaveSettings = &Hook[hook_resolver.Resolver]{}
	//m.onHandleSaveAutoDownloaderSettingsRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetStatusRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetStatus = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetLogFilenamesRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetLogFilenames = &Hook[hook_resolver.Resolver]{}
	//m.onHandleDeleteLogsRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetLatestLogContentRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetLatestLogContent = &Hook[hook_resolver.Resolver]{}
	//m.onHandleSyncGetTrackedMediaItemsRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleSyncGetTrackedMediaItems = &Hook[hook_resolver.Resolver]{}
	//m.onHandleSyncAddMediaRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleSyncRemoveMediaRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleSyncGetIsMediaTrackedRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleSyncLocalDataRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleSyncGetQueueStateRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleSyncGetQueueState = &Hook[hook_resolver.Resolver]{}
	//m.onHandleSyncAnilistDataRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleSyncSetHasLocalChangesRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleSyncGetHasLocalChangesRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleSyncGetLocalStorageSizeRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleSyncGetLocalStorageSize = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetThemeRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetTheme = &Hook[hook_resolver.Resolver]{}
	//m.onHandleUpdateThemeRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleUpdateTheme = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetActiveTorrentListRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetActiveTorrentList = &Hook[hook_resolver.Resolver]{}
	//m.onHandleTorrentClientActionRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleTorrentClientDownloadRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleTorrentClientAddMagnetFromRuleRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleSearchTorrentRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleSearchTorrent = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetTorrentstreamEpisodeCollectionRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetTorrentstreamEpisodeCollection = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetTorrentstreamSettingsRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetTorrentstreamSettings = &Hook[hook_resolver.Resolver]{}
	//m.onHandleSaveTorrentstreamSettingsRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleSaveTorrentstreamSettings = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetTorrentstreamTorrentFilePreviewsRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetTorrentstreamTorrentFilePreviews = &Hook[hook_resolver.Resolver]{}
	//m.onHandleTorrentstreamStartStreamRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleTorrentstreamStopStreamRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleTorrentstreamDropTorrentRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetTorrentstreamBatchHistoryRequested = &Hook[hook_resolver.Resolver]{}
	//m.onHandleGetTorrentstreamBatchHistory = &Hook[hook_resolver.Resolver]{}
	//m.onHandleTorrentstreamServeStreamRequested = &Hook[hook_resolver.Resolver]{}
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

func (m *ManagerImpl) OnGetCachedAnimeCollection() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onGetCachedAnimeCollection
}

func (m *ManagerImpl) OnGetCachedMangaCollection() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onGetCachedMangaCollection
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

func (m *ManagerImpl) OnGetCachedRawAnimeCollection() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onGetCachedRawAnimeCollection
}

func (m *ManagerImpl) OnGetCachedRawMangaCollection() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onGetCachedRawMangaCollection
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

func (m *ManagerImpl) OnAnimeMetadata() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onAnimeMetadata
}

func (m *ManagerImpl) OnAnimeEpisodeMetadataRequested() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onAnimeEpisodeMetadataRequested
}

func (m *ManagerImpl) OnAnimeEpisodeMetadata() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onAnimeEpisodeMetadata
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

func (m *ManagerImpl) OnWatchHistoryItemUpdated() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onWatchHistoryItemUpdated
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

// Discord RPC events

func (m *ManagerImpl) OnDiscordPresenceAnimeActivityRequested() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onDiscordPresenceAnimeActivityRequested
}

func (m *ManagerImpl) OnDiscordPresenceMangaActivityRequested() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onDiscordPresenceMangaActivityRequested
}

func (m *ManagerImpl) OnDiscordPresenceClientClosed() *Hook[hook_resolver.Resolver] {
	if m == nil {
		return &Hook[hook_resolver.Resolver]{}
	}
	return m.onDiscordPresenceClientClosed
}
