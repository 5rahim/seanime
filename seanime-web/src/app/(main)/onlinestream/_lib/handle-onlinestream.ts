import { getServerBaseUrl } from "@/api/client/server-url"
import { ExtensionRepo_OnlinestreamProviderExtensionItem, Onlinestream_EpisodeSource } from "@/api/generated/types"
import { useHandleCurrentMediaContinuity } from "@/api/hooks/continuity.hooks"
import { useGetOnlineStreamEpisodeList, useGetOnlineStreamEpisodeSource } from "@/api/hooks/onlinestream.hooks"
import { useNakamaStatus } from "@/app/(main)/_features/nakama/nakama-manager"
import { useWebsocketSender } from "@/app/(main)/_hooks/handle-websockets"
import { useHandleOnlinestreamProviderExtensions } from "@/app/(main)/onlinestream/_lib/handle-onlinestream-providers"
import {
    __onlinestream_qualityAtom,
    __onlinestream_selectedDubbedAtom,
    __onlinestream_selectedEpisodeNumberAtom,
    __onlinestream_selectedProviderAtom,
    __onlinestream_selectedServerAtom,
} from "@/app/(main)/onlinestream/_lib/onlinestream.atoms"
import { logger } from "@/lib/helpers/debug"
import { MediaPlayerInstance } from "@vidstack/react"
import { atom } from "jotai"
import { useAtom, useAtomValue, useSetAtom } from "jotai/react"
import { uniq } from "lodash"
import { usePathname, useRouter, useSearchParams } from "next/navigation"
import React from "react"
import { useUpdateEffect } from "react-use"
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

        logger("ONLINESTREAM").info("Stored quality", quality)
        logger("ONLINESTREAM").info("Selected server", selectedServer)

        if (selectedServer && videoSources.some(n => n.server === selectedServer)) {
            videoSources = videoSources.filter(s => s.server === selectedServer)
        }

        const hasQuality = videoSources.some(n => n.quality === quality)
        const hasAuto = videoSources.some(n => n.quality === "auto")

        logger("ONLINESTREAM").info("Selecting quality", {
            hasAuto,
            hasQuality,
        })

        // If quality is set, filter sources by quality
        // Only filter by quality if the quality is present in the sources
        if (quality && hasQuality) {
            videoSources = videoSources.filter(s => s.quality === quality)
        } else if (hasAuto) {
            videoSources = videoSources.filter(s => s.quality === "auto")
        } else {

            logger("ONLINESTREAM").info("Choosing a quality")

            if (videoSources.some(n => n.quality.includes("1080p"))) {
                videoSources = videoSources.filter(s => s.quality.includes("1080p"))
            } else if (videoSources.some(n => n.quality.includes("720p"))) {
                videoSources = videoSources.filter(s => s.quality.includes("720p"))
            } else if (videoSources.some(n => n.quality.includes("480p"))) {
                videoSources = videoSources.filter(s => s.quality.includes("480p"))
            } else if (videoSources.some(n => n.quality.includes("360p"))) {
                videoSources = videoSources.filter(s => s.quality.includes("360p"))
            }

            if (videoSources.some(n => n.quality.includes("default"))) {
                videoSources = videoSources.filter(s => s.quality.includes("default"))
            }
        }


        logger("ONLINESTREAM").info("videoSources", videoSources)

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

    // Nakama Watch Party
    const nakamaStatus = useNakamaStatus()
    const { streamToLoad, onLoadedStream, hostNotifyStreamStarted } = useNakamaOnlineStreamWatchParty()

    /**
     * 1. Get the list of episodes
     */
    const { episodes, media, isFetching, isLoading, isSuccess, isError } = useOnlinestreamEpisodeList(mediaId)

    /**
     * 2. Watch history
     */
    const { waitForWatchHistory } = useHandleCurrentMediaContinuity(mediaId)

    /**
     * 3. Get the current episode source
     */
    const {
        episodeSource,
        isLoading: isLoadingEpisodeSource,
        isFetching: isFetchingEpisodeSource,
        isError: isErrorEpisodeSource,
    } = useOnlinestreamEpisodeSource(providerExtensions, mediaId, (isSuccess && !waitForWatchHistory))

    /**
     * Variables used for episode source query
     */
    const setEpisodeNumber = useSetAtom(__onlinestream_selectedEpisodeNumberAtom)
    const setServer = useSetAtom(__onlinestream_selectedServerAtom)
    const setQuality = useSetAtom(__onlinestream_qualityAtom)
    const [dubbed, setDubbed] = useAtom(__onlinestream_selectedDubbedAtom)
    const [provider, setProvider] = useAtom(__onlinestream_selectedProviderAtom)

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
        if (!episodeSource) {
            logger("ONLINESTREAM").info("Updating servers, no episode source", [])
            return []
        }
        const servers = episodeSource.videoSources?.map((source) => source.server)
        logger("ONLINESTREAM").info("Updating servers", servers)
        return uniq(servers)
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
        logger("ONLINESTREAM").info("Setting stream URL to undefined")
        if (videoSource?.url) {
            setServer(videoSource.server)
            let _url = videoSource.url
            if (videoSource.headers && Object.keys(videoSource.headers).length > 0) {
                _url = `${getServerBaseUrl()}/api/v1/proxy?url=${encodeURIComponent(videoSource?.url)}&headers=${encodeURIComponent(JSON.stringify(
                    videoSource?.headers))}`
            } else {
                _url = videoSource.url
            }
            React.startTransition(() => {
                logger("ONLINESTREAM").info("Setting stream URL", { url: _url })
                setUrl(_url)
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

    useUpdateEffect(() => {
        if (!streamToLoad) return

        logger("ONLINESTREAM").info("Stream to load", { streamToLoad })

        // Check if we have the provider
        if (!providerExtensionOptions.some(p => p.value === streamToLoad.provider)) {
            logger("ONLINESTREAM").warning("Provider not found in options", { providerExtensionOptions, provider: streamToLoad.provider })
            toast.error("Watch Party: The provider used by the host is not installed.")
            return
        }

        setProvider(streamToLoad.provider)
        setDubbed(streamToLoad.dubbed)
        setEpisodeNumber(streamToLoad.episodeNumber)
        setServer(streamToLoad.server)
        setQuality(streamToLoad.quality)

        onLoadedStream()
    }, [streamToLoad])


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
    const onFatalError = () => {
        logger("ONLINESTREAM").error("onFatalError", {
            sameProvider: provider == currentProviderRef.current,
        })
        if (provider == currentProviderRef.current) {
            setUrl(undefined)
            logger("ONLINESTREAM").error("Setting stream URL to undefined")
            toast.warning("Playback error, trying another server...")
            logger("ONLINESTREAM").error("Player encountered a fatal error")
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
    }

    /**
     * Handle the onCanPlay event
     */
    const onCanPlay = () => {
        logger("ONLINESTREAM").info("Can play event", {
            previousCurrentTime: previousCurrentTimeRef.current,
            previousIsPlayingRef: previousIsPlayingRef.current,
        })

        // Handle Nakama Watch Party
        // Send the playback info to the server
        if (nakamaStatus?.isHost && nakamaStatus.currentWatchPartySession && !nakamaStatus?.currentWatchPartySession.isRelayMode) {
            const params = {
                mediaId: media?.id ?? 0,
                provider: currentProviderRef.current ?? "",
                episodeNumber: episodeSource?.number ?? 0,
                server: videoSource?.server ?? "",
                quality: videoSource?.quality ?? "",
                dubbed: dubbed,
            }
            logger("ONLINESTREAM").info("Watch Party: Notifying peers that stream has started", params)
            hostNotifyStreamStarted(params)
        }

        // When the onCanPlay event is received
        // Restore the previous time if set
        if (previousCurrentTimeRef.current > 0) {
            // Seek to the previous time
            Object.assign(playerRef.current ?? {}, { currentTime: previousCurrentTimeRef.current })
            // Reset the previous time ref
            previousCurrentTimeRef.current = 0
            logger("ONLINESTREAM").info("Seeking to previous time", { previousCurrentTime: previousCurrentTimeRef.current })
        }

        // If the player was playing before the onCanPlay event, resume playing
        setTimeout(() => {
            if (previousIsPlayingRef.current) {
                try {
                    playerRef.current?.play()
                }
                catch {
                }
                logger("ONLINESTREAM").info("Resuming playback since past video was playing before the onCanPlay event")
            }
        }, 500)
    }


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
            logger("ONLINESTREAM").info("Changing quality", { quality })
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
            logger("ONLINESTREAM").info("Changing provider", { provider })
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
            logger("ONLINESTREAM").info("Changing server", { server })
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
            logger("ONLINESTREAM").info("Toggling dubbed")
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


//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export type OnlineStreamParams = {
    mediaId: number
    provider: string
    episodeNumber: number
    server: string
    quality: string
    dubbed: boolean
}

const __onlinestream_streamToLoadAtom = atom<OnlineStreamParams | null>(null)

export function useNakamaOnlineStreamWatchParty() {
    const [streamToLoad, setStreamToLoad] = useAtom(__onlinestream_streamToLoadAtom)
    const pathname = usePathname()
    const searchParams = useSearchParams()
    const router = useRouter()
    const nakamaStatus = useNakamaStatus()

    const { sendMessage } = useWebsocketSender()


    return {
        // Host
        hostNotifyStreamStarted: (params: OnlineStreamParams) => {
            if (!nakamaStatus?.isHost) {
                logger("ONLINESTREAM").warning("hostNotifyStreamStarted called, but not a host")
                return
            }
            sendMessage({
                type: "nakama",
                payload: {
                    type: "online-stream-started",
                    payload: {
                        mediaId: params.mediaId,
                        provider: params.provider,
                        episodeNumber: params.episodeNumber,
                        server: params.server,
                        quality: params.quality,
                        dubbed: params.dubbed,
                    },
                },
            })
        },
        // Peers
        streamToLoad,
        onLoadedStream: () => {
            setStreamToLoad(null)
        },
        startOnlineStream(params: OnlineStreamParams) {
            if (nakamaStatus?.isHost) {
                logger("ONLINESTREAM").info("Start online stream event sent to peers", params)
                return
            }
            logger("ONLINESTREAM").info("Starting online stream", params)
            setStreamToLoad(params)
            if (pathname !== "/entry" || searchParams.get("id") !== String(params.mediaId)) {
                // Navigate to the onlinestream page
                router.push("/entry?id=" + params.mediaId + "&tab=onlinestream&provider=" + params.provider + "&episodeNumber=" + params.episodeNumber + "&server=" + params.server + "&quality=" + params.quality + "&dubbed=" + params.dubbed)
            }
        },
    }
}
