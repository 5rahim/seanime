import { BaseMangaFragment, MediaListStatus } from "@/lib/anilist/gql/graphql"

export const MANGA_PROVIDER_OPTIONS = [
    { value: "mangasee", label: "MangaSee" },
    { value: "comick", label: "ComicK" },
]

// +----------------------------------------------------------------+
// Query
// +----------------------------------------------------------------+

export type MangaChapterContainer_QueryVariables = {
    mediaId: number
    provider: string
}

export type MangaPageContainer_QueryVariables = {
    mediaId: number
    provider: string
    chapterId: string
    doublePage: boolean
}

export type ClearMangaCache_QueryVariables = {
    mediaId: number
}


// +----------------------------------------------------------------+
// Return data
// +----------------------------------------------------------------+

export type MangaCollection = {
    lists: MangaCollectionList[]
}

export type MangaCollectionList = {
    type: string
    status: string
    entries: MangaCollectionEntry[]
}

export type MangaCollectionEntry = {
    media: BaseMangaFragment
    mediaId: number
    listData?: MangaEntryListData
}

export type MangaEntry = {
    mediaId: number
    media: BaseMangaFragment
    listData?: MangaEntryListData
}

export type MangaEntryListData = {
    progress?: number
    score?: number
    status?: MediaListStatus
    startedAt?: string
    completedAt?: string
}

export type MangaEntryBackups = {
    mediaId: number
    provider: string
    chapterIds: Record<string, boolean>
}

export type MangaChapterContainer = {
    mediaId: number
    provider: string
    chapters?: MangaChapterDetails[]
}

export type MangaPageContainer = {
    mediaId: number
    provider: string
    chapterId: string
    pages?: MangaChapterPage[]
    pageDimensions?: Record<number, { width: number, height: number }>
    isDownloaded?: boolean
}

export type MangaChapterDetails = {
    provider: string
    id: string
    url: string
    title: string
    chapter: string
    index: number
    rating?: number
    updatedAt?: string
}

export type MangaChapterPage = {
    provider: string
    url: string
    index: number
}

