import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/websocket.hooks"
import { MangaCollectionQuery } from "@/lib/anilist/gql/graphql"
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

    useWebsocketMessageListener<MangaCollectionQuery>({
        type: WSEvents.REFRESHED_ANILIST_MANGA_COLLECTION,
        onMessage: data => {
            (async () => {
                await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.MANGA.GetAnilistMangaCollection.key] })
            })()
        },
    })

}

