import {
    MangaChapterContainer,
    MangaChapterDetails,
    MangaCollection,
    MangaEntry,
    MangaEntryBackups,
    MangaPageContainer,
} from "@/app/(main)/manga/_lib/types"
import { getChapterNumberFromChapter } from "@/app/(main)/manga/_lib/utils"
import { MangaDetailsByIdQuery } from "@/lib/anilist/gql/graphql"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation, useSeaQuery } from "@/lib/server/query"
import { useAtomValue } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import React from "react"

const enum MangaProvider {
    COMICK = "comick",
    MANGASEE = "mangasee",
}

export const __manga_selectedProviderAtom = atomWithStorage<string>("sea-manga-provider", MangaProvider.COMICK)

export function useMangaCollection() {
    const { data, isLoading } = useSeaQuery<MangaCollection>({
        endpoint: SeaEndpoints.MANGA_COLLECTION,
        queryKey: ["get-manga-collection"],
    })

    return {
        mangaCollection: data,
        mangaCollectionLoading: isLoading,
    }
}

export function useMangaEntry(mediaId: string | undefined | null) {
    const { data, isLoading } = useSeaQuery<MangaEntry>({
        endpoint: SeaEndpoints.MANGA_ENTRY.replace("{id}", mediaId ?? ""),
        queryKey: ["get-manga-entry", Number(mediaId)],
        enabled: !!mediaId,
    })

    return {
        mangaEntry: data,
        mangaEntryLoading: isLoading,
    }
}

export function useMangaEntryDetails(mediaId: string | undefined | null) {
    const { data, isLoading } = useSeaQuery<MangaDetailsByIdQuery["Media"]>({
        endpoint: SeaEndpoints.MANGA_ENTRY_DETAILS.replace("{id}", mediaId ?? ""),
        queryKey: ["get-manga-entry-details", Number(mediaId)],
        enabled: !!mediaId,
    })

    return {
        mangaDetails: data,
        mangaDetailsLoading: isLoading,
    }
}

export function useUpdateMangaProgress() {
    const { mutate, isPending } = useSeaMutation<boolean>({
        endpoint: SeaEndpoints.UPDATE_MANGA_PROGRESS,
        mutationKey: ["update-manga-progress"],
    })

    return {
        updateProgress: mutate,
        isUpdating: isPending,
    }
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Backups
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export function useMangaEntryBackups(mediaId: string | undefined | null) {
    // FIXME SHELVED
    // const provider = useAtomValue(__manga_selectedProviderAtom)
    //
    // const { data, isLoading, isFetching } = useSeaQuery<MangaEntryBackups>({
    //     endpoint: SeaEndpoints.MANGA_ENTRY_BACKUPS,
    //     method: "post",
    //     data: {
    //         mediaId: Number(mediaId),
    //         provider,
    //     },
    //     queryKey: ["get-manga-entry-backups", Number(mediaId), provider],
    //     enabled: !!mediaId,
    //     gcTime: 0,
    // })
    //
    // return {
    //     chapterBackups: data,
    //     chapterBackupsLoading: isLoading || isFetching,
    // }

    return {
        chapterBackups: {
            mediaId: Number(mediaId),
            provider: "comick",
            chapterIds: {},
        } as MangaEntryBackups,
        chapterBackupsLoading: false,
    }
}

export function useDownloadMangaChapter(mediaId: string | undefined | null) {
    const provider = useAtomValue(__manga_selectedProviderAtom)

    const { mutate, isPending } = useSeaMutation<void, { mediaId: number, provider: string, chapterId: string }>({
        endpoint: SeaEndpoints.DOWNLOAD_MANGA_CHAPTER,
        method: "post",
        mutationKey: ["download-manga-chapter", Number(mediaId), provider],
    })

    return {
        downloadChapter: (chapter: MangaChapterDetails) => {
            mutate({
                mediaId: Number(mediaId),
                provider,
                chapterId: chapter.id,
            })
        },
        isSendingDownloadRequest: isPending,
    }
}


//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Chapters and Pages
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export function useMangaChapterContainer(mediaId: string | undefined | null) {
    const provider = useAtomValue(__manga_selectedProviderAtom)

    const { data, isLoading, isFetching } = useSeaQuery<MangaChapterContainer>({
        endpoint: SeaEndpoints.MANGA_CHAPTERS,
        method: "post",
        data: {
            mediaId: Number(mediaId),
            provider,
        },
        queryKey: ["get-manga-chapters", Number(mediaId), provider],
        enabled: !!mediaId,
        gcTime: 0,
    })

    // Keep track of chapter numbers as integers
    // This is used to filter the chapters
    // [id]: number
    const chapterNumbersMap = React.useMemo(() => {
        const map = new Map<string, number>()

        for (const chapter of data?.chapters ?? []) {
            map.set(chapter.id, getChapterNumberFromChapter(chapter.chapter))
        }

        return map
    }, [data?.chapters])

    return {
        chapterContainer: data,
        chapterIdToNumbersMap: chapterNumbersMap,
        chapterContainerLoading: isLoading || isFetching,
    }
}

export function useMangaPageContainer(mediaId: string | undefined | null, chapterId: string | undefined | null) {
    const provider = useAtomValue(__manga_selectedProviderAtom)

    const { data, isLoading, isError } = useSeaQuery<MangaPageContainer>({
        endpoint: SeaEndpoints.MANGA_PAGES,
        method: "post",
        data: {
            mediaId: Number(mediaId),
            chapterId,
            provider,
        },
        queryKey: ["get-manga-pages", Number(mediaId), provider, chapterId],
        enabled: !!mediaId && !!chapterId,
    })

    return {
        pageContainer: data,
        pageContainerLoading: isLoading,
        pageContainerError: isError,
    }
}
