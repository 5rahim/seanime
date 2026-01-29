import { buildSeaQuery } from "@/api/client/requests"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { AL_ListAnime, AL_ListManga } from "@/api/generated/types"
import { serverAuthTokenAtom } from "@/app/(main)/_atoms/server-status.atoms"
import { __advancedSearch_getValue, __advancedSearch_paramsAtom } from "@/app/(main)/search/_lib/advanced-search.atoms"
import { useInfiniteQuery } from "@tanstack/react-query"
import { useAtomValue } from "jotai/react"
import React from "react"

export function useAnilistAdvancedSearch() {

    const params = useAtomValue(__advancedSearch_paramsAtom)
    const password = useAtomValue(serverAuthTokenAtom)

    const { isLoading: isLoading1, data: data1, fetchNextPage: fetchNextPage1, hasNextPage: hasNextPage1 } = useInfiniteQuery({
        queryKey: ["advanced-search-anime", params],
        initialPageParam: 1,
        queryFn: async ({ pageParam }) => {
            const variables = {
                page: pageParam,
                perPage: 48,
                format: __advancedSearch_getValue(params.format)?.toUpperCase(),
                search: (params.title === null || params.title === "") ? undefined : params.title,
                genres: __advancedSearch_getValue(params.genre),
                season: __advancedSearch_getValue(params.season),
                seasonYear: __advancedSearch_getValue(params.year),
                averageScore_greater: __advancedSearch_getValue(params.minScore) !== undefined
                    ? __advancedSearch_getValue(params.minScore)
                    : undefined,
                sort: (params.title?.length && params.title.length > 0) ? ["SEARCH_MATCH",
                    ...(__advancedSearch_getValue(params.sorting) || ["SCORE_DESC"])] : (__advancedSearch_getValue(params.sorting) || ["SCORE_DESC"]),
                status: params.sorting?.includes("START_DATE_DESC") ? (__advancedSearch_getValue(params.status)
                    ?.filter((n: string) => n !== "NOT_YET_RELEASED") || ["FINISHED", "RELEASING"]) : __advancedSearch_getValue(params.status),
                isAdult: params.isAdult,
            }

            return buildSeaQuery<AL_ListAnime>({
                endpoint: API_ENDPOINTS.ANILIST.AnilistListAnime.endpoint,
                method: "POST",
                data: variables,
                password: password,
            })
        },
        getNextPageParam: (lastPage, pages) => {
            const curr = lastPage?.Page?.pageInfo?.currentPage
            const hasNext = lastPage?.Page?.pageInfo?.hasNextPage
            // console.log("lastPage", lastPage, "pages", pages, "curr", curr, "hasNext", hasNext, "nextPage", (!!curr && hasNext) ? pages.length + 1
            // : undefined)
            return (!!curr && hasNext) ? pages.length + 1 : undefined
        },
        enabled: params.active && params.type === "anime",
        refetchOnMount: true,
    })

    const { isLoading: isLoading2, data: data2, fetchNextPage: fetchNextPage2, hasNextPage: hasNextPage2 } = useInfiniteQuery({
        queryKey: ["advanced-search-manga", params],
        initialPageParam: 1,
        queryFn: async ({ pageParam }) => {
            const variables = {
                page: pageParam,
                perPage: 48,
                search: (params.title === null || params.title === "") ? undefined : params.title,
                genres: __advancedSearch_getValue(params.genre),
                year: __advancedSearch_getValue(params.year),
                format: __advancedSearch_getValue(params.format)?.toUpperCase(),
                averageScore_greater: __advancedSearch_getValue(params.minScore) !== undefined
                    ? __advancedSearch_getValue(params.minScore)
                    : undefined,
                sort: (params.title?.length && params.title.length > 0) ? ["SEARCH_MATCH",
                    ...(__advancedSearch_getValue(params.sorting) || ["SCORE_DESC"])] : (__advancedSearch_getValue(params.sorting) || ["SCORE_DESC"]),
                status: params.sorting?.includes("START_DATE_DESC") ? (__advancedSearch_getValue(params.status)
                    ?.filter((n: string) => n !== "NOT_YET_RELEASED") || ["FINISHED", "RELEASING"]) : __advancedSearch_getValue(params.status),
                countryOfOrigin: __advancedSearch_getValue(params.countryOfOrigin),
                isAdult: params.isAdult,
            }

            return buildSeaQuery<AL_ListManga>({
                endpoint: API_ENDPOINTS.MANGA.AnilistListManga.endpoint,
                method: "POST",
                data: variables,
                password: password,
            })
        },
        getNextPageParam: (lastPage, pages) => {
            const curr = lastPage?.Page?.pageInfo?.currentPage
            const hasNext = lastPage?.Page?.pageInfo?.hasNextPage
            // console.log("lastPage", lastPage, "pages", pages, "curr", curr, "hasNext", hasNext, "nextPage", (!!curr && hasNext) ? pages.length + 1
            // : undefined)
            return (!!curr && hasNext) ? pages.length + 1 : undefined
        },
        enabled: params.active && params.type === "manga",
        refetchOnMount: true,
    })

    const isLoading = React.useMemo(() => params.type === "anime" ? isLoading1 : isLoading2, [isLoading1, isLoading2, params.type])
    const data = React.useMemo(() => params.type === "anime" ? data1 : data2, [data1, data2, params.type])
    const fetchNextPage = React.useMemo(() => params.type === "anime" ? fetchNextPage1 : fetchNextPage2,
        [fetchNextPage1, fetchNextPage2, params.type])
    const hasNextPage = React.useMemo(() => params.type === "anime" ? hasNextPage1 : hasNextPage2, [hasNextPage1, hasNextPage2, params.type])

    return {
        isLoading,
        data,
        fetchNextPage,
        hasNextPage,
        type: params.type,
    }
}
