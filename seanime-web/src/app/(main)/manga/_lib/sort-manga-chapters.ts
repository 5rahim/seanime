/** Return chapters in reading order without changing the provider's original list. */
export function sortMangaChaptersAscending<T extends { chapter: string }>(chapters: T[]): T[] {
    const chapterNumber = (chapter: string) => {
        const number = chapter.match(/\d+(?:\.\d+)?/)?.[0]
        return number === undefined ? Number.POSITIVE_INFINITY : Number(number)
    }

    return [...chapters].sort((a, b) => chapterNumber(a.chapter) - chapterNumber(b.chapter))
}
