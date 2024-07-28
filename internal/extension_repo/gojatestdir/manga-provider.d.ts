declare type SearchResult = {
    provider: string
    id: string
    title: string
    synonyms?: string[]
    year?: number
    image?: string
    searchRating?: number
}

declare type ChapterDetails = {
    provider: string
    id: string
    url: string
    title: string
    chapter: string
    index: number
    rating?: number
    updatedAt?: string
}

declare type ChapterPage = {
    provider: string
    url: string
    index: number
    headers: { [key: string]: string }
}

declare function $findBestMatchWithSorensenDice(title: string, allTitles: string[]): {
    originalValue: string
    value: string
    rating: number
} | undefined


declare abstract class MangaProvider {
    search(query: string): Promise<SearchResult[]>

    findChapters(id: string): Promise<ChapterDetails[]>

    findChapterPages(id: string): Promise<ChapterPage[]>
}
