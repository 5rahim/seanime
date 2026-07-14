import { Manga_MangaLatestChapterNumberItem } from "@/api/generated/types"
import { MangaEntryFilters } from "./manga-preferences"

export type MangaUnreadState = {
    known: boolean
    latest: number | null
    unread: number
}

export function getMangaEntryLatestChapterNumber(
    mangaId: string | number,
    latestChapterNumbers: Record<number, Manga_MangaLatestChapterNumberItem[]>,
    storedProviders: Record<string, string>,
    storedFilters: Record<string, MangaEntryFilters>,
) {
    const provider = storedProviders[String(mangaId)]
    if (!provider) return null

    const filters = storedFilters[String(mangaId)]
    const scanlators = filters?.scanlators?.filter(Boolean) ?? []
    const language = filters?.language
    const entries = latestChapterNumbers[Number(mangaId)]?.filter(item => {
        if (item.provider !== provider) return false
        if (language && item.language !== language) return false
        if (scanlators.length && !scanlators.includes(item.scanlator)) return false
        return true
    }) ?? []

    if (!entries.length) return null

    return entries.reduce((latest, item) => Math.max(latest, item.number), 0)
}

export function getMangaEntryUnreadState(
    mangaId: string | number,
    progress: number,
    latestChapterNumbers: Record<number, Manga_MangaLatestChapterNumberItem[]>,
    storedProviders: Record<string, string>,
    storedFilters: Record<string, MangaEntryFilters>,
): MangaUnreadState {
    const latest = getMangaEntryLatestChapterNumber(mangaId, latestChapterNumbers, storedProviders, storedFilters)
    if (latest === null || !Number.isFinite(latest)) {
        return { known: false, latest: null, unread: 0 }
    }

    return {
        known: true,
        latest,
        unread: Math.max(0, latest - progress),
    }
}
