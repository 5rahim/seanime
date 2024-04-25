import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { WSEvents } from "@/lib/server/ws-events"
import { useQueryClient } from "@tanstack/react-query"

/**
 * @description
 * - Listens to REFRESHED_ANILIST_COLLECTION events and re-fetches queries associated with AniList collection.
 */
export function useAnilistCollectionListener() {

    const qc = useQueryClient()

    useWebsocketMessageListener({
        type: WSEvents.REFRESHED_ANILIST_COLLECTION,
        onMessage: data => {
            (async () => {
                await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_COLLECTION.GetLibraryCollection.key] })
                await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.ANILIST.GetAnilistCollection.key] })
                await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.ANIME_ENTRIES.GetMissingEpisodes.key] })
            })()
        },
    })

    useWebsocketMessageListener({
        type: WSEvents.REFRESHED_ANILIST_MANGA_COLLECTION,
        onMessage: data => {
            (async () => {
                await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA.GetAnilistMangaCollection.key] })
            })()
        },
    })

}

