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
    return new Date(date).toLocaleString("en-US", {
        dateStyle: "short",
        timeStyle: "short",
    })
}
