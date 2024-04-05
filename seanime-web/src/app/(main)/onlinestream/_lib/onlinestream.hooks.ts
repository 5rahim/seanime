import {
    __onlinestream_qualityAtom,
    __onlinestream_selectedDubbedAtom,
    __onlinestream_selectedEpisodeNumberAtom,
    __onlinestream_selectedProviderAtom,
    __onlinestream_selectedServerAtom,
} from "@/app/(main)/onlinestream/_lib/onlinestream.atoms"
import { Onlinestream_EpisodeListResponse, Onlinestream_EpisodeSource } from "@/app/(main)/onlinestream/_lib/onlinestream.types"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaQuery } from "@/lib/server/query"
import { useAtomValue } from "jotai/react"
import { useRouter } from "next/navigation"
import React from "react"

export const ONLINESTREAM_PROVIDERS = [
    { value: "gogoanime", label: "Gogoanime" },
    { value: "zoro", label: "Hianime" },
]

export function useOnlinestreamEpisodeList(mId: string | null) {
    const router = useRouter()
    const provider = useAtomValue(__onlinestream_selectedProviderAtom)
    const dubbed = useAtomValue(__onlinestream_selectedDubbedAtom)

    const { data, isLoading, isFetching, isSuccess, isError } = useSeaQuery<Onlinestream_EpisodeListResponse, {
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
        isError,
    }
}


export function useOnlinestreamEpisodeSource(mId: string | null, isSuccess: boolean) {

    const provider = useAtomValue(__onlinestream_selectedProviderAtom)
    const episodeNumber = useAtomValue(__onlinestream_selectedEpisodeNumberAtom)
    const dubbed = useAtomValue(__onlinestream_selectedDubbedAtom)

    const { data, isLoading, isFetching, isError } = useSeaQuery<Onlinestream_EpisodeSource, {
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
        isError,
    }
}


export function useOnlinestreamVideoSource(episodeSource: Onlinestream_EpisodeSource | undefined) {

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
