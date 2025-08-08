import { getServerBaseUrl } from "@/api/client/server-url"
import { NativePlayerDrawer } from "@/app/(main)/_features/native-player/native-player-drawer"
import { NativePlayerState } from "@/app/(main)/_features/native-player/native-player.atoms"
import { AniSkipTime } from "@/app/(main)/_features/sea-media-player/aniskip"
import {
    __seaMediaPlayer_autoNextAtom,
    __seaMediaPlayer_autoPlayAtom,
    __seaMediaPlayer_autoSkipIntroOutroAtom,
    __seaMediaPlayer_mutedAtom,
    __seaMediaPlayer_playbackRateAtom,
    __seaMediaPlayer_volumeAtom,
} from "@/app/(main)/_features/sea-media-player/sea-media-player.atoms"
import { VideoCoreAudioManager } from "@/app/(main)/_features/video-core/video-core-audio"
import { VideoCoreControlBar } from "@/app/(main)/_features/video-core/video-core-control-bar"
import {
    FlashNotificationDisplay,
    VideoCoreKeybindingController,
    VideoCoreKeybindingsModal,
} from "@/app/(main)/_features/video-core/video-core-keybindings"
import { VideoCorePreviewManager } from "@/app/(main)/_features/video-core/video-core-preview"
import { VideoCoreSubtitleManager } from "@/app/(main)/_features/video-core/video-core-subtitles"
import { VideoCoreTimeRange } from "@/app/(main)/_features/video-core/video-core-time-range"
import { VideoCoreTopPlaybackInfo, VideoCoreTopSection } from "@/app/(main)/_features/video-core/video-core-top-section"
import { vc_settings } from "@/app/(main)/_features/video-core/video-core.atoms"
import {
    detectSubtitleType,
    isSubtitleFile,
    useVideoBindings,
    vc_createChapterCues,
    vc_createChaptersFromAniSkip,
} from "@/app/(main)/_features/video-core/video-core.utils"
import { TorrentStreamOverlay } from "@/app/(main)/entry/_containers/torrent-stream/torrent-stream-overlay"
import { LuffyError } from "@/components/shared/luffy-error"
import { Button, IconButton } from "@/components/ui/button"
import { useUpdateEffect } from "@/components/ui/core/hooks"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { logger } from "@/lib/helpers/debug"
import { __isDesktop__ } from "@/types/constants"
import { atom, useAtomValue } from "jotai"
import { derive } from "jotai-derive"
import { createIsolation } from "jotai-scope"
import { useAtom, useSetAtom } from "jotai/react"
import React, { useCallback, useEffect, useMemo, useRef, useState } from "react"
import { BiExpand, BiX } from "react-icons/bi"
import { FiMinimize2 } from "react-icons/fi"
import { PiSpinnerDuotone } from "react-icons/pi"
import { RemoveScrollBar } from "react-remove-scroll-bar"
import { useMeasure } from "react-use"
import { toast } from "sonner"

const VideoCoreIsolation = createIsolation()

const log = logger("VIDEO CORE")

export const VIDEOCORE_DEBUG_ELEMENTS = false

const DELAY_BEFORE_NOT_BUSY = 1_000 //ms

export const vc_videoSize = atom({ width: 1, height: 1 })
export const vc_duration = atom(1)
export const vc_currentTime = atom(0)
export const vc_playbackRate = atom(1)
export const vc_readyState = atom(0)
export const vc_isMuted = atom(false)
export const vc_volume = atom(1)
export const vc_subtitleDelay = atom(0)
export const vc_isFullscreen = atom(false)
export const vc_pip = atom(false)
export const vc_seeking = atom(false)
export const vc_seekingTargetProgress = atom(0) // 0-100
export const vc_timeRanges = atom<TimeRanges | null>(null)
export const vc_closestBufferedTime = derive([vc_timeRanges, vc_currentTime], (tr, currentTime) => {
    if (!tr) return 0
    let closest = 0
    for (let i = 0; i < tr.length; i++) {
        const start = tr.start(i)
        const end = tr.end(i)
        if (currentTime >= start && currentTime <= end) {
            return end
        }
        if (end >= currentTime && closest > end) {
            closest = end
        }
    }
    return closest
})
export const vc_ended = atom(false)
export const vc_paused = atom(true)
export const vc_miniPlayer = atom(false)
export const vc_cursorBusy = atom(false)
export const vc_cursorPosition = atom({ x: 0, y: 0 })
export const vc_busy = atom(true)

