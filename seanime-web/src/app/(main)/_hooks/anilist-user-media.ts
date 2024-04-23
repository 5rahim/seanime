import { AnimeCollectionQuery, BaseMediaFragment } from "@/lib/anilist/gql/graphql"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaQuery } from "@/lib/server/query"
import { atom } from "jotai/index"
import { useSetAtom } from "jotai/react"
import { useEffect } from "react"

export const anilistUserMediaAtom = atom<BaseMediaFragment[] | undefined>(undefined)

/**
 * @description
 * - Top level function
 * - Listens to "get-anilist-collection" query and updates `userMediaAtom`
 */
export function useAnilistUserMediaLoader() {
    const setter = useSetAtom(anilistUserMediaAtom)

    const { data, isLoading } = useSeaQuery<AnimeCollectionQuery>({
        endpoint: SeaEndpoints.ANILIST_COLLECTION,
        queryKey: ["get-anilist-collection"],
    })

    // Store the received data in `userMediaAtom`
    useEffect(() => {
        if (!!data) {
            const allMedia = data.MediaListCollection?.lists?.flatMap(n => n?.entries)?.filter(Boolean)?.map(n => n.media)?.filter(Boolean) ?? []
            setter(allMedia)
        }
    }, [data])

    return null
}
