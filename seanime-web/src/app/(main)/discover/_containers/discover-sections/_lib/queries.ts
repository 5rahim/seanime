import { ListMediaQuery, ListMediaQueryVariables, MediaSeason } from "@/lib/anilist/gql/graphql"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaQuery } from "@/lib/server/query"
import { useInView } from "framer-motion"
import { atom } from "jotai"
import { useAtomValue } from "jotai/react"
import React from "react"

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

export function useDiscoverPastSeasonAnime(ref: any) {
    const _isInView = useInView(ref)
    const [isInView, setIsInView] = React.useState(false)
    React.useEffect(() => {
        if (isInView) return
        setIsInView(_isInView)
    }, [_isInView])
    const currentMonth = new Date().getMonth() + 1
    const currentYear = new Date().getFullYear()
    let season: MediaSeason = "SUMMER"
    switch (currentMonth) {
        case 1:
        case 2:
        case 3:
            season = "WINTER"
            break
        case 4:
        case 5:
        case 6:
            season = "SPRING"
            break
        case 7:
        case 8:
        case 9:
            season = "SUMMER"
            break
        case 10:
        case 11:
        case 12:
            season = "FALL"
            break
    }
    const pastSeason = season === "WINTER" ? "FALL" : season === "SPRING" ? "WINTER" : season === "SUMMER" ? "SPRING" : "SUMMER"
    const pastYear = season === "WINTER" ? currentYear - 1 : currentYear

    return useSeaQuery<ListMediaQuery, ListMediaQueryVariables>({
        queryKey: ["discover-highest-rating-last-season-anime"],
        endpoint: SeaEndpoints.ANILIST_LIST_ANIME,
        method: "post",
        data: {
            page: 1,
            perPage: 20,
            sort: ["SCORE_DESC"],
            season: pastSeason,
            seasonYear: pastYear,
        },
        enabled: isInView,
    })
}

export function useDiscoverUpcomingAnime(ref: any) {
    const _isInView = useInView(ref)
    const [isInView, setIsInView] = React.useState(false)
    React.useEffect(() => {
        if (isInView) return
        setIsInView(_isInView)
    }, [_isInView])
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
        enabled: isInView,
    })
}

export function useDiscoverPopularAnime(ref: any) {
    const _isInView = useInView(ref)
    const [isInView, setIsInView] = React.useState(false)
    React.useEffect(() => {
        if (isInView) return
        setIsInView(_isInView)
    }, [_isInView])
    return useSeaQuery<ListMediaQuery>({
        queryKey: ["discover-popular-anime"],
        endpoint: SeaEndpoints.ANILIST_LIST_ANIME,
        method: "post",
        data: {
            page: 1,
            perPage: 20,
            sort: ["POPULARITY_DESC"],
        },
        enabled: isInView,
    })
}

export function useDiscoverTrendingMovies(ref: any) {
    const _isInView = useInView(ref)
    const [isInView, setIsInView] = React.useState(false)
    React.useEffect(() => {
        if (isInView) return
        setIsInView(_isInView)
    }, [_isInView])
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
        enabled: isInView,
    })
}
