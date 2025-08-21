import { getServerBaseUrl } from "@/api/client/server-url"
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
import { vc_doFlashAction, VideoCoreActionDisplay } from "@/app/(main)/_features/video-core/video-core-action-display"
import { vc_anime4kOption, VideoCoreAnime4K } from "@/app/(main)/_features/video-core/video-core-anime-4k"
import { Anime4KOption, VideoCoreAnime4KManager } from "@/app/(main)/_features/video-core/video-core-anime-4k-manager"
import { VideoCoreAudioManager } from "@/app/(main)/_features/video-core/video-core-audio"
import {
    vc_hoveringControlBar,
    VideoCoreAudioButton,
    VideoCoreControlBar,
    VideoCoreFullscreenButton,
    VideoCoreNextButton,
    VideoCorePipButton,
    VideoCorePlayButton,
    VideoCorePreviousButton,
    VideoCoreSettingsButton,
    VideoCoreSubtitleButton,
    VideoCoreTimestamp,
    VideoCoreVolumeButton,
} from "@/app/(main)/_features/video-core/video-core-control-bar"
import { VideoCoreDrawer } from "@/app/(main)/_features/video-core/video-core-drawer"
import { vc_fullscreenManager, VideoCoreFullscreenManager } from "@/app/(main)/_features/video-core/video-core-fullscreen"
import { VideoCoreKeybindingController, VideoCoreKeybindingsModal } from "@/app/(main)/_features/video-core/video-core-keybindings"
import { vc_pipElement, vc_pipManager, VideoCorePipManager } from "@/app/(main)/_features/video-core/video-core-pip"
import { useVideoCorePlaylistSetup } from "@/app/(main)/_features/video-core/video-core-playlist"
import { VideoCorePreviewManager } from "@/app/(main)/_features/video-core/video-core-preview"
import { VideoCoreSubtitleManager } from "@/app/(main)/_features/video-core/video-core-subtitles"
import { VideoCoreTimeRange } from "@/app/(main)/_features/video-core/video-core-time-range"
import { VideoCoreTopPlaybackInfo, VideoCoreTopSection } from "@/app/(main)/_features/video-core/video-core-top-section"
import { vc_beautifyImageAtom, vc_settings } from "@/app/(main)/_features/video-core/video-core.atoms"
import {
    detectSubtitleType,
    isSubtitleFile,
    useVideoCoreBindings,
    vc_createChapterCues,
    vc_createChaptersFromAniSkip,
    vc_formatTime,
    vc_logGeneralInfo,
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
export const vc_realVideoSize = atom({ width: 0, height: 0 })
export const vc_duration = atom(1)
export const vc_currentTime = atom(0)
export const vc_playbackRate = atom(1)
export const vc_readyState = atom(0)
export const vc_buffering = atom(false)
export const vc_isMuted = atom(false)
export const vc_volume = atom(1)
export const vc_subtitleDelay = atom(0)
export const vc_isFullscreen = atom(false)
export const vc_pip = derive([vc_pipElement], (pipElement) => pipElement !== null)
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
export const vc_menuOpen = atom(false)
export const vc_cursorBusy = derive([vc_hoveringControlBar, vc_menuOpen], (f1, f2) => {
    return f1 || f2
})
export const vc_cursorPosition = atom({ x: 0, y: 0 })
export const vc_busy = atom(true)

export const vc_videoElement = atom<HTMLVideoElement | null>(null)
export const vc_containerElement = atom<HTMLDivElement | null>(null)

export const vc_subtitleManager = atom<VideoCoreSubtitleManager | null>(null)
export const vc_audioManager = atom<VideoCoreAudioManager | null>(null)
export const vc_previewManager = atom<VideoCorePreviewManager | null>(null)
export const vc_anime4kManager = atom<VideoCoreAnime4KManager | null>(null)

export const vc_previousPausedState = atom(false)

export const vc_lastKnownProgress = atom(0)

type VideoCoreAction = "seekTo" | "seek" | "togglePlay" | "restoreProgress"

export const vc_dispatchAction = atom(null, (get, set, action: { type: VideoCoreAction; payload?: any }) => {
    const videoElement = get(vc_videoElement)
    const duration = get(vc_duration)
    let t = 0
    if (videoElement) {
        switch (action.type) {
            // for smooth seeking, we don't want to peg the current time to the actual video time
            // instead act like the target time is instantly reached
            case "seekTo":
                t = Math.min(duration, Math.max(0, action.payload.time))
                videoElement.currentTime = t
                set(vc_currentTime, t)
                if (action.payload.flashTime) {
                    set(vc_doFlashAction, { message: `${vc_formatTime(t)} / ${vc_formatTime(duration)}`, type: "message" })
                }
                break
            case "seek":
                const currentTime = get(vc_currentTime)
                t = Math.min(duration, Math.max(0, currentTime + action.payload.time))
                videoElement.currentTime = t
                set(vc_currentTime, t)
                if (action.payload.flashTime) {
                    set(vc_doFlashAction, { message: `${vc_formatTime(t)} / ${vc_formatTime(duration)}`, type: "message" })
                }
                break
            case "togglePlay":
                videoElement.paused ? videoElement.play() : videoElement.pause()
                break
            case "restoreProgress":
                // Restore time to the last known position
                set(vc_lastKnownProgress, Math.max(0, action.payload.time))
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

    const videoRef = useRef<HTMLVideoElement | null>(null)
    const containerRef = useRef<HTMLDivElement | null>(null)

    const setVideoElement = useSetAtom(vc_videoElement)
    const setRealVideoSize = useSetAtom(vc_realVideoSize)
    useVideoCoreBindings(state.playbackInfo)
    // useVideoCoreAnime4K()
    useVideoCorePlaylistSetup()

    const videoCompletedRef = useRef(false)
    const currentPlaybackRef = useRef<string | null>(null)

    const [, setContainerElement] = useAtom(vc_containerElement)

    const [subtitleManager, setSubtitleManager] = useAtom(vc_subtitleManager)
    const [audioManager, setAudioManager] = useAtom(vc_audioManager)
    const [previewManager, setPreviewManager] = useAtom(vc_previewManager)
    const [anime4kManager, setAnime4kManager] = useAtom(vc_anime4kManager)
    const [pipManager, setPipManager] = useAtom(vc_pipManager)
    const setPipElement = useSetAtom(vc_pipElement)
    const [fullscreenManager, setFullscreenManager] = useAtom(vc_fullscreenManager)
    const setIsFullscreen = useSetAtom(vc_isFullscreen)
    const action = useSetAtom(vc_dispatchAction)

    // States
    const settings = useAtomValue(vc_settings)
    const [isMiniPlayer, setIsMiniPlayer] = useAtom(vc_miniPlayer)
    const [busy, setBusy] = useAtom(vc_busy)
    const [buffering, setBuffering] = useAtom(vc_buffering)
    const duration = useAtomValue(vc_duration)
    const fullscreen = useAtomValue(vc_isFullscreen)
    const paused = useAtomValue(vc_paused)
    const readyState = useAtomValue(vc_readyState)
    const beautifyImage = useAtomValue(vc_beautifyImageAtom)
    const isPip = useAtomValue(vc_pip)
    const flashAction = useSetAtom(vc_doFlashAction)
    const dispatchAction = useSetAtom(vc_dispatchAction)

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
        setRealVideoSize({
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
        videoRef.current = instance
        if (instance) measureRef(instance)
        setVideoElement(instance)
    }
    const combineContainerRef = (instance: HTMLDivElement | null) => {
        containerRef.current = instance
        setContainerElement(instance)
    }

    // actions
    function togglePlay() {
        if (videoRef?.current?.paused) {
            videoRef?.current?.play()
            onPlay?.()
            flashAction({ message: "PLAY", type: "icon" })
        } else {
            videoRef?.current?.pause()
            onPause?.()
            flashAction({ message: "PAUSE", type: "icon" })
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
        if (!state.playbackInfo) {
            log.info("Cleaning up")
            setVideoElement(null)
            subtitleManager?.destroy?.()
            setSubtitleManager(null)
            previewManager?.cleanup?.()
            setPreviewManager(null)
            setAudioManager(null)
            anime4kManager?.destroy?.()
            setAnime4kManager(null)
            pipManager?.destroy?.()
            setPipManager(null)
            setPipElement(null)
            fullscreenManager?.destroy?.()
            setFullscreenManager(null)
            setIsFullscreen(false)
            currentPlaybackRef.current = null
            videoRef.current = null
        }

        if (!!state.playbackInfo && (!currentPlaybackRef.current || state.playbackInfo.id !== currentPlaybackRef.current)) {
            log.info("New stream loaded")
            vc_logGeneralInfo(videoRef.current)
        }
    }, [state.playbackInfo, videoRef.current])

    const streamUrl = state?.playbackInfo?.streamUrl?.replace?.("{{SERVER_URL}}", getServerBaseUrl())

    const [anime4kOption, setAnime4kOption] = useAtom(vc_anime4kOption)

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

        if (!state.playbackInfo) return // shouldn't happen

        // setHasUpdatedProgress(false)

        currentPlaybackRef.current = state.playbackInfo.id

        // Initialize the subtitle manager if the stream is MKV
        if (!!state.playbackInfo?.mkvMetadata) {
            setSubtitleManager(p => {
                if (p) p.destroy()
                return new VideoCoreSubtitleManager({
                    videoElement: v!,
                    playbackInfo: state.playbackInfo!,
                    jassubOffscreenRender: true,
                    settings: settings,
                })
            })

            setAudioManager(new VideoCoreAudioManager({
                videoElement: v!,
                playbackInfo: state.playbackInfo,
                settings: settings,
                onError: (error) => {
                    log.error("Audio manager error", error)
                    onError?.(error)
                },
            }))
        }

        // Initialize Anime4K manager
        setAnime4kManager(p => {
            if (p) p.destroy()
            return new VideoCoreAnime4KManager({
                videoElement: v!,
                settings: settings,
                onFallback: (message) => {
                    flashAction({ message, duration: 2000 })
                },
                onOptionChanged: (opt) => {
                    console.warn("here", opt)
                    setAnime4kOption(opt)
                },
            })
        })

        // Initialize PIP manager
        setPipManager(p => {
            if (p) p.destroy()
            const manager = new VideoCorePipManager((element) => {
                setPipElement(element)
            })
            manager.setVideo(v!, state.playbackInfo!)
            return manager
        })

        // Initialize fullscreen manager
        setFullscreenManager(p => {
            if (p) p.destroy()
            return new VideoCoreFullscreenManager((isFullscreen: boolean) => {
                setIsFullscreen(isFullscreen)
            })
        })

        log.info("Initializing preview manager")
        // TODO uncomment
        // setPreviewManager(p => {
        //     if (p) p.cleanup()
        //     return new VideoCorePreviewManager(v!, streamUrl)
        // })
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

    const handleDoubleClick = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        // fullscreenManager?.toggleFullscreen()
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

    const handleWaiting = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        setBuffering(true)
    }

    const handleCanPlay = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        setBuffering(false)
    }

    const handleStalled = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        setBuffering(true)
    }

    // external state

    // Handle volume changes
    React.useEffect(() => {
        if (videoRef.current && volume !== undefined && volume !== videoRef.current.volume) {
            videoRef.current.volume = volume
        }
    }, [volume, videoRef.current])

    // Handle mute changes
    React.useEffect(() => {
        if (videoRef.current && muted !== undefined && muted !== videoRef.current.muted) {
            videoRef.current.muted = muted
        }
    }, [muted, videoRef.current])

    // Handle playback rate changes
    React.useEffect(() => {
        if (videoRef.current && playbackRate !== undefined && playbackRate !== videoRef.current.playbackRate) {
            videoRef.current.playbackRate = playbackRate
        }
    }, [playbackRate, videoRef.current])

    // Update PIP manager
    React.useEffect(() => {
        if (pipManager && videoRef.current && state.playbackInfo) {
            pipManager.setVideo(videoRef.current, state.playbackInfo)
            if (subtitleManager) pipManager.setSubtitleManager(subtitleManager)
        }
    }, [pipManager, subtitleManager, videoRef.current, state.playbackInfo])

    // Update fullscreen manager
    React.useEffect(() => {
        if (fullscreenManager && containerRef.current) {
            fullscreenManager.setContainer(containerRef.current)
        }
    }, [fullscreenManager, containerRef.current])

    //

    // container events
    const cursorBusy = useAtomValue(vc_cursorBusy)
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

    /**
     * Restore last position
     */
    const [restoreProgressTo, setRestoreProgressTo] = useAtom(vc_lastKnownProgress)
    React.useEffect(() => {
        if (!anime4kManager || !restoreProgressTo) return

        if (anime4kOption === "off" || anime4kManager.canvas !== null) {
            dispatchAction({ type: "seekTo", payload: { time: restoreProgressTo } })
        } else if (anime4kOption !== ("off" as Anime4KOption) && anime4kManager.canvas === null) {
            anime4kManager.registerOnCanvasCreated(() => {
                dispatchAction({ type: "seekTo", payload: { time: restoreProgressTo } })
                setRestoreProgressTo(0)
            })
        }

    }, [anime4kManager, anime4kOption, restoreProgressTo])

    return (
        <>
            <VideoCoreAnime4K />
            <VideoCoreKeybindingsModal />
            {state.active && !isMiniPlayer && <RemoveScrollBar />}

            <VideoCoreDrawer
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
                onMiniPlayerClick={() => {
                    togglePlay()
                }}
            >
                {!(!!state.playbackInfo?.streamUrl && !state.loadingState) && <TorrentStreamOverlay isNativePlayerComponent />}

                {(state?.playbackError) && (
                    <div className="h-full w-full bg-black/80 flex items-center justify-center z-[200] absolute p-4">
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
                        (!busy && !isMiniPlayer) && "cursor-none", // show cursor when not busy
                    )}
                    onPointerMove={handleContainerPointerMove}
                    // onPointerLeave={() => setBusy(false)}
                    // onPointerCancel={() => setBusy(false)}
                >

                    {(!!state.playbackInfo?.streamUrl && !state.loadingState) ? (
                        <>

                            <VideoCoreKeybindingController
                                videoRef={videoRef}
                                active={state.active}
                                chapterCues={chapterCues ?? []}
                                introStartTime={aniSkipData?.op?.interval?.startTime}
                                introEndTime={aniSkipData?.op?.interval?.endTime}
                                endingStartTime={aniSkipData?.ed?.interval?.startTime}
                                endingEndTime={aniSkipData?.ed?.interval?.endTime}
                            />

                            <VideoCoreActionDisplay />

                            {buffering && (
                                <div className="absolute inset-0 flex items-center justify-center z-[50] pointer-events-none">
                                    <div className="bg-black/20 backdrop-blur-sm rounded-full p-4">
                                        <PiSpinnerDuotone className="size-12 text-white animate-spin" />
                                    </div>
                                </div>
                            )}

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

                            <div className="relative w-full h-full flex items-center justify-center">
                                <video
                                    data-video-core-element
                                    crossOrigin="anonymous"
                                    // preload="metadata"
                                    preload="auto"
                                    src={streamUrl!}
                                    ref={combineRef}
                                    onLoadedMetadata={handleLoadedMetadata}
                                    onTimeUpdate={handleTimeUpdate}
                                    onEnded={handleEnded}
                                    onPlay={handlePlay}
                                    onPause={handlePause}
                                    onClick={handleClick}
                                    onDoubleClick={handleDoubleClick}
                                    onLoadedData={handleLoadedData}
                                    onVolumeChange={handleVolumeChange}
                                    onRateChange={handleRateChange}
                                    onError={handleError}
                                    onWaiting={handleWaiting}
                                    onCanPlay={handleCanPlay}
                                    onStalled={handleStalled}
                                    autoPlay={autoPlay}
                                    muted={muted}
                                    playsInline
                                    controls={false}
                                    style={{
                                        border: "none",
                                        width: "100%",
                                        height: "auto",
                                        filter: (settings.videoEnhancement.enabled && beautifyImage)
                                            ? `contrast(${settings.videoEnhancement.contrast}) saturate(${settings.videoEnhancement.saturation}) brightness(${settings.videoEnhancement.brightness})`
                                            : "none",
                                        imageRendering: "crisp-edges",
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
                            </div>

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

                            {isPip && <div className="absolute top-0 left-0 w-full h-full z-[100] bg-black flex items-center justify-center">
                                <Button
                                    intent="gray-outline" size="xl" onClick={() => {
                                    pipManager?.togglePip()
                                }}
                                >
                                    Exit PiP
                                </Button>
                            </div>}

                            <VideoCoreControlBar
                                timeRange={<VideoCoreTimeRange
                                    chapterCues={chapterCues ?? []}
                                />}
                            >

                                <VideoCorePlayButton />

                                <VideoCorePreviousButton onClick={() => { }} />
                                <VideoCoreNextButton onClick={() => { }} />

                                <VideoCoreVolumeButton />

                                <VideoCoreTimestamp />

                                <div className="flex flex-1" />

                                {!isMiniPlayer && <TorrentStreamOverlay isNativePlayerComponent="control-bar" />}

                                <VideoCoreSettingsButton />

                                <VideoCoreAudioButton />

                                <VideoCoreSubtitleButton />

                                <VideoCorePipButton />

                                <VideoCoreFullscreenButton />


                            </VideoCoreControlBar>

                        </>
                    ) : (
                        <div
                            className="w-full h-full absolute flex justify-center items-center flex-col space-y-4 bg-black rounded-md"
                        >

                            {/* {!state.miniPlayer && <SquareBg className="absolute top-0 left-0 w-full h-full z-[0]" />} */}
                            <FloatingButtons part="loading" onTerminateStream={onTerminateStream} />

                            {state.loadingState && <LoadingSpinner
                                title={state.loadingState || "Loading..."}
                                spinner={<PiSpinnerDuotone className="size-20 text-white animate-spin" />}
                                containerClass="z-[1]"
                            />}
                        </div>
                    )}


                </div>
            </VideoCoreDrawer>
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
                    className="rounded-full absolute top-0 flex-none right-4 z-[999]"
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
                        "rounded-full text-2xl flex-none absolute z-[999] right-4 top-4 pointer-events-auto bg-black/30 hover:bg-black/40",
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
                        "rounded-full text-2xl flex-none absolute z-[999] left-4 top-4 pointer-events-auto",
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
