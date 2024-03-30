import { useWebsocketMessageListener } from "@/atoms/websocket"
import { AnimeCollectionQuery, MangaCollectionQuery } from "@/lib/anilist/gql/graphql"
import { SeaEndpoints, WSEvents } from "@/lib/server/endpoints"
import { useSeaQuery } from "@/lib/server/query"
import { useQueryClient } from "@tanstack/react-query"

/**
 * @description
 * Listens to REFRESHED_ANILIST_COLLECTION events and re-fetches queries associated with AniList collection.
 */
export function useAnilistCollectionListener() {

    const qc = useQueryClient()

    useWebsocketMessageListener<AnimeCollectionQuery>({
        type: WSEvents.REFRESHED_ANILIST_COLLECTION,
        onMessage: data => {
            (async () => {
                await qc.refetchQueries({ queryKey: ["get-library-collection"] })
                await qc.refetchQueries({ queryKey: ["get-anilist-collection"] })
                await qc.refetchQueries({ queryKey: ["get-missing-episodes"] })
            })()
        },
    })

    useWebsocketMessageListener<MangaCollectionQuery>({
        type: WSEvents.REFRESHED_ANILIST_MANGA_COLLECTION,
        onMessage: data => {
            (async () => {
                await qc.refetchQueries({ queryKey: ["get-manga-collection"] })
            })()
        },
    })

}

/**
 * @description
 * Fetches the (cached) AniList collection
 */
export function useAnilistCollection() {

    const { data, isLoading } = useSeaQuery<AnimeCollectionQuery>({
        endpoint: SeaEndpoints.ANILIST_COLLECTION,
        queryKey: ["get-anilist-collection"],
    })

    return {
        anilistLists: data?.MediaListCollection?.lists ?? [],
        isLoading,
    }

}