export const vc_videoElement = atom<HTMLVideoElement | null>(null)
export const vc_videoRef = atom(() => React.createRef<HTMLVideoElement>())
export const vc_containerElement = atom<HTMLDivElement | null>(null)
export const vc_containerRef = atom(() => React.createRef<HTMLDivElement>())

export const vc_subtitleManager = atom<VideoCoreSubtitleManager | null>(null)
export const vc_audioManager = atom<VideoCoreAudioManager | null>(null)
export const vc_previewManager = atom<VideoCorePreviewManager | null>(null)

export const vc_previousPausedState = atom(false)

export const vc_dispatchAction = atom(null, (get, set, action: { type: string; payload: any }) => {
    const videoElement = get(vc_videoElement)
    if (videoElement) {
        switch (action.type) {
            // for smooth seeking, we don't want to peg the current time to the actual video time
            // instead act like the target time is instantly reached
            case "seekTo":
                videoElement.currentTime = action.payload.time
                set(vc_currentTime, action.payload.time)
                break
            case "seek":
                const currentTime = get(vc_currentTime)
                const newTime = currentTime + action.payload.time
                videoElement.currentTime = newTime
                set(vc_currentTime, newTime)
                break
        }
    }
})

export function VideoCoreProvider(props: { children: React.ReactNode }) {
    const { children } = props
    return (
        <VideoCoreIsolation.Provider>
            {children}
        </VideoCoreIsolation.Provider>
    )
}

export type VideoCoreChapterCue = {
    startTime: number
    endTime: number
    text: string
}

export interface VideoCoreProps {
    state: NativePlayerState
    aniSkipData?: {
        op: AniSkipTime | null
        ed: AniSkipTime | null
    } | undefined
    onTerminateStream: () => void
    onEnded?: () => void
    onCompleted?: () => void
    onPlay?: () => void
    onPause?: () => void
    onTimeUpdate?: () => void
    onLoadedData?: () => void
    onLoadedMetadata?: () => void
    onVolumeChange?: () => void
    onSeeking?: () => void
    onSeeked?: (time: number) => void
    onError?: (error: string) => void
    onPlaybackRateChange?: () => void
    onFileUploaded: (data: { name: string, content: string }) => void
}

