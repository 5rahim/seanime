export type HlsQualityOption = {
    index: number
    height: number
    bitrate: number
}

export function getPreferredHlsQualityLevel(levels: HlsQualityOption[], quality: string | undefined) {
    if (!levels.length || !quality) return null

    const normalized = quality.trim().toLowerCase()
    const height = Number(normalized.match(/\b(\d{3,4})p\b/)?.[1])
    if (!height) return /\b(auto|default)\b/.test(normalized) ? -1 : null

    const preferredLevel = levels.find(level => level.height === height)
    if (preferredLevel) return preferredLevel.index

    return levels.reduce((lowest, level) => level.bitrate < lowest.bitrate ? level : lowest).index
}
