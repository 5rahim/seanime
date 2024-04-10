import { MangaChapterDetails, MangaDownloadData } from "@/app/(main)/manga/_lib/manga.types"
import { DataGridRowSelectedEvent } from "@/components/ui/datagrid/use-datagrid-row-selection"
import { RowSelectionState } from "@tanstack/react-table"
import React from "react"

export function getChapterNumberFromChapter(chapter: string): number {
    const chapterNumber = chapter.match(/(\d+(\.\d+)?)/)?.[0]
    return chapterNumber ? Math.floor(parseFloat(chapterNumber)) : 0
}

export function useMangaDownloadDataUtils(data: MangaDownloadData | undefined, loading: boolean) {

    const isChapterDownloaded = React.useCallback((chapter: MangaChapterDetails | undefined) => {
        if (!data || !chapter) return false
        return (data?.downloaded[chapter.provider]?.findIndex(n => n.chapterId === chapter.id) ?? -1) !== -1
    }, [data])

    const isChapterQueued = React.useCallback((chapter: MangaChapterDetails | undefined) => {
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
    }
}

export function useMangaChapterListRowSelection() {

    const [rowSelection, setRowSelection] = React.useState<RowSelectionState>({})

    const [selectedChapters, setSelectedChapters] = React.useState<MangaChapterDetails[]>([])

    const onSelectChange = React.useCallback((event: DataGridRowSelectedEvent<MangaChapterDetails>) => {
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
