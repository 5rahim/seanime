import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { MKVParser_SubtitleEvent, MKVParser_TrackInfo, NativePlayer_PlaybackInfo, NativePlayer_ServerEvent } from "@/api/generated/types"
import { useUpdateAnimeEntryProgress } from "@/api/hooks/anime_entries.hooks"
import { useHandleCurrentMediaContinuity } from "@/api/hooks/continuity.hooks"
import { __seaMediaPlayer_autoNextAtom } from "@/app/(main)/_features/sea-media-player/sea-media-player.atoms"
import { vc_dispatchAction, vc_miniPlayer, vc_subtitleManager, vc_videoElement, VideoCore } from "@/app/(main)/_features/video-core/video-core"
import { clientIdAtom } from "@/app/websocket-provider"
import { logger } from "@/lib/helpers/debug"
import { WSEvents } from "@/lib/server/ws-events"
import { useQueryClient } from "@tanstack/react-query"
import { useAtom, useAtomValue } from "jotai"
import { useSetAtom } from "jotai/react"
import React from "react"
import { toast } from "sonner"
import { useWebsocketMessageListener, useWebsocketSender } from "../../_hooks/handle-websockets"
import { useServerStatus } from "../../_hooks/use-server-status"
import { useSkipData } from "../sea-media-player/aniskip"
import { nativePlayer_stateAtom } from "./native-player.atoms"

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
    VIDEO_TERMINATED = "video-terminated",
    VIDEO_TIME_UPDATE = "video-time-update",
}

const log = logger("NATIVE PLAYER")

