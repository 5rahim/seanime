import { MKVParser_TrackInfo, VideoCore_ClientEventType, VideoCore_PlaybackState, VideoCore_ServerEvent } from "@/api/generated/types"
import {
    vc_activePlayerId,
    vc_anime4kManager,
    vc_audioManager,
    vc_mediaCaptionsManager,
    vc_subtitleManager,
} from "@/app/(main)/_features/video-core/video-core"
import { Anime4KManagerOptionChangedEvent } from "@/app/(main)/_features/video-core/video-core-anime-4k-manager"
import { AudioManagerHlsTrackChangedEvent, AudioManagerTrackChangedEvent } from "@/app/(main)/_features/video-core/video-core-audio"
import { FullscreenManagerChangedEvent, vc_fullscreenManager } from "@/app/(main)/_features/video-core/video-core-fullscreen"
import {
    MediaCaptionsTrackDeselectedEvent,
    MediaCaptionsTrackInfo,
    MediaCaptionsTrackSelectedEvent,
} from "@/app/(main)/_features/video-core/video-core-media-captions"
import { vc_showOverlayFeedback } from "@/app/(main)/_features/video-core/video-core-overlay-display"
import { PipManagerToggledEvent, vc_pipManager } from "@/app/(main)/_features/video-core/video-core-pip"
import { useVideoCorePlaylist, VideoCorePlaylistState } from "@/app/(main)/_features/video-core/video-core-playlist"
import { SubtitleManagerTrackDeselectedEvent, SubtitleManagerTrackSelectedEvent } from "@/app/(main)/_features/video-core/video-core-subtitles"
import { vc_autoNextAtom, VideoCore_VideoSubtitleTrack, VideoCoreLifecycleState } from "@/app/(main)/_features/video-core/video-core.atoms"
import { detectSubtitleType, isSubtitleFile } from "@/app/(main)/_features/video-core/video-core.utils"
import { useWebsocketMessageListener, useWebsocketSender } from "@/app/(main)/_hooks/handle-websockets"
import { clientIdAtom } from "@/app/websocket-provider"
import { logger } from "@/lib/helpers/debug"
import { WSEvents } from "@/lib/server/ws-events"
import { useAtomValue, useSetAtom } from "jotai"
import React, { useCallback, useRef } from "react"
import { toast } from "sonner"

export type ClientSubtitleFileUploadedEventPayload = {
    filename: string
    content: string
}

export type VideoCoreErrorEventPayload = {
    error: string
}

export type VideoCoreEndedEventPayload = {
    autoNext: boolean
}

export type VideoCoreStatusEventPayload = {
    id: string
    clientId: string
    currentTime: number
    duration: number
    paused: boolean
}


export type VideoCoreFullscreenEventPayload = {
    fullscreen: boolean
}

export type VideoCorePipEventPayload = {
    pip: boolean
}

export type VideoCoreSubtitleTrackEventPayload = {
    trackNumber: number
    kind?: "file" | "event"
}

export type VideoCoreMediaCaptionTrackEventPayload = {
    trackIndex: number
}

export type VideoCoreAudioTrackEventPayload = {
    trackNumber: number
    isHLS: boolean
}

export type VideoCoreAnime4KEventPayload = {
    option: string
}

export type VideoCoreLoadedPayload = {
    state: VideoCore_PlaybackState
}

const log = logger("VideoCoreEvents")

