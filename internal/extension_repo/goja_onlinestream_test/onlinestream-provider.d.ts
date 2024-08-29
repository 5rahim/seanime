declare type SearchResult = {
    id: string
    title: string
    url: string
    subOrDub: SubOrDub
}

declare type SubOrDub = "sub" | "dub" | "both"

declare type EpisodeDetails = {
    id: string
    number: number
    url: string
    title?: string
}

declare type EpisodeServer = {
    server: string
    headers: { [key: string]: string }
    videoSources: VideoSource[]
}

declare type VideoSourceType = "mp4" | "m3u8"

declare type VideoSource = {
    url: string
    type: VideoSourceType
    quality: string
    subtitles: VideoSubtitle[]
}

declare type VideoSubtitle = {
    id: string
    url: string
    language: string
    isDefault: boolean
}

declare type SearchOptions = {
    query: string
    dub: boolean
    year?: number
}

declare type Settings = {
    episodeServers: string[]
    supportsDub: boolean
}

declare abstract class AnimeProvider {
    search(opts: SearchOptions): Promise<SearchResult[]>

    findEpisodes(id: string): Promise<EpisodeDetails[]>

    findEpisodeServer(episode: EpisodeDetails, server: string): Promise<EpisodeServer>

    getSettings(): Settings
}
