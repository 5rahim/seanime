import { MangaChapterDetails } from "@/app/(main)/manga/_lib/manga.types"
import { Button } from "@/components/ui/button"
import React from "react"
import { FaDownload } from "react-icons/fa"

type ChapterListBulkActionsProps = {
    rowSelectedChapters: MangaChapterDetails[] | undefined
    onDownloadSelected: (chapters: MangaChapterDetails[]) => void
}

export function ChapterListBulkActions(props: ChapterListBulkActionsProps) {

    const {
        rowSelectedChapters,
        onDownloadSelected,
        ...rest
    } = props

    const handleDownloadSelected = React.useCallback(() => {
        onDownloadSelected(rowSelectedChapters || [])
    }, [onDownloadSelected, rowSelectedChapters])

    if (rowSelectedChapters?.length === 0) return null

    return (
        <>
            <div
                className=""
            >
                <Button
                    onClick={handleDownloadSelected}
                    intent="white"
                    size="sm"
                    leftIcon={<FaDownload />}
                    className="animate-pulse"
                >
                    Download selected chapters ({rowSelectedChapters?.length})
                </Button>
            </div>
        </>
    )
}
