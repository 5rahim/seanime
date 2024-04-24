import { useGetAutoDownloaderItems } from "@/api/hooks/auto_downloader.hooks"
import { autoDownloaderItemCountAtom, autoDownloaderItemsAtom } from "@/app/(main)/_atoms/autodownloader.atoms"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/websocket.hooks"
import { WSEvents } from "@/lib/server/endpoints"
import { useQueryClient } from "@tanstack/react-query"
import { useAtomValue, useSetAtom } from "jotai/react"
import { usePathname } from "next/navigation"
import { useEffect } from "react"

export function useAutoDownloaderQueueCount() {
    return useAtomValue(autoDownloaderItemCountAtom)
}

/**
 * @description
 * - When the user is not on the main page, send a request to get auto downloader queue items
 */
export function useAutoDownloaderItemListener() {
    const pathname = usePathname()
    const setter = useSetAtom(autoDownloaderItemsAtom)
    const qc = useQueryClient()

    const { data } = useGetAutoDownloaderItems(pathname !== "/auto-downloader")

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
