"use client"
import { Manga_Collection } from "@/api/generated/types"
import { useGetMangaCollection } from "@/api/hooks/manga.hooks"
import { useGetMangaDownloadsList } from "@/api/hooks/manga_download.hooks"
import { MediaEntryCard } from "@/app/(main)/_features/media/_components/media-entry-card"

import { useHandleMangaChapterDownloadQueue } from "@/app/(main)/manga/_lib/handle-manga-downloads"
import { LuffyError } from "@/components/shared/luffy-error"
import { SeaLink } from "@/components/shared/sea-link"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import { ProgressBar } from "@/components/ui/progress-bar"
import { ScrollArea } from "@/components/ui/scroll-area"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import React from "react"
import { MdClear } from "react-icons/md"
import { PiWarningOctagonDuotone } from "react-icons/pi"
import { TbWorldDownload } from "react-icons/tb"

export const __manga_chapterDownloadsDrawerIsOpenAtom = atom(false)

type ChapterDownloadQueueDrawerProps = {}

export function ChapterDownloadsDrawer(props: ChapterDownloadQueueDrawerProps) {

    const {} = props

    const [isOpen, setIsOpen] = useAtom(__manga_chapterDownloadsDrawerIsOpenAtom)

    const { data: mangaCollection } = useGetMangaCollection()

    return (
        <>
            <Modal
                open={isOpen}
                onOpenChange={setIsOpen}
                contentClass="max-w-5xl"
                title="Downloaded chapters"
                data-chapter-downloads-modal
            >

                <div className="py-4 space-y-8" data-chapter-downloads-modal-content>
                    <ChapterDownloadQueue mangaCollection={mangaCollection} />

                    <ChapterDownloadList />
                </div>

            </Modal>
        </>
    )
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type ChapterDownloadQueueProps = {
    mangaCollection: Manga_Collection | undefined
}

export function ChapterDownloadQueue(props: ChapterDownloadQueueProps) {

    const {
        mangaCollection,
        ...rest
    } = props

    const {
        downloadQueue,
        downloadQueueLoading,
        downloadQueueError,
        startDownloadQueue,
        stopDownloadQueue,
        isStartingDownloadQueue,
        isStoppingDownloadQueue,
        resetErroredChapters,
        isResettingErroredChapters,
        clearDownloadQueue,
        isClearingDownloadQueue,
    } = useHandleMangaChapterDownloadQueue()

    const isMutating = isStartingDownloadQueue || isStoppingDownloadQueue || isResettingErroredChapters || isClearingDownloadQueue

    return (
        <>
            <div className="space-y-4" data-chapter-download-queue-container>

                <div className="flex w-full items-center" data-chapter-download-queue-header>
                    <h3>Queue</h3>
                    <div className="flex flex-1" data-chapter-download-queue-header-spacer></div>
                    {(!downloadQueueLoading && !downloadQueueError) &&
                        <div className="flex gap-2 items-center" data-chapter-download-queue-header-actions>

                        {!!downloadQueue?.find(n => n.status === "errored") && <Button
                            intent="warning-outline"
                            size="sm"
                            disabled={isMutating}
                            onClick={() => resetErroredChapters()}
                            loading={isResettingErroredChapters}
                        >
                            Reset errored chapters
                        </Button>}

                        {!!downloadQueue?.find(n => n.status === "downloading") ? (
                            <>
                                <Button
                                    intent="alert-subtle"
                                    size="sm"
                                    onClick={() => stopDownloadQueue()}
                                    loading={isStoppingDownloadQueue}
                                >
                                    Stop
                                </Button>
                            </>
                        ) : (
                            <>
                                {!!downloadQueue?.length && <Button
                                    intent="alert-subtle"
                                    size="sm"
                                    disabled={isMutating}
                                    onClick={() => clearDownloadQueue()}
                                    leftIcon={<MdClear className="text-xl" />}
                                    loading={isClearingDownloadQueue}
                                >
                                    Clear all
                                </Button>}

                                {(!!downloadQueue?.length && !!downloadQueue?.find(n => n.status === "not_started")) && <Button
                                    intent="success"
                                    size="sm"
                                    disabled={isMutating}
                                    onClick={() => startDownloadQueue()}
                                    leftIcon={<TbWorldDownload className="text-xl" />}
                                    loading={isStartingDownloadQueue}
                                >
                                    Start
                                </Button>}
                            </>
                        )}
                    </div>}
                </div>

                <Card className="p-4 space-y-2" data-chapter-download-queue-card>

                    {downloadQueueLoading
                        ? <LoadingSpinner />
                        : (downloadQueueError ? <LuffyError title="Oops!">
                            <p>Could not fetch the download queue</p>
                        </LuffyError> : null)}

                    {!!downloadQueue?.length ? (
                        <ScrollArea className="h-[14rem]" data-chapter-download-queue-scroll-area>
                            <div className="space-y-2" data-chapter-download-queue-scroll-area-content>
                                {downloadQueue.map(item => {

                                    const media = mangaCollection?.lists?.flatMap(n => n.entries)?.find(n => n?.media?.id === item.mediaId)?.media

                                    return (
                                        <Card
                                            key={item.mediaId + item.provider + item.chapterId} className={cn(
                                            "px-3 py-2 bg-gray-900 space-y-1.5",
                                            item.status === "errored" && "border-[--orange]",
                                        )}
                                        >
                                            <div className="flex items-center gap-2">
                                                {!!media && <SeaLink
                                                    className="font-semibold max-w-[180px] text-ellipsis truncate underline"
                                                    href={`/manga/entry?id=${media.id}`}
                                                >{media.title?.userPreferred}</SeaLink>}
                                                <p>Chapter {item.chapterNumber} <span className="text-[--muted] italic">(id: {item.chapterId})</span>
                                                </p>
                                                {item.status === "errored" && (
                                                    <div className="flex gap-1 items-center text-[--orange]">
                                                        <PiWarningOctagonDuotone className="text-2xl text-[--orange]" />
                                                        <p>
                                                            Errored
                                                        </p>
                                                    </div>
                                                )}
                                            </div>
                                            {item.status === "downloading" && (
                                                <ProgressBar size="sm" isIndeterminate />
                                            )}
                                        </Card>
                                    )
                                })}
                            </div>
                        </ScrollArea>
                    ) : ((!downloadQueueLoading && !downloadQueueError) && (
                        <p className="text-center text-[--muted] italic" data-chapter-download-queue-empty-state>
                            Nothing in the queue
                        </p>
                    ))}

                </Card>

            </div>
        </>
    )
}

/////////////////////////////////////

type ChapterDownloadListProps = {}

export function ChapterDownloadList(props: ChapterDownloadListProps) {

    const {} = props

    const { data, isLoading, isError } = useGetMangaDownloadsList()

    return (
        <>
            <div className="space-y-4" data-chapter-download-list-container>

                <div className="flex w-full items-center" data-chapter-download-list-header>
                    <h3>Downloaded</h3>
                    <div className="flex flex-1" data-chapter-download-list-header-spacer></div>
                </div>

                <div className="py-4 space-y-2" data-chapter-download-list-content>

                    {isLoading
                        ? <LoadingSpinner />
                        : (isError ? <LuffyError title="Oops!">
                            <p>Could not fetch the download queue</p>
                        </LuffyError> : null)}

                    {!!data?.length ? (
                        <>
                            {data?.filter(n => !n.media)
                                .sort((a, b) => a.mediaId - b.mediaId)
                                .sort((a, b) => Object.values(b.downloadData).flatMap(n => n).length - Object.values(a.downloadData)
                                    .flatMap(n => n).length)
                                .map(item => {
                                    return (
                                        <Card
                                            key={item.mediaId} className={cn(
                                            "px-3 py-2 bg-gray-900 space-y-1",
                                        )}
                                        >
                                            <SeaLink
                                                className="font-semibold underline"
                                                href={`/manga/entry?id=${item.mediaId}`}
                                            >Media {item.mediaId}</SeaLink>

                                            <div className="flex items-center gap-2">
                                                <p>{Object.values(item.downloadData)
                                                    .flatMap(n => n).length} chapters</p> - <em className="text-[--muted]">Not in your AniList
                                                                                                                           collection</em>
                                            </div>
                                        </Card>
                                    )
                                })}

                            <div
                                data-chapter-download-list-media-grid
                                className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-4 xl:grid-cols-4 2xl:grid-cols-4 gap-4"
                            >
                                {data?.filter(n => !!n.media)
                                    .sort((a, b) => a.mediaId - b.mediaId)
                                    .sort((a, b) => Object.values(b.downloadData).flatMap(n => n).length - Object.values(a.downloadData)
                                        .flatMap(n => n).length)
                                    .map(item => {
                                        const nb = Object.values(item.downloadData).flatMap(n => n).length
                                        return <div key={item.media?.id!} className="col-span-1">
                                            <MediaEntryCard
                                                media={item.media!}
                                                type="manga"
                                                hideUnseenCountBadge
                                                overlay={<Badge
                                                    className="font-semibold text-white bg-gray-950 !bg-opacity-100 rounded-[--radius-md] text-base rounded-bl-none rounded-tr-none"
                                                    intent="gray"
                                                    size="lg"
                                                >{nb} chapter{nb === 1 ? "" : "s"}</Badge>}
                                            />
                                        </div>
                                    })}
                            </div>
                        </>
                    ) : ((!isLoading && !isError) && (
                        <p className="text-center text-[--muted] italic" data-chapter-download-list-empty-state>
                            No chapters downloaded
                        </p>
                    ))}

                </div>

            </div>
        </>
    )
}
