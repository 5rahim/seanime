import capitalize from "lodash/capitalize"

export function getLibraryCollectionTitle(type?: string) {
    switch (type) {
        case "CURRENT":
            return "Currently watching"
        default:
            return capitalize(type ?? "")
    }
}

export function getMangaCollectionTitle(type?: string) {
    switch (type) {
        case "CURRENT":
            return "Currently reading"
        default:
            return capitalize(type ?? "")
    }
}

export function formatDateAndTimeShort(date: string) {
    return new Date(date).toLocaleString(undefined, {
        dateStyle: "short",
        timeStyle: "short",
    })
}

export function isCustomSource(mId: number | null | undefined) {
    if (mId === null || mId === undefined) return false
    return mId >= 0x80000000
}

export function getCustomSourceExtensionId(m: { siteUrl?: string } | null | undefined) {
    if (!m?.siteUrl) return null
    let s = m.siteUrl.replace("ext_custom_source_", "")
    if (s.includes("|END|")) {
        s = s.split("|END|")[0]
    }
    return s
}

export function getCustomSourceMediaSiteUrl(m: { siteUrl?: string } | null | undefined) {
    if (!m?.siteUrl) return null
    let s = m.siteUrl.replace("ext_custom_source_", "")
    if (s.includes("|END|")) {
        s = s.split("|END|")?.[1] ?? null
    }
    return s
}
