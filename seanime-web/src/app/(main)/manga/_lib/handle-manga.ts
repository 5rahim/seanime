import { Manga_ChapterDetails, Manga_Collection, Manga_Provider } from "@/api/generated/types"
import { useGetMangaCollection } from "@/api/hooks/manga.hooks"
import {
    useClearAllChapterDownloadQueue,
    useDownloadMangaChapters,
    useGetMangaDownloadQueue,
    useResetErroredChapterDownloadQueue,
    useStartMangaDownloadQueue,
    useStopMangaDownloadQueue,
} from "@/api/hooks/manga_download.hooks"
import { useAtomValue } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import { useRouter } from "next/navigation"
import React, { useMemo } from "react"
import { toast } from "sonner"

export const __manga_selectedProviderAtom = atomWithStorage<Manga_Provider>("sea-manga-provider", "mangapill")

/**
 * Get the manga collection
 */
export function useMangaCollection() {
    const router = useRouter()
    const { data, isLoading, isError } = useGetMangaCollection()

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
        } as Manga_Collection
    }, [data])

    return {
        mangaCollection: sortedCollection,
        mangaCollectionLoading: isLoading,
    }
}


//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Downloads
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export function useHandleDownloadMangaChapter(mediaId: string | undefined | null) {
    const provider = useAtomValue(__manga_selectedProviderAtom)

    const { mutate, isPending } = useDownloadMangaChapters(mediaId, provider)

    return {
        downloadChapters: (chapters: Manga_ChapterDetails[]) => {
            mutate({
                mediaId: Number(mediaId),
                provider,
                chapterIds: chapters.map(ch => ch.id),
                startNow: false,
            }, {
                onSuccess: () => {
                    toast.success("Chapters added to download queue")
                },
            })
        },
        isSendingDownloadRequest: isPending,
    }
}


export function useHandleMangaChapterDownloadQueue() {

    const { data, isLoading, isError } = useGetMangaDownloadQueue()

    const { mutate: start, isPending: isStarting } = useStartMangaDownloadQueue()

    const { mutate: stop, isPending: isStopping } = useStopMangaDownloadQueue()

    const { mutate: resetErrored, isPending: isResettingErrored } = useResetErroredChapterDownloadQueue()

    const { mutate: clearQueue, isPending: isClearingQueue } = useClearAllChapterDownloadQueue()

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
