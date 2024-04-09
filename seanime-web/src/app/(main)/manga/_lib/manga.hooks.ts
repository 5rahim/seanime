import {
    ClearMangaCache_QueryVariables,
    MangaChapterContainer,
    MangaChapterContainer_QueryVariables,
    MangaChapterDetails,
    MangaChapterDownloadQueueItem,
    MangaCollection,
    MangaDownloadChapters_QueryVariables,
    MangaDownloadData,
    MangaDownloadData_QueryVariables,
    MangaDownloadListItem,
    MangaEntry,
    MangaPageContainer,
    MangaPageContainer_QueryVariables,
} from "@/app/(main)/manga/_lib/manga.types"
import { getChapterNumberFromChapter } from "@/app/(main)/manga/_lib/manga.utils"
import { __manga_readingModeAtom, MangaReadingMode } from "@/app/(main)/manga/entry/_containers/chapter-reader/_lib/manga-chapter-reader.atoms"
import { MangaDetailsByIdQuery } from "@/lib/anilist/gql/graphql"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation, useSeaQuery } from "@/lib/server/query"
import { useQueryClient } from "@tanstack/react-query"
import { useAtomValue } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import { useRouter } from "next/navigation"
import React, { useMemo } from "react"
import { toast } from "sonner"

const enum MangaProvider {
    COMICK = "comick",
    MANGASEE = "mangasee",
}

export const __manga_selectedProviderAtom = atomWithStorage<string>("sea-manga-provider", MangaProvider.COMICK)

export function useMangaCollection() {
    const router = useRouter()
    const { data, isLoading, isError } = useSeaQuery<MangaCollection>({
        endpoint: SeaEndpoints.MANGA_COLLECTION,
        queryKey: ["get-manga-collection"],
    })

    React.useEffect(() => {
        if (isError) {
            router.push("/")
        }
    }, [isError])

    const sortedCollection = useMemo(() => {
        if (!data || !data.lists) return data
        return {
            ...data,
            lists: [
                data.lists.find(n => n.type === "current"),
                data.lists.find(n => n.type === "paused"),
                data.lists.find(n => n.type === "planned"),
                data.lists.find(n => n.type === "completed"),
                data.lists.find(n => n.type === "dropped"),
            ].filter(Boolean),
        } as MangaCollection
    }, [data])

    return {
        mangaCollection: sortedCollection,
        mangaCollectionLoading: isLoading,
    }
}

