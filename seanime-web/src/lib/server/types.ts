/**
 * Requests
 */
import {
    BaseMediaFragment,
    BasicMediaFragment,
    GetViewerQuery,
    MediaFormat,
    MediaListStatus,
    MediaRelation,
    MediaSeason,
    MediaStatus,
    MediaType,
} from "@/lib/anilist/gql/graphql"

export type SeaErrorResponse = { error: string }
export type SeaDataResponse<T> = { data: T | undefined }
export type SeaResponse<T> = SeaDataResponse<T> | SeaErrorResponse

export type SeaWebsocketEvent<T> = { type: string, payload: T }


/**
 * Auth
 */

export type ServerStatus = {
    os: string,
    user: {
        viewer: GetViewerQuery["Viewer"],
        token: string
    } | null,
    settings: Settings | null
    mal: MalInfo | null
    version: string
    themeSettings?: ThemeSettings | null
}

/**
 * Settings
 */

export type Settings = {
    library?: LibrarySettings
    mediaPlayer?: MediaPlayerSettings
    torrent?: TorrentSettings
    anilist?: AnilistSettings
    listSync?: ListSyncSettings
    autoDownloader?: AutoDownloaderSettings
}

export type AnilistSettings = {
    hideAudienceScore: boolean
}

export type MediaPlayerSettings = {
    defaultPlayer: string
    host: string
    vlcUsername: string
    vlcPassword: string
    vlcPort: number
    vlcPath: string
    mpcPort: number
    mpcPath: string
    mpvSocket: string
    mpvPath: string
}

export type LibrarySettings = {
    libraryPath: string
    autoUpdateProgress: boolean
    disableUpdateCheck: boolean
    torrentProvider: string
    autoScan: boolean
}

export const DEFAULT_TORRENT_PROVIDER = "animetosho"
export const DEFAULT_TORRENT_CLIENT = "qbittorrent"

export type TorrentSettings = {
    defaultTorrentClient: string
    qbittorrentPath: string
    qbittorrentHost: string
    qbittorrentPort: number
    qbittorrentUsername: string
    qbittorrentPassword: string
    transmissionPath: string
    transmissionHost: string
    transmissionPort: number
    transmissionUsername: string
    transmissionPassword: string
}

/**
 * List Sync
 */

export type ListSyncSettings = {
    origin: string
    automatic: boolean
}

export const enum ListSyncOrigin {
    ANILIST = "anilist",
    MAL = "mal"
}

export const enum ListSyncAnimeDiffKind {
    MISSING_IN_ORIGIN = "missing_in_origin",
    MISSING_IN_TARGET = "missing_in_target",
    METADATA = "metadata",
}

export const enum ListSyncAnimeMetadataDiffKind {
    SCORE = "score",
    PROGRESS = "progress",
    STATUS = "status",
}


export type ListSyncAnimeEntry = {
    source: ListSyncOrigin
    sourceID: number
    malID: number
    displayTitle: string
    url: string
    progress: number
    totalEpisodes: number
    status: string
    image: string
    score: string
}

export type ListSyncAnimeDiff = {
    id: string
    targetSource: string
    originEntry?: ListSyncAnimeEntry
    targetEntry?: ListSyncAnimeEntry
    kind: ListSyncAnimeDiffKind
    metadataDiffKinds: ListSyncAnimeMetadataDiffKind[]
}

/**
 * MAL
 */

export type MalInfo = {
    username: string
    accessToken: string
    refreshToken: string
}

export type MalAuthResponse = {
    access_token: string
    refresh_token: string
    expires_in: number
    token_type: string
}

/**
 * Collection
 */

export type LibraryCollection = {
    continueWatchingList: MediaEntryEpisode[]
    lists: LibraryCollectionList[]
    unmatchedLocalFiles: LocalFile[]
    ignoredLocalFiles: LocalFile[]
    unmatchedGroups: UnmatchedGroup[]
    unknownGroups: UnknownGroup[]
}

export type LibraryCollectionListType = "current" | "planned" | "completed" | "paused" | "dropped"

export type LibraryCollectionList = {
    type: LibraryCollectionListType
    status: MediaListStatus
    entries: LibraryCollectionEntry[]
}

export type LibraryCollectionEntry = {
    media?: BaseMediaFragment
    mediaId: number
    listData?: MediaEntryListData
    libraryData?: MediaEntryLibraryData
}

export type UnmatchedGroup = {
    dir: string
    localFiles: LocalFile[]
    suggestions: BasicMediaFragment[]
}

export type UnknownGroup = {
    mediaId: number
    localFiles: LocalFile[]
}

