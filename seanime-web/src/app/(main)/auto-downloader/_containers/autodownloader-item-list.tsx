import { Models_AutoDownloaderItem } from "@/api/generated/types"
import { useDeleteAutoDownloaderItem } from "@/api/hooks/auto_downloader.hooks"
import { useTorrentClientAddMagnetFromRule } from "@/api/hooks/torrent_client.hooks"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { SeaLink } from "@/components/shared/sea-link"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { formatDateAndTimeShort } from "@/lib/server/utils"
import React from "react"
import { BiDownload, BiTrash } from "react-icons/bi"

type AutoDownloaderItemListProps = {
    children?: React.ReactNode
    items: Models_AutoDownloaderItem[] | undefined
    isLoading: boolean
}

export function AutoDownloaderItemList(props: AutoDownloaderItemListProps) {

    const {
        children,
        items: data,
        isLoading,
        ...rest
    } = props

    const serverStatus = useServerStatus()

    const { mutate: deleteItem, isPending } = useDeleteAutoDownloaderItem()

    const { mutate: addMagnet, isPending: isAdding } = useTorrentClientAddMagnetFromRule()

    if (isLoading) return <LoadingSpinner />

    return (
        <Card className="p-4 space-y-4">
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
            {data?.map((item) => (
                <div className="rounded-[--radius] p-3 bg-gray-900" key={item.id}>
                    <div className="flex items-center justify-between">
                        <div>
                            <h3 className="text-base font-medium tracking-wide">{item.torrentName}</h3>
                            <p className="text-base text-gray-400 flex gap-2 items-center">
                                {item.downloaded && <span className="text-green-200">File downloaded </span>}
                                {!item.downloaded && <span className="text-brand-300 italic">Queued </span>}
                                {item.createdAt && formatDateAndTimeShort(item.createdAt)}
                            </p>
                            {item.downloaded && (
                                <p className="text-sm text-[--muted]">
                                    Not yet scanned
                                </p>
                            )}
                        </div>
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
                                intent="alert"
                                onClick={() => {
                                    deleteItem({ id: item.id })
                                }}
                                disabled={isPending || isAdding}
                                loading={isPending}
                            >
                                Delete
                            </Button>
                        </div>
                    </div>
                </div>
            ))}
        </Card>
    )
}