// VideoCore augments the native video element.
// External states should be synced by listening to the video element's events.
export function VideoCore(props: VideoCoreProps) {
    const {
        state,
        aniSkipData,
        onTerminateStream,
        onEnded,
        onPlay,
        onCompleted,
        onPause,
        onTimeUpdate,
        onLoadedData,
        onLoadedMetadata,
        onVolumeChange,
        onSeeking,
        onSeeked,
        onError,
        onPlaybackRateChange,
        onFileUploaded,
        ...rest
    } = props


    // Ref
    const videoRef = useAtomValue(vc_videoRef)
    const [, setVideoElement] = useAtom(vc_videoElement)
    const setVideoSize = useSetAtom(vc_videoSize)
    useVideoBindings(videoRef)
    const action = useSetAtom(vc_dispatchAction)

    const videoCompletedRef = useRef(false)
    const currentPlaybackRef = useRef<string | null>(null)

    const containerRef = useAtomValue(vc_containerRef)
    const [, setContainerElement] = useAtom(vc_containerElement)

    const [subtitleManager, setSubtitleManager] = useAtom(vc_subtitleManager)
    const [audioManager, setAudioManager] = useAtom(vc_audioManager)
    const [previewManager, setPreviewManager] = useAtom(vc_previewManager)

    // States
    const settings = useAtomValue(vc_settings)
    const [isMiniPlayer, setIsMiniPlayer] = useAtom(vc_miniPlayer)
    const [busy, setBusy] = useAtom(vc_busy)
    const duration = useAtomValue(vc_duration)
    const fullscreen = useAtomValue(vc_isFullscreen)
    const paused = useAtomValue(vc_paused)

    const [showSkipIntroButton, setShowSkipIntroButton] = useState(false)
    const [showSkipEndingButton, setShowSkipEndingButton] = useState(false)

    const [autoNext, setAutoNext] = useAtom(__seaMediaPlayer_autoNextAtom)
    const [autoPlay] = useAtom(__seaMediaPlayer_autoPlayAtom)
    const [autoSkipIntroOutro] = useAtom(__seaMediaPlayer_autoSkipIntroOutroAtom)
    const [volume] = useAtom(__seaMediaPlayer_volumeAtom)
    const [muted] = useAtom(__seaMediaPlayer_mutedAtom)
    const [playbackRate, setPlaybackRate] = useAtom(__seaMediaPlayer_playbackRateAtom)

    const [measureRef, { width, height }] = useMeasure<HTMLVideoElement>()
    React.useEffect(() => {
        setVideoSize({
            width,
            height,
        })
    }, [width, height])


    React.useEffect(() => {
        if (state.active && videoRef.current && !!state.playbackInfo) {
            // Small delay to ensure the video element is fully rendered
            setTimeout(() => {
                videoRef.current?.focus()
            }, 100)
        }
    }, [state.active])

    const combineRef = (instance: HTMLVideoElement | null) => {
        if (videoRef as unknown instanceof Function) (videoRef as any)(instance)
        else if (videoRef) (videoRef as any).current = instance
        if (instance) measureRef(instance)
        setVideoElement(instance)
    }
    const combineContainerRef = (instance: HTMLDivElement | null) => {
        if (containerRef as unknown instanceof Function) (containerRef as any)(instance)
        else if (containerRef) (containerRef as any).current = instance
        setContainerElement(instance)
    }

    // actions
    function togglePlay() {
        if (videoRef?.current?.paused) {
            videoRef?.current?.play()
            onPlay?.()
        } else {
            videoRef?.current?.pause()
            onPause?.()
        }
    }

    function onCaptionsChange() {
        log.info("Subtitles changed", videoRef.current?.textTracks)
        if (videoRef.current) {
            let trackFound = false
            for (let i = 0; i < videoRef.current.textTracks.length; i++) {
                const track = videoRef.current.textTracks[i]
                if (track.mode === "showing") {
                    subtitleManager?.selectTrack(Number(track.id))
                    trackFound = true
                }
            }
            if (!trackFound) {
                subtitleManager?.setNoTrack()
            }
        }
    }

    function onAudioChange() {
        log.info("Audio changed", videoRef.current?.audioTracks)
        if (videoRef.current?.audioTracks) {
            for (let i = 0; i < videoRef.current.audioTracks.length; i++) {
                const track = videoRef.current.audioTracks[i]
                if (track.enabled) {
                    audioManager?.selectTrack(Number(track.id))
                    break
                }
            }
        }
        action({ type: "seek", payload: { time: -1 } })
    }

    useUpdateEffect(() => {
        if (!!state.playbackInfo && (!currentPlaybackRef.current || state.playbackInfo.id !== currentPlaybackRef.current)) {
            if (videoRef.current) {
                // MP4 container codec tests
                console.log("MP4 HEVC HVC1 main profile support ->", videoRef.current.canPlayType("video/mp4;codecs=\"hvc1\""))
                console.log("MP4 HEVC main profile support ->", videoRef.current.canPlayType("video/mp4;codecs=\"hev1.1.6.L120.90\""))
                console.log("MP4 HEVC main 10 profile support ->", videoRef.current.canPlayType("video/mp4;codecs=\"hev1.2.4.L120.90\""))
                console.log("MP4 HEVC main still-picture profile support ->", videoRef.current.canPlayType("video/mp4;codecs=\"hev1.3.E.L120.90\""))
                console.log("MP4 HEVC range extensions profile support ->", videoRef.current.canPlayType("video/mp4;codecs=\"hev1.4.10.L120.90\""))

                // Audio codec tests
                console.log("Dolby AC3 support ->", videoRef.current.canPlayType("audio/mp4; codecs=\"ac-3\""))
                console.log("Dolby EC3 support ->", videoRef.current.canPlayType("audio/mp4; codecs=\"ec-3\""))

                // GPU and hardware acceleration status
                const canvas = document.createElement("canvas")
                const gl = canvas.getContext("webgl2") || canvas.getContext("webgl")
                if (gl) {
                    const debugInfo = gl.getExtension("WEBGL_debug_renderer_info")
                    if (debugInfo) {
                        console.log("GPU Vendor:", gl.getParameter(debugInfo.UNMASKED_VENDOR_WEBGL))
                        console.log("GPU Renderer:", gl.getParameter(debugInfo.UNMASKED_RENDERER_WEBGL))
                    }
                }
                console.log("Hardware concurrency:", navigator.hardwareConcurrency)
                console.log("User agent:", navigator.userAgent)
            }
        }


        if (!state.playbackInfo && currentPlaybackRef.current) {
            log.info("Stream unloaded")
            subtitleManager?.terminate()
            previewManager?.cleanup()
            // setPreviewThumbnail(undefined)
            currentPlaybackRef.current = null
        }
    }, [state.playbackInfo, videoRef.current])

    // events
    const handleLoadedMetadata = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        onLoadedMetadata?.()
        if (!videoRef.current) return
        const v = videoRef.current

        log.info("Loaded metadata", v.duration)
        log.info("Audio tracks", v.audioTracks)
        log.info("Text tracks", v.textTracks)

        onCaptionsChange()
        onAudioChange()

        videoCompletedRef.current = false

        if (!state.playbackInfo || !videoRef.current) return // shouldn't happen

        // setHasUpdatedProgress(false)

        currentPlaybackRef.current = state.playbackInfo.id

        // Initialize the subtitle manager if the stream is MKV
        if (!!state.playbackInfo?.mkvMetadata) {
            setSubtitleManager(new VideoCoreSubtitleManager({
                videoElement: videoRef.current,
                playbackInfo: state.playbackInfo,
                jassubOffscreenRender: true,
                settings: settings,
            }))

            setAudioManager(new VideoCoreAudioManager({
                videoElement: videoRef.current,
                playbackInfo: state.playbackInfo,
                settings: settings,
                onError: (error) => {
                    log.error("Audio manager error", error)
                    onError?.(error)
                },
            }))
        }

        // Initialize thumbnailer
        if (state.playbackInfo?.streamUrl) {
            const streamUrl = state.playbackInfo.streamUrl.replace("{{SERVER_URL}}", getServerBaseUrl())
            log.info("Initializing thumbnailer with URL:", streamUrl)
            setPreviewManager(new VideoCorePreviewManager(videoRef.current, streamUrl))
            log.info("Thumbnailer initialized successfully")
        } else {
            log.info("No stream URL available for thumbnailer")
        }
    }

    const handleTimeUpdate = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        onTimeUpdate?.()
        if (!videoRef.current) return
        const v = videoRef.current

        // Video completed event
        const percent = v.currentTime / v.duration
        if (!!v.duration && !videoCompletedRef.current && percent >= 0.8) {
            videoCompletedRef.current = true

        }

        /**
         * AniSkip
         */
        if (
            aniSkipData?.op?.interval &&
            !!e.currentTarget.currentTime &&
            e.currentTarget.currentTime >= aniSkipData.op.interval.startTime &&
            e.currentTarget.currentTime <= aniSkipData.op.interval.endTime
        ) {
            setShowSkipIntroButton(true)
            if (autoSkipIntroOutro) {
                action({ type: "seekTo", payload: { time: aniSkipData?.op?.interval?.endTime || 0 } })
            }
        } else {
            setShowSkipIntroButton(false)
        }
        if (
            aniSkipData?.ed?.interval &&
            Math.abs(aniSkipData.ed.interval.startTime - (aniSkipData?.ed?.episodeLength)) < 500 &&
            !!e.currentTarget.currentTime &&
            e.currentTarget.currentTime >= aniSkipData.ed.interval.startTime &&
            e.currentTarget.currentTime <= aniSkipData.ed.interval.endTime
        ) {
            setShowSkipEndingButton(true)
            if (autoSkipIntroOutro) {
                action({ type: "seekTo", payload: { time: aniSkipData?.ed?.interval?.endTime || 0 } })
            }
        } else {
            setShowSkipEndingButton(false)
        }
    }

    const handleEnded = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        log.info("Video ended")
        onEnded?.()
    }

    const handleClick = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        log.info("Video clicked")
        togglePlay()
    }

    const handlePlay = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        log.info("Video resumed")
        onPlay?.()
    }

    const handlePause = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        log.info("Video paused")
        onPause?.()
    }

    const handleLoadedData = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        log.info("Loaded data")
        onLoadedData?.()
    }

    const handleVolumeChange = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        const v = e.currentTarget
        log.info("Volume changed", v.volume)
        onVolumeChange?.()
    }

    const handleRateChange = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        const v = e.currentTarget
        log.info("Playback rate changed", v.playbackRate)
        onPlaybackRateChange?.()

        if (v.playbackRate !== playbackRate) {
            setPlaybackRate(v.playbackRate)
        }
    }

    const handleError = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        onError?.("")
    }

    // external state

    // Handle volume changes
    React.useEffect(() => {
        if (videoRef.current && volume !== undefined && volume !== videoRef.current.volume) {
            videoRef.current.volume = volume
        }
    }, [volume])

    // Handle mute changes
    React.useEffect(() => {
        if (videoRef.current && muted !== undefined && muted !== videoRef.current.muted) {
            videoRef.current.muted = muted
        }
    }, [muted])

    //

    // container events
    const [cursorBusy, setCursorBusy] = useAtom(vc_cursorBusy)
    const setNotBusyTimeout = React.useRef<NodeJS.Timeout | null>(null)
    const lastPointerPosition = React.useRef({ x: 0, y: 0 })
    const handleContainerPointerMove = (e: React.PointerEvent<HTMLDivElement>) => {
        const { x, y } = e.nativeEvent
        const dx = x - lastPointerPosition.current.x
        const dy = y - lastPointerPosition.current.y
        if (Math.abs(dx) < 15 && Math.abs(dy) < 15) return
        if (setNotBusyTimeout?.current) {
            clearTimeout(setNotBusyTimeout.current)
        }
        setBusy(true)
        setNotBusyTimeout.current = setTimeout(() => {
            if (!cursorBusy) {
                setBusy(false)
            }
        }, DELAY_BEFORE_NOT_BUSY)
        lastPointerPosition.current = { x, y }
    }

    const chapterCues = useMemo(() => {
            // If we have MKV chapters, use them
            if (state.playbackInfo?.mkvMetadata?.chapters?.length) {
                const cues = vc_createChapterCues(state.playbackInfo.mkvMetadata.chapters, duration)
                log.info("Chapter cues from MKV", cues)
                return cues
            }

            // Otherwise, create chapters from AniSkip data if available
            if (!!aniSkipData?.op?.interval && duration > 0) {
                const chapters = vc_createChaptersFromAniSkip(aniSkipData, duration, state?.playbackInfo?.media?.format)
                const cues = vc_createChapterCues(chapters, duration)
                log.info("Chapter cues from AniSkip", cues)
                return cues
            }

            return []
        },
        [
            state.playbackInfo?.mkvMetadata?.chapters,
            aniSkipData?.op?.interval,
            aniSkipData?.ed?.interval,
            duration,
            state?.playbackInfo?.media?.format,
        ])

    /**
     * Upload subtitle files
     */
    type UploadEvent = {
        dataTransfer?: DataTransfer
        clipboardData?: DataTransfer
    }
    const handleUpload = useCallback(async (e: UploadEvent & Event) => {
        e.preventDefault()
        log.info("Upload event", e)
        const items = [...(e.dataTransfer ?? e.clipboardData)?.items ?? []]

        // First, try to get actual files
        const actualFiles = items
            .filter(item => item.kind === "file")
            .map(item => item.getAsFile())
            .filter(file => file !== null)

        if (actualFiles.length > 0) {
            // Process actual files
            for (const f of actualFiles) {
                if (f && isSubtitleFile(f.name)) {
                    const content = await f.text()
                    // console.log("Uploading subtitle file", f.name, content)
                    onFileUploaded({ name: f.name, content })
                }
            }
        } else {
            // If no actual files, try to process text content
            // Only process plain text, ignore RTF and HTML
            const textItems = items.filter(item =>
                item.kind === "string" &&
                item.type === "text/plain",
            )

            if (textItems.length > 0) {
                // Only take the first plain text item to avoid duplicates
                const textItem = textItems[0]
                textItem.getAsString(str => {
                    log.info("Uploading subtitle content from clipboard")
                    const type = detectSubtitleType(str)
                    log.info("Detected subtitle type", type)
                    if (type === "unknown") {
                        toast.error("Unknown subtitle type")
                        log.info("Unknown subtitle type, skipping")
                        return
                    }
                    const filename = `PLACEHOLDER.${type}`
                    onFileUploaded({ name: filename, content: str })
                })
            }
        }
    }, [])

    function suppressEvent(e: Event) {
        e.preventDefault()
    }

    useEffect(() => {
        const playerContainer = containerRef.current
        if (!playerContainer || !state.active) return

        playerContainer.addEventListener("paste", handleUpload)
        playerContainer.addEventListener("drop", handleUpload)
        playerContainer.addEventListener("dragover", suppressEvent)
        playerContainer.addEventListener("dragenter", suppressEvent)

        return () => {
            playerContainer.removeEventListener("paste", handleUpload)
            playerContainer.removeEventListener("drop", handleUpload)
            playerContainer.removeEventListener("dragover", suppressEvent)
            playerContainer.removeEventListener("dragenter", suppressEvent)
        }
    }, [handleUpload, state.active])

    return (
        <>
            <VideoCoreKeybindingsModal />
            {state.active && !isMiniPlayer && <RemoveScrollBar />}

            <NativePlayerDrawer
                open={state.active}
                onOpenChange={(v) => {
                    if (!v) {
                        // if (state.playbackError) {
                        //     handleTerminateStream()
                        //     return
                        // }
                        if (!isMiniPlayer) {
                            setIsMiniPlayer(true)
                        } else {
                            onTerminateStream()
                        }
                    }
                }}
                borderToBorder
                miniPlayer={isMiniPlayer}
                size={isMiniPlayer ? "md" : "full"}
                side={isMiniPlayer ? "right" : "bottom"}
                contentClass={cn(
                    "p-0 m-0",
                    !isMiniPlayer && "h-full",
                )}
                allowOutsideInteraction={true}
                overlayClass={cn(
                    isMiniPlayer && "hidden",
                )}
                hideCloseButton
                closeClass={cn(
                    "z-[99]",
                    __isDesktop__ && !isMiniPlayer && "top-8",
                    isMiniPlayer && "left-4",
                )}
                data-native-player-drawer
            >
                {!(!!state.playbackInfo?.streamUrl && !state.loadingState) && <TorrentStreamOverlay isNativePlayerComponent />}

                {(state?.playbackError) && (
                    <div className="h-full w-full bg-black/80 flex items-center justify-center z-[50] absolute p-4">
                        <div className="text-white text-center">
                            {!isMiniPlayer ? (
                                <LuffyError title="Playback Error" />
                            ) : (
                                <h1 className={cn("text-2xl font-bold", isMiniPlayer && "text-lg")}>Playback Error</h1>
                            )}
                            <p className={cn("text-base text-white/50", isMiniPlayer && "text-sm max-w-lg mx-auto")}>
                                {state.playbackError || "An error occurred while playing the stream. Please try again later."}
                            </p>
                        </div>
                    </div>
                )}

                <div
                    ref={combineContainerRef}
                    className={cn(
                        "relative w-full h-full bg-black overflow-clip flex items-center justify-center",
                        !busy && "cursor-none", // show cursor when not busy
                    )}
                    onPointerMove={handleContainerPointerMove}
                >

                    {(!!state.playbackInfo?.streamUrl && !state.loadingState) ? (
                        <>

                            <VideoCoreKeybindingController
                                active={state.active}
                                videoRef={videoRef}
                                chapterCues={chapterCues ?? []}
                                introStartTime={aniSkipData?.op?.interval?.startTime}
                                introEndTime={aniSkipData?.op?.interval?.endTime}
                            />

                            <FlashNotificationDisplay />

                            {/*<MediaLoadingIndicator*/}
                            {/*    slot="centered-chrome"*/}
                            {/*    loadingDelay={300}*/}
                            {/*    className="native-player-loading-indicator"*/}
                            {/*/>*/}

                            {/* Skip Intro/Ending Buttons */}
                            {busy && <>
                                {showSkipIntroButton && !isMiniPlayer && !state.playbackInfo?.mkvMetadata?.chapters?.length && (
                                    <div className="absolute left-5 bottom-28 z-[60] native-player-hide-on-fullscreen">
                                        <Button
                                            size="sm"
                                            intent="gray-basic"
                                            onClick={e => {
                                                e.stopPropagation()
                                                action({ type: "seekTo", payload: { time: aniSkipData?.op?.interval?.endTime || 0 } })
                                            }}
                                            onPointerMove={e => e.stopPropagation()}
                                        >
                                            Skip Opening
                                        </Button>
                                    </div>
                                )}

                                {showSkipEndingButton && !isMiniPlayer && !state.playbackInfo?.mkvMetadata?.chapters?.length && (
                                    <div className="absolute right-5 bottom-28 z-[60] native-player-hide-on-fullscreen">
                                        <Button
                                            size="sm"
                                            intent="gray-basic"
                                            onClick={e => {
                                                e.stopPropagation()
                                                action({ type: "seekTo", payload: { time: aniSkipData?.ed?.interval?.endTime || 0 } })
                                            }}
                                            onPointerMove={e => e.stopPropagation()}
                                        >
                                            Skip Ending
                                        </Button>
                                    </div>
                                )}
                            </>}

                            <video
                                data-video-core-element
                                crossOrigin="anonymous"
                                preload="auto"
                                src={state.playbackInfo.streamUrl.replace("{{SERVER_URL}}", getServerBaseUrl())}
                                ref={combineRef}
                                onLoadedMetadata={handleLoadedMetadata}
                                onTimeUpdate={handleTimeUpdate}
                                onEnded={handleEnded}
                                onPlay={handlePlay}
                                onPause={handlePause}
                                onClick={handleClick}
                                onDoubleClick={() => {}}
                                onLoadedData={handleLoadedData}
                                onVolumeChange={handleVolumeChange}
                                onRateChange={handleRateChange}
                                onError={handleError}
                                autoPlay={autoPlay}
                                muted={muted}
                                playsInline
                                controls={false}
                                style={{
                                    width: "100%",
                                    height: "100%",
                                    border: "none",
                                    filter: settings.videoEnhancement.enabled
                                        ? `contrast(${settings.videoEnhancement.contrast}) saturate(${settings.videoEnhancement.saturation}) brightness(${settings.videoEnhancement.brightness})`
                                        : "none",
                                    imageRendering: "auto",
                                }}
                            >
                                {state.playbackInfo?.mkvMetadata?.subtitleTracks?.map(track => (
                                    <track
                                        id={track.number.toString()}
                                        key={track.number}
                                        kind="subtitles"
                                        srcLang={track.language || "eng"}
                                        label={track.name}
                                    />
                                ))}
                            </video>

                            <VideoCoreTopSection>
                                <VideoCoreTopPlaybackInfo state={state} />

                                <div
                                    className={cn(
                                        "opacity-0",
                                        "transition-opacity duration-200 ease-in-out",
                                        (busy || paused) && "opacity-100",
                                    )}
                                >
                                    <FloatingButtons part="video" onTerminateStream={onTerminateStream} />
                                </div>
                                {/*<TorrentStreamOverlay isNativePlayerComponent="info" />*/}
                            </VideoCoreTopSection>

                            <VideoCoreControlBar
                                timeRange={<VideoCoreTimeRange
                                    chapterCues={chapterCues ?? []}
                                />}
                            >

                            </VideoCoreControlBar>

                        </>
                    ) : (
                        <div
                            className="w-full h-full absolute flex justify-center items-center flex-col space-y-4 bg-black rounded-md"
                        >

                            {/* {!state.miniPlayer && <SquareBg className="absolute top-0 left-0 w-full h-full z-[0]" />} */}
                            <FloatingButtons part="loading" onTerminateStream={onTerminateStream} />

                            <LoadingSpinner
                                title={state.loadingState || "Loading..."}
                                spinner={<PiSpinnerDuotone className="size-20 text-white animate-spin" />}
                                containerClass="z-[1]"
                            />
                        </div>
                    )}


                </div>
            </NativePlayerDrawer>
        </>
    )
}

