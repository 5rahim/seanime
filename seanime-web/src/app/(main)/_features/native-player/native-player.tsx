import { getServerBaseUrl } from "@/api/client/server-url"
import { MKVParser_SubtitleEvent, MKVParser_TrackInfo, NativePlayer_PlaybackInfo, NativePlayer_ServerEvent } from "@/api/generated/types"
import { NativePlayerIcons } from "@/app/(main)/_features/native-player/native-player-icons"
import {
    __seaMediaPlayer_autoNextAtom,
    __seaMediaPlayer_autoPlayAtom,
    __seaMediaPlayer_autoSkipIntroOutroAtom,
    __seaMediaPlayer_discreteControlsAtom,
    __seaMediaPlayer_mutedAtom,
    __seaMediaPlayer_volumeAtom,
} from "@/app/(main)/_features/sea-media-player/sea-media-player.atoms"
import { clientIdAtom } from "@/app/websocket-provider"
import { LuffyError } from "@/components/shared/luffy-error"
import { IconButton } from "@/components/ui/button"
import { useUpdateEffect } from "@/components/ui/core/hooks"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { logger } from "@/lib/helpers/debug"
import { WSEvents } from "@/lib/server/ws-events"
import { __isDesktop__ } from "@/types/constants"
import { useAtom, useAtomValue } from "jotai"
import {
    MediaControlBar,
    MediaController,
    MediaErrorDialog,
    MediaFullscreenButton,
    MediaMuteButton,
    MediaPipButton,
    MediaPlayButton,
    MediaPreviewChapterDisplay,
    MediaPreviewThumbnail,
    MediaPreviewTimeDisplay,
    MediaTimeDisplay,
    MediaTimeRange,
    MediaVolumeRange,
} from "media-chrome/react"
import { MediaProvider } from "media-chrome/react/media-store"
import {
    MediaAudioTrackMenu,
    MediaAudioTrackMenuButton,
    MediaCaptionsMenu,
    MediaCaptionsMenuButton,
    MediaPlaybackRateMenu,
    MediaRenditionMenu,
    MediaSettingsMenu,
    MediaSettingsMenuButton,
    MediaSettingsMenuItem,
} from "media-chrome/react/menu"
import React, { FormEvent, useCallback, useEffect, useMemo, useRef, useState } from "react"
import { BiExpand, BiX } from "react-icons/bi"
import { FiMinimize2 } from "react-icons/fi"
import { PiSpinnerDuotone } from "react-icons/pi"
import { useWebsocketMessageListener, useWebsocketSender } from "../../_hooks/handle-websockets"
import { StreamAudioManager, StreamSubtitleManager } from "./handle-native-player"
import { NativePlayerDrawer } from "./native-player-drawer"
import {
    FlashNotificationDisplay,
    NativePlayerKeybindingController,
    NativePlayerKeybindingsModal,
    nativePlayerKeybindingsModalAtom,
} from "./native-player-keybindings"
import { StreamPreviewCaptureIntervalSeconds, StreamPreviewManager } from "./native-player-preview"
import { nativePlayer_settingsAtom, nativePlayer_stateAtom, nativePlayerKeybindingsAtom } from "./native-player.atoms"
import { detectSubtitleType, isSubtitleFile, nativeplayer_createChapterCues, nativeplayer_createChapterVTT } from "./native-player.utils"

const enum VideoPlayerEvents {
    LOADED_METADATA = "loaded-metadata",
    VIDEO_SEEKED = "video-seeked",
    SUBTITLE_FILE_UPLOADED = "subtitle-file-uploaded",
    VIDEO_PAUSED = "video-paused",
    VIDEO_RESUMED = "video-resumed",
    VIDEO_ENDED = "video-ended",
    VIDEO_ERROR = "video-error",
    VIDEO_CAN_PLAY = "video-can-play",
    VIDEO_STARTED = "video-started",
    VIDEO_COMPLETED = "video-completed",
}

const log = logger("NATIVE PLAYER")

