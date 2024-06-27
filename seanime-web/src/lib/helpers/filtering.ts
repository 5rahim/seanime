import {
    AL_AnimeCollection_MediaListCollection_Lists_Entries,
    AL_MangaCollection_MediaListCollection_Lists_Entries,
    AL_MediaFormat,
    AL_MediaSeason,
    AL_MediaStatus,
} from "@/api/generated/types"
import sortBy from "lodash/sortBy"

type CollectionSorting =
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

export const COLLECTION_SORTING_OPTIONS = [
    { label: "Highest score", value: "SCORE_DESC" },
    { label: "Lowest score", value: "SCORE" },
    { label: "Highest progress", value: "PROGRESS_DESC" },
    { label: "Lowest progress", value: "PROGRESS" },
    { label: "Started recently", value: "START_DATE_DESC" },
    { label: "Oldest start date", value: "START_DATE" },
    { label: "Completed recently", value: "END_DATE_DESC" },
    { label: "Oldest completion date", value: "END_DATE" },
    { label: "Released recently", value: "RELEASE_DATE_DESC" },
    { label: "Oldest release", value: "RELEASE_DATE" },
]

export type CollectionParams = {
    sorting: CollectionSorting
    genre: string[] | null
    status: AL_MediaStatus | null
    format: AL_MediaFormat | null
    season: AL_MediaSeason | null
    year: string | null
    isAdult: boolean
}

export const DEFAULT_COLLECTION_PARAMS: CollectionParams = {
    sorting: "SCORE_DESC",
    genre: null,
    status: null,
    format: null,
    season: null,
    year: null,
    isAdult: false,
}


function getParamValue<T extends any>(value: T | ""): any {
    if (value === "") return undefined
    if (Array.isArray(value) && value.filter(Boolean).length === 0) return undefined
    if (typeof value === "string" && !isNaN(parseInt(value))) return Number(value)
    if (value === null) return undefined
    return value
}


export function filterEntriesByTitle(arr: AL_AnimeCollection_MediaListCollection_Lists_Entries[] | null | undefined, input: string) {
    if (!arr) return []
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

export function filterCollectionEntries<T extends AL_MangaCollection_MediaListCollection_Lists_Entries[] | AL_AnimeCollection_MediaListCollection_Lists_Entries[]>(
    entries: T | null | undefined,
    params: CollectionParams,
    showAdultContent: boolean | undefined,
) {
    if (!entries) return []
    let arr = [...entries]

    // Filter by isAdult
    if (!!arr && params.isAdult) arr = arr.filter(n => n.media?.isAdult)

    // Filter by showAdultContent
    if (!showAdultContent) arr = arr.filter(n => !n.media?.isAdult)

    // Filter by format
    if (!!arr && !!params.format) arr = arr.filter(n => n.media?.format === params.format)

    // Filter by season
    if (!!arr && !!params.season) arr = arr.filter(n => n.media?.season === params.season)

    // Filter by status
    if (!!arr && !!params.status) arr = arr.filter(n => n.media?.status === params.status)

    // Filter by year
    if (!!arr && !!params.year) arr = arr.filter(n => n.media?.startDate?.year === Number(params.year))

    // Filter by genre
    if (!!arr && !!params.genre?.length) {
        arr = arr.filter(n => {
            return params.genre?.every(genre => n.media?.genres?.includes(genre))
        })
    }

    // Sort by name
    arr = sortBy(arr, n => n?.media?.title?.userPreferred).reverse()

    // Sort by release date
    if (getParamValue(params.sorting) === "RELEASE_DATE" || getParamValue(params.sorting) === "RELEASE_DATE_DESC") {
        arr = arr?.filter(n => n.media?.startDate && !!n.media.startDate.year && !!n.media.startDate.month)
    }
    if (getParamValue(params.sorting) === "RELEASE_DATE")
        arr = sortBy(arr, n => new Date(n?.media?.startDate?.year!, n?.media?.startDate?.month! - 1))
    if (getParamValue(params.sorting) === "RELEASE_DATE_DESC")
        arr = sortBy(arr, n => new Date(n?.media?.startDate?.year!, n?.media?.startDate?.month! - 1)).reverse()

    // Sort by score
    if (getParamValue(params.sorting) === "SCORE")
        arr = sortBy(arr, n => n?.score)
    if (getParamValue(params.sorting) === "SCORE_DESC")
        arr = sortBy(arr, n => n?.score).reverse()

    // Sort by start date
    if (getParamValue(params.sorting) === "START_DATE" || getParamValue(params.sorting) === "START_DATE_DESC") {
        arr = arr?.filter(n => n.startedAt && !!n.startedAt.year && !!n.startedAt.month && !!n.startedAt.day)
    }
    if (getParamValue(params.sorting) === "START_DATE")
        arr = sortBy(arr, n => new Date(n?.startedAt?.year!, n?.startedAt?.month! - 1, n?.startedAt?.day))
    if (getParamValue(params.sorting) === "START_DATE_DESC")
        arr = sortBy(arr, n => new Date(n?.startedAt?.year!, n?.startedAt?.month! - 1, n?.startedAt?.day)).reverse()

    // Sort by end date
    if (getParamValue(params.sorting) === "END_DATE" || getParamValue(params.sorting) === "END_DATE_DESC") {
        arr = arr?.filter(n => n.completedAt && !!n.completedAt.year && !!n.completedAt.month && !!n.completedAt.day)
    }
    if (getParamValue(params.sorting) === "END_DATE")
        arr = sortBy(arr, n => new Date(n?.completedAt?.year!, n?.completedAt?.month! - 1, n?.completedAt?.day))
    if (getParamValue(params.sorting) === "END_DATE_DESC")
        arr = sortBy(arr, n => new Date(n?.completedAt?.year!, n?.completedAt?.month! - 1, n?.completedAt?.day)).reverse()

    // Sort by progress
    if (getParamValue(params.sorting) === "PROGRESS")
        arr = sortBy(arr, n => n?.progress || 0)
    if (getParamValue(params.sorting) === "PROGRESS_DESC")
        arr = sortBy(arr, n => n?.progress || 0).reverse()

    return arr
}
