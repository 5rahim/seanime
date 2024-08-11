import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { WSEvents } from "@/lib/server/ws-events"
import { useQueryClient } from "@tanstack/react-query"

/**
 * @description
 * - Re-fetches queries associated with extension data
 */
export function useExtensionListener() {

    const qc = useQueryClient()

    useWebsocketMessageListener<number>({
        type: WSEvents.EXTENSIONS_RELOADED,
        onMessage: () => {
            (async () => {
                await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.EXTENSIONS.ListAnimeTorrentProviderExtensions.key] })
                await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.EXTENSIONS.ListMangaProviderExtensions.key] })
                await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.EXTENSIONS.ListOnlinestreamProviderExtensions.key] })
                await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.EXTENSIONS.ListExtensionData.key] })
                await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.EXTENSIONS.GetAllExtensions.key] })
            })()
        },
    })

}
