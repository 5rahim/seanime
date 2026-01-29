import { HibikeManga_ChapterDetails, Manga_ChapterContainer, Manga_Entry } from "@/api/generated/types"
import { useGetMangaEntryDownloadedChapters } from "@/api/hooks/manga.hooks"
import { ChapterReaderDrawer } from "@/app/(main)/manga/_containers/chapter-reader/chapter-reader-drawer"
import { __manga_selectedChapterAtom } from "@/app/(main)/manga/_lib/handle-chapter-reader"
import { useHandleMangaDownloadData } from "@/app/(main)/manga/_lib/handle-manga-downloads"
import { getChapterNumberFromChapter } from "@/app/(main)/manga/_lib/handle-manga-utils"
import { monochromeCheckboxClasses } from "@/components/shared/classnames"
import { IconButton } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { DataGrid, defineDataGridColumns } from "@/components/ui/datagrid"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { useSetAtom } from "jotai"
import React from "react"
import { GiOpenBook } from "react-icons/gi"

type OfflineChapterListProps = {
    entry: Manga_Entry | undefined
    children?: React.ReactNode
}

export function OfflineChapterList(props: OfflineChapterListProps) {

    const {
        entry,
        children,
        ...rest
    } = props

    const { data: chapterContainers, isLoading } = useGetMangaEntryDownloadedChapters(entry?.mediaId)

    /**
     * Set selected chapter
     */
    const setSelectedChapter = useSetAtom(__manga_selectedChapterAtom)

    // Load download data
    useHandleMangaDownloadData(entry?.mediaId)

    const chapters = React.useMemo(() => {
        return chapterContainers?.flatMap(n => n.chapters)?.filter(Boolean) ?? []
    }, [chapterContainers])


    const chapterNumbersMap = React.useMemo(() => {
        const map = new Map<string, number>()

        for (const chapter of chapters) {
            map.set(chapter.id, getChapterNumberFromChapter(chapter.chapter))
        }

        return map
    }, [chapterContainers])

    const [selectedChapterContainer, setSelectedChapterContainer] = React.useState<Manga_ChapterContainer | undefined>(undefined)

    const columns = React.useMemo(() => defineDataGridColumns<HibikeManga_ChapterDetails>(() => [
        {
            accessorKey: "title",
            header: "Name",
            size: 90,
        },
        {
            accessorKey: "provider",
            header: "Provider",
            size: 10,
            enableSorting: true,
        },
        {
            id: "number",
            header: "Number",
            size: 10,
            enableSorting: true,
            accessorFn: (row) => {
                return chapterNumbersMap.get(row.id)
            },
        },
        {
            id: "_actions",
            size: 5,
            enableSorting: false,
            enableGlobalFilter: false,
            cell: ({ row }) => {
                return (
                    <div className="flex justify-end gap-2 items-center w-full">
                        <IconButton
                            intent="gray-subtle"
                            size="sm"
                            onClick={() => {
                                // setProvider(row.original.provider)
                                setSelectedChapterContainer(chapterContainers?.find(c => c.provider === row.original.provider))
                                React.startTransition(() => {
                                    setSelectedChapter({
                                        chapterId: row.original.id,
                                        chapterNumber: row.original.chapter,
                                        provider: row.original.provider,
                                        mediaId: Number(entry?.mediaId),
                                    })
                                })
                            }}
                            icon={<GiOpenBook />}
                        />
                    </div>
                )
            },
        },
    ]), [entry, chapterNumbersMap])

    const [showUnreadChapter, setShowUnreadChapter] = React.useState(false)

    const retainUnreadChapters = React.useCallback((chapter: HibikeManga_ChapterDetails) => {
        if (!entry?.listData || !chapterNumbersMap.has(chapter.id) || !entry?.listData?.progress) return true

        const chapterNumber = chapterNumbersMap.get(chapter.id)
        return !!chapterNumber && chapterNumber > entry.listData?.progress
    }, [chapterNumbersMap, chapterContainers, entry])

    const unreadChapters = React.useMemo(() => chapters.filter(ch => retainUnreadChapters(ch)) ?? [],
        [chapters, entry])

    React.useEffect(() => {
        setShowUnreadChapter(!!unreadChapters.length)
    }, [unreadChapters])

    const tableChapters = React.useMemo(() => {
        return showUnreadChapter ? unreadChapters : chapters
    }, [showUnreadChapter, chapters, unreadChapters])

    if (!entry || isLoading) return <LoadingSpinner />

    return (
        <>
            <div className="space-y-4 border rounded-[--radius-md] bg-[--paper] p-4">

                <div className="flex flex-wrap items-center gap-4">
                    <Checkbox
                        label="Show unread"
                        value={showUnreadChapter}
                        onValueChange={v => setShowUnreadChapter(v as boolean)}
                        fieldClass="w-fit"
                        {...monochromeCheckboxClasses}
                    />
                </div>

                <DataGrid<HibikeManga_ChapterDetails>
                    columns={columns}
                    data={tableChapters}
                    rowCount={tableChapters?.length || 0}
                    isLoading={!tableChapters}
                    rowSelectionPrimaryKey="id"
                    initialState={{
                        pagination: {
                            pageIndex: 0,
                            pageSize: 10,
                        },
                        sorting: [
                            {
                                id: "number",
                                desc: false,
                            },
                        ],
                    }}
                    className=""
                />

                {(!!selectedChapterContainer) && <ChapterReaderDrawer
                    entry={entry}
                    chapterIdToNumbersMap={chapterNumbersMap}
                    chapterContainer={selectedChapterContainer}
                />}
            </div>
        </>
    )
}
