import { MangaChapterDetails, MangaDownloadData } from "@/app/(main)/manga/_lib/manga.types"
import React from "react"

export function getChapterNumberFromChapter(chapter: string): number {
    const chapterNumber = chapter.match(/(\d+(\.\d+)?)/)?.[0]
    return chapterNumber ? Math.floor(parseFloat(chapterNumber)) : 0
}

export function useMangaDownloadDataUtils(data: MangaDownloadData | undefined, loading: boolean) {

    const isChapterDownloaded = React.useCallback((chapter: MangaChapterDetails | undefined) => {
        if (!data || !chapter) return false
        return !!data?.downloaded[chapter.provider]?.includes(chapter.id)
    }, [data])

    const isChapterQueued = React.useCallback((chapter: MangaChapterDetails | undefined) => {
        if (!data || !chapter) return false
        return !!data?.queued[chapter.provider]?.includes(chapter.id)
    }, [data])

    const getProviderNumberOfDownloadedChapters = React.useCallback((provider: string) => {
        if (!data) return 0
        return Object.keys(data.downloaded[provider] || {}).length
    }, [data])

    return {
        isChapterDownloaded,
        isChapterQueued,
        getProviderNumberOfDownloadedChapters,
        showActionButtons: !loading,
    }
}