export function NativePlayer() {
    const clientId = useAtomValue(clientIdAtom)
    const { sendMessage } = useWebsocketSender()
    //
    // Player
    //
    // The player reference
    const videoRef = useRef<HTMLVideoElement | null>(null)
    const videoCompletedRef = useRef(false)
    const playerContainerRef = useRef<HTMLDivElement | null>(null)
    const timeRangeRef = useRef<any>(null)

    //
    // Control settings
    //
    const [autoPlay, setAutoPlay] = useAtom(__seaMediaPlayer_autoPlayAtom)
    const [autoNext, setAutoNext] = useAtom(__seaMediaPlayer_autoNextAtom)
    const [discreteControls, setDiscreteControls] = useAtom(__seaMediaPlayer_discreteControlsAtom)
    const [autoSkipIntroOutro, setAutoSkipIntroOutro] = useAtom(__seaMediaPlayer_autoSkipIntroOutroAtom)
    const [volume, setVolume] = useAtom(__seaMediaPlayer_volumeAtom)
    const [muted, setMuted] = useAtom(__seaMediaPlayer_mutedAtom)

    const [settings, setSettings] = useAtom(nativePlayer_settingsAtom)

    // The state
    const [state, setState] = useAtom(nativePlayer_stateAtom)
    const [duration, setDuration] = useState(0)

    // Keybindings
    const keybindings = useAtomValue(nativePlayerKeybindingsAtom)

    const streamLoadedRef = useRef<string | null>(null)
    const subtitleManagerRef = useRef<StreamSubtitleManager | null>(null)
    const audioManagerRef = useRef<StreamAudioManager | null>(null)
    const previewManagerRef = useRef<StreamPreviewManager | null>(null)

    // Handle thumbnail preview updates
    const [previewThumbnail, setPreviewThumbnail] = useState<string | undefined>(undefined)

    // Create chapter track
    const [chapterTrackUrl, setChapterTrackUrl] = useState<string | null>(null)

    useEffect(() => {
        if (state.playbackInfo?.mkvMetadata?.chapters && duration > 0) {
            // Create VTT content for chapters
            const chapters = state.playbackInfo.mkvMetadata.chapters
            const vttContent = nativeplayer_createChapterVTT(chapters, duration)

            const blob = new Blob([vttContent], { type: "text/vtt" })
            const url = URL.createObjectURL(blob)
            setChapterTrackUrl(url)

            return () => {
                URL.revokeObjectURL(url)
            }
        }
    }, [state.playbackInfo?.mkvMetadata?.chapters, duration])

    //
    // Start
    //

    useUpdateEffect(() => {
        if (!!state.playbackInfo && (!streamLoadedRef.current || state.playbackInfo.id !== streamLoadedRef.current)) {
            if (videoRef.current) {
                // log.info("Stream loaded")
                // log.info("Can play", videoRef.current.canPlayType(state.playbackInfo.mkvMetadata?.mimeCodec || ""))

                console.log("HEVC HVC1 main profile support ->", videoRef.current.canPlayType("video/mp4;codecs=\"hvc1\""))
                console.log("HEVC main profile support ->", videoRef.current.canPlayType("video/mp4;codecs=\"hev1.1.6.L120.90\""))
                console.log("HEVC main 10 profile support ->", videoRef.current.canPlayType("video/mp4;codecs=\"hev1.2.4.L120.90\""))
                console.log("HEVC main still-picture profile support ->", videoRef.current.canPlayType("video/mp4;codecs=\"hev1.3.E.L120.90\""))
                console.log("HEVC range extensions profile support ->", videoRef.current.canPlayType("video/mp4;codecs=\"hev1.4.10.L120.90\""))
                console.log("Dolby AC3 support ->", videoRef.current.canPlayType("audio/mp4; codecs=\"ac-3\""))
                console.log("Dolby EC3 support ->", videoRef.current.canPlayType("audio/mp4; codecs=\"ec-3\""))


                // if (!!streamLoadedRef.current && state.playbackInfo.id !== streamLoadedRef.current) {
                //     log.info("Stream changed")
                // }


                // subtitleManagerRef.current?.loadTracks()
            }
        }


        if (!state.playbackInfo && streamLoadedRef.current) {
            log.info("Stream unloaded")
            subtitleManagerRef.current?.terminate()
            previewManagerRef.current?.cleanup()
            previewManagerRef.current = null
            setPreviewThumbnail(undefined)
            streamLoadedRef.current = null

        }
    }, [state.playbackInfo, videoRef.current])

    // Clean up player when unmounting or changing streams
    useEffect(() => {
        if (!videoRef.current) return

        return () => {
            if (videoRef.current) {
                log.info("Cleaning up player")
                // videoRef.current = null
            }
        }
    }, [state.playbackInfo?.streamUrl])

    // Handle volume changes
    useEffect(() => {
        if (videoRef.current) {
            videoRef.current.volume = volume
        }
    }, [volume])

    // Handle mute changes
    useEffect(() => {
        if (videoRef.current) {
            videoRef.current.muted = muted
        }
    }, [muted])

    useEffect(() => {
        if (state.active && videoRef.current && state.playbackInfo) {
            // Small delay to ensure the video element is fully rendered
            setTimeout(() => {
                videoRef.current?.focus()
            }, 100)
        }
    }, [state.active, state.playbackInfo])

    //
    // Functions
    //

    function seekTo(time: number) {
        if (videoRef.current) {
            videoRef.current.currentTime = time
        }
    }

    function seek(offset: number) {
        if (videoRef.current) {
            const newTime = videoRef.current.currentTime + offset
            videoRef.current.currentTime = newTime
        }
    }

    //
    // Event Handlers
    //
    const handleTimeUpdate = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        // log.info("Time update", e.currentTarget.currentTime)
        const percent = e.currentTarget.currentTime / e.currentTarget.duration
        if (!!e.currentTarget.duration && !videoCompletedRef.current && percent >= 0.8) {
            videoCompletedRef.current = true
            sendMessage({
                type: WSEvents.NATIVE_PLAYER,
                payload: {
                    clientId: clientId,
                    type: VideoPlayerEvents.VIDEO_COMPLETED,
                },
            })
        }
    }

    const handleDurationChange = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        log.info("Duration change", e.currentTarget.duration)
    }

    const handleCanPlay = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        setDuration(videoRef.current?.duration ?? 0)
    }

    const handleEnded = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        log.info("Ended")

        sendMessage({
            type: WSEvents.NATIVE_PLAYER,
            payload: {
                clientId: clientId,
                type: VideoPlayerEvents.VIDEO_ENDED,
                payload: {
                    autoNext: autoNext,
                },
            },
        })

    }

    const handleError = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        log.info("Media error", e.currentTarget.error)
        sendMessage({
            type: WSEvents.NATIVE_PLAYER,
            payload: {
                clientId: clientId,
                type: VideoPlayerEvents.VIDEO_ERROR,
                payload: { error: e.currentTarget.error?.message || "Unknown error" },
            },
        })
    }

    const handleVolumeChange = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        setVolume(e.currentTarget.volume)
        setMuted(e.currentTarget.muted)
    }

    const handleSeeked = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        const currentTime = e.currentTarget.currentTime
        log.info("Video seeked to", currentTime)

        sendMessage({
            type: WSEvents.NATIVE_PLAYER,
            payload: {
                clientId: clientId,
                type: VideoPlayerEvents.VIDEO_SEEKED,
                payload: { currentTime: currentTime },
            },
        })

    }

    const handleLoadedMetadata = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        log.info("Loaded metadata", e.currentTarget.duration)
        log.info("Audio tracks", videoRef.current?.audioTracks)
        log.info("Text tracks", videoRef.current?.textTracks)

        videoCompletedRef.current = false

        if (!state.playbackInfo || !videoRef.current) return // shouldn't happen

        streamLoadedRef.current = state.playbackInfo.id

        // Initialize the subtitle manager if the stream is MKV
        if (!!state.playbackInfo?.mkvMetadata) {
            subtitleManagerRef.current = new StreamSubtitleManager({
                videoElement: videoRef.current,
                playbackInfo: state.playbackInfo,
                jassubOffscreenRender: true,
                settings: settings,
            })

            audioManagerRef.current = new StreamAudioManager({
                videoElement: videoRef.current,
                playbackInfo: state.playbackInfo,
                settings: settings,
                onError: (error) => {
                    log.error("Audio manager error", error)
                    setState(draft => {
                        draft.playbackError = error
                        return
                    })
                },
            })
        }

        // Initialize thumbnailer
        if (state.playbackInfo?.streamUrl) {
            const streamUrl = state.playbackInfo.streamUrl.replace("{{SERVER_URL}}", getServerBaseUrl())
            log.info("Initializing thumbnailer with URL:", streamUrl)
            previewManagerRef.current = new StreamPreviewManager(videoRef.current, streamUrl)
            log.info("Thumbnailer initialized successfully")
        } else {
            log.info("No stream URL available for thumbnailer")
        }

        sendMessage({
            type: WSEvents.NATIVE_PLAYER,
            payload: {
                clientId: clientId,
                type: VideoPlayerEvents.LOADED_METADATA,
                payload: {},
            },
        })
    }

    const handlePause = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        log.info("Pause")

        sendMessage({
            type: WSEvents.NATIVE_PLAYER,
            payload: {
                clientId: clientId,
                type: VideoPlayerEvents.VIDEO_PAUSED,
                payload: {},
            },
        })
    }

    const handlePlay = (e: React.SyntheticEvent<HTMLVideoElement>) => {
        log.info("Play/Resume")

        sendMessage({
            type: WSEvents.NATIVE_PLAYER,
            payload: {
                clientId: clientId,
                type: VideoPlayerEvents.VIDEO_RESUMED,
                payload: {},
            },
        })
    }

    type UploadEvent = {
        dataTransfer?: DataTransfer
        clipboardData?: DataTransfer
    }
    const handleUpload = useCallback(async (e: UploadEvent & Event) => {
        e.preventDefault() // stop the default behavior
        log.info("Upload", e)
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
                    sendMessage({
                        type: WSEvents.NATIVE_PLAYER,
                        payload: {
                            clientId: clientId,
                            type: VideoPlayerEvents.SUBTITLE_FILE_UPLOADED,
                            payload: { filename: f.name, content },
                        },
                    })
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
                    if (type === "unknown") {
                        log.info("Unknown subtitle type, skipping")
                        return
                    }
                    const filename = `PLACEHOLDER.${type}`
                    sendMessage({
                        type: WSEvents.NATIVE_PLAYER,
                        payload: {
                            clientId: clientId,
                            type: VideoPlayerEvents.SUBTITLE_FILE_UPLOADED,
                            payload: { filename, content: str },
                        },
                    })
                })
            }
        }
    }, [clientId, sendMessage])

    function suppressEvent(e: Event) {
        e.preventDefault()
    }

    useEffect(() => {
        const playerContainer = playerContainerRef.current
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

    //
    // Server events
    //

    useWebsocketMessageListener({
        type: WSEvents.NATIVE_PLAYER,
        onMessage: ({ type, payload }: { type: NativePlayer_ServerEvent, payload: unknown }) => {
            switch (type) {
                // 1. Open and await
                // The server is loading the stream
                case "open-and-await":
                    log.info("Open and await event received", { payload })
                    setState(draft => {
                        draft.active = true
                        draft.miniPlayer = false
                        draft.loadingState = payload as string
                        draft.playbackInfo = null
                        draft.playbackError = null
                        return
                    })

                    break
                // 2. Watch
                // We received the playback info
                case "watch":
                    log.info("Watch event received", { payload })
                    setState(draft => {
                        draft.playbackInfo = payload as NativePlayer_PlaybackInfo
                        draft.loadingState = null
                        draft.playbackError = null
                        return
                    })
                    break
                // 3. Subtitle event (MKV)
                // We receive the subtitle events after the server received the loaded-metadata event
                case "subtitle-event":
                    subtitleManagerRef.current?.onSubtitleEvent(payload as MKVParser_SubtitleEvent)
                    break
                case "add-subtitle-track":
                    subtitleManagerRef.current?.onTrackAdded(payload as MKVParser_TrackInfo)
                    break
            }
        },
    })

    //
    // Handlers
    //

    function handleTerminateStream() {
        // Clean up player first
        if (videoRef.current) {
            log.info("Cleaning up media")
            videoRef.current.pause()
        }

        setState(draft => {
            draft.miniPlayer = true
            draft.playbackInfo = null
            draft.playbackError = null
            draft.loadingState = "Ending stream..."
            return
        })

        setTimeout(() => {
            setState(draft => {
                draft.active = false
                return
            })
        }, 1000)
        // Send terminate stream event
    }

    function onCaptionsChange(e: FormEvent<any>) {
        log.info("Subtitles changed", e, videoRef.current?.textTracks)
        if (videoRef.current) {
            let trackFound = false
            for (let i = 0; i < videoRef.current.textTracks.length; i++) {
                const track = videoRef.current.textTracks[i]
                if (track.mode === "showing") {
                    subtitleManagerRef.current?.selectTrack(Number(track.id))
                    trackFound = true
                }
            }
            if (!trackFound) {
                subtitleManagerRef.current?.setNoTrack()
            }
        }
    }

    function onAudioChange(e: FormEvent<any>) {
        log.info("Audio changed", e, videoRef.current?.audioTracks)
        if (videoRef.current?.audioTracks) {
            for (let i = 0; i < videoRef.current.audioTracks.length; i++) {
                const track = videoRef.current.audioTracks[i]
                if (track.enabled) {
                    audioManagerRef.current?.selectTrack(Number(track.id))
                    break
                }
            }
        }
        seek(-1)
    }

    const handleTimeRangePreview = useCallback(async (event: MouseEvent) => {
        if (!previewManagerRef.current || !duration) {
            return
        }

        const timeRange = timeRangeRef.current
        if (!timeRange) return

        // Calculate preview time based on mouse position
        const rect = timeRange.getBoundingClientRect()
        const x = event.clientX - rect.left
        const percentage = Math.max(0, Math.min(1, x / rect.width))
        const previewTime = percentage * duration

        if (previewTime >= 0 && previewTime <= duration) {
            const thumbnailIndex = Math.floor(previewTime / StreamPreviewCaptureIntervalSeconds)

            try {
                const thumbnail = await previewManagerRef.current.retrievePreviewForSegment(thumbnailIndex)
                if (thumbnail) {
                    timeRange.setAttribute("mediapreviewimage", thumbnail)
                    setPreviewThumbnail(thumbnail)
                }
            }
            catch (error) {
                log.error("Failed to get thumbnail", error)
            }
        }
    }, [duration])

    // Add event listener for preview events
    useEffect(() => {
        const timeRange = timeRangeRef.current
        if (!timeRange) {
            log.info("TimeRange ref not available")
            return
        }

        const handleMouseLeave = () => {
            timeRange.removeAttribute("mediapreviewimage")
            setPreviewThumbnail(undefined)
        }

        timeRange.addEventListener("mouseleave", handleMouseLeave)
        timeRange.addEventListener("mousemove", handleTimeRangePreview)

        return () => {
            timeRange.removeEventListener("mouseleave", handleMouseLeave)
            timeRange.removeEventListener("mousemove", handleTimeRangePreview)
        }
    }, [handleTimeRangePreview])

    const chapterCues = useMemo(() => {
        const cues = nativeplayer_createChapterCues(state.playbackInfo?.mkvMetadata?.chapters, duration)
        log.info("Chapter cues", cues)
        return cues
    }, [state.playbackInfo?.mkvMetadata?.chapters, duration])

    const [keybindingsModalOpen, setKeybindingsModalOpen] = useAtom(nativePlayerKeybindingsModalAtom)

    return (
        <>
            <NativePlayerKeybindingsModal />

            <NativePlayerDrawer
                open={state.active}
                onOpenChange={(v) => {
                    if (!v) {
                        if (state.playbackError) {
                            handleTerminateStream()
                            return
                        }
                        if (!state.miniPlayer) {
                            setState(draft => {
                                draft.miniPlayer = true
                            })
                        } else {
                            handleTerminateStream()
                        }
                    }
                }}
                borderToBorder
                miniPlayer={state.miniPlayer}
                size={state.miniPlayer ? "md" : "full"}
                side={state.miniPlayer ? "right" : "bottom"}
                contentClass={cn(
                    "p-0 m-0",
                    !state.miniPlayer && "h-full",
                )}
                allowOutsideInteraction={true}
                overlayClass={cn(
                    state.miniPlayer && "hidden",
                )}
                closeClass={cn(
                    "z-[99]",
                    __isDesktop__ && !state.miniPlayer && "top-8",
                    state.miniPlayer && "left-4",
                )}
                hideCloseButton
                data-native-player-drawer
            >

                {(state?.playbackError) && (
                    <div className="h-full w-full bg-black/80 flex items-center justify-center z-[50] absolute p-4">
                        <div className="text-white text-center">
                            {!state.miniPlayer ? (
                                <LuffyError title="Playback Error" />
                            ) : (
                                <h1 className={cn("text-2xl font-bold", state.miniPlayer && "text-lg")}>Playback Error</h1>
                            )}
                            <p className={cn("text-base text-white/50", state.miniPlayer && "text-sm")}>
                                {state.playbackError || "An error occurred while playing the stream. Please try again later."}
                            </p>
                        </div>
                    </div>
                )}


                <div
                    className="h-full w-full bg-black flex items-center z-[50]"
                    data-native-player-container
                    data-mini-player={state.miniPlayer}
                    tabIndex={-1}
                    ref={playerContainerRef}
                >
                    {(!!state.playbackInfo?.streamUrl && !state.loadingState) ? (
                        <MediaProvider>
                            <MediaController
                                className={cn(
                                    "w-full h-full",
                                    discreteControls && "discrete-controls",
                                )}
                                tabIndex={-1}
                            >

                                <NativePlayerKeybindingController
                                    {...{
                                        videoRef,
                                        chapterCues,
                                        seekTo,
                                        seek,
                                        setVolume,
                                        setMuted,
                                        volume,
                                        muted,
                                        subtitleManagerRef,
                                        audioManagerRef,
                                    }}
                                />

                                <FlashNotificationDisplay />

                                {!state.miniPlayer && <IconButton
                                    icon={<FiMinimize2 className="text-2xl" />}
                                    intent="gray-basic"
                                    className="rounded-full absolute top-8 right-4 native-player-hide-on-fullscreen"
                                    onClick={() => {
                                        setState(draft => {
                                            draft.miniPlayer = true
                                        })
                                    }}
                                />}

                                {state.miniPlayer && <>
                                    <IconButton
                                        type="button"
                                        intent="gray-basic"
                                        size="sm"
                                        className={cn(
                                            "rounded-full text-2xl flex-none absolute z-[99] right-4 top-4 pointer-events-auto native-player-hide-on-fullscreen",
                                            state.miniPlayer && "text-xl",
                                        )}
                                        icon={<BiExpand />}
                                        onClick={() => {
                                            setState(draft => {
                                                draft.miniPlayer = false
                                            })
                                        }}
                                    />
                                    <IconButton
                                        type="button"
                                        intent="alert-subtle"
                                        size="xs"
                                        className={cn(
                                            "rounded-full text-2xl flex-none absolute z-[99] left-4 top-4 pointer-events-auto native-player-hide-on-fullscreen",
                                            state.miniPlayer && "text-xl",
                                        )}
                                        icon={<BiX />}
                                        onClick={() => {
                                            handleTerminateStream()
                                        }}
                                    />
                                </>}

                                <video
                                    ref={videoRef}
                                    slot="media"
                                    src={state.playbackInfo?.streamUrl?.replace("{{SERVER_URL}}", getServerBaseUrl())}
                                    crossOrigin="anonymous"
                                    playsInline
                                    autoPlay={autoPlay}
                                    muted={muted}
                                    onTimeUpdate={handleTimeUpdate}
                                    onDurationChange={handleDurationChange}
                                    onEnded={handleEnded}
                                    onError={handleError}
                                    onVolumeChange={handleVolumeChange}
                                    onSeeked={handleSeeked}
                                    onLoadedMetadata={handleLoadedMetadata}
                                    onCanPlay={handleCanPlay}
                                    onPause={handlePause}
                                    onPlay={handlePlay}
                                    style={{
                                        width: "100%",
                                        height: "100%",
                                        border: "none",
                                        filter: settings.videoEnhancement.enabled
                                            ? `contrast(${settings.videoEnhancement.contrast}) saturate(${settings.videoEnhancement.saturation}) brightness(${settings.videoEnhancement.brightness})`
                                            : "none",
                                        imageRendering: "auto",
                                    }}
                                    tabIndex={0}
                                    className="outline-none"
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
                                    {chapterTrackUrl && (
                                        <track
                                            kind="chapters"
                                            src={chapterTrackUrl}
                                            default
                                            label="Chapters"
                                        />
                                    )}
                                </video>

                                <MediaErrorDialog slot="dialog" />

                                <div className="native-player-gradient-bottom"></div>

                                {!state.miniPlayer && <div
                                    className={cn(
                                        "top-8 absolute py-4 px-5",
                                        // state.miniPlayer && "hidden",
                                    )}
                                >
                                    <div className="">
                                        <p className="text-white font-bold text-lg">
                                            {state.playbackInfo?.episode?.displayTitle}
                                        </p>
                                        <p className="text-white/50 text-base !font-normal">
                                            {state.playbackInfo?.episode?.episodeTitle}
                                        </p>
                                    </div>
                                </div>}

                                <MediaSettingsMenu hidden anchor="auto">
                                    <MediaSettingsMenuItem>
                                        Playback Speed
                                        <MediaPlaybackRateMenu rates={[0.5, 0.75, 1, 1.10, 1.25, 1.5, 1.75, 2]} slot="submenu" hidden>
                                            <div slot="title">Playback Speed</div>
                                        </MediaPlaybackRateMenu>
                                    </MediaSettingsMenuItem>
                                    <MediaSettingsMenuItem className="quality-settings">
                                        Quality
                                        <MediaRenditionMenu slot="submenu" hidden>
                                            <div slot="title">Quality</div>
                                        </MediaRenditionMenu>
                                    </MediaSettingsMenuItem>
                                    <MediaSettingsMenuItem
                                        className="keyboard-shortcuts-settings" onClick={e => {
                                        e.preventDefault()
                                        e.stopPropagation()
                                        setKeybindingsModalOpen(true)
                                    }}
                                    >
                                        Keyboard Shortcuts
                                    </MediaSettingsMenuItem>
                                </MediaSettingsMenu>

                                <MediaCaptionsMenu anchor="auto" hidden onChange={onCaptionsChange}>
                                    <div slot="header">Subtitles/CC</div>
                                </MediaCaptionsMenu>

                                <MediaAudioTrackMenu
                                    anchor="auto" hidden onChange={onAudioChange}
                                >
                                    <div slot="header">Audio</div>
                                </MediaAudioTrackMenu>

                                <MediaTimeRange
                                    ref={timeRangeRef}
                                    mediaChaptersCues={chapterCues}
                                >
                                    <MediaPreviewThumbnail
                                        slot="preview"
                                        mediaPreviewImage={previewThumbnail}
                                        mediaPreviewCoords={previewThumbnail ? [0, 0, 200, 100] : undefined}
                                    />
                                    <MediaPreviewChapterDisplay slot="preview" />
                                    <MediaPreviewTimeDisplay slot="preview" />
                                </MediaTimeRange>

                                <MediaControlBar>
                                    <MediaPlayButton
                                        className="native-player-button flex justify-center items-center"
                                        data-mini-player={state.miniPlayer}
                                        dangerouslySetInnerHTML={{
                                            __html: NativePlayerIcons.Play,
                                        }}
                                    />

                                    <MediaMuteButton
                                        className="native-player-button" data-mini-player={state.miniPlayer}
                                        dangerouslySetInnerHTML={{
                                            __html: NativePlayerIcons.Mute,
                                        }}
                                    />

                                    <MediaVolumeRange />

                                    <MediaTimeDisplay showDuration />

                                    <span className="control-spacer" />

                                    <MediaAudioTrackMenuButton
                                        className="native-player-button" data-mini-player={state.miniPlayer}
                                        dangerouslySetInnerHTML={{
                                            __html: NativePlayerIcons.AudioTrack,
                                        }}
                                    />

                                    <MediaCaptionsMenuButton
                                        className="native-player-button" data-mini-player={state.miniPlayer}
                                        dangerouslySetInnerHTML={{
                                            __html: NativePlayerIcons.Captions,
                                        }}
                                    />

                                    <MediaSettingsMenuButton
                                        className="native-player-button" data-mini-player={state.miniPlayer}
                                        dangerouslySetInnerHTML={{
                                            __html: NativePlayerIcons.Settings,
                                        }}
                                    />

                                    <MediaPipButton
                                        className="native-player-button" data-mini-player={state.miniPlayer}
                                        dangerouslySetInnerHTML={{
                                            __html: NativePlayerIcons.Pip,
                                        }}
                                    />

                                    <MediaFullscreenButton
                                        className="native-player-button" data-mini-player={state.miniPlayer}
                                        dangerouslySetInnerHTML={{
                                            __html: NativePlayerIcons.Fullscreen,
                                        }}
                                    />

                                </MediaControlBar>

                            </MediaController>
                        </MediaProvider>
                    ) : (
                        <div
                            className="w-full h-full absolute flex justify-center items-center flex-col space-y-4 bg-black rounded-md"
                        >
                            <LoadingSpinner
                                title={state.loadingState || "Loading..."}
                                spinner={<PiSpinnerDuotone className="size-20 text-white animate-spin" />}
                            />
                        </div>
                    )}
                </div>
            </NativePlayerDrawer>
        </>
    )
}
