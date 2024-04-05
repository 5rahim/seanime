import { __advancedSearch_getValue, __advancedSearch_paramsAtom } from "@/app/(main)/discover/_containers/advanced-search/_lib/parameters"
import { ListMediaQuery } from "@/lib/anilist/gql/graphql"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { buildSeaQuery } from "@/lib/server/query"
import { useInfiniteQuery } from "@tanstack/react-query"
import { useAtomValue } from "jotai/react"

export function useAnilistAdvancedSearch() {

    const params = useAtomValue(__advancedSearch_paramsAtom)

    return useInfiniteQuery({
        queryKey: ["projects", params],
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
                averageScore_greater: __advancedSearch_getValue(params.minScore) !== undefined ? __advancedSearch_getValue(params.minScore) : undefined,
                sort: (params.title?.length && params.title.length > 0) ? ["SEARCH_MATCH", ...(__advancedSearch_getValue(params.sorting) || ["SCORE_DESC"])] : (__advancedSearch_getValue(params.sorting) || ["SCORE_DESC"]),
                status: params.sorting?.includes("START_DATE_DESC") ? (__advancedSearch_getValue(params.status)?.filter((n: string) => n !== "NOT_YET_RELEASED") || ["FINISHED", "RELEASING"]) : __advancedSearch_getValue(params.status),
                isAdult: params.isAdult,
            }

            return buildSeaQuery<ListMediaQuery>({
                endpoint: SeaEndpoints.ANILIST_LIST_ANIME,
                method: "post",
                data: variables,
            })
        },
        getNextPageParam: (lastPage, pages) => {
            const curr = lastPage?.Page?.pageInfo?.currentPage
            const hasNext = lastPage?.Page?.pageInfo?.hasNextPage
            // console.log("lastPage", lastPage, "pages", pages, "curr", curr, "hasNext", hasNext, "nextPage", (!!curr && hasNext) ? pages.length + 1 : undefined)
            return (!!curr && hasNext) ? pages.length + 1 : undefined
        },
        enabled: params.active,
        refetchOnMount: true,
    })
}
