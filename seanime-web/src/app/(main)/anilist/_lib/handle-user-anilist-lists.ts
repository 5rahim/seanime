import {
    AL_AnimeCollection_MediaListCollection_Lists_Entries,
    AL_MediaFormat,
    AL_MediaListStatus,
    AL_MediaSeason,
    AL_MediaStatus,
} from "@/api/generated/types"
import { useGetRawAnimeCollection } from "@/api/hooks/anilist.hooks"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { useAtomValue } from "jotai"
import { atomWithImmer } from "jotai-immer"
import sortBy from "lodash/sortBy"
import React from "react"
import { useDebounce } from "use-debounce"

type Sorting =
    "START_DATE"
    | "START_DATE_DESC"
    | "END_DATE"
    | "END_DATE_DESC"
    | "SCORE"
    | "SCORE_DESC"
    | "RELEASE_DATE"
    | "RELEASE_DATE_DESC"
    | "PROGRESS"
    | "PROGRESS_DESC"

export const MYLISTS_SORTING_OPTIONS = [
    { label: "Highest score", value: "SCORE_DESC" },
    { label: "Lowest score", value: "SCORE" },
    { label: "Highest progress", value: "PROGRESS_DESC" },
    { label: "Lowest progress", value: "PROGRESS" },
    { label: "Recent start date", value: "START_DATE_DESC" },
    { label: "Oldest start date", value: "START_DATE" },
    { label: "Recent completion date", value: "END_DATE_DESC" },
    { label: "Oldest completion date", value: "END_DATE" },
    { label: "Oldest release", value: "RELEASE_DATE" },
    { label: "Released recently", value: "RELEASE_DATE_DESC" },
]

type Params = {
    sorting: Sorting
    genre: string[] | null
    status: AL_MediaStatus | null
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

export const __myListsSearch_paramsInputAtom = atomWithImmer<Params>({
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

    const params = useAtomValue(__myListsSearch_paramsAtom)
    const debouncedParams = useDebounce(params, 500)

    const sortedLists = React.useMemo(() => {
        return lists?.map(obj => {
            if (!obj) return undefined
            let arr = structuredClone(obj?.entries)

            // Filter by isAdult
            if (!!arr && params.isAdult) arr = arr.filter(n => n.media?.isAdult)

            // Filter by format
            if (!!arr && params.format) arr = arr.filter(n => n.media?.format === params.format)

            // Filter by season
            if (!!arr && params.season) arr = arr.filter(n => n.media?.season === params.season)

            // Filter by status
            if (!!arr && params.status) arr = arr.filter(n => n.media?.status === params.status)

            // Filter by year
            if (!!arr && params.year) arr = arr.filter(n => n.media?.startDate?.year === Number(params.year))

            // Sort by name
            arr = sortBy(arr, n => n?.media?.title?.userPreferred).reverse()

            // Sort by release date
            if (__myListsSearch_getParamValue(params.sorting) === "RELEASE_DATE" || __myListsSearch_getParamValue(params.sorting) === "RELEASE_DATE_DESC") {
                arr = obj.entries?.filter(n => n.media?.startDate && !!n.media.startDate.year && !!n.media.startDate.month)
            }
            if (__myListsSearch_getParamValue(params.sorting) === "RELEASE_DATE") arr = sortBy(arr,
                n => new Date(n?.media?.startDate?.year!, n?.media?.startDate?.month! - 1))
            if (__myListsSearch_getParamValue(params.sorting) === "RELEASE_DATE_DESC") arr = sortBy(arr,
                n => new Date(n?.media?.startDate?.year!, n?.media?.startDate?.month! - 1)).reverse()

            // Sort by score
            if (__myListsSearch_getParamValue(params.sorting) === "SCORE") arr = sortBy(arr, n => n?.score)
            if (__myListsSearch_getParamValue(params.sorting) === "SCORE_DESC") arr = sortBy(arr, n => n?.score).reverse()

            // Sort by start date
            if (__myListsSearch_getParamValue(params.sorting) === "START_DATE" || __myListsSearch_getParamValue(params.sorting) === "START_DATE_DESC") {
                arr = obj.entries?.filter(n => n.startedAt && !!n.startedAt.year && !!n.startedAt.month && !!n.startedAt.day)
            }
            if (__myListsSearch_getParamValue(params.sorting) === "START_DATE") arr = sortBy(arr,
                n => new Date(n?.startedAt?.year!, n?.startedAt?.month! - 1, n?.startedAt?.day))
            if (__myListsSearch_getParamValue(params.sorting) === "START_DATE_DESC") arr = sortBy(arr,
                n => new Date(n?.startedAt?.year!, n?.startedAt?.month! - 1, n?.startedAt?.day)).reverse()

            // Sort by end date
            if (__myListsSearch_getParamValue(params.sorting) === "END_DATE" || __myListsSearch_getParamValue(params.sorting) === "END_DATE_DESC") {
                arr = obj.entries?.filter(n => n.completedAt && !!n.completedAt.year && !!n.completedAt.month && !!n.completedAt.day)
            }
            if (__myListsSearch_getParamValue(params.sorting) === "END_DATE") arr = sortBy(arr,
                n => new Date(n?.completedAt?.year!, n?.completedAt?.month! - 1, n?.completedAt?.day))
            if (__myListsSearch_getParamValue(params.sorting) === "END_DATE_DESC") arr = sortBy(arr,
                n => new Date(n?.completedAt?.year!, n?.completedAt?.month! - 1, n?.completedAt?.day)).reverse()

            // Sort by progress
            if (__myListsSearch_getParamValue(params.sorting) === "PROGRESS") arr = sortBy(arr, n => n?.progress || 0)
            if (__myListsSearch_getParamValue(params.sorting) === "PROGRESS_DESC") arr = sortBy(arr, n => n?.progress || 0).reverse()


            // Filter by year

            return {
                ...obj,
                entries: arr,
            }
        }).filter(Boolean)
    }, [lists, debouncedParams])

    const customLists = React.useMemo(() => {
        if (debouncedSearchInput === "") return sortedLists?.filter(obj => obj?.isCustomList)
        return sortedLists?.filter(obj => obj?.isCustomList)?.map(obj => {
            if (!obj.entries) return undefined
            const entries = filterEntriesByTitle(obj?.entries, debouncedSearchInput)
            return { ...obj, entries }
        }).filter(Boolean)
    }, [sortedLists, debouncedSearchInput])

    const getList = React.useCallback((status: AL_MediaListStatus, debouncedSearchInput: string) => {
        const list = sortedLists?.find(n => n?.status === status)
        if (!list) return undefined

        const entries = list.entries?.filter(entry => {
            const title = entry.media?.title?.userPreferred?.toLowerCase()
            return title && title.includes(debouncedSearchInput.toLowerCase())
        })

        if (!serverStatus?.settings?.anilist?.enableAdultContent) {
            return {
                ...list,
                entries: entries?.filter(entry => !entry.media?.isAdult),
            }
        }

        return { ...list, entries }
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