export function NativePlayer() {
    const serverStatus = useServerStatus()
    const clientId = useAtomValue(clientIdAtom)
    const { sendMessage } = useWebsocketSender()

    const autoPlayNext = useAtomValue(__seaMediaPlayer_autoNextAtom)
    const videoElement = useAtomValue(vc_videoElement)
    const [state, setState] = useAtom(nativePlayer_stateAtom)
    const [miniPlayer, setMiniPlayer] = useAtom(vc_miniPlayer)
    const subtitleManager = useAtomValue(vc_subtitleManager)
    const dispatchEvent = useSetAtom(vc_dispatchAction)

    // Continuity
    const { watchHistory, waitForWatchHistory, getEpisodeContinuitySeekTo } = useHandleCurrentMediaContinuity(state?.playbackInfo?.media?.id)

    // AniSkip
    const { data: aniSkipData } = useSkipData(state?.playbackInfo?.media?.idMal, state?.playbackInfo?.episode?.progressNumber ?? -1)

    //
    // Start
    //

    const qc = useQueryClient()

    React.useEffect(() => {
        qc.invalidateQueries({ queryKey: [API_ENDPOINTS.CONTINUITY.GetContinuityWatchHistoryItem.key] })
    }, [state])


    // Update progress
    const { mutate: updateProgress, isPending: isUpdatingProgress, isSuccess: isProgressUpdateSuccess } = useUpdateAnimeEntryProgress(
        state.playbackInfo?.media?.id,
        state.playbackInfo?.episode?.progressNumber ?? 0,
    )

    const handleTimeInterval = () => {
        if (videoElement) {
            sendMessage({
                type: WSEvents.NATIVE_PLAYER,
                payload: {
                    clientId: clientId,
                    type: VideoPlayerEvents.VIDEO_TIME_UPDATE,
                    payload: {
                        currentTime: videoElement.currentTime,
                        duration: videoElement.duration,
                        paused: videoElement.paused,
                    },
                },
            })
        }
    }

    // Time update interval
    React.useEffect(() => {
        const interval = setInterval(handleTimeInterval, 2000)
        return () => clearInterval(interval)
    }, [videoElement])

    //
    // Event Handlers
    //

    const handleCompleted = () => {
        const v = videoElement
        if (!v) return
        sendMessage({
            type: WSEvents.NATIVE_PLAYER,
            payload: {
                clientId: clientId,
                type: VideoPlayerEvents.VIDEO_COMPLETED,
                payload: {
                    currentTime: v.currentTime,
                    duration: v.duration,
                },
            },
        })
    }

    const handleTimeUpdate = () => {
        const v = videoElement
        if (!v) return

    }

    const handleEnded = () => {
        log.info("Ended")

        sendMessage({
            type: WSEvents.NATIVE_PLAYER,
            payload: {
                clientId: clientId,
                type: VideoPlayerEvents.VIDEO_ENDED,
                payload: {
                    autoNext: autoPlayNext,
                },
            },
        })

    }

    const handleError = (value: string) => {
        const v = videoElement
        if (!v) return

        const error = value || v.error
        let errorMessage = value || "Unknown error"
        let detailedInfo = ""

        if (error instanceof MediaError) {
            switch (error.code) {
                case MediaError.MEDIA_ERR_ABORTED:
                    errorMessage = "Media playback aborted"
                    break
                case MediaError.MEDIA_ERR_NETWORK:
                    errorMessage = "Network error occurred"
                    break
                case MediaError.MEDIA_ERR_DECODE:
                    errorMessage = "Media decode error - codec not supported or corrupted file"
                    detailedInfo = "This is likely a codec compatibility issue."
                    break
                case MediaError.MEDIA_ERR_SRC_NOT_SUPPORTED:
                    errorMessage = "Media format not supported"
                    detailedInfo = "The video codec/container format is not supported."
                    break
                default:
                    errorMessage = error.message || "Unknown media error"
            }
            log.error("Media error", {
                code: error?.code,
                message: error?.message,
                src: v.src,
                networkState: v.networkState,
                readyState: v.readyState,
            })
        }


        const fullErrorMessage = detailedInfo ? `${errorMessage}\n\n${detailedInfo}` : errorMessage

        log.error("Media error", fullErrorMessage)

        sendMessage({
            type: WSEvents.NATIVE_PLAYER,
            payload: {
                clientId: clientId,
                type: VideoPlayerEvents.VIDEO_ERROR,
                payload: { error: fullErrorMessage },
            },
        })
    }

    const handleSeeked = (currentTime: number) => {
        const v = videoElement
        if (!v) return

        log.info("Video seeked to", currentTime)

        sendMessage({
            type: WSEvents.NATIVE_PLAYER,
            payload: {
                clientId: clientId,
                type: VideoPlayerEvents.VIDEO_SEEKED,
                payload: { currentTime: currentTime, duration: v.duration },
            },
        })
    }

    /**
     * Metadata is loaded
     * - Handle captions
     * - Initialize the subtitle manager if the stream is MKV
     * - Initialize the audio manager if the stream is MKV
     * - Initialize the thumbnailer if the stream is local file
     */
    const handleLoadedMetadata = () => {
        const v = videoElement
        if (!v) return


        sendMessage({
            type: WSEvents.NATIVE_PLAYER,
            payload: {
                clientId: clientId,
                type: VideoPlayerEvents.LOADED_METADATA,
                payload: {
                    currentTime: v.currentTime,
                    duration: v.duration,
                },
            },
        })

        if (state.playbackInfo?.episode?.progressNumber && watchHistory?.found && watchHistory.item?.episodeNumber === state.playbackInfo?.episode?.progressNumber) {
            const lastWatchedTime = getEpisodeContinuitySeekTo(state.playbackInfo?.episode?.progressNumber,
                videoElement?.currentTime,
                videoElement?.duration)
            logger("MEDIA PLAYER").info("Watch continuity: Seeking to last watched time", { lastWatchedTime })
            if (lastWatchedTime > 0) {
                logger("MEDIA PLAYER").info("Watch continuity: Seeking to", lastWatchedTime)
                dispatchEvent({ type: "restoreProgress", payload: { time: lastWatchedTime } })
                // const isPaused = videoElement?.paused
                // videoElement?.play?.()
                // setTimeout(() => {
                //
                //     if (isPaused) {
                //         setTimeout(() => {
                //             videoElement?.pause?.()
                //         }, 200)
                //     }
                // }, 1000)
            }
        }
    }

    const handlePause = () => {
        const v = videoElement
        if (!v) return

        sendMessage({
            type: WSEvents.NATIVE_PLAYER,
            payload: {
                clientId: clientId,
                type: VideoPlayerEvents.VIDEO_PAUSED,
                payload: {
                    currentTime: v.currentTime,
                    duration: v.duration,
                },
            },
        })
    }

    const handlePlay = () => {
        const v = videoElement
        if (!v) return

        sendMessage({
            type: WSEvents.NATIVE_PLAYER,
            payload: {
                clientId: clientId,
                type: VideoPlayerEvents.VIDEO_RESUMED,
                payload: {
                    currentTime: v.currentTime,
                    duration: v.duration,
                },
            },
        })
    }

    function handleFileUploaded(data: { name: string, content: string }) {
        sendMessage({
            type: WSEvents.NATIVE_PLAYER,
            payload: {
                clientId: clientId,
                type: VideoPlayerEvents.SUBTITLE_FILE_UPLOADED,
                payload: { filename: data.name, content: data.content },
            },
        })
    }

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
                        draft.loadingState = payload as string
                        draft.playbackInfo = null
                        draft.playbackError = null
                        return
                    })
                    setMiniPlayer(false)

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
                    setMiniPlayer(false)
                    break
                // 3. Subtitle event (MKV)
                // We receive the subtitle events after the server received the loaded-metadata event
                case "subtitle-event":
                    subtitleManager?.onSubtitleEvent(payload as MKVParser_SubtitleEvent)
                    break
                case "add-subtitle-track":
                    subtitleManager?.onTrackAdded(payload as MKVParser_TrackInfo)
                    break
                case "terminate":
                    log.info("Terminate event received")
                    handleTerminateStream()
                    break
                case "error":
                    log.error("Error event received", payload)
                    toast.error("An error occurred while playing the stream. " + ((payload as { error: string }).error))
                    setState(draft => {
                        draft.playbackError = (payload as { error: string }).error
                        return
                    })
                    break
            }
        },
    })

    //
    // Handlers
    //

    function handleTerminateStream() {
        // Clean up player first
        if (videoElement) {
            log.info("Cleaning up media")
            videoElement.pause()
        }

        setMiniPlayer(true)
        setState(draft => {
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
        }, 700)

        sendMessage({
            type: WSEvents.NATIVE_PLAYER,
            payload: {
                clientId: clientId,
                type: VideoPlayerEvents.VIDEO_TERMINATED,
            },
        })
    }

    return (
        <>

            <VideoCore
                state={state}
                aniSkipData={aniSkipData}
                onTerminateStream={handleTerminateStream}
                onLoadedMetadata={handleLoadedMetadata}
                onTimeUpdate={handleTimeUpdate}
                onEnded={handleEnded}
                onSeeked={handleSeeked}
                onCompleted={handleCompleted}
                onError={handleError}
                onPlay={handlePlay}
                onPause={handlePause}
                onFileUploaded={handleFileUploaded}
            />
        </>
    )
}
