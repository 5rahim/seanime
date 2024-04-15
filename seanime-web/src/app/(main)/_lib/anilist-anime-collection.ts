import { AnimeCollectionQuery } from "@/lib/anilist/gql/graphql"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaQuery } from "@/lib/server/query"

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
