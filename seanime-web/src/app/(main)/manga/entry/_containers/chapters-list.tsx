import { __manga__chapterDownloadsDrawerIsOpenAtom } from "@/app/(main)/manga/_containers/chapter-downloads/chapter-downloads-drawer"
import {
    __manga_selectedProviderAtom,
    useClearMangaCache,
    useDownloadMangaChapter,
    useMangaChapterContainer,
} from "@/app/(main)/manga/_lib/manga.hooks"
import { MANGA_PROVIDER_OPTIONS, MangaChapterDetails, MangaDownloadData, MangaEntry } from "@/app/(main)/manga/_lib/manga.types"
import { useMangaDownloadDataUtils } from "@/app/(main)/manga/_lib/manga.utils"
import { __manga_selectedChapterAtom, ChapterReaderDrawer } from "@/app/(main)/manga/entry/_containers/chapter-reader/chapter-reader-drawer"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/application/confirmation-dialog"
import { LuffyError } from "@/components/shared/luffy-error"
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion"
import { Button, IconButton } from "@/components/ui/button"
import { DataGrid, defineDataGridColumns } from "@/components/ui/datagrid"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Select } from "@/components/ui/select"
import { MangaDetailsByIdQuery } from "@/lib/anilist/gql/graphql"
import { useAtom, useSetAtom } from "jotai/react"
import React from "react"
import { BiBookAlt } from "react-icons/bi"
import { FaDownload, FaRedo } from "react-icons/fa"
import { GiOpenBook } from "react-icons/gi"
import { IoBookOutline, IoLibrary } from "react-icons/io5"

type ChaptersListProps = {
    mediaId: string | null
    entry: MangaEntry
    details: MangaDetailsByIdQuery["Media"] | undefined
    downloadData: MangaDownloadData | undefined
    downloadDataLoading: boolean
}

