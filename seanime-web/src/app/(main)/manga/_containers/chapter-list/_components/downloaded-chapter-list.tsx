// /* -------------------------------------------------------------------------------------------------
//  * Download List
//  * -----------------------------------------------------------------------------------------------*/


import { Manga_Entry, Manga_MediaDownloadData, Manga_Provider } from "@/api/generated/types"
import { useDeleteMangaChapterDownload } from "@/api/hooks/manga_download.hooks"
import { __manga_selectedChapterAtom } from "@/app/(main)/manga/_containers/chapter-reader/chapter-reader-drawer"
import { __manga_selectedProviderAtom } from "@/app/(main)/manga/_lib/manga.hooks"
import { getChapterNumberFromChapter } from "@/app/(main)/manga/_lib/manga.utils"
import { primaryPillCheckboxClass } from "@/components/shared/styling/classnames"
import { Button, IconButton } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { DataGrid, defineDataGridColumns } from "@/components/ui/datagrid"
import { DataGridRowSelectedEvent } from "@/components/ui/datagrid/use-datagrid-row-selection"
import { RowSelectionState } from "@tanstack/react-table"
import { useSetAtom } from "jotai"
import { useAtom } from "jotai/react"
import React from "react"
import { BiTrash } from "react-icons/bi"
import { GiOpenBook } from "react-icons/gi"
import { IoLibrary } from "react-icons/io5"

type DownloadedChapterListProps = {
    entry: Manga_Entry
    data: Manga_MediaDownloadData | undefined
}

export type DownloadChapterItem = { provider: string, chapterId: string, chapterNumber: string, queued: boolean, downloaded: boolean }

export function DownloadedChapterList(props: DownloadedChapterListProps) {

    const {
        entry,
        data,
        ...rest
    } = props

    const [provider, setProvider] = useAtom(__manga_selectedProviderAtom)

    /**
     * Set selected chapter
     */
    const setSelectedChapter = useSetAtom(__manga_selectedChapterAtom)

    const [showQueued, setShowQueued] = React.useState(false)

    const { mutate: deleteChapter, isPending: isDeletingChapter } = useDeleteMangaChapterDownload(String(entry.mediaId), provider)

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

    const chapterIdsToNumber = React.useMemo(() => {
        const map = new Map<string, number>()

        for (const chapter of tableData ?? []) {
            map.set(chapter.chapterId, getChapterNumberFromChapter(chapter.chapterNumber))
        }

        return map
    }, [tableData])

    const columns = React.useMemo(() => defineDataGridColumns<DownloadChapterItem>(() => [
        {
            accessorKey: "chapterNumber",
            header: "Chapter",
            size: 90,
            cell: info => <span>Chapter {info.getValue<string>()}</span>,
        },
        {
            header: "Number",
            size: 10,
            enableSorting: true,
            accessorFn: (row) => {
                return chapterIdsToNumber.get(row.chapterId)
            },
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
            cell: info => <span className="text-[--muted] text-sm italic">{info.getValue<string>()}</span>,
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

                        {row.original.downloaded && <IconButton
                            intent="gray-subtle"
                            size="sm"
                            onClick={() => {
                                /**
                                 * Set the provider to the one of the selected chapter
                                 * This is because the provider is needed to fetch the chapter pages
                                 */
                                setProvider(row.original.provider as Manga_Provider)
                                React.startTransition(() => {
                                    setSelectedChapter({
                                        chapterId: row.original.chapterId,
                                        chapterNumber: row.original.chapterNumber,
                                        provider: row.original.provider,
                                        mediaId: Number(entry.mediaId),
                                    })
                                })
                            }}
                            icon={<GiOpenBook />}
                        />}
                    </div>
                )
            },
        },
    ]), [tableData, entry?.mediaId, chapterIdsToNumber])

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

    if (!data || Object.keys(data.downloaded).length === 0 && Object.keys(data.queued).length === 0) return null

    return (
        <>
            <h3 className="pt-8">Downloaded chapters</h3>

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
                        loading={isDeletingChapter}
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
                        sorting: [
                            {
                                id: "chapterNumber",
                                desc: false,
                            },
                        ],
                    }}
                    state={{
                        rowSelection,
                    }}
                    hideColumns={[
                        {
                            below: 1000,
                            hide: ["chapterId", "number", "provider"],
                        },
                    ]}
                    onSortingChange={console.log}
                    onRowSelect={onSelectChange}
                    onRowSelectionChange={setRowSelection}
                    className=""
                />
            </div>
        </>
    )
}