export function useVideoCoreSetupEvents(id: string,
    state: VideoCoreLifecycleState,
    videoRef: React.MutableRefObject<HTMLVideoElement | null>,
    onTerminateStream: () => void,
) {
    const { sendEvent } = useVideoCoreEvents()

    const clientId = useAtomValue(clientIdAtom)
    const activePlayer = useAtomValue(vc_activePlayerId)
    const fullscreenManager = useAtomValue(vc_fullscreenManager)
    const subtitleManager = useAtomValue(vc_subtitleManager)
    const anime4kManager = useAtomValue(vc_anime4kManager)
    const mediaCaptionsManager = useAtomValue(vc_mediaCaptionsManager)
    const pipManager = useAtomValue(vc_pipManager)
    const audioManager = useAtomValue(vc_audioManager)
    const autoNext = useAtomValue(vc_autoNextAtom)
    const showOverlayFeedback = useSetAtom(vc_showOverlayFeedback)
    const { playEpisode: playPlaylistEpisode, playlistState } = useVideoCorePlaylist()

    // React.useEffect(() => {
    //     log.trace(activePlayer, id)
    // })

    // Check if the current player is the active one
    // For native player, activePlayer should be null
    const isActivePlayer = React.useMemo(() => {
        if (id === "native-player") {
            return activePlayer === null || activePlayer === "native-player"
        }
        return activePlayer === id
    }, [activePlayer, id])

    // fullscreen events
    React.useLayoutEffect(() => {
        if (!isActivePlayer || !fullscreenManager) return

        function handleFullscreenChange(ev: FullscreenManagerChangedEvent) {
            log.trace(`Fullscreen changed: ${ev.detail.isFullscreen}`)
            sendEvent<VideoCoreFullscreenEventPayload>("video-fullscreen", { fullscreen: ev.detail.isFullscreen })
        }

        fullscreenManager.addEventListener("fullscreenchanged", handleFullscreenChange)
        return () => {
            fullscreenManager.removeEventListener("fullscreenchanged", handleFullscreenChange)
        }
    }, [isActivePlayer, state, fullscreenManager])

    // subtitle events
    React.useLayoutEffect(() => {
        if (!isActivePlayer || !subtitleManager) return

        function handleTrackSelected(ev: SubtitleManagerTrackSelectedEvent) {
            log.trace(`Subtitle track changed: ${ev.detail.trackNumber}`)
            sendEvent<VideoCoreSubtitleTrackEventPayload>("video-subtitle-track", { trackNumber: ev.detail.trackNumber, kind: ev.detail.kind })
        }

        function handleTrackDeselected(ev: SubtitleManagerTrackDeselectedEvent) {
            log.trace(`Subtitle track changed: -1`)
            sendEvent<VideoCoreSubtitleTrackEventPayload>("video-subtitle-track", { trackNumber: -1 })
        }

        subtitleManager.addEventListener("trackselected", handleTrackSelected)
        subtitleManager.addEventListener("trackdeselected", handleTrackDeselected)
        return () => {
            subtitleManager.removeEventListener("trackselected", handleTrackSelected)
            subtitleManager.addEventListener("trackdeselected", handleTrackDeselected)
        }
    }, [isActivePlayer, state, subtitleManager])

    // media captions events
    React.useLayoutEffect(() => {
        if (!isActivePlayer || !mediaCaptionsManager) return

        function handleTrackSelected(ev: MediaCaptionsTrackSelectedEvent) {
            log.trace(`Media caption track changed: ${ev.detail.trackIndex}`)
            sendEvent<VideoCoreMediaCaptionTrackEventPayload>("video-media-caption-track", { trackIndex: ev.detail.trackIndex })
        }

        function handleTrackDeselected(ev: MediaCaptionsTrackDeselectedEvent) {
            log.trace(`Media caption track changed: -1`)
            sendEvent<VideoCoreMediaCaptionTrackEventPayload>("video-media-caption-track", { trackIndex: -1 })
        }

        mediaCaptionsManager.addEventListener("trackselected", handleTrackSelected)
        mediaCaptionsManager.addEventListener("trackdeselected", handleTrackDeselected)
        return () => {
            mediaCaptionsManager.removeEventListener("trackselected", handleTrackSelected)
            mediaCaptionsManager.addEventListener("trackdeselected", handleTrackDeselected)
        }
    }, [isActivePlayer, state, mediaCaptionsManager])

    // audio events
    React.useLayoutEffect(() => {
        if (!isActivePlayer || !audioManager) return

        function handleAudioTrackChanged(ev: AudioManagerTrackChangedEvent) {
            log.trace(`Subtitle track changed: ${ev.detail.trackNumber}`)
            sendEvent<VideoCoreAudioTrackEventPayload>("video-audio-track", { trackNumber: ev.detail.trackNumber, isHLS: false })
        }

        function handleHlsAudioTrackChanged(ev: AudioManagerHlsTrackChangedEvent) {
            log.trace(`Subtitle track changed: ${ev.detail.trackId}`)
            sendEvent<VideoCoreAudioTrackEventPayload>("video-audio-track", { trackNumber: ev.detail.trackId, isHLS: true })
        }

        audioManager.addEventListener("trackchanged", handleAudioTrackChanged)
        audioManager.addEventListener("hlstrackchanged", handleHlsAudioTrackChanged)
        return () => {
            audioManager.removeEventListener("trackchanged", handleAudioTrackChanged)
            audioManager.addEventListener("hlstrackchanged", handleHlsAudioTrackChanged)
        }
    }, [isActivePlayer, state, audioManager])

    // pip events
    React.useLayoutEffect(() => {
        if (!isActivePlayer || !pipManager) return

        function handleToggledPip(ev: PipManagerToggledEvent) {
            log.trace(`PIP Changed: ${ev.detail.enabled}`)
            sendEvent<VideoCorePipEventPayload>("video-pip", { pip: ev.detail.enabled })
        }

        pipManager.addEventListener("toggledpip", handleToggledPip)
        return () => {
            pipManager.removeEventListener("toggledpip", handleToggledPip)
        }
    }, [isActivePlayer, state, pipManager])


    // anime4k events
    React.useLayoutEffect(() => {
        if (!isActivePlayer || !anime4kManager) return

        function handleOptionChanged(ev: Anime4KManagerOptionChangedEvent) {
            log.trace(`Anime4K Changed: ${ev.detail.newOption}`)
            sendEvent<VideoCoreAnime4KEventPayload>("video-anime-4k", { option: ev.detail.newOption })
        }

        function handleDestroyed(ev: any) {
            log.trace(`Anime4K Changed: Off`)
            sendEvent<VideoCoreAnime4KEventPayload>("video-anime-4k", { option: "off" })
        }

        anime4kManager.addEventListener("optionchanged", handleOptionChanged)
        anime4kManager.addEventListener("destroyed", handleDestroyed)
        anime4kManager.addEventListener("error", handleDestroyed)
        return () => {
            anime4kManager.removeEventListener("optionchanged", handleOptionChanged)
            anime4kManager.removeEventListener("destroyed", handleDestroyed)
            anime4kManager.removeEventListener("error", handleDestroyed)
        }
    }, [isActivePlayer, state, anime4kManager])

    const lastSeekedSent = useRef(Date.now() - 1000)

    // video events
    React.useEffect(() => {
        if (!isActivePlayer || !videoRef.current) return

        function handlePlay() {
            if (!videoRef.current) return
            log.trace("Video resumed")
            sendEvent("video-resumed", {
                currentTime: videoRef.current.currentTime,
                duration: videoRef.current.duration,
            })
        }

        function handlePaused() {
            if (!videoRef.current) return
            log.trace("Video paused")
            sendEvent("video-paused", {
                currentTime: videoRef.current.currentTime,
                duration: videoRef.current.duration,
                paused: videoRef.current.paused,
            })
        }

        function handleLoadedMetadata() {
            if (!videoRef.current) return
            log.trace("Video loaded metadata")
            sendEvent("video-loaded-metadata")
            sendEvent("video-loaded-metadata", {
                id: state.playbackInfo?.id || "",
                clientId: clientId,
                currentTime: videoRef.current.currentTime,
                duration: videoRef.current.duration,
                paused: videoRef.current.paused,
            })
        }

        function handleEnded() {
            log.trace("Video ended")
            sendEvent<VideoCoreEndedEventPayload>("video-ended", { autoNext: autoNext })
        }

        function handleSeeked() {
            if (Date.now() - lastSeekedSent.current < 1000 || !videoRef.current) return
            lastSeekedSent.current = Date.now()
            log.trace("Video seeked")
            sendEvent("video-seeked", {
                currentTime: videoRef.current.currentTime,
                duration: videoRef.current.duration,
                paused: videoRef.current.paused,
            })
        }

        videoRef.current?.addEventListener("play", handlePlay)
        videoRef.current?.addEventListener("pause", handlePaused)
        videoRef.current?.addEventListener("loadedmetadata", handleLoadedMetadata)
        videoRef.current?.addEventListener("ended", handleEnded)
        videoRef.current?.addEventListener("seeked", handleSeeked)

        return () => {
            videoRef.current?.removeEventListener("play", handlePlay)
            videoRef.current?.removeEventListener("pause", handlePaused)
            videoRef.current?.removeEventListener("loadedmetadata", handleLoadedMetadata)
            videoRef.current?.removeEventListener("ended", handleEnded)
            videoRef.current?.removeEventListener("seeked", handleSeeked)
        }
    }, [isActivePlayer, state])

    function dispatchTerminatedEvent() {
        log.trace("Video terminated")
        sendEvent("video-terminated")
    }

    function dispatchTranslateTextEvent(text: string) {
        if (!videoRef.current) return
        sendEvent("translate-text", {
            text: text,
        })
    }

    function dispatchTranslateSubtitleTrackEvent(track: VideoCore_VideoSubtitleTrack) {
        if (!videoRef.current) return
        sendEvent("translate-subtitle-file-track", track)
    }

    function dispatchCanPlayEvent() {
        if (!videoRef.current) return
        log.trace("Video can play")
        sendEvent("video-can-play", {
            currentTime: videoRef.current.currentTime,
            duration: videoRef.current.duration,
            paused: videoRef.current.paused,
        })
    }

    function dispatchVideoCompletedEvent() {
        if (!videoRef.current) return
        log.trace("Video completed")
        sendEvent("video-completed", {
            currentTime: videoRef.current.currentTime,
            duration: videoRef.current.duration,
            paused: videoRef.current.paused,
        })
    }

    function dispatchVideoErrorEvent(value: string) {
        log.trace("Video error")
        const v = videoRef.current
        if (!v) return


        const error = v.error || value
        let msg = value || "Unknown error"
        let detailedInfo = ""

        if (error instanceof MediaError) {
            switch (error.code) {
                case MediaError.MEDIA_ERR_ABORTED:
                    msg = "Media playback aborted"
                    break
                case MediaError.MEDIA_ERR_NETWORK:
                    msg = "Network error occurred: Check the console and network tab for more details"
                    break
                case MediaError.MEDIA_ERR_DECODE:
                    msg = "Media decode error: codec not supported or corrupted file"
                    detailedInfo = "This is likely a codec compatibility issue."
                    break
                case MediaError.MEDIA_ERR_SRC_NOT_SUPPORTED:
                    msg = "Media format not supported"
                    detailedInfo = "The video codec/container format is not supported."
                    break
                default:
                    msg = error.message || "Unknown media error"
            }
            log.error("Media error", {
                code: error?.code,
                message: error?.message,
                src: v.src,
                networkState: v.networkState,
                readyState: v.readyState,
            })
        }


        const fullErrorMessage = detailedInfo ? `${msg}\n\n${detailedInfo}` : msg

        log.error("Media error", fullErrorMessage)

        sendEvent<VideoCoreErrorEventPayload>("video-error", { error: fullErrorMessage })
    }


    function dispatchVideoLoadedEvent() {
        if (!state.playbackInfo || !clientId) return
        log.trace("Video loaded")
        sendEvent<VideoCoreLoadedPayload>("video-loaded", {
            state: {
                clientId: clientId,
                playerType: id === "native-player" ? "native" : "web",
                playbackInfo: state.playbackInfo,
            },
        })
    }

    React.useEffect(() => {
        if (!isActivePlayer || !state.playbackInfo || !videoRef.current || !clientId) return

        const interval = setInterval(() => {
            if (!videoRef.current) return
            // log.trace("Sending video status")
            sendEvent<VideoCoreStatusEventPayload>("video-status", {
                id: state.playbackInfo?.id || "",
                clientId: clientId,
                currentTime: videoRef.current.currentTime,
                duration: videoRef.current.duration,
                paused: videoRef.current.paused,
            })
        }, 1000)

        return () => {
            clearInterval(interval)
        }
    }, [state.playbackInfo, isActivePlayer, clientId])

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
                    sendEvent<ClientSubtitleFileUploadedEventPayload>("subtitle-file-uploaded", {
                        filename: f.name,
                        content: content,
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
                    log.info("Detected subtitle type", type)
                    if (type === "unknown") {
                        toast.error("Unknown subtitle type")
                        log.info("Unknown subtitle type, skipping")
                        return
                    }
                    const filename = `PLACEHOLDER.${type}`
                    sendEvent<ClientSubtitleFileUploadedEventPayload>("subtitle-file-uploaded", {
                        filename: filename,
                        content: str,
                    })
                })
            }
        }
    }, [])

    function suppressEvent(e: Event) {
        e.preventDefault()
    }

    React.useEffect(() => {
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

    useWebsocketMessageListener({
        type: WSEvents.VIDEOCORE,
        deps: [activePlayer, id],
        onMessage: ({ type, payload }: { type: VideoCore_ServerEvent, payload: unknown }) => {
            if (!isActivePlayer || !videoRef.current) return

            switch (type) {
                case "get-status":
                    if (!clientId || !state.playbackInfo) return
                    sendEvent<VideoCoreStatusEventPayload>("video-status", {
                        id: state.playbackInfo?.id || "",
                        clientId: clientId,
                        currentTime: videoRef.current.currentTime,
                        duration: videoRef.current.duration,
                        paused: videoRef.current.paused,
                    })
                    break
                case "pause":
                    log.info("Pause event received", payload)
                    videoRef.current.pause()
                    break
                case "resume":
                    log.info("Resume event received", payload)
                    videoRef.current.play().catch()
                    break
                case "seek":
                    log.info("Seek event received", payload)
                    const seekAmount = payload as number
                    const newTime = videoRef.current.currentTime + seekAmount
                    if (videoRef.current.duration) {
                        videoRef.current.currentTime = Math.max(0, Math.min(newTime, videoRef.current.duration))
                    }
                    break
                case "seek-to":
                    log.info("Seek to event received", payload)
                    const seekToTime = payload as number
                    if (videoRef.current.duration) {
                        videoRef.current.currentTime = Math.max(0, Math.min(seekToTime, videoRef.current.duration))
                    }
                    break
                case "set-fullscreen":
                    log.info("Set fullscreen event received", payload)
                    if (fullscreenManager) {
                        if (payload as boolean) {
                            fullscreenManager.enterFullscreen()
                        } else {
                            fullscreenManager.exitFullscreen()
                        }
                    }
                    break
                case "set-pip":
                    log.info("Set PIP event received", payload)
                    if (pipManager) {
                        const enablePip = payload as boolean
                        const isCurrentlyInPip = document.pictureInPictureElement !== null
                        if (enablePip && !isCurrentlyInPip) {
                            pipManager.togglePip(true)
                        } else if (!enablePip && isCurrentlyInPip) {
                            pipManager.togglePip(false)
                        }
                    }
                    break
                case "set-subtitle-track":
                    log.info("Set subtitle track event received", payload)
                    if (subtitleManager) {
                        const trackNumber = payload as number
                        if (trackNumber === -1) {
                            subtitleManager.setNoTrack()
                        } else {
                            subtitleManager.selectTrack(trackNumber)
                        }
                    }
                    break
                case "add-subtitle-track":
                    log.info("Add subtitle track event received", payload)
                    const subtitleTrack = payload as MKVParser_TrackInfo
                    if (subtitleManager) {
                        subtitleManager.addEventTrack(subtitleTrack)
                    }
                    break
                case "add-external-subtitle-track":
                    log.info("Add subtitle track event received", payload)
                    const fileTrack = payload as VideoCore_VideoSubtitleTrack
                    if (subtitleManager) {
                        subtitleManager.addFileTrack(fileTrack)
                    } else if (mediaCaptionsManager) {
                        mediaCaptionsManager.addCaptionTrack(fileTrack)
                    }
                    showOverlayFeedback({ message: `Subtitle track added: ${fileTrack.label}`, type: "message", duration: 1500 })
                    break
                case "set-media-caption-track":
                    log.info("Set media caption track event received", payload)
                    if (mediaCaptionsManager) {
                        const trackIndex = payload as number
                        if (trackIndex === -1) {
                            mediaCaptionsManager.selectTrack(-1)
                        } else {
                            mediaCaptionsManager.selectTrack(trackIndex)
                        }
                    }
                    break
                case "add-media-caption-track":
                    log.info("Add media caption track event received", payload)
                    const track = payload as MediaCaptionsTrackInfo
                    if (mediaCaptionsManager) {
                        mediaCaptionsManager.addCaptionTrack({ index: -1, ...track })
                    }
                    break
                case "set-audio-track":
                    log.info("Set audio track event received", payload)
                    if (audioManager) {
                        const trackNumber = payload as number
                        audioManager.selectTrack(trackNumber)
                    }
                    break
                case "terminate":
                    log.info("Terminate event received")
                    onTerminateStream()
                    dispatchTerminatedEvent()
                    break
                case "get-anime-4k":
                    log.info("Get anime 4k event received")
                    const anime4kOption = anime4kManager?.getCurrentOption() || "off"
                    sendEvent<VideoCoreAnime4KEventPayload>("video-anime-4k", { option: anime4kOption })
                    break
                case "get-audio-track":
                    log.info("Get audio track event received")
                    const trackNumber = audioManager?.getSelectedTrackNumberOrNull() ?? -1
                    sendEvent<VideoCoreAudioTrackEventPayload>("video-audio-track", { trackNumber: trackNumber, isHLS: audioManager?.isHLS ?? false })
                    break
                case "get-subtitle-track":
                    log.info("Get subtitle track event received")
                    const subtitleTrackNumber = subtitleManager?.getSelectedTrackNumberOrNull() ?? -1
                    sendEvent<VideoCoreSubtitleTrackEventPayload>("video-subtitle-track",
                        { trackNumber: subtitleTrackNumber, kind: !!subtitleManager?.getFileTrack(subtitleTrackNumber) ? "file" : "event" })
                    break
                case "get-media-caption-track":
                    log.info("Get media caption track event received")
                    const mediaCaptionTrackIndex = mediaCaptionsManager?.getSelectedTrackIndexOrNull() ?? -1
                    sendEvent<VideoCoreMediaCaptionTrackEventPayload>("video-media-caption-track", { trackIndex: mediaCaptionTrackIndex })
                    break
                case "get-fullscreen":
                    log.info("Get fullscreen event received")
                    sendEvent<VideoCoreFullscreenEventPayload>("video-fullscreen", { fullscreen: fullscreenManager?.isFullscreen ?? false })
                    break
                case "get-pip":
                    log.info("Get PIP event received")
                    sendEvent<VideoCorePipEventPayload>("video-pip", { pip: pipManager?.isPip ?? false })
                    break
                case "get-playback-state":
                    log.info("Get playback state event received")
                    if (!state.playbackInfo || !clientId) {
                        sendEvent<VideoCoreLoadedPayload>("video-playback-state", {
                            state: {} as VideoCore_PlaybackState,
                        })
                        break
                    }
                    sendEvent<VideoCoreLoadedPayload>("video-playback-state", {
                        state: {
                            clientId: clientId!,
                            playerType: id === "native-player" ? "native" : "web",
                            playbackInfo: state.playbackInfo!,
                        },
                    })
                    break
                case "show-message":
                    log.info("Show message event received", payload)
                    const data = payload as { message: string, duration?: number }
                    showOverlayFeedback({ message: data.message, type: "message", duration: data.duration ?? 2000 })
                    break
                case "get-playlist":
                    log.info("Get playlist event received")
                    if (!playlistState) return
                    sendEvent<{ playlist: VideoCorePlaylistState }>("video-playlist", {
                        playlist: playlistState,
                    })
                    break
                case "play-playlist-episode":
                    log.info("Play next episode event received")
                    playPlaylistEpisode(payload as string)
                    break
                case "start-onlinestream-watch-party":
                    break
                case "get-text-tracks":
                    log.info("Get text tracks event received")
                    let textTracks: { type: "subtitles" | "captions", label: string, language: string, number: number }[] = []
                    if (subtitleManager) {
                        textTracks = subtitleManager.getTracks().map(n => ({
                            number: n.number,
                            type: "subtitles",
                            label: n.label || "",
                            language: n.language || n.languageIETF || "",
                        }))
                    } else if (mediaCaptionsManager) {
                        textTracks = mediaCaptionsManager.getTracks().map(n => ({
                            number: n.number,
                            type: "captions",
                            label: n.label,
                            language: n.language,
                        }))
                    }
                    sendEvent("video-text-tracks", { textTracks })
                    break
                case "translated-text":
                    const p = payload as { original: string, translated: string }
                    subtitleManager?.processEventTranslationQueue?.(p.original, p.translated)
                    break
                default:
                    log.warn("Unknown event received", type)
            }
        },
    })

    return {
        dispatchTerminatedEvent,
        dispatchVideoCompletedEvent,
        dispatchVideoLoadedEvent,
        dispatchVideoErrorEvent,
        dispatchCanPlayEvent,
        dispatchTranslateTextEvent,
        dispatchTranslateSubtitleTrackEvent,
    }
}

export function useVideoCoreEvents() {
    const { sendMessage } = useWebsocketSender()
    const clientId = useAtomValue(clientIdAtom)

    function sendEvent<T extends Record<string, any> | void = void>(type: VideoCore_ClientEventType, payload?: T) {
        sendMessage({
            type: WSEvents.VIDEOCORE,
            payload: {
                clientId: clientId,
                type: type,
                payload: payload,
            },
        })
    }

    return {
        sendEvent: sendEvent,
    }
}