export function ChaptersList(props: ChaptersListProps) {

    const {
        mediaId,
        entry,
        details,
        downloadData,
        downloadDataLoading,
        ...rest
    } = props

    const { chapterContainer, chapterIdToNumbersMap, chapterContainerError, chapterContainerLoading } = useMangaChapterContainer(mediaId)

    const [provider, setProvider] = useAtom(__manga_selectedProviderAtom)

    const { clearMangaCache, isClearingMangaCache } = useClearMangaCache()

    const setSelectedChapter = useSetAtom(__manga_selectedChapterAtom)

    const { downloadChapter, isSendingDownloadRequest } = useDownloadMangaChapter(mediaId)

    const { isChapterQueued, isChapterDownloaded, getProviderNumberOfDownloadedChapters } = useMangaDownloadDataUtils(downloadData,
        downloadDataLoading)

    const openDownloadQueue = useSetAtom(__manga__chapterDownloadsDrawerIsOpenAtom)

    const retainUnreadChapters = React.useCallback((chapter: MangaChapterDetails) => {
        if (!entry.listData || !chapterIdToNumbersMap.has(chapter.id) || !entry.listData?.progress) return true

        const chapterNumber = chapterIdToNumbersMap.get(chapter.id)
        return chapterNumber && chapterNumber > entry.listData?.progress
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
     * Chapter tables
     */
    const columns = React.useMemo(() => defineDataGridColumns<MangaChapterDetails>(() => [
        {
            accessorKey: "title",
            header: "Name",
            size: 10,
        },
        {
            header: "Number",
            size: 90,
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
                            onClick={() => downloadChapter(row.original)}
                            icon={<FaDownload className="text-sm" />}
                        />}
                        {isChapterQueued(row.original) && <p className="text-[--muted]">Queued</p>}
                        {isChapterDownloaded(row.original) && <p className="text-[--muted] px-1"><IoLibrary className="text-lg" /></p>}
                        <IconButton
                            intent="gray-subtle"
                            size="sm"
                            onClick={() => setSelectedChapter(row.original)}
                            icon={<GiOpenBook />}
                        />
                    </div>
                )
            },
        },
    ]), [chapterIdToNumbersMap, isSendingDownloadRequest, isChapterDownloaded, downloadData])

    const unreadChapters = React.useMemo(() => chapterContainer?.chapters?.filter(ch => retainUnreadChapters(ch)) ?? [], [chapterContainer, entry])
    const chapters = React.useMemo(() => chapterContainer?.chapters?.toReversed() ?? [], [chapterContainer])


    return (
        <div
            className="space-y-2"
        >

            <Button onClick={() => openDownloadQueue(true)}>
                Queue
            </Button>

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
                chapterContainerError ? <LuffyError title="Oops!"><p>Failed to fetch chapters</p></LuffyError> : (
                    <>

                        {chapterContainer?.chapters?.length === 0 && (
                            <LuffyError title="No chapters found"><p>Try another source</p></LuffyError>
                        )}

                        <Accordion
                            type="single"
                            className="!py-4"
                            triggerClass="text-[--muted] dark:data-[state=open]:text-white px-0 dark:hover:bg-transparent hover:bg-transparent dark:hover:text-white hover:text-black"
                            itemClass="border-b"
                            contentClass="pb-8"
                            collapsible
                            defaultValue={!unreadChapters.length ? "all" : undefined}
                        >
                            <AccordionItem value="all">
                                <AccordionTrigger>
                                    <h3 className="flex p-1 gap-2 items-center"><BiBookAlt className="text-gray-300" /> All chapters</h3>
                                </AccordionTrigger>
                                <AccordionContent className="p-0 pb-1 space-y-2">
                                    <DataGrid<MangaChapterDetails>
                                        columns={columns}
                                        data={chapters}
                                        rowCount={chapters.length}
                                        isLoading={chapterContainerLoading}
                                        rowSelectionPrimaryKey={"id"}
                                        initialState={{
                                            pagination: {
                                                pageIndex: 0,
                                                pageSize: 10,
                                            },
                                        }}
                                        className="border rounded-md bg-[--paper] p-4"
                                    />
                                    {/*{chapterContainer?.chapters?.toReversed()?.map((chapter) => (*/}
                                    {/*    <ChapterItem*/}
                                    {/*        chapter={chapter}*/}
                                    {/*        key={chapter.id}*/}
                                    {/*        // chapterBackups={chapterBackups}*/}
                                    {/*        // handleDownloadChapter={handleDownloadChapter}*/}
                                    {/*        // downloadProgressMap={downloadProgressMap}*/}
                                    {/*        // isSendingDownloadRequest={false}*/}
                                    {/*    />*/}
                                    {/*))}*/}
                                </AccordionContent>
                            </AccordionItem>
                        </Accordion>


                        {!!unreadChapters?.length && (
                            <>
                                <div className="flex gap-2 items-center w-full pb-2">
                                    <h3 className="px-1">Unread chapters</h3>
                                    <div className="flex flex-1"></div>
                                    <div>
                                        <Button
                                            intent="white"
                                            rounded
                                            leftIcon={<IoBookOutline />}
                                            onClick={() => {
                                                setSelectedChapter(unreadChapters[0])
                                            }}
                                        >
                                            Continue reading
                                        </Button>
                                    </div>
                                </div>
                                <div className="space-y-2">
                                    {/*{chapterContainer?.chapters?.filter(ch => retainUnreadChapters(ch)).map((chapter) => (*/}
                                    {/*    <ChapterItem*/}
                                    {/*        chapter={chapter}*/}
                                    {/*        key={chapter.id}*/}
                                    {/*        // chapterBackups={chapterBackups}*/}
                                    {/*        // handleDownloadChapter={handleDownloadChapter}*/}
                                    {/*        // downloadProgressMap={downloadProgressMap}*/}
                                    {/*        // isSendingDownloadRequest={false}*/}
                                    {/*    />*/}
                                    {/*))}*/}
                                    <DataGrid<MangaChapterDetails>
                                        columns={columns}
                                        data={unreadChapters}
                                        rowCount={unreadChapters.length}
                                        isLoading={chapterContainerLoading}
                                        rowSelectionPrimaryKey={"id"}
                                        initialState={{
                                            pagination: {
                                                pageIndex: 0,
                                                pageSize: 10,
                                            },
                                        }}
                                        className="border rounded-md bg-[--paper] p-4"
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

            <ConfirmationDialog {...confirmReloadSource} />
        </div>
    )
}
