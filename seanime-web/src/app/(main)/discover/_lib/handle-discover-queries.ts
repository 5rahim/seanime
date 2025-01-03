import { AL_MediaSeason } from "@/api/generated/types"
import { useAnilistListAnime } from "@/api/hooks/anilist.hooks"
import { useInView } from "framer-motion"
import { atom } from "jotai"
import { useAtomValue } from "jotai/react"

export const __discover_trendingGenresAtom = atom<string[]>([])

export function useDiscoverTrendingAnime() {
    const genres = useAtomValue(__discover_trendingGenresAtom)

    return useAnilistListAnime({
        page: 1,
        perPage: 20,
        sort: ["TRENDING_DESC"],
        genres: genres.length > 0 ? genres : undefined,
    }, true)

}

export function useDiscoverPastSeasonAnime(ref: any) {
    const isInView = useInView(ref, { once: true })
    const currentMonth = new Date().getMonth() + 1
    const currentYear = new Date().getFullYear()
    let season: AL_MediaSeason = "SUMMER"
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

    return useAnilistListAnime({
        page: 1,
        perPage: 20,
        sort: ["SCORE_DESC"],
        season: pastSeason,
        seasonYear: pastYear,
    }, isInView)
}

export function useDiscoverUpcomingAnime(ref: any) {
    const isInView = useInView(ref, { once: true })
    return useAnilistListAnime({
        page: 1,
        perPage: 20,
        sort: ["TRENDING_DESC"],
        status: ["NOT_YET_RELEASED"],
    }, isInView)
}

export function useDiscoverPopularAnime(ref: any) {
    const isInView = useInView(ref, { once: true })
    return useAnilistListAnime({
        page: 1,
        perPage: 20,
        sort: ["POPULARITY_DESC"],
    }, isInView)
}

export function useDiscoverTrendingMovies(ref: any) {
    const isInView = useInView(ref, { once: true })
    return useAnilistListAnime({
        page: 1,
        perPage: 20,
        format: "MOVIE",
        sort: ["TRENDING_DESC"],
    }, isInView)
}
