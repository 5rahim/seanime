import { ListMediaQuery } from "@/lib/anilist/gql/graphql"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaQuery } from "@/lib/server/query"
import { atom } from "jotai"
import { useAtomValue } from "jotai/react"

export const __discover_trendingGenresAtom = atom<string[]>([])

export function useDiscoverTrendingAnime() {
    const genres = useAtomValue(__discover_trendingGenresAtom)

    return useSeaQuery<ListMediaQuery>({
        queryKey: ["discover-trending-anime", genres],
        endpoint: SeaEndpoints.ANILIST_LIST_ANIME,
        method: "post",
        data: {
            page: 1,
            perPage: 20,
            sort: ["TRENDING_DESC"],
            genres: genres.length > 0 ? genres : undefined,
        },
    })

}

export function useDiscoverUpcomingAnime() {
    return useSeaQuery<ListMediaQuery>({
        queryKey: ["discover-upcoming-anime"],
        endpoint: SeaEndpoints.ANILIST_LIST_ANIME,
        method: "post",
        data: {
            page: 1,
            perPage: 20,
            sort: ["TRENDING_DESC"],
            status: ["NOT_YET_RELEASED"],
        },
    })
}

export function useDiscoverPopularAnime() {
    return useSeaQuery<ListMediaQuery>({
        queryKey: ["discover-popular-anime"],
        endpoint: SeaEndpoints.ANILIST_LIST_ANIME,
        method: "post",
        data: {
            page: 1,
            perPage: 20,
            sort: ["POPULARITY_DESC"],
        },
    })
}

export function useDiscoverTrendingMovies() {
    return useSeaQuery<ListMediaQuery>({
        queryKey: ["discover-trending-movies"],
        endpoint: SeaEndpoints.ANILIST_LIST_ANIME,
        method: "post",
        data: {
            page: 1,
            perPage: 20,
            format: "MOVIE",
            sort: ["TRENDING_DESC"],
        },
    })
}