/**
 * Media
 */

export type MediaEntry = {
    mediaId: number
    media?: BaseMediaFragment | null | undefined
    currentEpisodeCount: number
    episodes?: MediaEntryEpisode[]
    nextEpisode?: MediaEntryEpisode
    localFiles: LocalFile[]
    listData?: MediaEntryListData
    libraryData?: MediaEntryLibraryData
    downloadInfo?: MediaEntryDownloadInfo
    aniDBId?: number
}

export type MediaEntryListData = {
    progress?: number
    score?: number
    status?: MediaListStatus
    startedAt?: string
    completedAt?: string
}

export type MediaEntryLibraryData = {
    allFilesLocked: boolean
    sharedPath: string
}

export type MediaEntryDownloadInfo = {
    episodesToDownload: MediaEntryDownloadEpisode[]
    canBatch: boolean
    batchAll: boolean
    hasInaccurateSchedule: boolean
    rewatch: boolean
    absoluteOffset: number
}

export type MediaEntryDownloadEpisode = {
    episodeNumber: number
    aniDBEpisode: string
    episode?: MediaEntryEpisode
}

export type MediaEntryEpisode = {
    type: LocalFileType
    displayTitle: string
    episodeTitle: string
    episodeNumber: number
    absoluteEpisodeNumber: number
    progressNumber: number
    localFile?: LocalFile
    isDownloaded: boolean
    episodeMetadata?: MediaEntryEpisodeMetadata
    fileMetadata?: LocalFileMetadata
    isInvalid: boolean
    metadataIssue?: string
    basicMedia?: BasicMediaFragment
}

export type MediaEntryEpisodeMetadata = {
    image?: string
    airDate?: string
    length?: number
    summary?: string
    overview?: string
    aniDBId?: string
}

/**
 * Local File
 */

export type LocalFile = {
    path: string
    name: string
    parsedInfo?: LocalFileParsedData
    parsedFolderInfo: LocalFileParsedData[]
    metadata: LocalFileMetadata
    locked: boolean
    ignored: boolean
    mediaId: number
}

export type LocalFileType = "main" | "special" | "nc"

export type LocalFileMetadata = {
    episode: number
    aniDBEpisode: string
    type: LocalFileType
}

export type LocalFileParsedData = {
    original: string
    title?: string
    releaseGroup?: string
    season?: string
    seasonRange?: string[]
    part?: string
    partRange?: string[]
    episode?: string
    episodeRange?: string[]
    episodeTitle?: string
    year?: string
}

/**
 * Media Player
 */

export type MediaPlayerPlaybackStatus = {
    completionPercentage: number
    filename: string
    playing: boolean
    duration: number
}

/**
 * Playback Manager
 */

export type PlaybackManagerPlaybackState = {
    state: "tracking" | "completed"
    filename: string
    mediaTitle: string
    mediaTotalEpisodes: number
    episodeNumber: number
    completionPercentage: number
    canPlayNext: boolean
    progressUpdated: boolean
    mediaId: number
}

export type PlaybackManagerPlaylistState = {
    current: PlaybackManagerPlaylistStateItem | null
    next: PlaybackManagerPlaylistStateItem | null
    remaining: number
}

export type PlaybackManagerPlaylistStateItem = {
    name: string
    mediaImage: string
}

/**
 * Torrent
 */

export type TorrentSearchData = {
    previews: TorrentPreview[]
    torrents: AnimeTorrent[]
}

export type AnimeTorrent = {
    name: string
    date: string
    size: number
    formattedSize: string
    seeders: number
    leechers: number
    downloadCount: number
    link: string
    downloadUrl: string
    infoHash: string
    resolution?: string
    isBatch: boolean
    episodeNumber?: number
    releaseGroup?: string
    provider: string
}

export type TorrentPreview = {
    torrent: AnimeTorrent
    episode: MediaEntryEpisode | null
}

//---

export type SeaTorrent = {
    name: string
    hash: string
    seeds: number
    upSpeed: string
    downSpeed: string
    progress: number
    size: string
    eta: string
    status: "downloading" | "paused" | "seeding"
    contentPath: string
}

export type SeaTorrentActionProps = { hash: string, action: "pause" | "resume" | "remove" | "open", dir: string }

/**
 * Scan Summary
 */

export type ScanSummary = {
    createdAt: string
    id: string
    groups: ScanSummaryGroup[] | undefined
    unmatchedFiles: ScanSummaryFile[] | undefined
}

export type ScanSummaryFile = {
    id: string
    localFile: LocalFile
    logs: ScanSummaryLog[]
}

