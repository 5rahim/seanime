import { getServerBaseUrl } from "@/api/client/server-url"
import { Anime_Episode, Mediastream_StreamType, Nullish } from "@/api/generated/types"
import { useHandleContinuityWithMediaPlayer, useHandleCurrentMediaContinuity } from "@/api/hooks/continuity.hooks"
import { useGetMediastreamSettings, useMediastreamShutdownTranscodeStream, useRequestMediastreamMediaContainer } from "@/api/hooks/mediastream.hooks"
import { usePlaylistManager } from "@/app/(main)/_features/playlists/_containers/global-playlist-manager"
import { useIsCodecSupported } from "@/app/(main)/_features/sea-media-player/hooks"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { useMediastreamCurrentFile, useMediastreamJassubOffscreenRender } from "@/app/(main)/mediastream/_lib/mediastream.atoms"
import { clientIdAtom } from "@/app/websocket-provider"
import { logger } from "@/lib/helpers/debug"
import { legacy_getAssetUrl } from "@/lib/server/assets"
import { WSEvents } from "@/lib/server/ws-events"
import {
    isHLSProvider,
    LibASSTextRenderer,
    MediaCanPlayDetail,
    MediaPlayerInstance,
    MediaProviderAdapter,
    MediaProviderChangeEvent,
    MediaProviderSetupEvent,
} from "@vidstack/react"
import HLS, { LoadPolicy } from "hls.js"
import { useAtomValue } from "jotai"
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
            maxLoadTimeMs: 300_000,
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
        startLevel: 0, // Start at level 0
        lowLatencyMode: false,
        initialLiveManifestSize: 0,
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

type HandleMediastreamProps = {
    playerRef: React.RefObject<MediaPlayerInstance>
    episodes: Anime_Episode[]
    mediaId: Nullish<string | number>
}

