import { searchAnilistMediaList } from "@/lib/anilist/queries/search-media"
import { useInfiniteQuery, useQuery } from "@tanstack/react-query"
import { atom } from "jotai"
import { useAtomValue } from "jotai/react"

export const __discover_trendingGenresAtom = atom<string[]>([])

export function useDiscoverTrendingAnime() {
    const genres = useAtomValue(__discover_trendingGenresAtom)

    return useInfiniteQuery({
        queryKey: ["discover-trending-anime", genres],
        initialPageParam: 1,
        queryFn: async ({ pageParam }) => {
            return searchAnilistMediaList({
                page: pageParam,
                perPage: 20,
                sort: ["TRENDING_DESC"],
                genres: genres.length > 0 ? genres : undefined,
            })
        },
        getNextPageParam: (lastPage, pages) => {
            const curr = lastPage?.Page?.pageInfo?.currentPage
            const hasNext = lastPage?.Page?.pageInfo?.hasNextPage
            return (!!curr && hasNext && curr < 4) ? pages.length + 1 : undefined
        },
    })

}

export function useDiscoverUpcomingAnime() {
    return useQuery({
        queryKey: ["discover-upcoming-anime"],
        queryFn: () => {
            return searchAnilistMediaList({
                page: 1,
                perPage: 20,
                sort: ["TRENDING_DESC"],
                status: ["NOT_YET_RELEASED"],
            })
        },
    })
}

export function useDiscoverPopularAnime() {
    return useQuery({
        queryKey: ["discover-popular-anime"],
        queryFn: () => {
            return searchAnilistMediaList({
                page: 1,
                perPage: 20,
                sort: ["POPULARITY_DESC"],
            })
        },
    })
}

export function useDiscoverTrendingMovies() {
    return useQuery({
        queryKey: ["discover-trending-movies"],
        queryFn: () => {
            return searchAnilistMediaList({
                page: 1,
                perPage: 20,
                format: "MOVIE",
                sort: ["TRENDING_DESC"],
            })
        },
    })
}
