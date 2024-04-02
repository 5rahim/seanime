import { BaseMediaFragment } from "@/lib/anilist/gql/graphql"

export type Onlinestream_Episode = {
    number: number
    title?: string
    description?: string
    image?: string
}

export type Onlinestream_EpisodeListResponse = {
    episodes: Onlinestream_Episode[]
    media: BaseMediaFragment
}

export type Onlinestream_EpisodeSource = {
    number: number
    videoSources: Onlinestream_VideoSource[]
    subtitles: Onlinestream_VideoSubtitles[] | undefined
}

export type Onlinestream_VideoSource = {
    headers: Record<string, string>
    server: string
    url: string
    quality: string
}

export type Onlinestream_VideoSubtitles = {
    url: string
    language: string
}

export const enum OnlinestreamProvider {
    GOGOANIME = "gogoanime",
    ZORO = "zoro",
}
