import { MediaFormat, MediaSeason, MediaSort, MediaStatus } from "@/lib/anilist/gql/graphql"
import { atomWithImmer } from "jotai-immer"

type Params = {
    active: boolean
    title: string | null
    sorting: MediaSort[] | null
    genre: string[] | null
    status: MediaStatus[] | null
    format: MediaFormat | null
    season: MediaSeason | null
    year: string | null
    minScore: string | null
}

export const __advancedSearch_paramsAtom = atomWithImmer<Params>({
    active: false,
    title: null,
    sorting: null,
    status: null,
    genre: null,
    format: null,
    season: null,
    year: null,
    minScore: null,
})

export function __advancedSearch_getValue<T extends any>(value: T | ""): any {
    if (value === "") return undefined
    if (Array.isArray(value) && value.filter(Boolean).length === 0) return undefined
    if (typeof value === "string" && !isNaN(parseInt(value))) return Number(value)
    if (value === null) return undefined
    return value
}