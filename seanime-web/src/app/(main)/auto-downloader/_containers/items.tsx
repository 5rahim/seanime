import { Button } from "@/components/ui/button"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation, useSeaQuery } from "@/lib/server/queries/utils"
import { AutoDownloaderItem } from "@/lib/server/types"
import { formatDateAndTimeShort } from "@/lib/server/utils"
import { BiDownload } from "@react-icons/all-files/bi/BiDownload"
import { BiTrash } from "@react-icons/all-files/bi/BiTrash"
import { useQueryClient } from "@tanstack/react-query"
import Link from "next/link"
import React from "react"
import toast from "react-hot-toast"

type AutoDownloaderItemsProps = {
    children?: React.ReactNode
}

export function AutoDownloaderItems(props: AutoDownloaderItemsProps) {

    const {
        children,
        ...rest
    } = props

    const qc = useQueryClient()

    const { data, isLoading } = useSeaQuery<AutoDownloaderItem[] | null>({
        queryKey: ["auto-downloader-items"],
        endpoint: SeaEndpoints.AUTO_DOWNLOADER_ITEMS,
    })

    const { mutate: deleteItem, isPending } = useSeaMutation<void, { id: number }>({
        mutationKey: ["delete-auto-downloader-items"],
        method: "delete",
        endpoint: SeaEndpoints.AUTO_DOWNLOADER_ITEM,
        onSuccess: async () => {
            await qc.refetchQueries({ queryKey: ["auto-downloader-items"] })
            toast.success("Item deleted")
        },
    })

    if (isLoading) return <LoadingSpinner />

    return (
        <div className="space-y-4">
            {!data?.length && (
                <p className="text-center text-[--muted]">
                    Queue is empty.
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
                        </div>
                        <div className="flex gap-2 items-center">
                            {!item.downloaded && (
                                <Link href={item.magnet} target="_blank">
                                    <Button
                                        leftIcon={<BiDownload />}
                                        size="sm"
                                        intent="primary-subtle"
                                    >
                                        Download
                                    </Button>
                                </Link>
                            )}
                            <Button
                                leftIcon={<BiTrash />}
                                size="sm"
                                intent="alert"
                                onClick={() => {
                                    deleteItem({ id: item.id })
                                }}
                                isDisabled={isPending}
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
