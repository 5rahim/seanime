import { __manga_selectedProviderAtom, useEmptyMangaCache, useMangaChapterContainer } from "@/app/(main)/manga/_lib/manga.hooks"
import { manga_providers_options, MangaChapterDetails, MangaEntry, MangaEntryBackups } from "@/app/(main)/manga/_lib/manga.types"
import { __manga_selectedChapterAtom, ChapterDrawer } from "@/app/(main)/manga/entry/_containers/chapter-drawer/chapter-drawer"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/application/confirmation-dialog"
import { LuffyError } from "@/components/shared/luffy-error"
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion"
import { Button, IconButton } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Select } from "@/components/ui/select"
import { MangaDetailsByIdQuery } from "@/lib/anilist/gql/graphql"
import { atomWithImmer } from "jotai-immer"
import { useAtom, useSetAtom } from "jotai/react"
import React from "react"
import { BiBookAlt } from "react-icons/bi"
import { FaRedo } from "react-icons/fa"
import { GiOpenBook } from "react-icons/gi"

type ChaptersListProps = {
    mediaId: string | null
    entry: MangaEntry
    details?: MangaDetailsByIdQuery["Media"]
}

const downloadProgressMapAtom = atomWithImmer<Record<string, number>>({})

export function ChaptersList(props: ChaptersListProps) {

    const {
        mediaId,
        entry,
        details,
        ...rest
    } = props


    const { chapterContainer, chapterIdToNumbersMap, chapterContainerError, chapterContainerLoading } = useMangaChapterContainer(mediaId)

    const [provider, setProvider] = useAtom(__manga_selectedProviderAtom)

    const { emptyMangaCache, isEmptyingMangaCache } = useEmptyMangaCache()

    // SHELVED
    // const { chapterBackups, chapterBackupsLoading } = useMangaEntryBackups(mediaId)
    // const [downloadProgressMap, setDownloadProgressMap] = useAtom(downloadProgressMapAtom)
    // const qc = useQueryClient()
    // const { downloadChapter, isSendingDownloadRequest } = useDownloadMangaChapter(mediaId)
    // useWebsocketMessageListener<{ chapterId: string, number: number } | null>({
    //     type: WSEvents.MANGA_DOWNLOADER_DOWNLOADING_PROGRESS,
    //     onMessage: data => {
    //         if (!data) return
    //
    //         if (data.number === 0) {
    //             setDownloadProgressMap(draft => {
    //                 delete draft[data.chapterId]
    //             })
    //             qc.refetchQueries({ queryKey: ["get-manga-entry-backups"] })
    //         } else {
    //             setDownloadProgressMap(draft => {
    //                 draft[data.chapterId] = data.number
    //             })
    //         }
    //     },
    // })
    // const handleDownloadChapter = React.useCallback((chapter: MangaChapterDetails) => {
    //     // shelved
    //     // downloadChapter(chapter)
    // }, [])

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
                emptyMangaCache({ mediaId: Number(mediaId) })
            }
        },
    })


    return (
        <div
            className="space-y-2"
        >

            <div className="flex gap-2 items-center">
                <Select
                    fieldClass="w-fit"
                    options={manga_providers_options}
                    value={provider}
                    onValueChange={setProvider}
                    leftAddon="Source"
                    intent="filled"
                    disabled={isEmptyingMangaCache}
                />

                <Button
                    leftIcon={<FaRedo />}
                    intent="white-subtle"
                    onClick={() => confirmReloadSource.open()}
                    loading={isEmptyingMangaCache}
                >
                    Reload sources
                </Button>
            </div>

            {(chapterContainerLoading || isEmptyingMangaCache) ? <LoadingSpinner /> : (
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
                        >
                            <AccordionItem value="all">
                                <AccordionTrigger>
                                    <h3 className="flex gap-2 items-center"><BiBookAlt className="text-gray-300" /> All chapters</h3>
                                </AccordionTrigger>
                                <AccordionContent className="p-0 space-y-2 max-h-[75dvh] overflow-y-auto">
                                    {chapterContainer?.chapters?.toReversed()?.map((chapter) => (
                                        <ChapterItem
                                            chapter={chapter}
                                            key={chapter.id}
                                            // chapterBackups={chapterBackups}
                                            // handleDownloadChapter={handleDownloadChapter}
                                            // downloadProgressMap={downloadProgressMap}
                                            // isSendingDownloadRequest={false}
                                        />
                                    ))}
                                </AccordionContent>
                            </AccordionItem>
                        </Accordion>


                        <h3>Unread chapters</h3>
                        <div className="p-0 space-y-2 max-h-[75dvh] overflow-y-auto">
                            {chapterContainer?.chapters?.filter(ch => retainUnreadChapters(ch)).map((chapter) => (
                                <ChapterItem
                                    chapter={chapter}
                                    key={chapter.id}
                                    // chapterBackups={chapterBackups}
                                    // handleDownloadChapter={handleDownloadChapter}
                                    // downloadProgressMap={downloadProgressMap}
                                    // isSendingDownloadRequest={false}
                                />
                            ))}
                        </div>

                        {chapterContainer && <ChapterDrawer
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


type ChapterItemProps = {
    chapter: MangaChapterDetails
    chapterBackups?: MangaEntryBackups | undefined
    handleDownloadChapter?: (chapter: MangaChapterDetails) => void
    downloadProgressMap?: Record<string, number>
    isSendingDownloadRequest?: boolean
}

export function ChapterItem(props: ChapterItemProps) {

    const {
        chapter,
        chapterBackups,
        handleDownloadChapter,
        downloadProgressMap,
        isSendingDownloadRequest,
        ...rest
    } = props

    const setSelectedChapter = useSetAtom(__manga_selectedChapterAtom)

    return (
        <>
            <Card
                key={chapter.id}
                className={cn(
                    "px-3 py-1.5 flex w-full gap-2 items-center",
                    "hover:bg-[--subtle]",
                )}
            >
                <p>{chapter.title}</p>
                <div className="flex flex-1"></div>
                <IconButton
                    intent="gray-basic"
                    size="sm"
                    onClick={() => setSelectedChapter(chapter)}
                    icon={<GiOpenBook />}
                />
                {/*SHELVED*/}
                {/*{!chapterBackups?.chapterIds[chapter.id] && <IconButton*/}
                {/*    intent="gray-basic"*/}
                {/*    size="sm"*/}
                {/*    loading={downloadProgressMap?.[chapter.id] !== undefined}*/}
                {/*    disabled={isSendingDownloadRequest}*/}
                {/*    onClick={() => handleDownloadChapter?.(chapter)}*/}
                {/*    icon={<FaDownload />}*/}
                {/*/>}*/}
            </Card>
        </>
    )
}
