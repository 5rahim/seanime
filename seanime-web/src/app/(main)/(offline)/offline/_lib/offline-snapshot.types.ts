import { MediaEntryEpisode } from "@/app/(main)/(library)/_lib/anime-library.types"
import { MangaChapterContainer } from "@/app/(main)/manga/_lib/manga.types"
import {
    AnimeCollectionQuery,
    BaseMangaFragment,
    BaseMediaFragment,
    GetViewerQuery,
    MangaCollectionQuery,
    MediaListStatus,
} from "@/lib/anilist/gql/graphql"

export type OfflineSnapshot = {
    dbId: number
    user?: {
        token: string,
        viewer: GetViewerQuery["Viewer"]
    }
    entries?: OfflineEntries
    libraryCollections?: Collections
    assetMap?: OfflineAssetMap
}

export type OfflineAssetMap = Record<string, string>

export type Collections = {
    animeCollection?: AnimeCollectionQuery
    mangaCollection?: MangaCollectionQuery
}

export type OfflineEntries = {
    animeEntries: (OfflineAnimeEntry | undefined)[]
    mangaEntries: (OfflineMangaEntry | undefined)[]
}

export type OfflineAnimeEntry = {
    mediaId: number
    listData?: OfflineListData
    media?: BaseMediaFragment
    episodes: (MediaEntryEpisode | undefined)[]
    downloadedAssets: boolean
}

export type OfflineMangaEntry = {
    mediaId: number
    listData: OfflineListData | undefined
    media: BaseMangaFragment | undefined
    chapterContainers: MangaChapterContainer[] | undefined
    downloadedAssets: boolean
}

export type OfflineListData = {
    score: number
    status: MediaListStatus
    progress: number
    startedAt: string
    completedAt: string
}
