import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/websocket.hooks"
import { WSEvents } from "@/lib/server/endpoints"
import { useQueryClient } from "@tanstack/react-query"

/**
 * @description
 * - Listens to DOWNLOADED_CHAPTER events and re-fetches queries associated with media ID
 */
export function useMangaListener() {

    const qc = useQueryClient()

    useWebsocketMessageListener<number>({
        type: WSEvents.DOWNLOADED_CHAPTER,
        onMessage: mediaId => {
            (async () => {
                await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA_DOWNLOAD.GetMangaDownloadData.key] })
                await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA_DOWNLOAD.GetMangaDownloadsList.key] })
            })()
        },
    })

    useWebsocketMessageListener({
        type: WSEvents.CHAPTER_DOWNLOAD_QUEUE_UPDATED,
        onMessage: data => {
            (async () => {
                await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA_DOWNLOAD.GetMangaDownloadData.key] })
                await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA_DOWNLOAD.GetMangaDownloadQueue.key] })
                await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA_DOWNLOAD.GetMangaDownloadsList.key] })
            })()
        },
    })

}
