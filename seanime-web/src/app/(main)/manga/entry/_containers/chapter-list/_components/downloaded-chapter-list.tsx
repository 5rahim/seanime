// /* -------------------------------------------------------------------------------------------------
//  * Download List
//  * -----------------------------------------------------------------------------------------------*/


import { useDeleteDownloadedMangaChapter } from "@/app/(main)/manga/_lib/manga.hooks"
import { MangaDownloadData, MangaEntry } from "@/app/(main)/manga/_lib/manga.types"
import { primaryPillCheckboxClass } from "@/components/shared/styling/classnames"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { DataGrid, defineDataGridColumns } from "@/components/ui/datagrid"
import { DataGridRowSelectedEvent } from "@/components/ui/datagrid/use-datagrid-row-selection"
import { RowSelectionState } from "@tanstack/react-table"
import React from "react"
import { BiTrash } from "react-icons/bi"
import { IoLibrary } from "react-icons/io5"

type DownloadedChapterListProps = {
    entry: MangaEntry
    data: MangaDownloadData | undefined
}

export type DownloadChapterItem = { provider: string, chapterId: string, chapterNumber: string, queued: boolean, downloaded: boolean }

export function DownloadedChapterList(props: DownloadedChapterListProps) {

    const {
        entry,
        data,
        ...rest
    } = props

    const [showQueued, setShowQueued] = React.useState(false)

    const { deleteChapter, isDeletingChapter } = useDeleteDownloadedMangaChapter(String(entry.mediaId))

    // Transforms {downloaded: Record<string, { chapterId: string, chapterNumber: string }[]>,
    //                            queued: Record<string, { chapterId: string, chapterNumber: string }[]>}
    // to [{provider: string, chapterId: string, queued: boolean, downloaded: boolean}, ...]
    const tableData = React.useMemo(() => {
        let d: DownloadChapterItem[] = []
        if (data) {
            if (!showQueued) {
                for (const provider in data.downloaded) {
                    d = d.concat(data.downloaded[provider].map(ch => ({
                        provider,
                        chapterId: ch.chapterId,
                        chapterNumber: ch.chapterNumber,
                        queued: false,
                        downloaded: true,
                    })))
                }
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
    }, [data, showQueued])

    const columns = React.useMemo(() => defineDataGridColumns<DownloadChapterItem>(() => [
        {
            accessorKey: "chapterNumber",
            header: "Chapter",
            size: 90,
            cell: info => <span>Chapter {info.getValue<string>()}</span>,
        },
        {
            accessorKey: "provider",
            header: "Provider",
            size: 10,
        },
        {
            accessorKey: "chapterId",
            header: "Chapter ID",
            size: 20,
        },
        {
            id: "_actions",
            size: 10,
            enableSorting: false,
            enableGlobalFilter: false,
            cell: ({ row }) => {
                return (
                    <div className="flex justify-end gap-2 items-center w-full">
                        {row.original.queued && <p className="text-[--muted]">Queued</p>}
                        {row.original.downloaded && <p className="text-[--muted] px-1"><IoLibrary className="text-lg" /></p>}
                    </div>
                )
            },
        },
    ]), [tableData])

    const [rowSelection, setRowSelection] = React.useState<RowSelectionState>({})

    const [selectedChapters, setSelectedChapters] = React.useState<DownloadChapterItem[]>([])

    const onSelectChange = React.useCallback((event: DataGridRowSelectedEvent<DownloadChapterItem>) => {
        setSelectedChapters(event.data)
    }, [])

    const handleDeleteSelectedChapters = React.useCallback(() => {
        if (!!selectedChapters.length) {
            for (const chapter of selectedChapters) {
                deleteChapter({
                    mediaId: entry.mediaId,
                    provider: chapter.provider,
                    chapterId: chapter.chapterId,
                    chapterNumber: chapter.chapterNumber,
                })
            }
            setRowSelection({})
            setSelectedChapters([])
        }
    }, [selectedChapters])

    if (!data || !tableData.length) return null

    return (
        <>
            <h3 className="pt-8">Downloads</h3>

            <div className="space-y-4 border rounded-md bg-[--paper] p-4">

                <div className="flex flex-wrap items-center gap-4">
                    <Checkbox
                        label="Show queued"
                        value={showQueued}
                        onValueChange={v => setShowQueued(v as boolean)}
                        fieldClass="w-fit"
                        {...primaryPillCheckboxClass}
                    />
                </div>

                {!!selectedChapters.length && <div
                    className=""
                >
                    <Button
                        onClick={handleDeleteSelectedChapters}
                        intent="alert"
                        size="sm"
                        leftIcon={<BiTrash />}
                        className=""
                    >
                        Delete selected chapters ({selectedChapters?.length})
                    </Button>
                </div>}

                <DataGrid<DownloadChapterItem>
                    columns={columns}
                    data={tableData}
                    rowCount={tableData.length}
                    isLoading={false}
                    rowSelectionPrimaryKey="chapterId"
                    enableRowSelection={row => (row.original.downloaded)}
                    initialState={{
                        pagination: {
                            pageIndex: 0,
                            pageSize: 10,
                        },
                    }}
                    state={{
                        rowSelection,
                    }}
                    onRowSelect={onSelectChange}
                    onRowSelectionChange={setRowSelection}
                    className=""
                />
            </div>
        </>
    )
}
