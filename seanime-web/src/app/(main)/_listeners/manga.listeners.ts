import { useWebsocketMessageListener } from "@/app/(main)/_hooks/websocket.hooks"
import { WSEvents } from "@/lib/server/endpoints"
import { useQueryClient } from "@tanstack/react-query"

/**
 * @description
 * Listens to DOWNLOADED_CHAPTER events and re-fetches queries associated with media ID
 */
export function useMangaListener() {

    const qc = useQueryClient()

    useWebsocketMessageListener<number>({
        type: WSEvents.DOWNLOADED_CHAPTER,
        onMessage: mediaId => {
            (async () => {
                await qc.refetchQueries({ queryKey: ["get-manga-download-data"] })
                await qc.refetchQueries({ queryKey: ["get-manga-downloads"] })
            })()
        },
    })

    useWebsocketMessageListener({
        type: WSEvents.CHAPTER_DOWNLOAD_QUEUE_UPDATED,
        onMessage: data => {
            (async () => {
                await qc.refetchQueries({ queryKey: ["get-manga-download-data"] })
                await qc.refetchQueries({ queryKey: ["get-manga-chapter-download-queue"] })
                await qc.refetchQueries({ queryKey: ["get-manga-downloads"] })
            })()
        },
    })

}
