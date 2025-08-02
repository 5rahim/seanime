declare namespace $app {

    /**
     * @package anilist
     */

    /**
     * @event ListMissedSequelsRequestedEvent
     * @file internal/api/anilist/hook_events.go
     * @description
     * ListMissedSequelsRequestedEvent is triggered when the list missed sequels request is requested.
     * Prevent default to skip the default behavior and return your own data.
     */
    function onListMissedSequelsRequested(cb: (event: ListMissedSequelsRequestedEvent) => void): void;

    interface ListMissedSequelsRequestedEvent {
        next(): void;

        preventDefault(): void;

        animeCollectionWithRelations?: AL_AnimeCollectionWithRelations;
        variables?: Record<string, any>;
        query: string;
        list?: Array<AL_BaseAnime>;
    }

    /**
     * @event ListMissedSequelsEvent
     * @file internal/api/anilist/hook_events.go
     */
    function onListMissedSequels(cb: (event: ListMissedSequelsEvent) => void): void;

    interface ListMissedSequelsEvent {
        next(): void;

        list?: Array<AL_BaseAnime>;
    }


    /**
     * @package anilist_platform
     */

    /**
     * @event GetAnimeEvent
     * @file internal/platforms/anilist_platform/hook_events.go
     */
    function onGetAnime(cb: (event: GetAnimeEvent) => void): void;

    interface GetAnimeEvent {
        next(): void;

        anime?: AL_BaseAnime;
    }

    /**
     * @event GetAnimeDetailsEvent
     * @file internal/platforms/anilist_platform/hook_events.go
     */
    function onGetAnimeDetails(cb: (event: GetAnimeDetailsEvent) => void): void;

    interface GetAnimeDetailsEvent {
        next(): void;

        anime?: AL_AnimeDetailsById_Media;
    }

    /**
     * @event GetMangaEvent
     * @file internal/platforms/anilist_platform/hook_events.go
     */
    function onGetManga(cb: (event: GetMangaEvent) => void): void;

    interface GetMangaEvent {
        next(): void;

        manga?: AL_BaseManga;
    }

    /**
     * @event GetMangaDetailsEvent
     * @file internal/platforms/anilist_platform/hook_events.go
     */
    function onGetMangaDetails(cb: (event: GetMangaDetailsEvent) => void): void;

    interface GetMangaDetailsEvent {
        next(): void;

        manga?: AL_MangaDetailsById_Media;
    }

    /**
     * @event GetCachedAnimeCollectionEvent
     * @file internal/platforms/anilist_platform/hook_events.go
     */
    function onGetCachedAnimeCollection(cb: (event: GetCachedAnimeCollectionEvent) => void): void;

    interface GetCachedAnimeCollectionEvent {
        next(): void;

        animeCollection?: AL_AnimeCollection;
    }

    /**
     * @event GetCachedMangaCollectionEvent
     * @file internal/platforms/anilist_platform/hook_events.go
     */
    function onGetCachedMangaCollection(cb: (event: GetCachedMangaCollectionEvent) => void): void;

    interface GetCachedMangaCollectionEvent {
        next(): void;

        mangaCollection?: AL_MangaCollection;
    }

    /**
     * @event GetAnimeCollectionEvent
     * @file internal/platforms/anilist_platform/hook_events.go
     */
    function onGetAnimeCollection(cb: (event: GetAnimeCollectionEvent) => void): void;

    interface GetAnimeCollectionEvent {
        next(): void;

        animeCollection?: AL_AnimeCollection;
    }

    /**
     * @event GetMangaCollectionEvent
     * @file internal/platforms/anilist_platform/hook_events.go
     */
    function onGetMangaCollection(cb: (event: GetMangaCollectionEvent) => void): void;

    interface GetMangaCollectionEvent {
        next(): void;

        mangaCollection?: AL_MangaCollection;
    }

    /**
     * @event GetCachedRawAnimeCollectionEvent
     * @file internal/platforms/anilist_platform/hook_events.go
     */
    function onGetCachedRawAnimeCollection(cb: (event: GetCachedRawAnimeCollectionEvent) => void): void;

    interface GetCachedRawAnimeCollectionEvent {
        next(): void;

        animeCollection?: AL_AnimeCollection;
    }

    /**
     * @event GetCachedRawMangaCollectionEvent
     * @file internal/platforms/anilist_platform/hook_events.go
     */
    function onGetCachedRawMangaCollection(cb: (event: GetCachedRawMangaCollectionEvent) => void): void;

    interface GetCachedRawMangaCollectionEvent {
        next(): void;

        mangaCollection?: AL_MangaCollection;
    }

    /**
     * @event GetRawAnimeCollectionEvent
     * @file internal/platforms/anilist_platform/hook_events.go
     */
    function onGetRawAnimeCollection(cb: (event: GetRawAnimeCollectionEvent) => void): void;

    interface GetRawAnimeCollectionEvent {
        next(): void;

        animeCollection?: AL_AnimeCollection;
    }

    /**
     * @event GetRawMangaCollectionEvent
     * @file internal/platforms/anilist_platform/hook_events.go
     */
    function onGetRawMangaCollection(cb: (event: GetRawMangaCollectionEvent) => void): void;

    interface GetRawMangaCollectionEvent {
        next(): void;

        mangaCollection?: AL_MangaCollection;
    }

    /**
     * @event GetStudioDetailsEvent
     * @file internal/platforms/anilist_platform/hook_events.go
     */
    function onGetStudioDetails(cb: (event: GetStudioDetailsEvent) => void): void;

    interface GetStudioDetailsEvent {
        next(): void;

        studio?: AL_StudioDetails;
    }

    /**
     * @event PreUpdateEntryEvent
     * @file internal/platforms/anilist_platform/hook_events.go
     * @description
     * PreUpdateEntryEvent is triggered when an entry is about to be updated.
     * Prevent default to skip the default update and override the update.
     */
    function onPreUpdateEntry(cb: (event: PreUpdateEntryEvent) => void): void;

    interface PreUpdateEntryEvent {
        next(): void;

        preventDefault(): void;

        mediaId?: number;
        status?: AL_MediaListStatus;
        scoreRaw?: number;
        progress?: number;
        startedAt?: AL_FuzzyDateInput;
        completedAt?: AL_FuzzyDateInput;
    }

    /**
     * @event PostUpdateEntryEvent
     * @file internal/platforms/anilist_platform/hook_events.go
     */
    function onPostUpdateEntry(cb: (event: PostUpdateEntryEvent) => void): void;

    interface PostUpdateEntryEvent {
        next(): void;

        mediaId?: number;
    }

    /**
     * @event PreUpdateEntryProgressEvent
     * @file internal/platforms/anilist_platform/hook_events.go
     * @description
     * PreUpdateEntryProgressEvent is triggered when an entry's progress is about to be updated.
     * Prevent default to skip the default update and override the update.
     */
    function onPreUpdateEntryProgress(cb: (event: PreUpdateEntryProgressEvent) => void): void;

    interface PreUpdateEntryProgressEvent {
        next(): void;

        preventDefault(): void;

        mediaId?: number;
        progress?: number;
        totalCount?: number;
        status?: AL_MediaListStatus;
    }

    /**
     * @event PostUpdateEntryProgressEvent
     * @file internal/platforms/anilist_platform/hook_events.go
     */
    function onPostUpdateEntryProgress(cb: (event: PostUpdateEntryProgressEvent) => void): void;

    interface PostUpdateEntryProgressEvent {
        next(): void;

        mediaId?: number;
    }

    /**
     * @event PreUpdateEntryRepeatEvent
     * @file internal/platforms/anilist_platform/hook_events.go
     * @description
     * PreUpdateEntryRepeatEvent is triggered when an entry's repeat is about to be updated.
     * Prevent default to skip the default update and override the update.
     */
    function onPreUpdateEntryRepeat(cb: (event: PreUpdateEntryRepeatEvent) => void): void;

    interface PreUpdateEntryRepeatEvent {
        next(): void;

        preventDefault(): void;

        mediaId?: number;
        repeat?: number;
    }

    /**
     * @event PostUpdateEntryRepeatEvent
     * @file internal/platforms/anilist_platform/hook_events.go
     */
    function onPostUpdateEntryRepeat(cb: (event: PostUpdateEntryRepeatEvent) => void): void;

    interface PostUpdateEntryRepeatEvent {
        next(): void;

        mediaId?: number;
    }


    /**
     * @package animap
     */

    /**
     * @event AnimapMediaRequestedEvent
     * @file internal/api/animap/hook_events.go
     * @description
     * AnimapMediaRequestedEvent is triggered when the Animap media is requested.
     * Prevent default to skip the default behavior and return your own data.
     */
    function onAnimapMediaRequested(cb: (event: AnimapMediaRequestedEvent) => void): void;

    interface AnimapMediaRequestedEvent {
        from: string;
        id: number;
        media?: Animap_Anime;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event AnimapMediaEvent
     * @file internal/api/animap/hook_events.go
     * @description
     * AnimapMediaEvent is triggered after processing AnimapMedia.
     */
    function onAnimapMedia(cb: (event: AnimapMediaEvent) => void): void;

    interface AnimapMediaEvent {
        media?: Animap_Anime;

        next(): void;
    }


    /**
     * @package anime
     */

    /**
     * @event AnimeEntryRequestedEvent
     * @file internal/library/anime/hook_events.go
     * @description
     * AnimeEntryRequestedEvent is triggered when an anime entry is requested.
     * Prevent default to skip the default behavior and return the modified entry.
     * This event is triggered before [AnimeEntryEvent].
     * If the modified entry is nil, an error will be returned.
     */
    function onAnimeEntryRequested(cb: (event: AnimeEntryRequestedEvent) => void): void;

    interface AnimeEntryRequestedEvent {
        next(): void;

        preventDefault(): void;

        mediaId: number;
        localFiles?: Array<Anime_LocalFile>;
        animeCollection?: AL_AnimeCollection;
        entry?: Anime_Entry;
    }

    /**
     * @event AnimeEntryEvent
     * @file internal/library/anime/hook_events.go
     * @description
     * AnimeEntryEvent is triggered when the media entry is being returned.
     * This event is triggered after [AnimeEntryRequestedEvent].
     */
    function onAnimeEntry(cb: (event: AnimeEntryEvent) => void): void;

    interface AnimeEntryEvent {
        next(): void;

        entry?: Anime_Entry;
    }

    /**
     * @event AnimeEntryFillerHydrationEvent
     * @file internal/library/anime/hook_events.go
     * @description
     * AnimeEntryFillerHydrationEvent is triggered when the filler data is being added to the media entry.
     * This event is triggered after [AnimeEntryEvent].
     * Prevent default to skip the filler data.
     */
    function onAnimeEntryFillerHydration(cb: (event: AnimeEntryFillerHydrationEvent) => void): void;

    interface AnimeEntryFillerHydrationEvent {
        next(): void;

        preventDefault(): void;

        entry?: Anime_Entry;
    }

    /**
     * @event AnimeEntryLibraryDataRequestedEvent
     * @file internal/library/anime/hook_events.go
     * @description
     * AnimeEntryLibraryDataRequestedEvent is triggered when the app requests the library data for a media entry.
     * This is triggered before [AnimeEntryLibraryDataEvent].
     */
    function onAnimeEntryLibraryDataRequested(cb: (event: AnimeEntryLibraryDataRequestedEvent) => void): void;

    interface AnimeEntryLibraryDataRequestedEvent {
        next(): void;

        entryLocalFiles?: Array<Anime_LocalFile>;
        mediaId: number;
        currentProgress: number;
    }

    /**
     * @event AnimeEntryLibraryDataEvent
     * @file internal/library/anime/hook_events.go
     * @description
     * AnimeEntryLibraryDataEvent is triggered when the library data is being added to the media entry.
     * This is triggered after [AnimeEntryLibraryDataRequestedEvent].
     */
    function onAnimeEntryLibraryData(cb: (event: AnimeEntryLibraryDataEvent) => void): void;

    interface AnimeEntryLibraryDataEvent {
        next(): void;

        entryLibraryData?: Anime_EntryLibraryData;
    }

    /**
     * @event AnimeEntryManualMatchBeforeSaveEvent
     * @file internal/library/anime/hook_events.go
     * @description
     * AnimeEntryManualMatchBeforeSaveEvent is triggered when the user manually matches local files to a media entry.
     * Prevent default to skip saving the local files.
     */
    function onAnimeEntryManualMatchBeforeSave(cb: (event: AnimeEntryManualMatchBeforeSaveEvent) => void): void;

    interface AnimeEntryManualMatchBeforeSaveEvent {
        next(): void;

        preventDefault(): void;

        mediaId: number;
        paths?: Array<string>;
        matchedLocalFiles?: Array<Anime_LocalFile>;
    }

    /**
     * @event MissingEpisodesRequestedEvent
     * @file internal/library/anime/hook_events.go
     * @description
     * MissingEpisodesRequestedEvent is triggered when the user requests the missing episodes for the entire library.
     * Prevent default to skip the default process and return the modified missing episodes.
     */
    function onMissingEpisodesRequested(cb: (event: MissingEpisodesRequestedEvent) => void): void;

    interface MissingEpisodesRequestedEvent {
        next(): void;

        preventDefault(): void;

        animeCollection?: AL_AnimeCollection;
        localFiles?: Array<Anime_LocalFile>;
        silencedMediaIds?: Array<number>;
        missingEpisodes?: Anime_MissingEpisodes;
    }

    /**
     * @event MissingEpisodesEvent
     * @file internal/library/anime/hook_events.go
     * @description
     * MissingEpisodesEvent is triggered when the missing episodes are being returned.
     */
    function onMissingEpisodes(cb: (event: MissingEpisodesEvent) => void): void;

    interface MissingEpisodesEvent {
        next(): void;

        missingEpisodes?: Anime_MissingEpisodes;
    }

    /**
     * @event AnimeLibraryCollectionRequestedEvent
     * @file internal/library/anime/hook_events.go
     * @description
     * AnimeLibraryCollectionRequestedEvent is triggered when the user requests the library collection.
     * Prevent default to skip the default process and return the modified library collection.
     * If the modified library collection is nil, an error will be returned.
     */
    function onAnimeLibraryCollectionRequested(cb: (event: AnimeLibraryCollectionRequestedEvent) => void): void;

    interface AnimeLibraryCollectionRequestedEvent {
        next(): void;

        preventDefault(): void;

        animeCollection?: AL_AnimeCollection;
        localFiles?: Array<Anime_LocalFile>;
        libraryCollection?: Anime_LibraryCollection;
    }

    /**
     * @event AnimeLibraryCollectionEvent
     * @file internal/library/anime/hook_events.go
     * @description
     * AnimeLibraryCollectionEvent is triggered when the user requests the library collection.
     */
    function onAnimeLibraryCollection(cb: (event: AnimeLibraryCollectionEvent) => void): void;

    interface AnimeLibraryCollectionEvent {
        next(): void;

        libraryCollection?: Anime_LibraryCollection;
    }

    /**
     * @event AnimeLibraryStreamCollectionRequestedEvent
     * @file internal/library/anime/hook_events.go
     * @description
     * AnimeLibraryStreamCollectionRequestedEvent is triggered when the user requests the library stream collection.
     * This is called when the user enables "Include in library" for either debrid/online/torrent streamings.
     */
    function onAnimeLibraryStreamCollectionRequested(cb: (event: AnimeLibraryStreamCollectionRequestedEvent) => void): void;

    interface AnimeLibraryStreamCollectionRequestedEvent {
        next(): void;

        animeCollection?: AL_AnimeCollection;
        libraryCollection?: Anime_LibraryCollection;
    }

    /**
     * @event AnimeLibraryStreamCollectionEvent
     * @file internal/library/anime/hook_events.go
     * @description
     * AnimeLibraryStreamCollectionEvent is triggered when the library stream collection is being returned.
     */
    function onAnimeLibraryStreamCollection(cb: (event: AnimeLibraryStreamCollectionEvent) => void): void;

    interface AnimeLibraryStreamCollectionEvent {
        next(): void;

        streamCollection?: Anime_StreamCollection;
    }

    /**
     * @event AnimeEntryDownloadInfoRequestedEvent
     * @file internal/library/anime/hook_events.go
     * @description
     * AnimeEntryDownloadInfoRequestedEvent is triggered when the app requests the download info for a media entry.
     * This is triggered before [AnimeEntryDownloadInfoEvent].
     */
    function onAnimeEntryDownloadInfoRequested(cb: (event: AnimeEntryDownloadInfoRequestedEvent) => void): void;

    interface AnimeEntryDownloadInfoRequestedEvent {
        next(): void;

        localFiles?: Array<Anime_LocalFile>;
        AnimeMetadata?: Metadata_AnimeMetadata;
        Media?: AL_BaseAnime;
        Progress?: number;
        Status?: AL_MediaListStatus;
        entryDownloadInfo?: Anime_EntryDownloadInfo;
    }

    /**
     * @event AnimeEntryDownloadInfoEvent
     * @file internal/library/anime/hook_events.go
     * @description
     * AnimeEntryDownloadInfoEvent is triggered when the download info is being returned.
     */
    function onAnimeEntryDownloadInfo(cb: (event: AnimeEntryDownloadInfoEvent) => void): void;

    interface AnimeEntryDownloadInfoEvent {
        next(): void;

        entryDownloadInfo?: Anime_EntryDownloadInfo;
    }

    /**
     * @event AnimeEpisodeCollectionRequestedEvent
     * @file internal/library/anime/hook_events.go
     * @description
     * AnimeEpisodeCollectionRequestedEvent is triggered when the episode collection is being requested.
     * Prevent default to skip the default behavior and return your own data.
     */
    function onAnimeEpisodeCollectionRequested(cb: (event: AnimeEpisodeCollectionRequestedEvent) => void): void;

    interface AnimeEpisodeCollectionRequestedEvent {
        next(): void;

        preventDefault(): void;

        media?: AL_BaseAnime;
        metadata?: Metadata_AnimeMetadata;
        episodeCollection?: Anime_EpisodeCollection;
    }

    /**
     * @event AnimeEpisodeCollectionEvent
     * @file internal/library/anime/hook_events.go
     * @description
     * AnimeEpisodeCollectionEvent is triggered when the episode collection is being returned.
     */
    function onAnimeEpisodeCollection(cb: (event: AnimeEpisodeCollectionEvent) => void): void;

    interface AnimeEpisodeCollectionEvent {
        next(): void;

        episodeCollection?: Anime_EpisodeCollection;
    }


    /**
     * @package anizip
     */

    /**
     * @event AnizipMediaRequestedEvent
     * @file internal/api/anizip/hook_events.go
     * @description
     * AnizipMediaRequestedEvent is triggered when the AniZip media is requested.
     * Prevent default to skip the default behavior and return your own data.
     */
    function onAnizipMediaRequested(cb: (event: AnizipMediaRequestedEvent) => void): void;

    interface AnizipMediaRequestedEvent {
        next(): void;

        preventDefault(): void;

        from: string;
        id: number;
        media?: Anizip_Media;
    }

    /**
     * @event AnizipMediaEvent
     * @file internal/api/anizip/hook_events.go
     * @description
     * AnizipMediaEvent is triggered after processing AnizipMedia.
     */
    function onAnizipMedia(cb: (event: AnizipMediaEvent) => void): void;

    interface AnizipMediaEvent {
        next(): void;

        media?: Anizip_Media;
    }


    /**
     * @package autodownloader
     */

    /**
     * @event AutoDownloaderRunStartedEvent
     * @file internal/library/autodownloader/hook_events.go
     * @description
     * AutoDownloaderRunStartedEvent is triggered when the autodownloader starts checking for new episodes.
     * Prevent default to abort the run.
     */
    function onAutoDownloaderRunStarted(cb: (event: AutoDownloaderRunStartedEvent) => void): void;

    interface AutoDownloaderRunStartedEvent {
        next(): void;

        preventDefault(): void;

        rules?: Array<Anime_AutoDownloaderRule>;
    }

    /**
     * @event AutoDownloaderTorrentsFetchedEvent
     * @file internal/library/autodownloader/hook_events.go
     * @description
     * AutoDownloaderTorrentsFetchedEvent is triggered at the beginning of a run, when the autodownloader fetches torrents from the provider.
     */
    function onAutoDownloaderTorrentsFetched(cb: (event: AutoDownloaderTorrentsFetchedEvent) => void): void;

    interface AutoDownloaderTorrentsFetchedEvent {
        next(): void;

        torrents?: Array<AutoDownloader_NormalizedTorrent>;
    }

    /**
     * @event AutoDownloaderMatchVerifiedEvent
     * @file internal/library/autodownloader/hook_events.go
     * @description
     * AutoDownloaderMatchVerifiedEvent is triggered when a torrent is verified to follow a rule.
     * Prevent default to abort the download if the match is found.
     */
    function onAutoDownloaderMatchVerified(cb: (event: AutoDownloaderMatchVerifiedEvent) => void): void;

    interface AutoDownloaderMatchVerifiedEvent {
        next(): void;

        preventDefault(): void;

        torrent?: AutoDownloader_NormalizedTorrent;
        rule?: Anime_AutoDownloaderRule;
        listEntry?: AL_AnimeListEntry;
        localEntry?: Anime_LocalFileWrapperEntry;
        episode: number;
        matchFound: boolean;
    }

    /**
     * @event AutoDownloaderSettingsUpdatedEvent
     * @file internal/library/autodownloader/hook_events.go
     * @description
     * AutoDownloaderSettingsUpdatedEvent is triggered when the autodownloader settings are updated
     */
    function onAutoDownloaderSettingsUpdated(cb: (event: AutoDownloaderSettingsUpdatedEvent) => void): void;

    interface AutoDownloaderSettingsUpdatedEvent {
        next(): void;

        settings?: Models_AutoDownloaderSettings;
    }

    /**
     * @event AutoDownloaderBeforeDownloadTorrentEvent
     * @file internal/library/autodownloader/hook_events.go
     * @description
     * AutoDownloaderBeforeDownloadTorrentEvent is triggered when the autodownloader is about to download a torrent.
     * Prevent default to abort the download.
     */
    function onAutoDownloaderBeforeDownloadTorrent(cb: (event: AutoDownloaderBeforeDownloadTorrentEvent) => void): void;

    interface AutoDownloaderBeforeDownloadTorrentEvent {
        next(): void;

        preventDefault(): void;

        torrent?: AutoDownloader_NormalizedTorrent;
        rule?: Anime_AutoDownloaderRule;
        items?: Array<Models_AutoDownloaderItem>;
    }

    /**
     * @event AutoDownloaderAfterDownloadTorrentEvent
     * @file internal/library/autodownloader/hook_events.go
     * @description
     * AutoDownloaderAfterDownloadTorrentEvent is triggered when the autodownloader has downloaded a torrent.
     */
    function onAutoDownloaderAfterDownloadTorrent(cb: (event: AutoDownloaderAfterDownloadTorrentEvent) => void): void;

    interface AutoDownloaderAfterDownloadTorrentEvent {
        next(): void;

        torrent?: AutoDownloader_NormalizedTorrent;
        rule?: Anime_AutoDownloaderRule;
    }


    /**
     * @package continuity
     */

    /**
     * @event WatchHistoryItemRequestedEvent
     * @file internal/continuity/hook_events.go
     * @description
     * WatchHistoryItemRequestedEvent is triggered when a watch history item is requested.
     * Prevent default to skip getting the watch history item from the file cache, in this case the event should have a valid WatchHistoryItem object
     *     or set it to nil to indicate that the watch history item was not found.
     */
    function onWatchHistoryItemRequested(cb: (event: WatchHistoryItemRequestedEvent) => void): void;

    interface WatchHistoryItemRequestedEvent {
        next(): void;

        preventDefault(): void;

        mediaId: number;
        watchHistoryItem?: Continuity_WatchHistoryItem;
    }

    /**
     * @event WatchHistoryItemUpdatedEvent
     * @file internal/continuity/hook_events.go
     * @description
     * WatchHistoryItemUpdatedEvent is triggered when a watch history item is updated.
     */
    function onWatchHistoryItemUpdated(cb: (event: WatchHistoryItemUpdatedEvent) => void): void;

    interface WatchHistoryItemUpdatedEvent {
        next(): void;

        watchHistoryItem?: Continuity_WatchHistoryItem;
    }

    /**
     * @event WatchHistoryLocalFileEpisodeItemRequestedEvent
     * @file internal/continuity/hook_events.go
     */
    function onWatchHistoryLocalFileEpisodeItemRequested(cb: (event: WatchHistoryLocalFileEpisodeItemRequestedEvent) => void): void;

    interface WatchHistoryLocalFileEpisodeItemRequestedEvent {
        next(): void;

        Path: string;
        LocalFiles?: Array<Anime_LocalFile>;
        watchHistoryItem?: Continuity_WatchHistoryItem;
    }

    /**
     * @event WatchHistoryStreamEpisodeItemRequestedEvent
     * @file internal/continuity/hook_events.go
     */
    function onWatchHistoryStreamEpisodeItemRequested(cb: (event: WatchHistoryStreamEpisodeItemRequestedEvent) => void): void;

    interface WatchHistoryStreamEpisodeItemRequestedEvent {
        next(): void;

        Episode: number;
        MediaId: number;
        watchHistoryItem?: Continuity_WatchHistoryItem;
    }


    /**
     * @package debrid_client
     */

    /**
     * @event DebridAutoSelectTorrentsFetchedEvent
     * @file internal/debrid/client/hook_events.go
     * @description
     * DebridAutoSelectTorrentsFetchedEvent is triggered when the torrents are fetched for auto select.
     * The torrents are sorted by seeders from highest to lowest.
     * This event is triggered before the top 3 torrents are analyzed.
     */
    function onDebridAutoSelectTorrentsFetched(cb: (event: DebridAutoSelectTorrentsFetchedEvent) => void): void;

    interface DebridAutoSelectTorrentsFetchedEvent {
        next(): void;

        Torrents?: Array<HibikeTorrent_AnimeTorrent>;
    }

    /**
     * @event DebridSkipStreamCheckEvent
     * @file internal/debrid/client/hook_events.go
     * @description
     * DebridSkipStreamCheckEvent is triggered when the debrid client is about to skip the stream check.
     * Prevent default to enable the stream check.
     */
    function onDebridSkipStreamCheck(cb: (event: DebridSkipStreamCheckEvent) => void): void;

    interface DebridSkipStreamCheckEvent {
        next(): void;

        preventDefault(): void;

        streamURL: string;
        retries: number;
    /**
     * in seconds
     */
        retryDelay: number;
    }

    /**
     * @event DebridSendStreamToMediaPlayerEvent
     * @file internal/debrid/client/hook_events.go
     * @description
     * DebridSendStreamToMediaPlayerEvent is triggered when the debrid client is about to send a stream to the media player.
     * Prevent default to skip the playback.
     */
    function onDebridSendStreamToMediaPlayer(cb: (event: DebridSendStreamToMediaPlayerEvent) => void): void;

    interface DebridSendStreamToMediaPlayerEvent {
        next(): void;

        preventDefault(): void;

        windowTitle: string;
        streamURL: string;
        media?: AL_BaseAnime;
        aniDbEpisode: string;
        playbackType: string;
    }

    /**
     * @event DebridLocalDownloadRequestedEvent
     * @file internal/debrid/client/hook_events.go
     * @description
     * DebridLocalDownloadRequestedEvent is triggered when Seanime is about to download a debrid torrent locally.
     * Prevent default to skip the default download and override the download.
     */
    function onDebridLocalDownloadRequested(cb: (event: DebridLocalDownloadRequestedEvent) => void): void;

    interface DebridLocalDownloadRequestedEvent {
        next(): void;

        preventDefault(): void;

        torrentName: string;
        destination: string;
        downloadUrl: string;
    }


    /**
     * @package discordrpc_presence
     */

    /**
     * @event DiscordPresenceAnimeActivityRequestedEvent
     * @file internal/discordrpc/presence/hook_events.go
     * @description
     * DiscordPresenceAnimeActivityRequestedEvent is triggered when anime activity is requested, after the [animeActivity] is processed, and right
     *     before the activity is sent to queue. There is no guarantee as to when or if the activity will be successfully sent to discord. Note that
     *     this event is triggered every 6 seconds or so, avoid heavy processing or perform it only when the activity is changed. Prevent default to
     *     stop the activity from being sent to discord.
     */
    function onDiscordPresenceAnimeActivityRequested(cb: (event: DiscordPresenceAnimeActivityRequestedEvent) => void): void;

    interface DiscordPresenceAnimeActivityRequestedEvent {
        next(): void;

        preventDefault(): void;

        animeActivity?: DiscordRPC_AnimeActivity;
        name: string;
        details: string;
        state: string;
        startTimestamp?: number;
        endTimestamp?: number;
        largeImage: string;
        largeText: string;
        smallImage: string;
        smallText: string;
        buttons?: Array<DiscordRPC_Button>;
        instance: boolean;
        type: number;
    }

    /**
     * @event DiscordPresenceMangaActivityRequestedEvent
     * @file internal/discordrpc/presence/hook_events.go
     * @description
     * DiscordPresenceMangaActivityRequestedEvent is triggered when manga activity is requested, after the [mangaActivity] is processed, and right
     *     before the activity is sent to queue. There is no guarantee as to when or if the activity will be successfully sent to discord. Note that
     *     this event is triggered every 6 seconds or so, avoid heavy processing or perform it only when the activity is changed. Prevent default to
     *     stop the activity from being sent to discord.
     */
    function onDiscordPresenceMangaActivityRequested(cb: (event: DiscordPresenceMangaActivityRequestedEvent) => void): void;

    interface DiscordPresenceMangaActivityRequestedEvent {
        next(): void;

        preventDefault(): void;

        mangaActivity?: DiscordRPC_MangaActivity;
        name: string;
        details: string;
        state: string;
        startTimestamp?: number;
        endTimestamp?: number;
        largeImage: string;
        largeText: string;
        smallImage: string;
        smallText: string;
        buttons?: Array<DiscordRPC_Button>;
        instance: boolean;
        type: number;
    }

    /**
     * @event DiscordPresenceClientClosedEvent
     * @file internal/discordrpc/presence/hook_events.go
     * @description
     * DiscordPresenceClientClosedEvent is triggered when the discord rpc client is closed.
     */
    function onDiscordPresenceClientClosed(cb: (event: DiscordPresenceClientClosedEvent) => void): void;

    interface DiscordPresenceClientClosedEvent {
        next(): void;

    }


    /**
     * @package manga
     */

    /**
     * @event MangaEntryRequestedEvent
     * @file internal/manga/hook_events.go
     * @description
     * MangaEntryRequestedEvent is triggered when a manga entry is requested.
     * Prevent default to skip the default behavior and return the modified entry.
     * If the modified entry is nil, an error will be returned.
     */
    function onMangaEntryRequested(cb: (event: MangaEntryRequestedEvent) => void): void;

    interface MangaEntryRequestedEvent {
        next(): void;

        preventDefault(): void;

        mediaId: number;
        mangaCollection?: AL_MangaCollection;
        entry?: Manga_Entry;
    }

    /**
     * @event MangaEntryEvent
     * @file internal/manga/hook_events.go
     * @description
     * MangaEntryEvent is triggered when the manga entry is being returned.
     */
    function onMangaEntry(cb: (event: MangaEntryEvent) => void): void;

    interface MangaEntryEvent {
        next(): void;

        entry?: Manga_Entry;
    }

    /**
     * @event MangaLibraryCollectionRequestedEvent
     * @file internal/manga/hook_events.go
     * @description
     * MangaLibraryCollectionRequestedEvent is triggered when the manga library collection is being requested.
     */
    function onMangaLibraryCollectionRequested(cb: (event: MangaLibraryCollectionRequestedEvent) => void): void;

    interface MangaLibraryCollectionRequestedEvent {
        next(): void;

        mangaCollection?: AL_MangaCollection;
    }

    /**
     * @event MangaLibraryCollectionEvent
     * @file internal/manga/hook_events.go
     * @description
     * MangaLibraryCollectionEvent is triggered when the manga library collection is being returned.
     */
    function onMangaLibraryCollection(cb: (event: MangaLibraryCollectionEvent) => void): void;

    interface MangaLibraryCollectionEvent {
        next(): void;

        libraryCollection?: Manga_Collection;
    }

    /**
     * @event MangaDownloadedChapterContainersRequestedEvent
     * @file internal/manga/hook_events.go
     * @description
     * MangaDownloadedChapterContainersRequestedEvent is triggered when the manga downloaded chapter containers are being requested.
     * Prevent default to skip the default behavior and return the modified chapter containers.
     * If the modified chapter containers are nil, an error will be returned.
     */
    function onMangaDownloadedChapterContainersRequested(cb: (event: MangaDownloadedChapterContainersRequestedEvent) => void): void;

    interface MangaDownloadedChapterContainersRequestedEvent {
        next(): void;

        preventDefault(): void;

        mangaCollection?: AL_MangaCollection;
        chapterContainers?: Array<Manga_ChapterContainer>;
    }

    /**
     * @event MangaDownloadedChapterContainersEvent
     * @file internal/manga/hook_events.go
     * @description
     * MangaDownloadedChapterContainersEvent is triggered when the manga downloaded chapter containers are being returned.
     */
    function onMangaDownloadedChapterContainers(cb: (event: MangaDownloadedChapterContainersEvent) => void): void;

    interface MangaDownloadedChapterContainersEvent {
        next(): void;

        chapterContainers?: Array<Manga_ChapterContainer>;
    }

    /**
     * @event MangaLatestChapterNumbersMapEvent
     * @file internal/manga/hook_events.go
     * @description
     * MangaLatestChapterNumbersMapEvent is triggered when the manga latest chapter numbers map is being returned.
     */
    function onMangaLatestChapterNumbersMap(cb: (event: MangaLatestChapterNumbersMapEvent) => void): void;

    interface MangaLatestChapterNumbersMapEvent {
        next(): void;

        latestChapterNumbersMap?: Record<number, Array<Manga_MangaLatestChapterNumberItem>>;
    }

    /**
     * @event MangaDownloadMapEvent
     * @file internal/manga/hook_events.go
     * @description
     * MangaDownloadMapEvent is triggered when the manga download map has been updated.
     * This map is used to tell the client which chapters have been downloaded.
     */
    function onMangaDownloadMap(cb: (event: MangaDownloadMapEvent) => void): void;

    interface MangaDownloadMapEvent {
        next(): void;

        mediaMap?: Manga_MediaMap;
    }

    /**
     * @event MangaChapterContainerRequestedEvent
     * @file internal/manga/hook_events.go
     * @description
     * MangaChapterContainerRequestedEvent is triggered when the manga chapter container is being requested.
     * This event happens before the chapter container is fetched from the cache or provider.
     * Prevent default to skip the default behavior and return the modified chapter container.
     * If the modified chapter container is nil, an error will be returned.
     */
    function onMangaChapterContainerRequested(cb: (event: MangaChapterContainerRequestedEvent) => void): void;

    interface MangaChapterContainerRequestedEvent {
        next(): void;

        preventDefault(): void;

        provider: string;
        mediaId: number;
        titles?: Array<string>;
        year: number;
        chapterContainer?: Manga_ChapterContainer;
    }

    /**
     * @event MangaChapterContainerEvent
     * @file internal/manga/hook_events.go
     * @description
     * MangaChapterContainerEvent is triggered when the manga chapter container is being returned.
     * This event happens after the chapter container is fetched from the cache or provider.
     */
    function onMangaChapterContainer(cb: (event: MangaChapterContainerEvent) => void): void;

    interface MangaChapterContainerEvent {
        next(): void;

        chapterContainer?: Manga_ChapterContainer;
    }


    /**
     * @package mediaplayer
     */

    /**
     * @event MediaPlayerLocalFileTrackingRequestedEvent
     * @file internal/mediaplayers/mediaplayer/hook_events.go
     * @description
     * MediaPlayerLocalFileTrackingRequestedEvent is triggered when the playback manager wants to track the progress of a local file.
     * Prevent default to stop tracking.
     */
    function onMediaPlayerLocalFileTrackingRequested(cb: (event: MediaPlayerLocalFileTrackingRequestedEvent) => void): void;

    interface MediaPlayerLocalFileTrackingRequestedEvent {
        next(): void;

        preventDefault(): void;

        startRefreshDelay: number;
        refreshDelay: number;
        maxRetries: number;
    }

    /**
     * @event MediaPlayerStreamTrackingRequestedEvent
     * @file internal/mediaplayers/mediaplayer/hook_events.go
     * @description
     * MediaPlayerStreamTrackingRequestedEvent is triggered when the playback manager wants to track the progress of a stream.
     * Prevent default to stop tracking.
     */
    function onMediaPlayerStreamTrackingRequested(cb: (event: MediaPlayerStreamTrackingRequestedEvent) => void): void;

    interface MediaPlayerStreamTrackingRequestedEvent {
        next(): void;

        preventDefault(): void;

        startRefreshDelay: number;
        refreshDelay: number;
        maxRetries: number;
        maxRetriesAfterStart: number;
    }


    /**
     * @package metadata
     */

    /**
     * @event AnimeMetadataRequestedEvent
     * @file internal/api/metadata/hook_events.go
     * @description
     * AnimeMetadataRequestedEvent is triggered when anime metadata is requested and right before the metadata is processed.
     * This event is followed by [AnimeMetadataEvent] which is triggered when the metadata is available.
     * Prevent default to skip the default behavior and return the modified metadata.
     * If the modified metadata is nil, an error will be returned.
     */
    function onAnimeMetadataRequested(cb: (event: AnimeMetadataRequestedEvent) => void): void;

    interface AnimeMetadataRequestedEvent {
        next(): void;

        preventDefault(): void;

        mediaId: number;
        animeMetadata?: Metadata_AnimeMetadata;
    }

    /**
     * @event AnimeMetadataEvent
     * @file internal/api/metadata/hook_events.go
     * @description
     * AnimeMetadataEvent is triggered when anime metadata is available and is about to be returned.
     * Anime metadata can be requested in many places, ranging from displaying the anime entry to starting a torrent stream.
     * This event is triggered after [AnimeMetadataRequestedEvent].
     * If the modified metadata is nil, an error will be returned.
     */
    function onAnimeMetadata(cb: (event: AnimeMetadataEvent) => void): void;

    interface AnimeMetadataEvent {
        next(): void;

        mediaId: number;
        animeMetadata?: Metadata_AnimeMetadata;
    }

    /**
     * @event AnimeEpisodeMetadataRequestedEvent
     * @file internal/api/metadata/hook_events.go
     * @description
     * AnimeEpisodeMetadataRequestedEvent is triggered when anime episode metadata is requested.
     * Prevent default to skip the default behavior and return the overridden metadata.
     * This event is triggered before [AnimeEpisodeMetadataEvent].
     * If the modified episode metadata is nil, an empty EpisodeMetadata object will be returned.
     */
    function onAnimeEpisodeMetadataRequested(cb: (event: AnimeEpisodeMetadataRequestedEvent) => void): void;

    interface AnimeEpisodeMetadataRequestedEvent {
        next(): void;

        preventDefault(): void;

        animeEpisodeMetadata?: Metadata_EpisodeMetadata;
        episodeNumber: number;
        mediaId: number;
    }

    /**
     * @event AnimeEpisodeMetadataEvent
     * @file internal/api/metadata/hook_events.go
     * @description
     * AnimeEpisodeMetadataEvent is triggered when anime episode metadata is available and is about to be returned.
     * In the current implementation, episode metadata is requested for display purposes. It is used to get a more complete metadata object since the
     *     original AnimeMetadata object is not complete. This event is triggered after [AnimeEpisodeMetadataRequestedEvent]. If the modified episode
     *     metadata is nil, an empty EpisodeMetadata object will be returned.
     */
    function onAnimeEpisodeMetadata(cb: (event: AnimeEpisodeMetadataEvent) => void): void;

    interface AnimeEpisodeMetadataEvent {
        next(): void;

        animeEpisodeMetadata?: Metadata_EpisodeMetadata;
        episodeNumber: number;
        mediaId: number;
    }


    /**
     * @package playbackmanager
     */

    /**
     * @event LocalFilePlaybackRequestedEvent
     * @file internal/library/playbackmanager/hook_events.go
     * @description
     * LocalFilePlaybackRequestedEvent is triggered when a local file is requested to be played.
     * Prevent default to skip the default playback and override the playback.
     */
    function onLocalFilePlaybackRequested(cb: (event: LocalFilePlaybackRequestedEvent) => void): void;

    interface LocalFilePlaybackRequestedEvent {
        next(): void;

        preventDefault(): void;

        path: string;
    }

    /**
     * @event StreamPlaybackRequestedEvent
     * @file internal/library/playbackmanager/hook_events.go
     * @description
     * StreamPlaybackRequestedEvent is triggered when a stream is requested to be played.
     * Prevent default to skip the default playback and override the playback.
     */
    function onStreamPlaybackRequested(cb: (event: StreamPlaybackRequestedEvent) => void): void;

    interface StreamPlaybackRequestedEvent {
        next(): void;

        preventDefault(): void;

        windowTitle: string;
        payload: string;
        media?: AL_BaseAnime;
        aniDbEpisode: string;
    }

    /**
     * @event PlaybackBeforeTrackingEvent
     * @file internal/library/playbackmanager/hook_events.go
     * @description
     * PlaybackBeforeTrackingEvent is triggered just before the playback tracking starts.
     * Prevent default to skip playback tracking.
     */
    function onPlaybackBeforeTracking(cb: (event: PlaybackBeforeTrackingEvent) => void): void;

    interface PlaybackBeforeTrackingEvent {
        next(): void;

        preventDefault(): void;

        isStream: boolean;
    }

    /**
     * @event PlaybackLocalFileDetailsRequestedEvent
     * @file internal/library/playbackmanager/hook_events.go
     * @description
     * PlaybackLocalFileDetailsRequestedEvent is triggered when the local files details for a specific path are requested.
     * This event is triggered right after the media player loads an episode.
     * The playback manager uses the local files details to track the progress, propose next episodes, etc.
     * In the current implementation, the details are fetched by selecting the local file from the database and making requests to retrieve the media
     *     and anime list entry. Prevent default to skip the default fetching and override the details.
     */
    function onPlaybackLocalFileDetailsRequested(cb: (event: PlaybackLocalFileDetailsRequestedEvent) => void): void;

    interface PlaybackLocalFileDetailsRequestedEvent {
        next(): void;

        preventDefault(): void;

        path: string;
        localFiles?: Array<Anime_LocalFile>;
        animeListEntry?: AL_AnimeListEntry;
        localFile?: Anime_LocalFile;
        localFileWrapperEntry?: Anime_LocalFileWrapperEntry;
    }

    /**
     * @event PlaybackStreamDetailsRequestedEvent
     * @file internal/library/playbackmanager/hook_events.go
     * @description
     * PlaybackStreamDetailsRequestedEvent is triggered when the stream details are requested.
     * Prevent default to skip the default fetching and override the details.
     * In the current implementation, the details are fetched by selecting the anime from the anime collection. If nothing is found, the stream is
     *     still tracked.
     */
    function onPlaybackStreamDetailsRequested(cb: (event: PlaybackStreamDetailsRequestedEvent) => void): void;

    interface PlaybackStreamDetailsRequestedEvent {
        next(): void;

        preventDefault(): void;

        animeCollection?: AL_AnimeCollection;
        mediaId: number;
        animeListEntry?: AL_AnimeListEntry;
    }


    /**
     * @package scanner
     */

    /**
     * @event ScanStartedEvent
     * @file internal/library/scanner/hook_events.go
     * @description
     * ScanStartedEvent is triggered when the scanning process begins.
     * Prevent default to skip the rest of the scanning process and return the local files.
     */
    function onScanStarted(cb: (event: ScanStartedEvent) => void): void;

    interface ScanStartedEvent {
        next(): void;

        preventDefault(): void;

        libraryPath: string;
        otherLibraryPaths?: Array<string>;
        enhanced: boolean;
        skipLocked: boolean;
        skipIgnored: boolean;
        localFiles?: Array<Anime_LocalFile>;
    }

    /**
     * @event ScanFilePathsRetrievedEvent
     * @file internal/library/scanner/hook_events.go
     * @description
     * ScanFilePathsRetrievedEvent is triggered when the file paths to scan are retrieved.
     * The event includes file paths from all directories to scan.
     * The event includes file paths of local files that will be skipped.
     */
    function onScanFilePathsRetrieved(cb: (event: ScanFilePathsRetrievedEvent) => void): void;

    interface ScanFilePathsRetrievedEvent {
        next(): void;

        filePaths?: Array<string>;
    }

    /**
     * @event ScanLocalFilesParsedEvent
     * @file internal/library/scanner/hook_events.go
     * @description
     * ScanLocalFilesParsedEvent is triggered right after the file paths are parsed into local file objects.
     * The event does not include local files that are skipped.
     */
    function onScanLocalFilesParsed(cb: (event: ScanLocalFilesParsedEvent) => void): void;

    interface ScanLocalFilesParsedEvent {
        next(): void;

        localFiles?: Array<Anime_LocalFile>;
    }

    /**
     * @event ScanCompletedEvent
     * @file internal/library/scanner/hook_events.go
     * @description
     * ScanCompletedEvent is triggered when the scanning process finishes.
     * The event includes all the local files (skipped and scanned) to be inserted as a new entry.
     * Right after this event, the local files will be inserted as a new entry.
     */
    function onScanCompleted(cb: (event: ScanCompletedEvent) => void): void;

    interface ScanCompletedEvent {
        next(): void;

        localFiles?: Array<Anime_LocalFile>;
    /**
     * in milliseconds
     */
        duration: number;
    }

    /**
     * @event ScanMediaFetcherStartedEvent
     * @file internal/library/scanner/hook_events.go
     * @description
     * ScanMediaFetcherStartedEvent is triggered right before Seanime starts fetching media to be matched against the local files.
     */
    function onScanMediaFetcherStarted(cb: (event: ScanMediaFetcherStartedEvent) => void): void;

    interface ScanMediaFetcherStartedEvent {
        next(): void;

        enhanced: boolean;
    }

    /**
     * @event ScanMediaFetcherCompletedEvent
     * @file internal/library/scanner/hook_events.go
     * @description
     * ScanMediaFetcherCompletedEvent is triggered when the media fetcher completes.
     * The event includes all the media fetched from AniList.
     * The event includes the media IDs that are not in the user's collection.
     */
    function onScanMediaFetcherCompleted(cb: (event: ScanMediaFetcherCompletedEvent) => void): void;

    interface ScanMediaFetcherCompletedEvent {
        next(): void;

        allMedia?: Array<AL_CompleteAnime>;
        unknownMediaIds?: Array<number>;
    }

    /**
     * @event ScanMatchingStartedEvent
     * @file internal/library/scanner/hook_events.go
     * @description
     * ScanMatchingStartedEvent is triggered when the matching process begins.
     * Prevent default to skip the default matching, in which case modified local files will be used.
     */
    function onScanMatchingStarted(cb: (event: ScanMatchingStartedEvent) => void): void;

    interface ScanMatchingStartedEvent {
        next(): void;

        preventDefault(): void;

        localFiles?: Array<Anime_LocalFile>;
        normalizedMedia?: Array<Anime_NormalizedMedia>;
        algorithm: string;
        threshold: number;
    }

    /**
     * @event ScanLocalFileMatchedEvent
     * @file internal/library/scanner/hook_events.go
     * @description
     * ScanLocalFileMatchedEvent is triggered when a local file is matched with media and before the match is analyzed.
     * Prevent default to skip the default analysis and override the match.
     */
    function onScanLocalFileMatched(cb: (event: ScanLocalFileMatchedEvent) => void): void;

    interface ScanLocalFileMatchedEvent {
        next(): void;

        preventDefault(): void;

        match?: Anime_NormalizedMedia;
        found: boolean;
        localFile?: Anime_LocalFile;
        score: number;
    }

    /**
     * @event ScanMatchingCompletedEvent
     * @file internal/library/scanner/hook_events.go
     * @description
     * ScanMatchingCompletedEvent is triggered when the matching process completes.
     */
    function onScanMatchingCompleted(cb: (event: ScanMatchingCompletedEvent) => void): void;

    interface ScanMatchingCompletedEvent {
        next(): void;

        localFiles?: Array<Anime_LocalFile>;
    }

    /**
     * @event ScanHydrationStartedEvent
     * @file internal/library/scanner/hook_events.go
     * @description
     * ScanHydrationStartedEvent is triggered when the file hydration process begins.
     * Prevent default to skip the rest of the hydration process, in which case the event's local files will be used.
     */
    function onScanHydrationStarted(cb: (event: ScanHydrationStartedEvent) => void): void;

    interface ScanHydrationStartedEvent {
        next(): void;

        preventDefault(): void;

        localFiles?: Array<Anime_LocalFile>;
        allMedia?: Array<Anime_NormalizedMedia>;
    }

    /**
     * @event ScanLocalFileHydrationStartedEvent
     * @file internal/library/scanner/hook_events.go
     * @description
     * ScanLocalFileHydrationStartedEvent is triggered when a local file's metadata is about to be hydrated.
     * Prevent default to skip the default hydration and override the hydration.
     */
    function onScanLocalFileHydrationStarted(cb: (event: ScanLocalFileHydrationStartedEvent) => void): void;

    interface ScanLocalFileHydrationStartedEvent {
        next(): void;

        preventDefault(): void;

        localFile?: Anime_LocalFile;
        media?: Anime_NormalizedMedia;
    }

    /**
     * @event ScanLocalFileHydratedEvent
     * @file internal/library/scanner/hook_events.go
     * @description
     * ScanLocalFileHydratedEvent is triggered when a local file's metadata is hydrated
     */
    function onScanLocalFileHydrated(cb: (event: ScanLocalFileHydratedEvent) => void): void;

    interface ScanLocalFileHydratedEvent {
        next(): void;

        localFile?: Anime_LocalFile;
        mediaId: number;
        episode: number;
    }


    /**
     * @package torrentstream
     */

    /**
     * @event TorrentStreamAutoSelectTorrentsFetchedEvent
     * @file internal/torrentstream/hook_events.go
     * @description
     * TorrentStreamAutoSelectTorrentsFetchedEvent is triggered when the torrents are fetched for auto select.
     * The torrents are sorted by seeders from highest to lowest.
     * This event is triggered before the top 3 torrents are analyzed.
     */
    function onTorrentStreamAutoSelectTorrentsFetched(cb: (event: TorrentStreamAutoSelectTorrentsFetchedEvent) => void): void;

    interface TorrentStreamAutoSelectTorrentsFetchedEvent {
        next(): void;

        Torrents?: Array<HibikeTorrent_AnimeTorrent>;
    }

    /**
     * @event TorrentStreamSendStreamToMediaPlayerEvent
     * @file internal/torrentstream/hook_events.go
     * @description
     * TorrentStreamSendStreamToMediaPlayerEvent is triggered when the torrent stream is about to send a stream to the media player.
     * Prevent default to skip the default playback and override the playback.
     */
    function onTorrentStreamSendStreamToMediaPlayer(cb: (event: TorrentStreamSendStreamToMediaPlayerEvent) => void): void;

    interface TorrentStreamSendStreamToMediaPlayerEvent {
        next(): void;

        preventDefault(): void;

        windowTitle: string;
        streamURL: string;
        media?: AL_BaseAnime;
        aniDbEpisode: string;
        playbackType: string;
    }

    ///////////////////////////////////////////////////////////////////////////////////////////////////////////////
    ///////////////////////////////////////////////////////////////////////////////////////////////////////////////
    ///////////////////////////////////////////////////////////////////////////////////////////////////////////////

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_AnimeCollection {
        MediaListCollection?: AL_AnimeCollection_MediaListCollection;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_AnimeCollectionWithRelations {
        MediaListCollection?: AL_AnimeCollectionWithRelations_MediaListCollection;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_AnimeCollectionWithRelations_MediaListCollection {
        lists?: Array<AL_AnimeCollectionWithRelations_MediaListCollection_Lists>;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_AnimeCollectionWithRelations_MediaListCollection_Lists {
        status?: AL_MediaListStatus;
        name?: string;
        isCustomList?: boolean;
        entries?: Array<AL_AnimeCollectionWithRelations_MediaListCollection_Lists_Entries>;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_AnimeCollectionWithRelations_MediaListCollection_Lists_Entries {
        id: number;
        score?: number;
        progress?: number;
        status?: AL_MediaListStatus;
        notes?: string;
        repeat?: number;
        private?: boolean;
        startedAt?: AL_AnimeCollectionWithRelations_MediaListCollection_Lists_Entries_StartedAt;
        completedAt?: AL_AnimeCollectionWithRelations_MediaListCollection_Lists_Entries_CompletedAt;
        media?: AL_CompleteAnime;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_AnimeCollectionWithRelations_MediaListCollection_Lists_Entries_CompletedAt {
        year?: number;
        month?: number;
        day?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_AnimeCollectionWithRelations_MediaListCollection_Lists_Entries_StartedAt {
        year?: number;
        month?: number;
        day?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_AnimeCollection_MediaListCollection {
        lists?: Array<AL_AnimeCollection_MediaListCollection_Lists>;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_AnimeCollection_MediaListCollection_Lists {
        status?: AL_MediaListStatus;
        name?: string;
        isCustomList?: boolean;
        entries?: Array<AL_AnimeCollection_MediaListCollection_Lists_Entries>;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_AnimeCollection_MediaListCollection_Lists_Entries {
        id: number;
        score?: number;
        progress?: number;
        status?: AL_MediaListStatus;
        notes?: string;
        repeat?: number;
        private?: boolean;
        startedAt?: AL_AnimeCollection_MediaListCollection_Lists_Entries_StartedAt;
        completedAt?: AL_AnimeCollection_MediaListCollection_Lists_Entries_CompletedAt;
        media?: AL_BaseAnime;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_AnimeCollection_MediaListCollection_Lists_Entries_CompletedAt {
        year?: number;
        month?: number;
        day?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_AnimeCollection_MediaListCollection_Lists_Entries_StartedAt {
        year?: number;
        month?: number;
        day?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_AnimeDetailsById_Media {
        siteUrl?: string;
        id: number;
        duration?: number;
        genres?: Array<string>;
        averageScore?: number;
        popularity?: number;
        meanScore?: number;
        description?: string;
        trailer?: AL_AnimeDetailsById_Media_Trailer;
        startDate?: AL_AnimeDetailsById_Media_StartDate;
        endDate?: AL_AnimeDetailsById_Media_EndDate;
        studios?: AL_AnimeDetailsById_Media_Studios;
        characters?: AL_AnimeDetailsById_Media_Characters;
        staff?: AL_AnimeDetailsById_Media_Staff;
        rankings?: Array<AL_AnimeDetailsById_Media_Rankings>;
        recommendations?: AL_AnimeDetailsById_Media_Recommendations;
        relations?: AL_AnimeDetailsById_Media_Relations;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_AnimeDetailsById_Media_Characters {
        edges?: Array<AL_AnimeDetailsById_Media_Characters_Edges>;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_AnimeDetailsById_Media_Characters_Edges {
        id?: number;
        role?: AL_CharacterRole;
        name?: string;
        node?: AL_BaseCharacter;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_AnimeDetailsById_Media_EndDate {
        year?: number;
        month?: number;
        day?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_AnimeDetailsById_Media_Rankings {
        context: string;
        type: AL_MediaRankType;
        rank: number;
        year?: number;
        format: AL_MediaFormat;
        allTime?: boolean;
        season?: AL_MediaSeason;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_AnimeDetailsById_Media_Recommendations {
        edges?: Array<AL_AnimeDetailsById_Media_Recommendations_Edges>;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_AnimeDetailsById_Media_Recommendations_Edges {
        node?: AL_AnimeDetailsById_Media_Recommendations_Edges_Node;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_AnimeDetailsById_Media_Recommendations_Edges_Node {
        mediaRecommendation?: AL_AnimeDetailsById_Media_Recommendations_Edges_Node_MediaRecommendation;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_AnimeDetailsById_Media_Recommendations_Edges_Node_MediaRecommendation {
        id: number;
        idMal?: number;
        siteUrl?: string;
        status?: AL_MediaStatus;
        isAdult?: boolean;
        season?: AL_MediaSeason;
        type?: AL_MediaType;
        format?: AL_MediaFormat;
        meanScore?: number;
        description?: string;
        episodes?: number;
        trailer?: AL_AnimeDetailsById_Media_Recommendations_Edges_Node_MediaRecommendation_Trailer;
        startDate?: AL_AnimeDetailsById_Media_Recommendations_Edges_Node_MediaRecommendation_StartDate;
        coverImage?: AL_AnimeDetailsById_Media_Recommendations_Edges_Node_MediaRecommendation_CoverImage;
        bannerImage?: string;
        title?: AL_AnimeDetailsById_Media_Recommendations_Edges_Node_MediaRecommendation_Title;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_AnimeDetailsById_Media_Recommendations_Edges_Node_MediaRecommendation_CoverImage {
        extraLarge?: string;
        large?: string;
        medium?: string;
        color?: string;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_AnimeDetailsById_Media_Recommendations_Edges_Node_MediaRecommendation_StartDate {
        year?: number;
        month?: number;
        day?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_AnimeDetailsById_Media_Recommendations_Edges_Node_MediaRecommendation_Title {
        romaji?: string;
        english?: string;
        native?: string;
        userPreferred?: string;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_AnimeDetailsById_Media_Recommendations_Edges_Node_MediaRecommendation_Trailer {
        id?: string;
        site?: string;
        thumbnail?: string;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_AnimeDetailsById_Media_Relations {
        edges?: Array<AL_AnimeDetailsById_Media_Relations_Edges>;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_AnimeDetailsById_Media_Relations_Edges {
        relationType?: AL_MediaRelation;
        node?: AL_BaseAnime;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_AnimeDetailsById_Media_Staff {
        edges?: Array<AL_AnimeDetailsById_Media_Staff_Edges>;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_AnimeDetailsById_Media_Staff_Edges {
        role?: string;
        node?: AL_AnimeDetailsById_Media_Staff_Edges_Node;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_AnimeDetailsById_Media_Staff_Edges_Node {
        name?: AL_AnimeDetailsById_Media_Staff_Edges_Node_Name;
        id: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_AnimeDetailsById_Media_Staff_Edges_Node_Name {
        full?: string;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_AnimeDetailsById_Media_StartDate {
        year?: number;
        month?: number;
        day?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_AnimeDetailsById_Media_Studios {
        nodes?: Array<AL_AnimeDetailsById_Media_Studios_Nodes>;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_AnimeDetailsById_Media_Studios_Nodes {
        name: string;
        id: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_AnimeDetailsById_Media_Trailer {
        id?: string;
        site?: string;
        thumbnail?: string;
    }

    /**
     * - Filepath: internal/api/anilist/collection_helper.go
     */
    export type AL_AnimeListEntry = AL_AnimeCollection_MediaListCollection_Lists_Entries;

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_BaseAnime {
        id: number;
        idMal?: number;
        siteUrl?: string;
        status?: AL_MediaStatus;
        season?: AL_MediaSeason;
        type?: AL_MediaType;
        format?: AL_MediaFormat;
        seasonYear?: number;
        bannerImage?: string;
        episodes?: number;
        synonyms?: Array<string>;
        isAdult?: boolean;
        countryOfOrigin?: string;
        meanScore?: number;
        description?: string;
        genres?: Array<string>;
        duration?: number;
        trailer?: AL_BaseAnime_Trailer;
        title?: AL_BaseAnime_Title;
        coverImage?: AL_BaseAnime_CoverImage;
        startDate?: AL_BaseAnime_StartDate;
        endDate?: AL_BaseAnime_EndDate;
        nextAiringEpisode?: AL_BaseAnime_NextAiringEpisode;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_BaseAnime_CoverImage {
        extraLarge?: string;
        large?: string;
        medium?: string;
        color?: string;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_BaseAnime_EndDate {
        year?: number;
        month?: number;
        day?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_BaseAnime_NextAiringEpisode {
        airingAt: number;
        timeUntilAiring: number;
        episode: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_BaseAnime_StartDate {
        year?: number;
        month?: number;
        day?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_BaseAnime_Title {
        userPreferred?: string;
        romaji?: string;
        english?: string;
        native?: string;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_BaseAnime_Trailer {
        id?: string;
        site?: string;
        thumbnail?: string;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_BaseCharacter {
        id: number;
        isFavourite: boolean;
        gender?: string;
        age?: string;
        dateOfBirth?: AL_BaseCharacter_DateOfBirth;
        name?: AL_BaseCharacter_Name;
        image?: AL_BaseCharacter_Image;
        description?: string;
        siteUrl?: string;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_BaseCharacter_DateOfBirth {
        year?: number;
        month?: number;
        day?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_BaseCharacter_Image {
        large?: string;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_BaseCharacter_Name {
        full?: string;
        native?: string;
        alternative?: Array<string>;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_BaseManga {
        id: number;
        idMal?: number;
        siteUrl?: string;
        status?: AL_MediaStatus;
        season?: AL_MediaSeason;
        type?: AL_MediaType;
        format?: AL_MediaFormat;
        bannerImage?: string;
        chapters?: number;
        volumes?: number;
        synonyms?: Array<string>;
        isAdult?: boolean;
        countryOfOrigin?: string;
        meanScore?: number;
        description?: string;
        genres?: Array<string>;
        title?: AL_BaseManga_Title;
        coverImage?: AL_BaseManga_CoverImage;
        startDate?: AL_BaseManga_StartDate;
        endDate?: AL_BaseManga_EndDate;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_BaseManga_CoverImage {
        extraLarge?: string;
        large?: string;
        medium?: string;
        color?: string;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_BaseManga_EndDate {
        year?: number;
        month?: number;
        day?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_BaseManga_StartDate {
        year?: number;
        month?: number;
        day?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_BaseManga_Title {
        userPreferred?: string;
        romaji?: string;
        english?: string;
        native?: string;
    }

    /**
     * - Filepath: internal/api/anilist/models_gen.go
     * @description
     *  The role the character plays in the media
     */
    export type AL_CharacterRole = "MAIN" | "SUPPORTING" | "BACKGROUND";

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_CompleteAnime {
        id: number;
        idMal?: number;
        siteUrl?: string;
        status?: AL_MediaStatus;
        season?: AL_MediaSeason;
        seasonYear?: number;
        type?: AL_MediaType;
        format?: AL_MediaFormat;
        bannerImage?: string;
        episodes?: number;
        synonyms?: Array<string>;
        isAdult?: boolean;
        countryOfOrigin?: string;
        meanScore?: number;
        description?: string;
        genres?: Array<string>;
        duration?: number;
        trailer?: AL_CompleteAnime_Trailer;
        title?: AL_CompleteAnime_Title;
        coverImage?: AL_CompleteAnime_CoverImage;
        startDate?: AL_CompleteAnime_StartDate;
        endDate?: AL_CompleteAnime_EndDate;
        nextAiringEpisode?: AL_CompleteAnime_NextAiringEpisode;
        relations?: AL_CompleteAnime_Relations;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_CompleteAnime_CoverImage {
        extraLarge?: string;
        large?: string;
        medium?: string;
        color?: string;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_CompleteAnime_EndDate {
        year?: number;
        month?: number;
        day?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_CompleteAnime_NextAiringEpisode {
        airingAt: number;
        timeUntilAiring: number;
        episode: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_CompleteAnime_Relations {
        edges?: Array<AL_CompleteAnime_Relations_Edges>;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_CompleteAnime_Relations_Edges {
        relationType?: AL_MediaRelation;
        node?: AL_BaseAnime;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_CompleteAnime_StartDate {
        year?: number;
        month?: number;
        day?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_CompleteAnime_Title {
        userPreferred?: string;
        romaji?: string;
        english?: string;
        native?: string;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_CompleteAnime_Trailer {
        id?: string;
        site?: string;
        thumbnail?: string;
    }

    /**
     * - Filepath: internal/api/anilist/models_gen.go
     * @description
     *  Date object that allows for incomplete date values (fuzzy)
     */
    interface AL_FuzzyDateInput {
        year?: number;
        month?: number;
        day?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_ListAnime {
        Page?: AL_ListAnime_Page;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_ListAnime_Page {
        pageInfo?: AL_ListAnime_Page_PageInfo;
        media?: Array<AL_BaseAnime>;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_ListAnime_Page_PageInfo {
        hasNextPage?: boolean;
        total?: number;
        perPage?: number;
        currentPage?: number;
        lastPage?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_ListManga {
        Page?: AL_ListManga_Page;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_ListManga_Page {
        pageInfo?: AL_ListManga_Page_PageInfo;
        media?: Array<AL_BaseManga>;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_ListManga_Page_PageInfo {
        hasNextPage?: boolean;
        total?: number;
        perPage?: number;
        currentPage?: number;
        lastPage?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_ListRecentAnime {
        Page?: AL_ListRecentAnime_Page;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_ListRecentAnime_Page {
        pageInfo?: AL_ListRecentAnime_Page_PageInfo;
        airingSchedules?: Array<AL_ListRecentAnime_Page_AiringSchedules>;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_ListRecentAnime_Page_AiringSchedules {
        id: number;
        airingAt: number;
        episode: number;
        timeUntilAiring: number;
        media?: AL_BaseAnime;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_ListRecentAnime_Page_PageInfo {
        hasNextPage?: boolean;
        total?: number;
        perPage?: number;
        currentPage?: number;
        lastPage?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_MangaCollection {
        MediaListCollection?: AL_MangaCollection_MediaListCollection;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_MangaCollection_MediaListCollection {
        lists?: Array<AL_MangaCollection_MediaListCollection_Lists>;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_MangaCollection_MediaListCollection_Lists {
        status?: AL_MediaListStatus;
        name?: string;
        isCustomList?: boolean;
        entries?: Array<AL_MangaCollection_MediaListCollection_Lists_Entries>;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_MangaCollection_MediaListCollection_Lists_Entries {
        id: number;
        score?: number;
        progress?: number;
        status?: AL_MediaListStatus;
        notes?: string;
        repeat?: number;
        private?: boolean;
        startedAt?: AL_MangaCollection_MediaListCollection_Lists_Entries_StartedAt;
        completedAt?: AL_MangaCollection_MediaListCollection_Lists_Entries_CompletedAt;
        media?: AL_BaseManga;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_MangaCollection_MediaListCollection_Lists_Entries_CompletedAt {
        year?: number;
        month?: number;
        day?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_MangaCollection_MediaListCollection_Lists_Entries_StartedAt {
        year?: number;
        month?: number;
        day?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_MangaDetailsById_Media {
        siteUrl?: string;
        id: number;
        duration?: number;
        genres?: Array<string>;
        rankings?: Array<AL_MangaDetailsById_Media_Rankings>;
        characters?: AL_MangaDetailsById_Media_Characters;
        recommendations?: AL_MangaDetailsById_Media_Recommendations;
        relations?: AL_MangaDetailsById_Media_Relations;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_MangaDetailsById_Media_Characters {
        edges?: Array<AL_MangaDetailsById_Media_Characters_Edges>;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_MangaDetailsById_Media_Characters_Edges {
        id?: number;
        role?: AL_CharacterRole;
        name?: string;
        node?: AL_BaseCharacter;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_MangaDetailsById_Media_Rankings {
        context: string;
        type: AL_MediaRankType;
        rank: number;
        year?: number;
        format: AL_MediaFormat;
        allTime?: boolean;
        season?: AL_MediaSeason;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_MangaDetailsById_Media_Recommendations {
        edges?: Array<AL_MangaDetailsById_Media_Recommendations_Edges>;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_MangaDetailsById_Media_Recommendations_Edges {
        node?: AL_MangaDetailsById_Media_Recommendations_Edges_Node;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_MangaDetailsById_Media_Recommendations_Edges_Node {
        mediaRecommendation?: AL_MangaDetailsById_Media_Recommendations_Edges_Node_MediaRecommendation;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_MangaDetailsById_Media_Recommendations_Edges_Node_MediaRecommendation {
        id: number;
        idMal?: number;
        siteUrl?: string;
        status?: AL_MediaStatus;
        season?: AL_MediaSeason;
        type?: AL_MediaType;
        format?: AL_MediaFormat;
        bannerImage?: string;
        chapters?: number;
        volumes?: number;
        synonyms?: Array<string>;
        isAdult?: boolean;
        countryOfOrigin?: string;
        meanScore?: number;
        description?: string;
        title?: AL_MangaDetailsById_Media_Recommendations_Edges_Node_MediaRecommendation_Title;
        coverImage?: AL_MangaDetailsById_Media_Recommendations_Edges_Node_MediaRecommendation_CoverImage;
        startDate?: AL_MangaDetailsById_Media_Recommendations_Edges_Node_MediaRecommendation_StartDate;
        endDate?: AL_MangaDetailsById_Media_Recommendations_Edges_Node_MediaRecommendation_EndDate;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_MangaDetailsById_Media_Recommendations_Edges_Node_MediaRecommendation_CoverImage {
        extraLarge?: string;
        large?: string;
        medium?: string;
        color?: string;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_MangaDetailsById_Media_Recommendations_Edges_Node_MediaRecommendation_EndDate {
        year?: number;
        month?: number;
        day?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_MangaDetailsById_Media_Recommendations_Edges_Node_MediaRecommendation_StartDate {
        year?: number;
        month?: number;
        day?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_MangaDetailsById_Media_Recommendations_Edges_Node_MediaRecommendation_Title {
        userPreferred?: string;
        romaji?: string;
        english?: string;
        native?: string;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_MangaDetailsById_Media_Relations {
        edges?: Array<AL_MangaDetailsById_Media_Relations_Edges>;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_MangaDetailsById_Media_Relations_Edges {
        relationType?: AL_MediaRelation;
        node?: AL_BaseManga;
    }

    /**
     * - Filepath: internal/api/anilist/models_gen.go
     * @description
     *  The format the media was released in
     */
    export type AL_MediaFormat = "TV" |
    "TV_SHORT" |
    "MOVIE" |
    "SPECIAL" |
    "OVA" |
    "ONA" |
    "MUSIC" |
    "MANGA" |
    "NOVEL" |
    "ONE_SHOT";

    /**
     * - Filepath: internal/api/anilist/models_gen.go
     * @description
     *  Media list watching/reading status enum.
     */
    export type AL_MediaListStatus = "CURRENT" |
    "PLANNING" |
    "COMPLETED" |
    "DROPPED" |
    "PAUSED" |
    "REPEATING";

    /**
     * - Filepath: internal/api/anilist/models_gen.go
     * @description
     *  The type of ranking
     */
    export type AL_MediaRankType = "RATED" | "POPULAR";

    /**
     * - Filepath: internal/api/anilist/models_gen.go
     * @description
     *  Type of relation media has to its parent.
     */
    export type AL_MediaRelation = "ADAPTATION" |
    "PREQUEL" |
    "SEQUEL" |
    "PARENT" |
    "SIDE_STORY" |
    "CHARACTER" |
    "SUMMARY" |
    "ALTERNATIVE" |
    "SPIN_OFF" |
    "OTHER" |
    "SOURCE" |
    "COMPILATION" |
    "CONTAINS";

    /**
     * - Filepath: internal/api/anilist/models_gen.go
     */
    export type AL_MediaSeason = "WINTER" | "SPRING" | "SUMMER" | "FALL";

    /**
     * - Filepath: internal/api/anilist/models_gen.go
     * @description
     *  Media sort enums
     */
    export type AL_MediaSort = "ID" |
    "ID_DESC" |
    "TITLE_ROMAJI" |
    "TITLE_ROMAJI_DESC" |
    "TITLE_ENGLISH" |
    "TITLE_ENGLISH_DESC" |
    "TITLE_NATIVE" |
    "TITLE_NATIVE_DESC" |
    "TYPE" |
    "TYPE_DESC" |
    "FORMAT" |
    "FORMAT_DESC" |
    "START_DATE" |
    "START_DATE_DESC" |
    "END_DATE" |
    "END_DATE_DESC" |
    "SCORE" |
    "SCORE_DESC" |
    "POPULARITY" |
    "POPULARITY_DESC" |
    "TRENDING" |
    "TRENDING_DESC" |
    "EPISODES" |
    "EPISODES_DESC" |
    "DURATION" |
    "DURATION_DESC" |
    "STATUS" |
    "STATUS_DESC" |
    "CHAPTERS" |
    "CHAPTERS_DESC" |
    "VOLUMES" |
    "VOLUMES_DESC" |
    "UPDATED_AT" |
    "UPDATED_AT_DESC" |
    "SEARCH_MATCH" |
    "FAVOURITES" |
    "FAVOURITES_DESC";

    /**
     * - Filepath: internal/api/anilist/models_gen.go
     * @description
     *  The current releasing status of the media
     */
    export type AL_MediaStatus = "FINISHED" | "RELEASING" | "NOT_YET_RELEASED" | "CANCELLED" | "HIATUS";

    /**
     * - Filepath: internal/api/anilist/models_gen.go
     * @description
     *  Media type enum, anime or manga.
     */
    export type AL_MediaType = "ANIME" | "MANGA";

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_StudioDetails {
        Studio?: AL_StudioDetails_Studio;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_StudioDetails_Studio {
        id: number;
        isAnimationStudio: boolean;
        name: string;
        media?: AL_StudioDetails_Studio_Media;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_StudioDetails_Studio_Media {
        nodes?: Array<AL_BaseAnime>;
    }

    /**
     * - Filepath: internal/api/animap/animap.go
     */
    interface Animap_Anime {
        title: string;
        titles?: Record<string, string>;
        /**
         * YYYY-MM-DD
         */
        startDate?: string;
        /**
         * YYYY-MM-DD
         */
        endDate?: string;
        /**
         * Finished, Airing, Upcoming, etc.
         */
        status: string;
        /**
         * TV, OVA, Movie, etc.
         */
        type: string;
        /**
         * Indexed by AniDB episode number, "1", "S1", etc.
         */
        episodes?: Record<string, Animap_Episode>;
        mappings?: Animap_AnimeMapping;
    }

    /**
     * - Filepath: internal/api/animap/animap.go
     */
    interface Animap_AnimeMapping {
        anidb_id?: number;
        anilist_id?: number;
        kitsu_id?: number;
        thetvdb_id?: number;
        /**
         * Can be int or string, forced to string
         */
        themoviedb_id?: string;
        mal_id?: number;
        livechart_id?: number;
        /**
         * Can be int or string, forced to string
         */
        anime
        -
        planet_id?: string;
        anisearch_id?: number;
        simkl_id?: number;
        notify
        .
        moe_id?: string;
        animecountdown_id?: number;
        type?: string;
    }

    /**
     * - Filepath: internal/api/animap/animap.go
     */
    interface Animap_Episode {
        anidbEpisode: string;
        anidbEid: number;
        tvdbEid?: number;
        tvdbShowId?: number;
        /**
         * YYYY-MM-DD
         */
        airDate?: string;
        /**
         * Title of the episode from AniDB
         */
        anidbTitle?: string;
        /**
         * Title of the episode from TVDB
         */
        tvdbTitle?: string;
        overview?: string;
        image?: string;
        /**
         * minutes
         */
        runtime?: number;
        /**
         * Xm
         */
        length?: string;
        seasonNumber?: number;
        seasonName?: string;
        number: number;
        absoluteNumber?: number;
    }

    /**
     * - Filepath: internal/library/anime/autodownloader_rule.go
     */
    interface Anime_AutoDownloaderRule {
        /**
         * Will be set when fetched from the database
         */
        dbId: number;
        enabled: boolean;
        mediaId: number;
        releaseGroups?: Array<string>;
        resolutions?: Array<string>;
        comparisonTitle: string;
        titleComparisonType: Anime_AutoDownloaderRuleTitleComparisonType;
        episodeType: Anime_AutoDownloaderRuleEpisodeType;
        episodeNumbers?: Array<number>;
        destination: string;
        additionalTerms?: Array<string>;
    }

    /**
     * - Filepath: internal/library/anime/autodownloader_rule.go
     */
    export type Anime_AutoDownloaderRuleEpisodeType = "recent" | "selected";

    /**
     * - Filepath: internal/library/anime/autodownloader_rule.go
     */
    export type Anime_AutoDownloaderRuleTitleComparisonType = "contains" | "likely";

    /**
     * - Filepath: internal/library/anime/entry.go
     */
    interface Anime_Entry {
        mediaId: number;
        media?: AL_BaseAnime;
        listData?: Anime_EntryListData;
        libraryData?: Anime_EntryLibraryData;
        downloadInfo?: Anime_EntryDownloadInfo;
        episodes?: Array<Anime_Episode>;
        nextEpisode?: Anime_Episode;
        localFiles?: Array<Anime_LocalFile>;
        anidbId: number;
        currentEpisodeCount: number;
        _isNakamaEntry: boolean;
        nakamaLibraryData?: Anime_NakamaEntryLibraryData;
    }

    /**
     * - Filepath: internal/library/anime/entry_download_info.go
     */
    interface Anime_EntryDownloadEpisode {
        episodeNumber: number;
        aniDBEpisode: string;
        episode?: Anime_Episode;
    }

    /**
     * - Filepath: internal/library/anime/entry_download_info.go
     */
    interface Anime_EntryDownloadInfo {
        episodesToDownload?: Array<Anime_EntryDownloadEpisode>;
        canBatch: boolean;
        batchAll: boolean;
        hasInaccurateSchedule: boolean;
        rewatch: boolean;
        absoluteOffset: number;
    }

    /**
     * - Filepath: internal/library/anime/entry_library_data.go
     */
    interface Anime_EntryLibraryData {
        allFilesLocked: boolean;
        sharedPath: string;
        unwatchedCount: number;
        mainFileCount: number;
    }

    /**
     * - Filepath: internal/library/anime/entry.go
     */
    interface Anime_EntryListData {
        progress?: number;
        score?: number;
        status?: AL_MediaListStatus;
        repeat?: number;
        startedAt?: string;
        completedAt?: string;
    }

    /**
     * - Filepath: internal/library/anime/episode.go
     */
    interface Anime_Episode {
        type: Anime_LocalFileType;
        /**
         * e.g, Show: "Episode 1", Movie: "Violet Evergarden The Movie"
         */
        displayTitle: string;
        /**
         * e.g, "Shibuya Incident - Gate, Open"
         */
        episodeTitle: string;
        episodeNumber: number;
        /**
         * AniDB episode number
         */
        aniDBEpisode?: string;
        absoluteEpisodeNumber: number;
        /**
         * Usually the same as EpisodeNumber, unless there is a discrepancy between AniList and AniDB
         */
        progressNumber: number;
        localFile?: Anime_LocalFile;
        /**
         * Is in the local files
         */
        isDownloaded: boolean;
        /**
         * (image, airDate, length, summary, overview)
         */
        episodeMetadata?: Anime_EpisodeMetadata;
        /**
         * (episode, aniDBEpisode, type...)
         */
        fileMetadata?: Anime_LocalFileMetadata;
        /**
         * No AniDB data
         */
        isInvalid: boolean;
        /**
         * Alerts the user that there is a discrepancy between AniList and AniDB
         */
        metadataIssue?: string;
        baseAnime?: AL_BaseAnime;
        _isNakamaEpisode: boolean;
    }

    /**
     * - Filepath: internal/library/anime/episode_collection.go
     */
    interface Anime_EpisodeCollection {
        hasMappingError: boolean;
        episodes?: Array<Anime_Episode>;
        metadata?: Metadata_AnimeMetadata;
    }

    /**
     * - Filepath: internal/library/anime/episode.go
     */
    interface Anime_EpisodeMetadata {
        anidbId?: number;
        image?: string;
        airDate?: string;
        length?: number;
        summary?: string;
        overview?: string;
        isFiller?: boolean;
        /**
         * Indicates if the episode has a real image
         */
        hasImage?: boolean;
    }

    /**
     * - Filepath: internal/library/anime/collection.go
     */
    interface Anime_LibraryCollection {
        continueWatchingList?: Array<Anime_Episode>;
        lists?: Array<Anime_LibraryCollectionList>;
        unmatchedLocalFiles?: Array<Anime_LocalFile>;
        unmatchedGroups?: Array<Anime_UnmatchedGroup>;
        ignoredLocalFiles?: Array<Anime_LocalFile>;
        unknownGroups?: Array<Anime_UnknownGroup>;
        stats?: Anime_LibraryCollectionStats;
        /**
         * Hydrated by the route handler
         */
        stream?: Anime_StreamCollection;
    }

    /**
     * - Filepath: internal/library/anime/collection.go
     */
    interface Anime_LibraryCollectionEntry {
        media?: AL_BaseAnime;
        mediaId: number;
        /**
         * Library data
         */
        libraryData?: Anime_EntryLibraryData;
        /**
         * Library data from Nakama
         */
        nakamaLibraryData?: Anime_NakamaEntryLibraryData;
        /**
         * AniList list data
         */
        listData?: Anime_EntryListData;
    }

    /**
     * - Filepath: internal/library/anime/collection.go
     */
    interface Anime_LibraryCollectionList {
        type?: AL_MediaListStatus;
        status?: AL_MediaListStatus;
        entries?: Array<Anime_LibraryCollectionEntry>;
    }

    /**
     * - Filepath: internal/library/anime/collection.go
     */
    interface Anime_LibraryCollectionStats {
        totalEntries: number;
        totalFiles: number;
        totalShows: number;
        totalMovies: number;
        totalSpecials: number;
        totalSize: string;
    }

    /**
     * - Filepath: internal/library/anime/localfile.go
     */
    interface Anime_LocalFile {
        path: string;
        name: string;
        parsedInfo?: Anime_LocalFileParsedData;
        parsedFolderInfo?: Array<Anime_LocalFileParsedData>;
        metadata?: Anime_LocalFileMetadata;
        locked: boolean;
        /**
         * Unused for now
         */
        ignored: boolean;
        mediaId: number;
    }

    /**
     * - Filepath: internal/library/anime/localfile.go
     */
    interface Anime_LocalFileMetadata {
        episode: number;
        aniDBEpisode: string;
        type: Anime_LocalFileType;
    }

    /**
     * - Filepath: internal/library/anime/localfile.go
     */
    interface Anime_LocalFileParsedData {
        original: string;
        title?: string;
        releaseGroup?: string;
        season?: string;
        seasonRange?: Array<string>;
        part?: string;
        partRange?: Array<string>;
        episode?: string;
        episodeRange?: Array<string>;
        episodeTitle?: string;
        year?: string;
    }

    /**
     * - Filepath: internal/library/anime/localfile.go
     */
    export type Anime_LocalFileType = "main" | "special" | "nc";

    /**
     * - Filepath: internal/library/anime/localfile_wrapper.go
     */
    interface Anime_LocalFileWrapperEntry {
        mediaId: number;
        localFiles?: Array<Anime_LocalFile>;
    }

    /**
     * - Filepath: internal/library/anime/missing_episodes.go
     */
    interface Anime_MissingEpisodes {
        episodes?: Array<Anime_Episode>;
        silencedEpisodes?: Array<Anime_Episode>;
    }

    /**
     * - Filepath: internal/library/anime/entry_library_data.go
     */
    interface Anime_NakamaEntryLibraryData {
        unwatchedCount: number;
        mainFileCount: number;
    }

    /**
     * - Filepath: internal/library/anime/normalized_media.go
     */
    interface Anime_NormalizedMedia {
        id: number;
        idMal?: number;
        siteUrl?: string;
        status?: AL_MediaStatus;
        season?: AL_MediaSeason;
        type?: AL_MediaType;
        format?: AL_MediaFormat;
        seasonYear?: number;
        bannerImage?: string;
        episodes?: number;
        synonyms?: Array<string>;
        isAdult?: boolean;
        countryOfOrigin?: string;
        meanScore?: number;
        description?: string;
        genres?: Array<string>;
        duration?: number;
        trailer?: AL_BaseAnime_Trailer;
        title?: AL_BaseAnime_Title;
        coverImage?: AL_BaseAnime_CoverImage;
        startDate?: AL_BaseAnime_StartDate;
        endDate?: AL_BaseAnime_EndDate;
        nextAiringEpisode?: AL_BaseAnime_NextAiringEpisode;
    }

    /**
     * - Filepath: internal/library/anime/collection.go
     */
    interface Anime_StreamCollection {
        continueWatchingList?: Array<Anime_Episode>;
        anime?: Array<AL_BaseAnime>;
        listData?: Record<number, Anime_EntryListData>;
    }

    /**
     * - Filepath: internal/library/anime/collection.go
     */
    interface Anime_UnknownGroup {
        mediaId: number;
        localFiles?: Array<Anime_LocalFile>;
    }

    /**
     * - Filepath: internal/library/anime/collection.go
     */
    interface Anime_UnmatchedGroup {
        dir: string;
        localFiles?: Array<Anime_LocalFile>;
        suggestions?: Array<AL_BaseAnime>;
    }

    /**
     * - Filepath: internal/api/anizip/anizip.go
     */
    interface Anizip_Episode {
        tvdbEid?: number;
        airdate?: string;
        seasonNumber?: number;
        episodeNumber?: number;
        absoluteEpisodeNumber?: number;
        title?: Record<string, string>;
        image?: string;
        summary?: string;
        overview?: string;
        runtime?: number;
        length?: number;
        episode?: string;
        anidbEid?: number;
        rating?: string;
    }

    /**
     * - Filepath: internal/api/anizip/anizip.go
     */
    interface Anizip_Mappings {
        animeplanet_id?: string;
        kitsu_id?: number;
        mal_id?: number;
        type?: string;
        anilist_id?: number;
        anisearch_id?: number;
        anidb_id?: number;
        notifymoe_id?: string;
        livechart_id?: number;
        thetvdb_id?: number;
        imdb_id?: string;
        themoviedb_id?: string;
    }

    /**
     * - Filepath: internal/api/anizip/anizip.go
     */
    interface Anizip_Media {
        titles?: Record<string, string>;
        episodes?: Record<string, Anizip_Episode>;
        episodeCount: number;
        specialCount: number;
        mappings?: Anizip_Mappings;
    }

    /**
     * - Filepath: internal/library/autodownloader/autodownloader_torrent.go
     */
    interface AutoDownloader_NormalizedTorrent {
        parsedData?: $habari.Metadata;
        /**
         * Access using GetMagnet()
         */
        magnet: string;
        provider?: string;
        name: string;
        date: string;
        size: number;
        formattedSize: string;
        seeders: number;
        leechers: number;
        downloadCount: number;
        link: string;
        downloadUrl: string;
        magnetLink?: string;
        infoHash?: string;
        resolution?: string;
        isBatch?: boolean;
        episodeNumber?: number;
        releaseGroup?: string;
        isBestRelease: boolean;
        confirmed: boolean;
    }

    /**
     * - Filepath: internal/continuity/manager.go
     */
    export type Continuity_Kind = "onlinestream" | "mediastream" | "external_player";

    /**
     * - Filepath: internal/continuity/history.go
     */
    interface Continuity_UpdateWatchHistoryItemOptions {
        currentTime: number;
        duration: number;
        mediaId: number;
        episodeNumber: number;
        filepath?: string;
        kind: Continuity_Kind;
    }

    /**
     * - Filepath: internal/continuity/history.go
     */
    export type Continuity_WatchHistory = Record<number, Continuity_WatchHistoryItem>;

    /**
     * - Filepath: internal/continuity/history.go
     */
    interface Continuity_WatchHistoryItem {
        kind: Continuity_Kind;
        filepath: string;
        mediaId: number;
        episodeNumber: number;
        currentTime: number;
        duration: number;
        timeAdded?: string;
        timeUpdated?: string;
    }

    /**
     * - Filepath: internal/continuity/history.go
     */
    interface Continuity_WatchHistoryItemResponse {
        item?: Continuity_WatchHistoryItem;
        found: boolean;
    }

    /**
     * - Filepath: internal/discordrpc/presence/presence.go
     */
    interface DiscordRPC_AnimeActivity {
        id: number;
        title: string;
        image: string;
        isMovie: boolean;
        episodeNumber: number;
        paused: boolean;
        progress: number;
        duration: number;
        totalEpisodes?: number;
        currentEpisodeCount?: number;
        episodeTitle?: string;
    }

    /**
     * - Filepath: internal/discordrpc/client/activity.go
     */
    interface DiscordRPC_Button {
        label?: string;
        url?: string;
    }

    /**
     * - Filepath: internal/discordrpc/presence/presence.go
     */
    interface DiscordRPC_LegacyAnimeActivity {
        id: number;
        title: string;
        image: string;
        isMovie: boolean;
        episodeNumber: number;
    }

    /**
     * - Filepath: internal/discordrpc/presence/presence.go
     */
    interface DiscordRPC_MangaActivity {
        id: number;
        title: string;
        image: string;
        chapter: string;
    }

    /**
     * - Filepath: internal/extension/hibike/manga/types.go
     */
    interface HibikeManga_ChapterDetails {
        provider: string;
        id: string;
        url: string;
        title: string;
        chapter: string;
        index: number;
        scanlator?: string;
        language?: string;
        rating?: number;
        updatedAt?: string;
        localIsPDF?: boolean;
    }

    /**
     * - Filepath: internal/extension/hibike/torrent/types.go
     */
    interface HibikeTorrent_AnimeTorrent {
        provider?: string;
        name: string;
        date: string;
        size: number;
        formattedSize: string;
        seeders: number;
        leechers: number;
        downloadCount: number;
        link: string;
        downloadUrl: string;
        magnetLink?: string;
        infoHash?: string;
        resolution?: string;
        isBatch?: boolean;
        episodeNumber?: number;
        releaseGroup?: string;
        isBestRelease: boolean;
        confirmed: boolean;
    }

    /**
     * - Filepath: internal/manga/chapter_container.go
     */
    interface Manga_ChapterContainer {
        mediaId: number;
        provider: string;
        chapters?: Array<HibikeManga_ChapterDetails>;
    }

    /**
     * - Filepath: internal/manga/collection.go
     */
    interface Manga_Collection {
        lists?: Array<Manga_CollectionList>;
    }

    /**
     * - Filepath: internal/manga/collection.go
     */
    interface Manga_CollectionEntry {
        media?: AL_BaseManga;
        mediaId: number;
        /**
         * AniList list data
         */
        listData?: Manga_EntryListData;
    }

    /**
     * - Filepath: internal/manga/collection.go
     */
    interface Manga_CollectionList {
        type?: AL_MediaListStatus;
        status?: AL_MediaListStatus;
        entries?: Array<Manga_CollectionEntry>;
    }

    /**
     * - Filepath: internal/manga/manga_entry.go
     */
    interface Manga_Entry {
        mediaId: number;
        media?: AL_BaseManga;
        listData?: Manga_EntryListData;
    }

    /**
     * - Filepath: internal/manga/manga_entry.go
     */
    interface Manga_EntryListData {
        progress?: number;
        score?: number;
        status?: AL_MediaListStatus;
        repeat?: number;
        startedAt?: string;
        completedAt?: string;
    }

    /**
     * - Filepath: internal/manga/chapter_container.go
     */
    interface Manga_MangaLatestChapterNumberItem {
        provider: string;
        scanlator: string;
        language: string;
        number: number;
    }

    /**
     * - Filepath: internal/manga/download.go
     */
    export type Manga_MediaMap = Record<number, Manga_ProviderDownloadMap>;

    /**
     * - Filepath: internal/manga/download.go
     */
    export type Manga_ProviderDownloadMap = Record<string, Array<Manga_ProviderDownloadMapChapterInfo>>;

    /**
     * - Filepath: internal/manga/download.go
     */
    interface Manga_ProviderDownloadMapChapterInfo {
        chapterId: string;
        chapterNumber: string;
    }

    /**
     * - Filepath: internal/api/metadata/types.go
     */
    interface Metadata_AnimeMappings {
        animeplanetId: string;
        kitsuId: number;
        malId: number;
        type: string;
        anilistId: number;
        anisearchId: number;
        anidbId: number;
        notifymoeId: string;
        livechartId: number;
        thetvdbId: number;
        imdbId: string;
        themoviedbId: string;
    }

    /**
     * - Filepath: internal/api/metadata/types.go
     */
    interface Metadata_AnimeMetadata {
        titles?: Record<string, string>;
        episodes?: Record<string, Metadata_EpisodeMetadata>;
        episodeCount: number;
        specialCount: number;
        mappings?: Metadata_AnimeMappings;
    }

    /**
     * - Filepath: internal/api/metadata/types.go
     */
    interface Metadata_EpisodeMetadata {
        anidbId: number;
        tvdbId: number;
        title: string;
        image: string;
        airDate: string;
        length: number;
        summary: string;
        overview: string;
        episodeNumber: number;
        episode: string;
        seasonNumber: number;
        absoluteEpisodeNumber: number;
        anidbEid: number;
        /**
         * Indicates if the episode has a real image
         */
        hasImage: boolean;
    }

    /**
     * - Filepath: internal/database/models/models.go
     */
    interface Models_AutoDownloaderItem {
        ruleId: number;
        mediaId: number;
        episode: number;
        link: string;
        hash: string;
        magnet: string;
        torrentName: string;
        downloaded: boolean;
        id: number;
        createdAt?: string;
        updatedAt?: string;
    }

    /**
     * - Filepath: internal/database/models/models.go
     */
    interface Models_AutoDownloaderSettings {
        provider: string;
        interval: number;
        enabled: boolean;
        downloadAutomatically: boolean;
        enableEnhancedQueries: boolean;
        enableSeasonCheck: boolean;
        useDebrid: boolean;
    }

    /**
     * - Filepath: internal/onlinestream/repository.go
     */
    interface Onlinestream_Episode {
        number: number;
        title?: string;
        image?: string;
        description?: string;
        isFiller?: boolean;
    }

    /**
     * - Filepath: internal/torrent_clients/torrent_client/torrent.go
     */
    interface TorrentClient_Torrent {
        name: string;
        hash: string;
        seeds: number;
        upSpeed: string;
        downSpeed: string;
        progress: number;
        size: string;
        eta: string;
        status: TorrentClient_TorrentStatus;
        contentPath: string;
    }

    /**
     * - Filepath: internal/torrent_clients/torrent_client/torrent.go
     */
    export type TorrentClient_TorrentStatus = "downloading" | "seeding" | "paused" | "other" | "stopped";

}
