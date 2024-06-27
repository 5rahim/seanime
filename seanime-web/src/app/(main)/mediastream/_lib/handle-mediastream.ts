import { Anime_MediaEntryEpisode, Mediastream_StreamType } from "@/api/generated/types"
import {
    useGetMediastreamSettings,
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
import { isMobile } from "@/lib/utils/browser-detection"
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
            maxTimeToFirstByteMs: Number.POSITIVE_INFINITY,
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
        autoStartLoad: true,
        abrEwmaDefaultEstimate: 35_000_000,
        abrEwmaDefaultEstimateMax: 50_000_000,
        // debug: true,
        lowLatencyMode: false,
        fragLoadPolicy: {
            default: {
                maxTimeToFirstByteMs: Number.POSITIVE_INFINITY,
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
}

export const __mediastream_progressItemAtom = atom<ProgressItem | undefined>(undefined)

export const __mediastream_currentProgressAtom = atom(0)

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

    const { data: mediastreamSettings, isFetching: mediastreamSettingsLoading } = useGetMediastreamSettings(true)

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
        path: filePath,
        streamType: streamType,
        clientId: sessionId,
    }, !!mediastreamSettings && !mediastreamSettingsLoading)

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


    const isCodecSupported = React.useCallback((codec: string) => {
        if (isMobile()) return false
        if (navigator.userAgent.search("Firefox") === -1)
            codec = codec.replace("video/x-matroska", "video/mp4")
        const videos = document.getElementsByTagName("video")
        const video = videos.item(0) ?? document.createElement("video")
        return video.canPlayType(codec) === "probably"
    }, [])

    /**
     * Effect triggered when media container is available
     * - Check compatibility
     * - Set URL and stream type when media container is available
     */
    React.useEffect(() => {


        /**
         * Check if codec is supported, if it is, switch to direct play
         */
        const codecSupported = isCodecSupported(mediaContainer?.mediaInfo?.mimeCodec ?? "")
        // If the codec is supported, switch to direct play
        if (mediaContainer?.streamType === "transcode") {
            if (codecSupported && !mediastreamSettings?.disableAutoSwitchToDirectPlay) {
                logger("MEDIASTREAM").info("Codec supported", mediaContainer?.mediaInfo?.mimeCodec)
                logger("MEDIASTREAM").warning("Switching to direct play")
                setStreamType("direct")
                changeUrl(undefined)
                return
            } else {
                logger("MEDIASTREAM").info("Codec not supported for direct play", mediaContainer?.mediaInfo?.mimeCodec)
            }
        }
        // If the codec is not supported, switch to transcode
        if (mediaContainer?.streamType === "direct") {
            if (!codecSupported) {
                logger("MEDIASTREAM").warning("Codec not supported for direct play", mediaContainer?.mediaInfo?.mimeCodec)
                logger("MEDIASTREAM").warning("Switching to transcode")
                setStreamType("transcode")
                changeUrl(undefined)
                return
            }
        }

        if (mediaContainer?.streamUrl) {
            logger("MEDIASTREAM").info("Media container", mediaContainer)

            const _newUrl = typeof window !== "undefined" ? (`${window?.location?.protocol}//` + (process.env.NODE_ENV === "development"
                ? `${window?.location?.hostname}:${__DEV_SERVER_PORT}`
                : window?.location?.host) + mediaContainer.streamUrl) : undefined

            logger("MEDIASTREAM").info("Received new stream URL", _newUrl, "streamType:", mediaContainer.streamType)

            changeUrl(_newUrl)
        } else {
            changeUrl(undefined)
        }

    }, [mediaContainer?.streamUrl, mediastreamSettings?.disableAutoSwitchToDirectPlay])

    /**
     * Effect used to set LibASS renderer
     * Add subtitle renderer
     */
    React.useEffect(() => {
        if (playerRef.current && !!mediaContainer?.mediaInfo?.fonts) {
            const legacyWasmUrl = process.env.NODE_ENV === "development"
                ? "/jassub/jassub-worker.wasm.js" : getAssetUrl("/jassub/jassub-worker.wasm.js")

            logger("MEDIASTREAM").info("Loading JASSUB renderer")

            const fonts = mediaContainer?.mediaInfo?.fonts?.map(name => `${window?.location?.protocol}//` + (process.env.NODE_ENV === "development"
                ? `${window?.location?.hostname}:${__DEV_SERVER_PORT}`
                : window?.location?.host) + `/api/v1/mediastream/att/${name}`) || []

            // Extracted fonts
            let availableFonts: Record<string, string> = {}
            let firstFont = ""
            if (!!fonts?.length) {
                for (const font of fonts) {
                    const name = font.split("/").pop()?.split(".")[0]
                    if (name) {
                        if (!firstFont) {
                            firstFont = name.toLowerCase()
                        }
                        availableFonts[name.toLowerCase()] = font
                    }
                }
            }

            // Fallback font if no fonts are available
            if (!firstFont) {
                firstFont = "liberation sans"
            }
            if (Object.keys(availableFonts).length === 0) {
                availableFonts = {
                    "liberation sans": `${window?.location?.protocol}//` + (process.env.NODE_ENV === "development"
                        ? `${window?.location?.hostname}:${__DEV_SERVER_PORT}`
                        : window?.location?.host) + `/jassub/default.woff2`,
                }
            }

            logger("MEDIASTREAM").info("Available fonts:", availableFonts)
            logger("MEDIASTREAM").info("Fallback font:", firstFont)


            // @ts-expect-error
            const renderer = new LibASSTextRenderer(() => import("jassub"), {
                wasmUrl: "/jassub/jassub-worker.wasm",
                workerUrl: "/jassub/jassub-worker.js",
                legacyWasmUrl: legacyWasmUrl,
                // Both parameters needed for subs to work on iOS, ref: jellyfin-vue
                offscreenRender: false,
                prescaleFactor: 0.8,
                onDemandRender: false,
                fonts: fonts,
                availableFonts: availableFonts,
                fallbackFont: firstFont,
            })
            playerRef.current!.textRenderers.add(renderer)

            logger("MEDIASTREAM").info("JASSUB renderer added to player")

            return () => {
                playerRef.current!.textRenderers.remove(renderer)
            }
        }
    }, [playerRef.current, mediaContainer?.streamUrl, mediaContainer?.mediaInfo?.fonts])

    function changeUrl(newUrl: string | undefined) {
        logger("MEDIASTREAM").info("Changing URL", "newURL:", newUrl)
        if (prevUrlRef.current !== newUrl) {
            logger("MEDIASTREAM").info("Resetting playback error status")
            setPlaybackErrored(false)
        }
        setUrl(prevUrl => {
            if (prevUrl === newUrl) {
                logger("MEDIASTREAM").info("URL has not changed")
                return prevUrl
            }
            prevUrlRef.current = prevUrl
            logger("MEDIASTREAM").info("URL changed")
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
        logger("MEDIASTREAM").info("Provider changed", provider, nativeEvent)
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

                    provider.instance?.on(HLS.Events.MANIFEST_PARSED, function (event, data) {
                        logger("MEDIASTREAM").info("onManifestParsed", "attaching media")
                    })

                    provider.instance?.on(HLS.Events.MEDIA_ATTACHED, (event) => {
                        logger("MEDIASTREAM").info("onMediaAttached")
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
                            logger("MEDIASTREAM").error("handleFatalError", data)
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
                    logger("MEDIASTREAM").info("HLS not supported, using native HLS")
                    provider.video.src = url
                } else {
                    logger("MEDIASTREAM").info("HLS not supported, using native HLS")
                    provider.video.src = url
                }
            } else {
                logger("MEDIASTREAM").error("Provider setup - no URL")
            }
        } else {
            logger("MEDIASTREAM").info("Provider setup - not HLS")
        }
    }

    const preloadedNextFileForRef = React.useRef<string | undefined>(undefined)

    const [progressItem, setProgressItem] = useAtom(__mediastream_progressItemAtom)

    const episode = React.useMemo(() => episodes.find(ep => !!ep.localFile?.path && ep.localFile?.path === filePath), [episodes, filePath])

    const onTimeUpdate = React.useCallback((e: MediaTimeUpdateEventDetail) => {
        // DEVNOTE: Disable preloading next file, it causes issues
        // if (!!filePath && duration > 0 && (e.currentTime / duration) > 0.7 && preloadedNextFileForRef.current !== filePath) {
        //     const currentEpisodeIndex = episodes.findIndex(ep => !!ep.localFile?.path && ep.localFile?.path === filePath)
        //     const nextFile = currentEpisodeIndex !== -1 ? episodes[currentEpisodeIndex + 1] : undefined
        //     if (nextFile?.localFile?.path && nextFile?.localFile?.path !== preloadedNextFileForRef.current) {
        //         logger("MEDIASTREAM").info("Preloading next file")
        //         preloadedNextFileForRef.current = filePath
        //         preloadMediaContainer({ path: nextFile?.localFile?.path, streamType: streamType, audioStreamIndex: 0 })
        //     }
        // }
        if (
            (!progressItem || (!!episode?.progressNumber && episode?.progressNumber > progressItem.episodeNumber)) &&
            duration > 0 && (e.currentTime / duration) > 0.8
        ) {
            if (episode) {
                setProgressItem({
                    episodeNumber: episode.progressNumber,
                })
            }
        }
    }, [duration, filePath, episodes, episode, progressItem])

    const onCanPlay = React.useCallback((e: MediaCanPlayDetail) => {
        logger("MEDIASTREAM").info("Can play event received", e)
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
        const baseUri = typeof window !== "undefined" ? (`${window?.location?.protocol}//` + (process.env.NODE_ENV === "development"
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
        disabledAutoSwitchToDirectPlay: mediastreamSettings?.disableAutoSwitchToDirectPlay,

        setStreamType: (type: Mediastream_StreamType) => {
            setStreamType(type)
            changeUrl(undefined)
        },

        onTimeUpdate,
        onCanPlay,
        onEnded,
        onProviderChange,
        onProviderSetup,
        isCodecSupported,
    }

}
