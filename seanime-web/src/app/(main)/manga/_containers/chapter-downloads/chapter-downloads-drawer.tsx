"use client"
import { MediaEntryCard } from "@/app/(main)/_components/features/media/media-entry-card"
import { useMangaChapterDownloadQueue, useMangaChapterDownloads } from "@/app/(main)/manga/_lib/manga.hooks"
import { MangaCollection } from "@/app/(main)/manga/_lib/manga.types"
import { LuffyError } from "@/components/shared/luffy-error"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { cn } from "@/components/ui/core/styling"
import { Drawer } from "@/components/ui/drawer"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { ProgressBar } from "@/components/ui/progress-bar"
import { ScrollArea } from "@/components/ui/scroll-area"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaQuery } from "@/lib/server/query"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import Link from "next/link"
import React from "react"
import { MdClear } from "react-icons/md"
import { PiWarningOctagonDuotone } from "react-icons/pi"
import { TbWorldDownload } from "react-icons/tb"

export const __manga__chapterDownloadsDrawerIsOpenAtom = atom(false)

type ChapterDownloadQueueDrawerProps = {}

export function ChapterDownloadsDrawer(props: ChapterDownloadQueueDrawerProps) {

    const {
        ...rest
    } = props

    const [isOpen, setIsOpen] = useAtom(__manga__chapterDownloadsDrawerIsOpenAtom)

    const { data: mangaCollection } = useSeaQuery<MangaCollection>({
        endpoint: SeaEndpoints.MANGA_COLLECTION,
        queryKey: ["get-manga-collection"],
    })

    return (
        <>
            <Drawer
                open={isOpen}
                onOpenChange={setIsOpen}
                size="xl"
                title="Downloaded chapters"
            >

                <div className="py-4 space-y-8">
                    <ChapterDownloadQueue mangaCollection={mangaCollection} />

                    <ChapterDownloadList />
                </div>

            </Drawer>
        </>
    )
}

/////////////////////////////////////

type ChapterDownloadQueueProps = {
    mangaCollection: MangaCollection | undefined
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
    } = useMangaChapterDownloadQueue()

    const isMutating = isStartingDownloadQueue || isStoppingDownloadQueue || isResettingErroredChapters || isClearingDownloadQueue

    return (
        <>
            <div className="space-y-4">

                <div className="flex w-full items-center">
                    <h3>Queue</h3>
                    <div className="flex flex-1"></div>
                    {(!downloadQueueLoading && !downloadQueueError) && <div className="flex gap-2 items-center">

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

                <Card className="p-4 space-y-2">

                    {downloadQueueLoading
                        ? <LoadingSpinner />
                        : (downloadQueueError ? <LuffyError title="Oops!">
                            <p>Could not fetch the download queue</p>
                        </LuffyError> : null)}

                    {!!downloadQueue?.length ? (
                        <ScrollArea className="h-[14rem]">
                            <div className="space-y-2">
                                {downloadQueue.map(item => {

                                    const media = mangaCollection?.lists?.flatMap(n => n.entries)?.find(n => n.media?.id === item.mediaId)?.media

                                    return (
                                        <Card
                                            key={item.mediaId + item.provider + item.chapterId} className={cn(
                                            "px-3 py-2 bg-gray-900 space-y-1.5",
                                            item.status === "errored" && "border-[--orange]",
                                        )}
                                        >
                                            <div className="flex items-center gap-2">
                                                {!!media && <Link
                                                    className="font-semibold max-w-[180px] text-ellipsis truncate underline"
                                                    href={`/manga/entry?id=${media.id}`}
                                                >{media.title?.userPreferred}</Link>}
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
                        <p className="text-center text-[--muted] italic">
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

    const {
        ...rest
    } = props

    const {
        data,
        isLoading,
        isError,
    } = useMangaChapterDownloads()

    return (
        <>
            <div className="space-y-4">

                <div className="flex w-full items-center">
                    <h3>Downloaded</h3>
                    <div className="flex flex-1"></div>
                </div>

                <div className="py-4 space-y-2">

                    {isLoading
                        ? <LoadingSpinner />
                        : (isError ? <LuffyError title="Oops!">
                            <p>Could not fetch the download queue</p>
                        </LuffyError> : null)}

                    {!!data?.length ? (
                        <>
                            {data?.filter(n => !n.media)
                                .sort((a, b) => Object.values(b.downloadData).flatMap(n => n).length - Object.values(a.downloadData)
                                    .flatMap(n => n).length)
                                .map(item => {
                                    return (
                                        <Card
                                            key={item.mediaId} className={cn(
                                            "px-3 py-2 bg-gray-900 space-y-1",
                                        )}
                                        >
                                            <Link
                                                className="font-semibold underline"
                                                href={`/manga/entry?id=${item.mediaId}`}
                                            >Media {item.mediaId}</Link>

                                            <div className="flex items-center gap-2">
                                                <p>{Object.values(item.downloadData)
                                                    .flatMap(n => n).length} chapters</p> - <em className="text-[--muted]">Not in your AniList
                                                                                                                           collection</em>
                                            </div>
                                        </Card>
                                    )
                                })}

                            <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-4 xl:grid-cols-4 2xl:grid-cols-4 gap-4">
                                {data?.filter(n => !!n.media)
                                    .sort((a, b) => Object.values(b.downloadData).flatMap(n => n).length - Object.values(a.downloadData)
                                        .flatMap(n => n).length)
                                    .map(item => {
                                        const nb = Object.values(item.downloadData).flatMap(n => n).length
                                        return <div key={item.media?.id!} className="col-span-1">
                                            <MediaEntryCard
                                                media={item.media!}
                                                showLibraryBadge
                                                showTrailer
                                                isManga
                                                overlay={<Badge
                                                    className="font-semibold text-white bg-gray-950 !bg-opacity-100 rounded-md text-base rounded-bl-none rounded-tr-none"
                                                    intent="gray"
                                                    size="lg"
                                                >{nb} chapter{nb === 1 ? "" : "s"}</Badge>}
                                            />
                                        </div>
                                    })}
                            </div>
                        </>
                    ) : ((!isLoading && !isError) && (
                        <p className="text-center text-[--muted] italic">
                            No chapters downloaded
                        </p>
                    ))}

                </div>

            </div>
        </>
    )
}
