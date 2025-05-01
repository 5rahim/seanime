declare namespace $app {

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
     * AnimeLibraryCollectionRequestedEvent is triggered when the user requests the library collection.
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
     * @package handlers
     */

    /**
     * @event HandleGetAnimeCollectionRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetAnimeCollectionRequestedEvent is triggered when GetAnimeCollection is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetAnimeCollectionRequested(cb: (event: HandleGetAnimeCollectionRequestedEvent) => void): void;

    interface HandleGetAnimeCollectionRequestedEvent {
        data?: AL_AnimeCollection;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetAnimeCollectionEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetAnimeCollectionEvent is triggered after processing GetAnimeCollection.
     */
    function onHandleGetAnimeCollection(cb: (event: HandleGetAnimeCollectionEvent) => void): void;

    interface HandleGetAnimeCollectionEvent {
        data?: AL_AnimeCollection;

        next(): void;
    }

    /**
     * @event HandleGetRawAnimeCollectionRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetRawAnimeCollectionRequestedEvent is triggered when GetRawAnimeCollection is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetRawAnimeCollectionRequested(cb: (event: HandleGetRawAnimeCollectionRequestedEvent) => void): void;

    interface HandleGetRawAnimeCollectionRequestedEvent {
        data?: AL_AnimeCollection;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetRawAnimeCollectionEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetRawAnimeCollectionEvent is triggered after processing GetRawAnimeCollection.
     */
    function onHandleGetRawAnimeCollection(cb: (event: HandleGetRawAnimeCollectionEvent) => void): void;

    interface HandleGetRawAnimeCollectionEvent {
        data?: AL_AnimeCollection;

        next(): void;
    }

    /**
     * @event HandleEditAnilistListEntryRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleEditAnilistListEntryRequestedEvent is triggered when EditAnilistListEntry is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleEditAnilistListEntryRequested(cb: (event: HandleEditAnilistListEntryRequestedEvent) => void): void;

    interface HandleEditAnilistListEntryRequestedEvent {
        mediaId: number;
        status?: AL_MediaListStatus;
        score: number;
        progress: number;
        startedAt?: AL_FuzzyDateInput;
        completedAt?: AL_FuzzyDateInput;
        type: string;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetAnilistAnimeDetailsRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetAnilistAnimeDetailsRequestedEvent is triggered when GetAnilistAnimeDetails is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetAnilistAnimeDetailsRequested(cb: (event: HandleGetAnilistAnimeDetailsRequestedEvent) => void): void;

    interface HandleGetAnilistAnimeDetailsRequestedEvent {
        id: number;
        data?: AL_AnimeDetailsById_Media;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetAnilistAnimeDetailsEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetAnilistAnimeDetailsEvent is triggered after processing GetAnilistAnimeDetails.
     */
    function onHandleGetAnilistAnimeDetails(cb: (event: HandleGetAnilistAnimeDetailsEvent) => void): void;

    interface HandleGetAnilistAnimeDetailsEvent {
        data?: AL_AnimeDetailsById_Media;

        next(): void;
    }

    /**
     * @event HandleGetAnilistStudioDetailsRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetAnilistStudioDetailsRequestedEvent is triggered when GetAnilistStudioDetails is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetAnilistStudioDetailsRequested(cb: (event: HandleGetAnilistStudioDetailsRequestedEvent) => void): void;

    interface HandleGetAnilistStudioDetailsRequestedEvent {
        id: number;
        data?: AL_StudioDetails;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetAnilistStudioDetailsEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetAnilistStudioDetailsEvent is triggered after processing GetAnilistStudioDetails.
     */
    function onHandleGetAnilistStudioDetails(cb: (event: HandleGetAnilistStudioDetailsEvent) => void): void;

    interface HandleGetAnilistStudioDetailsEvent {
        data?: AL_StudioDetails;

        next(): void;
    }

    /**
     * @event HandleDeleteAnilistListEntryRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleDeleteAnilistListEntryRequestedEvent is triggered when DeleteAnilistListEntry is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleDeleteAnilistListEntryRequested(cb: (event: HandleDeleteAnilistListEntryRequestedEvent) => void): void;

    interface HandleDeleteAnilistListEntryRequestedEvent {
        mediaId: number;
        type: string;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleAnilistListAnimeRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleAnilistListAnimeRequestedEvent is triggered when AnilistListAnime is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleAnilistListAnimeRequested(cb: (event: HandleAnilistListAnimeRequestedEvent) => void): void;

    interface HandleAnilistListAnimeRequestedEvent {
        page: number;
        search: string;
        perPage: number;
        sort?: Array<MediaSort>;
        status?: Array<MediaStatus>;
        genres?: Array<string>;
        averageScore_greater: number;
        season?: AL_MediaSeason;
        seasonYear: number;
        format?: AL_MediaFormat;
        isAdult: boolean;
        data?: AL_ListAnime;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleAnilistListAnimeEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleAnilistListAnimeEvent is triggered after processing AnilistListAnime.
     */
    function onHandleAnilistListAnime(cb: (event: HandleAnilistListAnimeEvent) => void): void;

    interface HandleAnilistListAnimeEvent {
        data?: AL_ListAnime;

        next(): void;
    }

    /**
     * @event HandleAnilistListRecentAiringAnimeRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleAnilistListRecentAiringAnimeRequestedEvent is triggered when AnilistListRecentAiringAnime is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleAnilistListRecentAiringAnimeRequested(cb: (event: HandleAnilistListRecentAiringAnimeRequestedEvent) => void): void;

    interface HandleAnilistListRecentAiringAnimeRequestedEvent {
        page: number;
        search: string;
        perPage: number;
        airingAt_greater: number;
        airingAt_lesser: number;
        notYetAired: boolean;
        sort?: Array<AiringSort>;
        data?: AL_ListRecentAnime;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleAnilistListRecentAiringAnimeEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleAnilistListRecentAiringAnimeEvent is triggered after processing AnilistListRecentAiringAnime.
     */
    function onHandleAnilistListRecentAiringAnime(cb: (event: HandleAnilistListRecentAiringAnimeEvent) => void): void;

    interface HandleAnilistListRecentAiringAnimeEvent {
        data?: AL_ListRecentAnime;

        next(): void;
    }

    /**
     * @event HandleAnilistListMissedSequelsRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleAnilistListMissedSequelsRequestedEvent is triggered when AnilistListMissedSequels is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleAnilistListMissedSequelsRequested(cb: (event: HandleAnilistListMissedSequelsRequestedEvent) => void): void;

    interface HandleAnilistListMissedSequelsRequestedEvent {
        data?: AL_BaseAnime;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleAnilistListMissedSequelsEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleAnilistListMissedSequelsEvent is triggered after processing AnilistListMissedSequels.
     */
    function onHandleAnilistListMissedSequels(cb: (event: HandleAnilistListMissedSequelsEvent) => void): void;

    interface HandleAnilistListMissedSequelsEvent {
        data?: AL_BaseAnime;

        next(): void;
    }

    /**
     * @event HandleGetAniListStatsRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetAniListStatsRequestedEvent is triggered when GetAniListStats is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetAniListStatsRequested(cb: (event: HandleGetAniListStatsRequestedEvent) => void): void;

    interface HandleGetAniListStatsRequestedEvent {
        data?: AL_Stats;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetAniListStatsEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetAniListStatsEvent is triggered after processing GetAniListStats.
     */
    function onHandleGetAniListStats(cb: (event: HandleGetAniListStatsEvent) => void): void;

    interface HandleGetAniListStatsEvent {
        data?: AL_Stats;

        next(): void;
    }

    /**
     * @event HandleGetLibraryCollectionRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetLibraryCollectionRequestedEvent is triggered when GetLibraryCollection is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetLibraryCollectionRequested(cb: (event: HandleGetLibraryCollectionRequestedEvent) => void): void;

    interface HandleGetLibraryCollectionRequestedEvent {
        data?: Anime_LibraryCollection;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetLibraryCollectionEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetLibraryCollectionEvent is triggered after processing GetLibraryCollection.
     */
    function onHandleGetLibraryCollection(cb: (event: HandleGetLibraryCollectionEvent) => void): void;

    interface HandleGetLibraryCollectionEvent {
        data?: Anime_LibraryCollection;

        next(): void;
    }

    /**
     * @event HandleAddUnknownMediaRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleAddUnknownMediaRequestedEvent is triggered when AddUnknownMedia is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleAddUnknownMediaRequested(cb: (event: HandleAddUnknownMediaRequestedEvent) => void): void;

    interface HandleAddUnknownMediaRequestedEvent {
        mediaIds?: Array<number>;
        data?: AL_AnimeCollection;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleAddUnknownMediaEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleAddUnknownMediaEvent is triggered after processing AddUnknownMedia.
     */
    function onHandleAddUnknownMedia(cb: (event: HandleAddUnknownMediaEvent) => void): void;

    interface HandleAddUnknownMediaEvent {
        data?: AL_AnimeCollection;

        next(): void;
    }

    /**
     * @event HandleGetAnimeEntryRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetAnimeEntryRequestedEvent is triggered when GetAnimeEntry is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetAnimeEntryRequested(cb: (event: HandleGetAnimeEntryRequestedEvent) => void): void;

    interface HandleGetAnimeEntryRequestedEvent {
        id: number;
        data?: Anime_Entry;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetAnimeEntryEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetAnimeEntryEvent is triggered after processing GetAnimeEntry.
     */
    function onHandleGetAnimeEntry(cb: (event: HandleGetAnimeEntryEvent) => void): void;

    interface HandleGetAnimeEntryEvent {
        data?: Anime_Entry;

        next(): void;
    }

    /**
     * @event HandleAnimeEntryBulkActionRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleAnimeEntryBulkActionRequestedEvent is triggered when AnimeEntryBulkAction is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleAnimeEntryBulkActionRequested(cb: (event: HandleAnimeEntryBulkActionRequestedEvent) => void): void;

    interface HandleAnimeEntryBulkActionRequestedEvent {
        mediaId: number;
        action: string;
        data?: Anime_LocalFile;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleAnimeEntryBulkActionEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleAnimeEntryBulkActionEvent is triggered after processing AnimeEntryBulkAction.
     */
    function onHandleAnimeEntryBulkAction(cb: (event: HandleAnimeEntryBulkActionEvent) => void): void;

    interface HandleAnimeEntryBulkActionEvent {
        data?: Anime_LocalFile;

        next(): void;
    }

    /**
     * @event HandleOpenAnimeEntryInExplorerRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleOpenAnimeEntryInExplorerRequestedEvent is triggered when OpenAnimeEntryInExplorer is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleOpenAnimeEntryInExplorerRequested(cb: (event: HandleOpenAnimeEntryInExplorerRequestedEvent) => void): void;

    interface HandleOpenAnimeEntryInExplorerRequestedEvent {
        mediaId: number;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleFetchAnimeEntrySuggestionsRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleFetchAnimeEntrySuggestionsRequestedEvent is triggered when FetchAnimeEntrySuggestions is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleFetchAnimeEntrySuggestionsRequested(cb: (event: HandleFetchAnimeEntrySuggestionsRequestedEvent) => void): void;

    interface HandleFetchAnimeEntrySuggestionsRequestedEvent {
        dir: string;
        data?: AL_BaseAnime;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleFetchAnimeEntrySuggestionsEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleFetchAnimeEntrySuggestionsEvent is triggered after processing FetchAnimeEntrySuggestions.
     */
    function onHandleFetchAnimeEntrySuggestions(cb: (event: HandleFetchAnimeEntrySuggestionsEvent) => void): void;

    interface HandleFetchAnimeEntrySuggestionsEvent {
        data?: AL_BaseAnime;

        next(): void;
    }

    /**
     * @event HandleAnimeEntryManualMatchRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleAnimeEntryManualMatchRequestedEvent is triggered when AnimeEntryManualMatch is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleAnimeEntryManualMatchRequested(cb: (event: HandleAnimeEntryManualMatchRequestedEvent) => void): void;

    interface HandleAnimeEntryManualMatchRequestedEvent {
        paths?: Array<string>;
        mediaId: number;
        data?: Anime_LocalFile;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleAnimeEntryManualMatchEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleAnimeEntryManualMatchEvent is triggered after processing AnimeEntryManualMatch.
     */
    function onHandleAnimeEntryManualMatch(cb: (event: HandleAnimeEntryManualMatchEvent) => void): void;

    interface HandleAnimeEntryManualMatchEvent {
        data?: Anime_LocalFile;

        next(): void;
    }

    /**
     * @event HandleGetMissingEpisodesRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetMissingEpisodesRequestedEvent is triggered when GetMissingEpisodes is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetMissingEpisodesRequested(cb: (event: HandleGetMissingEpisodesRequestedEvent) => void): void;

    interface HandleGetMissingEpisodesRequestedEvent {
        data?: Anime_MissingEpisodes;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetMissingEpisodesEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetMissingEpisodesEvent is triggered after processing GetMissingEpisodes.
     */
    function onHandleGetMissingEpisodes(cb: (event: HandleGetMissingEpisodesEvent) => void): void;

    interface HandleGetMissingEpisodesEvent {
        data?: Anime_MissingEpisodes;

        next(): void;
    }

    /**
     * @event HandleGetAnimeEntrySilenceStatusRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetAnimeEntrySilenceStatusRequestedEvent is triggered when GetAnimeEntrySilenceStatus is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetAnimeEntrySilenceStatusRequested(cb: (event: HandleGetAnimeEntrySilenceStatusRequestedEvent) => void): void;

    interface HandleGetAnimeEntrySilenceStatusRequestedEvent {
        id: number;
        data?: Models_SilencedMediaEntry;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetAnimeEntrySilenceStatusEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetAnimeEntrySilenceStatusEvent is triggered after processing GetAnimeEntrySilenceStatus.
     */
    function onHandleGetAnimeEntrySilenceStatus(cb: (event: HandleGetAnimeEntrySilenceStatusEvent) => void): void;

    interface HandleGetAnimeEntrySilenceStatusEvent {
        data?: Models_SilencedMediaEntry;

        next(): void;
    }

    /**
     * @event HandleToggleAnimeEntrySilenceStatusRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleToggleAnimeEntrySilenceStatusRequestedEvent is triggered when ToggleAnimeEntrySilenceStatus is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleToggleAnimeEntrySilenceStatusRequested(cb: (event: HandleToggleAnimeEntrySilenceStatusRequestedEvent) => void): void;

    interface HandleToggleAnimeEntrySilenceStatusRequestedEvent {
        mediaId: number;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleUpdateAnimeEntryProgressRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleUpdateAnimeEntryProgressRequestedEvent is triggered when UpdateAnimeEntryProgress is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleUpdateAnimeEntryProgressRequested(cb: (event: HandleUpdateAnimeEntryProgressRequestedEvent) => void): void;

    interface HandleUpdateAnimeEntryProgressRequestedEvent {
        mediaId: number;
        malId: number;
        episodeNumber: number;
        totalEpisodes: number;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleUpdateAnimeEntryRepeatRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleUpdateAnimeEntryRepeatRequestedEvent is triggered when UpdateAnimeEntryRepeat is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleUpdateAnimeEntryRepeatRequested(cb: (event: HandleUpdateAnimeEntryRepeatRequestedEvent) => void): void;

    interface HandleUpdateAnimeEntryRepeatRequestedEvent {
        mediaId: number;
        repeat: number;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleLoginRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleLoginRequestedEvent is triggered when Login is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleLoginRequested(cb: (event: HandleLoginRequestedEvent) => void): void;

    interface HandleLoginRequestedEvent {
        token: string;
        data?: Status;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleLoginEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleLoginEvent is triggered after processing Login.
     */
    function onHandleLogin(cb: (event: HandleLoginEvent) => void): void;

    interface HandleLoginEvent {
        data?: Status;

        next(): void;
    }

    /**
     * @event HandleLogoutRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleLogoutRequestedEvent is triggered when Logout is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleLogoutRequested(cb: (event: HandleLogoutRequestedEvent) => void): void;

    interface HandleLogoutRequestedEvent {
        data?: Status;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleLogoutEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleLogoutEvent is triggered after processing Logout.
     */
    function onHandleLogout(cb: (event: HandleLogoutEvent) => void): void;

    interface HandleLogoutEvent {
        data?: Status;

        next(): void;
    }

    /**
     * @event HandleRunAutoDownloaderRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleRunAutoDownloaderRequestedEvent is triggered when RunAutoDownloader is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleRunAutoDownloaderRequested(cb: (event: HandleRunAutoDownloaderRequestedEvent) => void): void;

    interface HandleRunAutoDownloaderRequestedEvent {
        next(): void;

        preventDefault(): void;

    }

    /**
     * @event HandleGetAutoDownloaderRuleRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetAutoDownloaderRuleRequestedEvent is triggered when GetAutoDownloaderRule is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetAutoDownloaderRuleRequested(cb: (event: HandleGetAutoDownloaderRuleRequestedEvent) => void): void;

    interface HandleGetAutoDownloaderRuleRequestedEvent {
        id: number;
        data?: Anime_AutoDownloaderRule;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetAutoDownloaderRuleEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetAutoDownloaderRuleEvent is triggered after processing GetAutoDownloaderRule.
     */
    function onHandleGetAutoDownloaderRule(cb: (event: HandleGetAutoDownloaderRuleEvent) => void): void;

    interface HandleGetAutoDownloaderRuleEvent {
        data?: Anime_AutoDownloaderRule;

        next(): void;
    }

    /**
     * @event HandleGetAutoDownloaderRulesByAnimeRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetAutoDownloaderRulesByAnimeRequestedEvent is triggered when GetAutoDownloaderRulesByAnime is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetAutoDownloaderRulesByAnimeRequested(cb: (event: HandleGetAutoDownloaderRulesByAnimeRequestedEvent) => void): void;

    interface HandleGetAutoDownloaderRulesByAnimeRequestedEvent {
        id: number;
        data?: Anime_AutoDownloaderRule;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetAutoDownloaderRulesByAnimeEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetAutoDownloaderRulesByAnimeEvent is triggered after processing GetAutoDownloaderRulesByAnime.
     */
    function onHandleGetAutoDownloaderRulesByAnime(cb: (event: HandleGetAutoDownloaderRulesByAnimeEvent) => void): void;

    interface HandleGetAutoDownloaderRulesByAnimeEvent {
        data?: Anime_AutoDownloaderRule;

        next(): void;
    }

    /**
     * @event HandleGetAutoDownloaderRulesRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetAutoDownloaderRulesRequestedEvent is triggered when GetAutoDownloaderRules is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetAutoDownloaderRulesRequested(cb: (event: HandleGetAutoDownloaderRulesRequestedEvent) => void): void;

    interface HandleGetAutoDownloaderRulesRequestedEvent {
        data?: Anime_AutoDownloaderRule;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetAutoDownloaderRulesEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetAutoDownloaderRulesEvent is triggered after processing GetAutoDownloaderRules.
     */
    function onHandleGetAutoDownloaderRules(cb: (event: HandleGetAutoDownloaderRulesEvent) => void): void;

    interface HandleGetAutoDownloaderRulesEvent {
        data?: Anime_AutoDownloaderRule;

        next(): void;
    }

    /**
     * @event HandleCreateAutoDownloaderRuleRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleCreateAutoDownloaderRuleRequestedEvent is triggered when CreateAutoDownloaderRule is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleCreateAutoDownloaderRuleRequested(cb: (event: HandleCreateAutoDownloaderRuleRequestedEvent) => void): void;

    interface HandleCreateAutoDownloaderRuleRequestedEvent {
        enabled: boolean;
        mediaId: number;
        releaseGroups?: Array<string>;
        resolutions?: Array<string>;
        additionalTerms?: Array<string>;
        comparisonTitle: string;
        titleComparisonType?: Anime_AutoDownloaderRuleTitleComparisonType;
        episodeType?: Anime_AutoDownloaderRuleEpisodeType;
        episodeNumbers?: Array<number>;
        destination: string;
        data?: Anime_AutoDownloaderRule;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleCreateAutoDownloaderRuleEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleCreateAutoDownloaderRuleEvent is triggered after processing CreateAutoDownloaderRule.
     */
    function onHandleCreateAutoDownloaderRule(cb: (event: HandleCreateAutoDownloaderRuleEvent) => void): void;

    interface HandleCreateAutoDownloaderRuleEvent {
        data?: Anime_AutoDownloaderRule;

        next(): void;
    }

    /**
     * @event HandleUpdateAutoDownloaderRuleRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleUpdateAutoDownloaderRuleRequestedEvent is triggered when UpdateAutoDownloaderRule is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleUpdateAutoDownloaderRuleRequested(cb: (event: HandleUpdateAutoDownloaderRuleRequestedEvent) => void): void;

    interface HandleUpdateAutoDownloaderRuleRequestedEvent {
        rule?: Anime_AutoDownloaderRule;
        data?: Anime_AutoDownloaderRule;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleUpdateAutoDownloaderRuleEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleUpdateAutoDownloaderRuleEvent is triggered after processing UpdateAutoDownloaderRule.
     */
    function onHandleUpdateAutoDownloaderRule(cb: (event: HandleUpdateAutoDownloaderRuleEvent) => void): void;

    interface HandleUpdateAutoDownloaderRuleEvent {
        data?: Anime_AutoDownloaderRule;

        next(): void;
    }

    /**
     * @event HandleDeleteAutoDownloaderRuleRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleDeleteAutoDownloaderRuleRequestedEvent is triggered when DeleteAutoDownloaderRule is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleDeleteAutoDownloaderRuleRequested(cb: (event: HandleDeleteAutoDownloaderRuleRequestedEvent) => void): void;

    interface HandleDeleteAutoDownloaderRuleRequestedEvent {
        id: number;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetAutoDownloaderItemsRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetAutoDownloaderItemsRequestedEvent is triggered when GetAutoDownloaderItems is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetAutoDownloaderItemsRequested(cb: (event: HandleGetAutoDownloaderItemsRequestedEvent) => void): void;

    interface HandleGetAutoDownloaderItemsRequestedEvent {
        data?: Models_AutoDownloaderItem;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetAutoDownloaderItemsEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetAutoDownloaderItemsEvent is triggered after processing GetAutoDownloaderItems.
     */
    function onHandleGetAutoDownloaderItems(cb: (event: HandleGetAutoDownloaderItemsEvent) => void): void;

    interface HandleGetAutoDownloaderItemsEvent {
        data?: Models_AutoDownloaderItem;

        next(): void;
    }

    /**
     * @event HandleDeleteAutoDownloaderItemRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleDeleteAutoDownloaderItemRequestedEvent is triggered when DeleteAutoDownloaderItem is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleDeleteAutoDownloaderItemRequested(cb: (event: HandleDeleteAutoDownloaderItemRequestedEvent) => void): void;

    interface HandleDeleteAutoDownloaderItemRequestedEvent {
        id: number;
        id: number;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleUpdateContinuityWatchHistoryItemRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleUpdateContinuityWatchHistoryItemRequestedEvent is triggered when UpdateContinuityWatchHistoryItem is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleUpdateContinuityWatchHistoryItemRequested(cb: (event: HandleUpdateContinuityWatchHistoryItemRequestedEvent) => void): void;

    interface HandleUpdateContinuityWatchHistoryItemRequestedEvent {
        options?: Continuity_UpdateWatchHistoryItemOptions;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetContinuityWatchHistoryItemRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetContinuityWatchHistoryItemRequestedEvent is triggered when GetContinuityWatchHistoryItem is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetContinuityWatchHistoryItemRequested(cb: (event: HandleGetContinuityWatchHistoryItemRequestedEvent) => void): void;

    interface HandleGetContinuityWatchHistoryItemRequestedEvent {
        id: number;
        data?: Continuity_WatchHistoryItemResponse;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetContinuityWatchHistoryItemEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetContinuityWatchHistoryItemEvent is triggered after processing GetContinuityWatchHistoryItem.
     */
    function onHandleGetContinuityWatchHistoryItem(cb: (event: HandleGetContinuityWatchHistoryItemEvent) => void): void;

    interface HandleGetContinuityWatchHistoryItemEvent {
        data?: Continuity_WatchHistoryItemResponse;

        next(): void;
    }

    /**
     * @event HandleGetContinuityWatchHistoryRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetContinuityWatchHistoryRequestedEvent is triggered when GetContinuityWatchHistory is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetContinuityWatchHistoryRequested(cb: (event: HandleGetContinuityWatchHistoryRequestedEvent) => void): void;

    interface HandleGetContinuityWatchHistoryRequestedEvent {
        data?: Continuity_WatchHistory;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetContinuityWatchHistoryEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetContinuityWatchHistoryEvent is triggered after processing GetContinuityWatchHistory.
     */
    function onHandleGetContinuityWatchHistory(cb: (event: HandleGetContinuityWatchHistoryEvent) => void): void;

    interface HandleGetContinuityWatchHistoryEvent {
        data?: Continuity_WatchHistory;

        next(): void;
    }

    /**
     * @event HandleGetDebridSettingsRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetDebridSettingsRequestedEvent is triggered when GetDebridSettings is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetDebridSettingsRequested(cb: (event: HandleGetDebridSettingsRequestedEvent) => void): void;

    interface HandleGetDebridSettingsRequestedEvent {
        data?: Models_DebridSettings;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetDebridSettingsEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetDebridSettingsEvent is triggered after processing GetDebridSettings.
     */
    function onHandleGetDebridSettings(cb: (event: HandleGetDebridSettingsEvent) => void): void;

    interface HandleGetDebridSettingsEvent {
        data?: Models_DebridSettings;

        next(): void;
    }

    /**
     * @event HandleSaveDebridSettingsRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleSaveDebridSettingsRequestedEvent is triggered when SaveDebridSettings is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleSaveDebridSettingsRequested(cb: (event: HandleSaveDebridSettingsRequestedEvent) => void): void;

    interface HandleSaveDebridSettingsRequestedEvent {
        settings?: Models_DebridSettings;
        data?: Models_DebridSettings;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleSaveDebridSettingsEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleSaveDebridSettingsEvent is triggered after processing SaveDebridSettings.
     */
    function onHandleSaveDebridSettings(cb: (event: HandleSaveDebridSettingsEvent) => void): void;

    interface HandleSaveDebridSettingsEvent {
        data?: Models_DebridSettings;

        next(): void;
    }

    /**
     * @event HandleDebridAddTorrentsRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleDebridAddTorrentsRequestedEvent is triggered when DebridAddTorrents is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleDebridAddTorrentsRequested(cb: (event: HandleDebridAddTorrentsRequestedEvent) => void): void;

    interface HandleDebridAddTorrentsRequestedEvent {
        torrents?: Array<AnimeTorrent>;
        media?: AL_BaseAnime;
        destination: string;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleDebridDownloadTorrentRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleDebridDownloadTorrentRequestedEvent is triggered when DebridDownloadTorrent is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleDebridDownloadTorrentRequested(cb: (event: HandleDebridDownloadTorrentRequestedEvent) => void): void;

    interface HandleDebridDownloadTorrentRequestedEvent {
        torrentItem?: Debrid_TorrentItem;
        destination: string;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleDebridCancelDownloadRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleDebridCancelDownloadRequestedEvent is triggered when DebridCancelDownload is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleDebridCancelDownloadRequested(cb: (event: HandleDebridCancelDownloadRequestedEvent) => void): void;

    interface HandleDebridCancelDownloadRequestedEvent {
        itemID: string;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleDebridDeleteTorrentRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleDebridDeleteTorrentRequestedEvent is triggered when DebridDeleteTorrent is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleDebridDeleteTorrentRequested(cb: (event: HandleDebridDeleteTorrentRequestedEvent) => void): void;

    interface HandleDebridDeleteTorrentRequestedEvent {
        torrentItem?: Debrid_TorrentItem;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleDebridGetTorrentsRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleDebridGetTorrentsRequestedEvent is triggered when DebridGetTorrents is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleDebridGetTorrentsRequested(cb: (event: HandleDebridGetTorrentsRequestedEvent) => void): void;

    interface HandleDebridGetTorrentsRequestedEvent {
        data?: Debrid_TorrentItem;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleDebridGetTorrentsEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleDebridGetTorrentsEvent is triggered after processing DebridGetTorrents.
     */
    function onHandleDebridGetTorrents(cb: (event: HandleDebridGetTorrentsEvent) => void): void;

    interface HandleDebridGetTorrentsEvent {
        data?: Debrid_TorrentItem;

        next(): void;
    }

    /**
     * @event HandleDebridGetTorrentInfoRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleDebridGetTorrentInfoRequestedEvent is triggered when DebridGetTorrentInfo is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleDebridGetTorrentInfoRequested(cb: (event: HandleDebridGetTorrentInfoRequestedEvent) => void): void;

    interface HandleDebridGetTorrentInfoRequestedEvent {
        torrent?: HibikeTorrent_AnimeTorrent;
        data?: Debrid_TorrentInfo;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleDebridGetTorrentInfoEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleDebridGetTorrentInfoEvent is triggered after processing DebridGetTorrentInfo.
     */
    function onHandleDebridGetTorrentInfo(cb: (event: HandleDebridGetTorrentInfoEvent) => void): void;

    interface HandleDebridGetTorrentInfoEvent {
        data?: Debrid_TorrentInfo;

        next(): void;
    }

    /**
     * @event HandleDebridGetTorrentFilePreviewsRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleDebridGetTorrentFilePreviewsRequestedEvent is triggered when DebridGetTorrentFilePreviews is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleDebridGetTorrentFilePreviewsRequested(cb: (event: HandleDebridGetTorrentFilePreviewsRequestedEvent) => void): void;

    interface HandleDebridGetTorrentFilePreviewsRequestedEvent {
        torrent?: HibikeTorrent_AnimeTorrent;
        episodeNumber: number;
        media?: AL_BaseAnime;
        data?: DebridClient_FilePreview;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleDebridGetTorrentFilePreviewsEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleDebridGetTorrentFilePreviewsEvent is triggered after processing DebridGetTorrentFilePreviews.
     */
    function onHandleDebridGetTorrentFilePreviews(cb: (event: HandleDebridGetTorrentFilePreviewsEvent) => void): void;

    interface HandleDebridGetTorrentFilePreviewsEvent {
        data?: DebridClient_FilePreview;

        next(): void;
    }

    /**
     * @event HandleDebridStartStreamRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleDebridStartStreamRequestedEvent is triggered when DebridStartStream is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleDebridStartStreamRequested(cb: (event: HandleDebridStartStreamRequestedEvent) => void): void;

    interface HandleDebridStartStreamRequestedEvent {
        mediaId: number;
        episodeNumber: number;
        aniDBEpisode: string;
        autoSelect: boolean;
        torrent?: HibikeTorrent_AnimeTorrent;
        fileId: string;
        fileIndex: number;
        playbackType?: DebridClient_StreamPlaybackType;
        clientId: string;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleDebridCancelStreamRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleDebridCancelStreamRequestedEvent is triggered when DebridCancelStream is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleDebridCancelStreamRequested(cb: (event: HandleDebridCancelStreamRequestedEvent) => void): void;

    interface HandleDebridCancelStreamRequestedEvent {
        options?: DebridClient_CancelStreamOptions;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleDirectorySelectorRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleDirectorySelectorRequestedEvent is triggered when DirectorySelector is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleDirectorySelectorRequested(cb: (event: HandleDirectorySelectorRequestedEvent) => void): void;

    interface HandleDirectorySelectorRequestedEvent {
        input: string;
        data?: DirectorySelectorResponse;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleDirectorySelectorEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleDirectorySelectorEvent is triggered after processing DirectorySelector.
     */
    function onHandleDirectorySelector(cb: (event: HandleDirectorySelectorEvent) => void): void;

    interface HandleDirectorySelectorEvent {
        data?: DirectorySelectorResponse;

        next(): void;
    }

    /**
     * @event HandleSetDiscordMangaActivityRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleSetDiscordMangaActivityRequestedEvent is triggered when SetDiscordMangaActivity is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleSetDiscordMangaActivityRequested(cb: (event: HandleSetDiscordMangaActivityRequestedEvent) => void): void;

    interface HandleSetDiscordMangaActivityRequestedEvent {
        mediaId: number;
        title: string;
        image: string;
        chapter: string;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleSetDiscordLegacyAnimeActivityRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleSetDiscordLegacyAnimeActivityRequestedEvent is triggered when SetDiscordLegacyAnimeActivity is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleSetDiscordLegacyAnimeActivityRequested(cb: (event: HandleSetDiscordLegacyAnimeActivityRequestedEvent) => void): void;

    interface HandleSetDiscordLegacyAnimeActivityRequestedEvent {
        mediaId: number;
        title: string;
        image: string;
        isMovie: boolean;
        episodeNumber: number;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleSetDiscordAnimeActivityWithProgressRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleSetDiscordAnimeActivityWithProgressRequestedEvent is triggered when SetDiscordAnimeActivityWithProgress is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleSetDiscordAnimeActivityWithProgressRequested(cb: (event: HandleSetDiscordAnimeActivityWithProgressRequestedEvent) => void): void;

    interface HandleSetDiscordAnimeActivityWithProgressRequestedEvent {
        mediaId: number;
        title: string;
        image: string;
        isMovie: boolean;
        episodeNumber: number;
        progress: number;
        duration: number;
        totalEpisodes: number;
        currentEpisodeCount: number;
        episodeTitle: string;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleUpdateDiscordAnimeActivityWithProgressRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleUpdateDiscordAnimeActivityWithProgressRequestedEvent is triggered when UpdateDiscordAnimeActivityWithProgress is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleUpdateDiscordAnimeActivityWithProgressRequested(cb: (event: HandleUpdateDiscordAnimeActivityWithProgressRequestedEvent) => void): void;

    interface HandleUpdateDiscordAnimeActivityWithProgressRequestedEvent {
        progress: number;
        duration: number;
        paused: boolean;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleCancelDiscordActivityRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleCancelDiscordActivityRequestedEvent is triggered when CancelDiscordActivity is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleCancelDiscordActivityRequested(cb: (event: HandleCancelDiscordActivityRequestedEvent) => void): void;

    interface HandleCancelDiscordActivityRequestedEvent {
        next(): void;

        preventDefault(): void;

    }

    /**
     * @event HandleGetDocsRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetDocsRequestedEvent is triggered when GetDocs is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetDocsRequested(cb: (event: HandleGetDocsRequestedEvent) => void): void;

    interface HandleGetDocsRequestedEvent {
        data?: ApiDocsGroup;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetDocsEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetDocsEvent is triggered after processing GetDocs.
     */
    function onHandleGetDocs(cb: (event: HandleGetDocsEvent) => void): void;

    interface HandleGetDocsEvent {
        data?: ApiDocsGroup;

        next(): void;
    }

    /**
     * @event HandleDownloadTorrentFileRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleDownloadTorrentFileRequestedEvent is triggered when DownloadTorrentFile is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleDownloadTorrentFileRequested(cb: (event: HandleDownloadTorrentFileRequestedEvent) => void): void;

    interface HandleDownloadTorrentFileRequestedEvent {
        download_urls?: Array<string>;
        destination: string;
        media?: AL_BaseAnime;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleDownloadReleaseRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleDownloadReleaseRequestedEvent is triggered when DownloadRelease is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleDownloadReleaseRequested(cb: (event: HandleDownloadReleaseRequestedEvent) => void): void;

    interface HandleDownloadReleaseRequestedEvent {
        download_url: string;
        destination: string;
        data?: DownloadReleaseResponse;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleDownloadReleaseEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleDownloadReleaseEvent is triggered after processing DownloadRelease.
     */
    function onHandleDownloadRelease(cb: (event: HandleDownloadReleaseEvent) => void): void;

    interface HandleDownloadReleaseEvent {
        data?: DownloadReleaseResponse;

        next(): void;
    }

    /**
     * @event HandleOpenInExplorerRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleOpenInExplorerRequestedEvent is triggered when OpenInExplorer is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleOpenInExplorerRequested(cb: (event: HandleOpenInExplorerRequestedEvent) => void): void;

    interface HandleOpenInExplorerRequestedEvent {
        path: string;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleFetchExternalExtensionDataRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleFetchExternalExtensionDataRequestedEvent is triggered when FetchExternalExtensionData is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleFetchExternalExtensionDataRequested(cb: (event: HandleFetchExternalExtensionDataRequestedEvent) => void): void;

    interface HandleFetchExternalExtensionDataRequestedEvent {
        manifestUri: string;
        data?: Extension_Extension;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleFetchExternalExtensionDataEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleFetchExternalExtensionDataEvent is triggered after processing FetchExternalExtensionData.
     */
    function onHandleFetchExternalExtensionData(cb: (event: HandleFetchExternalExtensionDataEvent) => void): void;

    interface HandleFetchExternalExtensionDataEvent {
        data?: Extension_Extension;

        next(): void;
    }

    /**
     * @event HandleInstallExternalExtensionRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleInstallExternalExtensionRequestedEvent is triggered when InstallExternalExtension is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleInstallExternalExtensionRequested(cb: (event: HandleInstallExternalExtensionRequestedEvent) => void): void;

    interface HandleInstallExternalExtensionRequestedEvent {
        manifestUri: string;
        data?: ExtensionRepo_ExtensionInstallResponse;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleInstallExternalExtensionEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleInstallExternalExtensionEvent is triggered after processing InstallExternalExtension.
     */
    function onHandleInstallExternalExtension(cb: (event: HandleInstallExternalExtensionEvent) => void): void;

    interface HandleInstallExternalExtensionEvent {
        data?: ExtensionRepo_ExtensionInstallResponse;

        next(): void;
    }

    /**
     * @event HandleUninstallExternalExtensionRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleUninstallExternalExtensionRequestedEvent is triggered when UninstallExternalExtension is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleUninstallExternalExtensionRequested(cb: (event: HandleUninstallExternalExtensionRequestedEvent) => void): void;

    interface HandleUninstallExternalExtensionRequestedEvent {
        id: string;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleUpdateExtensionCodeRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleUpdateExtensionCodeRequestedEvent is triggered when UpdateExtensionCode is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleUpdateExtensionCodeRequested(cb: (event: HandleUpdateExtensionCodeRequestedEvent) => void): void;

    interface HandleUpdateExtensionCodeRequestedEvent {
        id: string;
        payload: string;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleReloadExternalExtensionsRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleReloadExternalExtensionsRequestedEvent is triggered when ReloadExternalExtensions is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleReloadExternalExtensionsRequested(cb: (event: HandleReloadExternalExtensionsRequestedEvent) => void): void;

    interface HandleReloadExternalExtensionsRequestedEvent {
        next(): void;

        preventDefault(): void;

    }

    /**
     * @event HandleReloadExternalExtensionRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleReloadExternalExtensionRequestedEvent is triggered when ReloadExternalExtension is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleReloadExternalExtensionRequested(cb: (event: HandleReloadExternalExtensionRequestedEvent) => void): void;

    interface HandleReloadExternalExtensionRequestedEvent {
        id: string;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleListExtensionDataRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleListExtensionDataRequestedEvent is triggered when ListExtensionData is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleListExtensionDataRequested(cb: (event: HandleListExtensionDataRequestedEvent) => void): void;

    interface HandleListExtensionDataRequestedEvent {
        data?: Extension_Extension;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleListExtensionDataEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleListExtensionDataEvent is triggered after processing ListExtensionData.
     */
    function onHandleListExtensionData(cb: (event: HandleListExtensionDataEvent) => void): void;

    interface HandleListExtensionDataEvent {
        data?: Extension_Extension;

        next(): void;
    }

    /**
     * @event HandleGetExtensionPayloadRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetExtensionPayloadRequestedEvent is triggered when GetExtensionPayload is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetExtensionPayloadRequested(cb: (event: HandleGetExtensionPayloadRequestedEvent) => void): void;

    interface HandleGetExtensionPayloadRequestedEvent {
        data: string;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetExtensionPayloadEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetExtensionPayloadEvent is triggered after processing GetExtensionPayload.
     */
    function onHandleGetExtensionPayload(cb: (event: HandleGetExtensionPayloadEvent) => void): void;

    interface HandleGetExtensionPayloadEvent {
        data: string;

        next(): void;
    }

    /**
     * @event HandleListDevelopmentModeExtensionsRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleListDevelopmentModeExtensionsRequestedEvent is triggered when ListDevelopmentModeExtensions is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleListDevelopmentModeExtensionsRequested(cb: (event: HandleListDevelopmentModeExtensionsRequestedEvent) => void): void;

    interface HandleListDevelopmentModeExtensionsRequestedEvent {
        data?: Extension_Extension;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleListDevelopmentModeExtensionsEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleListDevelopmentModeExtensionsEvent is triggered after processing ListDevelopmentModeExtensions.
     */
    function onHandleListDevelopmentModeExtensions(cb: (event: HandleListDevelopmentModeExtensionsEvent) => void): void;

    interface HandleListDevelopmentModeExtensionsEvent {
        data?: Extension_Extension;

        next(): void;
    }

    /**
     * @event HandleGetAllExtensionsRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetAllExtensionsRequestedEvent is triggered when GetAllExtensions is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetAllExtensionsRequested(cb: (event: HandleGetAllExtensionsRequestedEvent) => void): void;

    interface HandleGetAllExtensionsRequestedEvent {
        withUpdates: boolean;
        data?: ExtensionRepo_AllExtensions;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetAllExtensionsEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetAllExtensionsEvent is triggered after processing GetAllExtensions.
     */
    function onHandleGetAllExtensions(cb: (event: HandleGetAllExtensionsEvent) => void): void;

    interface HandleGetAllExtensionsEvent {
        data?: ExtensionRepo_AllExtensions;

        next(): void;
    }

    /**
     * @event HandleGetExtensionUpdateDataRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetExtensionUpdateDataRequestedEvent is triggered when GetExtensionUpdateData is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetExtensionUpdateDataRequested(cb: (event: HandleGetExtensionUpdateDataRequestedEvent) => void): void;

    interface HandleGetExtensionUpdateDataRequestedEvent {
        data?: ExtensionRepo_UpdateData;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetExtensionUpdateDataEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetExtensionUpdateDataEvent is triggered after processing GetExtensionUpdateData.
     */
    function onHandleGetExtensionUpdateData(cb: (event: HandleGetExtensionUpdateDataEvent) => void): void;

    interface HandleGetExtensionUpdateDataEvent {
        data?: ExtensionRepo_UpdateData;

        next(): void;
    }

    /**
     * @event HandleListMangaProviderExtensionsRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleListMangaProviderExtensionsRequestedEvent is triggered when ListMangaProviderExtensions is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleListMangaProviderExtensionsRequested(cb: (event: HandleListMangaProviderExtensionsRequestedEvent) => void): void;

    interface HandleListMangaProviderExtensionsRequestedEvent {
        data?: ExtensionRepo_MangaProviderExtensionItem;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleListMangaProviderExtensionsEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleListMangaProviderExtensionsEvent is triggered after processing ListMangaProviderExtensions.
     */
    function onHandleListMangaProviderExtensions(cb: (event: HandleListMangaProviderExtensionsEvent) => void): void;

    interface HandleListMangaProviderExtensionsEvent {
        data?: ExtensionRepo_MangaProviderExtensionItem;

        next(): void;
    }

    /**
     * @event HandleListOnlinestreamProviderExtensionsRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleListOnlinestreamProviderExtensionsRequestedEvent is triggered when ListOnlinestreamProviderExtensions is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleListOnlinestreamProviderExtensionsRequested(cb: (event: HandleListOnlinestreamProviderExtensionsRequestedEvent) => void): void;

    interface HandleListOnlinestreamProviderExtensionsRequestedEvent {
        data?: ExtensionRepo_OnlinestreamProviderExtensionItem;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleListOnlinestreamProviderExtensionsEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleListOnlinestreamProviderExtensionsEvent is triggered after processing ListOnlinestreamProviderExtensions.
     */
    function onHandleListOnlinestreamProviderExtensions(cb: (event: HandleListOnlinestreamProviderExtensionsEvent) => void): void;

    interface HandleListOnlinestreamProviderExtensionsEvent {
        data?: ExtensionRepo_OnlinestreamProviderExtensionItem;

        next(): void;
    }

    /**
     * @event HandleListAnimeTorrentProviderExtensionsRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleListAnimeTorrentProviderExtensionsRequestedEvent is triggered when ListAnimeTorrentProviderExtensions is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleListAnimeTorrentProviderExtensionsRequested(cb: (event: HandleListAnimeTorrentProviderExtensionsRequestedEvent) => void): void;

    interface HandleListAnimeTorrentProviderExtensionsRequestedEvent {
        data?: ExtensionRepo_AnimeTorrentProviderExtensionItem;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleListAnimeTorrentProviderExtensionsEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleListAnimeTorrentProviderExtensionsEvent is triggered after processing ListAnimeTorrentProviderExtensions.
     */
    function onHandleListAnimeTorrentProviderExtensions(cb: (event: HandleListAnimeTorrentProviderExtensionsEvent) => void): void;

    interface HandleListAnimeTorrentProviderExtensionsEvent {
        data?: ExtensionRepo_AnimeTorrentProviderExtensionItem;

        next(): void;
    }

    /**
     * @event HandleGetPluginSettingsRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetPluginSettingsRequestedEvent is triggered when GetPluginSettings is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetPluginSettingsRequested(cb: (event: HandleGetPluginSettingsRequestedEvent) => void): void;

    interface HandleGetPluginSettingsRequestedEvent {
        data?: ExtensionRepo_StoredPluginSettingsData;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetPluginSettingsEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetPluginSettingsEvent is triggered after processing GetPluginSettings.
     */
    function onHandleGetPluginSettings(cb: (event: HandleGetPluginSettingsEvent) => void): void;

    interface HandleGetPluginSettingsEvent {
        data?: ExtensionRepo_StoredPluginSettingsData;

        next(): void;
    }

    /**
     * @event HandleSetPluginSettingsPinnedTraysRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleSetPluginSettingsPinnedTraysRequestedEvent is triggered when SetPluginSettingsPinnedTrays is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleSetPluginSettingsPinnedTraysRequested(cb: (event: HandleSetPluginSettingsPinnedTraysRequestedEvent) => void): void;

    interface HandleSetPluginSettingsPinnedTraysRequestedEvent {
        pinnedTrayPluginIds?: Array<string>;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGrantPluginPermissionsRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGrantPluginPermissionsRequestedEvent is triggered when GrantPluginPermissions is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGrantPluginPermissionsRequested(cb: (event: HandleGrantPluginPermissionsRequestedEvent) => void): void;

    interface HandleGrantPluginPermissionsRequestedEvent {
        id: string;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleRunExtensionPlaygroundCodeRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleRunExtensionPlaygroundCodeRequestedEvent is triggered when RunExtensionPlaygroundCode is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleRunExtensionPlaygroundCodeRequested(cb: (event: HandleRunExtensionPlaygroundCodeRequestedEvent) => void): void;

    interface HandleRunExtensionPlaygroundCodeRequestedEvent {
        params?: RunPlaygroundCodeParams;
        data?: RunPlaygroundCodeResponse;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleRunExtensionPlaygroundCodeEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleRunExtensionPlaygroundCodeEvent is triggered after processing RunExtensionPlaygroundCode.
     */
    function onHandleRunExtensionPlaygroundCode(cb: (event: HandleRunExtensionPlaygroundCodeEvent) => void): void;

    interface HandleRunExtensionPlaygroundCodeEvent {
        data?: RunPlaygroundCodeResponse;

        next(): void;
    }

    /**
     * @event HandleGetExtensionUserConfigRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetExtensionUserConfigRequestedEvent is triggered when GetExtensionUserConfig is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetExtensionUserConfigRequested(cb: (event: HandleGetExtensionUserConfigRequestedEvent) => void): void;

    interface HandleGetExtensionUserConfigRequestedEvent {
        data?: ExtensionRepo_ExtensionUserConfig;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetExtensionUserConfigEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetExtensionUserConfigEvent is triggered after processing GetExtensionUserConfig.
     */
    function onHandleGetExtensionUserConfig(cb: (event: HandleGetExtensionUserConfigEvent) => void): void;

    interface HandleGetExtensionUserConfigEvent {
        data?: ExtensionRepo_ExtensionUserConfig;

        next(): void;
    }

    /**
     * @event HandleSaveExtensionUserConfigRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleSaveExtensionUserConfigRequestedEvent is triggered when SaveExtensionUserConfig is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleSaveExtensionUserConfigRequested(cb: (event: HandleSaveExtensionUserConfigRequestedEvent) => void): void;

    interface HandleSaveExtensionUserConfigRequestedEvent {
        id: string;
        version: number;
        values?: Record<string, string>;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetMarketplaceExtensionsRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetMarketplaceExtensionsRequestedEvent is triggered when GetMarketplaceExtensions is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetMarketplaceExtensionsRequested(cb: (event: HandleGetMarketplaceExtensionsRequestedEvent) => void): void;

    interface HandleGetMarketplaceExtensionsRequestedEvent {
        data?: Extension_Extension;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetMarketplaceExtensionsEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetMarketplaceExtensionsEvent is triggered after processing GetMarketplaceExtensions.
     */
    function onHandleGetMarketplaceExtensions(cb: (event: HandleGetMarketplaceExtensionsEvent) => void): void;

    interface HandleGetMarketplaceExtensionsEvent {
        data?: Extension_Extension;

        next(): void;
    }

    /**
     * @event HandleGetFileCacheTotalSizeRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetFileCacheTotalSizeRequestedEvent is triggered when GetFileCacheTotalSize is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetFileCacheTotalSizeRequested(cb: (event: HandleGetFileCacheTotalSizeRequestedEvent) => void): void;

    interface HandleGetFileCacheTotalSizeRequestedEvent {
        data: string;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetFileCacheTotalSizeEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetFileCacheTotalSizeEvent is triggered after processing GetFileCacheTotalSize.
     */
    function onHandleGetFileCacheTotalSize(cb: (event: HandleGetFileCacheTotalSizeEvent) => void): void;

    interface HandleGetFileCacheTotalSizeEvent {
        data: string;

        next(): void;
    }

    /**
     * @event HandleRemoveFileCacheBucketRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleRemoveFileCacheBucketRequestedEvent is triggered when RemoveFileCacheBucket is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleRemoveFileCacheBucketRequested(cb: (event: HandleRemoveFileCacheBucketRequestedEvent) => void): void;

    interface HandleRemoveFileCacheBucketRequestedEvent {
        bucket: string;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetFileCacheMediastreamVideoFilesTotalSizeRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetFileCacheMediastreamVideoFilesTotalSizeRequestedEvent is triggered when GetFileCacheMediastreamVideoFilesTotalSize is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetFileCacheMediastreamVideoFilesTotalSizeRequested(cb: (event: HandleGetFileCacheMediastreamVideoFilesTotalSizeRequestedEvent) => void): void;

    interface HandleGetFileCacheMediastreamVideoFilesTotalSizeRequestedEvent {
        data: string;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetFileCacheMediastreamVideoFilesTotalSizeEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetFileCacheMediastreamVideoFilesTotalSizeEvent is triggered after processing GetFileCacheMediastreamVideoFilesTotalSize.
     */
    function onHandleGetFileCacheMediastreamVideoFilesTotalSize(cb: (event: HandleGetFileCacheMediastreamVideoFilesTotalSizeEvent) => void): void;

    interface HandleGetFileCacheMediastreamVideoFilesTotalSizeEvent {
        data: string;

        next(): void;
    }

    /**
     * @event HandleClearFileCacheMediastreamVideoFilesRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleClearFileCacheMediastreamVideoFilesRequestedEvent is triggered when ClearFileCacheMediastreamVideoFiles is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleClearFileCacheMediastreamVideoFilesRequested(cb: (event: HandleClearFileCacheMediastreamVideoFilesRequestedEvent) => void): void;

    interface HandleClearFileCacheMediastreamVideoFilesRequestedEvent {
        next(): void;

        preventDefault(): void;

    }

    /**
     * @event HandleGetLocalFilesRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetLocalFilesRequestedEvent is triggered when GetLocalFiles is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetLocalFilesRequested(cb: (event: HandleGetLocalFilesRequestedEvent) => void): void;

    interface HandleGetLocalFilesRequestedEvent {
        data?: Anime_LocalFile;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetLocalFilesEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetLocalFilesEvent is triggered after processing GetLocalFiles.
     */
    function onHandleGetLocalFiles(cb: (event: HandleGetLocalFilesEvent) => void): void;

    interface HandleGetLocalFilesEvent {
        data?: Anime_LocalFile;

        next(): void;
    }

    /**
     * @event HandleImportLocalFilesRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleImportLocalFilesRequestedEvent is triggered when ImportLocalFiles is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleImportLocalFilesRequested(cb: (event: HandleImportLocalFilesRequestedEvent) => void): void;

    interface HandleImportLocalFilesRequestedEvent {
        dataFilePath: string;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleLocalFileBulkActionRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleLocalFileBulkActionRequestedEvent is triggered when LocalFileBulkAction is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleLocalFileBulkActionRequested(cb: (event: HandleLocalFileBulkActionRequestedEvent) => void): void;

    interface HandleLocalFileBulkActionRequestedEvent {
        action: string;
        data?: Anime_LocalFile;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleLocalFileBulkActionEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleLocalFileBulkActionEvent is triggered after processing LocalFileBulkAction.
     */
    function onHandleLocalFileBulkAction(cb: (event: HandleLocalFileBulkActionEvent) => void): void;

    interface HandleLocalFileBulkActionEvent {
        data?: Anime_LocalFile;

        next(): void;
    }

    /**
     * @event HandleUpdateLocalFileDataRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleUpdateLocalFileDataRequestedEvent is triggered when UpdateLocalFileData is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleUpdateLocalFileDataRequested(cb: (event: HandleUpdateLocalFileDataRequestedEvent) => void): void;

    interface HandleUpdateLocalFileDataRequestedEvent {
        path: string;
        metadata?: Anime_LocalFileMetadata;
        locked: boolean;
        ignored: boolean;
        mediaId: number;
        data?: Anime_LocalFile;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleUpdateLocalFileDataEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleUpdateLocalFileDataEvent is triggered after processing UpdateLocalFileData.
     */
    function onHandleUpdateLocalFileData(cb: (event: HandleUpdateLocalFileDataEvent) => void): void;

    interface HandleUpdateLocalFileDataEvent {
        data?: Anime_LocalFile;

        next(): void;
    }

    /**
     * @event HandleUpdateLocalFilesRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleUpdateLocalFilesRequestedEvent is triggered when UpdateLocalFiles is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleUpdateLocalFilesRequested(cb: (event: HandleUpdateLocalFilesRequestedEvent) => void): void;

    interface HandleUpdateLocalFilesRequestedEvent {
        paths?: Array<string>;
        action: string;
        mediaId: number;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleDeleteLocalFilesRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleDeleteLocalFilesRequestedEvent is triggered when DeleteLocalFiles is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleDeleteLocalFilesRequested(cb: (event: HandleDeleteLocalFilesRequestedEvent) => void): void;

    interface HandleDeleteLocalFilesRequestedEvent {
        paths?: Array<string>;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleRemoveEmptyDirectoriesRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleRemoveEmptyDirectoriesRequestedEvent is triggered when RemoveEmptyDirectories is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleRemoveEmptyDirectoriesRequested(cb: (event: HandleRemoveEmptyDirectoriesRequestedEvent) => void): void;

    interface HandleRemoveEmptyDirectoriesRequestedEvent {
        next(): void;

        preventDefault(): void;

    }

    /**
     * @event HandleMALAuthRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleMALAuthRequestedEvent is triggered when MALAuth is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleMALAuthRequested(cb: (event: HandleMALAuthRequestedEvent) => void): void;

    interface HandleMALAuthRequestedEvent {
        code: string;
        state: string;
        code_verifier: string;
        data?: MalAuthResponse;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleMALAuthEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleMALAuthEvent is triggered after processing MALAuth.
     */
    function onHandleMALAuth(cb: (event: HandleMALAuthEvent) => void): void;

    interface HandleMALAuthEvent {
        data?: MalAuthResponse;

        next(): void;
    }

    /**
     * @event HandleEditMALListEntryProgressRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleEditMALListEntryProgressRequestedEvent is triggered when EditMALListEntryProgress is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleEditMALListEntryProgressRequested(cb: (event: HandleEditMALListEntryProgressRequestedEvent) => void): void;

    interface HandleEditMALListEntryProgressRequestedEvent {
        mediaId: number;
        progress: number;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleMALLogoutRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleMALLogoutRequestedEvent is triggered when MALLogout is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleMALLogoutRequested(cb: (event: HandleMALLogoutRequestedEvent) => void): void;

    interface HandleMALLogoutRequestedEvent {
        next(): void;

        preventDefault(): void;

    }

    /**
     * @event HandleGetAnilistMangaCollectionRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetAnilistMangaCollectionRequestedEvent is triggered when GetAnilistMangaCollection is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetAnilistMangaCollectionRequested(cb: (event: HandleGetAnilistMangaCollectionRequestedEvent) => void): void;

    interface HandleGetAnilistMangaCollectionRequestedEvent {
        bypassCache: boolean;
        data?: AL_MangaCollection;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetAnilistMangaCollectionEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetAnilistMangaCollectionEvent is triggered after processing GetAnilistMangaCollection.
     */
    function onHandleGetAnilistMangaCollection(cb: (event: HandleGetAnilistMangaCollectionEvent) => void): void;

    interface HandleGetAnilistMangaCollectionEvent {
        data?: AL_MangaCollection;

        next(): void;
    }

    /**
     * @event HandleGetRawAnilistMangaCollectionRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetRawAnilistMangaCollectionRequestedEvent is triggered when GetRawAnilistMangaCollection is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetRawAnilistMangaCollectionRequested(cb: (event: HandleGetRawAnilistMangaCollectionRequestedEvent) => void): void;

    interface HandleGetRawAnilistMangaCollectionRequestedEvent {
        data?: AL_MangaCollection;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetRawAnilistMangaCollectionEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetRawAnilistMangaCollectionEvent is triggered after processing GetRawAnilistMangaCollection.
     */
    function onHandleGetRawAnilistMangaCollection(cb: (event: HandleGetRawAnilistMangaCollectionEvent) => void): void;

    interface HandleGetRawAnilistMangaCollectionEvent {
        data?: AL_MangaCollection;

        next(): void;
    }

    /**
     * @event HandleGetMangaCollectionRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetMangaCollectionRequestedEvent is triggered when GetMangaCollection is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetMangaCollectionRequested(cb: (event: HandleGetMangaCollectionRequestedEvent) => void): void;

    interface HandleGetMangaCollectionRequestedEvent {
        data?: Manga_Collection;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetMangaCollectionEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetMangaCollectionEvent is triggered after processing GetMangaCollection.
     */
    function onHandleGetMangaCollection(cb: (event: HandleGetMangaCollectionEvent) => void): void;

    interface HandleGetMangaCollectionEvent {
        data?: Manga_Collection;

        next(): void;
    }

    /**
     * @event HandleGetMangaEntryRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetMangaEntryRequestedEvent is triggered when GetMangaEntry is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetMangaEntryRequested(cb: (event: HandleGetMangaEntryRequestedEvent) => void): void;

    interface HandleGetMangaEntryRequestedEvent {
        id: number;
        data?: Manga_Entry;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetMangaEntryEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetMangaEntryEvent is triggered after processing GetMangaEntry.
     */
    function onHandleGetMangaEntry(cb: (event: HandleGetMangaEntryEvent) => void): void;

    interface HandleGetMangaEntryEvent {
        data?: Manga_Entry;

        next(): void;
    }

    /**
     * @event HandleGetMangaEntryDetailsRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetMangaEntryDetailsRequestedEvent is triggered when GetMangaEntryDetails is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetMangaEntryDetailsRequested(cb: (event: HandleGetMangaEntryDetailsRequestedEvent) => void): void;

    interface HandleGetMangaEntryDetailsRequestedEvent {
        id: number;
        data?: AL_MangaDetailsById_Media;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetMangaEntryDetailsEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetMangaEntryDetailsEvent is triggered after processing GetMangaEntryDetails.
     */
    function onHandleGetMangaEntryDetails(cb: (event: HandleGetMangaEntryDetailsEvent) => void): void;

    interface HandleGetMangaEntryDetailsEvent {
        data?: AL_MangaDetailsById_Media;

        next(): void;
    }

    /**
     * @event HandleGetMangaLatestChapterNumbersMapRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetMangaLatestChapterNumbersMapRequestedEvent is triggered when GetMangaLatestChapterNumbersMap is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetMangaLatestChapterNumbersMapRequested(cb: (event: HandleGetMangaLatestChapterNumbersMapRequestedEvent) => void): void;

    interface HandleGetMangaLatestChapterNumbersMapRequestedEvent {
        data?: Manga_MangaLatestChapterNumberItem;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetMangaLatestChapterNumbersMapEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetMangaLatestChapterNumbersMapEvent is triggered after processing GetMangaLatestChapterNumbersMap.
     */
    function onHandleGetMangaLatestChapterNumbersMap(cb: (event: HandleGetMangaLatestChapterNumbersMapEvent) => void): void;

    interface HandleGetMangaLatestChapterNumbersMapEvent {
        data?: Manga_MangaLatestChapterNumberItem;

        next(): void;
    }

    /**
     * @event HandleRefetchMangaChapterContainersRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleRefetchMangaChapterContainersRequestedEvent is triggered when RefetchMangaChapterContainers is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleRefetchMangaChapterContainersRequested(cb: (event: HandleRefetchMangaChapterContainersRequestedEvent) => void): void;

    interface HandleRefetchMangaChapterContainersRequestedEvent {
        selectedProviderMap?: Record<number, string>;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleEmptyMangaEntryCacheRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleEmptyMangaEntryCacheRequestedEvent is triggered when EmptyMangaEntryCache is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleEmptyMangaEntryCacheRequested(cb: (event: HandleEmptyMangaEntryCacheRequestedEvent) => void): void;

    interface HandleEmptyMangaEntryCacheRequestedEvent {
        mediaId: number;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetMangaEntryChaptersRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetMangaEntryChaptersRequestedEvent is triggered when GetMangaEntryChapters is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetMangaEntryChaptersRequested(cb: (event: HandleGetMangaEntryChaptersRequestedEvent) => void): void;

    interface HandleGetMangaEntryChaptersRequestedEvent {
        mediaId: number;
        provider: string;
        data?: Manga_ChapterContainer;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetMangaEntryChaptersEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetMangaEntryChaptersEvent is triggered after processing GetMangaEntryChapters.
     */
    function onHandleGetMangaEntryChapters(cb: (event: HandleGetMangaEntryChaptersEvent) => void): void;

    interface HandleGetMangaEntryChaptersEvent {
        data?: Manga_ChapterContainer;

        next(): void;
    }

    /**
     * @event HandleGetMangaEntryPagesRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetMangaEntryPagesRequestedEvent is triggered when GetMangaEntryPages is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetMangaEntryPagesRequested(cb: (event: HandleGetMangaEntryPagesRequestedEvent) => void): void;

    interface HandleGetMangaEntryPagesRequestedEvent {
        mediaId: number;
        provider: string;
        chapterId: string;
        doublePage: boolean;
        data?: Manga_PageContainer;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetMangaEntryPagesEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetMangaEntryPagesEvent is triggered after processing GetMangaEntryPages.
     */
    function onHandleGetMangaEntryPages(cb: (event: HandleGetMangaEntryPagesEvent) => void): void;

    interface HandleGetMangaEntryPagesEvent {
        data?: Manga_PageContainer;

        next(): void;
    }

    /**
     * @event HandleGetMangaEntryDownloadedChaptersRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetMangaEntryDownloadedChaptersRequestedEvent is triggered when GetMangaEntryDownloadedChapters is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetMangaEntryDownloadedChaptersRequested(cb: (event: HandleGetMangaEntryDownloadedChaptersRequestedEvent) => void): void;

    interface HandleGetMangaEntryDownloadedChaptersRequestedEvent {
        id: number;
        data?: Manga_ChapterContainer;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetMangaEntryDownloadedChaptersEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetMangaEntryDownloadedChaptersEvent is triggered after processing GetMangaEntryDownloadedChapters.
     */
    function onHandleGetMangaEntryDownloadedChapters(cb: (event: HandleGetMangaEntryDownloadedChaptersEvent) => void): void;

    interface HandleGetMangaEntryDownloadedChaptersEvent {
        data?: Manga_ChapterContainer;

        next(): void;
    }

    /**
     * @event HandleAnilistListMangaRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleAnilistListMangaRequestedEvent is triggered when AnilistListManga is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleAnilistListMangaRequested(cb: (event: HandleAnilistListMangaRequestedEvent) => void): void;

    interface HandleAnilistListMangaRequestedEvent {
        page: number;
        search: string;
        perPage: number;
        sort?: Array<MediaSort>;
        status?: Array<MediaStatus>;
        genres?: Array<string>;
        averageScore_greater: number;
        year: number;
        countryOfOrigin: string;
        isAdult: boolean;
        format?: AL_MediaFormat;
        data?: AL_ListManga;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleAnilistListMangaEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleAnilistListMangaEvent is triggered after processing AnilistListManga.
     */
    function onHandleAnilistListManga(cb: (event: HandleAnilistListMangaEvent) => void): void;

    interface HandleAnilistListMangaEvent {
        data?: AL_ListManga;

        next(): void;
    }

    /**
     * @event HandleUpdateMangaProgressRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleUpdateMangaProgressRequestedEvent is triggered when UpdateMangaProgress is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleUpdateMangaProgressRequested(cb: (event: HandleUpdateMangaProgressRequestedEvent) => void): void;

    interface HandleUpdateMangaProgressRequestedEvent {
        mediaId: number;
        malId: number;
        chapterNumber: number;
        totalChapters: number;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleMangaManualSearchRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleMangaManualSearchRequestedEvent is triggered when MangaManualSearch is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleMangaManualSearchRequested(cb: (event: HandleMangaManualSearchRequestedEvent) => void): void;

    interface HandleMangaManualSearchRequestedEvent {
        provider: string;
        query: string;
        data?: HibikeManga_SearchResult;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleMangaManualSearchEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleMangaManualSearchEvent is triggered after processing MangaManualSearch.
     */
    function onHandleMangaManualSearch(cb: (event: HandleMangaManualSearchEvent) => void): void;

    interface HandleMangaManualSearchEvent {
        data?: HibikeManga_SearchResult;

        next(): void;
    }

    /**
     * @event HandleMangaManualMappingRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleMangaManualMappingRequestedEvent is triggered when MangaManualMapping is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleMangaManualMappingRequested(cb: (event: HandleMangaManualMappingRequestedEvent) => void): void;

    interface HandleMangaManualMappingRequestedEvent {
        provider: string;
        mediaId: number;
        mangaId: string;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetMangaMappingRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetMangaMappingRequestedEvent is triggered when GetMangaMapping is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetMangaMappingRequested(cb: (event: HandleGetMangaMappingRequestedEvent) => void): void;

    interface HandleGetMangaMappingRequestedEvent {
        provider: string;
        mediaId: number;
        data?: Manga_MappingResponse;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetMangaMappingEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetMangaMappingEvent is triggered after processing GetMangaMapping.
     */
    function onHandleGetMangaMapping(cb: (event: HandleGetMangaMappingEvent) => void): void;

    interface HandleGetMangaMappingEvent {
        data?: Manga_MappingResponse;

        next(): void;
    }

    /**
     * @event HandleRemoveMangaMappingRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleRemoveMangaMappingRequestedEvent is triggered when RemoveMangaMapping is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleRemoveMangaMappingRequested(cb: (event: HandleRemoveMangaMappingRequestedEvent) => void): void;

    interface HandleRemoveMangaMappingRequestedEvent {
        provider: string;
        mediaId: number;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleDownloadMangaChaptersRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleDownloadMangaChaptersRequestedEvent is triggered when DownloadMangaChapters is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleDownloadMangaChaptersRequested(cb: (event: HandleDownloadMangaChaptersRequestedEvent) => void): void;

    interface HandleDownloadMangaChaptersRequestedEvent {
        mediaId: number;
        provider: string;
        chapterIds?: Array<string>;
        startNow: boolean;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetMangaDownloadDataRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetMangaDownloadDataRequestedEvent is triggered when GetMangaDownloadData is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetMangaDownloadDataRequested(cb: (event: HandleGetMangaDownloadDataRequestedEvent) => void): void;

    interface HandleGetMangaDownloadDataRequestedEvent {
        mediaId: number;
        cached: boolean;
        data?: Manga_MediaDownloadData;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetMangaDownloadDataEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetMangaDownloadDataEvent is triggered after processing GetMangaDownloadData.
     */
    function onHandleGetMangaDownloadData(cb: (event: HandleGetMangaDownloadDataEvent) => void): void;

    interface HandleGetMangaDownloadDataEvent {
        data?: Manga_MediaDownloadData;

        next(): void;
    }

    /**
     * @event HandleGetMangaDownloadQueueRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetMangaDownloadQueueRequestedEvent is triggered when GetMangaDownloadQueue is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetMangaDownloadQueueRequested(cb: (event: HandleGetMangaDownloadQueueRequestedEvent) => void): void;

    interface HandleGetMangaDownloadQueueRequestedEvent {
        data?: Models_ChapterDownloadQueueItem;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetMangaDownloadQueueEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetMangaDownloadQueueEvent is triggered after processing GetMangaDownloadQueue.
     */
    function onHandleGetMangaDownloadQueue(cb: (event: HandleGetMangaDownloadQueueEvent) => void): void;

    interface HandleGetMangaDownloadQueueEvent {
        data?: Models_ChapterDownloadQueueItem;

        next(): void;
    }

    /**
     * @event HandleStartMangaDownloadQueueRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleStartMangaDownloadQueueRequestedEvent is triggered when StartMangaDownloadQueue is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleStartMangaDownloadQueueRequested(cb: (event: HandleStartMangaDownloadQueueRequestedEvent) => void): void;

    interface HandleStartMangaDownloadQueueRequestedEvent {
        next(): void;

        preventDefault(): void;

    }

    /**
     * @event HandleStopMangaDownloadQueueRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleStopMangaDownloadQueueRequestedEvent is triggered when StopMangaDownloadQueue is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleStopMangaDownloadQueueRequested(cb: (event: HandleStopMangaDownloadQueueRequestedEvent) => void): void;

    interface HandleStopMangaDownloadQueueRequestedEvent {
        next(): void;

        preventDefault(): void;

    }

    /**
     * @event HandleClearAllChapterDownloadQueueRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleClearAllChapterDownloadQueueRequestedEvent is triggered when ClearAllChapterDownloadQueue is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleClearAllChapterDownloadQueueRequested(cb: (event: HandleClearAllChapterDownloadQueueRequestedEvent) => void): void;

    interface HandleClearAllChapterDownloadQueueRequestedEvent {
        next(): void;

        preventDefault(): void;

    }

    /**
     * @event HandleResetErroredChapterDownloadQueueRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleResetErroredChapterDownloadQueueRequestedEvent is triggered when ResetErroredChapterDownloadQueue is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleResetErroredChapterDownloadQueueRequested(cb: (event: HandleResetErroredChapterDownloadQueueRequestedEvent) => void): void;

    interface HandleResetErroredChapterDownloadQueueRequestedEvent {
        next(): void;

        preventDefault(): void;

    }

    /**
     * @event HandleDeleteMangaDownloadedChaptersRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleDeleteMangaDownloadedChaptersRequestedEvent is triggered when DeleteMangaDownloadedChapters is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleDeleteMangaDownloadedChaptersRequested(cb: (event: HandleDeleteMangaDownloadedChaptersRequestedEvent) => void): void;

    interface HandleDeleteMangaDownloadedChaptersRequestedEvent {
        downloadIds?: Array<DownloadID>;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetMangaDownloadsListRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetMangaDownloadsListRequestedEvent is triggered when GetMangaDownloadsList is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetMangaDownloadsListRequested(cb: (event: HandleGetMangaDownloadsListRequestedEvent) => void): void;

    interface HandleGetMangaDownloadsListRequestedEvent {
        data?: Manga_DownloadListItem;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetMangaDownloadsListEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetMangaDownloadsListEvent is triggered after processing GetMangaDownloadsList.
     */
    function onHandleGetMangaDownloadsList(cb: (event: HandleGetMangaDownloadsListEvent) => void): void;

    interface HandleGetMangaDownloadsListEvent {
        data?: Manga_DownloadListItem;

        next(): void;
    }

    /**
     * @event HandleTestDumpRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleTestDumpRequestedEvent is triggered when TestDump is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleTestDumpRequested(cb: (event: HandleTestDumpRequestedEvent) => void): void;

    interface HandleTestDumpRequestedEvent {
        next(): void;

        preventDefault(): void;

    }

    /**
     * @event HandleStartDefaultMediaPlayerRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleStartDefaultMediaPlayerRequestedEvent is triggered when StartDefaultMediaPlayer is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleStartDefaultMediaPlayerRequested(cb: (event: HandleStartDefaultMediaPlayerRequestedEvent) => void): void;

    interface HandleStartDefaultMediaPlayerRequestedEvent {
        next(): void;

        preventDefault(): void;

    }

    /**
     * @event HandleGetMediastreamSettingsRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetMediastreamSettingsRequestedEvent is triggered when GetMediastreamSettings is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetMediastreamSettingsRequested(cb: (event: HandleGetMediastreamSettingsRequestedEvent) => void): void;

    interface HandleGetMediastreamSettingsRequestedEvent {
        data?: Models_MediastreamSettings;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetMediastreamSettingsEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetMediastreamSettingsEvent is triggered after processing GetMediastreamSettings.
     */
    function onHandleGetMediastreamSettings(cb: (event: HandleGetMediastreamSettingsEvent) => void): void;

    interface HandleGetMediastreamSettingsEvent {
        data?: Models_MediastreamSettings;

        next(): void;
    }

    /**
     * @event HandleSaveMediastreamSettingsRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleSaveMediastreamSettingsRequestedEvent is triggered when SaveMediastreamSettings is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleSaveMediastreamSettingsRequested(cb: (event: HandleSaveMediastreamSettingsRequestedEvent) => void): void;

    interface HandleSaveMediastreamSettingsRequestedEvent {
        settings?: Models_MediastreamSettings;
        data?: Models_MediastreamSettings;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleSaveMediastreamSettingsEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleSaveMediastreamSettingsEvent is triggered after processing SaveMediastreamSettings.
     */
    function onHandleSaveMediastreamSettings(cb: (event: HandleSaveMediastreamSettingsEvent) => void): void;

    interface HandleSaveMediastreamSettingsEvent {
        data?: Models_MediastreamSettings;

        next(): void;
    }

    /**
     * @event HandleRequestMediastreamMediaContainerRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleRequestMediastreamMediaContainerRequestedEvent is triggered when RequestMediastreamMediaContainer is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleRequestMediastreamMediaContainerRequested(cb: (event: HandleRequestMediastreamMediaContainerRequestedEvent) => void): void;

    interface HandleRequestMediastreamMediaContainerRequestedEvent {
        path: string;
        streamType?: Mediastream_StreamType;
        audioStreamIndex: number;
        clientId: string;
        data?: Mediastream_MediaContainer;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleRequestMediastreamMediaContainerEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleRequestMediastreamMediaContainerEvent is triggered after processing RequestMediastreamMediaContainer.
     */
    function onHandleRequestMediastreamMediaContainer(cb: (event: HandleRequestMediastreamMediaContainerEvent) => void): void;

    interface HandleRequestMediastreamMediaContainerEvent {
        data?: Mediastream_MediaContainer;

        next(): void;
    }

    /**
     * @event HandlePreloadMediastreamMediaContainerRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandlePreloadMediastreamMediaContainerRequestedEvent is triggered when PreloadMediastreamMediaContainer is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandlePreloadMediastreamMediaContainerRequested(cb: (event: HandlePreloadMediastreamMediaContainerRequestedEvent) => void): void;

    interface HandlePreloadMediastreamMediaContainerRequestedEvent {
        path: string;
        streamType?: Mediastream_StreamType;
        audioStreamIndex: number;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleMediastreamShutdownTranscodeStreamRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleMediastreamShutdownTranscodeStreamRequestedEvent is triggered when MediastreamShutdownTranscodeStream is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleMediastreamShutdownTranscodeStreamRequested(cb: (event: HandleMediastreamShutdownTranscodeStreamRequestedEvent) => void): void;

    interface HandleMediastreamShutdownTranscodeStreamRequestedEvent {
        next(): void;

        preventDefault(): void;

    }

    /**
     * @event HandlePopulateTVDBEpisodesRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandlePopulateTVDBEpisodesRequestedEvent is triggered when PopulateTVDBEpisodes is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandlePopulateTVDBEpisodesRequested(cb: (event: HandlePopulateTVDBEpisodesRequestedEvent) => void): void;

    interface HandlePopulateTVDBEpisodesRequestedEvent {
        mediaId: number;
        data?: TVDB_Episode;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandlePopulateTVDBEpisodesEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandlePopulateTVDBEpisodesEvent is triggered after processing PopulateTVDBEpisodes.
     */
    function onHandlePopulateTVDBEpisodes(cb: (event: HandlePopulateTVDBEpisodesEvent) => void): void;

    interface HandlePopulateTVDBEpisodesEvent {
        data?: TVDB_Episode;

        next(): void;
    }

    /**
     * @event HandleEmptyTVDBEpisodesRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleEmptyTVDBEpisodesRequestedEvent is triggered when EmptyTVDBEpisodes is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleEmptyTVDBEpisodesRequested(cb: (event: HandleEmptyTVDBEpisodesRequestedEvent) => void): void;

    interface HandleEmptyTVDBEpisodesRequestedEvent {
        mediaId: number;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandlePopulateFillerDataRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandlePopulateFillerDataRequestedEvent is triggered when PopulateFillerData is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandlePopulateFillerDataRequested(cb: (event: HandlePopulateFillerDataRequestedEvent) => void): void;

    interface HandlePopulateFillerDataRequestedEvent {
        mediaId: number;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleRemoveFillerDataRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleRemoveFillerDataRequestedEvent is triggered when RemoveFillerData is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleRemoveFillerDataRequested(cb: (event: HandleRemoveFillerDataRequestedEvent) => void): void;

    interface HandleRemoveFillerDataRequestedEvent {
        mediaId: number;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetOnlineStreamEpisodeListRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetOnlineStreamEpisodeListRequestedEvent is triggered when GetOnlineStreamEpisodeList is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetOnlineStreamEpisodeListRequested(cb: (event: HandleGetOnlineStreamEpisodeListRequestedEvent) => void): void;

    interface HandleGetOnlineStreamEpisodeListRequestedEvent {
        mediaId: number;
        dubbed: boolean;
        provider: string;
        data?: Onlinestream_EpisodeListResponse;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetOnlineStreamEpisodeListEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetOnlineStreamEpisodeListEvent is triggered after processing GetOnlineStreamEpisodeList.
     */
    function onHandleGetOnlineStreamEpisodeList(cb: (event: HandleGetOnlineStreamEpisodeListEvent) => void): void;

    interface HandleGetOnlineStreamEpisodeListEvent {
        data?: Onlinestream_EpisodeListResponse;

        next(): void;
    }

    /**
     * @event HandleGetOnlineStreamEpisodeSourceRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetOnlineStreamEpisodeSourceRequestedEvent is triggered when GetOnlineStreamEpisodeSource is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetOnlineStreamEpisodeSourceRequested(cb: (event: HandleGetOnlineStreamEpisodeSourceRequestedEvent) => void): void;

    interface HandleGetOnlineStreamEpisodeSourceRequestedEvent {
        episodeNumber: number;
        mediaId: number;
        provider: string;
        dubbed: boolean;
        data?: Onlinestream_EpisodeSource;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetOnlineStreamEpisodeSourceEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetOnlineStreamEpisodeSourceEvent is triggered after processing GetOnlineStreamEpisodeSource.
     */
    function onHandleGetOnlineStreamEpisodeSource(cb: (event: HandleGetOnlineStreamEpisodeSourceEvent) => void): void;

    interface HandleGetOnlineStreamEpisodeSourceEvent {
        data?: Onlinestream_EpisodeSource;

        next(): void;
    }

    /**
     * @event HandleOnlineStreamEmptyCacheRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleOnlineStreamEmptyCacheRequestedEvent is triggered when OnlineStreamEmptyCache is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleOnlineStreamEmptyCacheRequested(cb: (event: HandleOnlineStreamEmptyCacheRequestedEvent) => void): void;

    interface HandleOnlineStreamEmptyCacheRequestedEvent {
        mediaId: number;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleOnlinestreamManualSearchRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleOnlinestreamManualSearchRequestedEvent is triggered when OnlinestreamManualSearch is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleOnlinestreamManualSearchRequested(cb: (event: HandleOnlinestreamManualSearchRequestedEvent) => void): void;

    interface HandleOnlinestreamManualSearchRequestedEvent {
        provider: string;
        query: string;
        dubbed: boolean;
        data?: HibikeOnlinestream_SearchResult;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleOnlinestreamManualSearchEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleOnlinestreamManualSearchEvent is triggered after processing OnlinestreamManualSearch.
     */
    function onHandleOnlinestreamManualSearch(cb: (event: HandleOnlinestreamManualSearchEvent) => void): void;

    interface HandleOnlinestreamManualSearchEvent {
        data?: HibikeOnlinestream_SearchResult;

        next(): void;
    }

    /**
     * @event HandleOnlinestreamManualMappingRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleOnlinestreamManualMappingRequestedEvent is triggered when OnlinestreamManualMapping is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleOnlinestreamManualMappingRequested(cb: (event: HandleOnlinestreamManualMappingRequestedEvent) => void): void;

    interface HandleOnlinestreamManualMappingRequestedEvent {
        provider: string;
        mediaId: number;
        animeId: string;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetOnlinestreamMappingRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetOnlinestreamMappingRequestedEvent is triggered when GetOnlinestreamMapping is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetOnlinestreamMappingRequested(cb: (event: HandleGetOnlinestreamMappingRequestedEvent) => void): void;

    interface HandleGetOnlinestreamMappingRequestedEvent {
        provider: string;
        mediaId: number;
        data?: Onlinestream_MappingResponse;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetOnlinestreamMappingEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetOnlinestreamMappingEvent is triggered after processing GetOnlinestreamMapping.
     */
    function onHandleGetOnlinestreamMapping(cb: (event: HandleGetOnlinestreamMappingEvent) => void): void;

    interface HandleGetOnlinestreamMappingEvent {
        data?: Onlinestream_MappingResponse;

        next(): void;
    }

    /**
     * @event HandleRemoveOnlinestreamMappingRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleRemoveOnlinestreamMappingRequestedEvent is triggered when RemoveOnlinestreamMapping is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleRemoveOnlinestreamMappingRequested(cb: (event: HandleRemoveOnlinestreamMappingRequestedEvent) => void): void;

    interface HandleRemoveOnlinestreamMappingRequestedEvent {
        provider: string;
        mediaId: number;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandlePlaybackPlayVideoRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandlePlaybackPlayVideoRequestedEvent is triggered when PlaybackPlayVideo is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandlePlaybackPlayVideoRequested(cb: (event: HandlePlaybackPlayVideoRequestedEvent) => void): void;

    interface HandlePlaybackPlayVideoRequestedEvent {
        path: string;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandlePlaybackPlayRandomVideoRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandlePlaybackPlayRandomVideoRequestedEvent is triggered when PlaybackPlayRandomVideo is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandlePlaybackPlayRandomVideoRequested(cb: (event: HandlePlaybackPlayRandomVideoRequestedEvent) => void): void;

    interface HandlePlaybackPlayRandomVideoRequestedEvent {
        next(): void;

        preventDefault(): void;

    }

    /**
     * @event HandlePlaybackSyncCurrentProgressRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandlePlaybackSyncCurrentProgressRequestedEvent is triggered when PlaybackSyncCurrentProgress is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandlePlaybackSyncCurrentProgressRequested(cb: (event: HandlePlaybackSyncCurrentProgressRequestedEvent) => void): void;

    interface HandlePlaybackSyncCurrentProgressRequestedEvent {
        data: number;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandlePlaybackSyncCurrentProgressEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandlePlaybackSyncCurrentProgressEvent is triggered after processing PlaybackSyncCurrentProgress.
     */
    function onHandlePlaybackSyncCurrentProgress(cb: (event: HandlePlaybackSyncCurrentProgressEvent) => void): void;

    interface HandlePlaybackSyncCurrentProgressEvent {
        data: number;

        next(): void;
    }

    /**
     * @event HandlePlaybackPlayNextEpisodeRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandlePlaybackPlayNextEpisodeRequestedEvent is triggered when PlaybackPlayNextEpisode is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandlePlaybackPlayNextEpisodeRequested(cb: (event: HandlePlaybackPlayNextEpisodeRequestedEvent) => void): void;

    interface HandlePlaybackPlayNextEpisodeRequestedEvent {
        next(): void;

        preventDefault(): void;

    }

    /**
     * @event HandlePlaybackGetNextEpisodeRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandlePlaybackGetNextEpisodeRequestedEvent is triggered when PlaybackGetNextEpisode is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandlePlaybackGetNextEpisodeRequested(cb: (event: HandlePlaybackGetNextEpisodeRequestedEvent) => void): void;

    interface HandlePlaybackGetNextEpisodeRequestedEvent {
        data?: Anime_LocalFile;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandlePlaybackGetNextEpisodeEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandlePlaybackGetNextEpisodeEvent is triggered after processing PlaybackGetNextEpisode.
     */
    function onHandlePlaybackGetNextEpisode(cb: (event: HandlePlaybackGetNextEpisodeEvent) => void): void;

    interface HandlePlaybackGetNextEpisodeEvent {
        data?: Anime_LocalFile;

        next(): void;
    }

    /**
     * @event HandlePlaybackAutoPlayNextEpisodeRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandlePlaybackAutoPlayNextEpisodeRequestedEvent is triggered when PlaybackAutoPlayNextEpisode is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandlePlaybackAutoPlayNextEpisodeRequested(cb: (event: HandlePlaybackAutoPlayNextEpisodeRequestedEvent) => void): void;

    interface HandlePlaybackAutoPlayNextEpisodeRequestedEvent {
        next(): void;

        preventDefault(): void;

    }

    /**
     * @event HandlePlaybackStartPlaylistRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandlePlaybackStartPlaylistRequestedEvent is triggered when PlaybackStartPlaylist is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandlePlaybackStartPlaylistRequested(cb: (event: HandlePlaybackStartPlaylistRequestedEvent) => void): void;

    interface HandlePlaybackStartPlaylistRequestedEvent {
        dbId: number;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandlePlaybackCancelCurrentPlaylistRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandlePlaybackCancelCurrentPlaylistRequestedEvent is triggered when PlaybackCancelCurrentPlaylist is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandlePlaybackCancelCurrentPlaylistRequested(cb: (event: HandlePlaybackCancelCurrentPlaylistRequestedEvent) => void): void;

    interface HandlePlaybackCancelCurrentPlaylistRequestedEvent {
        next(): void;

        preventDefault(): void;

    }

    /**
     * @event HandlePlaybackPlaylistNextRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandlePlaybackPlaylistNextRequestedEvent is triggered when PlaybackPlaylistNext is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandlePlaybackPlaylistNextRequested(cb: (event: HandlePlaybackPlaylistNextRequestedEvent) => void): void;

    interface HandlePlaybackPlaylistNextRequestedEvent {
        next(): void;

        preventDefault(): void;

    }

    /**
     * @event HandlePlaybackStartManualTrackingRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandlePlaybackStartManualTrackingRequestedEvent is triggered when PlaybackStartManualTracking is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandlePlaybackStartManualTrackingRequested(cb: (event: HandlePlaybackStartManualTrackingRequestedEvent) => void): void;

    interface HandlePlaybackStartManualTrackingRequestedEvent {
        mediaId: number;
        episodeNumber: number;
        clientId: string;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandlePlaybackCancelManualTrackingRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandlePlaybackCancelManualTrackingRequestedEvent is triggered when PlaybackCancelManualTracking is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandlePlaybackCancelManualTrackingRequested(cb: (event: HandlePlaybackCancelManualTrackingRequestedEvent) => void): void;

    interface HandlePlaybackCancelManualTrackingRequestedEvent {
        next(): void;

        preventDefault(): void;

    }

    /**
     * @event HandleCreatePlaylistRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleCreatePlaylistRequestedEvent is triggered when CreatePlaylist is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleCreatePlaylistRequested(cb: (event: HandleCreatePlaylistRequestedEvent) => void): void;

    interface HandleCreatePlaylistRequestedEvent {
        name: string;
        paths?: Array<string>;
        data?: Anime_Playlist;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleCreatePlaylistEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleCreatePlaylistEvent is triggered after processing CreatePlaylist.
     */
    function onHandleCreatePlaylist(cb: (event: HandleCreatePlaylistEvent) => void): void;

    interface HandleCreatePlaylistEvent {
        data?: Anime_Playlist;

        next(): void;
    }

    /**
     * @event HandleGetPlaylistsRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetPlaylistsRequestedEvent is triggered when GetPlaylists is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetPlaylistsRequested(cb: (event: HandleGetPlaylistsRequestedEvent) => void): void;

    interface HandleGetPlaylistsRequestedEvent {
        data?: Anime_Playlist;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetPlaylistsEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetPlaylistsEvent is triggered after processing GetPlaylists.
     */
    function onHandleGetPlaylists(cb: (event: HandleGetPlaylistsEvent) => void): void;

    interface HandleGetPlaylistsEvent {
        data?: Anime_Playlist;

        next(): void;
    }

    /**
     * @event HandleUpdatePlaylistRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleUpdatePlaylistRequestedEvent is triggered when UpdatePlaylist is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleUpdatePlaylistRequested(cb: (event: HandleUpdatePlaylistRequestedEvent) => void): void;

    interface HandleUpdatePlaylistRequestedEvent {
        id: number;
        dbId: number;
        name: string;
        paths?: Array<string>;
        data?: Anime_Playlist;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleUpdatePlaylistEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleUpdatePlaylistEvent is triggered after processing UpdatePlaylist.
     */
    function onHandleUpdatePlaylist(cb: (event: HandleUpdatePlaylistEvent) => void): void;

    interface HandleUpdatePlaylistEvent {
        data?: Anime_Playlist;

        next(): void;
    }

    /**
     * @event HandleDeletePlaylistRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleDeletePlaylistRequestedEvent is triggered when DeletePlaylist is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleDeletePlaylistRequested(cb: (event: HandleDeletePlaylistRequestedEvent) => void): void;

    interface HandleDeletePlaylistRequestedEvent {
        dbId: number;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetPlaylistEpisodesRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetPlaylistEpisodesRequestedEvent is triggered when GetPlaylistEpisodes is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetPlaylistEpisodesRequested(cb: (event: HandleGetPlaylistEpisodesRequestedEvent) => void): void;

    interface HandleGetPlaylistEpisodesRequestedEvent {
        id: number;
        progress: number;
        data?: Anime_LocalFile;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetPlaylistEpisodesEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetPlaylistEpisodesEvent is triggered after processing GetPlaylistEpisodes.
     */
    function onHandleGetPlaylistEpisodes(cb: (event: HandleGetPlaylistEpisodesEvent) => void): void;

    interface HandleGetPlaylistEpisodesEvent {
        data?: Anime_LocalFile;

        next(): void;
    }

    /**
     * @event HandleInstallLatestUpdateRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleInstallLatestUpdateRequestedEvent is triggered when InstallLatestUpdate is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleInstallLatestUpdateRequested(cb: (event: HandleInstallLatestUpdateRequestedEvent) => void): void;

    interface HandleInstallLatestUpdateRequestedEvent {
        fallback_destination: string;
        data?: Status;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleInstallLatestUpdateEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleInstallLatestUpdateEvent is triggered after processing InstallLatestUpdate.
     */
    function onHandleInstallLatestUpdate(cb: (event: HandleInstallLatestUpdateEvent) => void): void;

    interface HandleInstallLatestUpdateEvent {
        data?: Status;

        next(): void;
    }

    /**
     * @event HandleGetLatestUpdateRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetLatestUpdateRequestedEvent is triggered when GetLatestUpdate is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetLatestUpdateRequested(cb: (event: HandleGetLatestUpdateRequestedEvent) => void): void;

    interface HandleGetLatestUpdateRequestedEvent {
        data?: Updater_Update;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetLatestUpdateEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetLatestUpdateEvent is triggered after processing GetLatestUpdate.
     */
    function onHandleGetLatestUpdate(cb: (event: HandleGetLatestUpdateEvent) => void): void;

    interface HandleGetLatestUpdateEvent {
        data?: Updater_Update;

        next(): void;
    }

    /**
     * @event HandleGetChangelogRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetChangelogRequestedEvent is triggered when GetChangelog is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetChangelogRequested(cb: (event: HandleGetChangelogRequestedEvent) => void): void;

    interface HandleGetChangelogRequestedEvent {
        data: string;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetChangelogEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetChangelogEvent is triggered after processing GetChangelog.
     */
    function onHandleGetChangelog(cb: (event: HandleGetChangelogEvent) => void): void;

    interface HandleGetChangelogEvent {
        data: string;

        next(): void;
    }

    /**
     * @event HandleSaveIssueReportRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleSaveIssueReportRequestedEvent is triggered when SaveIssueReport is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleSaveIssueReportRequested(cb: (event: HandleSaveIssueReportRequestedEvent) => void): void;

    interface HandleSaveIssueReportRequestedEvent {
        clickLogs?: Array<ClickLog>;
        networkLogs?: Array<NetworkLog>;
        reactQueryLogs?: Array<ReactQueryLog>;
        consoleLogs?: Array<ConsoleLog>;
        isAnimeLibraryIssue: boolean;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleDownloadIssueReportRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleDownloadIssueReportRequestedEvent is triggered when DownloadIssueReport is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleDownloadIssueReportRequested(cb: (event: HandleDownloadIssueReportRequestedEvent) => void): void;

    interface HandleDownloadIssueReportRequestedEvent {
        data?: Report_IssueReport;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleDownloadIssueReportEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleDownloadIssueReportEvent is triggered after processing DownloadIssueReport.
     */
    function onHandleDownloadIssueReport(cb: (event: HandleDownloadIssueReportEvent) => void): void;

    interface HandleDownloadIssueReportEvent {
        data?: Report_IssueReport;

        next(): void;
    }

    /**
     * @event HandleScanLocalFilesRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleScanLocalFilesRequestedEvent is triggered when ScanLocalFiles is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleScanLocalFilesRequested(cb: (event: HandleScanLocalFilesRequestedEvent) => void): void;

    interface HandleScanLocalFilesRequestedEvent {
        enhanced: boolean;
        skipLockedFiles: boolean;
        skipIgnoredFiles: boolean;
        data?: Anime_LocalFile;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleScanLocalFilesEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleScanLocalFilesEvent is triggered after processing ScanLocalFiles.
     */
    function onHandleScanLocalFiles(cb: (event: HandleScanLocalFilesEvent) => void): void;

    interface HandleScanLocalFilesEvent {
        data?: Anime_LocalFile;

        next(): void;
    }

    /**
     * @event HandleGetScanSummariesRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetScanSummariesRequestedEvent is triggered when GetScanSummaries is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetScanSummariesRequested(cb: (event: HandleGetScanSummariesRequestedEvent) => void): void;

    interface HandleGetScanSummariesRequestedEvent {
        data?: Summary_ScanSummaryItem;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetScanSummariesEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetScanSummariesEvent is triggered after processing GetScanSummaries.
     */
    function onHandleGetScanSummaries(cb: (event: HandleGetScanSummariesEvent) => void): void;

    interface HandleGetScanSummariesEvent {
        data?: Summary_ScanSummaryItem;

        next(): void;
    }

    /**
     * @event HandleGetSettingsRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetSettingsRequestedEvent is triggered when GetSettings is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetSettingsRequested(cb: (event: HandleGetSettingsRequestedEvent) => void): void;

    interface HandleGetSettingsRequestedEvent {
        data?: Models_Settings;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetSettingsEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetSettingsEvent is triggered after processing GetSettings.
     */
    function onHandleGetSettings(cb: (event: HandleGetSettingsEvent) => void): void;

    interface HandleGetSettingsEvent {
        data?: Models_Settings;

        next(): void;
    }

    /**
     * @event HandleGettingStartedRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGettingStartedRequestedEvent is triggered when GettingStarted is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGettingStartedRequested(cb: (event: HandleGettingStartedRequestedEvent) => void): void;

    interface HandleGettingStartedRequestedEvent {
        library?: Models_LibrarySettings;
        mediaPlayer?: Models_MediaPlayerSettings;
        torrent?: Models_TorrentSettings;
        anilist?: Models_AnilistSettings;
        discord?: Models_DiscordSettings;
        manga?: Models_MangaSettings;
        notifications?: Models_NotificationSettings;
        enableTranscode: boolean;
        enableTorrentStreaming: boolean;
        debridProvider: string;
        debridApiKey: string;
        data?: Status;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGettingStartedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGettingStartedEvent is triggered after processing GettingStarted.
     */
    function onHandleGettingStarted(cb: (event: HandleGettingStartedEvent) => void): void;

    interface HandleGettingStartedEvent {
        data?: Status;

        next(): void;
    }

    /**
     * @event HandleSaveSettingsRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleSaveSettingsRequestedEvent is triggered when SaveSettings is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleSaveSettingsRequested(cb: (event: HandleSaveSettingsRequestedEvent) => void): void;

    interface HandleSaveSettingsRequestedEvent {
        library?: Models_LibrarySettings;
        mediaPlayer?: Models_MediaPlayerSettings;
        torrent?: Models_TorrentSettings;
        anilist?: Models_AnilistSettings;
        discord?: Models_DiscordSettings;
        manga?: Models_MangaSettings;
        notifications?: Models_NotificationSettings;
        data?: Status;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleSaveSettingsEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleSaveSettingsEvent is triggered after processing SaveSettings.
     */
    function onHandleSaveSettings(cb: (event: HandleSaveSettingsEvent) => void): void;

    interface HandleSaveSettingsEvent {
        data?: Status;

        next(): void;
    }

    /**
     * @event HandleSaveAutoDownloaderSettingsRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleSaveAutoDownloaderSettingsRequestedEvent is triggered when SaveAutoDownloaderSettings is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleSaveAutoDownloaderSettingsRequested(cb: (event: HandleSaveAutoDownloaderSettingsRequestedEvent) => void): void;

    interface HandleSaveAutoDownloaderSettingsRequestedEvent {
        interval: number;
        enabled: boolean;
        downloadAutomatically: boolean;
        enableEnhancedQueries: boolean;
        enableSeasonCheck: boolean;
        useDebrid: boolean;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetStatusRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetStatusRequestedEvent is triggered when GetStatus is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetStatusRequested(cb: (event: HandleGetStatusRequestedEvent) => void): void;

    interface HandleGetStatusRequestedEvent {
        data?: Status;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetStatusEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetStatusEvent is triggered after processing GetStatus.
     */
    function onHandleGetStatus(cb: (event: HandleGetStatusEvent) => void): void;

    interface HandleGetStatusEvent {
        data?: Status;

        next(): void;
    }

    /**
     * @event HandleGetLogFilenamesRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetLogFilenamesRequestedEvent is triggered when GetLogFilenames is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetLogFilenamesRequested(cb: (event: HandleGetLogFilenamesRequestedEvent) => void): void;

    interface HandleGetLogFilenamesRequestedEvent {
        data: string;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetLogFilenamesEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetLogFilenamesEvent is triggered after processing GetLogFilenames.
     */
    function onHandleGetLogFilenames(cb: (event: HandleGetLogFilenamesEvent) => void): void;

    interface HandleGetLogFilenamesEvent {
        data: string;

        next(): void;
    }

    /**
     * @event HandleDeleteLogsRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleDeleteLogsRequestedEvent is triggered when DeleteLogs is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleDeleteLogsRequested(cb: (event: HandleDeleteLogsRequestedEvent) => void): void;

    interface HandleDeleteLogsRequestedEvent {
        filenames?: Array<string>;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetLatestLogContentRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetLatestLogContentRequestedEvent is triggered when GetLatestLogContent is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetLatestLogContentRequested(cb: (event: HandleGetLatestLogContentRequestedEvent) => void): void;

    interface HandleGetLatestLogContentRequestedEvent {
        data: string;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetLatestLogContentEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetLatestLogContentEvent is triggered after processing GetLatestLogContent.
     */
    function onHandleGetLatestLogContent(cb: (event: HandleGetLatestLogContentEvent) => void): void;

    interface HandleGetLatestLogContentEvent {
        data: string;

        next(): void;
    }

    /**
     * @event HandleSyncGetTrackedMediaItemsRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleSyncGetTrackedMediaItemsRequestedEvent is triggered when SyncGetTrackedMediaItems is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleSyncGetTrackedMediaItemsRequested(cb: (event: HandleSyncGetTrackedMediaItemsRequestedEvent) => void): void;

    interface HandleSyncGetTrackedMediaItemsRequestedEvent {
        data?: Sync_TrackedMediaItem;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleSyncGetTrackedMediaItemsEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleSyncGetTrackedMediaItemsEvent is triggered after processing SyncGetTrackedMediaItems.
     */
    function onHandleSyncGetTrackedMediaItems(cb: (event: HandleSyncGetTrackedMediaItemsEvent) => void): void;

    interface HandleSyncGetTrackedMediaItemsEvent {
        data?: Sync_TrackedMediaItem;

        next(): void;
    }

    /**
     * @event HandleSyncAddMediaRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleSyncAddMediaRequestedEvent is triggered when SyncAddMedia is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleSyncAddMediaRequested(cb: (event: HandleSyncAddMediaRequestedEvent) => void): void;

    interface HandleSyncAddMediaRequestedEvent {
        media?: Array<{ mediaId: number; type: string; }>;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleSyncRemoveMediaRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleSyncRemoveMediaRequestedEvent is triggered when SyncRemoveMedia is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleSyncRemoveMediaRequested(cb: (event: HandleSyncRemoveMediaRequestedEvent) => void): void;

    interface HandleSyncRemoveMediaRequestedEvent {
        mediaId: number;
        type: string;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleSyncGetIsMediaTrackedRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleSyncGetIsMediaTrackedRequestedEvent is triggered when SyncGetIsMediaTracked is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleSyncGetIsMediaTrackedRequested(cb: (event: HandleSyncGetIsMediaTrackedRequestedEvent) => void): void;

    interface HandleSyncGetIsMediaTrackedRequestedEvent {
        id: number;
        type: string;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleSyncLocalDataRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleSyncLocalDataRequestedEvent is triggered when SyncLocalData is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleSyncLocalDataRequested(cb: (event: HandleSyncLocalDataRequestedEvent) => void): void;

    interface HandleSyncLocalDataRequestedEvent {
        next(): void;

        preventDefault(): void;

    }

    /**
     * @event HandleSyncGetQueueStateRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleSyncGetQueueStateRequestedEvent is triggered when SyncGetQueueState is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleSyncGetQueueStateRequested(cb: (event: HandleSyncGetQueueStateRequestedEvent) => void): void;

    interface HandleSyncGetQueueStateRequestedEvent {
        data?: Sync_QueueState;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleSyncGetQueueStateEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleSyncGetQueueStateEvent is triggered after processing SyncGetQueueState.
     */
    function onHandleSyncGetQueueState(cb: (event: HandleSyncGetQueueStateEvent) => void): void;

    interface HandleSyncGetQueueStateEvent {
        data?: Sync_QueueState;

        next(): void;
    }

    /**
     * @event HandleSyncAnilistDataRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleSyncAnilistDataRequestedEvent is triggered when SyncAnilistData is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleSyncAnilistDataRequested(cb: (event: HandleSyncAnilistDataRequestedEvent) => void): void;

    interface HandleSyncAnilistDataRequestedEvent {
        next(): void;

        preventDefault(): void;

    }

    /**
     * @event HandleSyncSetHasLocalChangesRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleSyncSetHasLocalChangesRequestedEvent is triggered when SyncSetHasLocalChanges is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleSyncSetHasLocalChangesRequested(cb: (event: HandleSyncSetHasLocalChangesRequestedEvent) => void): void;

    interface HandleSyncSetHasLocalChangesRequestedEvent {
        updated: boolean;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleSyncGetHasLocalChangesRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleSyncGetHasLocalChangesRequestedEvent is triggered when SyncGetHasLocalChanges is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleSyncGetHasLocalChangesRequested(cb: (event: HandleSyncGetHasLocalChangesRequestedEvent) => void): void;

    interface HandleSyncGetHasLocalChangesRequestedEvent {
        next(): void;

        preventDefault(): void;

    }

    /**
     * @event HandleSyncGetLocalStorageSizeRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleSyncGetLocalStorageSizeRequestedEvent is triggered when SyncGetLocalStorageSize is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleSyncGetLocalStorageSizeRequested(cb: (event: HandleSyncGetLocalStorageSizeRequestedEvent) => void): void;

    interface HandleSyncGetLocalStorageSizeRequestedEvent {
        data: string;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleSyncGetLocalStorageSizeEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleSyncGetLocalStorageSizeEvent is triggered after processing SyncGetLocalStorageSize.
     */
    function onHandleSyncGetLocalStorageSize(cb: (event: HandleSyncGetLocalStorageSizeEvent) => void): void;

    interface HandleSyncGetLocalStorageSizeEvent {
        data: string;

        next(): void;
    }

    /**
     * @event HandleGetThemeRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetThemeRequestedEvent is triggered when GetTheme is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetThemeRequested(cb: (event: HandleGetThemeRequestedEvent) => void): void;

    interface HandleGetThemeRequestedEvent {
        data?: Models_Theme;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetThemeEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetThemeEvent is triggered after processing GetTheme.
     */
    function onHandleGetTheme(cb: (event: HandleGetThemeEvent) => void): void;

    interface HandleGetThemeEvent {
        data?: Models_Theme;

        next(): void;
    }

    /**
     * @event HandleUpdateThemeRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleUpdateThemeRequestedEvent is triggered when UpdateTheme is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleUpdateThemeRequested(cb: (event: HandleUpdateThemeRequestedEvent) => void): void;

    interface HandleUpdateThemeRequestedEvent {
        theme?: Models_Theme;
        data?: Models_Theme;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleUpdateThemeEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleUpdateThemeEvent is triggered after processing UpdateTheme.
     */
    function onHandleUpdateTheme(cb: (event: HandleUpdateThemeEvent) => void): void;

    interface HandleUpdateThemeEvent {
        data?: Models_Theme;

        next(): void;
    }

    /**
     * @event HandleGetActiveTorrentListRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetActiveTorrentListRequestedEvent is triggered when GetActiveTorrentList is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetActiveTorrentListRequested(cb: (event: HandleGetActiveTorrentListRequestedEvent) => void): void;

    interface HandleGetActiveTorrentListRequestedEvent {
        data?: TorrentClient_Torrent;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetActiveTorrentListEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetActiveTorrentListEvent is triggered after processing GetActiveTorrentList.
     */
    function onHandleGetActiveTorrentList(cb: (event: HandleGetActiveTorrentListEvent) => void): void;

    interface HandleGetActiveTorrentListEvent {
        data?: TorrentClient_Torrent;

        next(): void;
    }

    /**
     * @event HandleTorrentClientActionRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleTorrentClientActionRequestedEvent is triggered when TorrentClientAction is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleTorrentClientActionRequested(cb: (event: HandleTorrentClientActionRequestedEvent) => void): void;

    interface HandleTorrentClientActionRequestedEvent {
        hash: string;
        action: string;
        dir: string;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleTorrentClientDownloadRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleTorrentClientDownloadRequestedEvent is triggered when TorrentClientDownload is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleTorrentClientDownloadRequested(cb: (event: HandleTorrentClientDownloadRequestedEvent) => void): void;

    interface HandleTorrentClientDownloadRequestedEvent {
        torrents?: Array<AnimeTorrent>;
        destination: string;
        smartSelect?: { enabled: boolean; missingEpisodeNumbers: Array<number>; };
        media?: AL_BaseAnime;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleTorrentClientAddMagnetFromRuleRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleTorrentClientAddMagnetFromRuleRequestedEvent is triggered when TorrentClientAddMagnetFromRule is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleTorrentClientAddMagnetFromRuleRequested(cb: (event: HandleTorrentClientAddMagnetFromRuleRequestedEvent) => void): void;

    interface HandleTorrentClientAddMagnetFromRuleRequestedEvent {
        magnetUrl: string;
        ruleId: number;
        queuedItemId: number;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleSearchTorrentRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleSearchTorrentRequestedEvent is triggered when SearchTorrent is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleSearchTorrentRequested(cb: (event: HandleSearchTorrentRequestedEvent) => void): void;

    interface HandleSearchTorrentRequestedEvent {
        type: string;
        provider: string;
        query: string;
        episodeNumber: number;
        batch: boolean;
        media?: AL_BaseAnime;
        absoluteOffset: number;
        resolution: string;
        bestRelease: boolean;
        data?: Torrent_SearchData;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleSearchTorrentEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleSearchTorrentEvent is triggered after processing SearchTorrent.
     */
    function onHandleSearchTorrent(cb: (event: HandleSearchTorrentEvent) => void): void;

    interface HandleSearchTorrentEvent {
        data?: Torrent_SearchData;

        next(): void;
    }

    /**
     * @event HandleGetTorrentstreamEpisodeCollectionRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetTorrentstreamEpisodeCollectionRequestedEvent is triggered when GetTorrentstreamEpisodeCollection is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetTorrentstreamEpisodeCollectionRequested(cb: (event: HandleGetTorrentstreamEpisodeCollectionRequestedEvent) => void): void;

    interface HandleGetTorrentstreamEpisodeCollectionRequestedEvent {
        id: number;
        data?: Torrentstream_EpisodeCollection;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetTorrentstreamEpisodeCollectionEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetTorrentstreamEpisodeCollectionEvent is triggered after processing GetTorrentstreamEpisodeCollection.
     */
    function onHandleGetTorrentstreamEpisodeCollection(cb: (event: HandleGetTorrentstreamEpisodeCollectionEvent) => void): void;

    interface HandleGetTorrentstreamEpisodeCollectionEvent {
        data?: Torrentstream_EpisodeCollection;

        next(): void;
    }

    /**
     * @event HandleGetTorrentstreamSettingsRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetTorrentstreamSettingsRequestedEvent is triggered when GetTorrentstreamSettings is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetTorrentstreamSettingsRequested(cb: (event: HandleGetTorrentstreamSettingsRequestedEvent) => void): void;

    interface HandleGetTorrentstreamSettingsRequestedEvent {
        data?: Models_TorrentstreamSettings;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetTorrentstreamSettingsEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetTorrentstreamSettingsEvent is triggered after processing GetTorrentstreamSettings.
     */
    function onHandleGetTorrentstreamSettings(cb: (event: HandleGetTorrentstreamSettingsEvent) => void): void;

    interface HandleGetTorrentstreamSettingsEvent {
        data?: Models_TorrentstreamSettings;

        next(): void;
    }

    /**
     * @event HandleSaveTorrentstreamSettingsRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleSaveTorrentstreamSettingsRequestedEvent is triggered when SaveTorrentstreamSettings is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleSaveTorrentstreamSettingsRequested(cb: (event: HandleSaveTorrentstreamSettingsRequestedEvent) => void): void;

    interface HandleSaveTorrentstreamSettingsRequestedEvent {
        settings?: Models_TorrentstreamSettings;
        data?: Models_TorrentstreamSettings;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleSaveTorrentstreamSettingsEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleSaveTorrentstreamSettingsEvent is triggered after processing SaveTorrentstreamSettings.
     */
    function onHandleSaveTorrentstreamSettings(cb: (event: HandleSaveTorrentstreamSettingsEvent) => void): void;

    interface HandleSaveTorrentstreamSettingsEvent {
        data?: Models_TorrentstreamSettings;

        next(): void;
    }

    /**
     * @event HandleGetTorrentstreamTorrentFilePreviewsRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetTorrentstreamTorrentFilePreviewsRequestedEvent is triggered when GetTorrentstreamTorrentFilePreviews is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetTorrentstreamTorrentFilePreviewsRequested(cb: (event: HandleGetTorrentstreamTorrentFilePreviewsRequestedEvent) => void): void;

    interface HandleGetTorrentstreamTorrentFilePreviewsRequestedEvent {
        torrent?: HibikeTorrent_AnimeTorrent;
        episodeNumber: number;
        media?: AL_BaseAnime;
        data?: Torrentstream_FilePreview;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetTorrentstreamTorrentFilePreviewsEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetTorrentstreamTorrentFilePreviewsEvent is triggered after processing GetTorrentstreamTorrentFilePreviews.
     */
    function onHandleGetTorrentstreamTorrentFilePreviews(cb: (event: HandleGetTorrentstreamTorrentFilePreviewsEvent) => void): void;

    interface HandleGetTorrentstreamTorrentFilePreviewsEvent {
        data?: Torrentstream_FilePreview;

        next(): void;
    }

    /**
     * @event HandleTorrentstreamStartStreamRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleTorrentstreamStartStreamRequestedEvent is triggered when TorrentstreamStartStream is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleTorrentstreamStartStreamRequested(cb: (event: HandleTorrentstreamStartStreamRequestedEvent) => void): void;

    interface HandleTorrentstreamStartStreamRequestedEvent {
        mediaId: number;
        episodeNumber: number;
        aniDBEpisode: string;
        autoSelect: boolean;
        torrent?: HibikeTorrent_AnimeTorrent;
        fileIndex: number;
        playbackType?: Torrentstream_PlaybackType;
        clientId: string;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleTorrentstreamStopStreamRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleTorrentstreamStopStreamRequestedEvent is triggered when TorrentstreamStopStream is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleTorrentstreamStopStreamRequested(cb: (event: HandleTorrentstreamStopStreamRequestedEvent) => void): void;

    interface HandleTorrentstreamStopStreamRequestedEvent {
        next(): void;

        preventDefault(): void;

    }

    /**
     * @event HandleTorrentstreamDropTorrentRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleTorrentstreamDropTorrentRequestedEvent is triggered when TorrentstreamDropTorrent is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleTorrentstreamDropTorrentRequested(cb: (event: HandleTorrentstreamDropTorrentRequestedEvent) => void): void;

    interface HandleTorrentstreamDropTorrentRequestedEvent {
        next(): void;

        preventDefault(): void;

    }

    /**
     * @event HandleGetTorrentstreamBatchHistoryRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetTorrentstreamBatchHistoryRequestedEvent is triggered when GetTorrentstreamBatchHistory is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleGetTorrentstreamBatchHistoryRequested(cb: (event: HandleGetTorrentstreamBatchHistoryRequestedEvent) => void): void;

    interface HandleGetTorrentstreamBatchHistoryRequestedEvent {
        mediaId: number;
        data?: Torrentstream_BatchHistoryResponse;

        next(): void;

        preventDefault(): void;
    }

    /**
     * @event HandleGetTorrentstreamBatchHistoryEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleGetTorrentstreamBatchHistoryEvent is triggered after processing GetTorrentstreamBatchHistory.
     */
    function onHandleGetTorrentstreamBatchHistory(cb: (event: HandleGetTorrentstreamBatchHistoryEvent) => void): void;

    interface HandleGetTorrentstreamBatchHistoryEvent {
        data?: Torrentstream_BatchHistoryResponse;

        next(): void;
    }

    /**
     * @event HandleTorrentstreamServeStreamRequestedEvent
     * @file internal/handlers/hook_events.go
     * @description
     * HandleTorrentstreamServeStreamRequestedEvent is triggered when TorrentstreamServeStream is requested.
     * Prevent default to skip the default behavior and provide your own implementation.
     */
    function onHandleTorrentstreamServeStreamRequested(cb: (event: HandleTorrentstreamServeStreamRequestedEvent) => void): void;

    interface HandleTorrentstreamServeStreamRequestedEvent {
        next(): void;

        preventDefault(): void;

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
     * - Filepath: internal/api/anilist/stats.go
     */
    interface AL_AnimeStats {
        count: number;
        minutesWatched: number;
        episodesWatched: number;
        meanScore: number;
        genres?: Array<AL_UserGenreStats>;
        formats?: Array<AL_UserFormatStats>;
        statuses?: Array<AL_UserStatusStats>;
        studios?: Array<AL_UserStudioStats>;
        scores?: Array<AL_UserScoreStats>;
        startYears?: Array<AL_UserStartYearStats>;
        releaseYears?: Array<AL_UserReleaseYearStats>;
    }

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
    interface AL_GetViewer_Viewer {
        name: string;
        avatar?: AL_GetViewer_Viewer_Avatar;
        bannerImage?: string;
        isBlocked?: boolean;
        options?: AL_GetViewer_Viewer_Options;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_GetViewer_Viewer_Avatar {
        large?: string;
        medium?: string;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_GetViewer_Viewer_Options {
        displayAdultContent?: boolean;
        airingNotifications?: boolean;
        profileColor?: string;
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
     * - Filepath: internal/api/anilist/manga.go
     */
    export type AL_MangaListEntry = AL_MangaCollection_MediaListCollection_Lists_Entries;

    /**
     * - Filepath: internal/api/anilist/stats.go
     */
    interface AL_MangaStats {
        count: number;
        chaptersRead: number;
        meanScore: number;
        genres?: Array<AL_UserGenreStats>;
        statuses?: Array<AL_UserStatusStats>;
        scores?: Array<AL_UserScoreStats>;
        startYears?: Array<AL_UserStartYearStats>;
        releaseYears?: Array<AL_UserReleaseYearStats>;
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
     * - Filepath: internal/api/anilist/stats.go
     */
    interface AL_Stats {
        animeStats?: AL_AnimeStats;
        mangaStats?: AL_MangaStats;
    }

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
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_UserFormatStats {
        format?: AL_MediaFormat;
        meanScore: number;
        count: number;
        minutesWatched: number;
        mediaIds?: Array<number>;
        chaptersRead: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_UserGenreStats {
        genre?: string;
        meanScore: number;
        count: number;
        minutesWatched: number;
        mediaIds?: Array<number>;
        chaptersRead: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_UserReleaseYearStats {
        releaseYear?: number;
        meanScore: number;
        count: number;
        minutesWatched: number;
        mediaIds?: Array<number>;
        chaptersRead: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_UserScoreStats {
        score?: number;
        meanScore: number;
        count: number;
        minutesWatched: number;
        mediaIds?: Array<number>;
        chaptersRead: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_UserStartYearStats {
        startYear?: number;
        meanScore: number;
        count: number;
        minutesWatched: number;
        mediaIds?: Array<number>;
        chaptersRead: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_UserStatusStats {
        status?: AL_MediaListStatus;
        meanScore: number;
        count: number;
        minutesWatched: number;
        mediaIds?: Array<number>;
        chaptersRead: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_UserStudioStats {
        studio?: AL_UserStudioStats_Studio;
        meanScore: number;
        count: number;
        minutesWatched: number;
        mediaIds?: Array<number>;
        chaptersRead: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     */
    interface AL_UserStudioStats_Studio {
        id: number;
        name: string;
        isAnimationStudio: boolean;
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
     * - Filepath: internal/library/anime/playlist.go
     */
    interface Anime_Playlist {
        /**
         * DbId is the database ID of the models.PlaylistEntry
         */
        dbId: number;
        /**
         * Name is the name of the playlist
         */
        name: string;
        /**
         * LocalFiles is a list of local files in the playlist, in order
         */
        localFiles?: Array<Anime_LocalFile>;
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
     * - Filepath: internal/library/anime/user.go
     */
    interface Anime_User {
        viewer?: AL_GetViewer_Viewer;
        token: string;
    }

    /**
     * - Filepath: internal/handlers/docs.go
     */
    interface ApiDocsGroup {
        filename: string;
        name: string;
        handlers?: Array<RouteHandler>;
    }

    /**
     * - Filepath: internal/mediastream/videofile/info.go
     */
    interface Audio {
        index: number;
        title?: string;
        language?: string;
        codec: string;
        mimeCodec?: string;
        isDefault: boolean;
        isForced: boolean;
        channels: number;
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
     * - Filepath: internal/mediastream/videofile/info.go
     */
    interface Chapter {
        startTime: number;
        endTime: number;
        name: string;
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
     * - Filepath: internal/debrid/client/stream.go
     */
    interface DebridClient_CancelStreamOptions {
        removeTorrent: boolean;
    }

    /**
     * - Filepath: internal/debrid/client/previews.go
     */
    interface DebridClient_FilePreview {
        path: string;
        displayPath: string;
        displayTitle: string;
        episodeNumber: number;
        relativeEpisodeNumber: number;
        isLikely: boolean;
        index: number;
        fileId: string;
    }

    /**
     * - Filepath: internal/debrid/client/stream.go
     */
    export type DebridClient_StreamPlaybackType = "default" | "externalPlayerLink";

    /**
     * - Filepath: internal/debrid/debrid/debrid.go
     */
    interface Debrid_CachedFile {
        size: number;
        name: string;
    }

    /**
     * - Filepath: internal/debrid/debrid/debrid.go
     */
    interface Debrid_TorrentInfo {
        /**
         * ID of the torrent if added to the debrid service
         */
        id?: string;
        name: string;
        hash: string;
        size: number;
        files?: Array<Debrid_TorrentItemFile>;
    }

    /**
     * - Filepath: internal/debrid/debrid/debrid.go
     */
    interface Debrid_TorrentItem {
        id: string;
        /**
         * Name of the torrent or file
         */
        name: string;
        /**
         * SHA1 hash of the torrent
         */
        hash: string;
        /**
         * Size of the selected files (size in bytes)
         */
        size: number;
        /**
         * Formatted size of the selected files
         */
        formattedSize: string;
        /**
         * Progress percentage (0 to 100)
         */
        completionPercentage: number;
        /**
         * Formatted estimated time remaining
         */
        eta: string;
        /**
         * Current download status
         */
        status: Debrid_TorrentItemStatus;
        /**
         * Date when the torrent was added, RFC3339 format
         */
        added: string;
        /**
         * Current download speed (optional, present in downloading state)
         */
        speed?: string;
        /**
         * Number of seeders (optional, present in downloading state)
         */
        seeders?: number;
        /**
         * Whether the torrent is ready to be downloaded
         */
        isReady: boolean;
        /**
         * List of files in the torrent (optional)
         */
        files?: Array<Debrid_TorrentItemFile>;
    }

    /**
     * - Filepath: internal/debrid/debrid/debrid.go
     */
    interface Debrid_TorrentItemFile {
        /**
         * ID of the file, usually the index
         */
        id: string;
        index: number;
        name: string;
        path: string;
        size: number;
    }

    /**
     * - Filepath: internal/debrid/debrid/debrid.go
     */
    interface Debrid_TorrentItemInstantAvailability {
        /**
         * Key is the file ID (or index)
         */
        cachedFiles?: Record<string, Debrid_CachedFile>;
    }

    /**
     * - Filepath: internal/debrid/debrid/debrid.go
     */
    export type Debrid_TorrentItemStatus = "downloading" |
        "completed" |
        "seeding" |
        "error" |
        "stalled" |
        "paused" |
        "other";

    /**
     * - Filepath: internal/handlers/directory_selector.go
     */
    interface DirectoryInfo {
        fullPath: string;
        folderName: string;
    }

    /**
     * - Filepath: internal/handlers/directory_selector.go
     */
    interface DirectorySelectorResponse {
        fullPath: string;
        exists: boolean;
        basePath: string;
        suggestions?: Array<DirectoryInfo>;
        content?: Array<DirectoryInfo>;
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
     * - Filepath: internal/handlers/download.go
     */
    interface DownloadReleaseResponse {
        destination: string;
        error?: string;
    }

    /**
     * - Filepath: internal/extension_repo/repository.go
     */
    interface ExtensionRepo_AllExtensions {
        extensions?: Array<Extension_Extension>;
        invalidExtensions?: Array<Extension_InvalidExtension>;
        invalidUserConfigExtensions?: Array<Extension_InvalidExtension>;
        hasUpdate?: Array<ExtensionRepo_UpdateData>;
    }

    /**
     * - Filepath: internal/extension_repo/repository.go
     */
    interface ExtensionRepo_AnimeTorrentProviderExtensionItem {
        id: string;
        name: string;
        /**
         * ISO 639-1 language code
         */
        lang: string;
        settings?: HibikeTorrent_AnimeProviderSettings;
    }

    /**
     * - Filepath: internal/extension_repo/external.go
     */
    interface ExtensionRepo_ExtensionInstallResponse {
        message: string;
    }

    /**
     * - Filepath: internal/extension_repo/userconfig.go
     */
    interface ExtensionRepo_ExtensionUserConfig {
        userConfig?: Extension_UserConfig;
        savedUserConfig?: Extension_SavedUserConfig;
    }

    /**
     * - Filepath: internal/extension_repo/repository.go
     */
    interface ExtensionRepo_MangaProviderExtensionItem {
        id: string;
        name: string;
        /**
         * ISO 639-1 language code
         */
        lang: string;
        settings?: HibikeManga_Settings;
    }

    /**
     * - Filepath: internal/extension_repo/repository.go
     */
    interface ExtensionRepo_OnlinestreamProviderExtensionItem {
        id: string;
        name: string;
        /**
         * ISO 639-1 language code
         */
        lang: string;
        episodeServers?: Array<string>;
        supportsDub: boolean;
    }

    /**
     * - Filepath: internal/extension_repo/external_plugin.go
     */
    interface ExtensionRepo_StoredPluginSettingsData {
        pinnedTrayPluginIds?: Array<string>;
        /**
         * Extension ID -> Permission Hash
         */
        pluginGrantedPermissions?: Record<string, string>;
    }

    /**
     * - Filepath: internal/extension_repo/repository.go
     */
    interface ExtensionRepo_UpdateData {
        extensionID: string;
        manifestURI: string;
        version: string;
    }

    /**
     * - Filepath: internal/extension/plugin.go
     * @description
     *  CommandArg represents an argument for a command
     */
    interface Extension_CommandArg {
        value?: string;
        validator?: string;
    }

    /**
     * - Filepath: internal/extension/plugin.go
     * @description
     *  CommandScope defines a specific command or set of commands that can be executed
     *  with specific arguments and validation rules.
     */
    interface Extension_CommandScope {
        description?: string;
        command: string;
        args?: Array<Extension_CommandArg>;
    }

    /**
     * - Filepath: internal/extension/extension.go
     */
    interface Extension_ConfigField {
        type: Extension_ConfigFieldType;
        name: string;
        label: string;
        options?: Array<Extension_ConfigFieldSelectOption>;
        default?: string;
    }

    /**
     * - Filepath: internal/extension/extension.go
     */
    interface Extension_ConfigFieldSelectOption {
        value: string;
        label: string;
    }

    /**
     * - Filepath: internal/extension/extension.go
     */
    export type Extension_ConfigFieldType = "text" | "switch" | "select";

    /**
     * - Filepath: internal/extension/extension.go
     */
    interface Extension_Extension {
        /**
         * e.g. "extension-example"
         */
        id: string;
        /**
         * e.g. "Extension"
         */
        name: string;
        /**
         * e.g. "1.0.0"
         */
        version: string;
        semverConstraint?: string;
        /**
         * e.g. "http://cdn.something.app/extensions/extension-example/manifest.json"
         */
        manifestURI: string;
        /**
         * e.g. "go"
         */
        language: Extension_Language;
        /**
         * e.g. "anime-torrent-provider"
         */
        type: Extension_Type;
        /**
         * e.g. "This extension provides torrents"
         */
        description: string;
        /**
         * e.g. "Seanime"
         */
        author: string;
        icon: string;
        website: string;
        lang: string;
        /**
         * NOT IMPLEMENTED
         */
        permissions?: Array<string>;
        userConfig?: Extension_UserConfig;
        payload: string;
        payloadURI?: string;
        plugin?: Extension_PluginManifest;
        isDevelopment?: boolean;
        /**
         * Contains the saved user config for the extension
         */?: Extension_SavedUserConfig;
    }

    /**
     * - Filepath: internal/extension/extension.go
     */
    interface Extension_InvalidExtension {
        id: string;
        path: string;
        extension: Extension_Extension;
        reason: string;
        code: Extension_InvalidExtensionErrorCode;
        pluginPermissionDescription?: string;
    }

    /**
     * - Filepath: internal/extension/extension.go
     */
    export type Extension_InvalidExtensionErrorCode = "invalid_manifest" |
        "invalid_payload" |
        "user_config_error" |
        "invalid_authorization" |
        "plugin_permissions_not_granted" |
        "invalid_semver_constraint";

    /**
     * - Filepath: internal/extension/extension.go
     */
    export type Extension_Language = "javascript" | "typescript" | "go";

    /**
     * - Filepath: internal/extension/plugin.go
     * @description
     *  PluginAllowlist is a list of system permissions that the plugin is asking for.
     *
     *  The user must acknowledge these permissions before the plugin can be loaded.
     */
    interface Extension_PluginAllowlist {
        readPaths?: Array<string>;
        writePaths?: Array<string>;
        commandScopes?: Array<Extension_CommandScope>;
    }

    /**
     * - Filepath: internal/extension/plugin.go
     */
    interface Extension_PluginManifest {
        version: string;
        permissions?: Extension_PluginPermissions;
    }

    /**
     * - Filepath: internal/extension/extension.go
     */
    export type Extension_PluginPermissionScope = string;

    /**
     * - Filepath: internal/extension/plugin.go
     */
    interface Extension_PluginPermissions {
        scopes?: Array<Extension_PluginPermissionScope>;
        allow?: Extension_PluginAllowlist;
    }

    /**
     * - Filepath: internal/extension/extension.go
     */
    interface Extension_SavedUserConfig {
        version: number;
        values?: Record<string, string>;
    }

    /**
     * - Filepath: internal/extension/extension.go
     */
    export type Extension_Type = "anime-torrent-provider" | "manga-provider" | "onlinestream-provider" | "plugin";

    /**
     * - Filepath: internal/extension/extension.go
     */
    interface Extension_UserConfig {
        version: number;
        requiresConfig: boolean;
        fields?: Array<Extension_ConfigField>;
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
    }

    /**
     * - Filepath: internal/extension/hibike/manga/types.go
     */
    interface HibikeManga_ChapterPage {
        provider: string;
        url: string;
        index: number;
        headers?: Record<string, string>;
    }

    /**
     * - Filepath: internal/extension/hibike/manga/types.go
     */
    interface HibikeManga_SearchResult {
        provider: string;
        id: string;
        title: string;
        synonyms?: Array<string>;
        year?: number;
        image?: string;
        searchRating?: number;
    }

    /**
     * - Filepath: internal/extension/hibike/manga/types.go
     */
    interface HibikeManga_Settings {
        supportsMultiScanlator: boolean;
        supportsMultiLanguage: boolean;
    }

    /**
     * - Filepath: internal/extension/hibike/onlinestream/types.go
     */
    interface HibikeOnlinestream_SearchResult {
        id: string;
        title: string;
        url: string;
        subOrDub: HibikeOnlinestream_SubOrDub;
    }

    /**
     * - Filepath: internal/extension/hibike/onlinestream/types.go
     */
    export type HibikeOnlinestream_SubOrDub = "sub" | "dub" | "both";

    /**
     * - Filepath: internal/extension/hibike/torrent/types.go
     */
    interface HibikeTorrent_AnimeProviderSettings {
        canSmartSearch: boolean;
        smartSearchFilters?: Array<HibikeTorrent_AnimeProviderSmartSearchFilter>;
        supportsAdult: boolean;
        type: HibikeTorrent_AnimeProviderType;
    }

    /**
     * - Filepath: internal/extension/hibike/torrent/types.go
     */
    export type HibikeTorrent_AnimeProviderSmartSearchFilter = "batch" | "episodeNumber" | "resolution" | "query" | "bestReleases";

    /**
     * - Filepath: internal/extension/hibike/torrent/types.go
     */
    export type HibikeTorrent_AnimeProviderType = "main" | "special";

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
     * - Filepath: internal/core/feature_flags.go
     */
    interface INTERNAL_FeatureFlags {
        MainServerTorrentStreaming: boolean;
    }

    /**
     * - Filepath: internal/handlers/mal.go
     */
    interface MalAuthResponse {
        access_token: string;
        refresh_token: string;
        expires_in: number;
        token_type: string;
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
     * - Filepath: internal/manga/download.go
     */
    interface Manga_DownloadListItem {
        mediaId: number;
        media?: AL_BaseManga;
        downloadData: Manga_ProviderDownloadMap;
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
     * - Filepath: internal/manga/chapter_container_mapping.go
     */
    interface Manga_MappingResponse {
        mangaId?: string;
    }

    /**
     * - Filepath: internal/manga/download.go
     */
    interface Manga_MediaDownloadData {
        downloaded: Manga_ProviderDownloadMap;
        queued: Manga_ProviderDownloadMap;
    }

    /**
     * - Filepath: internal/manga/download.go
     */
    export type Manga_MediaMap = Record<number, Manga_ProviderDownloadMap>;

    /**
     * - Filepath: internal/manga/chapter_page_container.go
     */
    interface Manga_PageContainer {
        mediaId: number;
        provider: string;
        chapterId: string;
        pages?: Array<HibikeManga_ChapterPage>;
        /**
         * Indexed by page number
         */
        pageDimensions?: Record<number, Manga_PageDimension>;
        /**
         * TODO remove
         */
        isDownloaded: boolean;
    }

    /**
     * - Filepath: internal/manga/chapter_page_container.go
     */
    interface Manga_PageDimension {
        width: number;
        height: number;
    }

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
     * - Filepath: internal/mediastream/videofile/info.go
     */
    interface MediaInfo {
        ready: any;
        sha: string;
        path: string;
        extension: string;
        mimeCodec?: string;
        size: number;
        duration: number;
        container?: string;
        video?: Video;
        videos?: Array<Video>;
        audios?: Array<Audio>;
        subtitles?: Array<Subtitle>;
        fonts?: Array<string>;
        chapters?: Array<Chapter>;
    }

    /**
     * - Filepath: internal/mediastream/playback.go
     */
    interface Mediastream_MediaContainer {
        filePath: string;
        hash: string;
        /**
         * Tells the frontend how to play the media.
         */
        streamType: Mediastream_StreamType;
        /**
         * The relative endpoint to stream the media.
         */
        streamUrl: string;
        mediaInfo?: MediaInfo;
    }

    /**
     * - Filepath: internal/mediastream/playback.go
     */
    export type Mediastream_StreamType = "transcode" | "optimized" | "direct";

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
    }

    /**
     * - Filepath: internal/database/models/models.go
     */
    interface Models_AnilistSettings {
        hideAudienceScore: boolean;
        enableAdultContent: boolean;
        blurAdultContent: boolean;
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
     * - Filepath: internal/database/models/models.go
     */
    interface Models_ChapterDownloadQueueItem {
        provider: string;
        mediaId: number;
        chapterId: string;
        chapterNumber: string;
        /**
         * Contains map of page index to page details
         */
        pageData?: Array<string>;
        status: string;
        id: number;
        createdAt?: string;
        updatedAt?: string;
    }

    /**
     * - Filepath: internal/database/models/models.go
     */
    interface Models_DebridSettings {
        enabled: boolean;
        provider: string;
        apiKey: string;
        includeDebridStreamInLibrary: boolean;
        streamAutoSelect: boolean;
        streamPreferredResolution: string;
        id: number;
        createdAt?: string;
        updatedAt?: string;
    }

    /**
     * - Filepath: internal/database/models/models.go
     */
    interface Models_DiscordSettings {
        enableRichPresence: boolean;
        enableAnimeRichPresence: boolean;
        enableMangaRichPresence: boolean;
        richPresenceHideSeanimeRepositoryButton: boolean;
        richPresenceShowAniListMediaButton: boolean;
        richPresenceShowAniListProfileButton: boolean;
    }

    /**
     * - Filepath: internal/database/models/models.go
     */
    export type Models_LibraryPaths = Array<string>;

    /**
     * - Filepath: internal/database/models/models.go
     */
    interface Models_LibrarySettings {
        libraryPath: string;
        autoUpdateProgress: boolean;
        disableUpdateCheck: boolean;
        torrentProvider: string;
        autoScan: boolean;
        enableOnlinestream: boolean;
        includeOnlineStreamingInLibrary: boolean;
        disableAnimeCardTrailers: boolean;
        enableManga: boolean;
        dohProvider: string;
        openTorrentClientOnStart: boolean;
        openWebURLOnStart: boolean;
        refreshLibraryOnStart: boolean;
        autoPlayNextEpisode: boolean;
        enableWatchContinuity: boolean;
        libraryPaths: Models_LibraryPaths;
        autoSyncOfflineLocalData: boolean;
        scannerMatchingThreshold: number;
        scannerMatchingAlgorithm: string;
    }

    /**
     * - Filepath: internal/database/models/models.go
     */
    interface Models_ListSyncSettings {
        automatic: boolean;
        origin: string;
    }

    /**
     * - Filepath: internal/database/models/models.go
     */
    interface Models_MangaSettings {
        defaultMangaProvider: string;
        mangaAutoUpdateProgress: boolean;
    }

    /**
     * - Filepath: internal/database/models/models.go
     */
    interface Models_MediaPlayerSettings {
        /**
         * "vlc" or "mpc-hc"
         */
        defaultPlayer: string;
        host: string;
        vlcUsername: string;
        vlcPassword: string;
        vlcPort: number;
        vlcPath: string;
        mpcPort: number;
        mpcPath: string;
        mpvSocket: string;
        mpvPath: string;
    }

    /**
     * - Filepath: internal/database/models/models.go
     */
    interface Models_MediastreamSettings {
        transcodeEnabled: boolean;
        transcodeHwAccel: string;
        transcodeThreads: number;
        transcodePreset: string;
        disableAutoSwitchToDirectPlay: boolean;
        directPlayOnly: boolean;
        preTranscodeEnabled: boolean;
        preTranscodeLibraryDir: string;
        ffmpegPath: string;
        ffprobePath: string;
        transcodeHwAccelCustomSettings: string;
        id: number;
        createdAt?: string;
        updatedAt?: string;
    }

    /**
     * - Filepath: internal/database/models/models.go
     */
    interface Models_NotificationSettings {
        disableNotifications: boolean;
        disableAutoDownloaderNotifications: boolean;
        disableAutoScannerNotifications: boolean;
    }

    /**
     * - Filepath: internal/database/models/models.go
     */
    interface Models_Settings {
        library?: Models_LibrarySettings;
        mediaPlayer?: Models_MediaPlayerSettings;
        torrent?: Models_TorrentSettings;
        manga?: Models_MangaSettings;
        anilist?: Models_AnilistSettings;
        listSync?: Models_ListSyncSettings;
        autoDownloader?: Models_AutoDownloaderSettings;
        discord?: Models_DiscordSettings;
        notifications?: Models_NotificationSettings;
        id: number;
        createdAt?: string;
        updatedAt?: string;
    }

    /**
     * - Filepath: internal/database/models/models.go
     */
    interface Models_SilencedMediaEntry {
        id: number;
        createdAt?: string;
        updatedAt?: string;
    }

    /**
     * - Filepath: internal/database/models/models.go
     */
    interface Models_Theme {
        enableColorSettings: boolean;
        backgroundColor: string;
        accentColor: string;
        /**
         * DEPRECATED
         */
        sidebarBackgroundColor: string;
        /**
         * DEPRECATED
         */
        animeEntryScreenLayout: string;
        expandSidebarOnHover: boolean;
        hideTopNavbar: boolean;
        enableMediaCardBlurredBackground: boolean;
        libraryScreenCustomBackgroundImage: string;
        libraryScreenCustomBackgroundOpacity: number;
        smallerEpisodeCarouselSize: boolean;
        libraryScreenBannerType: string;
        libraryScreenCustomBannerImage: string;
        libraryScreenCustomBannerPosition: string;
        libraryScreenCustomBannerOpacity: number;
        disableLibraryScreenGenreSelector: boolean;
        libraryScreenCustomBackgroundBlur: string;
        enableMediaPageBlurredBackground: boolean;
        disableSidebarTransparency: boolean;
        /**
         * DEPRECATED
         */
        useLegacyEpisodeCard: boolean;
        disableCarouselAutoScroll: boolean;
        mediaPageBannerType: string;
        mediaPageBannerSize: string;
        mediaPageBannerInfoBoxSize: string;
        showEpisodeCardAnimeInfo: boolean;
        continueWatchingDefaultSorting: string;
        animeLibraryCollectionDefaultSorting: string;
        mangaLibraryCollectionDefaultSorting: string;
        showAnimeUnwatchedCount: boolean;
        showMangaUnreadCount: boolean;
        hideEpisodeCardDescription: boolean;
        hideDownloadedEpisodeCardFilename: boolean;
        customCSS: string;
        mobileCustomCSS: string;
        id: number;
        createdAt?: string;
        updatedAt?: string;
    }

    /**
     * - Filepath: internal/database/models/models.go
     */
    interface Models_TorrentSettings {
        defaultTorrentClient: string;
        qbittorrentPath: string;
        qbittorrentHost: string;
        qbittorrentPort: number;
        qbittorrentUsername: string;
        qbittorrentPassword: string;
        qbittorrentTags: string;
        transmissionPath: string;
        transmissionHost: string;
        transmissionPort: number;
        transmissionUsername: string;
        transmissionPassword: string;
        showActiveTorrentCount: boolean;
        hideTorrentList: boolean;
    }

    /**
     * - Filepath: internal/database/models/models.go
     */
    interface Models_TorrentstreamSettings {
        enabled: boolean;
        autoSelect: boolean;
        preferredResolution: string;
        disableIPV6: boolean;
        downloadDir: string;
        addToLibrary: boolean;
        torrentClientHost: string;
        torrentClientPort: number;
        streamingServerHost: string;
        streamingServerPort: number;
        includeInLibrary: boolean;
        streamUrlAddress: string;
        slowSeeding: boolean;
        id: number;
        createdAt?: string;
        updatedAt?: string;
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
     * - Filepath: internal/onlinestream/repository.go
     */
    interface Onlinestream_EpisodeListResponse {
        episodes?: Array<Onlinestream_Episode>;
        media?: AL_BaseAnime;
    }

    /**
     * - Filepath: internal/onlinestream/repository.go
     */
    interface Onlinestream_EpisodeSource {
        number: number;
        videoSources?: Array<Onlinestream_VideoSource>;
        subtitles?: Array<Onlinestream_Subtitle>;
    }

    /**
     * - Filepath: internal/onlinestream/manual_mapping.go
     */
    interface Onlinestream_MappingResponse {
        animeId?: string;
    }

    /**
     * - Filepath: internal/onlinestream/repository.go
     */
    interface Onlinestream_Subtitle {
        url: string;
        language: string;
    }

    /**
     * - Filepath: internal/onlinestream/repository.go
     */
    interface Onlinestream_VideoSource {
        server: string;
        headers?: Record<string, string>;
        url: string;
        quality: string;
    }

    /**
     * - Filepath: internal/mediastream/videofile/video_quality.go
     */
    export type Quality = "240p" |
        "360p" |
        "480p" |
        "720p" |
        "1080p" |
        "1440p" |
        "4k" |
        "8k" |
        "original";

    /**
     * - Filepath: internal/report/report.go
     */
    interface Report_ClickLog {
        timestamp?: string;
        element: string;
        pageUrl: string;
        text?: string;
        className?: string;
    }

    /**
     * - Filepath: internal/report/report.go
     */
    interface Report_ConsoleLog {
        type: string;
        content: string;
        pageUrl: string;
        timestamp?: string;
    }

    /**
     * - Filepath: internal/report/report.go
     */
    interface Report_IssueReport {
        createdAt?: string;
        userAgent: string;
        appVersion: string;
        os: string;
        arch: string;
        clickLogs?: Array<Report_ClickLog>;
        networkLogs?: Array<Report_NetworkLog>;
        reactQueryLogs?: Array<Report_ReactQueryLog>;
        consoleLogs?: Array<Report_ConsoleLog>;
        unlockedLocalFiles?: Array<Report_UnlockedLocalFile>;
        scanLogs?: Array<string>;
        serverLogs?: string;
        status?: string;
    }

    /**
     * - Filepath: internal/report/report.go
     */
    interface Report_NetworkLog {
        type: string;
        method: string;
        url: string;
        pageUrl: string;
        status: number;
        duration: number;
        dataPreview: string;
        body: string;
        timestamp?: string;
    }

    /**
     * - Filepath: internal/report/report.go
     */
    interface Report_ReactQueryLog {
        type: string;
        pageUrl: string;
        status: string;
        hash: string;
        error: any;
        dataPreview: string;
        dataType: string;
        timestamp?: string;
    }

    /**
     * - Filepath: internal/report/report.go
     */
    interface Report_UnlockedLocalFile {
        path: string;
        mediaId: number;
    }

    /**
     * - Filepath: internal/handlers/docs.go
     */
    interface RouteHandler {
        name: string;
        trimmedName: string;
        comments?: Array<string>;
        filepath: string;
        filename: string;
        api?: RouteHandlerApi;
    }

    /**
     * - Filepath: internal/handlers/docs.go
     */
    interface RouteHandlerApi {
        summary: string;
        descriptions?: Array<string>;
        endpoint: string;
        methods?: Array<string>;
        params?: Array<RouteHandlerParam>;
        bodyFields?: Array<RouteHandlerParam>;
        returns: string;
        returnGoType: string;
        returnTypescriptType: string;
    }

    /**
     * - Filepath: internal/handlers/docs.go
     */
    interface RouteHandlerParam {
        name: string;
        jsonName: string;
        /**
         * e.g., []models.User
         */
        goType: string;
        /**
         * e.g., models.User
         */
        usedStructType: string;
        /**
         * e.g., Array<User>
         */
        typescriptType: string;
        required: boolean;
        descriptions?: Array<string>;
    }

    /**
     * - Filepath: internal/extension_playground/playground.go
     */
    interface RunPlaygroundCodeParams {
        type?: Extension_Type;
        language?: Extension_Language;
        code: string;
        inputs?: Record<string, any>;
        function: string;
    }

    /**
     * - Filepath: internal/extension_playground/playground.go
     */
    interface RunPlaygroundCodeResponse {
        logs: string;
        value: string;
    }

    /**
     * - Filepath: internal/handlers/status.go
     * @description
     *  Status is a struct containing the user data, settings, and OS.
     *  It is used by the client in various places to access necessary information.
     */
    interface Status {
        os: string;
        clientDevice: string;
        clientPlatform: string;
        clientUserAgent: string;
        dataDir: string;
        user?: Anime_User;
        settings?: Models_Settings;
        version: string;
        versionName: string;
        themeSettings?: Models_Theme;
        isOffline: boolean;
        mediastreamSettings?: Models_MediastreamSettings;
        torrentstreamSettings?: Models_TorrentstreamSettings;
        debridSettings?: Models_DebridSettings;
        anilistClientId: string;
        /**
         * If true, a new screen will be displayed
         */
        updating: boolean;
        /**
         * The server is running as a desktop sidecar
         */
        isDesktopSidecar: boolean;
        featureFlags?: INTERNAL_FeatureFlags;
        serverReady: boolean;
    }

    /**
     * - Filepath: internal/mediastream/videofile/info.go
     */
    interface Subtitle {
        index: number;
        title?: string;
        language?: string;
        codec: string;
        extension?: string;
        isDefault: boolean;
        isForced: boolean;
        isExternal: boolean;
        link?: string;
    }

    /**
     * - Filepath: internal/library/summary/scan_summary.go
     */
    interface Summary_ScanSummary {
        id: string;
        groups?: Array<Summary_ScanSummaryGroup>;
        unmatchedFiles?: Array<Summary_ScanSummaryFile>;
    }

    /**
     * - Filepath: internal/library/summary/scan_summary.go
     */
    interface Summary_ScanSummaryFile {
        id: string;
        localFile?: Anime_LocalFile;
        logs?: Array<Summary_ScanSummaryLog>;
    }

    /**
     * - Filepath: internal/library/summary/scan_summary.go
     */
    interface Summary_ScanSummaryGroup {
        id: string;
        files?: Array<Summary_ScanSummaryFile>;
        mediaId: number;
        mediaTitle: string;
        mediaImage: string;
        /**
         * Whether the media is in the user's AniList collection
         */
        mediaIsInCollection: boolean;
    }

    /**
     * - Filepath: internal/library/summary/scan_summary.go
     */
    interface Summary_ScanSummaryItem {
        createdAt?: string;
        scanSummary?: Summary_ScanSummary;
    }

    /**
     * - Filepath: internal/library/summary/scan_summary.go
     */
    interface Summary_ScanSummaryLog {
        id: string;
        filePath: string;
        level: string;
        message: string;
    }

    /**
     * - Filepath: internal/sync/sync.go
     */
    interface Sync_QueueMediaTask {
        mediaId: number;
        image: string;
        title: string;
        type: string;
    }

    /**
     * - Filepath: internal/sync/sync.go
     */
    interface Sync_QueueState {
        animeTasks?: Record<number, Sync_QueueMediaTask>;
        mangaTasks?: Record<number, Sync_QueueMediaTask>;
    }

    /**
     * - Filepath: internal/sync/manager.go
     */
    interface Sync_TrackedMediaItem {
        mediaId: number;
        type: string;
        animeEntry?: AL_AnimeListEntry;
        mangaEntry?: AL_MangaListEntry;
    }

    /**
     * - Filepath: internal/api/tvdb/types.go
     */
    interface TVDB_Episode {
        id: number;
        image: string;
        number: number;
        airedAt: string;
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

    /**
     * - Filepath: internal/torrents/torrent/search.go
     */
    interface Torrent_Preview {
        /**
         * nil if batch
         */
        episode?: Anime_Episode;
        torrent?: HibikeTorrent_AnimeTorrent;
    }

    /**
     * - Filepath: internal/torrents/torrent/search.go
     */
    interface Torrent_SearchData {
        /**
         * Torrents found
         */
        torrents?: Array<HibikeTorrent_AnimeTorrent>;
        /**
         * TorrentPreview for each torrent
         */
        previews?: Array<Torrent_Preview>;
        /**
         * Torrent metadata
         */
        torrentMetadata?: Record<string, Torrent_TorrentMetadata>;
        /**
         * Debrid instant availability
         */
        debridInstantAvailability?: Record<string, Debrid_TorrentItemInstantAvailability>;
        /**
         * AniZip media
         */
        animeMetadata?: Metadata_AnimeMetadata;
    }

    /**
     * - Filepath: internal/torrents/torrent/search.go
     */
    interface Torrent_TorrentMetadata {
        distance: number;
        metadata?: $habari.Metadata;
    }

    /**
     * - Filepath: internal/torrentstream/history.go
     */
    interface Torrentstream_BatchHistoryResponse {
        torrent?: HibikeTorrent_AnimeTorrent;
    }

    /**
     * - Filepath: internal/torrentstream/list.go
     */
    interface Torrentstream_EpisodeCollection {
        episodes?: Array<Anime_Episode>;
        hasMappingError: boolean;
    }

    /**
     * - Filepath: internal/torrentstream/previews.go
     */
    interface Torrentstream_FilePreview {
        path: string;
        displayPath: string;
        displayTitle: string;
        episodeNumber: number;
        relativeEpisodeNumber: number;
        isLikely: boolean;
        index: number;
    }

    /**
     * - Filepath: internal/torrentstream/stream.go
     */
    export type Torrentstream_PlaybackType = "default" | "externalPlayerLink";

    /**
     * - Filepath: internal/updater/check.go
     */
    interface Updater_Release {
        url: string;
        html_url: string;
        node_id: string;
        tag_name: string;
        name: string;
        body: string;
        published_at: string;
        released: boolean;
        version: string;
        assets?: Array<Updater_ReleaseAsset>;
    }

    /**
     * - Filepath: internal/updater/check.go
     */
    interface Updater_ReleaseAsset {
        url: string;
        id: number;
        node_id: string;
        name: string;
        content_type: string;
        uploaded: boolean;
        size: number;
        browser_download_url: string;
    }

    /**
     * - Filepath: internal/updater/updater.go
     */
    interface Updater_Update {
        release?: Updater_Release;
        current_version?: string;
        type: string;
    }

    /**
     * - Filepath: internal/mediastream/videofile/info.go
     */
    interface Video {
        codec: string;
        mimeCodec?: string;
        language?: string;
        quality: Quality;
        width: number;
        height: number;
        bitrate: number;
    }

}
