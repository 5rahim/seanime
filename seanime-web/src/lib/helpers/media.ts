import { AL_AnimeListEntry, AL_BaseAnime, AL_MangaListEntry, Nullish } from "@/api/generated/types"

export function anilist_getTotalEpisodes(anime: Nullish<AL_BaseAnime>) {
    if (!anime) return -1
    let maxEp = anime?.episodes ?? -1
    if (maxEp === -1) {
        if (anime.nextAiringEpisode && anime.nextAiringEpisode.episode) {
            maxEp = anime.nextAiringEpisode.episode - 1
        }
    }
    if (maxEp === -1) {
        return 0
    }
    return maxEp
}

export function anilist_getCurrentEpisodes(anime: Nullish<AL_BaseAnime>) {
    if (!anime) return -1
    let maxEp = -1
    if (anime.nextAiringEpisode && anime.nextAiringEpisode.episode) {
        maxEp = anime.nextAiringEpisode.episode - 1
    }
    if (maxEp === -1) {
        maxEp = anime.episodes ?? 0
    }
    return maxEp
}

export function anilist_getListDataFromEntry(entry: Nullish<AL_AnimeListEntry | AL_MangaListEntry>) {
    return {
        progress: entry?.progress,
        score: entry?.score,
        status: entry?.status,
        startedAt: new Date(entry?.startedAt?.year || 0,
            entry?.startedAt?.month ? entry?.startedAt?.month - 1 : 0,
            entry?.startedAt?.day || 0).toUTCString(),
        completedAt: new Date(entry?.completedAt?.year || 0,
            entry?.completedAt?.month ? entry?.completedAt?.month - 1 : 0,
            entry?.completedAt?.day || 0).toUTCString(),
    }
}


export function anilist_animeIsMovie(anime: Nullish<AL_BaseAnime>) {
    if (!anime) return false
    return anime?.format === "MOVIE"

}

export function anilist_animeIsSingleEpisode(anime: Nullish<AL_BaseAnime>) {
    if (!anime) return false
    return anime?.format === "MOVIE" || anime?.episodes === 1
}


export function anilist_getUnwatchedCount(anime: Nullish<AL_BaseAnime>, progress: Nullish<number>) {
    if (!anime) return false
    const maxEp = anilist_getCurrentEpisodes(anime)
    return maxEp - (progress ?? 0)
}

