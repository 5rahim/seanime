import { HibikeManga_ChapterDetails, Manga_MediaDownloadData, Nullish } from "@/api/generated/types"
import {
    useClearAllChapterDownloadQueue,
    useDownloadMangaChapters,
    useGetMangaDownloadData,
    useGetMangaDownloadQueue,
    useResetErroredChapterDownloadQueue,
    useStartMangaDownloadQueue,
    useStopMangaDownloadQueue,
} from "@/api/hooks/manga_download.hooks"
import { useSelectedMangaProvider } from "@/app/(main)/manga/_lib/handle-manga-selected-provider"
import { atom } from "jotai"
import { useAtomValue, useSetAtom } from "jotai/react"
import React from "react"
import { toast } from "sonner"

/**
 * Stores fetched manga download data
 */
const __manga_entryDownloadDataAtom = atom<Manga_MediaDownloadData | undefined>(undefined)

export type MangaDownloadChapterItem = { provider: string, chapterId: string, chapterNumber: string, queued: boolean, downloaded: boolean }

/**
 * @description
 * - This atom transforms the download data into a list of chapters
 */
const __manga_entryDownloadedChaptersAtom = atom<MangaDownloadChapterItem[]>(get => {
    let d: MangaDownloadChapterItem[] = []
    const data = get(__manga_entryDownloadDataAtom)
    if (data) {
        for (const provider in data.downloaded) {
            d = d.concat(data.downloaded[provider].map(ch => ({
                provider,
                chapterId: ch.chapterId,
                chapterNumber: ch.chapterNumber,
                queued: false,
                downloaded: true,
            })))
        }
        for (const provider in data.queued) {
            d = d.concat(data.queued[provider].map(ch => ({
                provider,
                chapterId: ch.chapterId,
                chapterNumber: ch.chapterNumber,
                queued: true,
                downloaded: false,
            })))
        }
    }
    return d
})

export function useMangaEntryDownloadedChapters() {
    return useAtomValue(__manga_entryDownloadedChaptersAtom)
}

/**
 * @description
 * - Fetch manga download data and store it in a state
 * - We store the download data in a state, so we can handle chapter pagination.
 *      For example, clicking "next chapter" will look for a downloaded chapter, and make a request with the appropriate provider
 */
export function useHandleMangaDownloadData(mediaId: Nullish<string | number>) {
    const { data, isLoading, isError } = useGetMangaDownloadData({
        mediaId: mediaId ? Number(mediaId) : undefined,
    })

    const setDownloadData = useSetAtom(__manga_entryDownloadDataAtom)
    React.useEffect(() => {
        setDownloadData(data)
    }, [data])

    return {
        downloadData: data,
        downloadDataLoading: isLoading,
        downloadDataError: isError,
    }
}

export function useMangaEntryDownloadData() {
    return useAtomValue(__manga_entryDownloadDataAtom)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

/**
 * Handle downloading manga chapters
 */
export function useHandleDownloadMangaChapter(mediaId: string | undefined | null) {
    const { selectedProvider } = useSelectedMangaProvider(mediaId)

    const { mutate, isPending } = useDownloadMangaChapters(mediaId, selectedProvider)

    return {
        downloadChapters: (chapters: HibikeManga_ChapterDetails[]) => {
            if (selectedProvider) {
                mutate({
                    mediaId: Number(mediaId),
                    provider: selectedProvider,
                    chapterIds: chapters.map(ch => ch.id),
                    startNow: false,
                }, {
                    onSuccess: () => {
                        toast.success("Chapters added to download queue")
                    },
                })
            }
        },
        isSendingDownloadRequest: isPending,
    }
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

/**
 * Handle the manga chapter download queue
 */
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

