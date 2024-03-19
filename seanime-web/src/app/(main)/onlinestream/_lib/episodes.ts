import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaQuery } from "@/lib/server/query"

type OnlinestreamEpisode = {
    number: number
    title?: string
    description?: string
    image?: string
}

type OnlinestreamEpisodeSource = {
    number: number
    sources: {
        url: string
        quality: string
    }[]
    subtitles: {
        url: string
        language: string
    }[] | undefined
}

const enum Provider {
    GOGOANIME = "gogoanime",
    ZORO = "zoro",
}

export function useOnlinestreamEpisodes(mId: string | null, dubbed: boolean) {

    const { data, isLoading, isFetching } = useSeaQuery<OnlinestreamEpisode[], { mediaId: number, provider: string, dubbed: boolean }>({
        endpoint: SeaEndpoints.ONLINESTREAM_EPISODES,
        method: "post",
        queryKey: ["onlinestream-episodes", mId, dubbed],
        data: {
            mediaId: Number(mId),
            dubbed,
            provider: Provider.ZORO,
        },
        enabled: !!mId,
    })

    return {
        episodes: data,
        isLoading,
        isFetching,
    }
}

export function useOnlinestreamEpisodeSources(mId: string | null, episodeNumber: number, dubbed: boolean) {

    const { data, isLoading, isFetching } = useSeaQuery<OnlinestreamEpisodeSource[], {
        mediaId: number,
        episodeNumber: number,
        provider: string,
        dubbed: boolean
    }>({
        endpoint: SeaEndpoints.ONLINESTREAM_EPISODE_SOURCES,
        method: "post",
        queryKey: ["onlinestream-episode-sources", mId, episodeNumber, dubbed],
        data: {
            mediaId: Number(mId),
            episodeNumber,
            dubbed,
            provider: Provider.ZORO,
        },
        enabled: !!mId,
    })

    return {
        sources: data,
        isLoading,
        isFetching,
    }
}
