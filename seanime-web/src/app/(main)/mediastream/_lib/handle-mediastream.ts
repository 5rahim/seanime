import { Mediastream_StreamType } from "@/api/generated/types"
import { useMediastreamShutdownTranscodeStream, useRequestMediastreamMediaContainer } from "@/api/hooks/mediastream.hooks"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { logger } from "@/lib/helpers/debug"
import { getAssetUrl } from "@/lib/server/assets"
import { __DEV_SERVER_PORT } from "@/lib/server/config"
import { WSEvents } from "@/lib/server/ws-events"
import {
    isHLSProvider,
    LibASSTextRenderer,
    MediaPlayerInstance,
    MediaProviderAdapter,
    MediaProviderChangeEvent,
    MediaProviderSetupEvent,
} from "@vidstack/react"
import Hls from "hls.js"
import HLS, { LoadPolicy } from "hls.js"
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
        debug: true,
        lowLatencyMode: false,
        fragLoadPolicy: {
            default: {
                maxTimeToFirstByteMs: Infinity,
                maxLoadTimeMs: 60_000,
                timeoutRetry: {
                    maxNumRetry: 5,
                    retryDelayMs: 100,
                    maxRetryDelayMs: 0,
                },
                errorRetry: {
                    maxNumRetry: 5,
                    retryDelayMs: 0,
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

type HandleMediastreamProps = {
    playerRef: React.RefObject<MediaPlayerInstance>
}

export function useHandleMediastream(props: HandleMediastreamProps) {

    const {
        playerRef,
    } = props

    /**
     * Stream URL
     */
    const prevUrlRef = React.useRef<string | undefined>(undefined)
    const [url, setUrl] = React.useState<string | undefined>(undefined)
    const [streamType, setStreamType] = React.useState<Mediastream_StreamType>("direct")

    const { data: _mediaContainer, isError: isMediaContainerError, isPending, refetch } = useRequestMediastreamMediaContainer({
        path: "E:\\ANIME\\Dungeon Meshi\\[EMBER] Dungeon Meshi - 15.mkv",
        streamType: "transcode",
    })
    const mediaContainer = !isPending ? _mediaContainer : undefined

    React.useEffect(() => {
        if (isPending) {
            logger("MEDIASTREAM").info("Loading media container")
            setUrl(undefined)
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

            setUrl(prevUrl => {
                if (prevUrl === _newUrl) return prevUrl
                prevUrlRef.current = prevUrl
                return _newUrl
            })
        } else {
            setUrl(undefined)
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

    function onProviderSetup(provider: MediaProviderAdapter, nativeEvent: MediaProviderSetupEvent) {
        if (isHLSProvider(provider)) {
            if (url) {
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
                        if (mediaContainer?.streamType === "transcode") {
                            shutdownTranscode()
                        }
                        setUrl(undefined)
                    })

                    provider.instance?.on(HLS.Events.ERROR, (event, data) => {
                        if (data?.fatal) {
                            logger("MEDIASTREAM").error("handleFatalError")
                            if (mediaContainer?.streamType === "transcode") {
                                shutdownTranscode()
                            }
                            setUrl(undefined)
                            toast.error("Playback error")
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
            setUrl(undefined)
        },
    })

    //////////////////////////////////////////////////////////////

    // Subtitle endpoint URI
    const subtitleEndpointUri = React.useMemo(() => {
        const baseUri = typeof window !== "undefined" ? ("http://" + (process.env.NODE_ENV === "development"
            ? `${window?.location?.hostname}:${__DEV_SERVER_PORT}`
            : window?.location?.host)) : ""
        if (mediaContainer?.streamUrl && mediaContainer?.streamType) {
            return `${baseUri}/api/v1/mediastream/transcode-subs`
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

        onProviderChange,
        onProviderSetup,
    }

}
