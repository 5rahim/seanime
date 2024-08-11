declare type SearchResult = {
    id: string
    title: string
    synonyms?: string[]
    year?: number
    image?: string
}

declare type ChapterDetails = {
    id: string
    url: string
    title: string
    chapter: string
    index: number
    rating?: number
    updatedAt?: string
}

declare type ChapterPage = {
    url: string
    index: number
    headers: { [key: string]: string }
}

declare type QueryOptions = {
    query: string
    year?: number
}


declare abstract class MangaProvider {
    search(opts: QueryOptions): Promise<SearchResult[]>
    findChapters(id: string): Promise<ChapterDetails[]>
    findChapterPages(id: string): Promise<ChapterPage[]>
}
