import type { OnlinestreamSubtitlePreference } from "@/app/(main)/onlinestream/_lib/onlinestream.atoms"

export type OnlinestreamSubtitleTrack = {
    number: number
    language?: string
    label?: string
}

function normalize(value: string | null | undefined) {
    return value?.trim().toLowerCase() ?? ""
}

export function isDefaultSubtitleTrack(tracks: Array<{ isDefault: boolean }>, index: number) {
    return tracks[index]?.isDefault || (!tracks.some(track => track.isDefault) && index === 0)
}

export function findPreferredSubtitleTrack<T extends OnlinestreamSubtitleTrack>(
    tracks: T[],
    preference: OnlinestreamSubtitlePreference | undefined,
) {
    if (!preference || preference.off) return null

    const language = normalize(preference.language)
    const label = normalize(preference.label)
    const exact = tracks.find(track => {
        if (language && normalize(track.language) !== language) return false
        if (label && normalize(track.label) !== label) return false
        return !!language || !!label
    })
    if (exact) return exact

    if (language) {
        const byLanguage = tracks.find(track => normalize(track.language) === language)
        if (byLanguage) return byLanguage
    }

    if (label) {
        return tracks.find(track => normalize(track.label) === label) ?? null
    }

    return null
}
