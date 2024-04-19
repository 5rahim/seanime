import {
    __manga_selectedProviderAtom,
    useClearMangaCache,
    useDownloadMangaChapter,
    useMangaChapterContainer,
} from "@/app/(main)/manga/_lib/manga.hooks"
import { MANGA_PROVIDER_OPTIONS, MangaChapterDetails, MangaDownloadData, MangaEntry } from "@/app/(main)/manga/_lib/manga.types"
import { useMangaChapterListRowSelection, useMangaDownloadDataUtils } from "@/app/(main)/manga/_lib/manga.utils"
import { ChapterListBulkActions } from "@/app/(main)/manga/entry/_containers/chapter-list/_components/chapter-list-bulk-actions"
import { DownloadedChapterList } from "@/app/(main)/manga/entry/_containers/chapter-list/_components/downloaded-chapter-list"
import { __manga_selectedChapterAtom, ChapterReaderDrawer } from "@/app/(main)/manga/entry/_containers/chapter-reader/chapter-reader-drawer"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/application/confirmation-dialog"
import { LuffyError } from "@/components/shared/luffy-error"
import { primaryPillCheckboxClass } from "@/components/shared/styling/classnames"
import { Button, IconButton } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { DataGrid, defineDataGridColumns } from "@/components/ui/datagrid"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Select } from "@/components/ui/select"
import { MangaDetailsByIdQuery } from "@/lib/anilist/gql/graphql"
import { useAtom, useSetAtom } from "jotai/react"
import React from "react"
import { FaDownload, FaRedo } from "react-icons/fa"
import { GiOpenBook } from "react-icons/gi"
import { IoBookOutline, IoLibrary } from "react-icons/io5"

type ChapterListProps = {
    mediaId: string | null
    entry: MangaEntry
    details: MangaDetailsByIdQuery["Media"] | undefined
    downloadData: MangaDownloadData | undefined
    downloadDataLoading: boolean
}

