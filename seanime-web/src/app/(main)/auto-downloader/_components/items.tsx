import { Button } from "@/components/ui/button"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation } from "@/lib/server/query"
import { AutoDownloaderItem } from "@/lib/server/types"
import { formatDateAndTimeShort } from "@/lib/server/utils"
import { BiDownload } from "@react-icons/all-files/bi/BiDownload"
import { BiTrash } from "@react-icons/all-files/bi/BiTrash"
import { useQueryClient } from "@tanstack/react-query"
import React from "react"
import toast from "react-hot-toast"

type AutoDownloaderItemsProps = {
    children?: React.ReactNode
    items: AutoDownloaderItem[] | undefined
    isLoading: boolean
}

export function AutoDownloaderItems(props: AutoDownloaderItemsProps) {

    const {
        children,
        items: data,
        isLoading,
        ...rest
    } = props

    const qc = useQueryClient()

    const { mutate: deleteItem, isPending } = useSeaMutation<void, { id: number }>({
        mutationKey: ["delete-auto-downloader-items"],
        method: "delete",
        endpoint: SeaEndpoints.AUTO_DOWNLOADER_ITEM,
        onSuccess: async () => {
            await qc.refetchQueries({ queryKey: ["auto-downloader-items"] })
            toast.success("Item deleted")
        },
    })

    const { mutate: addMagnet, isPending: isAdding } = useSeaMutation<void, { magnetUrl: string, ruleId: number, queuedItemId: number }>({
        mutationKey: ["torrent-client-add-rule-magnet"],
        method: "post",
        endpoint: SeaEndpoints.TORRENT_CLIENT_RULE_MAGNET,
        onSuccess: async () => {
            await qc.refetchQueries({ queryKey: ["auto-downloader-items"] })
            toast.success("Magnet added")
        },
    })

    if (isLoading) return <LoadingSpinner />

    return (
        <div className="space-y-4">
            <ul className="text-base text-[--muted]">
                <li>
                    The queue shows items waiting to be downloaded or scanned.
                </li>
            </ul>
            {!data?.length && (
                <p className="text-center text-[--muted]">
                    Queue is empty
                </p>
            )}
            {data?.map((item) => (
                <div className="rounded-[--radius] p-3 bg-[--background-color]" key={item.id}>
                    <div className="flex items-center justify-between">
                        <div>
                            <h3 className="text-base font-medium tracking-wide">{item.torrentName}</h3>
                            <p className="text-base text-gray-400 flex gap-2 items-center">
                                {item.downloaded && <span className="text-green-200">File downloaded </span>}
                                {!item.downloaded && <span className="text-brand-300 italic">Queued </span>}
                                {formatDateAndTimeShort(item.createdAt)}
                            </p>
                            {item.downloaded && (
                                <p className="text-sm text-[--muted]">
                                    Not yet scanned
                                </p>
                            )}
                        </div>
                        <div className="flex gap-2 items-center">
                            {!item.downloaded && (
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
                                    isLoading={isAdding}
                                    isDisabled={isPending}
                                >
                                    Download
                                </Button>
                            )}
                            <Button
                                leftIcon={<BiTrash />}
                                size="sm"
                                intent="alert"
                                onClick={() => {
                                    deleteItem({ id: item.id })
                                }}
                                isDisabled={isPending || isAdding}
                                isLoading={isPending}
                            >
                                Delete
                            </Button>
                        </div>
                    </div>
                </div>
            ))}
        </div>
    )
}
