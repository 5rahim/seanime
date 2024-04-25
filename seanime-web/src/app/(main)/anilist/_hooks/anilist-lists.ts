import { AL_AnimeCollection_MediaListCollection_Lists_Entries, AL_MediaListStatus } from "@/api/generated/types"
import { useGetAnilistCollection } from "@/api/hooks/anilist.hooks"
import { useServerStatus } from "@/app/(main)/_hooks/server-status.hooks"
import sortBy from "lodash/sortBy"
import React, { useCallback } from "react"

export function getUserAnilistLists(debouncedSearchInput: string) {

    const serverStatus = useServerStatus()
    const { data } = useGetAnilistCollection()

    const lists = React.useMemo(() => data?.MediaListCollection?.lists, [data])

    const sortedLists = React.useMemo(() => {
        return lists?.map(obj => {
            if (!obj) return undefined
            let arr = obj?.entries
            // Sort by name
            arr = sortBy(arr, n => n?.media?.title?.userPreferred).reverse()
            // Sort by score
            arr = (sortBy(arr, n => n?.score).reverse())
            obj.entries = arr
            return obj
        })
    }, [lists])

    const getList = useCallback((status: AL_MediaListStatus) => {
        let obj = structuredClone(sortedLists?.find(n => n?.status === status))
        if (!obj || !obj.entries) return undefined
        if (!serverStatus?.settings?.anilist?.enableAdultContent) {
            obj.entries = obj.entries?.filter(entry => !entry?.media?.isAdult)
        }
        obj.entries = filterEntriesByTitle(obj.entries, debouncedSearchInput)
        return obj
    }, [sortedLists, debouncedSearchInput, serverStatus?.settings?.anilist?.enableAdultContent])

    const currentList = React.useMemo(() => getList("CURRENT"), [debouncedSearchInput, getList, lists])
    const planningList = React.useMemo(() => getList("PLANNING"), [debouncedSearchInput, getList, lists])
    const pausedList = React.useMemo(() => getList("PAUSED"), [debouncedSearchInput, getList, lists])
    const completedList = React.useMemo(() => getList("COMPLETED"), [debouncedSearchInput, getList, lists])
    const droppedList = React.useMemo(() => getList("DROPPED"), [debouncedSearchInput, getList, lists])

    return {
        currentList,
        planningList,
        pausedList,
        completedList,
        droppedList,
    }
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

function filterEntriesByTitle(arr: AL_AnimeCollection_MediaListCollection_Lists_Entries[], input: string) {
    if (arr.length > 0 && input.length > 0) {
        const _input = input.toLowerCase().trim().replace(/\s+/g, " ")
        return arr.filter(entry => (
            entry.media?.title?.english?.toLowerCase().includes(_input)
            || entry.media?.title?.userPreferred?.toLowerCase().includes(_input)
            || entry.media?.title?.romaji?.toLowerCase().includes(_input)
            || entry.media?.synonyms?.some(syn => syn?.toLowerCase().includes(_input))
        ))
    }
    return arr
}
