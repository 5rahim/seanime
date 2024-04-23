/**
 * Collection
 */
import { BaseMediaFragment, BasicMediaFragment, MediaListStatus } from "@/lib/anilist/gql/graphql"


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
