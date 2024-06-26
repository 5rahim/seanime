import {
    AL_AnimeCollection_MediaListCollection_Lists_Entries,
    AL_MediaFormat,
    AL_MediaListStatus,
    AL_MediaSeason,
    AL_MediaStatus,
} from "@/api/generated/types"
import { useGetRawAnimeCollection } from "@/api/hooks/anilist.hooks"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { atomWithImmer } from "jotai-immer"
import sortBy from "lodash/sortBy"
import React, { useCallback } from "react"

type Sorting = "START_DATE" | "START_DATE_DESC" | "END_DATE" | "END_DATE_DESC" | "SCORE" | "SCORE_DESC" | "RELEASE_DATE" | "RELEASE_DATE_DESC"

type Params = {
    sorting: Sorting
    genre: string[] | null
    status: AL_MediaStatus[] | null
    format: AL_MediaFormat | null
    season: AL_MediaSeason | null
    year: string | null
    isAdult: boolean
}

export const __myListsSearch_paramsAtom = atomWithImmer<Params>({
    sorting: "SCORE_DESC",
    genre: null,
    status: null,
    format: null,
    season: null,
    year: null,
    isAdult: false,
})

function __myListsSearch_getParamValue<T extends any>(value: T | ""): any {
    if (value === "") return undefined
    if (Array.isArray(value) && value.filter(Boolean).length === 0) return undefined
    if (typeof value === "string" && !isNaN(parseInt(value))) return Number(value)
    if (value === null) return undefined
    return value
}

export function useHandleUserAnilistLists(debouncedSearchInput: string) {

    const serverStatus = useServerStatus()
    const { data } = useGetRawAnimeCollection()

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

    const customLists = React.useMemo(() => {
        return lists?.filter(obj => obj?.isCustomList)
    }, [lists])

    const getList = useCallback((status: AL_MediaListStatus, debouncedSearchInput: string) => {
        let obj = structuredClone(sortedLists?.find(n => n?.status === status))
        if (!obj || !obj.entries) return undefined
        if (!serverStatus?.settings?.anilist?.enableAdultContent) {
            obj.entries = obj.entries?.filter(entry => !entry?.media?.isAdult)
        }
        obj.entries = filterEntriesByTitle(obj.entries, debouncedSearchInput)
        return obj
    }, [sortedLists, serverStatus?.settings?.anilist?.enableAdultContent])

    const currentList = React.useMemo(() => getList("CURRENT", debouncedSearchInput), [debouncedSearchInput, getList, lists])
    const planningList = React.useMemo(() => getList("PLANNING", debouncedSearchInput), [debouncedSearchInput, getList, lists])
    const pausedList = React.useMemo(() => getList("PAUSED", debouncedSearchInput), [debouncedSearchInput, getList, lists])
    const completedList = React.useMemo(() => getList("COMPLETED", debouncedSearchInput), [debouncedSearchInput, getList, lists])
    const droppedList = React.useMemo(() => getList("DROPPED", debouncedSearchInput), [debouncedSearchInput, getList, lists])

    return {
        currentList,
        planningList,
        pausedList,
        completedList,
        droppedList,
        customLists,
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
