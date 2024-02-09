import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaQuery } from "@/lib/server/queries/utils"
import { AutoDownloaderItem } from "@/lib/server/types"
import { atom } from "jotai"
import { useAtomValue, useSetAtom } from "jotai/react"
import { usePathname } from "next/navigation"
import { useEffect } from "react"

export const autoDownloaderItemsAtom = atom<AutoDownloaderItem[]>([])

const countAtom = atom(get => get(autoDownloaderItemsAtom).length)

export function useAutoDownloaderQueueCount() {
    return useAtomValue(countAtom)
}

/**
 * @description
 * - When the user is not on the main page, send a request to get auto downloader queue items
 */
export function useListenToAutoDownloaderItems() {
    const pathname = usePathname()
    const setter = useSetAtom(autoDownloaderItemsAtom)

    const { data } = useSeaQuery<AutoDownloaderItem[]>({
        queryKey: ["auto-downloader-items"],
        endpoint: SeaEndpoints.AUTO_DOWNLOADER_ITEMS,
        enabled: pathname !== "/auto-downloader",
    })

    useEffect(() => {
        setter(data ?? [])
    }, [data])

    return null
}
