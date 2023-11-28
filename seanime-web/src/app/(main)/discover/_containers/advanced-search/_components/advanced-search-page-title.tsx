import { useMemo } from "react"
import startCase from "lodash/startCase"
import { useAtomValue } from "jotai/react"
import {
    __advancedSearch_getValue,
    __advancedSearch_paramsAtom,
} from "@/app/(main)/discover/_containers/advanced-search/_lib/parameters"

export function AdvancedSearchPageTitle() {

    const params = useAtomValue(__advancedSearch_paramsAtom)

    const title = useMemo(() => {
        if (params.title && params.title.length > 0) return startCase(params.title)
        if (!!__advancedSearch_getValue(params.genre)) return params.genre?.join(", ")
        if (__advancedSearch_getValue(params.sorting)?.includes("SCORE_DESC")) return "Most liked shows"
        if (__advancedSearch_getValue(params.sorting)?.includes("TRENDING_DESC")) return "Trending"
        if (__advancedSearch_getValue(params.sorting)?.includes("POPULARITY_DESC")) return "Popular"
        if (__advancedSearch_getValue(params.sorting)?.includes("START_DATE_DESC")) return "Latest"
        if (__advancedSearch_getValue(params.sorting)?.includes("EPISODES_DESC")) return "Most episodes"
        return "Most liked shows"
    }, [params.title, params.genre, params.sorting])

    return (
        <h2>{title}</h2>
    )
}