export function ChapterList(props: ChapterListProps) {

    const {
        mediaId,
        entry,
        details,
        downloadData,
        downloadDataLoading,
        ...rest
    } = props

    /**
     * Fetch chapter container
     */
    const { chapterContainer, chapterIdToNumbersMap, chapterContainerError, chapterContainerLoading } = useMangaChapterContainer(mediaId)

    const [showUnreadChapter, setShowUnreadChapter] = React.useState(false)
    const [showDownloadedChapters, setShowDownloadedChapters] = React.useState(false)

    /**
     * Current provider
     */
    const [provider, setProvider] = useAtom(__manga_selectedProviderAtom)
    /**
     * Set selected chapter
     */
    const setSelectedChapter = useSetAtom(__manga_selectedChapterAtom)
    /**
     * Clear manga cache
     */
    const { clearMangaCache, isClearingMangaCache } = useClearMangaCache()
    /**
     * Download chapter
     */
    const { downloadChapters, isSendingDownloadRequest } = useDownloadMangaChapter(mediaId)
    /**
     * Download data utils
     */
    const {
        isChapterQueued,
        isChapterDownloaded,
        getProviderNumberOfDownloadedChapters,
    } = useMangaDownloadDataUtils(downloadData, downloadDataLoading)

    /**
     * Function to filter unread chapters
     */
    const retainUnreadChapters = React.useCallback((chapter: MangaChapterDetails) => {
        if (!entry.listData || !chapterIdToNumbersMap.has(chapter.id) || !entry.listData?.progress) return true

        const chapterNumber = chapterIdToNumbersMap.get(chapter.id)
        return !!chapterNumber && chapterNumber > entry.listData?.progress
    }, [chapterIdToNumbersMap, chapterContainer, entry])

    const confirmReloadSource = useConfirmationDialog({
        title: "Reload sources",
        actionIntent: "primary",
        actionText: "Reload",
        description: "This action will empty the cache for this manga and fetch the latest data from the selected source.",
        onConfirm: () => {
            if (mediaId) {
                clearMangaCache({ mediaId: Number(mediaId) })
            }
        },
    })

    /**
     * Chapter columns
     */
    const columns = React.useMemo(() => defineDataGridColumns<MangaChapterDetails>(() => [
        {
            accessorKey: "title",
            header: "Name",
            size: 90,
        },
        {
            header: "Number",
            size: 10,
            enableSorting: true,
            accessorFn: (row) => {
                return chapterIdToNumbersMap.get(row.id)
            },
        },
        {
            id: "_actions",
            size: 10,
            enableSorting: false,
            enableGlobalFilter: false,
            cell: ({ row }) => {
                return (
                    <div className="flex justify-end gap-2 items-center w-full">
                        {(!isChapterDownloaded(row.original) && !isChapterQueued(row.original)) && <IconButton
                            intent="gray-basic"
                            size="sm"
                            disabled={isSendingDownloadRequest}
                            onClick={() => downloadChapters([row.original])}
                            icon={<FaDownload className="text-sm" />}
                        />}
                        {isChapterQueued(row.original) && <p className="text-[--muted]">Queued</p>}
                        {isChapterDownloaded(row.original) && <p className="text-[--muted] px-1"><IoLibrary className="text-lg" /></p>}
                        <IconButton
                            intent="gray-subtle"
                            size="sm"
                            onClick={() => setSelectedChapter({
                                chapterId: row.original.id,
                                chapterNumber: row.original.chapter,
                                provider: row.original.provider,
                                mediaId: Number(mediaId),
                            })}
                            icon={<GiOpenBook />}
                        />
                    </div>
                )
            },
        },
    ]), [chapterIdToNumbersMap, isSendingDownloadRequest, isChapterDownloaded, downloadData, mediaId])

    const unreadChapters = React.useMemo(() => chapterContainer?.chapters?.filter(ch => retainUnreadChapters(ch)) ?? [], [chapterContainer, entry])
    const allChapters = React.useMemo(() => chapterContainer?.chapters?.toReversed() ?? [], [chapterContainer])

    /**
     * Set "showUnreadChapter" state if there are unread chapters
     */
    React.useEffect(() => {
        setShowUnreadChapter(!!unreadChapters.length)
    }, [unreadChapters])

    /**
     * Filter chapters based on state
     */
    const chapters = React.useMemo(() => {
        let d = showUnreadChapter ? unreadChapters : allChapters
        if (showDownloadedChapters) {
            d = d.filter(ch => isChapterDownloaded(ch) || isChapterQueued(ch))
        }
        return d
    }, [showUnreadChapter, unreadChapters, allChapters, showDownloadedChapters, isChapterDownloaded, isChapterQueued, downloadData])


    const {
        rowSelectedChapters,
        onRowSelectionChange,
        rowSelection,
        setRowSelection,
        resetRowSelection,
    } = useMangaChapterListRowSelection()

    React.useEffect(() => {
        resetRowSelection()
    }, [chapters])

    return (
        <div
            className="space-y-2"
        >

            <div className="flex gap-2 items-center">
                <Select
                    fieldClass="w-fit"
                    options={MANGA_PROVIDER_OPTIONS}
                    value={provider}
                    onValueChange={setProvider}
                    leftAddon="Source"
                    intent="filled"
                    size="sm"
                    disabled={isClearingMangaCache}
                />

                <Button
                    leftIcon={<FaRedo />}
                    intent="white-subtle"
                    onClick={() => confirmReloadSource.open()}
                    loading={isClearingMangaCache}
                    size="sm"
                >
                    Reload sources
                </Button>
            </div>

            {(chapterContainerLoading || isClearingMangaCache) ? <LoadingSpinner /> : (
                chapterContainerError ? <LuffyError title="Oops!"><p>No chapters found</p></LuffyError> : (
                    <>

                        {chapterContainer?.chapters?.length === 0 && (
                            <LuffyError title="No chapters found"><p>Try another source</p></LuffyError>
                        )}

                        {!!chapterContainer?.chapters?.length && (
                            <>
                                <div className="flex gap-2 items-center w-full pb-2">
                                    <h3 className="px-1">Chapters</h3>
                                    <div className="flex flex-1"></div>
                                    <div>
                                        {!!unreadChapters?.length && <Button
                                            intent="white"
                                            rounded
                                            leftIcon={<IoBookOutline />}
                                            onClick={() => {
                                                setSelectedChapter({
                                                    chapterId: unreadChapters[0].id,
                                                    chapterNumber: unreadChapters[0].chapter,
                                                    provider: unreadChapters[0].provider,
                                                    mediaId: Number(mediaId),
                                                })
                                            }}
                                        >
                                            Continue reading
                                        </Button>}
                                    </div>
                                </div>

                                <div className="space-y-4 border rounded-md bg-[--paper] p-4">

                                    <div className="flex flex-wrap items-center gap-4">
                                        <Checkbox
                                            label="Show unread"
                                            value={showUnreadChapter}
                                            onValueChange={v => setShowUnreadChapter(v as boolean)}
                                            fieldClass="w-fit"
                                            {...primaryPillCheckboxClass}
                                        />
                                        <Checkbox
                                            label={<span className="flex gap-2 items-center"><IoLibrary /> Show downloaded</span>}
                                            value={showDownloadedChapters}
                                            onValueChange={v => setShowDownloadedChapters(v as boolean)}
                                            fieldClass="w-fit"
                                            {...primaryPillCheckboxClass}
                                        />
                                    </div>

                                    <ChapterListBulkActions
                                        rowSelectedChapters={rowSelectedChapters}
                                        onDownloadSelected={chapters => {
                                            downloadChapters(chapters)
                                            resetRowSelection()
                                        }}
                                    />

                                    <DataGrid<MangaChapterDetails>
                                        columns={columns}
                                        data={chapters}
                                        rowCount={chapters.length}
                                        isLoading={chapterContainerLoading}
                                        rowSelectionPrimaryKey="id"
                                        enableRowSelection={row => (!isChapterDownloaded(row.original) && !isChapterQueued(row.original))}
                                        initialState={{
                                            pagination: {
                                                pageIndex: 0,
                                                pageSize: 10,
                                            },
                                        }}
                                        state={{
                                            rowSelection,
                                        }}
                                        onRowSelect={onRowSelectionChange}
                                        onRowSelectionChange={setRowSelection}
                                        className=""
                                    />
                                </div>
                            </>
                        )}

                        {chapterContainer && <ChapterReaderDrawer
                            entry={entry}
                            chapterContainer={chapterContainer}
                            chapterIdToNumbersMap={chapterIdToNumbersMap}
                        />}
                    </>
                )
            )}

            <DownloadedChapterList
                entry={entry}
                data={downloadData}
            />

            <ConfirmationDialog {...confirmReloadSource} />
        </div>
    )
}

