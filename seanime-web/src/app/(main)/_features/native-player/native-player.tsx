import { getServerBaseUrl } from "@/api/client/server-url"
import { MKVParser_SubtitleEvent, MKVParser_TrackInfo, NativePlayer_PlaybackInfo, NativePlayer_ServerEvent } from "@/api/generated/types"
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
import React, { FormEvent, useCallback, useEffect, useRef } from "react"
import { BiExpand } from "react-icons/bi"
import { FiMinimize2 } from "react-icons/fi"
import { PiSpinnerDuotone } from "react-icons/pi"
import { useWebsocketMessageListener, useWebsocketSender } from "../../_hooks/handle-websockets"
import { StreamAudioManager, StreamSubtitleManager } from "./handle-native-player"
import { NativePlayerDrawer } from "./native-player-drawer"
import { nativePlayer_settingsAtom, nativePlayer_stateAtom } from "./native-player.atoms"
import { detectSubtitleType, isSubtitleFile } from "./native-player.utils"

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

    const streamLoadedRef = useRef<string | null>(null)
    const subtitleManagerRef = useRef<StreamSubtitleManager | null>(null)
    const audioManagerRef = useRef<StreamAudioManager | null>(null)

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
        log.info("Can play")

        // Check for audio and subtitle tracks if needed
        if (videoRef.current) {
            // Using textTracks which is standard
            log.info("Text tracks", videoRef.current.textTracks)
            log.info("Audio tracks", videoRef.current.audioTracks)
        }
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
    }

    const handleMuteChange = (e: React.SyntheticEvent<HTMLVideoElement>) => {
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
            actualFiles.forEach(async f => {
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
            })
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

    return (
        <>
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
                                return
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
                closeButton={!state.miniPlayer ? <IconButton
                    icon={<FiMinimize2 className="text-2xl" />}
                    intent="gray-basic"
                    className="rounded-full"
                /> : undefined}
                data-native-player-drawer
            >
                {state.miniPlayer && (
                    <IconButton
                        type="button"
                        intent="gray-basic"
                        size="sm"
                        className={cn(
                            "rounded-full text-2xl flex-none absolute z-[99] right-4 top-4 pointer-events-auto",
                            state.miniPlayer && "text-xl",
                        )}
                        icon={<BiExpand />}
                        onClick={() => {
                            setState(draft => {
                                draft.miniPlayer = false
                            })
                        }}
                    />
                )}

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
                        <MediaController
                            className={cn(
                                "w-full h-full",
                                discreteControls && "discrete-controls",
                            )}
                            tabIndex={-1}
                        >
                            <video
                                ref={videoRef}
                                slot="media"
                                src={state.playbackInfo?.streamUrl?.replace("{{SERVER_URL}}", getServerBaseUrl())}
                                crossOrigin="anonymous"
                                playsInline
                                autoPlay={autoPlay}
                                muted={muted}
                                // onTimeUpdate={handleTimeUpdate}
                                onDurationChange={handleDurationChange}
                                // onCanPlay={handleCanPlay}
                                onEnded={handleEnded}
                                onError={handleError}
                                onVolumeChange={handleVolumeChange}
                                onSeeked={handleSeeked}
                                onLoadedMetadata={handleLoadedMetadata}
                                onPause={handlePause}
                                onPlay={handlePlay}
                                style={{
                                    width: "100%",
                                    height: "100%",
                                    border: "none",
                                    // Enhanced color settings for MPV-like rendering
                                    filter: settings.videoEnhancement.enabled
                                        ? `contrast(${settings.videoEnhancement.contrast}) saturate(${settings.videoEnhancement.saturation}) brightness(${settings.videoEnhancement.brightness})`
                                        : "none",
                                    imageRendering: "auto",
                                }}
                                tabIndex={-1}
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
                            </video>

                            <MediaErrorDialog slot="dialog" />

                            <div className="yt-gradient-bottom"></div>

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
                            </MediaSettingsMenu>

                            <MediaCaptionsMenu anchor="auto" hidden onChange={onCaptionsChange}>
                                <div slot="header">Subtitles/CC</div>
                            </MediaCaptionsMenu>

                            <MediaAudioTrackMenu
                                anchor="auto" hidden onChange={onAudioChange}
                            >
                                <div slot="header">Audio</div>
                            </MediaAudioTrackMenu>

                            <MediaTimeRange>
                                <MediaPreviewThumbnail slot="preview" />
                                <MediaPreviewChapterDisplay slot="preview" />
                                <MediaPreviewTimeDisplay slot="preview" />
                            </MediaTimeRange>

                            <MediaControlBar>
                                <MediaPlayButton
                                    className="native-player-button flex justify-center items-center"
                                    data-mini-player={state.miniPlayer}
                                    dangerouslySetInnerHTML={{
                                        __html: `
<svg xmlns="http://www.w3.org/2000/svg" slot="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="0">
<path
id="play-p1"
        fill="currentColor"
        d="M8 18.392V5.608L18.226 12zM6 3.804v16.392a1 1 0 0 0 1.53.848l13.113-8.196a1 1 0 0 0 0-1.696L7.53 2.956A1 1 0 0 0 6 3.804"
      ></path>
            <path id="pause-p1" fill="currentColor" d="M6 3h2v18H6zm10 0h2v18h-2z"></path>
</svg>
`,
                                    }}
                                >

                                </MediaPlayButton>

                                <MediaMuteButton
                                    className="native-player-button" data-mini-player={state.miniPlayer} dangerouslySetInnerHTML={{
                                    __html: `
<span slot="high">
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4.702a.705.705 0 0 0-1.203-.498L6.413 7.587A1.4 1.4 0 0 1 5.416 8H3a1 1 0 0 0-1 1v6a1 1 0 0 0 1 1h2.416a1.4 1.4 0 0 1 .997.413l3.383 3.384A.705.705 0 0 0 11 19.298z"/><path d="M16 9a5 5 0 0 1 0 6"/><path d="M19.364 18.364a9 9 0 0 0 0-12.728"/></svg>
</span>
<span slot="medium">
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4.702a.705.705 0 0 0-1.203-.498L6.413 7.587A1.4 1.4 0 0 1 5.416 8H3a1 1 0 0 0-1 1v6a1 1 0 0 0 1 1h2.416a1.4 1.4 0 0 1 .997.413l3.383 3.384A.705.705 0 0 0 11 19.298z"/><path d="M16 9a5 5 0 0 1 0 6"/></svg>
</span>
<span slot="low">
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4.702a.705.705 0 0 0-1.203-.498L6.413 7.587A1.4 1.4 0 0 1 5.416 8H3a1 1 0 0 0-1 1v6a1 1 0 0 0 1 1h2.416a1.4 1.4 0 0 1 .997.413l3.383 3.384A.705.705 0 0 0 11 19.298z"/><path d="M16 9a5 5 0 0 1 0 6"/></svg>
</span>
<span slot="muted">
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4.702a.705.705 0 0 0-1.203-.498L6.413 7.587A1.4 1.4 0 0 1 5.416 8H3a1 1 0 0 0-1 1v6a1 1 0 0 0 1 1h2.416a1.4 1.4 0 0 1 .997.413l3.383 3.384A.705.705 0 0 0 11 19.298z"/></svg>
</span>
<span slot="off">
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4.702a.705.705 0 0 0-1.203-.498L6.413 7.587A1.4 1.4 0 0 1 5.416 8H3a1 1 0 0 0-1 1v6a1 1 0 0 0 1 1h2.416a1.4 1.4 0 0 1 .997.413l3.383 3.384A.705.705 0 0 0 11 19.298z"/><line x1="22" x2="16" y1="9" y2="15"/><line x1="16" x2="22" y1="9" y2="15"/></svg>
</span>
                                    `,
                                }}
                                >

                                </MediaMuteButton>
                                <MediaVolumeRange />

                                <MediaTimeDisplay showDuration />

                                <span className="control-spacer" />

                                <MediaAudioTrackMenuButton
                                    className="native-player-button" data-mini-player={state.miniPlayer} dangerouslySetInnerHTML={{
                                    __html: `
<svg slot="icon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M2 10v3"/><path d="M6 6v11"/><path d="M10 3v18"/><path d="M14 8v7"/><path d="M18 5v13"/><path d="M22 10v3"/></svg>
`,
                                }}
                                >

                                </MediaAudioTrackMenuButton>

                                <MediaCaptionsMenuButton
                                    className="native-player-button" data-mini-player={state.miniPlayer} dangerouslySetInnerHTML={{
                                    __html: `
<svg slot="icon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><rect width="18" height="14" x="3" y="5" rx="2" ry="2" fill="none"/><path d="M7 15h4M15 15h2M7 11h2M13 11h4"/></svg>
                                    `,
                                }}
                                >

                                </MediaCaptionsMenuButton>

                                <MediaSettingsMenuButton
                                    className="native-player-button" data-mini-player={state.miniPlayer} dangerouslySetInnerHTML={{
                                    __html: `    <svg
                                    slot="icon"
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
        >
      <path
      id="settings-icon"
        fill="currentColor"
        d="M2 12c0-.865.11-1.704.316-2.504A3 3 0 0 0 4.99 4.867a10 10 0 0 1 4.335-2.506a3 3 0 0 0 5.348 0a10 10 0 0 1 4.335 2.506a3 3 0 0 0 2.675 4.63c.206.8.316 1.638.316 2.503c0 .864-.11 1.703-.316 2.503a3 3 0 0 0-2.675 4.63a10 10 0 0 1-4.335 2.505a3 3 0 0 0-5.348 0a10 10 0 0 1-4.335-2.505a3 3 0 0 0-2.675-4.63C2.11 13.703 2 12.864 2 12m4.804 3c.63 1.091.81 2.346.564 3.524q.613.436 1.297.75A5 5 0 0 1 12 18c1.26 0 2.438.471 3.335 1.274q.684-.314 1.297-.75A5 5 0 0 1 17.196 15a5 5 0 0 1 2.77-2.25a8 8 0 0 0 0-1.5A5 5 0 0 1 17.196 9a5 5 0 0 1-.564-3.524a8 8 0 0 0-1.297-.75A5 5 0 0 1 12 6a5 5 0 0 1-3.335-1.274a8 8 0 0 0-1.297.75A5 5 0 0 1 6.804 9a5 5 0 0 1-2.77 2.25a8 8 0 0 0 0 1.5A5 5 0 0 1 6.805 15M12 15a3 3 0 1 1 0-6a3 3 0 0 1 0 6m0-2a1 1 0 1 0 0-2a1 1 0 0 0 0 2"
      ></path>
    </svg>`,
                                }}
                                >

                                </MediaSettingsMenuButton>

                                <MediaPipButton
                                    className="native-player-button" data-mini-player={state.miniPlayer} dangerouslySetInnerHTML={{
                                    __html: `   <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      slot="icon"
    >
      <path
      id="pip-icon"
        fill="currentColor"
        d="M21 3a1 1 0 0 1 1 1v7h-2V5H4v14h6v2H3a1 1 0 0 1-1-1V4a1 1 0 0 1 1-1zm0 10a1 1 0 0 1 1 1v6a1 1 0 0 1-1 1h-8a1 1 0 0 1-1-1v-6a1 1 0 0 1 1-1zm-1 2h-6v4h6z"
      ></path>
            <path
        id="pip-icon-2"
        fill="currentColor"
        d="M21 3a1 1 0 0 1 1 1v7h-2V5H4v14h6v2H3a1 1 0 0 1-1-1V4a1 1 0 0 1 1-1zm0 10a1 1 0 0 1 1 1v6a1 1 0 0 1-1 1h-8a1 1 0 0 1-1-1v-6a1 1 0 0 1 1-1zm-1 2h-6v4h6zm-8.5-8L9.457 9.043l2.25 2.25l-1.414 1.414l-2.25-2.25L6 12.5V7z"
      ></path>
    </svg>`,
                                }}
                                >
                                </MediaPipButton>

                                <MediaFullscreenButton
                                    className="native-player-button" data-mini-player={state.miniPlayer} dangerouslySetInnerHTML={{
                                    __html: `
                <span slot="enter">
                <svg slot="icon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-maximize-icon lucide-maximize">
                <g id="fs-enter-paths">
                <path d="M8 3H5a2 2 0 0 0-2 2v3"/>
                <path d="M21 8V5a2 2 0 0 0-2-2h-3"/>
                <path d="M3 16v3a2 2 0 0 0 2 2h3"/>
                <path d="M16 21h3a2 2 0 0 0 2-2v-3"/>
                </g>
                </svg>
                </span>
                <span slot="exit">
                <svg slot="icon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-minimize-icon lucide-minimize">
                <g id="fs-exit-paths">
                <path d="M8 3v3a2 2 0 0 1-2 2H3"/><path d="M21 8h-3a2 2 0 0 1-2-2V3"/><path d="M3 16h3a2 2 0 0 1 2 2v3"/><path d="M16 21v-3a2 2 0 0 1 2-2h3"/>

                </g>
                </svg>
                </span>`
                                }}
                                >
                                </MediaFullscreenButton>
                            </MediaControlBar>

                        </MediaController>
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
