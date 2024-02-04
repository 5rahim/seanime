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
}

export type LibrarySettings = {
    libraryPath: string
}

export type TorrentSettings = {
    qbittorrentPath: string
    qbittorrentHost: string
    qbittorrentPort: number
    qbittorrentUsername: string
    qbittorrentPassword: string
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

export type MediaPlayerPlaybackStatus = {
    completionPercentage: number
    filename: string
    playing: boolean
    duration: number
}

//---

export type TorrentSearchData = {
    previews: TorrentPreview[]
    torrents: SearchTorrent[]
}


export type SearchTorrent = {
    category: string
    name: string
    description: string
    date: string
    size: string
    seeders: string
    leechers: string
    downloads: string
    isTrusted: string
    isRemake: string
    comments: string
    link: string
    guid: string
    categoryId: string
    infoHash: string
    resolution: string
    // /!\ will be true if the torrent is a movie
    isBatch: boolean
}

export type SearchTorrentComment = {
    user: string
    date: string
    text: string
}

export type TorrentPreview = {
    torrent: SearchTorrent
    episode: MediaEntryEpisode | null
    isBatch: boolean
    resolution: string
    releaseGroup: string
    episodeNumber?: number
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

export type SeaTorrentActionProps = { hash: string, action: "pause" | "resume" | "open", dir: string }


//---

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
