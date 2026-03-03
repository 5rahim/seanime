import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { MKVParser_SubtitleEvent, NativePlayer_PlaybackInfo, NativePlayer_ServerEvent } from "@/api/generated/types"
import { vc_subtitleManager } from "@/app/(main)/_features/video-core/video-core"
import { VideoCore } from "@/app/(main)/_features/video-core/video-core"
import { vc_miniPlayer } from "@/app/(main)/_features/video-core/video-core-atoms"
import { vc_videoElement } from "@/app/(main)/_features/video-core/video-core-atoms"
import { VideoCoreLifecycleState } from "@/app/(main)/_features/video-core/video-core.atoms"
import { clientIdAtom } from "@/app/websocket-provider"
import { logger } from "@/lib/helpers/debug"
import { WSEvents } from "@/lib/server/ws-events"
import { useQueryClient } from "@tanstack/react-query"
import { useAtom, useAtomValue } from "jotai"
import React from "react"
import { toast } from "sonner"
import { useWebsocketMessageListener, useWebsocketSender } from "../../_hooks/handle-websockets"
import { useSkipData } from "../video-core/_lib/aniskip"
import { nativePlayer_stateAtom } from "./native-player.atoms"

const log = logger("NATIVE PLAYER")

// minimum interval between subtitle event flushes
const SUBTITLE_FLUSH_INTERVAL_MS = 300

export function NativePlayer() {
    const qc = useQueryClient()
    const clientId = useAtomValue(clientIdAtom)
    const { sendMessage } = useWebsocketSender()

    const videoElement = useAtomValue(vc_videoElement)
    const [state, setState] = useAtom(nativePlayer_stateAtom)
    const [miniPlayer, setMiniPlayer] = useAtom(vc_miniPlayer)
    const subtitleManager = useAtomValue(vc_subtitleManager)

    // AniSkip
    const { data: aniSkipData } = useSkipData(state?.playbackInfo?.media?.idMal, state?.playbackInfo?.episode?.progressNumber ?? -1)

    React.useEffect(() => {
        qc.invalidateQueries({ queryKey: [API_ENDPOINTS.CONTINUITY.GetContinuityWatchHistoryItem.key] })
    }, [state])

    //
    // Subtitle event buffering
    // Accumulate incoming subtitle events and flush them to the subtitle manager
    //

    const subtitleBufferRef = React.useRef<MKVParser_SubtitleEvent[]>([])
    const subtitleFlushTimerRef = React.useRef<ReturnType<typeof setTimeout> | null>(null)
    const subtitleIdleHandleRef = React.useRef<number | null>(null)
    const subtitleManagerRef = React.useRef(subtitleManager)
    subtitleManagerRef.current = subtitleManager

    const flushSubtitleBuffer = React.useCallback(() => {
        subtitleFlushTimerRef.current = null
        subtitleIdleHandleRef.current = null

        const events = subtitleBufferRef.current
        if (events.length === 0) return
        subtitleBufferRef.current = []

        // process outside the websocket message handler
        subtitleManagerRef.current?.onSubtitleEvents(events)?.then()
    }, [])

    const scheduleSubtitleFlush = React.useCallback(() => {
        if (subtitleFlushTimerRef.current !== null) return // already scheduled

        // with a deadline so events don't pile up
        if (typeof requestIdleCallback !== "undefined") {
            subtitleIdleHandleRef.current = requestIdleCallback(() => {
                flushSubtitleBuffer()
            }, { timeout: SUBTITLE_FLUSH_INTERVAL_MS })
        }

        // guarantee a flush even if idle callback doesn't fire in time
        subtitleFlushTimerRef.current = setTimeout(() => {
            if (subtitleIdleHandleRef.current !== null) {
                cancelIdleCallback(subtitleIdleHandleRef.current)
                subtitleIdleHandleRef.current = null
            }
            flushSubtitleBuffer()
        }, SUBTITLE_FLUSH_INTERVAL_MS)
    }, [flushSubtitleBuffer])

    // cleanup subtitle buffer timers on unmount
    React.useEffect(() => {
        return () => {
            if (subtitleFlushTimerRef.current !== null) {
                clearTimeout(subtitleFlushTimerRef.current)
            }
            if (subtitleIdleHandleRef.current !== null && typeof cancelIdleCallback !== "undefined") {
                cancelIdleCallback(subtitleIdleHandleRef.current)
            }
        }
    }, [])

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
                case "abort-open":
                    log.info("Abort open event received", { payload })
                    setState(draft => {
                        draft.loadingState = "An error occurred while loading the stream: " + ((payload as string) || "Unknown error")
                        draft.playbackError = payload as string
                        draft.playbackInfo = null
                        return
                    })
                    setTimeout(() => {
                        handleTerminateStream()
                    }, 3000)

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
                // We receive the subtitle events after the server received the loaded-metadata event.
                // Buffer the events and process them off the main thread
                case "subtitle-event":
                    if (Array.isArray(payload)) {
                        subtitleBufferRef.current.push(...(payload as MKVParser_SubtitleEvent[]))
                    } else {
                        subtitleBufferRef.current.push(payload as MKVParser_SubtitleEvent)
                    }
                    scheduleSubtitleFlush()
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
            type: WSEvents.VIDEOCORE,
            payload: {
                clientId: clientId,
                type: "video-terminated",
            },
        })
    }

    const ps = React.useMemo<VideoCoreLifecycleState>(() => {
        return {
            active: state.active,
            loadingState: state.loadingState,
            playbackError: state.playbackError,
            playbackInfo: {
                id: state.playbackInfo?.id!,
                playbackType: state.playbackInfo?.streamType!,
                streamUrl: state.playbackInfo?.streamUrl!,
                streamPath: state.playbackInfo?.streamPath,
                mkvMetadata: state.playbackInfo?.mkvMetadata,
                media: state.playbackInfo?.media,
                episode: state.playbackInfo?.episode,
                localFile: state.playbackInfo?.localFile,
                streamType: "native",
            },
        }
    }, [state])

    return (
        <>
            <VideoCore
                id="native-player"
                state={ps}
                aniSkipData={aniSkipData}
                onTerminateStream={handleTerminateStream}
            />
        </>
    )
}
