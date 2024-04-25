import { Manga_ChapterContainer, Manga_ChapterDetails, Offline_MangaEntry } from "@/api/generated/types"
import { __manga_selectedChapterAtom, ChapterReaderDrawer } from "@/app/(main)/manga/_containers/chapter-reader/chapter-reader-drawer"
import { __manga_selectedProviderAtom } from "@/app/(main)/manga/_lib/handle-manga"
import { getChapterNumberFromChapter } from "@/app/(main)/manga/_lib/handle-manga-utils"
import { primaryPillCheckboxClass } from "@/components/shared/classnames"
import { IconButton } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { DataGrid, defineDataGridColumns } from "@/components/ui/datagrid"
import { useSetAtom } from "jotai"
import React from "react"
import { GiOpenBook } from "react-icons/gi"

type OfflineChapterListProps = {
    entry: Offline_MangaEntry | undefined
    children?: React.ReactNode
}

export function OfflineChapterList(props: OfflineChapterListProps) {

    const {
        entry,
        children,
        ...rest
    } = props

    /**
     * Set selected chapter
     */
    const setSelectedChapter = useSetAtom(__manga_selectedChapterAtom)

    const setProvider = useSetAtom(__manga_selectedProviderAtom)

    const chapters = React.useMemo(() => {
        return entry?.chapterContainers?.flatMap(n => n.chapters)?.filter(Boolean) ?? []
    }, [entry?.chapterContainers])


    const chapterNumbersMap = React.useMemo(() => {
        const map = new Map<string, number>()

        for (const chapter of chapters) {
            map.set(chapter.id, getChapterNumberFromChapter(chapter.chapter))
        }

        return map
    }, [entry?.chapterContainers])

    const [selectedChapterContainer, setSelectedChapterContainer] = React.useState<Manga_ChapterContainer | undefined>(undefined)

    const columns = React.useMemo(() => defineDataGridColumns<Manga_ChapterDetails>(() => [
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
                                setProvider(row.original.provider)
                                setSelectedChapterContainer(entry?.chapterContainers?.find(c => c.provider === row.original.provider))
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

    const retainUnreadChapters = React.useCallback((chapter: Manga_ChapterDetails) => {
        if (!entry?.listData || !chapterNumbersMap.has(chapter.id) || !entry?.listData?.progress) return true

        const chapterNumber = chapterNumbersMap.get(chapter.id)
        return !!chapterNumber && chapterNumber > entry.listData?.progress
    }, [chapterNumbersMap, entry?.chapterContainers, entry])

    const unreadChapters = React.useMemo(() => chapters.filter(ch => retainUnreadChapters(ch)) ?? [],
        [chapters, entry])

    React.useEffect(() => {
        setShowUnreadChapter(!!unreadChapters.length)
    }, [unreadChapters])

    const tableChapters = React.useMemo(() => {
        return showUnreadChapter ? unreadChapters : chapters
    }, [showUnreadChapter, chapters, unreadChapters])

    if (!entry) return null

    return (
        <>
            <div className="space-y-4 border rounded-md bg-[--paper] p-4">

                <div className="flex flex-wrap items-center gap-4">
                    <Checkbox
                        label="Show unread"
                        value={showUnreadChapter}
                        onValueChange={v => setShowUnreadChapter(v as boolean)}
                        fieldClass="w-fit"
                        {...primaryPillCheckboxClass}
                    />
                </div>

                <DataGrid<Manga_ChapterDetails>
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
