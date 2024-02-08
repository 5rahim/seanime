import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaQuery } from "@/lib/server/queries/utils"
import { AutoDownloaderItem } from "@/lib/server/types"
import React from "react"

type AutoDownloaderItemsProps = {
    children?: React.ReactNode
}

export function AutoDownloaderItems(props: AutoDownloaderItemsProps) {

    const {
        children,
        ...rest
    } = props

    const { data, isLoading } = useSeaQuery<AutoDownloaderItem[] | null>({
        queryKey: ["auto-downloader-items"],
        endpoint: SeaEndpoints.AUTO_DOWNLOADER_ITEMS,
    })

    if (isLoading) return <LoadingSpinner />

    return (
        <>
            <pre>
                {JSON.stringify(data, null, 2)}
            </pre>
        </>
    )
}