export type ScanSummaryGroup = {
    id: string
    files: ScanSummaryFile[]
    mediaId: number
    mediaTitle: string
    mediaImage: string
    mediaIsInCollection: boolean
}

export type ScanSummaryLog = {
    id: string
    filePath: string
    message: string
    level: "info" | "warning" | "error"
}

/**
 * Auto Downloader
 */

export type AutoDownloaderRule = {
    dbId: number
    enabled: boolean
    mediaId: number
    releaseGroups: string[]
    resolutions: string[]
    comparisonTitle: string
    titleComparisonType: string
    episodeType: string
    episodeNumbers?: number[]
    destination: string
}

export type AutoDownloaderSettings = {
    provider: string
    interval: number
    enabled: boolean
    downloadAutomatically: boolean
}

export type AutoDownloaderItem = {
    id: number
    createdAt: string
    updatedAt: string
    ruleId: number
    mediaId: number
    episode: number
    link: string
    hash: string
    magnet: string
    torrentName: string
    downloaded: boolean
}

/**
 * Updates / Releases
 */

export type Update = {
    release?: Release
    type: string
}

export type LatestReleaseResponse = {
    release: Release
}

export type Release = {
    url: string
    html_url: string
    node_id: string
    tag_name: string
    name: string
    body: string
    published_at: string
    released: boolean
    version: string
    assets: ReleaseAsset[]
}

export type ReleaseAsset = {
    url: string
    id: number
    node_id: string
    name: string
    content_type: string
    uploaded: boolean
    size: number
    browser_download_url: string
}

export type ThemeSettings = {
    animeEntryScreenLayout: string
    smallerEpisodeCarouselSize: boolean
    expandSidebarOnHover: boolean
    backgroundColor: string
    sidebarBackgroundColor: string
    libraryScreenBannerType: string
    libraryScreenCustomBannerImage: string
    libraryScreenCustomBannerPosition: string
    libraryScreenCustomBannerOpacity: number
    libraryScreenCustomBackgroundImage: string
    libraryScreenCustomBackgroundOpacity: number
}

/**
 *
 */

export type Playlist = {
    dbId: number
    name: string
    localFiles: LocalFile[]
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export type AnilistCollectionEntry = {
    id: number,
    score?: number | null,
    progress?: number | null,
    status?: MediaListStatus | null,
    notes?: string | null,
    repeat?: number | null,
    private?: boolean | null,
    startedAt?: { year?: number | null, month?: number | null, day?: number | null } | null,
    completedAt?: { year?: number | null, month?: number | null, day?: number | null } | null,
    media?: {
        id: number,
        idMal?: number | null,
        siteUrl?: string | null,
        status?: MediaStatus | null,
        season?: MediaSeason | null,
        type?: MediaType | null,
        format?: MediaFormat | null,
        bannerImage?: string | null,
        episodes?: number | null,
        synonyms?: Array<string | null> | null,
        isAdult?: boolean | null,
        countryOfOrigin?: any | null,
        title?: {
            userPreferred?: string | null,
            romaji?: string | null,
            english?: string | null,
            native?: string | null
        } | null,
        coverImage?: {
            extraLarge?: string | null,
            large?: string | null,
            medium?: string | null,
            color?: string | null
        } | null,
        startDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
        endDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
        nextAiringEpisode?: { airingAt: number, timeUntilAiring: number, episode: number } | null,
        relations?: {
            edges?: Array<{
                relationType?: MediaRelation | null,
                node?: {
                    id: number,
                    idMal?: number | null,
                    siteUrl?: string | null,
                    status?: MediaStatus | null,
                    season?: MediaSeason | null,
                    type?: MediaType | null,
                    format?: MediaFormat | null,
                    bannerImage?: string | null,
                    episodes?: number | null,
                    synonyms?: Array<string | null> | null,
                    isAdult?: boolean | null,
                    countryOfOrigin?: any | null,
                    title?: {
                        userPreferred?: string | null,
                        romaji?: string | null,
                        english?: string | null,
                        native?: string | null
                    } | null,
                    coverImage?: {
                        extraLarge?: string | null,
                        large?: string | null,
                        medium?: string | null,
                        color?: string | null
                    } | null,
                    startDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
                    endDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
                    nextAiringEpisode?: {
                        airingAt: number,
                        timeUntilAiring: number,
                        episode: number
                    } | null
                } | null
            } | null> | null
        } | null
    } | null
}

export type AnilistCollectionList = {
    status?: MediaListStatus | null, entries?: Array<AnilistCollectionEntry | null> | null
}