export function useHandleMediastream(props: HandleMediastreamProps) {

    const {
        playerRef,
        episodes,
        mediaId,
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

    // Refs
    const previousCurrentTimeRef = React.useRef(0)
    const previousIsPlayingRef = React.useRef(false)

    const sessionId = useAtomValue(clientIdAtom)

    /**
     * Watch history
     */
    const { waitForWatchHistory } = useHandleCurrentMediaContinuity(mediaId)

    /**
     * Fetch media container containing stream URL
     */
    const { data: _mediaContainer, isError: isMediaContainerError, isPending, isFetching, refetch } = useRequestMediastreamMediaContainer({
        path: filePath,
        streamType: streamType,
        clientId: sessionId ?? uuidv4(),
    }, !!mediastreamSettings && !mediastreamSettingsLoading && !waitForWatchHistory)

    const mediaContainer = React.useMemo(() => (!isPending && !isFetching) ? _mediaContainer : undefined, [_mediaContainer, isPending, isFetching])

    // const { mutate: preloadMediaContainer } = usePreloadMediastreamMediaContainer()
    // const [preloadedFilePath, setPreloadedFilePath] = React.useState<string | undefined>(undefined)


    // Whether the playback has errored
    const [playbackErrored, setPlaybackErrored] = React.useState<boolean>(false)

    // Duration
    const [duration, setDuration] = React.useState<number>(0)

    React.useEffect(() => {
        if (isPending) {
            logger("MEDIASTREAM").info("Loading media container")
            changeUrl(undefined)
            logger("MEDIASTREAM").info("Setting URL to undefined")
        }
    }, [isPending])

    const { mutate: shutdownTranscode } = useMediastreamShutdownTranscodeStream()

    /**
     * This error happens when the media container is available but the URL has been set to undefined
     * - This is usually the case when the transcoder has errored out
     */
    const isStreamError = !!mediaContainer && !url

    const { isCodecSupported } = useIsCodecSupported()

    /**
     * Effect triggered when media container is available
     * - Check compatibility
     * - Set URL and stream type when media container is available
     */
    React.useEffect(() => {
        logger("MEDIASTREAM").info("Media container changed, running effect", mediaContainer)

        /**
         * Check if codec is supported, if it is, switch to direct play
         */
        const codecSupported = isCodecSupported(mediaContainer?.mediaInfo?.mimeCodec ?? "")
        logger("MEDIASTREAM").info("Is codec supported", codecSupported)

        // If the codec is supported, switch to direct play
        if (mediaContainer?.streamType === "transcode") {
            logger("MEDIASTREAM").info("Stream type is transcode")

            if (!codecSupported && mediastreamSettings?.directPlayOnly) {
                logger("MEDIASTREAM").warning("Codec not supported for direct play", mediaContainer?.mediaInfo?.mimeCodec)
                logger("MEDIASTREAM").warning("Stopping playback")
                toast.warning("Codec not supported for direct play")
                changeUrl(undefined)
                logger("MEDIASTREAM").info("Setting URL to undefined")
                return
            }

            if (codecSupported && (!mediastreamSettings?.disableAutoSwitchToDirectPlay || mediastreamSettings?.directPlayOnly)) {
                logger("MEDIASTREAM").info("Codec supported", mediaContainer?.mediaInfo?.mimeCodec)
                logger("MEDIASTREAM").warning("Switching to direct play")
                setStreamType("direct")
                changeUrl(undefined)
                logger("MEDIASTREAM").info("Setting URL to undefined")
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
                logger("MEDIASTREAM").info("Setting URL to undefined")
                return
            }
        }

        if (mediaContainer?.streamUrl) {
            logger("MEDIASTREAM").info("Stream URL available", mediaContainer.streamUrl)

            const _newUrl = `${getServerBaseUrl()}${mediaContainer.streamUrl}`

            logger("MEDIASTREAM").info("Changing URL", _newUrl, "streamType:", mediaContainer.streamType)

            changeUrl(_newUrl)
        } else {
            changeUrl(undefined)
            logger("MEDIASTREAM").info("Setting URL to undefined")
        }

    }, [mediaContainer?.streamUrl, mediastreamSettings?.disableAutoSwitchToDirectPlay])

    //////////////////////////////////////////////////////////////
    // JASSUB
    //////////////////////////////////////////////////////////////

    const { jassubOffscreenRender } = useMediastreamJassubOffscreenRender()

    /**
     * Effect used to set LibASS renderer
     * Add subtitle renderer
     */
    React.useEffect(() => {
        if (playerRef.current && !!mediaContainer?.mediaInfo?.fonts?.length) {
            logger("MEDIASTREAM").info("Adding JASSUB renderer to player", mediaContainer?.mediaInfo?.fonts?.length, "fonts")
            const legacyWasmUrl = process.env.NODE_ENV === "development"
                ? "/jassub/jassub-worker.wasm.js" : legacy_getAssetUrl("/jassub/jassub-worker.wasm.js")

            logger("MEDIASTREAM").info("Loading JASSUB renderer")

            const fonts = mediaContainer?.mediaInfo?.fonts?.map(name => `${getServerBaseUrl()}/api/v1/mediastream/att/${name}`) || []

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
                    "liberation sans": getServerBaseUrl() + `/jassub/default.woff2`,
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
                offscreenRender: jassubOffscreenRender, // should be false for iOS
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
    }, [
        playerRef.current,
        mediaContainer?.streamUrl,
        mediaContainer?.mediaInfo?.fonts,
        jassubOffscreenRender,
    ])

    /**
     * Changes the stream URL
     * @param newUrl
     */
    function changeUrl(newUrl: string | undefined) {
        logger("MEDIASTREAM").info("[changeUrl] called,", "request url:", newUrl)
        if (prevUrlRef.current !== newUrl) {
            logger("MEDIASTREAM").info("Resetting playback error status")
            setPlaybackErrored(false)
        }
        setUrl(prevUrl => {
            if (prevUrl === newUrl) {
                logger("MEDIASTREAM").info("[changeUrl] URL has not changed")
                return prevUrl
            }
            prevUrlRef.current = prevUrl
            logger("MEDIASTREAM").info("[changeUrl] URL updated")
            return newUrl
        })
        if (newUrl) {
            definedUrlRef.current = newUrl
        }
    }

    //////////////////////////////////////////////////////////////
    // Media player
    //////////////////////////////////////////////////////////////

    function onProviderChange(provider: MediaProviderAdapter | null, nativeEvent: MediaProviderChangeEvent) {
        if (isHLSProvider(provider) && mediaContainer?.streamType === "transcode") {
            logger("MEDIASTREAM").info("[onProviderChange] Provider changed to HLS")
            provider.library = HLS
            provider.config = {
                ...mediastream_getHlsConfig(),
            }
        } else {
            logger("MEDIASTREAM").info("[onProviderChange] Provider changed to native")
        }
    }

    function onProviderSetup(provider: MediaProviderAdapter, nativeEvent: MediaProviderSetupEvent) {
        if (isHLSProvider(provider)) {
            if (url) {

                if (HLS.isSupported() && url.endsWith(".m3u8")) {

                    logger("MEDIASTREAM").info("[onProviderSetup] HLS Provider setup")
                    logger("MEDIASTREAM").info("[onProviderSetup] Loading source", url)

                    provider.instance?.on(HLS.Events.MANIFEST_PARSED, function (event, data) {
                        logger("MEDIASTREAM").info("onManifestParsed", data)
                        // Check if the manifest is live or VOD
                        data.levels.forEach((level) => {
                            logger("MEDIASTREAM").info(`Level ${level.id} is live:`, level.details?.live)
                        })
                    })

                    provider.instance?.on(HLS.Events.MEDIA_ATTACHED, (event) => {
                        logger("MEDIASTREAM").info("onMediaAttached")
                    })

                    provider.instance?.on(HLS.Events.MEDIA_DETACHED, (event) => {
                        logger("MEDIASTREAM").warning("onMediaDetached")
                        // When the media is detached, stop the transcoder but only if there was no playback error
                        if (!playbackErrored) {
                            if (mediaContainer?.streamType === "transcode") {
                                // DEVNOTE: Code below kills the transcoder AFTER changing episode due to delay
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
                logger("MEDIASTREAM").error("[onProviderSetup] Provider setup - no URL")
            }
        } else {
            logger("MEDIASTREAM").info("[onProviderSetup] Provider setup - not HLS")
        }
    }


    /**
     * Current episode
     */
    const episode = React.useMemo(() => {
        return episodes.find(ep => !!ep.localFile?.path && ep.localFile?.path === filePath)
    }, [episodes, filePath])

    /**
     * Continuity
     */
    const { handleUpdateWatchHistory } = useHandleContinuityWithMediaPlayer(playerRef, episode?.episodeNumber, mediaId)


    const preloadedNextFileForRef = React.useRef<string | undefined>(undefined) // unused

    const onCanPlay = (e: MediaCanPlayDetail) => {
        logger("MEDIASTREAM").info("[onCanPlay] called", e)
        preloadedNextFileForRef.current = undefined
        setDuration(e.duration)
    }

    const { currentPlaylist, playEpisode: playPlaylistEpisode, nextPlaylistEpisode, prevPlaylistEpisode } = usePlaylistManager()

    const currentEpisodeIndex = episodes.findIndex(ep => !!ep.localFile?.path && ep.localFile?.path === filePath)

    const nextFile = currentEpisodeIndex === -1 ? undefined : episodes?.[currentEpisodeIndex + 1]
    const prevFile = currentEpisodeIndex === -1 ? undefined : episodes?.[currentEpisodeIndex - 1]

    const hasNextEpisode = !!nextFile || (currentPlaylist && !!nextPlaylistEpisode)
    const hasPreviousEpisode = !!prevFile || (currentPlaylist && !!prevPlaylistEpisode)

    const playNextEpisode = () => {
        logger("MEDIASTREAM").info("[playNextEpisode] called")
        if (currentPlaylist) {
            playPlaylistEpisode("next", true)
            return
        }
        if (nextFile) {
            if (nextFile?.localFile?.path) {
                onPlayFile(nextFile.localFile.path)
            }
        }
    }

    const playPreviousEpisode = () => {
        logger("MEDIASTREAM").info("[playPreviousEpisode] called")
        if (currentPlaylist) {
            playPlaylistEpisode("previous", false)
            return
        }
        if (prevFile) {
            if (prevFile?.localFile?.path) {
                onPlayFile(prevFile.localFile.path)
            }
        }
    }

    const onPlayFile = (filepath: string) => {
        logger("MEDIASTREAM").info("Playing file", filepath)
        playerRef.current?.destroy?.()
        previousCurrentTimeRef.current = 0
        setFilePath(filepath)
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
            changeUrl(undefined)
        },
    })

    //////////////////////////////////////////////////////////////

    // Subtitle endpoint URI
    const subtitleEndpointUri = React.useMemo(() => {
        if (mediaContainer?.streamUrl && mediaContainer?.streamType) {
            return `${getServerBaseUrl()}/api/v1/mediastream/subs`
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
        episode,
        duration,
        disabledAutoSwitchToDirectPlay: mediastreamSettings?.disableAutoSwitchToDirectPlay,
        setStreamType: (type: Mediastream_StreamType) => {
            logger("MEDIASTREAM").info("[setStreamType] Setting stream type", type)
            setStreamType(type)
            playerRef.current?.destroy?.()
            changeUrl(undefined)
        },
        onCanPlay,
        playNextEpisode,
        playPreviousEpisode,
        hasPreviousEpisode,
        hasNextEpisode,
        onProviderChange,
        onProviderSetup,
        isCodecSupported,
        handleUpdateWatchHistory,
    }

}
