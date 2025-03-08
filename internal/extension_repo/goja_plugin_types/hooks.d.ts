declare namespace $app {

    /**
     * @package anilist_platform
     */

    /**
     * @event GetAnimeEvent
     * @file internal/platforms/anilist_platform/hook_events.go
     */
    function onGetAnime(cb: (event: GetAnimeEvent) => void);

    interface GetAnimeEvent {
        next();

        anime?: AL_BaseAnime;
    }

    /**
     * @event GetAnimeDetailsEvent
     * @file internal/platforms/anilist_platform/hook_events.go
     */
    function onGetAnimeDetails(cb: (event: GetAnimeDetailsEvent) => void);

    interface GetAnimeDetailsEvent {
        next();

        anime?: AL_AnimeDetailsById_Media;
    }

    /**
     * @event GetMangaEvent
     * @file internal/platforms/anilist_platform/hook_events.go
     */
    function onGetManga(cb: (event: GetMangaEvent) => void);

    interface GetMangaEvent {
        next();

        manga?: AL_BaseManga;
    }

    /**
     * @event GetMangaDetailsEvent
     * @file internal/platforms/anilist_platform/hook_events.go
     */
    function onGetMangaDetails(cb: (event: GetMangaDetailsEvent) => void);

    interface GetMangaDetailsEvent {
        next();

        manga?: AL_MangaDetailsById_Media;
    }

    /**
     * @event GetAnimeCollectionEvent
     * @file internal/platforms/anilist_platform/hook_events.go
     */
    function onGetAnimeCollection(cb: (event: GetAnimeCollectionEvent) => void);

    interface GetAnimeCollectionEvent {
        next();

        animeCollection?: AL_AnimeCollection;
    }

    /**
     * @event GetMangaCollectionEvent
     * @file internal/platforms/anilist_platform/hook_events.go
     */
    function onGetMangaCollection(cb: (event: GetMangaCollectionEvent) => void);

    interface GetMangaCollectionEvent {
        next();

        mangaCollection?: AL_MangaCollection;
    }

    /**
     * @event GetRawAnimeCollectionEvent
     * @file internal/platforms/anilist_platform/hook_events.go
     */
    function onGetRawAnimeCollection(cb: (event: GetRawAnimeCollectionEvent) => void);

    interface GetRawAnimeCollectionEvent {
        next();

        animeCollection?: AL_AnimeCollection;
    }

    /**
     * @event GetRawMangaCollectionEvent
     * @file internal/platforms/anilist_platform/hook_events.go
     */
    function onGetRawMangaCollection(cb: (event: GetRawMangaCollectionEvent) => void);

    interface GetRawMangaCollectionEvent {
        next();

        mangaCollection?: AL_MangaCollection;
    }

    /**
     * @event GetStudioDetailsEvent
     * @file internal/platforms/anilist_platform/hook_events.go
     */
    function onGetStudioDetails(cb: (event: GetStudioDetailsEvent) => void);

    interface GetStudioDetailsEvent {
        next();

        studio?: AL_StudioDetails;
    }

    /**
     * @event PreUpdateEntryEvent
     * @file internal/platforms/anilist_platform/hook_events.go
     * @description
     * PreUpdateEntryEvent is triggered when an entry is about to be updated.
     * Prevent default to skip the default update and override the update.
     */
    function onPreUpdateEntry(cb: (event: PreUpdateEntryEvent) => void);

    interface PreUpdateEntryEvent {
        next();

        preventDefault();

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
    function onPostUpdateEntry(cb: (event: PostUpdateEntryEvent) => void);

    interface PostUpdateEntryEvent {
        next();

        mediaId?: number;
    }

    /**
     * @event PreUpdateEntryProgressEvent
     * @file internal/platforms/anilist_platform/hook_events.go
     * @description
     * PreUpdateEntryProgressEvent is triggered when an entry's progress is about to be updated.
     * Prevent default to skip the default update and override the update.
     */
    function onPreUpdateEntryProgress(cb: (event: PreUpdateEntryProgressEvent) => void);

    interface PreUpdateEntryProgressEvent {
        next();

        preventDefault();

        mediaId?: number;
        progress?: number;
        totalCount?: number;
        status?: AL_MediaListStatus;
    }

    /**
     * @event PostUpdateEntryProgressEvent
     * @file internal/platforms/anilist_platform/hook_events.go
     */
    function onPostUpdateEntryProgress(cb: (event: PostUpdateEntryProgressEvent) => void);

    interface PostUpdateEntryProgressEvent {
        next();

        mediaId?: number;
    }

    /**
     * @event PreUpdateEntryRepeatEvent
     * @file internal/platforms/anilist_platform/hook_events.go
     * @description
     * PreUpdateEntryRepeatEvent is triggered when an entry's repeat is about to be updated.
     * Prevent default to skip the default update and override the update.
     */
    function onPreUpdateEntryRepeat(cb: (event: PreUpdateEntryRepeatEvent) => void);

    interface PreUpdateEntryRepeatEvent {
        next();

        preventDefault();

        mediaId?: number;
        repeat?: number;
    }

    /**
     * @event PostUpdateEntryRepeatEvent
     * @file internal/platforms/anilist_platform/hook_events.go
     */
    function onPostUpdateEntryRepeat(cb: (event: PostUpdateEntryRepeatEvent) => void);

    interface PostUpdateEntryRepeatEvent {
        next();

        mediaId?: number;
    }


    /**
     * @package anime
     */

    /**
     * @event AnimeEntryRequestedEvent
     * @file internal/library/anime/hook_events.go
     * @description
     * AnimeEntryRequestedEvent is triggered when a new media entry is being created.
     */
    function onAnimeEntryRequested(cb: (event: AnimeEntryRequestedEvent) => void);

    interface AnimeEntryRequestedEvent {
        next();

        mediaId: number;
        localFiles?: Array<Anime_LocalFile>;
        animeCollection?: AL_AnimeCollection;
    }

    /**
     * @event AnimeEntryEvent
     * @file internal/library/anime/hook_events.go
     * @description
     * AnimeEntryEvent is triggered when the media entry is being returned.
     */
    function onAnimeEntry(cb: (event: AnimeEntryEvent) => void);

    interface AnimeEntryEvent {
        next();

        entry?: Anime_Entry;
    }

    /**
     * @event AnimeEntryFillerHydrationEvent
     * @file internal/library/anime/hook_events.go
     * @description
     * AnimeEntryFillerHydrationEvent is triggered when the filler data is being added to the media entry.
     * Prevent default to avoid adding the filler data.
     */
    function onAnimeEntryFillerHydration(cb: (event: AnimeEntryFillerHydrationEvent) => void);

    interface AnimeEntryFillerHydrationEvent {
        next();

        preventDefault();

        entry?: Anime_Entry;
    }

    /**
     * @event AnimeEntryLibraryDataRequestedEvent
     * @file internal/library/anime/hook_events.go
     * @description
     * AnimeEntryLibraryDataRequestedEvent is triggered when the app requests the library data for a media entry.
     */
    function onAnimeEntryLibraryDataRequested(cb: (event: AnimeEntryLibraryDataRequestedEvent) => void);

    interface AnimeEntryLibraryDataRequestedEvent {
        next();

        entryLocalFiles?: Array<Anime_LocalFile>;
        mediaId: number;
        currentProgress: number;
    }

    /**
     * @event AnimeEntryLibraryDataEvent
     * @file internal/library/anime/hook_events.go
     * @description
     * AnimeEntryLibraryDataEvent is triggered when the library data is being added to the media entry.
     */
    function onAnimeEntryLibraryData(cb: (event: AnimeEntryLibraryDataEvent) => void);

    interface AnimeEntryLibraryDataEvent {
        next();

        entryLibraryData?: Anime_EntryLibraryData;
    }

    /**
     * @event AnimeEntryManualMatchBeforeSaveEvent
     * @file internal/library/anime/hook_events.go
     * @description
     * AnimeEntryManualMatchBeforeSaveEvent is triggered when the user manually matches local files to a media entry.
     */
    function onAnimeEntryManualMatchBeforeSave(cb: (event: AnimeEntryManualMatchBeforeSaveEvent) => void);

    interface AnimeEntryManualMatchBeforeSaveEvent {
        next();

        mediaId: number;
        paths?: Array<string>;
        matchedLocalFiles?: Array<Anime_LocalFile>;
    }

    /**
     * @event MissingEpisodesRequestedEvent
     * @file internal/library/anime/hook_events.go
     * @description
     * MissingEpisodesRequestedEvent is triggered when the user requests the missing episodes for a media entry.
     */
    function onMissingEpisodesRequested(cb: (event: MissingEpisodesRequestedEvent) => void);

    interface MissingEpisodesRequestedEvent {
        next();

        animeCollection?: AL_AnimeCollection;
        localFiles?: Array<Anime_LocalFile>;
        silencedMediaIds?: Array<number>;
    }

    /**
     * @event MissingEpisodesEvent
     * @file internal/library/anime/hook_events.go
     */
    function onMissingEpisodes(cb: (event: MissingEpisodesEvent) => void);

    interface MissingEpisodesEvent {
        next();

        missingEpisodes?: Anime_MissingEpisodes;
    }

    /**
     * @event AnimeLibraryCollectionEvent
     * @file internal/library/anime/hook_events.go
     */
    function onAnimeLibraryCollection(cb: (event: AnimeLibraryCollectionEvent) => void);

    interface AnimeLibraryCollectionEvent {
        next();

        libraryCollection?: Anime_LibraryCollection;
    }

    /**
     * @event AnimeLibraryStreamCollectionEvent
     * @file internal/library/anime/hook_events.go
     */
    function onAnimeLibraryStreamCollection(cb: (event: AnimeLibraryStreamCollectionEvent) => void);

    interface AnimeLibraryStreamCollectionEvent {
        next();

        streamCollection?: Anime_StreamCollection;
    }

    /**
     * @event AnimeLibraryCollectionRequestedEvent
     * @file internal/library/anime/hook_events.go
     */
    function onAnimeLibraryCollectionRequested(cb: (event: AnimeLibraryCollectionRequestedEvent) => void);

    interface AnimeLibraryCollectionRequestedEvent {
        next();

        animeCollection?: AL_AnimeCollection;
        localFiles?: Array<Anime_LocalFile>;
    }

    /**
     * @event AnimeLibraryStreamCollectionRequestedEvent
     * @file internal/library/anime/hook_events.go
     */
    function onAnimeLibraryStreamCollectionRequested(cb: (event: AnimeLibraryStreamCollectionRequestedEvent) => void);

    interface AnimeLibraryStreamCollectionRequestedEvent {
        next();

        animeCollection?: AL_AnimeCollection;
        libraryCollection?: Anime_LibraryCollection;
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
    function onAutoDownloaderRunStarted(cb: (event: AutoDownloaderRunStartedEvent) => void);

    interface AutoDownloaderRunStartedEvent {
        next();

        preventDefault();

        rules?: Array<Anime_AutoDownloaderRule>;
    }

    /**
     * @event AutoDownloaderTorrentsFetchedEvent
     * @file internal/library/autodownloader/hook_events.go
     * @description
     * AutoDownloaderTorrentsFetchedEvent is triggered when the autodownloader fetches torrents from the provider.
     */
    function onAutoDownloaderTorrentsFetched(cb: (event: AutoDownloaderTorrentsFetchedEvent) => void);

    interface AutoDownloaderTorrentsFetchedEvent {
        next();

        torrents?: Array<AutoDownloader_NormalizedTorrent>;
    }

    /**
     * @event AutoDownloaderMatchVerifiedEvent
     * @file internal/library/autodownloader/hook_events.go
     * @description
     * AutoDownloaderMatchVerifiedEvent is triggered when a torrent is verified to follow a rule.
     */
    function onAutoDownloaderMatchVerified(cb: (event: AutoDownloaderMatchVerifiedEvent) => void);

    interface AutoDownloaderMatchVerifiedEvent {
        next();

        torrent?: AutoDownloader_NormalizedTorrent;
        rule?: Anime_AutoDownloaderRule;
        listEntry?: AL_AnimeListEntry;
        localEntry?: Anime_LocalFileWrapperEntry;
        episode: number;
        ok: boolean;
    }

    /**
     * @event AutoDownloaderSettingsUpdatedEvent
     * @file internal/library/autodownloader/hook_events.go
     * @description
     * AutoDownloaderSettingsUpdatedEvent is triggered when the autodownloader settings are updated
     */
    function onAutoDownloaderSettingsUpdated(cb: (event: AutoDownloaderSettingsUpdatedEvent) => void);

    interface AutoDownloaderSettingsUpdatedEvent {
        next();

        settings?: Models_AutoDownloaderSettings;
    }


    /**
     * @package debrid_client
     */

    /**
     * @event DebridSendStreamToMediaPlayerEvent
     * @file internal/debrid/client/hook_events.go
     * @description
     * DebridSendStreamToMediaPlayerEvent is triggered when the debrid client is about to send a stream to the media player.
     * Prevent default to skip the default playback and override the playback.
     */
    function onDebridSendStreamToMediaPlayer(cb: (event: DebridSendStreamToMediaPlayerEvent) => void);

    interface DebridSendStreamToMediaPlayerEvent {
        next();

        preventDefault();

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
    function onDebridLocalDownloadRequested(cb: (event: DebridLocalDownloadRequestedEvent) => void);

    interface DebridLocalDownloadRequestedEvent {
        next();

        preventDefault();

        torrentName: string;
        destination: string;
        downloadUrl: string;
    }


    /**
     * @package manga
     */

    /**
     * @event MangaEntryRequestedEvent
     * @file internal/manga/hook_events.go
     * @description
     * MangaEntryRequestedEvent is triggered when a new media entry is being created.
     */
    function onMangaEntryRequested(cb: (event: MangaEntryRequestedEvent) => void);

    interface MangaEntryRequestedEvent {
        next();

        mediaId: number;
        mangaCollection?: AL_MangaCollection;
    }

    /**
     * @event MangaEntryEvent
     * @file internal/manga/hook_events.go
     * @description
     * MangaEntryEvent is triggered when the media entry is being returned.
     */
    function onMangaEntry(cb: (event: MangaEntryEvent) => void);

    interface MangaEntryEvent {
        next();

        entry?: Manga_Entry;
    }

    /**
     * @event MangaLibraryCollectionEvent
     * @file internal/manga/hook_events.go
     */
    function onMangaLibraryCollection(cb: (event: MangaLibraryCollectionEvent) => void);

    interface MangaLibraryCollectionEvent {
        next();

        libraryCollection?: Manga_Collection;
    }

    /**
     * @event MangaLibraryCollectionRequestedEvent
     * @file internal/manga/hook_events.go
     */
    function onMangaLibraryCollectionRequested(cb: (event: MangaLibraryCollectionRequestedEvent) => void);

    interface MangaLibraryCollectionRequestedEvent {
        next();

        mangaCollection?: AL_MangaCollection;
    }


    /**
     * @package metadata
     */

    /**
     * @event AnimeMetadataRequestedEvent
     * @file internal/api/metadata/hook_events.go
     * @description
     * AnimeMetadataRequestedEvent is triggered when anime metadata is requested.
     * Prevent default to skip the default behavior and return the overridden metadata.
     */
    function onAnimeMetadataRequested(cb: (event: AnimeMetadataRequestedEvent) => void);

    interface AnimeMetadataRequestedEvent {
        next();

        preventDefault();

        mediaId: number;
        overrideAnimeMetadata?: Metadata_AnimeMetadata;
    }

    /**
     * @event AnimeMetadataEvent
     * @file internal/api/metadata/hook_events.go
     * @description
     * AnimeMetadataEvent is triggered when anime metadata is available.
     */
    function onAnimeMetadata(cb: (event: AnimeMetadataEvent) => void);

    interface AnimeMetadataEvent {
        next();

        mediaId: number;
        animeMetadata?: Metadata_AnimeMetadata;
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
    function onLocalFilePlaybackRequested(cb: (event: LocalFilePlaybackRequestedEvent) => void);

    interface LocalFilePlaybackRequestedEvent {
        next();

        preventDefault();

        path: string;
    }

    /**
     * @event StreamPlaybackRequestedEvent
     * @file internal/library/playbackmanager/hook_events.go
     * @description
     * StreamPlaybackRequestedEvent is triggered when a stream is requested to be played.
     * Prevent default to skip the default playback and override the playback.
     */
    function onStreamPlaybackRequested(cb: (event: StreamPlaybackRequestedEvent) => void);

    interface StreamPlaybackRequestedEvent {
        next();

        preventDefault();

        windowTitle: string;
        payload: string;
        media?: AL_BaseAnime;
        aniDbEpisode: string;
    }

    /**
     * @event PrePlaybackTrackingEvent
     * @file internal/library/playbackmanager/hook_events.go
     * @description
     * PrePlaybackTrackingEvent is triggered just before the playback tracking starts.
     * Prevent default to skip the default playback tracking.
     */
    function onPrePlaybackTracking(cb: (event: PrePlaybackTrackingEvent) => void);

    interface PrePlaybackTrackingEvent {
        next();

        preventDefault();

        isStream: boolean;
    }


    /**
     * @package scanner
     */

    /**
     * @event ScanStartedEvent
     * @file internal/library/scanner/hook_events.go
     * @description
     * ScanStartedEvent is triggered when a scan operation begins
     */
    function onScanStarted(cb: (event: ScanStartedEvent) => void);

    interface ScanStartedEvent {
        next();

        dirPath: string;
        otherDirPaths?: Array<string>;
        enhanced: boolean;
        skipLocked: boolean;
        skipIgnored: boolean;
    }

    /**
     * @event ScanCompletedEvent
     * @file internal/library/scanner/hook_events.go
     * @description
     * ScanCompletedEvent is triggered when a scan operation finishes
     */
    function onScanCompleted(cb: (event: ScanCompletedEvent) => void);

    interface ScanCompletedEvent {
        next();

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
     * ScanMediaFetcherStartedEvent is triggered when the media fetcher begins
     */
    function onScanMediaFetcherStarted(cb: (event: ScanMediaFetcherStartedEvent) => void);

    interface ScanMediaFetcherStartedEvent {
        next();

        enhanced: boolean;
    }

    /**
     * @event ScanMediaFetcherCompletedEvent
     * @file internal/library/scanner/hook_events.go
     * @description
     * ScanMediaFetcherCompletedEvent is triggered when the media fetcher completes
     */
    function onScanMediaFetcherCompleted(cb: (event: ScanMediaFetcherCompletedEvent) => void);

    interface ScanMediaFetcherCompletedEvent {
        next();

        allMedia?: Array<AL_CompleteAnime>;
        unknownMediaIds?: Array<number>;
    }

    /**
     * @event ScanMatchingStartedEvent
     * @file internal/library/scanner/hook_events.go
     * @description
     * ScanMatchingStartedEvent is triggered when the matching process begins.
     * Prevent default to skip the default matching and override the matching.
     */
    function onScanMatchingStarted(cb: (event: ScanMatchingStartedEvent) => void);

    interface ScanMatchingStartedEvent {
        next();

        preventDefault();

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
    function onScanLocalFileMatched(cb: (event: ScanLocalFileMatchedEvent) => void);

    interface ScanLocalFileMatchedEvent {
        next();

        preventDefault();

        match?: Anime_NormalizedMedia;
        found: boolean;
        localFile?: Anime_LocalFile;
        score: number;
    }

    /**
     * @event ScanMatchingCompletedEvent
     * @file internal/library/scanner/hook_events.go
     * @description
     * ScanMatchingCompletedEvent is triggered when the matching process completes
     */
    function onScanMatchingCompleted(cb: (event: ScanMatchingCompletedEvent) => void);

    interface ScanMatchingCompletedEvent {
        next();

        localFiles?: Array<Anime_LocalFile>;
    }

    /**
     * @event ScanHydrationStartedEvent
     * @file internal/library/scanner/hook_events.go
     * @description
     * ScanHydrationStartedEvent is triggered when the file hydration process begins
     */
    function onScanHydrationStarted(cb: (event: ScanHydrationStartedEvent) => void);

    interface ScanHydrationStartedEvent {
        next();

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
    function onScanLocalFileHydrationStarted(cb: (event: ScanLocalFileHydrationStartedEvent) => void);

    interface ScanLocalFileHydrationStartedEvent {
        next();

        preventDefault();

        localFile?: Anime_LocalFile;
        media?: Anime_NormalizedMedia;
    }

    /**
     * @event ScanLocalFileHydratedEvent
     * @file internal/library/scanner/hook_events.go
     * @description
     * ScanLocalFileHydratedEvent is triggered when a local file's metadata is hydrated
     */
    function onScanLocalFileHydrated(cb: (event: ScanLocalFileHydratedEvent) => void);

    interface ScanLocalFileHydratedEvent {
        next();

        localFile?: Anime_LocalFile;
        mediaId: number;
        episode: number;
    }

    /**
     * @event ScanHydrationCompletedEvent
     * @file internal/library/scanner/hook_events.go
     * @description
     * ScanHydrationCompletedEvent is triggered when the file hydration process completes
     */
    function onScanHydrationCompleted(cb: (event: ScanHydrationCompletedEvent) => void);

    interface ScanHydrationCompletedEvent {
        next();

        localFiles?: Array<Anime_LocalFile>;
    }


    /**
     * @package torrentstream
     */

    /**
     * @event TorrentStreamSendStreamToMediaPlayerEvent
     * @file internal/torrentstream/hook_events.go
     * @description
     * TorrentStreamSendStreamToMediaPlayerEvent is triggered when the torrent stream is about to send a stream to the media player.
     * Prevent default to skip the default playback and override the playback.
     */
    function onTorrentStreamSendStreamToMediaPlayer(cb: (event: TorrentStreamSendStreamToMediaPlayerEvent) => void);

    interface TorrentStreamSendStreamToMediaPlayerEvent {
        next();

        preventDefault();

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
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_AnimeCollection {
        mediaListCollection?: AL_AnimeCollection_MediaListCollection;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_AnimeCollection_MediaListCollection {
        lists?: Array<AL_AnimeCollection_MediaListCollection_Lists>;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_AnimeCollection_MediaListCollection_Lists {
        status?: AL_MediaListStatus;
        name?: string;
        isCustomList?: boolean;
        entries?: Array<AL_AnimeCollection_MediaListCollection_Lists_Entries>;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
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
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_AnimeCollection_MediaListCollection_Lists_Entries_CompletedAt {
        year?: number;
        month?: number;
        day?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_AnimeCollection_MediaListCollection_Lists_Entries_StartedAt {
        year?: number;
        month?: number;
        day?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
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
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_AnimeDetailsById_Media_Characters {
        edges?: Array<AL_AnimeDetailsById_Media_Characters_Edges>;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_AnimeDetailsById_Media_Characters_Edges {
        id?: number;
        role?: AL_CharacterRole;
        name?: string;
        node?: AL_BaseCharacter;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_AnimeDetailsById_Media_EndDate {
        year?: number;
        month?: number;
        day?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
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
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_AnimeDetailsById_Media_Recommendations {
        edges?: Array<AL_AnimeDetailsById_Media_Recommendations_Edges>;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_AnimeDetailsById_Media_Recommendations_Edges {
        node?: AL_AnimeDetailsById_Media_Recommendations_Edges_Node;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_AnimeDetailsById_Media_Recommendations_Edges_Node {
        mediaRecommendation?: AL_AnimeDetailsById_Media_Recommendations_Edges_Node_MediaRecommendation;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
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
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_AnimeDetailsById_Media_Recommendations_Edges_Node_MediaRecommendation_CoverImage {
        extraLarge?: string;
        large?: string;
        medium?: string;
        color?: string;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_AnimeDetailsById_Media_Recommendations_Edges_Node_MediaRecommendation_StartDate {
        year?: number;
        month?: number;
        day?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_AnimeDetailsById_Media_Recommendations_Edges_Node_MediaRecommendation_Title {
        romaji?: string;
        english?: string;
        native?: string;
        userPreferred?: string;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_AnimeDetailsById_Media_Recommendations_Edges_Node_MediaRecommendation_Trailer {
        id?: string;
        site?: string;
        thumbnail?: string;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_AnimeDetailsById_Media_Relations {
        edges?: Array<AL_AnimeDetailsById_Media_Relations_Edges>;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_AnimeDetailsById_Media_Relations_Edges {
        relationType?: AL_MediaRelation;
        node?: AL_BaseAnime;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_AnimeDetailsById_Media_Staff {
        edges?: Array<AL_AnimeDetailsById_Media_Staff_Edges>;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_AnimeDetailsById_Media_Staff_Edges {
        role?: string;
        node?: AL_AnimeDetailsById_Media_Staff_Edges_Node;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_AnimeDetailsById_Media_Staff_Edges_Node {
        name?: AL_AnimeDetailsById_Media_Staff_Edges_Node_Name;
        id: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_AnimeDetailsById_Media_Staff_Edges_Node_Name {
        full?: string;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_AnimeDetailsById_Media_StartDate {
        year?: number;
        month?: number;
        day?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_AnimeDetailsById_Media_Studios {
        nodes?: Array<AL_AnimeDetailsById_Media_Studios_Nodes>;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_AnimeDetailsById_Media_Studios_Nodes {
        name: string;
        id: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_AnimeDetailsById_Media_Trailer {
        id?: string;
        site?: string;
        thumbnail?: string;
    }

    /**
     * - Filepath: internal/api/anilist/collection_helper.go
     * - Filename: collection_helper.go
     * - Package: anilist
     */
    export type AL_AnimeListEntry = AL_AnimeCollection_MediaListCollection_Lists_Entries;

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
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
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_BaseAnime_CoverImage {
        extraLarge?: string;
        large?: string;
        medium?: string;
        color?: string;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_BaseAnime_EndDate {
        year?: number;
        month?: number;
        day?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_BaseAnime_NextAiringEpisode {
        airingAt: number;
        timeUntilAiring: number;
        episode: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_BaseAnime_StartDate {
        year?: number;
        month?: number;
        day?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_BaseAnime_Title {
        userPreferred?: string;
        romaji?: string;
        english?: string;
        native?: string;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_BaseAnime_Trailer {
        id?: string;
        site?: string;
        thumbnail?: string;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
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
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_BaseCharacter_DateOfBirth {
        year?: number;
        month?: number;
        day?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_BaseCharacter_Image {
        large?: string;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_BaseCharacter_Name {
        full?: string;
        native?: string;
        alternative?: Array<string>;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
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
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_BaseManga_CoverImage {
        extraLarge?: string;
        large?: string;
        medium?: string;
        color?: string;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_BaseManga_EndDate {
        year?: number;
        month?: number;
        day?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_BaseManga_StartDate {
        year?: number;
        month?: number;
        day?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_BaseManga_Title {
        userPreferred?: string;
        romaji?: string;
        english?: string;
        native?: string;
    }

    /**
     * - Filepath: internal/api/anilist/models_gen.go
     * - Filename: models_gen.go
     * - Package: anilist
     * @description
     *  The role the character plays in the media
     */
    export type AL_CharacterRole = "MAIN" | "SUPPORTING" | "BACKGROUND";

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
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
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_CompleteAnime_CoverImage {
        extraLarge?: string;
        large?: string;
        medium?: string;
        color?: string;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_CompleteAnime_EndDate {
        year?: number;
        month?: number;
        day?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_CompleteAnime_NextAiringEpisode {
        airingAt: number;
        timeUntilAiring: number;
        episode: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_CompleteAnime_Relations {
        edges?: Array<AL_CompleteAnime_Relations_Edges>;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_CompleteAnime_Relations_Edges {
        relationType?: AL_MediaRelation;
        node?: AL_BaseAnime;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_CompleteAnime_StartDate {
        year?: number;
        month?: number;
        day?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_CompleteAnime_Title {
        userPreferred?: string;
        romaji?: string;
        english?: string;
        native?: string;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_CompleteAnime_Trailer {
        id?: string;
        site?: string;
        thumbnail?: string;
    }

    /**
     * - Filepath: internal/api/anilist/models_gen.go
     * - Filename: models_gen.go
     * - Package: anilist
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
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_MangaCollection {
        mediaListCollection?: AL_MangaCollection_MediaListCollection;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_MangaCollection_MediaListCollection {
        lists?: Array<AL_MangaCollection_MediaListCollection_Lists>;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_MangaCollection_MediaListCollection_Lists {
        status?: AL_MediaListStatus;
        name?: string;
        isCustomList?: boolean;
        entries?: Array<AL_MangaCollection_MediaListCollection_Lists_Entries>;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
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
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_MangaCollection_MediaListCollection_Lists_Entries_CompletedAt {
        year?: number;
        month?: number;
        day?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_MangaCollection_MediaListCollection_Lists_Entries_StartedAt {
        year?: number;
        month?: number;
        day?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
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
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_MangaDetailsById_Media_Characters {
        edges?: Array<AL_MangaDetailsById_Media_Characters_Edges>;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_MangaDetailsById_Media_Characters_Edges {
        id?: number;
        role?: AL_CharacterRole;
        name?: string;
        node?: AL_BaseCharacter;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
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
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_MangaDetailsById_Media_Recommendations {
        edges?: Array<AL_MangaDetailsById_Media_Recommendations_Edges>;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_MangaDetailsById_Media_Recommendations_Edges {
        node?: AL_MangaDetailsById_Media_Recommendations_Edges_Node;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_MangaDetailsById_Media_Recommendations_Edges_Node {
        mediaRecommendation?: AL_MangaDetailsById_Media_Recommendations_Edges_Node_MediaRecommendation;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
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
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_MangaDetailsById_Media_Recommendations_Edges_Node_MediaRecommendation_CoverImage {
        extraLarge?: string;
        large?: string;
        medium?: string;
        color?: string;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_MangaDetailsById_Media_Recommendations_Edges_Node_MediaRecommendation_EndDate {
        year?: number;
        month?: number;
        day?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_MangaDetailsById_Media_Recommendations_Edges_Node_MediaRecommendation_StartDate {
        year?: number;
        month?: number;
        day?: number;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_MangaDetailsById_Media_Recommendations_Edges_Node_MediaRecommendation_Title {
        userPreferred?: string;
        romaji?: string;
        english?: string;
        native?: string;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_MangaDetailsById_Media_Relations {
        edges?: Array<AL_MangaDetailsById_Media_Relations_Edges>;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_MangaDetailsById_Media_Relations_Edges {
        relationType?: AL_MediaRelation;
        node?: AL_BaseManga;
    }

    /**
     * - Filepath: internal/api/anilist/models_gen.go
     * - Filename: models_gen.go
     * - Package: anilist
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
     * - Filename: models_gen.go
     * - Package: anilist
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
     * - Filename: models_gen.go
     * - Package: anilist
     * @description
     *  The type of ranking
     */
    export type AL_MediaRankType = "RATED" | "POPULAR";

    /**
     * - Filepath: internal/api/anilist/models_gen.go
     * - Filename: models_gen.go
     * - Package: anilist
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
     * - Filename: models_gen.go
     * - Package: anilist
     */
    export type AL_MediaSeason = "WINTER" | "SPRING" | "SUMMER" | "FALL";

    /**
     * - Filepath: internal/api/anilist/models_gen.go
     * - Filename: models_gen.go
     * - Package: anilist
     * @description
     *  The current releasing status of the media
     */
    export type AL_MediaStatus = "FINISHED" | "RELEASING" | "NOT_YET_RELEASED" | "CANCELLED" | "HIATUS";

    /**
     * - Filepath: internal/api/anilist/models_gen.go
     * - Filename: models_gen.go
     * - Package: anilist
     * @description
     *  Media type enum, anime or manga.
     */
    export type AL_MediaType = "ANIME" | "MANGA";

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_StudioDetails {
        studio?: AL_StudioDetails_Studio;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_StudioDetails_Studio {
        id: number;
        isAnimationStudio: boolean;
        name: string;
        media?: AL_StudioDetails_Studio_Media;
    }

    /**
     * - Filepath: internal/api/anilist/client_gen.go
     * - Filename: client_gen.go
     * - Package: anilist
     */
    interface AL_StudioDetails_Studio_Media {
        nodes?: Array<AL_BaseAnime>;
    }

    /**
     * - Filepath: internal/library/anime/autodownloader_rule.go
     * - Filename: autodownloader_rule.go
     * - Package: anime
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
     * - Filename: autodownloader_rule.go
     * - Package: anime
     */
    export type Anime_AutoDownloaderRuleEpisodeType = "recent" | "selected";

    /**
     * - Filepath: internal/library/anime/autodownloader_rule.go
     * - Filename: autodownloader_rule.go
     * - Package: anime
     */
    export type Anime_AutoDownloaderRuleTitleComparisonType = "contains" | "likely";

    /**
     * - Filepath: internal/library/anime/entry.go
     * - Filename: entry.go
     * - Package: anime
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
     * - Filename: entry_download_info.go
     * - Package: anime
     */
    interface Anime_EntryDownloadEpisode {
        episodeNumber: number;
        aniDBEpisode: string;
        episode?: Anime_Episode;
    }

    /**
     * - Filepath: internal/library/anime/entry_download_info.go
     * - Filename: entry_download_info.go
     * - Package: anime
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
     * - Filename: entry_library_data.go
     * - Package: anime
     */
    interface Anime_EntryLibraryData {
        allFilesLocked: boolean;
        sharedPath: string;
        unwatchedCount: number;
        mainFileCount: number;
    }

    /**
     * - Filepath: internal/library/anime/entry.go
     * - Filename: entry.go
     * - Package: anime
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
     * - Filename: episode.go
     * - Package: anime
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
     * - Filename: episode.go
     * - Package: anime
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
     * - Filename: collection.go
     * - Package: anime
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
     * - Filename: collection.go
     * - Package: anime
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
     * - Filename: collection.go
     * - Package: anime
     */
    interface Anime_LibraryCollectionList {
        type?: AL_MediaListStatus;
        status?: AL_MediaListStatus;
        entries?: Array<Anime_LibraryCollectionEntry>;
    }

    /**
     * - Filepath: internal/library/anime/collection.go
     * - Filename: collection.go
     * - Package: anime
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
     * - Filename: localfile.go
     * - Package: anime
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
     * - Filename: localfile.go
     * - Package: anime
     */
    interface Anime_LocalFileMetadata {
        episode: number;
        aniDBEpisode: string;
        type: Anime_LocalFileType;
    }

    /**
     * - Filepath: internal/library/anime/localfile.go
     * - Filename: localfile.go
     * - Package: anime
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
     * - Filename: localfile.go
     * - Package: anime
     */
    export type Anime_LocalFileType = "main" | "special" | "nc";

    /**
     * - Filepath: internal/library/anime/localfile_wrapper.go
     * - Filename: localfile_wrapper.go
     * - Package: anime
     */
    interface Anime_LocalFileWrapperEntry {
        mediaId: number;
        localFiles?: Array<Anime_LocalFile>;
    }

    /**
     * - Filepath: internal/library/anime/missing_episodes.go
     * - Filename: missing_episodes.go
     * - Package: anime
     */
    interface Anime_MissingEpisodes {
        episodes?: Array<Anime_Episode>;
        silencedEpisodes?: Array<Anime_Episode>;
    }

    /**
     * - Filepath: internal/library/anime/normalized_media.go
     * - Filename: normalized_media.go
     * - Package: anime
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
     * - Filename: collection.go
     * - Package: anime
     */
    interface Anime_StreamCollection {
        continueWatchingList?: Array<Anime_Episode>;
        anime?: Array<AL_BaseAnime>;
        listData?: Record<number, Anime_EntryListData>;
    }

    /**
     * - Filepath: internal/library/anime/collection.go
     * - Filename: collection.go
     * - Package: anime
     */
    interface Anime_UnknownGroup {
        mediaId: number;
        localFiles?: Array<Anime_LocalFile>;
    }

    /**
     * - Filepath: internal/library/anime/collection.go
     * - Filename: collection.go
     * - Package: anime
     */
    interface Anime_UnmatchedGroup {
        dir: string;
        localFiles?: Array<Anime_LocalFile>;
        suggestions?: Array<AL_BaseAnime>;
    }

    /**
     * - Filepath: internal/library/autodownloader/autodownloader_torrent.go
     * - Filename: autodownloader_torrent.go
     * - Package: autodownloader
     */
    interface AutoDownloader_NormalizedTorrent {
        parsedData?: Metadata;
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
     * - Filepath: internal/manga/collection.go
     * - Filename: collection.go
     * - Package: manga
     */
    interface Manga_Collection {
        lists?: Array<Manga_CollectionList>;
    }

    /**
     * - Filepath: internal/manga/collection.go
     * - Filename: collection.go
     * - Package: manga
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
     * - Filename: collection.go
     * - Package: manga
     */
    interface Manga_CollectionList {
        type?: AL_MediaListStatus;
        status?: AL_MediaListStatus;
        entries?: Array<Manga_CollectionEntry>;
    }

    /**
     * - Filepath: internal/manga/manga_entry.go
     * - Filename: manga_entry.go
     * - Package: manga
     */
    interface Manga_Entry {
        mediaId: number;
        media?: AL_BaseManga;
        listData?: Manga_EntryListData;
    }

    /**
     * - Filepath: internal/manga/manga_entry.go
     * - Filename: manga_entry.go
     * - Package: manga
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
     * - Filepath: internal/api/metadata/types.go
     * - Filename: types.go
     * - Package: metadata
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
     * - Filename: types.go
     * - Package: metadata
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
     * - Filename: types.go
     * - Package: metadata
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
     * - Filename: models.go
     * - Package: models
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

}