function FloatingButtons(props: { part: "video" | "loading", onTerminateStream: () => void }) {
    const { part, onTerminateStream } = props
    const fullscreen = useAtomValue(vc_isFullscreen)
    const [isMiniPlayer, setIsMiniPlayer] = useAtom(vc_miniPlayer)
    if (fullscreen) return null
    const Content = () => (
        <>
            {!isMiniPlayer && <>
                <IconButton
                    icon={<FiMinimize2 className="text-2xl" />}
                    intent="gray-basic"
                    className="rounded-full absolute top-0 flex-none right-4"
                    onClick={() => {
                        setIsMiniPlayer(true)
                    }}
                />
            </>}

            {isMiniPlayer && <>
                <IconButton
                    type="button"
                    intent="gray"
                    size="sm"
                    className={cn(
                        "rounded-full text-2xl flex-none absolute z-[99] right-4 top-4 pointer-events-auto bg-black/30 hover:bg-black/40",
                        isMiniPlayer && "text-xl",
                    )}
                    icon={<BiExpand />}
                    onClick={() => {
                        setIsMiniPlayer(false)
                    }}
                />
                <IconButton
                    type="button"
                    intent="alert-subtle"
                    size="sm"
                    className={cn(
                        "rounded-full text-2xl flex-none absolute z-[99] left-4 top-4 pointer-events-auto",
                        isMiniPlayer && "text-xl",
                    )}
                    icon={<BiX />}
                    onClick={() => {
                        onTerminateStream()
                    }}
                />
            </>}
        </>
    )

    if (part === "loading") {
        return (
            <div
                className={cn(
                    "absolute top-8 w-full z-[100]",
                    isMiniPlayer && "top-0",
                )}
            >
                <Content />
            </div>
        )
    }

    return <Content />
}
