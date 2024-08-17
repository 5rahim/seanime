import { ExtensionRepo_OnlinestreamProviderExtensionItem, Onlinestream_EpisodeSource } from "@/api/generated/types"
import { useGetOnlineStreamEpisodeList, useGetOnlineStreamEpisodeSource } from "@/api/hooks/onlinestream.hooks"
import { useHandleOnlinestreamProviderExtensions } from "@/app/(main)/onlinestream/_lib/handle-onlinestream-providers"
import {
    __onlinestream_autoPlayAtom,
    __onlinestream_qualityAtom,
    __onlinestream_selectedDubbedAtom,
    __onlinestream_selectedEpisodeNumberAtom,
    __onlinestream_selectedProviderAtom,
    __onlinestream_selectedServerAtom,
} from "@/app/(main)/onlinestream/_lib/onlinestream.atoms"
import { logger } from "@/lib/helpers/debug"
import { MediaPlayerInstance } from "@vidstack/react"
import { useAtom, useAtomValue, useSetAtom } from "jotai/react"
import { uniq } from "lodash"
import { useRouter } from "next/navigation"
import React from "react"
import { toast } from "sonner"

export function useOnlinestreamEpisodeList(mId: string | null) {
    const router = useRouter()
    const provider = useAtomValue(__onlinestream_selectedProviderAtom)
    const dubbed = useAtomValue(__onlinestream_selectedDubbedAtom)

    const { data, isLoading, isFetching, isSuccess, isError } = useGetOnlineStreamEpisodeList(mId, provider, dubbed)

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


export function useOnlinestreamEpisodeSource(extensions: ExtensionRepo_OnlinestreamProviderExtensionItem[], mId: string | null, isSuccess: boolean) {

    const provider = useAtomValue(__onlinestream_selectedProviderAtom)
    const episodeNumber = useAtomValue(__onlinestream_selectedEpisodeNumberAtom)
    const dubbed = useAtomValue(__onlinestream_selectedDubbedAtom)

    const extension = React.useMemo(() => extensions.find(p => p.id === provider), [extensions, provider])

    const { data, isLoading, isFetching, isError } = useGetOnlineStreamEpisodeSource(
        mId,
        provider,
        episodeNumber,
        (!!extension?.supportsDub) && dubbed,
        !!mId && episodeNumber !== undefined && isSuccess,
    )

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
        if (!episodeSource || !episodeSource.videoSources) return undefined

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


type HandleOnlinestreamProps = {
    mediaId: string | null
    ref: React.RefObject<MediaPlayerInstance>
}

export function useHandleOnlinestream(props: HandleOnlinestreamProps) {

    const { mediaId, ref: playerRef } = props

    const { providerExtensions, providerExtensionOptions } = useHandleOnlinestreamProviderExtensions()

    /**
     * 1. Get the list of episodes
     */
    const { episodes, media, isFetching, isLoading, isSuccess, isError } = useOnlinestreamEpisodeList(mediaId)

    /**
     * 1. Get the current episode source
     */
    const {
        episodeSource,
        isLoading: isLoadingEpisodeSource,
        isFetching: isFetchingEpisodeSource,
        isError: isErrorEpisodeSource,
    } = useOnlinestreamEpisodeSource(providerExtensions, mediaId, isSuccess)

    /**
     * Variables used for episode source query
     */
    const setEpisodeNumber = useSetAtom(__onlinestream_selectedEpisodeNumberAtom)
    const setServer = useSetAtom(__onlinestream_selectedServerAtom)
    const setQuality = useSetAtom(__onlinestream_qualityAtom)
    const setDubbed = useSetAtom(__onlinestream_selectedDubbedAtom)
    const [provider, setProvider] = useAtom(__onlinestream_selectedProviderAtom)

    const autoPlay = useAtomValue(__onlinestream_autoPlayAtom)
    const [url, setUrl] = React.useState<string | undefined>(undefined)

    // Refs
    const currentProviderRef = React.useRef<string | null>(null)
    const previousCurrentTimeRef = React.useRef(0)
    const previousIsPlayingRef = React.useRef(false)

    // Get current episode details when [episodes] or [episodeSource] changes
    const episodeDetails = React.useMemo(() => {
        return episodes?.find((episode) => episode.number === episodeSource?.number)
    }, [episodes, episodeSource])

    // Get the list of servers
    const servers = React.useMemo(() => {
        if (!episodeSource) return []
        return uniq(episodeSource.videoSources?.map((source) => source.server))
    }, [episodeSource])

    /**
     * Keep episodeSource number in sync with the episode number
     */
    // React.useEffect(() => {
    //     logger("ONLINESTREAM").info("Episode source has changed", { episodeSource })
    //     if (episodeSource) {
    //         setEpisodeNumber(episodeSource.number)
    //     }
    // }, [episodeSource])

    /**
     * 2. Get the current video source
     * This handles selecting the best source
     */
    const { videoSource } = useOnlinestreamVideoSource(episodeSource)

    /**
     * 3. Change the stream URL when the video source changes
     */
    React.useEffect(() => {
        logger("ONLINESTREAM").info("Changing stream URL using videoSource", { videoSource })
        setUrl(undefined)
        if (videoSource?.url) {
            setServer(videoSource.server)
            React.startTransition(() => {
                setUrl(videoSource?.url)
            })
        }
    }, [videoSource?.url])

    // When the provider changes, set the currentProviderRef
    React.useEffect(() => {
        logger("ONLINESTREAM").info("Provider changed", { provider })
        currentProviderRef.current = provider
    }, [provider])

    React.useEffect(() => {
        logger("ONLINESTREAM").info("URL changed", { url })
    }, [url])


    //////////////////////////////////////////////////////////////
    // Video player
    //////////////////////////////////////////////////////////////

    // Store the errored servers, so we can switch to the next server
    const [erroredServers, setErroredServers] = React.useState<string[]>([])
    // Clear errored servers when the episode details change
    React.useEffect(() => {
        setErroredServers([])
    }, [episodeDetails])
    // When the media is detached
    const onMediaDetached = React.useCallback(() => {
        logger("ONLINESTREAM").warning("onMediaDetached")
    }, [])

    /**
     * Handle fatal errors
     * This function is called when the player encounters a fatal error
     * - Change the server if the server is errored
     * - Change the provider if all servers are errored
     */
    const onFatalError = React.useCallback(() => {
        logger("ONLINESTREAM").error("onFatalError", {
            sameProvider: provider == currentProviderRef.current,
        })
        if (provider == currentProviderRef.current) {
            setUrl(undefined)
            toast.info("Playback error, changing server")
            setTimeout(() => {
                logger("ONLINESTREAM").error("erroredServers", erroredServers)
                if (videoSource?.server) {
                    const otherServers = servers.filter((server) => server !== videoSource?.server && !erroredServers.includes(server))
                    if (otherServers.length > 0) {
                        setErroredServers((prev) => [...prev, videoSource?.server])
                        setServer(otherServers[0])
                    } else {
                        setProvider((prev) => providerExtensionOptions.find((p) => p.value !== prev)?.value ?? null)
                    }
                }
            }, 500)
        }
    }, [provider, videoSource, providerExtensionOptions])

    /**
     * Handle provider setup
     */
    const onProviderSetup = React.useCallback(() => {

    }, [provider, videoSource, autoPlay])

    const onCanPlay = React.useCallback(() => {
        logger("ONLINESTREAM").info("Can play event", {
            previousCurrentTime: previousCurrentTimeRef.current,
            previousIsPlayingRef: previousIsPlayingRef.current,
        })
        // When the onCanPlay event is received
        // Restore the time if set
        if (previousCurrentTimeRef.current > 0) {
            Object.assign(playerRef.current ?? {}, { currentTime: previousCurrentTimeRef.current })
            // Resume playing if it was playing
            previousCurrentTimeRef.current = 0
        }
        setTimeout(() => {
            if (previousIsPlayingRef.current) {
                playerRef.current?.play()
            }
        }, 500)
    }, [])


    // Quality
    const [hasCustomQualities, customQualities] = React.useMemo(() => {
        return [
            !!episodeSource?.videoSources?.map(n => n.quality)?.filter(q => q.includes("p"))?.length,
            uniq(episodeSource?.videoSources?.map(n => n.quality)),
        ]
    }, [episodeSource])

    const changeQuality = React.useCallback((quality: string) => {
        try {
            previousCurrentTimeRef.current = playerRef.current?.currentTime ?? 0
            previousIsPlayingRef.current = playerRef.current?.paused === false
        }
        catch {
        }
        setQuality(quality)
    }, [videoSource])

    // Provider
    const changeProvider = React.useCallback((provider: string) => {
        try {
            previousCurrentTimeRef.current = playerRef.current?.currentTime ?? 0
            previousIsPlayingRef.current = playerRef.current?.paused === false
        }
        catch {
        }
        setProvider(provider)
    }, [videoSource])

    // Server
    const changeServer = React.useCallback((server: string) => {
        try {
            previousCurrentTimeRef.current = playerRef.current?.currentTime ?? 0
            previousIsPlayingRef.current = playerRef.current?.paused === false
        }
        catch {
        }
        setServer(server)
    }, [videoSource])

    // Dubbed
    const toggleDubbed = React.useCallback(() => {
        try {
            previousCurrentTimeRef.current = playerRef.current?.currentTime ?? 0
            previousIsPlayingRef.current = playerRef.current?.paused === false
        }
        catch {
        }
        setDubbed((prev) => !prev)
    }, [videoSource])

    // Episode
    const handleChangeEpisodeNumber = (epNumber: number) => {
        setEpisodeNumber(_ => {
            return epNumber
        })
    }

    const selectedExtension = React.useMemo(() => providerExtensions.find(p => p.id === provider), [providerExtensions, provider])

    return {
        currentEpisodeDetails: episodeDetails,
        provider,
        servers,
        videoSource,
        onMediaDetached,
        onProviderSetup,
        onFatalError,
        onCanPlay,
        url,
        episodes,
        media: media!,
        episodeSource,
        loadPage: !isFetching && !isLoading,
        currentEpisodeNumber: episodeSource?.number ?? 0,
        handleChangeEpisodeNumber,
        episodeLoading: isLoadingEpisodeSource || isFetchingEpisodeSource,
        isErrorEpisodeSource,
        isErrorProvider: isError,
        opts: {
            selectedExtension,
            currentEpisodeDetails: episodeDetails,
            providerExtensions,
            providerExtensionOptions,
            servers,
            videoSource,
            customQualities,
            hasCustomQualities,
            changeQuality,
            changeProvider,
            changeServer,
            toggleDubbed,
        },
    }

}

export type OnlinestreamManagerOpts = ReturnType<typeof useHandleOnlinestream>