export function useMangaEntry(mediaId: string | undefined | null) {
    const router = useRouter()
    const { data, isLoading, isError } = useSeaQuery<MangaEntry>({
        endpoint: SeaEndpoints.MANGA_ENTRY.replace("{id}", mediaId ?? ""),
        queryKey: ["get-manga-entry", Number(mediaId)],
        enabled: !!mediaId,
    })

    React.useEffect(() => {
        if (isError) {
            router.push("/")
        }
    }, [isError])

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

export function useUpdateMangaProgress(mediaId: number) {
    const qc = useQueryClient()
    const { mutate: updateProgress, isPending: isUpdatingProgress } = useSeaMutation<boolean, {
        chapterNumber: number,
        mediaId: number,
        malId?: number,
        totalChapters: number,
    }>({
        endpoint: SeaEndpoints.UPDATE_MANGA_PROGRESS,
        mutationKey: ["update-manga-progress", mediaId],
        method: "post",
        onSuccess: async () => {
            await qc.refetchQueries({ queryKey: ["get-manga-entry", Number(mediaId)] })
            await qc.refetchQueries({ queryKey: ["get-manga-collection"] })
            toast.success("Progress updated")
        },
    })

    return {
        updateProgress,
        isUpdatingProgress,
    }
}


//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Chapters and Pages
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export function useClearMangaCache() {
    const qc = useQueryClient()
    const { mutate, isPending } = useSeaMutation<boolean, ClearMangaCache_QueryVariables>({
        endpoint: SeaEndpoints.MANGA_ENTRY_CACHE,
        method: "delete",
        mutationKey: ["clear-manga-cache"],
        onSuccess: async () => {
            await qc.refetchQueries({ queryKey: ["get-manga-chapters"] })
            toast.success("Sources reloaded successfully")
        },
    })

    return {
        clearMangaCache: mutate,
        isClearingMangaCache: isPending,
    }
}

export function useMangaChapterContainer(mediaId: string | undefined | null) {
    const provider = useAtomValue(__manga_selectedProviderAtom)

    const { data, isLoading, isError, isFetching } = useSeaQuery<MangaChapterContainer, MangaChapterContainer_QueryVariables>({
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
        chapterContainerError: isError,
    }
}

export function useMangaPageContainer(mediaId: string | undefined | null, chapterId: string | undefined | null) {
    const provider = useAtomValue(__manga_selectedProviderAtom)
    const readingMode = useAtomValue(__manga_readingModeAtom)

    const isDoublePage = React.useMemo(() => readingMode === MangaReadingMode.DOUBLE_PAGE, [readingMode])

    const { data, isLoading, isError, isFetching } = useSeaQuery<MangaPageContainer, MangaPageContainer_QueryVariables>({
        endpoint: SeaEndpoints.MANGA_PAGES,
        method: "post",
        data: {
            mediaId: Number(mediaId),
            chapterId: chapterId!,
            provider,
            doublePage: isDoublePage,
        },
        queryKey: ["get-manga-pages", Number(mediaId), provider, chapterId, isDoublePage],
        enabled: !!mediaId && !!chapterId,
    })

    return {
        pageContainer: data,
        pageContainerLoading: isLoading || isFetching,
        pageContainerError: isError,
    }
}


//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Downloads
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export function useMangaDownloadData(mediaId: string | undefined | null, entry: MangaEntry | undefined | null) {
    const { data, isLoading, isFetching } = useSeaQuery<MangaDownloadData, MangaDownloadData_QueryVariables>({
        endpoint: SeaEndpoints.MANGA_DOWNLOAD_DATA,
        method: "post",
        data: {
            mediaId: Number(mediaId),
            cached: true,
        },
        queryKey: ["get-manga-download-data", Number(mediaId)],
        enabled: !!mediaId && !!entry,
    })

    return {
        mangaDownloadData: data,
        mangaDownloadDataLoading: isLoading || isFetching,
    }
}

export function useDownloadMangaChapter(mediaId: string | undefined | null) {
    const qc = useQueryClient()
    const provider = useAtomValue(__manga_selectedProviderAtom)

    const { mutate, isPending } = useSeaMutation<void, MangaDownloadChapters_QueryVariables>({
        endpoint: SeaEndpoints.MANGA_DOWNLOAD_CHAPTERS,
        method: "post",
        mutationKey: ["download-manga-chapters", Number(mediaId), provider],
        onSuccess: async () => {
            await qc.refetchQueries({ queryKey: ["get-manga-download-data", Number(mediaId)] })
        },
    })

    return {
        downloadChapter: (chapter: MangaChapterDetails) => {
            mutate({
                mediaId: Number(mediaId),
                provider,
                chapterIds: [chapter.id],
                startNow: false,
            }, {
                onSuccess: () => {
                    toast.success("Chapter added to download queue")
                },
            })
        },
        isSendingDownloadRequest: isPending,
    }
}

export function useMangaChapterDownloadQueue() {
    const qc = useQueryClient()

    const { data, isLoading, isError } = useSeaQuery<MangaChapterDownloadQueueItem[]>({
        endpoint: SeaEndpoints.MANGA_DOWNLOAD_QUEUE,
        method: "get",
        queryKey: ["get-manga-chapter-download-queue"],
    })

    const { mutate: start, isPending: isStarting } = useSeaMutation<void>({
        endpoint: SeaEndpoints.MANGA_DOWNLOAD_QUEUE_START,
        method: "post",
        mutationKey: ["start-manga-chapter-download-queue"],
        onSuccess: async () => {
            await qc.refetchQueries({ queryKey: ["get-manga-chapter-download-queue"] })
            toast.info("Downloading chapters")
        },
    })

    const { mutate: stop, isPending: isStopping } = useSeaMutation<void>({
        endpoint: SeaEndpoints.MANGA_DOWNLOAD_QUEUE_STOP,
        method: "post",
        mutationKey: ["stop-manga-chapter-download-queue"],
        onSuccess: async () => {
            await qc.refetchQueries({ queryKey: ["get-manga-chapter-download-queue"] })
            toast.success("Download queue stopped")
        },
    })

    const { mutate: resetErrored, isPending: isResettingErrored } = useSeaMutation<void>({
        endpoint: SeaEndpoints.MANGA_DOWNLOAD_QUEUE_RESET_ERRORED,
        method: "post",
        mutationKey: ["manga-chapter-download-queue-reset-errored"],
        onSuccess: async () => {
            await qc.refetchQueries({ queryKey: ["get-manga-chapter-download-queue"] })
            toast.success("Reset errored chapters")
        },
    })

    const { mutate: clearQueue, isPending: isClearingQueue } = useSeaMutation<void>({
        endpoint: SeaEndpoints.MANGA_DOWNLOAD_QUEUE,
        method: "delete",
        mutationKey: ["clear-manga-chapter-download-queue"],
        onSuccess: async () => {
            await qc.refetchQueries({ queryKey: ["get-manga-chapter-download-queue"] })
            toast.success("Download queue cleared")
        },
    })

    return {
        downloadQueue: data,
        downloadQueueLoading: isLoading,
        downloadQueueError: isError,
        startDownloadQueue: start,
        isStartingDownloadQueue: isStarting,
        stopDownloadQueue: stop,
        isStoppingDownloadQueue: isStopping,
        resetErroredChapters: resetErrored,
        isResettingErroredChapters: isResettingErrored,
        clearDownloadQueue: clearQueue,
        isClearingDownloadQueue: isClearingQueue,
    }
}


export function useMangaChapterDownloads() {
    const qc = useQueryClient()

    const { data, isLoading, isError } = useSeaQuery<MangaDownloadListItem[]>({
        endpoint: SeaEndpoints.MANGA_DOWNLOADS,
        method: "get",
        queryKey: ["get-manga-downloads"],
    })


    return {
        data,
        isLoading,
        isError,
    }
}
