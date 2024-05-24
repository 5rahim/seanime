import { Anime_MediaEntryEpisode, Mediastream_StreamType } from "@/api/generated/types"
import {
    useMediastreamShutdownTranscodeStream,
    usePreloadMediastreamMediaContainer,
    useRequestMediastreamMediaContainer,
} from "@/api/hooks/mediastream.hooks"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { useMediastreamCurrentFile } from "@/app/(main)/mediastream/_lib/mediastream.atoms"
import { logger } from "@/lib/helpers/debug"
import { getAssetUrl } from "@/lib/server/assets"
import { __DEV_SERVER_PORT } from "@/lib/server/config"
import { WSEvents } from "@/lib/server/ws-events"
import {
    isHLSProvider,
    LibASSTextRenderer,
    MediaCanPlayDetail,
    MediaEndedEvent,
    MediaPlayerInstance,
    MediaProviderAdapter,
    MediaProviderChangeEvent,
    MediaProviderSetupEvent,
    MediaTimeUpdateEventDetail,
} from "@vidstack/react"
import Hls from "hls.js"
import HLS, { LoadPolicy } from "hls.js"
import { atom } from "jotai/index"
import { useAtom } from "jotai/react"
import { useRouter } from "next/navigation"
import React from "react"
import { toast } from "sonner"

function uuidv4(): string {
    // @ts-ignore
    return ([1e7] + -1e3 + -4e3 + -8e3 + -1e11).replace(/[018]/g, (c) =>
        (c ^ (crypto.getRandomValues(new Uint8Array(1))[0] & (15 >> (c / 4)))).toString(16),
    )
}

let cId = typeof window === "undefined" ? "-" : uuidv4()

const mediastream_getHlsConfig = () => {
    const loadPolicy: LoadPolicy = {
        default: {
            maxTimeToFirstByteMs: Infinity,
            maxLoadTimeMs: 60_000,
            timeoutRetry: {
                maxNumRetry: 2,
                retryDelayMs: 0,
                maxRetryDelayMs: 0,
            },
            errorRetry: {
                maxNumRetry: 1,
                retryDelayMs: 0,
                maxRetryDelayMs: 0,
            },
        },
    }
    return {
        startLevel: Infinity,
        autoStartLoad: true,
        // autoStartLoad: false,
        abrEwmaDefaultEstimate: 35_000_000,
        abrEwmaDefaultEstimateMax: 50_000_000,
        // debug: true,
        lowLatencyMode: false,
        fragLoadPolicy: {
            default: {
                maxTimeToFirstByteMs: Infinity,
                maxLoadTimeMs: 60_000,
                timeoutRetry: {
                    // maxNumRetry: 15,
                    maxNumRetry: 5,
                    retryDelayMs: 100,
                    maxRetryDelayMs: 0,
                },
                errorRetry: {
                    // maxNumRetry: 5,
                    // retryDelayMs: 0,
                    maxNumRetry: 15,
                    retryDelayMs: 100,
                    maxRetryDelayMs: 100,
                },
            },
        },
        keyLoadPolicy: loadPolicy,
        certLoadPolicy: loadPolicy,
        playlistLoadPolicy: loadPolicy,
        manifestLoadPolicy: loadPolicy,
        steeringManifestLoadPolicy: loadPolicy,
    }
}

type ProgressItem = {
    episodeNumber: number
    updated: boolean
}

export const __mediastream_progressItemAtom = atom<ProgressItem | undefined>(undefined)

type HandleMediastreamProps = {
    playerRef: React.RefObject<MediaPlayerInstance>
    episodes: Anime_MediaEntryEpisode[]
}

