import { __advancedSearch_getValue, __advancedSearch_paramsAtom } from "@/app/(main)/search/_lib/parameters"
import { useAtomValue } from "jotai/react"
import startCase from "lodash/startCase"
import React from "react"

export function AdvancedSearchPageTitle() {

    const params = useAtomValue(__advancedSearch_paramsAtom)

    const title = React.useMemo(() => {
        if (params.title && params.title.length > 0) return startCase(params.title)
        if (!!__advancedSearch_getValue(params.genre)) return params.genre?.join(", ")
        if (__advancedSearch_getValue(params.sorting)?.includes("SCORE_DESC")) {
            if (params.type === "anime") {
                return "Most liked shows"
            } else {
                return "Most liked manga"
            }
        }
        if (__advancedSearch_getValue(params.sorting)?.includes("TRENDING_DESC")) return "Trending"
        if (__advancedSearch_getValue(params.sorting)?.includes("POPULARITY_DESC")) return "Popular"
        if (__advancedSearch_getValue(params.sorting)?.includes("START_DATE_DESC")) return "Latest"
        if (__advancedSearch_getValue(params.sorting)?.includes("EPISODES_DESC")) return "Most episodes"
        if (__advancedSearch_getValue(params.sorting)?.includes("CHAPTERS_DESC")) return "Most chapters"
        return params.type === "anime" ? "Most liked shows" : "Most liked manga"
    }, [params.title, params.genre, params.sorting, params.type])

    return (
        <h2>{title}</h2>
    )
}
