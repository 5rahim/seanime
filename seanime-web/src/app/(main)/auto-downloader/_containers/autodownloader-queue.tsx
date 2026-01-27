import { Models_AutoDownloaderItem } from "@/api/generated/types"
import { useDeleteAutoDownloaderItem } from "@/api/hooks/auto_downloader.hooks"
import { useTorrentClientAddMagnetFromRule } from "@/api/hooks/torrent_client.hooks"
import { useMediaPreviewModal } from "@/app/(main)/_features/media/_containers/media-preview-modal"
import { useAnilistUserAnime } from "@/app/(main)/_hooks/anilist-collection-loader"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { SeaLink } from "@/components/shared/sea-link"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { openTab } from "@/lib/helpers/browser"
import { formatDistanceToNowSafe } from "@/lib/helpers/date"
import Image from "next/image"
import React from "react"
import { BiDownload, BiTrash } from "react-icons/bi"

type AutoDownloaderQueueProps = {
    children?: React.ReactNode
    items: Models_AutoDownloaderItem[] | undefined
    isLoading: boolean
}

export function AutodownloaderQueue(props: AutoDownloaderQueueProps) {

    const {
        children,
        items: data,
        isLoading,
        ...rest
    } = props

    const serverStatus = useServerStatus()

    const userMedia = useAnilistUserAnime()

    const { mutate: deleteItem, isPending } = useDeleteAutoDownloaderItem()

    const { mutate: addMagnet, isPending: isAdding } = useTorrentClientAddMagnetFromRule()

    const { setPreviewModalMediaId } = useMediaPreviewModal()

    if (isLoading) return <LoadingSpinner />

    return (
        <Card className="p-4 space-y-2">
            <ul className="text-base text-[--muted]">
                <li>
                    The queue shows items waiting to be downloaded or scanned.
                </li>
                {/* <li>
                 Removing an item from the queue can cause it to be re-added if the rule is still active and the episode isn't downloaded and scanned.
                 </li> */}
            </ul>
            {!data?.length && (
                <p className="text-center text-[--muted]">
                    Queue is empty
                </p>
            )}
            {data?.toSorted((a, b) => (b.updatedAt ?? "").localeCompare(a.updatedAt ?? ""))?.map((item) => {
                const media = userMedia?.find(m => m.id === item.mediaId)
                return <div className="rounded-[--radius] p-3 bg-gray-900" key={item.id}>
                    <div className="flex items-center gap-4">
                        <div
                            onClick={() => setPreviewModalMediaId(item.mediaId, "anime")}
                            className="cursor-pointer size-10 rounded-full bg-gray-800 flex items-center justify-center relative overflow-hidden flex-none"
                        >
                            <Image
                                src={media?.coverImage?.medium ?? "/no-cover.png"}
                                alt="cover"
                                sizes="2rem"
                                fill
                                className="object-cover object-center"
                            />
                        </div>
                        <div>
                            <h3
                                className="text-sm font-medium tracking-wide cursor-pointer"
                                onClick={() => openTab(item.link)}
                            >{item.torrentName}</h3>
                            <p className="text-md text-gray-400 flex gap-2 items-center">
                                {item.downloaded && <span className="text-green-200">File downloaded</span>}
                                {!item.downloaded && !item.isDelayed && <span className="text-blue-300 italic">Manual action required</span>}
                                {item.isDelayed && <span className="text-indigo-300 italic">Delayed</span>}
                                {item.isDelayed && item.delayUntil &&
                                    <span>for {formatDistanceToNowSafe(item.delayUntil, { addSuffix: false })}.</span>}
                            </p>
                            {item.createdAt && <span className="text-[--muted] text-sm">Added {formatDistanceToNowSafe(item.createdAt)}</span>}
                            {item.downloaded && (
                                <p className="text-sm text-[--muted]">
                                    Not yet scanned
                                </p>
                            )}
                        </div>
                        <div className="flex-1"></div>
                        <div className="flex gap-2 items-center">
                            {!item.downloaded && (
                                <>
                                    {!serverStatus?.settings?.autoDownloader?.useDebrid ? (
                                        <Button
                                            leftIcon={<BiDownload />}
                                            size="sm"
                                            intent="primary-subtle"
                                            onClick={() => {
                                                addMagnet({
                                                    magnetUrl: item.magnet,
                                                    ruleId: item.ruleId,
                                                    queuedItemId: item.id,
                                                })
                                            }}
                                            loading={isAdding}
                                            disabled={isPending}
                                        >
                                            Download
                                        </Button>
                                    ) : (
                                        <SeaLink href="/debrid">
                                            <Button
                                                leftIcon={<BiDownload />}
                                                size="sm"
                                                intent="primary-subtle"
                                                disabled={isPending}
                                            >
                                                Download
                                            </Button>
                                        </SeaLink>
                                    )}
                                </>
                            )}
                            <Button
                                leftIcon={<BiTrash />}
                                size="sm"
                                intent="alert-basic"
                                onClick={() => {
                                    deleteItem({ id: item.id })
                                }}
                                disabled={isPending || isAdding}
                                loading={isPending}
                            >
                                Remove
                            </Button>
                        </div>
                    </div>
                </div>
            })}
        </Card>
    )
}