export function useHandleMediastream(props: HandleMediastreamProps) {

    const {
        playerRef,
        episodes,
    } = props
    const router = useRouter()
    const { filePath, setFilePath } = useMediastreamCurrentFile()

    /**
     * Stream URL
     */
    const prevUrlRef = React.useRef<string | undefined>(undefined)
    const definedUrlRef = React.useRef<string | undefined>(undefined)
    const [url, setUrl] = React.useState<string | undefined>(undefined)
    const [streamType, setStreamType] = React.useState<Mediastream_StreamType>("transcode") // do not chance

    const [sessionId, setSessionId] = React.useState<string>(uuidv4())

    /**
     * Fetch media container containing stream URL
     */
    const { data: _mediaContainer, isError: isMediaContainerError, isPending, isFetching, refetch } = useRequestMediastreamMediaContainer({
        // path: filePath,
        streamType: streamType,
        clientId: sessionId,
    })

    const mediaContainer = React.useMemo(() => (!isPending && !isFetching) ? _mediaContainer : undefined, [_mediaContainer, isPending, isFetching])

    /**
     * Preload next file
     */
    const { mutate: preloadMediaContainer } = usePreloadMediastreamMediaContainer()
    // const [preloadedFilePath, setPreloadedFilePath] = React.useState<string | undefined>(undefined)


    // Whether the playback has errored
    const [playbackErrored, setPlaybackErrored] = React.useState<boolean>(false)

    // Duration
    const [duration, setDuration] = React.useState<number>(0)

    // useUpdateEffect(() => {
    //     if (!filePath?.length) {
    //         toast.error("No file path provided")
    //         router.push("/")
    //     }
    // }, [filePath])

    React.useEffect(() => {
        if (isPending) {
            logger("MEDIASTREAM").info("Loading media container")
            changeUrl(undefined)
        }
    }, [isPending])

    const { mutate: shutdownTranscode } = useMediastreamShutdownTranscodeStream()

    /**
     * This error is thrown when the media container is available but the URL has been set to undefined
     * - This is usually when the transcoder has errored out
     */
    const isStreamError = !!mediaContainer && !url

    /**
     * Set URL and stream type when media container is available
     */
    React.useEffect(() => {

        if (mediaContainer?.streamUrl) {
            logger("MEDIASTREAM").info("Media container", mediaContainer)

            const _newUrl = typeof window !== "undefined" ? ("http://" + (process.env.NODE_ENV === "development"
                ? `${window?.location?.hostname}:${__DEV_SERVER_PORT}`
                : window?.location?.host) + mediaContainer.streamUrl) : undefined

            logger("MEDIASTREAM").info("Setting stream URL available", _newUrl, mediaContainer.streamType)

            changeUrl(_newUrl)
        } else {
            changeUrl(undefined)
        }

    }, [mediaContainer?.streamUrl])

    /**
     * Add subtitle renderer
     */
    React.useEffect(() => {
        if (playerRef.current) {
            // @ts-ignore
            const renderer = new LibASSTextRenderer(() => import("jassub"), {
                workerUrl: getAssetUrl("/jassub/jassub-worker.js"),
                legacyWorkerUrl: getAssetUrl("/jassub/jassub-worker-legacy.js"),
            })
            playerRef.current!.textRenderers.add(renderer)

            return () => {
                playerRef.current!.textRenderers.remove(renderer)
            }
        }
    }, [mediaContainer?.streamUrl])

    function changeUrl(newUrl: string | undefined) {
        if (prevUrlRef.current !== newUrl) {
            setPlaybackErrored(false)
        }
        setUrl(prevUrl => {
            if (prevUrl === newUrl) return prevUrl
            prevUrlRef.current = prevUrl
            return newUrl
        })
        if (newUrl) {
            definedUrlRef.current = newUrl
        }
    }

    //////////////////////////////////////////////////////////////
    // Video player
    //////////////////////////////////////////////////////////////

    function onProviderChange(
        provider: MediaProviderAdapter | null,
        nativeEvent: MediaProviderChangeEvent,
    ) {
        logger("MEDIASTREAM").info("Provider change")
        if (isHLSProvider(provider)) {
            provider.library = HLS
            if (mediaContainer?.streamType === "transcode") {
                provider.config = {
                    // xhrSetup: async (xhr) => {
                    //     xhr.setRequestHeader("X-Seanime-Mediastream-Client-Id", cId)
                    // },
                    ...mediastream_getHlsConfig(),
                }
            }
        }
    }

    const previousCurrentTimeRef = React.useRef<number>(0)

    function onProviderSetup(provider: MediaProviderAdapter, nativeEvent: MediaProviderSetupEvent) {
        if (isHLSProvider(provider)) {
            if (url) {

                if (definedUrlRef.current === url && playbackErrored) {
                    if (previousCurrentTimeRef.current > 0) {
                        Object.assign(playerRef.current ?? {}, { currentTime: previousCurrentTimeRef.current })
                        // setTimeout(() => {
                        //     if (previousIsPlayingRef.current) {
                        //         playerRef.current?.play()
                        //     }
                        // }, 500)
                        previousCurrentTimeRef.current = 0
                        setPlaybackErrored(false)
                    }
                }

                if (HLS.isSupported() && url.endsWith(".m3u8")) {

                    logger("MEDIASTREAM").info("HLS Provider setup")

                    logger("MEDIASTREAM").info("Loading source", url)
                    // provider.instance?.loadSource(url)

                    provider.instance?.on(Hls.Events.MANIFEST_PARSED, function (event, data) {
                        logger("MEDIASTREAM").info("onManifestParsed", "attaching media")
                        // provider?.instance?.attachMedia(provider.video)
                    })

                    provider.instance?.on(HLS.Events.MEDIA_ATTACHED, (event) => {
                        logger("MEDIASTREAM").info("onMediaAttached")
                        // provider.instance?.startLoad(0)
                    })

                    provider.instance?.on(HLS.Events.MEDIA_DETACHED, (event) => {
                        logger("MEDIASTREAM").warning("onMediaDetached")
                        // When the media is detached, stop the transcoder but only if there was no playback error
                        if (!playbackErrored) {
                            if (mediaContainer?.streamType === "transcode") {
                                // DEVNOTE: Comment code below kills the transcoder AFTER changing episode due to delay
                                // shutdownTranscode()
                            }
                            changeUrl(undefined)
                        }
                        // refetch()
                    })

                    provider.instance?.on(HLS.Events.FRAG_LOADED, (event, data) => {
                        previousCurrentTimeRef.current = playerRef.current?.currentTime ?? 0
                    })

                    /**
                     * Fatal error
                     */
                    provider.instance?.on(HLS.Events.ERROR, (event, data) => {
                        if (data?.fatal) {
                            // Record current time
                            previousCurrentTimeRef.current = playerRef.current?.currentTime ?? 0
                            logger("MEDIASTREAM").error("handleFatalError")
                            // Shut down transcoder
                            if (mediaContainer?.streamType === "transcode") {
                                shutdownTranscode()
                            }
                            // Set playback errored
                            setPlaybackErrored(true)
                            // Delete URL
                            changeUrl(undefined)
                            toast.error("Playback error")
                            // Refetch media container
                            refetch()
                        }
                    })
                } else if (!HLS.isSupported() && url.endsWith(".m3u8") && provider.video.canPlayType("application/vnd.apple.mpegurl")) {
                    provider.video.src = url
                } else {
                    provider.video.src = url
                }
            } else {
                logger("MEDIASTREAM").error("Provider setup - no URL")
            }
        }
    }

    const preloadedNextFileForRef = React.useRef<string | undefined>(undefined)

    const [progressItem, setProgressItem] = useAtom(__mediastream_progressItemAtom)

    const onTimeUpdate = React.useCallback((e: MediaTimeUpdateEventDetail) => {
        if (!!filePath && duration > 0 && (e.currentTime / duration) > 0.7 && preloadedNextFileForRef.current !== filePath) {
            logger("MEDIASTREAM").info("Preloading next file")

            const currentEpisodeIndex = episodes.findIndex(ep => !!ep.localFile?.path && ep.localFile?.path === filePath)
            const nextFile = currentEpisodeIndex !== -1 ? episodes[currentEpisodeIndex + 1] : undefined
            if (nextFile?.localFile?.path) {
                preloadedNextFileForRef.current = filePath
                preloadMediaContainer({ path: filePath, streamType: streamType, audioStreamIndex: 0 })
            }
        }
        if ((!progressItem || !progressItem.updated) && duration > 0 && (e.currentTime / duration) > 0.8) {
            const episode = episodes.find(ep => !!ep.localFile?.path && ep.localFile?.path === filePath)
            if (episode) {
                setProgressItem({
                    episodeNumber: episode.progressNumber,
                    updated: false,
                })
            }
        }
    }, [duration, filePath, episodes, progressItem])

    const onCanPlay = React.useCallback((e: MediaCanPlayDetail) => {
        preloadedNextFileForRef.current = undefined
        setDuration(e.duration)
    }, [])

    const onEnded = React.useCallback((e: MediaEndedEvent) => {

    }, [])

    const onPlayFile = React.useCallback((filepath: string) => {
        logger("MEDIASTREAM").info("Playing file", filepath)
        setFilePath(filepath)
    }, [])

    //////////////////////////////////////////////////////////////
    // Events
    //////////////////////////////////////////////////////////////

    /**
     * Listen for shutdown stream event
     * - This event is sent when something goes wrong internally
     * - Settings the URL to undefined will unmount the player and thus avoid spamming the server
     */
    useWebsocketMessageListener<string | null>({
        type: WSEvents.MEDIASTREAM_SHUTDOWN_STREAM,
        onMessage: log => {
            if (log) {
                toast.error(log)
            }
            logger("MEDIASTREAM").warning("Shutdown stream event received")
            changeUrl(undefined)
        },
    })

    //////////////////////////////////////////////////////////////

    // Subtitle endpoint URI
    const subtitleEndpointUri = React.useMemo(() => {
        const baseUri = typeof window !== "undefined" ? ("http://" + (process.env.NODE_ENV === "development"
            ? `${window?.location?.hostname}:${__DEV_SERVER_PORT}`
            : window?.location?.host)) : ""
        if (mediaContainer?.streamUrl && mediaContainer?.streamType) {
            return `${baseUri}/api/v1/mediastream/subs`
        }
        return ""
    }, [mediaContainer?.streamUrl, mediaContainer?.streamType])

    return {
        url,
        streamType,
        subtitles: mediaContainer?.mediaInfo?.subtitles,
        isMediaContainerLoading: isPending,
        isError: isMediaContainerError || isStreamError,
        subtitleEndpointUri,
        mediaContainer: _mediaContainer,
        onPlayFile,
        filePath,

        onTimeUpdate,
        onCanPlay,
        onEnded,
        onProviderChange,
        onProviderSetup,
    }

}
