import { MediaFormat, MediaListStatus, MediaRelation, MediaSeason, MediaStatus, MediaType } from "@/lib/anilist/gql/graphql"

export type AnilistCollectionEntry = {
    id: number,
    score?: number | null,
    progress?: number | null,
    status?: MediaListStatus | null,
    notes?: string | null,
    repeat?: number | null,
    private?: boolean | null,
    startedAt?: { year?: number | null, month?: number | null, day?: number | null } | null,
    completedAt?: { year?: number | null, month?: number | null, day?: number | null } | null,
    media?: {
        id: number,
        idMal?: number | null,
        siteUrl?: string | null,
        status?: MediaStatus | null,
        season?: MediaSeason | null,
        type?: MediaType | null,
        format?: MediaFormat | null,
        bannerImage?: string | null,
        episodes?: number | null,
        synonyms?: Array<string | null> | null,
        isAdult?: boolean | null,
        countryOfOrigin?: any | null,
        meanScore?: number | null,
        description?: string | null,
        trailer?: { id?: string | null, site?: string | null, thumbnail?: string | null } | null,
        title?: { userPreferred?: string | null, romaji?: string | null, english?: string | null, native?: string | null } | null,
        coverImage?: { extraLarge?: string | null, large?: string | null, medium?: string | null, color?: string | null } | null,
        startDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
        endDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
        nextAiringEpisode?: { airingAt: number, timeUntilAiring: number, episode: number } | null,
        relations?: {
            edges?: Array<{
                relationType?: MediaRelation | null,
                node?: {
                    id: number,
                    idMal?: number | null,
                    siteUrl?: string | null,
                    status?: MediaStatus | null,
                    season?: MediaSeason | null,
                    type?: MediaType | null,
                    format?: MediaFormat | null,
                    bannerImage?: string | null,
                    episodes?: number | null,
                    synonyms?: Array<string | null> | null,
                    isAdult?: boolean | null,
                    countryOfOrigin?: any | null,
                    meanScore?: number | null,
                    description?: string | null,
                    trailer?: { id?: string | null, site?: string | null, thumbnail?: string | null } | null,
                    title?: { userPreferred?: string | null, romaji?: string | null, english?: string | null, native?: string | null } | null,
                    coverImage?: { extraLarge?: string | null, large?: string | null, medium?: string | null, color?: string | null } | null,
                    startDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
                    endDate?: { year?: number | null, month?: number | null, day?: number | null } | null,
                    nextAiringEpisode?: { airingAt: number, timeUntilAiring: number, episode: number } | null
                } | null
            } | null> | null
        } | null
    } | null
}

export type AnilistCollectionList = {
    status?: MediaListStatus | null, entries?: Array<AnilistCollectionEntry | null> | null
}
