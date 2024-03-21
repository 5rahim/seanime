import { BaseMediaFragment } from "@/lib/anilist/gql/graphql"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaQuery } from "@/lib/server/query"
import { atom } from "jotai"
import { useAtomValue } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import { useRouter } from "next/navigation"
import React from "react"

export type OnlinestreamEpisode = {
    number: number
    title?: string
    description?: string
    image?: string
}

export type OnlinestreamEpisodeListResponse = {
    episodes: OnlinestreamEpisode[]
    media: BaseMediaFragment
}

export type OnlinestreamEpisodeSource = {
    number: number
    videoSources: OnlinestreamVideoSource[]
    subtitles: OnlinestreamVideoSubtitles[] | undefined
}

export type OnlinestreamVideoSource = {
    headers: Record<string, string>
    server: string
    url: string
    quality: string
}

export type OnlinestreamVideoSubtitles = {
    url: string
    language: string
}

const enum Provider {
    GOGOANIME = "gogoanime",
    ZORO = "zoro",
}

export const onlinestream_providers = [
    { value: "gogoanime", label: "Gogoanime" },
    { value: "zoro", label: "Hianime" },
]

export const __onlinestream_mediaIdAtom = atom<string | null>(null)
export const __onlinestream_selectedProviderAtom = atomWithStorage<string>("sea-onlinestream-provider", Provider.GOGOANIME)
export const __onlinestream_selectedDubbedAtom = atom<boolean>(false)
export const __onlinestream_selectedEpisodeNumberAtom = atom<number | undefined>(undefined)

export const __onlinestream_autoPlayAtom = atomWithStorage("sea-onlinestream-autoplay", false)
export const __onlinestream_autoNextAtom = atomWithStorage("sea-onlinestream-autonext", false)


export function useOnlinestreamEpisodeList(mId: string | null) {
    const router = useRouter()
    const provider = useAtomValue(__onlinestream_selectedProviderAtom)
    const dubbed = useAtomValue(__onlinestream_selectedDubbedAtom)

    const { data, isLoading, isFetching, isSuccess, isError } = useSeaQuery<OnlinestreamEpisodeListResponse, {
        mediaId: number,
        provider: string,
        dubbed: boolean
    }>({
        endpoint: SeaEndpoints.ONLINESTREAM_EPISODE_LIST,
        method: "post",
        queryKey: ["onlinestream-episode-list", mId, provider, dubbed],
        data: {
            mediaId: Number(mId),
            dubbed,
            provider: provider,
        },
        enabled: !!mId,
    })

    React.useEffect(() => {
        if (isError) {
            router.push("/")
        }
    }, [isError])

    return {
        media: data?.media,
        episodes: data?.episodes,
        isLoading,
        isFetching,
        isSuccess,
    }
}


export function useOnlinestreamEpisodeSource(mId: string | null, isSuccess: boolean) {

    const provider = useAtomValue(__onlinestream_selectedProviderAtom)
    const episodeNumber = useAtomValue(__onlinestream_selectedEpisodeNumberAtom)
    const dubbed = useAtomValue(__onlinestream_selectedDubbedAtom)

    const { data, isLoading, isFetching } = useSeaQuery<OnlinestreamEpisodeSource, {
        mediaId: number,
        episodeNumber: number,
        provider: string,
        dubbed: boolean
    }>({
        endpoint: SeaEndpoints.ONLINESTREAM_EPISODE_SOURCE,
        method: "post",
        queryKey: ["onlinestream-episode-source", mId, provider, episodeNumber, dubbed],
        data: {
            mediaId: Number(mId),
            episodeNumber: episodeNumber!,
            dubbed,
            provider: provider,
        },
        enabled: !!mId && episodeNumber !== undefined && isSuccess,
    })

    return {
        episodeSource: data,
        isLoading,
        isFetching,
    }
}


export const __onlinestream_selectedServerAtom = atomWithStorage<string | undefined>("sea-onlinestream-server", undefined)
export const __onlinestream_qualityAtom = atomWithStorage<string | undefined>("sea-onlinestream-quality", undefined)
export const onlinestream_qualityOptions = ["360p", "480p", "720p", "1080p", "auto"]

export function useOnlinestreamVideoSource(episodeSource: OnlinestreamEpisodeSource | undefined) {

    const quality = useAtomValue(__onlinestream_qualityAtom)
    const selectedServer = useAtomValue(__onlinestream_selectedServerAtom)

    const videoSource = React.useMemo(() => {
        if (!episodeSource) return undefined

        let videoSources = episodeSource.videoSources

        if (selectedServer && videoSources.some(n => n.server === selectedServer)) {
            videoSources = videoSources.filter(s => s.server === selectedServer)
        }

        const hasQuality = videoSources.some(n => n.quality === quality)
        const hasAuto = videoSources.some(n => n.quality === "auto")

        // If quality is set, filter sources by quality
        // Only filter by quality if the quality is present in the sources
        if (quality && hasQuality) {
            videoSources = videoSources.filter(s => s.quality === quality)
        } else if (quality && !hasAuto) {
            if (videoSources.some(n => n.quality === "1080p")) {
                videoSources = videoSources.filter(s => s.quality === "1080p")
            } else if (videoSources.some(n => n.quality === "default")) {
                videoSources = videoSources.filter(s => s.quality === "default")
            } else if (videoSources.some(n => n.quality === "720p")) {
                videoSources = videoSources.filter(s => s.quality === "720p")
            } else if (videoSources.some(n => n.quality === "480p")) {
                videoSources = videoSources.filter(s => s.quality === "480p")
            } else if (videoSources.some(n => n.quality === "360p")) {
                videoSources = videoSources.filter(s => s.quality === "360p")
            }
        } else if (quality && hasAuto) {
            videoSources = videoSources.filter(s => s.quality === "auto")
        }

        return videoSources[0]
    }, [episodeSource, selectedServer, quality])

    return {
        videoSource,
    }
}
