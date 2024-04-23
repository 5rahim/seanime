import { useWebsocketMessageListener } from "@/app/(main)/_atoms/websocket"
import { AutoDownloaderItem } from "@/app/(main)/auto-downloader/_lib/autodownloader.types"
import { SeaEndpoints, WSEvents } from "@/lib/server/endpoints"
import { useSeaQuery } from "@/lib/server/query"
import { useQueryClient } from "@tanstack/react-query"
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
export function useAutoDownloaderItemListener() {
    const pathname = usePathname()
    const setter = useSetAtom(autoDownloaderItemsAtom)
    const qc = useQueryClient()

    const { data, refetch } = useSeaQuery<AutoDownloaderItem[]>({
        queryKey: ["auto-downloader-items"],
        endpoint: SeaEndpoints.AUTO_DOWNLOADER_ITEMS,
        enabled: pathname !== "/auto-downloader",
    })

    useWebsocketMessageListener<string>({
        type: WSEvents.AUTO_DOWNLOADER_ITEM_ADDED,
        onMessage: data => {
            qc.refetchQueries({ queryKey: ["auto-downloader-items"] })
        },
    })

    useEffect(() => {
        setter(data ?? [])
    }, [data])

    return null
}
