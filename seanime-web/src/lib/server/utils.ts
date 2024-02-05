import capitalize from "lodash/capitalize"

export function getLibraryCollectionTitle(type: string) {
    switch (type) {
        case "current":
            return "Currently watching"
        default:
            return capitalize(type)
    }
}

export function getAniDBEpisodeInteger<T extends {
    metadata: { aniDBEpisode?: string }
}>(props: T | null | undefined) {
    const metadata = props?.metadata
    if (!metadata || !metadata.aniDBEpisode) return undefined
    const parsed = Number(metadata.aniDBEpisode.replace(/\D/g, ""))
    return !isNaN(parsed) ? parsed : undefined
}

export function formatDateAndTimeShort(date: string) {
    return new Date(date).toLocaleString("en-US", {
        dateStyle: "short",
        timeStyle: "short",
    })
}
