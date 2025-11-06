import { getServerBaseUrl } from "@/api/client/server-url"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import {
    useCancelDiscordActivity,
    useSetDiscordAnimeActivityWithProgress,
    useUpdateDiscordAnimeActivityWithProgress,
} from "@/api/hooks/discord.hooks"
import { NativePlayerState } from "@/app/(main)/_features/native-player/native-player.atoms"
import { AniSkipTime } from "@/app/(main)/_features/sea-media-player/aniskip"
import { vc_doFlashAction, VideoCoreActionDisplay } from "@/app/(main)/_features/video-core/video-core-action-display"
import { vc_anime4kOption, VideoCoreAnime4K } from "@/app/(main)/_features/video-core/video-core-anime-4k"
import { Anime4KOption, VideoCoreAnime4KManager } from "@/app/(main)/_features/video-core/video-core-anime-4k-manager"
import { VideoCoreAudioManager } from "@/app/(main)/_features/video-core/video-core-audio"
import {
    vc_hoveringControlBar,
    VideoCoreAudioButton,
    VideoCoreControlBar,
    VideoCoreFullscreenButton,
    VideoCorePipButton,
    VideoCorePlayButton,
    VideoCoreSettingsButton,
    VideoCoreSubtitleButton,
    VideoCoreTimestamp,
    VideoCoreVolumeButton,
} from "@/app/(main)/_features/video-core/video-core-control-bar"
import { VideoCoreDrawer } from "@/app/(main)/_features/video-core/video-core-drawer"
import { vc_fullscreenManager, VideoCoreFullscreenManager } from "@/app/(main)/_features/video-core/video-core-fullscreen"
import { VideoCoreKeybindingController, VideoCoreKeybindingsModal } from "@/app/(main)/_features/video-core/video-core-keybindings"
import { vc_mediaSessionManager, VideoCoreMediaSessionManager } from "@/app/(main)/_features/video-core/video-core-media-session"
import { vc_menuOpen } from "@/app/(main)/_features/video-core/video-core-menu"
import { vc_pipElement, vc_pipManager, VideoCorePipManager } from "@/app/(main)/_features/video-core/video-core-pip"
import { useVideoCorePlaylist, useVideoCorePlaylistSetup, VideoCorePlaylistControl } from "@/app/(main)/_features/video-core/video-core-playlist"
import { VideoCorePreviewManager } from "@/app/(main)/_features/video-core/video-core-preview"
import { VideoCoreSubtitleManager } from "@/app/(main)/_features/video-core/video-core-subtitles"
import { VideoCoreTimeRange } from "@/app/(main)/_features/video-core/video-core-time-range"
import { VideoCoreTopPlaybackInfo, VideoCoreTopSection } from "@/app/(main)/_features/video-core/video-core-top-section"
import {
    vc_autoNextAtom,
    vc_autoPlayVideoAtom,
    vc_autoSkipOPEDAtom,
    vc_beautifyImageAtom,
    vc_settings,
    vc_storedMutedAtom,
    vc_storedPlaybackRateAtom,
    vc_storedVolumeAtom,
} from "@/app/(main)/_features/video-core/video-core.atoms"
import {
    detectSubtitleType,
    isSubtitleFile,
    useVideoCoreBindings,
    vc_createChapterCues,
    vc_createChaptersFromAniSkip,
    vc_formatTime,
    vc_logGeneralInfo,
} from "@/app/(main)/_features/video-core/video-core.utils"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { __torrentSearch_selectedTorrentsAtom } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-container"
import {
    __torrentSearch_selectionAtom,
    __torrentSearch_selectionEpisodeAtom,
} from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { TorrentStreamOverlay } from "@/app/(main)/entry/_containers/torrent-stream/torrent-stream-overlay"
import { GradientBackground } from "@/components/shared/gradient-background"
import { LuffyError } from "@/components/shared/luffy-error"
import { Button, IconButton } from "@/components/ui/button"
import { useUpdateEffect } from "@/components/ui/core/hooks"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { logger } from "@/lib/helpers/debug"
import { __isDesktop__ } from "@/types/constants"
import { useQueryClient } from "@tanstack/react-query"
import { atom, useAtomValue } from "jotai"
import { derive } from "jotai-derive"
import { createIsolation, ScopeProvider } from "jotai-scope"
import { useAtom, useSetAtom } from "jotai/react"
import React, { useCallback, useEffect, useMemo, useRef } from "react"
import { BiExpand, BiX } from "react-icons/bi"
import { FiMinimize2 } from "react-icons/fi"
import { ImSpinner2 } from "react-icons/im"
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
export const vc_cursorBusy = derive([vc_hoveringControlBar, vc_menuOpen], (f1, f2) => {
    return f1 || !!f2
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

export const vc_lastKnownProgress = atom<{ mediaId: number, progressNumber: number, time: number } | null>(null)

export const vc_skipOpeningTime = atom<number | null>(null)
export const vc_skipEndingTime = atom<number | null>(null)

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
                if (isNaN(duration) || duration <= 1) return
                t = Math.min(duration, Math.max(0, action.payload.time))
                videoElement.currentTime = t
                set(vc_currentTime, t)
                if (action.payload.flashTime) {
                    set(vc_doFlashAction, { message: `${vc_formatTime(t)} / ${vc_formatTime(duration)}`, type: "message" })
                }
                break
            case "seek":
                if (isNaN(duration) || duration <= 1) return
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
                if (action.payload) {
                    set(vc_lastKnownProgress, {
                        mediaId: action.payload.mediaId,
                        progressNumber: action.payload.progressNumber,
                        time: action.payload.time,
                    })
                } else {
                    set(vc_lastKnownProgress, null)
                }
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
    const serverStatus = useServerStatus()

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
    const [mediaSessionManager, setMediaSessionManager] = useAtom(vc_mediaSessionManager)
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
    const cursorBusy = useAtomValue(vc_cursorBusy)

    const [skipOpeningTime, setSkipOpeningTime] = useAtom(vc_skipOpeningTime)
    const [skipEndingTime, setSkipEndingTime] = useAtom(vc_skipEndingTime)

    const [autoNext] = useAtom(vc_autoNextAtom)
    const [autoPlay] = useAtom(vc_autoPlayVideoAtom)
    const [autoSkipOpeningOutro] = useAtom(vc_autoSkipOPEDAtom)
    const [volume] = useAtom(vc_storedVolumeAtom)
    const [muted] = useAtom(vc_storedMutedAtom)
    const [playbackRate, setPlaybackRate] = useAtom(vc_storedPlaybackRateAtom)

    const { mutate: setAnimeDiscordActivity } = useSetDiscordAnimeActivityWithProgress()
    const { mutate: updateAnimeDiscordActivity } = useUpdateDiscordAnimeActivityWithProgress()
    const { mutate: cancelDiscordActivity } = useCancelDiscordActivity()

    React.useEffect(() => {
        const interval = setInterval(() => {
            if (!videoRef.current) return

            if (serverStatus?.settings?.discord?.enableRichPresence && serverStatus?.settings?.discord?.enableAnimeRichPresence) {
                updateAnimeDiscordActivity({
                    progress: Math.floor(videoRef.current?.currentTime ?? 0),
                    duration: Math.floor(videoRef.current?.duration ?? 0),
                    paused: videoRef.current?.paused ?? false,
                })
            }
        }, 6000)

        return () => clearInterval(interval)
    }, [serverStatus?.settings?.discord, videoRef.current])

    const [measureRef, { width, height }] = useMeasure<HTMLVideoElement>()
    React.useEffect(() => {
        setRealVideoSize({
            width,
            height,
        })
    }, [width, height])

    const qc = useQueryClient()
    React.useEffect(() => {
        qc.invalidateQueries({ queryKey: [API_ENDPOINTS.CONTINUITY.GetContinuityWatchHistory.key] })
        qc.invalidateQueries({ queryKey: [API_ENDPOINTS.CONTINUITY.GetContinuityWatchHistoryItem.key] })

        if (!state.playbackInfo || state.playbackInfo.id !== currentPlaybackRef.current) {
            cancelDiscordActivity()
        }
    }, [state.playbackInfo])


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
            cancelDiscordActivity()
            if (videoRef.current) {
                videoRef.current.pause()
                videoRef.current.removeAttribute("src")
                videoRef.current.load()
                videoRef.current = null
            }
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
            setRestoreProgressTo(null)
            if (mediaSessionManager) {
                mediaSessionManager.setVideo(null)
                mediaSessionManager.destroy()
            }
            setMediaSessionManager(null)
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

        setSkipOpeningTime(null)
        setSkipEndingTime(null)

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

        // Initialize media session manager
        setMediaSessionManager(p => {
            if (p) p.destroy()
            const manager = new VideoCoreMediaSessionManager()

            manager.on("media-play-request", () => {
                if (videoRef.current?.paused) {
                    videoRef.current.play().catch(err => {
                        log.error("Failed to play video from media session", err)
                    })
                }
            })

            manager.on("media-pause-request", () => {
                if (videoRef.current && !videoRef.current.paused) {
                    videoRef.current.pause()
                }
            })

            manager.on("media-seek-request", (event: Event) => {
                const { seekTime } = (event as CustomEvent).detail
                action({ type: "seekTo", payload: { time: seekTime } })
            })
            manager.setPlaybackInfo(state.playbackInfo!)
            manager.setVideo(v!)
            return manager
        })

        log.info("Initializing preview manager")
        setPreviewManager(p => {
            if (p) p.cleanup()
            return new VideoCorePreviewManager(v!, streamUrl)
        })

        if (
            serverStatus?.settings?.discord?.enableRichPresence &&
            serverStatus?.settings?.discord?.enableAnimeRichPresence &&
            !!state.playbackInfo?.media?.id &&
            !!state.playbackInfo?.episode?.progressNumber
        ) {
            const media = state.playbackInfo.media
            const videoProgress = videoRef.current?.currentTime ?? 0
            const videoDuration = videoRef.current?.duration ?? 0

            log.info("Setting discord activity", {
                videoProgress,
                videoDuration,
            })
            setAnimeDiscordActivity({
                mediaId: media?.id ?? 0,
                title: media?.title?.userPreferred || media?.title?.romaji || media?.title?.english || "Watching",
                image: media?.coverImage?.large || media?.coverImage?.medium || "",
                isMovie: media?.format === "MOVIE",
                episodeNumber: state.playbackInfo?.episode?.progressNumber ?? 0,
                progress: Math.floor(videoProgress),
                duration: Math.floor(videoDuration),
                totalEpisodes: media?.episodes,
                currentEpisodeCount: media?.nextAiringEpisode?.episode ? media?.nextAiringEpisode?.episode - 1 : media?.episodes,
                episodeTitle: state.playbackInfo.episode.episodeTitle || undefined,
            })
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
            onCompleted?.()
        }
    }

    const { playEpisode } = useVideoCorePlaylist()
    const handleEnded = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        log.info("Video ended")
        onEnded?.()
        if (autoNext) {
            // videoRef?.current?.pause()
            playEpisode("next")
        }
    }

    const [debouncedMenuOpen, setDebouncedMenuOpen] = React.useState(false)
    const menuOpen = useAtomValue(vc_menuOpen)
    React.useEffect(() => {
        if (!!menuOpen) {
            setDebouncedMenuOpen(true)
            return
        }
        let t = setTimeout(() => {
            setDebouncedMenuOpen(false)
        }, 800)
        return () => {
            clearTimeout(t)
        }
    }, [menuOpen])

    let lastClickTime = React.useRef(0)

    const handleClick = (e: React.SyntheticEvent<HTMLDivElement>) => {
        log.info("Video clicked")
        // check if right click
        if (e.type === "click") {
            if (!debouncedMenuOpen) {
                togglePlay()
            }
            setTimeout(() => {
                setBusy(false)
            }, 100)
        }

        if (e.type === "contextmenu") {
            const now = Date.now()
            if (lastClickTime.current && now - lastClickTime.current < 500) {
                fullscreenManager?.toggleFullscreen()
            }
            lastClickTime.current = now
        }
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
        if (!videoRef.current) return
        if (autoPlay) {
            videoRef.current.play().catch()
        }
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

    React.useEffect(() => {
        if (mediaSessionManager && videoRef.current && state.playbackInfo && state.active) {
            mediaSessionManager.setVideo(videoRef.current)
            mediaSessionManager.setPlaybackInfo(state.playbackInfo)
            mediaSessionManager.activate()
        } else if (mediaSessionManager && !state.active) {
            log.info("Video inactive, deactivating media session")
            mediaSessionManager.deactivate()
        }
    }, [mediaSessionManager, videoRef.current, state.playbackInfo, state.active, playEpisode])

    React.useEffect(() => {
        return () => {
            log.info("VideoCore unmounting, cleaning up media session")
            if (mediaSessionManager) {
                mediaSessionManager.setVideo(null)
                mediaSessionManager.destroy()
            }
        }
    }, [mediaSessionManager])

    //

    // container events
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
            if (!duration || duration <= 1) return []
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
        toast.info("Adding subtitle file...")
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
        const player = videoRef.current
        if (!player || !state.active) return

        player.addEventListener("paste", handleUpload)
        player.addEventListener("drop", handleUpload)
        player.addEventListener("dragover", suppressEvent)
        player.addEventListener("dragenter", suppressEvent)

        return () => {
            player.removeEventListener("paste", handleUpload)
            player.removeEventListener("drop", handleUpload)
            player.removeEventListener("dragover", suppressEvent)
            player.removeEventListener("dragenter", suppressEvent)
        }
    }, [handleUpload, state.active, videoRef.current])

    /**
     * Restore last position
     */
    const [restoreProgressTo, setRestoreProgressTo] = useAtom(vc_lastKnownProgress)
    React.useEffect(() => {
        if (!anime4kManager || !restoreProgressTo) return

        if (restoreProgressTo.mediaId !== state.playbackInfo?.media?.id || restoreProgressTo.progressNumber !== state.playbackInfo?.episode?.progressNumber) {
            setRestoreProgressTo(null)
            return
        }

        if (anime4kOption === "off" || anime4kManager.canvas !== null) {
            flashAction({ message: "Progress restored", duration: 1500 })
            dispatchAction({ type: "seekTo", payload: { time: restoreProgressTo.time } })
        } else if (anime4kOption !== ("off" as Anime4KOption)) {
            flashAction({ message: "Restoring progress", duration: 1500 })
            anime4kManager.registerOnCanvasCreatedOnce(() => {
                dispatchAction({ type: "seekTo", payload: { time: restoreProgressTo.time } })
            })
        }
        setRestoreProgressTo(null)
        if (autoPlay) {
            videoRef.current?.play()
        }

    }, [anime4kManager, anime4kOption, restoreProgressTo, autoPlay])

    return (
        <>
            <ScopeProvider atoms={[__torrentSearch_selectionAtom, __torrentSearch_selectionEpisodeAtom, __torrentSearch_selectedTorrentsAtom]}>

                <VideoCoreAnime4K />
                <VideoCoreKeybindingsModal />
                {state.active && !isMiniPlayer && <RemoveScrollBar />}

                <TorrentStreamOverlay isNativePlayerComponent="overlay" show={(state.active && isMiniPlayer)} />

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
                                fullscreenManager?.exitFullscreen()
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
                    <TorrentStreamOverlay
                        isNativePlayerComponent="top-section"
                        show={!isMiniPlayer && !(!!state.playbackInfo?.streamUrl && !state.loadingState)}
                    />

                    {(state?.playbackError) && (
                        <div className="h-full w-full bg-black/100 flex items-center justify-center z-[20] absolute p-4">
                            <div className="text-white text-center">
                                {!isMiniPlayer ? (
                                    <LuffyError title="Playback Error" />
                                ) : (
                                    <h1 className={cn("text-2xl font-bold", isMiniPlayer && "text-lg")}>Playback Error</h1>
                                )}
                                <p className={cn("text-base text-white/50 max-w-xl", isMiniPlayer && "text-sm max-w-lg mx-auto")}>
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
                                    {!!skipOpeningTime && !isMiniPlayer && (
                                        <div className="absolute left-5 bottom-28 z-[60] native-player-hide-on-fullscreen">
                                            <Button
                                                size="sm"
                                                intent="gray-basic"
                                                onClick={e => {
                                                    e.stopPropagation()
                                                    action({ type: "seekTo", payload: { time: skipOpeningTime || 0 } })
                                                }}
                                                onPointerMove={e => e.stopPropagation()}
                                            >
                                                Skip Opening
                                            </Button>
                                        </div>
                                    )}

                                    {!!skipEndingTime && !isMiniPlayer && (
                                        <div className="absolute right-5 bottom-28 z-[60] native-player-hide-on-fullscreen">
                                            <Button
                                                size="sm"
                                                intent="gray-basic"
                                                onClick={e => {
                                                    e.stopPropagation()
                                                    action({ type: "seekTo", payload: { time: skipEndingTime || 0 } })
                                                }}
                                                onPointerMove={e => e.stopPropagation()}
                                            >
                                                Skip Ending
                                            </Button>
                                        </div>
                                    )}
                                </>}

                                <div
                                    className="relative w-full h-full flex items-center justify-center"
                                    onClick={handleClick}
                                    onContextMenu={handleClick}
                                >
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
                                            height: "100%",
                                            objectFit: "contain",
                                            objectPosition: "center",
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

                                    <VideoCorePlaylistControl />

                                    <VideoCoreVolumeButton />

                                    <VideoCoreTimestamp />

                                    <div className="flex flex-1" />

                                    <TorrentStreamOverlay isNativePlayerComponent="control-bar" show={!isMiniPlayer} />

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
                                    // spinner={<PiSpinnerDuotone className="size-20 text-white animate-spin" />}
                                    spinner={<ImSpinner2 className="size-20 text-white animate-spin" />}
                                    containerClass="z-[1]"
                                />}

                                {!isMiniPlayer && <div className="opacity-50 absolute inset-0 z-[0] overflow-hidden">
                                    <GradientBackground
                                        duration={10} breathingRange={5}
                                        // gradientColors={[
                                        //     "transparent",
                                        //     "#312887",
                                        //     "#3D5AFE",
                                        // ]} gradientStops={[35, 50, 100]}
                                    />
                                </div>}
                            </div>
                        )}


                    </div>
                </VideoCoreDrawer>

            </ScopeProvider>
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

