"use client"
import { getServerBaseUrl } from "@/api/client/server-url"
import { HibikeManga_ChapterDetails, Manga_MediaDownloadData } from "@/api/generated/types"
import { DataGridRowSelectedEvent } from "@/components/ui/datagrid/use-datagrid-row-selection"
import { RowSelectionState } from "@tanstack/react-table"
import React from "react"

export function getChapterNumberFromChapter(chapter: string): number {
    const chapterNumber = chapter.match(/(\d+(\.\d+)?)/)?.[0]
    return chapterNumber ? Math.floor(parseFloat(chapterNumber)) : 0
}

export function getDecimalFromChapter(chapter: string): number {
    const chapterNumber = chapter.match(/(\d+(\.\d+)?)/)?.[0]
    return chapterNumber ? parseFloat(chapterNumber) : 0
}

export function isChapterBefore(a: string, b: string): boolean {
    // compare the decimal part of the chapter number
    return getDecimalFromChapter(a) < getDecimalFromChapter(b)
}

export function isChapterAfter(a: string, b: string): boolean {
    // compare the decimal part of the chapter number
    return getDecimalFromChapter(a) > getDecimalFromChapter(b)
}

export function useMangaReaderUtils() {

    const getChapterPageUrl = React.useCallback((url: string, isDownloaded: boolean | undefined, headers?: Record<string, string>) => {
        if (url.startsWith("{{manga-local-assets}}")) {
            return `${getServerBaseUrl()}/api/v1/manga/local-page/${encodeURIComponent(url)}`
        }

        if (!isDownloaded) {
            if (headers && Object.keys(headers).length > 0) {
                return `${getServerBaseUrl()}/api/v1/image-proxy?url=${encodeURIComponent(url)}&headers=${encodeURIComponent(
                    JSON.stringify(headers))}`
            }
            return url
        }

        return `${getServerBaseUrl()}/manga-downloads/${url}`
    }, [])
    return {
        getChapterPageUrl,
    }

}

export function useMangaDownloadDataUtils(data: Manga_MediaDownloadData | undefined, loading: boolean) {

    const isChapterLocal = React.useCallback((chapter: HibikeManga_ChapterDetails | undefined) => {
        if (!chapter) return false
        return chapter.provider === "local-manga"
    }, [])

    const isChapterDownloaded = React.useCallback((chapter: HibikeManga_ChapterDetails | undefined) => {
        if (!data || !chapter) return false
        return (data?.downloaded[chapter.provider]?.findIndex(n => n.chapterId === chapter.id) ?? -1) !== -1
    }, [data])

    const isChapterQueued = React.useCallback((chapter: HibikeManga_ChapterDetails | undefined) => {
        if (!data || !chapter) return false
        return (data?.queued[chapter.provider]?.findIndex(n => n.chapterId === chapter.id) ?? -1) !== -1
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
        isChapterLocal,
    }

}

export function useMangaChapterListRowSelection() {

    const [rowSelection, setRowSelection] = React.useState<RowSelectionState>({})

    const [selectedChapters, setSelectedChapters] = React.useState<HibikeManga_ChapterDetails[]>([])

    const onSelectChange = React.useCallback((event: DataGridRowSelectedEvent<HibikeManga_ChapterDetails>) => {
        setSelectedChapters(event.data)
    }, [])
    return {
        rowSelection, setRowSelection,
        rowSelectedChapters: selectedChapters,
        onRowSelectionChange: onSelectChange,
        resetRowSelection: () => {
            setRowSelection({})
            setSelectedChapters([])
        },
    }
}
