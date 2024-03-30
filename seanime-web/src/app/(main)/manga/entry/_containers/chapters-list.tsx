import { useMangaChapterContainer, useMangaEntryBackups } from "@/app/(main)/manga/_lib/queries"
import { MangaChapterDetails, MangaEntry, MangaEntryBackups } from "@/app/(main)/manga/_lib/types"
import { __manga_selectedChapterAtom, ChapterDrawer } from "@/app/(main)/manga/entry/_containers/chapter-drawer"
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion"
import { IconButton } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { MangaDetailsByIdQuery } from "@/lib/anilist/gql/graphql"
import { atomWithImmer } from "jotai-immer"
import { useAtom, useSetAtom } from "jotai/react"
import React from "react"
import { BiBookAlt } from "react-icons/bi"
import { FaBookOpenReader } from "react-icons/fa6"

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


    const { chapterContainer, chapterIdToNumbersMap, chapterContainerLoading } = useMangaChapterContainer(mediaId)

    const { chapterBackups, chapterBackupsLoading } = useMangaEntryBackups(mediaId)
    const [downloadProgressMap, setDownloadProgressMap] = useAtom(downloadProgressMapAtom)

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

    const retainUnreadChapters = React.useCallback((chapter: MangaChapterDetails) => {
        if (!entry.listData || !chapterIdToNumbersMap.has(chapter.id) || !entry.listData?.progress) return true

        const chapterNumber = chapterIdToNumbersMap.get(chapter.id)
        return chapterNumber && chapterNumber > entry.listData?.progress
    }, [chapterIdToNumbersMap, chapterContainer, entry])

    const handleDownloadChapter = React.useCallback((chapter: MangaChapterDetails) => {
        // shelved
        // downloadChapter(chapter)
    }, [])

    if (!chapterContainer || chapterContainerLoading) return <LoadingSpinner />

    return (
        <div
            className="space-y-2"
        >

            <Accordion
                type="single"
                className=""
                triggerClass="text-[--muted] dark:data-[state=open]:text-white px-0 dark:hover:bg-transparent hover:bg-transparent dark:hover:text-white hover:text-black"
                itemClass="border-b"
                contentClass="pb-8"
                collapsible
            >
                <AccordionItem value="all">
                    <AccordionTrigger>
                        <h3 className="flex gap-2 items-center"><BiBookAlt className="text-gray-300" /> All chapters</h3>
                    </AccordionTrigger>
                    <AccordionContent className="p-0 py-4 space-y-2">
                        {chapterContainer?.chapters?.toReversed()?.map((chapter) => (
                            <ChapterItem
                                chapter={chapter}
                                key={chapter.id}
                                chapterBackups={chapterBackups}
                                handleDownloadChapter={handleDownloadChapter}
                                downloadProgressMap={downloadProgressMap}
                                isSendingDownloadRequest={false}
                            />
                        ))}
                    </AccordionContent>
                </AccordionItem>
            </Accordion>


            <h3>Unread chapters</h3>
            {chapterContainer?.chapters?.filter(ch => retainUnreadChapters(ch)).map((chapter) => (
                <ChapterItem
                    chapter={chapter}
                    key={chapter.id}
                    chapterBackups={chapterBackups}
                    handleDownloadChapter={handleDownloadChapter}
                    downloadProgressMap={downloadProgressMap}
                    isSendingDownloadRequest={false}
                />
            ))}

            <ChapterDrawer
                entry={entry}
                chapterContainer={chapterContainer}
                chapterIdToNumbersMap={chapterIdToNumbersMap}
            />
        </div>
    )
}


type ChapterItemProps = {
    chapter: MangaChapterDetails
    chapterBackups: MangaEntryBackups | undefined
    handleDownloadChapter: (chapter: MangaChapterDetails) => void
    downloadProgressMap: Record<string, number>
    isSendingDownloadRequest: boolean
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
                    "p-3 flex w-full gap-2 items-center",
                    "hover:bg-[--subtle]",
                )}
            >
                <p>{chapter.title}</p>
                <div className="flex flex-1"></div>
                <IconButton
                    intent="gray-basic"
                    size="sm"
                    onClick={() => setSelectedChapter(chapter)}
                    icon={<FaBookOpenReader />}
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
