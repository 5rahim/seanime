import { AL_MediaFormat, AL_MediaSeason, AL_MediaSort, AL_MediaStatus } from "@/api/generated/types"
import { atomWithImmer } from "jotai-immer"

type Params = {
    active: boolean
    title: string | null
    sorting: AL_MediaSort[] | null
    genre: string[] | null
    status: AL_MediaStatus[] | null
    format: AL_MediaFormat | null
    season: AL_MediaSeason | null
    year: string | null
    minScore: string | null
    isAdult: boolean
    countryOfOrigin: string | null
    type: "anime" | "manga"
}

export const __advancedSearch_paramsAtom = atomWithImmer<Params>({
    active: true,
    title: null,
    sorting: null,
    status: null,
    genre: null,
    format: null,
    season: null,
    year: null,
    minScore: null,
    isAdult: false,
    countryOfOrigin: null,
    type: "anime",
})

export function __advancedSearch_getValue<T extends any>(value: T | ""): any {
    if (value === "") return undefined
    if (Array.isArray(value) && value.filter(Boolean).length === 0) return undefined
    if (typeof value === "string" && !isNaN(parseInt(value))) return Number(value)
    if (value === null) return undefined
    return value
}
