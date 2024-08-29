declare type AnimeProviderSmartSearchFilter = "batch" | "episodeNumber" | "resolution" | "query" | "bestReleases"

declare type AnimeProviderType = "main" | "special"

declare interface AnimeProviderSettings {
    canSmartSearch: boolean
    smartSearchFilters: AnimeProviderSmartSearchFilter[]
    supportsAdult: boolean
    type: AnimeProviderType
}

declare interface Media {
    id: number
    idMal?: number
    status?: string
    format?: string
    englishTitle?: string
    romajiTitle?: string
    episodeCount?: number
    absoluteSeasonOffset?: number
    synonyms: string[]
    isAdult: boolean
    startDate?: FuzzyDate
}

declare interface FuzzyDate {
    year: number
    month?: number
    day?: number
}

declare interface AnimeSearchOptions {
    media: Media
    query: string
}

declare interface AnimeSmartSearchOptions {
    media: Media
    query: string
    batch: boolean
    episodeNumber: number
    resolution: string
    anidbAID: number
    anidbEID: number
    bestReleases: boolean
}

declare interface AnimeTorrent {
    name: string
    date: string
    size: number
    formattedSize: string
    seeders: number
    leechers: number
    downloadCount: number
    link: string
    downloadUrl: string
    magnetLink?: string
    infoHash?: string
    resolution?: string
    isBatch?: boolean
    episodeNumber?: number
    releaseGroup?: string
    isBestRelease: boolean
    confirmed: boolean
}

declare interface AnimeTorrentProvider {
    // Returns the search results depending on the query.
    search(opts: AnimeSearchOptions): Promise<AnimeTorrent[]>

    // Returns the search results depending on the search options.
    smartSearch(opts: AnimeSmartSearchOptions): Promise<AnimeTorrent[]>

    // Returns the info hash of the torrent.
    // This should just return the info hash without scraping the torrent page if already available.
    getTorrentInfoHash(torrent: AnimeTorrent): Promise<string>

    // Returns the magnet link of the torrent.
    // This should just return the magnet link without scraping the torrent page if already available.
    getTorrentMagnetLink(torrent: AnimeTorrent): Promise<string>

    // Returns the latest torrents.
    getLatest(): Promise<AnimeTorrent[]>

    // Returns the provider settings.
    getSettings(): AnimeProviderSettings
}
