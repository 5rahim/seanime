import { OfflineMangaEntry } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot.types"
import { __manga_selectedProviderAtom } from "@/app/(main)/manga/_lib/manga.hooks"
import { MangaChapterDetails } from "@/app/(main)/manga/_lib/manga.types"
import { getChapterNumberFromChapter } from "@/app/(main)/manga/_lib/manga.utils"
import { __manga_selectedChapterAtom, ChapterReaderDrawer } from "@/app/(main)/manga/entry/_containers/chapter-reader/chapter-reader-drawer"
import { primaryPillCheckboxClass } from "@/components/shared/styling/classnames"
import { PageWrapper } from "@/components/shared/styling/page-wrapper"
import { IconButton } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { DataGrid, defineDataGridColumns } from "@/components/ui/datagrid"
import { useSetAtom } from "jotai"
import React from "react"
import { GiOpenBook } from "react-icons/gi"

type OfflineChapterListProps = {
    entry: OfflineMangaEntry | undefined
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


    const chapterNumbersMap = React.useMemo(() => {
        const map = new Map<string, number>()

        for (const chapter of entry?.chapterContainer?.chapters ?? []) {
            map.set(chapter.id, getChapterNumberFromChapter(chapter.chapter))
        }

        return map
    }, [entry?.chapterContainer])

    const columns = React.useMemo(() => defineDataGridColumns<MangaChapterDetails>(() => [
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

    const retainUnreadChapters = React.useCallback((chapter: MangaChapterDetails) => {
        if (!entry?.listData || !chapterNumbersMap.has(chapter.id) || !entry?.listData?.progress) return true

        const chapterNumber = chapterNumbersMap.get(chapter.id)
        return !!chapterNumber && chapterNumber > entry.listData?.progress
    }, [chapterNumbersMap, entry?.chapterContainer, entry])

    const unreadChapters = React.useMemo(() => entry?.chapterContainer?.chapters?.filter(ch => retainUnreadChapters(ch)) ?? [],
        [entry?.chapterContainer, entry])

    React.useEffect(() => {
        setShowUnreadChapter(!!unreadChapters.length)
    }, [unreadChapters])

    const chapters = React.useMemo(() => {
        return showUnreadChapter ? unreadChapters : entry?.chapterContainer?.chapters
    }, [showUnreadChapter, entry?.chapterContainer?.chapters, unreadChapters])

    return (
        <PageWrapper className="p-4">
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

                <DataGrid<MangaChapterDetails>
                    columns={columns}
                    data={chapters}
                    rowCount={chapters?.length || 0}
                    isLoading={!chapters}
                    rowSelectionPrimaryKey="id"
                    initialState={{
                        pagination: {
                            pageIndex: 0,
                            pageSize: 10,
                        },
                    }}
                    className=""
                />

                {!!entry?.chapterContainer && <ChapterReaderDrawer
                    entry={entry}
                    chapterIdToNumbersMap={chapterNumbersMap}
                    chapterContainer={entry?.chapterContainer}
                />}
            </div>
        </PageWrapper>
    )
}
