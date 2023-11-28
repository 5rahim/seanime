import { useInfiniteQuery, useQuery } from "@tanstack/react-query"

import { searchAnilistMediaList } from "@/lib/anilist/queries/search-media"

export function useDiscoverTrendingAnime() {

    return useInfiniteQuery({
        queryKey: ["discover-trending-anime"],
        initialPageParam: 1,
        queryFn: async ({ pageParam }) => {
            return searchAnilistMediaList({
                page: pageParam,
                perPage: 20,
                sort: ["TRENDING_DESC"],
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