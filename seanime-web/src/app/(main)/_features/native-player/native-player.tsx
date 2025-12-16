import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { MKVParser_SubtitleEvent, NativePlayer_PlaybackInfo, NativePlayer_ServerEvent } from "@/api/generated/types"
import { useUpdateAnimeEntryProgress } from "@/api/hooks/anime_entries.hooks"
import { vc_miniPlayer, vc_subtitleManager, vc_videoElement, VideoCore } from "@/app/(main)/_features/video-core/video-core"
import { VideoCoreLifecycleState } from "@/app/(main)/_features/video-core/video-core.atoms"
import { clientIdAtom } from "@/app/websocket-provider"
import { logger } from "@/lib/helpers/debug"
import { WSEvents } from "@/lib/server/ws-events"
import { useQueryClient } from "@tanstack/react-query"
import { useAtom, useAtomValue } from "jotai"
import React from "react"
import { toast } from "sonner"
import { useWebsocketMessageListener, useWebsocketSender } from "../../_hooks/handle-websockets"
import { useSkipData } from "../sea-media-player/aniskip"
import { nativePlayer_stateAtom } from "./native-player.atoms"

const log = logger("NATIVE PLAYER")

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


    // Update progress
    const { mutate: updateProgress, isPending: isUpdatingProgress, isSuccess: isProgressUpdateSuccess } = useUpdateAnimeEntryProgress(
        state.playbackInfo?.media?.id,
        state.playbackInfo?.episode?.progressNumber ?? 0,
        false,
    )

    //
    // Event Handlers
    //

    const handleCompleted = () => {
        const v = videoElement
        if (!v) return
        if (state.playbackInfo?.media && state.playbackInfo?.episode) {
            updateProgress({
                mediaId: state.playbackInfo?.media?.id,
                totalEpisodes: state.playbackInfo?.media?.episodes || 0,
                episodeNumber: state.playbackInfo?.episode?.progressNumber,
            })
        }
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
                // We receive the subtitle events after the server received the loaded-metadata event
                case "subtitle-event":
                    subtitleManager?.onSubtitleEvent(payload as MKVParser_SubtitleEvent)
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
                onCompleted={handleCompleted}
            />
        </>
    )
}
